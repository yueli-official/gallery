package controller

import (
	"context"
	"slices"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"platform/gokit/authjwt"
	"platform/products/gallery/api/internal/galleryerr"
	"platform/products/gallery/api/internal/model"
)

func optionalSubject(ctx context.Context) (model.Subject, bool) {
	principal, ok := authjwt.From(ctx)
	if !ok || principal == nil || strings.TrimSpace(principal.Subject) == "" {
		return model.Subject{}, false
	}
	kind := valueString(principal.Claims["subject_kind"])
	if kind == "" {
		kind = "user"
	}
	return model.Subject{Kind: kind, ID: principal.Subject, Verified: claimBool(principal.Claims["email_verified"]), Bearer: bearerOf(ctx)}, true
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

func requireAdmin(ctx context.Context) (string, error) {
	principal, ok := authjwt.From(ctx)
	if !ok || !slices.Contains(g.Cfg().MustGet(ctx, "gallery.operatorSubs").Strings(), principal.Subject) {
		return "", galleryerr.Forbidden()
	}
	return principal.Subject, nil
}
