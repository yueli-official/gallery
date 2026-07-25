package controller

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
	"platform/products/gallery/api/internal/galleryauthz"
	"platform/products/gallery/api/internal/galleryerr"
	"platform/products/gallery/api/internal/model"
)

type authorizationContextKey struct{}

func AuthorizationMiddleware(service *galleryauthz.Service) ghttp.HandlerFunc {
	return func(request *ghttp.Request) {
		ctx := context.WithValue(request.Context(), authorizationContextKey{}, service)
		correlationID := strings.TrimSpace(request.Header.Get("X-Trace-Id"))
		if correlationID == "" {
			correlationID = strings.TrimSpace(request.Header.Get("X-Request-Id"))
		}
		ctx = authorization.WithRequestMetadata(ctx, authorization.RequestMetadata{CorrelationID: correlationID})
		request.SetCtx(ctx)
		request.Middleware.Next()
	}
}

func authorizationService(ctx context.Context) *galleryauthz.Service {
	service, _ := ctx.Value(authorizationContextKey{}).(*galleryauthz.Service)
	return service
}

func optionalSubject(ctx context.Context) (model.Subject, bool) {
	principal, ok := foundationauth.FromContext(ctx)
	if !ok || principal == nil || strings.TrimSpace(principal.Subject) == "" {
		return model.Subject{}, false
	}
	kindClaim, _ := principal.Claim("subject_kind")
	kind := valueString(kindClaim)
	if kind == "" {
		kind = "user"
	}
	verifiedClaim, _ := principal.Claim("email_verified")
	return model.Subject{Kind: kind, ID: principal.Subject, Verified: claimBool(verifiedClaim), Bearer: bearerOf(ctx)}, true
}

func bearerOf(ctx context.Context) string {
	request := ghttp.RequestFromCtx(ctx)
	if request == nil {
		return ""
	}
	value := strings.TrimSpace(request.Request.Header.Get("Authorization"))
	if len(value) < 7 || !strings.EqualFold(value[:7], "Bearer ") {
		return ""
	}
	return strings.TrimSpace(value[7:])
}

func requiredSubject(ctx context.Context) (model.Subject, error) {
	subject, ok := optionalSubject(ctx)
	if !ok {
		return model.Subject{}, galleryerr.Forbidden()
	}
	return subject, nil
}

func requiredUser(ctx context.Context) (model.Subject, error) {
	subject, err := requiredSubject(ctx)
	if err != nil {
		return model.Subject{}, err
	}
	if subject.Kind != "user" {
		return model.Subject{}, galleryerr.Forbidden()
	}
	return subject, nil
}

func isAdmin(ctx context.Context) bool {
	service := authorizationService(ctx)
	return service != nil && service.IsAdministrator(ctx)
}

func requireAdmin(ctx context.Context) (string, error) {
	return requireCapability(ctx, authorization.CapabilityManage)
}

func requireCapability(ctx context.Context, capability authorization.CapabilityKey) (string, error) {
	service := authorizationService(ctx)
	if service == nil {
		return "", galleryerr.AuthorizationUnavailable()
	}
	decision, err := service.Decide(ctx, capability)
	if err != nil {
		if authorization.Is(err, authorization.ErrorUnavailable) {
			return "", galleryerr.AuthorizationUnavailable()
		}
		return "", galleryerr.Forbidden()
	}
	if !decision.Allowed {
		return "", galleryerr.Forbidden()
	}
	subject := service.Subject(ctx)
	if subject.ID == "" {
		return "", galleryerr.Forbidden()
	}
	return subject.ID, nil
}

func mapAuthorizationError(err error) error {
	switch {
	case authorization.Is(err, authorization.ErrorDenied):
		return galleryerr.Forbidden()
	case authorization.Is(err, authorization.ErrorUnavailable):
		return galleryerr.AuthorizationUnavailable()
	case authorization.Is(err, authorization.ErrorNotFound):
		return galleryerr.NotFound("authorization", "")
	case authorization.Is(err, authorization.ErrorInvalidInput),
		authorization.Is(err, authorization.ErrorConflict),
		authorization.Is(err, authorization.ErrorExpired):
		return galleryerr.Validation("authorization", err.Error())
	default:
		return galleryerr.AuthorizationUnavailable()
	}
}
