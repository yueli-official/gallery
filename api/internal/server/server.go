package server

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"platform/gokit/authjwt"
	"platform/gokit/ghttpx"
	"platform/gokit/healthcheck"
	"platform/products/gallery/api/internal/assetclient"
	"platform/products/gallery/api/internal/controller"
	galleryservice "platform/products/gallery/api/internal/gallery"
)

type Deps struct {
	Gallery     *galleryservice.Service
	Verifier    *authjwt.Verifier
	Assets      assetclient.Client
	ReadyChecks map[string]healthcheck.Check
}

func Configure(s *ghttp.Server, deps Deps) {
	checks := deps.ReadyChecks
	if checks == nil {
		checks = map[string]healthcheck.Check{"database": healthcheck.Database}
	}
	s.Use(ghttpx.TraceRouteMiddleware)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttpx.Middleware)
		group.GET("/healthz", controller.Healthz)
		group.GET("/readyz", healthcheck.Handler(checks))
	})
	if deps.Gallery != nil {
		s.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(ghttpx.Middleware)
			group.Bind(controller.NewPublic(deps.Gallery))
		})
		s.Group("/", func(group *ghttp.RouterGroup) {
			if deps.Verifier != nil {
				group.Middleware(ghttpx.Middleware, authjwt.Middleware(deps.Verifier))
			} else {
				// OpenAPI export has no runtime verifier, but protected route shapes
				// still belong in the generated contract.
				group.Middleware(ghttpx.Middleware)
			}
			group.Bind(controller.NewWorkflow(deps.Gallery, deps.Assets))
			group.Bind(controller.NewAdmin(deps.Gallery))
		})
	}
}
