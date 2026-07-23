package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/yueli-official/foundation/go/work"
	workpostgres "github.com/yueli-official/foundation/go/work/postgres"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"

	"platform/gokit/authsetup"
	"platform/gokit/observability"
	"platform/gokit/openapiexport"
	"platform/gokit/postgresdb"
	"platform/products/gallery/api/internal/appconfig"
	"platform/products/gallery/api/internal/assetclient"
	"platform/products/gallery/api/internal/dao"
	galleryservice "platform/products/gallery/api/internal/gallery"
	"platform/products/gallery/api/internal/server"
)

func main() {
	ctx := gctx.New()
	shutdown, err := observability.StartFromEnvironment(ctx, "gallery-api")
	if err != nil {
		panic(err)
	}
	defer observability.ShutdownWithTimeout(shutdown)

	httpServer := g.Server()
	if os.Getenv("PLATFORM_OPENAPI_OUTPUT") != "" {
		server.Configure(httpServer, server.Deps{Gallery: galleryservice.New(nil)})
		if handled, exportErr := openapiexport.ExportIfRequested(httpServer); handled {
			if exportErr != nil {
				panic(exportErr)
			}
			return
		}
	}

	workDB, err := postgresdb.OpenDefault(ctx)
	if err != nil {
		panic(err)
	}
	defer workDB.Close()
	workCatalog, err := work.Compile(galleryservice.WorkDefinition())
	if err != nil {
		panic(err)
	}
	workAdapter, err := workpostgres.New(ctx, workCatalog, workpostgres.Options{
		DB: workDB, InstanceKey: "gallery:" + appconfig.SiteSlug(ctx),
	})
	if err != nil {
		panic(err)
	}
	service := galleryservice.New(dao.NewPG(g.DB(), workAdapter))
	assetCfg := appconfig.LoadAssetClient(ctx)
	assetPort, err := assetclient.NewHTTP(assetclient.Config{
		BaseURL: assetCfg.BaseURL, TokenURL: assetCfg.TokenURL, ClientID: assetCfg.ClientID,
		ClientSecret: assetCfg.ClientSecret, Scope: assetCfg.Scope, SiteKey: appconfig.SiteSlug(ctx),
	})
	if err != nil {
		panic(fmt.Sprintf("gallery asset client: %v", err))
	}
	service.SetAssetReferencePort(assetPort)
	processorCtx, stopProcessor := context.WithCancel(context.Background())
	defer stopProcessor()
	go runSubmissionProcessor(processorCtx, service)
	workerID := "gallery-worker"
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		workerID += ":" + hostname
	}
	runner, err := work.NewRunner(workCatalog, workAdapter, map[work.Kind]work.Handler{
		galleryservice.ClassificationRefreshKind: service.ClassificationRefreshHandler(),
	}, work.RunnerOptions{
		WorkerID: workerID, PollInterval: time.Second,
		OnError: func(err error) {
			g.Log().Warning(ctx, "gallery work runner error", "error", err)
		},
	})
	if err != nil {
		panic(err)
	}
	go func() {
		if err := runner.Run(processorCtx); err != nil && !errors.Is(err, context.Canceled) {
			g.Log().Error(ctx, "gallery work runner stopped", "error", err)
		}
	}()
	jwks := appconfig.LoadJWKS(ctx)
	verifier, err := authsetup.NewRemoteVerifier(authsetup.RemoteVerifierConfig{
		JWKSURL: jwks.URL, Issuer: jwks.Issuer, Audience: jwks.Audience,
		AllowLoopbackHTTP: jwks.AllowLoopbackHTTP,
	})
	if err != nil {
		panic(err)
	}
	server.Configure(httpServer, server.Deps{Gallery: service, Verifier: verifier})
	g.Log().Info(ctx, "gallery service starting")
	httpServer.Run()
}

func runSubmissionProcessor(ctx context.Context, service *galleryservice.Service) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	nextPublicationSweep := time.Time{}
	for {
		if time.Now().After(nextPublicationSweep) {
			if _, err := service.ReconcilePublishedImages(ctx, 100); err != nil && ctx.Err() == nil {
				g.Log().Warning(ctx, "gallery public rendition reconciliation failed", "error", err)
				nextPublicationSweep = time.Now().Add(30 * time.Second)
			} else {
				nextPublicationSweep = time.Now().Add(10 * time.Minute)
			}
		}
		processed, err := service.ProcessQueuedSubmissions(ctx, 8)
		if err != nil && ctx.Err() == nil {
			g.Log().Warning(ctx, "gallery submission processing batch failed", "error", err)
		}
		if processed >= 8 {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
