package main

import (
	"os"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"

	"platform/gokit/observability"
	"platform/gokit/openapiexport"
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
	server.Configure(httpServer, server.Deps{Gallery: service})
	g.Log().Info(ctx, "gallery service starting")
	httpServer.Run()
}
