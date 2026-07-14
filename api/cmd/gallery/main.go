package main

import (
	"os"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"

	"platform/gokit/authjwt"
	"platform/gokit/observability"
	"platform/gokit/openapiexport"
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

	service := galleryservice.New(dao.NewPG(g.DB()))
	assets := assetclient.NewHTTP(appconfig.AssetBaseURL(ctx), appconfig.SiteSlug(ctx))
	jwks := appconfig.LoadJWKS(ctx)
	verifier, err := authjwt.NewVerifier(authjwt.VerifierConfig{
		Keys: authjwt.NewRemoteKeySource(jwks.URL), Issuer: jwks.Issuer, Audience: jwks.Audience,
	})
	if err != nil {
		panic(err)
	}
	server.Configure(httpServer, server.Deps{Gallery: service, Verifier: verifier, Assets: assets})
	g.Log().Info(ctx, "gallery service starting")
	httpServer.Run()
}
