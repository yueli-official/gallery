package server

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"platform/gokit/ghttpx"
	"platform/gokit/healthcheck"
	"platform/products/gallery/api/internal/controller"
	galleryservice "platform/products/gallery/api/internal/gallery"
)

type Deps struct {
	Gallery     *galleryservice.Service
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
	}
}
