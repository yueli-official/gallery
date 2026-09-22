package server

import (
	"github.com/gogf/gf/v2/net/ghttp"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	goframeauth "github.com/yueli-official/foundation/go/goframe/auth"
	"github.com/yueli-official/gallery/api/internal/controller"
	galleryservice "github.com/yueli-official/gallery/api/internal/gallery"
	"github.com/yueli-official/gallery/api/internal/galleryauthz"
	"github.com/yueli-official/gallery/api/internal/gallerycomments"
	"github.com/yueli-official/gallery/api/internal/runtime"
)

type Deps struct {
	Gallery          *galleryservice.Service
	Verifier         *foundationauth.Verifier
	PersonalVerifier *foundationauth.PersonalTokenVerifier
	PersonalSite     string
	Authorization    *galleryauthz.Service
	Comments         *gallerycomments.Module
	ReadyChecks      map[string]runtime.ReadinessCheck
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
		var verifier goframeauth.TokenVerifier
		if deps.Verifier != nil {
			verifier = foundationauth.CompositeVerifier{JWT: deps.Verifier, Personal: deps.PersonalVerifier}
		}
		s.Group("/", func(group *ghttp.RouterGroup) {
			middlewares := []ghttp.HandlerFunc{apiMiddleware}
			if verifier != nil {
				middlewares = append(middlewares, runtime.OptionalAuth(verifier), controller.PersonalTokenRoutes)
			}
			middlewares = append(middlewares, controller.AuthorizationMiddleware(deps.Authorization))
			group.Middleware(middlewares...)
			group.Bind(controller.NewPublic(deps.Gallery))
			if deps.Comments != nil {
				group.Bind(controller.NewPublicComments(deps.Comments))
			}
		})
		s.Group("/", func(group *ghttp.RouterGroup) {
			if verifier != nil {
				group.Middleware(
					apiMiddleware,
					runtime.RequiredAuth(verifier),
					controller.PersonalTokenRoutes,
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
			group.Bind(controller.NewPersonalPermissions(deps.PersonalSite, deps.Authorization))
			if deps.Comments != nil {
				group.Bind(controller.NewComments(deps.Comments))
			}
		})
	}
}
