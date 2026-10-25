package assetclient

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	foundationhttpclient "github.com/yueli-official/foundation/go/httpclient"
	"github.com/yueli-official/foundation/go/oauthclient"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/model"
)

type Config struct {
	BaseURL      string
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scope        string
	SiteKey      string
	HTTPClient   *http.Client
}

type HTTP struct {
	baseURL string
	siteKey string
	http    *http.Client
	tokens  *oauthclient.Credentials
}

type SubmissionAssetFacts = model.SubmissionAssetFacts

func NewHTTP(cfg Config) (*HTTP, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" || strings.TrimSpace(cfg.TokenURL) == "" || strings.TrimSpace(cfg.ClientID) == "" || strings.TrimSpace(cfg.ClientSecret) == "" || strings.TrimSpace(cfg.SiteKey) == "" {
		return nil, fmt.Errorf("asset reference client requires base url, token url, client credentials, and site key")
	}
	client := cfg.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	scope := strings.TrimSpace(cfg.Scope)
	if scope == "" {
		scope = "asset:sign"
	}
	tokens, err := oauthclient.NewCredentials(oauthclient.Config{Endpoint: strings.TrimSpace(cfg.TokenURL), ClientID: strings.TrimSpace(cfg.ClientID), ClientSecret: cfg.ClientSecret, Scopes: strings.Fields(scope), HTTPClient: client})
	if err != nil {
		return nil, err
	}
	return &HTTP{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"), tokens: tokens,
		siteKey: strings.TrimSpace(cfg.SiteKey), http: client,
	}, nil
}

func (c *HTTP) RegisterSubmission(ctx context.Context, _ string, assetID, submissionID, title string) error {
	return c.referenceRequest(ctx, http.MethodPost, c.baseURL+"/api/v1/asset-references", map[string]any{
		"assetId": assetID, "siteKey": c.siteKey, "refType": "gallery-submission-image",
		"refId": submissionID, "refLabel": title, "refUrl": "/submissions/" + submissionID,
	})
}

func (c *HTTP) UnregisterSubmission(ctx context.Context, _ string, assetID, submissionID string) error {
	query := url.Values{"assetId": {assetID}, "siteKey": {c.siteKey}, "refType": {"gallery-submission-image"}, "refId": {submissionID}}
	return c.referenceRequest(ctx, http.MethodDelete, c.baseURL+"/api/v1/asset-references?"+query.Encode(), nil)
}

func (c *HTTP) PublishImage(ctx context.Context, assetID, imageID, title string) error {
	return c.referenceRequest(ctx, http.MethodPost, c.baseURL+"/api/v1/assets/"+url.PathEscape(strings.TrimSpace(assetID))+"/publications", map[string]any{
		"siteKey": c.siteKey, "refId": imageID, "refLabel": title, "refUrl": "/images/" + strings.TrimSpace(imageID),
	})
}

func (c *HTTP) PrepareSubmission(ctx context.Context, assetID string) (SubmissionAssetFacts, error) {
	token, err := c.accessToken(ctx)
	if err != nil {
		return SubmissionAssetFacts{}, galleryerr.UpstreamFailed("asset.token_unavailable")
	}
	endpoint := c.baseURL + "/api/v1/assets/" + url.PathEscape(strings.TrimSpace(assetID)) + "/sign?preset=thumbnail"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return SubmissionAssetFacts{}, galleryerr.UpstreamFailed("asset.unreachable")
	}
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := c.http.Do(request)
	if err != nil {
		return SubmissionAssetFacts{}, galleryerr.UpstreamFailed("asset.unreachable")
	}
	defer response.Body.Close()
	facts, decodeErr := foundationhttpclient.DecodeJSON[struct {
		URL         string `json:"url"`
		ContentHash string `json:"contentHash"`
		Mime        string `json:"mime"`
		Width       *int   `json:"width"`
		Height      *int   `json:"height"`
	}](response, foundationhttpclient.Limits{})
	if decodeErr != nil {
		return SubmissionAssetFacts{}, galleryerr.UpstreamFailed(remoteCode(decodeErr))
	}
	_, hashErr := hex.DecodeString(facts.ContentHash)
	previewURL, previewURLErr := url.Parse(strings.TrimSpace(facts.URL))
	if previewURLErr != nil || previewURL.Host == "" ||
		(previewURL.Scheme != "http" && previewURL.Scheme != "https") ||
		previewURL.User != nil || len(facts.ContentHash) != 64 || hashErr != nil ||
		!strings.HasPrefix(facts.Mime, "image/") || facts.Width == nil ||
		facts.Height == nil || *facts.Width <= 0 || *facts.Height <= 0 {
		return SubmissionAssetFacts{}, galleryerr.UpstreamFailed("asset.invalid_media_facts")
	}
	return SubmissionAssetFacts{
		PreviewURL: previewURL.String(), ContentHash: strings.ToLower(facts.ContentHash), Mime: facts.Mime,
		Width: *facts.Width, Height: *facts.Height,
	}, nil
}

func (c *HTTP) SignOriginal(ctx context.Context, assetID string) (*model.ImageDownload, error) {
	token, err := c.accessToken(ctx)
	if err != nil {
		return nil, galleryerr.UpstreamFailed("asset.token_unavailable")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/assets/"+url.PathEscape(assetID)+"/sign", nil)
	if err != nil {
		return nil, galleryerr.UpstreamFailed("asset.unreachable")
	}
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := c.http.Do(request)
	if err != nil {
		return nil, galleryerr.UpstreamFailed("asset.unreachable")
	}
	defer response.Body.Close()
	value, err := foundationhttpclient.DecodeJSON[model.ImageDownload](response, foundationhttpclient.Limits{})
	if err != nil {
		return nil, galleryerr.UpstreamFailed(remoteCode(err))
	}
	parsed, err := url.Parse(value.URL)
	if err != nil || parsed.Host == "" || parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, galleryerr.UpstreamFailed("asset.invalid_media_facts")
	}
	return &value, nil
}

func (c *HTTP) referenceRequest(ctx context.Context, method, endpoint string, body map[string]any) error {
	token, err := c.accessToken(ctx)
	if err != nil {
		return galleryerr.UpstreamFailed("asset.token_unavailable")
	}
	var payload *bytes.Reader
	if body == nil {
		payload = bytes.NewReader(nil)
	} else {
		raw, _ := json.Marshal(body)
		payload = bytes.NewReader(raw)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, payload)
	if err != nil {
		return galleryerr.UpstreamFailed("asset.unreachable")
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.http.Do(request)
	if err != nil {
		return galleryerr.UpstreamFailed("asset.unreachable")
	}
	defer response.Body.Close()
	if _, decodeErr := foundationhttpclient.DecodeJSON[any](response, foundationhttpclient.Limits{}); decodeErr != nil {
		return galleryerr.UpstreamFailed(remoteCode(decodeErr))
	}
	return nil
}

func remoteCode(err error) string {
	var remote *foundationhttpclient.RemoteError
	if errors.As(err, &remote) {
		return remote.Problem.Code
	}
	return "foundation.response.invalid"
}

func (c *HTTP) accessToken(ctx context.Context) (string, error) {
	return c.tokens.Token(ctx)
}
