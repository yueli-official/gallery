package deploy

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	jose "github.com/go-jose/go-jose/v4"
)

func TestWaitInterfaceValidatesOIDCServiceIdentityAndAssetContract(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	set := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{
		Key: privateKey.Public(), KeyID: "binding-test", Algorithm: "RS256", Use: "sig",
	}}}
	var identityURL string
	identity := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(response).Encode(map[string]any{
				"issuer": identityURL, "authorization_endpoint": identityURL + "/oauth2/authorize",
				"token_endpoint": identityURL + "/oauth2/token", "jwks_uri": identityURL + "/oauth2/jwks.json",
			})
		case "/oauth2/jwks.json":
			_ = json.NewEncoder(response).Encode(set)
		case "/oauth2/token":
			if err := request.ParseForm(); err != nil {
				t.Errorf("ParseForm() error = %v", err)
			}
			if request.Form.Get("client_id") != "gallery-asset-svc" ||
				request.Form.Get("client_secret") != "test-secret" ||
				request.Form.Get("scope") != "asset:sign" {
				t.Errorf("unexpected token form %#v", request.Form)
			}
			_ = json.NewEncoder(response).Encode(map[string]any{
				"access_token": "binding-token", "token_type": "Bearer",
			})
		default:
			http.NotFound(response, request)
		}
	}))
	defer identity.Close()
	identityURL = identity.URL

	asset := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/readyz":
			_, _ = response.Write([]byte(`{"status":"ok"}`))
		case "/api.json":
			paths := map[string]any{}
			for _, path := range requiredAssetPaths {
				paths[path] = map[string]any{"get": map[string]any{}}
			}
			_ = json.NewEncoder(response).Encode(map[string]any{"paths": paths})
		case "/api/v1/assets/00000000-0000-7000-8000-000000000000/sign":
			if request.Header.Get("Authorization") != "Bearer binding-token" {
				t.Errorf("authorization = %q", request.Header.Get("Authorization"))
			}
			response.WriteHeader(http.StatusNotFound)
			_, _ = response.Write([]byte(`{"code":"asset.not_found"}`))
		default:
			http.NotFound(response, request)
		}
	}))
	defer asset.Close()

	config := BindingConfig{
		IdentityIssuer: identity.URL, AssetBaseURL: asset.URL,
		ClientID: "gallery-asset-svc", ClientSecret: "test-secret", Scope: "asset:sign",
		AllowInsecureIdentityHTTP: true,
		AllowInsecureAssetHTTP:    true,
	}
	if err := WaitForBindings(context.Background(), config, time.Millisecond); err != nil {
		t.Fatalf("WaitForBindings() error = %v", err)
	}
}

func TestBindingImplementationRejectsIssuerMismatch(t *testing.T) {
	config, cleanup := bindingConfigForFailureTest(t)
	defer cleanup()
	config.IdentityIssuer = "https://identity.example.com"
	config.IdentityDiscoveryURL = strings.Replace(
		config.IdentityDiscoveryURL,
		"/.well-known",
		"/wrong/.well-known",
		1,
	)
	if err := checkBindings(context.Background(), config); err == nil ||
		!strings.Contains(err.Error(), "issuer") {
		t.Fatalf("checkBindings() error = %v, want issuer mismatch", err)
	}
}

func TestBindingImplementationRejectsPlainExternalHTTP(t *testing.T) {
	config := BindingConfig{
		IdentityIssuer: "http://identity.example.test",
		AssetBaseURL:   "https://asset.example.test",
		ClientID:       "gallery", ClientSecret: "secret", Scope: "asset:sign",
	}
	if err := checkBindings(context.Background(), config); err == nil ||
		!strings.Contains(err.Error(), "plain HTTP") {
		t.Fatalf("checkBindings() error = %v, want plain HTTP rejection", err)
	}
}

func bindingConfigForFailureTest(t *testing.T) (BindingConfig, func()) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	set := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{
		Key: privateKey.Public(), KeyID: "binding-test", Algorithm: "RS256", Use: "sig",
	}}}
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(request.URL.Path, "/.well-known/openid-configuration"):
			_ = json.NewEncoder(response).Encode(map[string]string{
				"issuer": serverURL, "authorization_endpoint": serverURL + "/oauth2/authorize",
				"token_endpoint": serverURL + "/oauth2/token", "jwks_uri": serverURL + "/oauth2/jwks.json",
			})
		case request.URL.Path == "/oauth2/jwks.json":
			_ = json.NewEncoder(response).Encode(set)
		case request.URL.Path == "/oauth2/token":
			_ = json.NewEncoder(response).Encode(map[string]string{
				"access_token": "token", "token_type": "Bearer",
			})
		case request.URL.Path == "/readyz":
			_, _ = response.Write([]byte(`{"status":"ok"}`))
		case request.URL.Path == "/api.json":
			paths := map[string]any{}
			for _, path := range requiredAssetPaths {
				paths[path] = map[string]any{}
			}
			_ = json.NewEncoder(response).Encode(map[string]any{"paths": paths})
		default:
			response.WriteHeader(http.StatusNotFound)
			_, _ = response.Write([]byte(`{}`))
		}
	}))
	serverURL = server.URL
	parsed, _ := url.Parse(server.URL)
	return BindingConfig{
		IdentityIssuer:            server.URL,
		IdentityDiscoveryURL:      server.URL + "/.well-known/openid-configuration",
		IdentityJWKSURL:           server.URL + "/oauth2/jwks.json",
		AssetBaseURL:              parsed.Scheme + "://" + parsed.Host,
		TokenURL:                  server.URL + "/oauth2/token",
		ClientID:                  "gallery",
		ClientSecret:              "secret",
		Scope:                     "asset:sign",
		AllowInsecureIdentityHTTP: true,
		AllowInsecureAssetHTTP:    true,
	}, server.Close
}
