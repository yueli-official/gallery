package appconfig

import (
	"context"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

type JWKS struct {
	URL               string
	Issuer            string
	Audience          string
	AllowLoopbackHTTP bool
}

type AssetClient struct {
	BaseURL      string
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scope        string
}

func LoadJWKS(ctx context.Context) JWKS {
	return JWKS{
		URL:               g.Cfg().MustGet(ctx, "gallery.jwks.url", "http://localhost:8081/oauth2/jwks.json").String(),
		Issuer:            g.Cfg().MustGet(ctx, "gallery.jwks.issuer", "http://localhost:8081").String(),
		Audience:          g.Cfg().MustGet(ctx, "gallery.jwks.audience", "").String(),
		AllowLoopbackHTTP: g.Cfg().MustGet(ctx, "gallery.jwks.allowLoopbackHttp", false).Bool(),
	}
}

func IdentityBaseURL(ctx context.Context) string {
	if configured := strings.TrimSpace(g.Cfg().MustGet(ctx, "gallery.identity.baseUrl", "").String()); configured != "" {
		return strings.TrimRight(configured, "/")
	}
	return identityBaseFromJWKS(LoadJWKS(ctx).URL)
}

func identityBaseFromJWKS(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "http://localhost:8081"
	}
	parsed.Path = ""
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/")
}

func SiteSlug(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "gallery.siteSlug", "gallery-main").String()
}

func AssetNamespace(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "gallery.assetNamespace", "gallery").String()
}

func BootstrapAdministratorSubs(ctx context.Context) []string {
	return g.Cfg().MustGet(ctx, "gallery.authorization.bootstrapAdministratorSubs").Strings()
}

func AssetBaseURL(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "gallery.asset.baseUrl", "http://localhost:8082").String()
}

func LoadAssetClient(ctx context.Context) AssetClient {
	return AssetClient{
		BaseURL:      AssetBaseURL(ctx),
		TokenURL:     g.Cfg().MustGet(ctx, "gallery.asset.tokenUrl", "http://localhost:8081/oauth2/token").String(),
		ClientID:     g.Cfg().MustGet(ctx, "gallery.asset.clientId").String(),
		ClientSecret: g.Cfg().MustGet(ctx, "gallery.asset.clientSecret").String(),
		Scope:        g.Cfg().MustGet(ctx, "gallery.asset.scope", "asset:sign").String(),
	}
}
