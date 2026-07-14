package assetclient

import (
	"context"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"

	"platform/products/gallery/api/internal/galleryerr"
)

type ReferenceInput struct {
	AssetID string
	RefID   string
	Label   string
	URL     string
}

type Client interface {
	RegisterReference(context.Context, string, ReferenceInput) error
	UnregisterReference(context.Context, string, ReferenceInput) error
}

type HTTP struct {
	baseURL string
	siteKey string
}

func NewHTTP(baseURL, siteKey string) *HTTP {
	return &HTTP{baseURL: strings.TrimRight(baseURL, "/"), siteKey: siteKey}
}

func (c *HTTP) RegisterReference(ctx context.Context, bearer string, input ReferenceInput) error {
	client := g.Client().SetHeader("Authorization", "Bearer "+bearer).ContentJson()
	response, err := client.Post(ctx, c.baseURL+"/api/v1/asset-references", g.Map{
		"assetId": input.AssetID, "siteKey": c.siteKey, "refType": "gallery-artwork-image",
		"refId": input.RefID, "refLabel": input.Label, "refUrl": input.URL,
	})
	if err != nil {
		return galleryerr.UpstreamFailed("asset.unreachable")
	}
	defer response.Close()
	code := gjson.New(response.ReadAllString()).Get("code").String()
	if code != "ok" {
		return galleryerr.UpstreamFailed(code)
	}
	return nil
}

func (c *HTTP) UnregisterReference(ctx context.Context, bearer string, input ReferenceInput) error {
	query := url.Values{}
	query.Set("assetId", input.AssetID)
	query.Set("siteKey", c.siteKey)
	query.Set("refType", "gallery-artwork-image")
	query.Set("refId", input.RefID)
	client := g.Client().SetHeader("Authorization", "Bearer "+bearer)
	response, err := client.Delete(ctx, c.baseURL+"/api/v1/asset-references?"+query.Encode())
	if err != nil {
		return galleryerr.UpstreamFailed("asset.unreachable")
	}
	defer response.Close()
	code := gjson.New(response.ReadAllString()).Get("code").String()
	if code != "ok" {
		return galleryerr.UpstreamFailed(code)
	}
	return nil
}
