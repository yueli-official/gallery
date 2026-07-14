package appconfig

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

type JWKS struct {
	URL      string
	Issuer   string
	Audience string
}

func LoadJWKS(ctx context.Context) JWKS {
	return JWKS{
		URL:      g.Cfg().MustGet(ctx, "gallery.jwks.url", "http://localhost:8081/oauth2/jwks.json").String(),
		Issuer:   g.Cfg().MustGet(ctx, "gallery.jwks.issuer", "http://localhost:8081").String(),
		Audience: g.Cfg().MustGet(ctx, "gallery.jwks.audience", "").String(),
	}
}

func SiteBrand(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "gallery.brand", "月离图库").String()
}

func SiteSlug(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "gallery.siteSlug", "gallery-ae").String()
}

func AssetBaseURL(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "gallery.asset.baseUrl", "http://localhost:8082").String()
}
