package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	foundationabuse "github.com/yueli-official/foundation/go/abuse"
	"github.com/yueli-official/foundation/go/abuse/turnstile"
	"github.com/yueli-official/foundation/go/authorization"
	authorizationpostgres "github.com/yueli-official/foundation/go/authorization/postgres"
	"github.com/yueli-official/foundation/go/work"
	workpostgres "github.com/yueli-official/foundation/go/work/postgres"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"

	"platform/gokit/authsetup"
	"platform/gokit/observability"
	"platform/gokit/openapiexport"
	"platform/gokit/postgresdb"
	"platform/gokit/webhooksetup"
	"platform/products/gallery/api/internal/appconfig"
	"platform/products/gallery/api/internal/assetclient"
	"platform/products/gallery/api/internal/dao"
	galleryservice "platform/products/gallery/api/internal/gallery"
	"platform/products/gallery/api/internal/galleryabuse"
	"platform/products/gallery/api/internal/galleryauthz"
	"platform/products/gallery/api/internal/gallerywebhook"
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
	authorizationDefinition, err := authorization.Compile(galleryauthz.Definition())
	if err != nil {
		panic(err)
	}
	if os.Getenv("PLATFORM_OPENAPI_OUTPUT") != "" {
		authz, err := authorization.NewMemory(authorizationDefinition, authorization.MemoryOptions{
			RootScopeID: galleryauthz.RootScopeID,
			ProtectedSubjects: []authorization.SubjectRef{{
				Kind: authorization.SubjectUser, ID: "openapi-export-admin",
			}},
			Predicates: galleryauthz.PredicateEvaluators(),
		})
		if err != nil {
			panic(err)
		}
		server.Configure(httpServer, server.Deps{
			Gallery: galleryservice.New(nil), Authorization: galleryauthz.New(authz),
		})
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
	bootstrapSubs := appconfig.BootstrapAdministratorSubs(ctx)
	protected := make([]authorization.SubjectRef, 0, len(bootstrapSubs))
	for _, sub := range bootstrapSubs {
		if sub != "" {
			protected = append(protected, authorization.SubjectRef{
				Kind: authorization.SubjectUser, ID: sub,
			})
		}
	}
	authz, err := authorizationpostgres.New(ctx, authorizationDefinition, authorizationpostgres.Options{
		DB: workDB, InstanceKey: "gallery:" + appconfig.SiteSlug(ctx),
		Memory: authorization.MemoryOptions{
			RootScopeID:       galleryauthz.RootScopeID,
			ProtectedSubjects: protected,
			Predicates:        galleryauthz.PredicateEvaluators(),
		},
	})
	if err != nil {
		panic(err)
	}
	if authz.InstanceWasCreated() && len(protected) == 0 {
		panic("gallery authorization bootstrap requires at least one administrator subject")
	}
	authorizationService := galleryauthz.New(authz)
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
	store := dao.NewPG(g.DB(), workAdapter)
	var webhookRuntime *webhooksetup.Runtime
	if g.Cfg().MustGet(ctx, "gallery.webhook.enabled").Bool() {
		masterKey, err := webhooksetup.DecodeMasterKey(
			g.Cfg().MustGet(ctx, "gallery.webhook.masterKey").String(),
		)
		if err != nil {
			panic(err)
		}
		webhookWorkerID := "gallery-webhook-worker"
		if hostname, hostErr := os.Hostname(); hostErr == nil && hostname != "" {
			webhookWorkerID += ":" + hostname
		}
		webhookRuntime, err = webhooksetup.New(ctx, webhooksetup.Options{
			DB: workDB, InstanceKey: "gallery:" + appconfig.SiteSlug(ctx) + ":webhook",
			Definition: gallerywebhook.Definition(appconfig.SiteSlug(ctx)),
			MasterKey:  masterKey, WorkerID: webhookWorkerID,
			OnError: func(runErr error) {
				g.Log().Warning(ctx, "gallery webhook runner error", "error", runErr)
			},
		})
		if err != nil {
			panic(err)
		}
		store.SetWebhook(webhookRuntime.Hooks)
		webhookContext, stopWebhook := context.WithCancel(context.Background())
		defer stopWebhook()
		go func() {
			if runErr := webhookRuntime.Runner.Run(webhookContext); runErr != nil && !errors.Is(runErr, context.Canceled) {
				g.Log().Error(ctx, "gallery webhook runner stopped", "error", runErr)
			}
		}()
	}
	service := galleryservice.New(store)
	var (
		abuseChallenge *foundationabuse.ChallengeDefinition
		abuseVerifiers map[foundationabuse.ChallengeKind]foundationabuse.ChallengeVerifier
	)
	if secret := g.Cfg().MustGet(ctx, "gallery.abuse.turnstile.secret").String(); secret != "" {
		hostnames := g.Cfg().MustGet(ctx, "gallery.abuse.turnstile.hostnames").Strings()
		if len(hostnames) == 0 {
			panic("gallery.abuse.turnstile.hostnames is required when Turnstile is enabled")
		}
		challengeVerifier, err := turnstile.New(turnstile.Options{
			Secret:   secret,
			Endpoint: g.Cfg().MustGet(ctx, "gallery.abuse.turnstile.endpoint").String(),
		})
		if err != nil {
			panic(err)
		}
		abuseChallenge = &foundationabuse.ChallengeDefinition{
			Kind: "turnstile", ExpectedAction: "gallery-submission",
			AllowedHosts: hostnames,
		}
		abuseVerifiers = map[foundationabuse.ChallengeKind]foundationabuse.ChallengeVerifier{
			"turnstile": challengeVerifier,
		}
	}
	abuseCatalog := foundationabuse.MustCompile(galleryabuse.Definition(galleryabuse.Policy{
		Challenge: abuseChallenge,
	}))
	abuseModule, err := foundationabuse.NewPostgres(ctx, abuseCatalog, foundationabuse.PostgresOptions{
		DB: workDB, InstanceKey: "gallery:" + appconfig.SiteSlug(ctx),
		Verifiers: abuseVerifiers,
	})
	if err != nil {
		panic(err)
	}
	if err := service.SetAbuse(abuseModule); err != nil {
		panic(err)
	}
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
	server.Configure(httpServer, server.Deps{
		Gallery: service, Verifier: verifier, Authorization: authorizationService,
	})
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
