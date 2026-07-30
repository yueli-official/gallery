package deploy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	jose "github.com/go-jose/go-jose/v4"
)

const maxResponseBytes = int64(1 << 20)

var requiredAssetPaths = []string{
	"/api/v1/asset-references",
	"/api/v1/assets/finalize",
	"/api/v1/assets/upload-init",
	"/api/v1/assets/{id}/publications",
	"/api/v1/assets/{id}/sign",
}

type BindingConfig struct {
	IdentityIssuer            string
	IdentityDiscoveryURL      string
	IdentityJWKSURL           string
	AssetBaseURL              string
	TokenURL                  string
	ClientID                  string
	ClientSecret              string
	Scope                     string
	AllowInsecureIdentityHTTP bool
	AllowInsecureAssetHTTP    bool
}

func checkBindings(ctx context.Context, config BindingConfig) error {
	config = normalizeBindingConfig(config)
	if err := validateBindingConfig(config); err != nil {
		return err
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	if err := checkIdentity(ctx, client, config); err != nil {
		return err
	}
	token, err := acquireToken(ctx, client, config)
	if err != nil {
		return err
	}
	if err := checkAsset(ctx, client, config.AssetBaseURL, token); err != nil {
		return err
	}
	return nil
}

func WaitForBindings(
	ctx context.Context,
	config BindingConfig,
	interval time.Duration,
) error {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	var lastErr error
	for {
		if err := checkBindings(ctx, config); err == nil {
			return nil
		} else {
			lastErr = err
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("binding check did not converge: %w", errors.Join(lastErr, ctx.Err()))
		case <-timer.C:
		}
	}
}

func normalizeBindingConfig(config BindingConfig) BindingConfig {
	config.IdentityIssuer = strings.TrimRight(strings.TrimSpace(config.IdentityIssuer), "/")
	config.IdentityDiscoveryURL = strings.TrimSpace(config.IdentityDiscoveryURL)
	config.IdentityJWKSURL = strings.TrimSpace(config.IdentityJWKSURL)
	config.AssetBaseURL = strings.TrimRight(strings.TrimSpace(config.AssetBaseURL), "/")
	config.TokenURL = strings.TrimSpace(config.TokenURL)
	config.ClientID = strings.TrimSpace(config.ClientID)
	config.ClientSecret = strings.TrimSpace(config.ClientSecret)
	config.Scope = strings.TrimSpace(config.Scope)
	if config.IdentityDiscoveryURL == "" && config.IdentityIssuer != "" {
		config.IdentityDiscoveryURL = config.IdentityIssuer + "/.well-known/openid-configuration"
	}
	if config.IdentityJWKSURL == "" && config.IdentityIssuer != "" {
		config.IdentityJWKSURL = config.IdentityIssuer + "/oauth2/jwks.json"
	}
	if config.TokenURL == "" && config.IdentityIssuer != "" {
		config.TokenURL = config.IdentityIssuer + "/oauth2/token"
	}
	if config.Scope == "" {
		config.Scope = "asset:sign"
	}
	return config
}

func validateBindingConfig(config BindingConfig) error {
	required := map[string]string{
		"identity issuer":        config.IdentityIssuer,
		"identity discovery URL": config.IdentityDiscoveryURL,
		"identity JWKS URL":      config.IdentityJWKSURL,
		"asset base URL":         config.AssetBaseURL,
		"identity token URL":     config.TokenURL,
		"asset client ID":        config.ClientID,
		"asset client secret":    config.ClientSecret,
		"asset scope":            config.Scope,
	}
	for name, value := range required {
		if value == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	for name, value := range map[string]string{
		"identity issuer":        config.IdentityIssuer,
		"identity discovery URL": config.IdentityDiscoveryURL,
		"identity JWKS URL":      config.IdentityJWKSURL,
		"identity token URL":     config.TokenURL,
	} {
		if err := validateBindingURL(value, config.AllowInsecureIdentityHTTP); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	if err := validateBindingURL(config.AssetBaseURL, config.AllowInsecureAssetHTTP); err != nil {
		return fmt.Errorf("asset base URL: %w", err)
	}
	return nil
}

func validateBindingURL(value string, allowInsecureHTTP bool) error {
	parsed, err := url.Parse(value)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" {
		return errors.New("must be an absolute URL")
	}
	if parsed.User != nil || parsed.Fragment != "" {
		return errors.New("must not contain credentials or a fragment")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "https":
		return nil
	case "http":
		if allowInsecureHTTP || isLoopbackHost(parsed.Hostname()) {
			return nil
		}
		return errors.New("plain HTTP requires an explicit insecure-HTTP binding opt-in")
	default:
		return errors.New("scheme must be HTTP or HTTPS")
	}
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	address := net.ParseIP(host)
	return address != nil && address.IsLoopback()
}

func checkIdentity(ctx context.Context, client *http.Client, config BindingConfig) error {
	var discovery struct {
		Issuer                string `json:"issuer"`
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		TokenEndpoint         string `json:"token_endpoint"`
		JWKSURI               string `json:"jwks_uri"`
	}
	if err := getJSON(ctx, client, config.IdentityDiscoveryURL, "", &discovery); err != nil {
		return fmt.Errorf("identity discovery: %w", err)
	}
	if strings.TrimRight(discovery.Issuer, "/") != config.IdentityIssuer {
		return fmt.Errorf(
			"identity discovery issuer %q does not match expected %q",
			discovery.Issuer,
			config.IdentityIssuer,
		)
	}
	if discovery.AuthorizationEndpoint == "" || discovery.TokenEndpoint == "" || discovery.JWKSURI == "" {
		return errors.New("identity discovery is missing OIDC endpoints")
	}

	var set jose.JSONWebKeySet
	if err := getJSON(ctx, client, config.IdentityJWKSURL, "", &set); err != nil {
		return fmt.Errorf("identity JWKS: %w", err)
	}
	usable := 0
	for index := range set.Keys {
		key := &set.Keys[index]
		if key.Valid() && key.IsPublic() && strings.TrimSpace(key.KeyID) != "" &&
			(key.Use == "" || key.Use == "sig") {
			usable++
		}
	}
	if usable == 0 {
		return errors.New("identity JWKS has no usable public signing key")
	}
	return nil
}

func acquireToken(
	ctx context.Context,
	client *http.Client,
	config BindingConfig,
) (string, error) {
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {config.ClientID},
		"client_secret": {config.ClientSecret},
		"scope":         {config.Scope},
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		config.TokenURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("create identity token request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("identity token request: %w", err)
	}
	defer response.Body.Close()
	body, err := readBody(response.Body)
	if err != nil {
		return "", fmt.Errorf("identity token response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("identity token request returned %s", response.Status)
	}
	var payload struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("decode identity token response: %w", err)
	}
	if strings.TrimSpace(payload.AccessToken) == "" ||
		!strings.EqualFold(strings.TrimSpace(payload.TokenType), "Bearer") {
		return "", errors.New("identity token response is missing a bearer access token")
	}
	return payload.AccessToken, nil
}

func checkAsset(ctx context.Context, client *http.Client, baseURL, token string) error {
	if err := getJSON(ctx, client, baseURL+"/readyz", "", &map[string]any{}); err != nil {
		return fmt.Errorf("asset readiness: %w", err)
	}
	var document struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	if err := getJSON(ctx, client, baseURL+"/api.json", "", &document); err != nil {
		return fmt.Errorf("asset OpenAPI: %w", err)
	}
	for _, path := range requiredAssetPaths {
		if _, ok := document.Paths[path]; !ok {
			return fmt.Errorf("asset contract is missing required path %s", path)
		}
	}

	probeURL := baseURL +
		"/api/v1/assets/00000000-0000-7000-8000-000000000000/sign?preset=thumbnail"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)
	if err != nil {
		return fmt.Errorf("create asset authorization probe: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("asset authorization probe: %w", err)
	}
	defer response.Body.Close()
	_, readErr := readBody(response.Body)
	if readErr != nil {
		return fmt.Errorf("asset authorization probe: %w", readErr)
	}
	switch {
	case response.StatusCode == http.StatusUnauthorized,
		response.StatusCode == http.StatusForbidden:
		return fmt.Errorf("asset rejected the configured service identity with %s", response.Status)
	case response.StatusCode >= http.StatusInternalServerError:
		return fmt.Errorf("asset authorization probe returned %s", response.Status)
	}
	return nil
}

func getJSON(
	ctx context.Context,
	client *http.Client,
	endpoint string,
	bearer string,
	target any,
) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	if bearer != "" {
		request.Header.Set("Authorization", "Bearer "+bearer)
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := readBody(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("GET %s returned %s", endpoint, response.Status)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode GET %s: %w", endpoint, err)
	}
	return nil
}

func readBody(reader io.Reader) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(reader, maxResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxResponseBytes {
		return nil, fmt.Errorf("response exceeds %d bytes", maxResponseBytes)
	}
	return body, nil
}
