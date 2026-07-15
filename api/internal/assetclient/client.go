package assetclient

import (
	"context"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"

	"platform/products/gallery/api/internal/galleryerr"
)

type HTTP struct {
	baseURL string
	siteKey string
}

func NewHTTP(baseURL, siteKey string) *HTTP {
	return &HTTP{baseURL: strings.TrimRight(baseURL, "/"), siteKey: siteKey}
}

func (c *HTTP) RegisterSubmission(ctx context.Context, bearer, assetID, submissionID, title string) error {
	response, err := g.Client().SetHeader("Authorization", "Bearer "+bearer).ContentJson().Post(ctx, c.baseURL+"/api/v1/asset-references", g.Map{
		"assetId": assetID, "siteKey": c.siteKey, "refType": "gallery-submission-image",
		"refId": submissionID, "refLabel": title, "refUrl": "/submissions/" + submissionID,
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

func (c *HTTP) UnregisterSubmission(ctx context.Context, bearer, assetID, submissionID string) error {
	query := url.Values{}
	query.Set("assetId", assetID)
	query.Set("siteKey", c.siteKey)
	query.Set("refType", "gallery-submission-image")
	query.Set("refId", submissionID)
	response, err := g.Client().SetHeader("Authorization", "Bearer "+bearer).Delete(ctx, c.baseURL+"/api/v1/asset-references?"+query.Encode())
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
