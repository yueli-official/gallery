package server

import (
	"github.com/gogf/gf/v2/net/ghttp"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/gallery/api/internal/controller"
	galleryservice "github.com/yueli-official/gallery/api/internal/gallery"
	"github.com/yueli-official/gallery/api/internal/galleryauthz"
	"github.com/yueli-official/gallery/api/internal/gallerycomments"
	"github.com/yueli-official/gallery/api/internal/runtime"
)

type Deps struct {
	Gallery       *galleryservice.Service
	Verifier      *foundationauth.Verifier
	Authorization *galleryauthz.Service
	Comments      *gallerycomments.Module
	ReadyChecks   map[string]runtime.ReadinessCheck
}

func Configure(s *ghttp.Server, deps Deps) {
	apiMiddleware := runtime.MustAPIMiddleware(runtime.MustRateLimiterFromEnvironment()).Handle
	checks := deps.ReadyChecks
	if checks == nil {
		checks = map[string]runtime.ReadinessCheck{"database": runtime.DatabaseReadiness}
	}
	s.Use(runtime.TraceRouteMiddleware)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(apiMiddleware)
		group.GET("/healthz", controller.Healthz)
		group.GET("/readyz", runtime.ReadinessHandler(checks))
	})
	if deps.Gallery != nil {
		s.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(
				apiMiddleware,
				runtime.OptionalAuth(deps.Verifier),
				controller.AuthorizationMiddleware(deps.Authorization),
			)
			group.Bind(controller.NewPublic(deps.Gallery))
			if deps.Comments != nil {
				group.Bind(controller.NewPublicComments(deps.Comments))
			}
		})
		s.Group("/", func(group *ghttp.RouterGroup) {
			if deps.Verifier != nil {
				group.Middleware(
					apiMiddleware,
					runtime.RequiredAuth(deps.Verifier),
					controller.AuthorizationMiddleware(deps.Authorization),
				)
			} else {
				// OpenAPI export has no runtime verifier, but protected route shapes
				// still belong in the generated contract.
				group.Middleware(apiMiddleware, controller.AuthorizationMiddleware(deps.Authorization))
			}
			group.Bind(controller.NewWorkflow(deps.Gallery))
			group.Bind(controller.NewAdmin(deps.Gallery))
			group.Bind(controller.NewMe())
			group.Bind(controller.NewAuthorization())
			if deps.Comments != nil {
				group.Bind(controller.NewComments(deps.Comments))
			}
		})
	}
}
