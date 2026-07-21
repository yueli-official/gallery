package server

import (
	"github.com/gogf/gf/v2/net/ghttp"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"platform/gokit/authhttp"
	"platform/gokit/ghttpx"
	"platform/gokit/healthcheck"
	"platform/products/gallery/api/internal/controller"
	galleryservice "platform/products/gallery/api/internal/gallery"
)

type Deps struct {
	Gallery     *galleryservice.Service
	Verifier    *foundationauth.Verifier
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
			group.Middleware(ghttpx.Middleware, authhttp.Optional(deps.Verifier))
			group.Bind(controller.NewPublic(deps.Gallery))
		})
		s.Group("/", func(group *ghttp.RouterGroup) {
			if deps.Verifier != nil {
				group.Middleware(ghttpx.Middleware, authhttp.Required(deps.Verifier))
			} else {
				// OpenAPI export has no runtime verifier, but protected route shapes
				// still belong in the generated contract.
				group.Middleware(ghttpx.Middleware)
			}
			group.Bind(controller.NewWorkflow(deps.Gallery))
			group.Bind(controller.NewAdmin(deps.Gallery))
		})
	}
}
