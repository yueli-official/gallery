package controller

import (
	"context"
	"slices"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"platform/gokit/authjwt"
	"platform/products/gallery/api/internal/galleryerr"
)

func subject(ctx context.Context) (string, error) {
	principal, ok := authjwt.From(ctx)
	if !ok {
		return "", galleryerr.Forbidden()
	}
	return principal.Subject, nil
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

func requireAdmin(ctx context.Context) (string, error) {
	principal, ok := authjwt.From(ctx)
	if !ok || !slices.Contains(g.Cfg().MustGet(ctx, "gallery.operatorSubs").Strings(), principal.Subject) {
		return "", galleryerr.Forbidden()
	}
	return principal.Subject, nil
}
