package assetclient

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"platform/products/gallery/api/internal/galleryerr"
	"platform/products/gallery/api/internal/model"
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
	baseURL      string
	tokenURL     string
	clientID     string
	clientSecret string
	scope        string
	siteKey      string
	http         *http.Client
	tokenMu      sync.Mutex
	token        string
	tokenExpires time.Time
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
	return &HTTP{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"), tokenURL: strings.TrimSpace(cfg.TokenURL),
		clientID: strings.TrimSpace(cfg.ClientID), clientSecret: cfg.ClientSecret, scope: scope,
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
	var envelope struct {
		Code string `json:"code"`
		Data struct {
			URL         string `json:"url"`
			ContentHash string `json:"contentHash"`
			Mime        string `json:"mime"`
			Width       *int   `json:"width"`
			Height      *int   `json:"height"`
		} `json:"data"`
	}
	if json.NewDecoder(response.Body).Decode(&envelope) != nil || response.StatusCode < 200 || response.StatusCode >= 300 || envelope.Code != "ok" {
		return SubmissionAssetFacts{}, galleryerr.UpstreamFailed(envelope.Code)
	}
	_, hashErr := hex.DecodeString(envelope.Data.ContentHash)
	if envelope.Data.URL == "" || len(envelope.Data.ContentHash) != 64 || hashErr != nil || !strings.HasPrefix(envelope.Data.Mime, "image/") || envelope.Data.Width == nil || envelope.Data.Height == nil || *envelope.Data.Width <= 0 || *envelope.Data.Height <= 0 {
		return SubmissionAssetFacts{}, galleryerr.UpstreamFailed("asset.invalid_media_facts")
	}
	return SubmissionAssetFacts{
		PreviewURL: envelope.Data.URL, ContentHash: strings.ToLower(envelope.Data.ContentHash), Mime: envelope.Data.Mime,
		Width: *envelope.Data.Width, Height: *envelope.Data.Height,
	}, nil
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
	var envelope struct {
		Code string `json:"code"`
	}
	if json.NewDecoder(response.Body).Decode(&envelope) != nil || response.StatusCode < 200 || response.StatusCode >= 300 || envelope.Code != "ok" {
		return galleryerr.UpstreamFailed(envelope.Code)
	}
	return nil
}

func (c *HTTP) accessToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.token != "" && time.Until(c.tokenExpires) > 30*time.Second {
		return c.token, nil
	}
	form := url.Values{
		"grant_type": {"client_credentials"}, "client_id": {c.clientID},
		"client_secret": {c.clientSecret}, "scope": {c.scope},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := c.http.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	var token struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
		return "", err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 || strings.TrimSpace(token.AccessToken) == "" {
		return "", fmt.Errorf("client credentials grant failed with %s", response.Status)
	}
	ttl := time.Duration(token.ExpiresIn) * time.Second
	if ttl <= 0 {
		ttl = time.Minute
	}
	c.token = token.AccessToken
	c.tokenExpires = time.Now().Add(ttl)
	return c.token, nil
}
