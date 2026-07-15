package main

import (
	"context"
	"os"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"

	"platform/gokit/authjwt"
	"platform/gokit/observability"
	"platform/gokit/openapiexport"
	"platform/products/gallery/api/internal/appconfig"
	"platform/products/gallery/api/internal/assetclient"
	"platform/products/gallery/api/internal/classificationwatcher"
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

	service := galleryservice.New(dao.NewPG(g.DB()))
	service.SetAssetReferencePort(assetclient.NewHTTP(appconfig.AssetBaseURL(ctx), appconfig.SiteSlug(ctx)))
	watcher, watcherErr := classificationwatcher.Start(g.DB().GetConfig(), func() {
		refreshCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if refreshErr := service.RefreshClassificationCatalog(refreshCtx); refreshErr != nil {
			g.Log().Warning(refreshCtx, "gallery classification catalog refresh failed", "error", refreshErr)
		}
	})
	if watcherErr != nil {
		g.Log().Warning(ctx, "gallery classification listener unavailable; request-time revision checks remain active", "error", watcherErr)
	} else {
		defer watcher.Close()
	}
	jwks := appconfig.LoadJWKS(ctx)
	verifier, err := authjwt.NewVerifier(authjwt.VerifierConfig{
		Keys: authjwt.NewRemoteKeySource(jwks.URL), Issuer: jwks.Issuer, Audience: jwks.Audience,
	})
	if err != nil {
		panic(err)
	}
	server.Configure(httpServer, server.Deps{Gallery: service, Verifier: verifier})
	g.Log().Info(ctx, "gallery service starting")
	httpServer.Run()
}
