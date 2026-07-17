package assetclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterSubmissionUsesServiceCredentialInsteadOfUserBearer(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(response http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if request.Form.Get("grant_type") != "client_credentials" || request.Form.Get("client_id") != "gallery-asset-svc" || request.Form.Get("client_secret") != "secret" || request.Form.Get("scope") != "asset:sign" {
			t.Fatalf("token form = %v", request.Form)
		}
		_ = json.NewEncoder(response).Encode(map[string]any{"access_token": "service-token", "expires_in": 600})
	})
	mux.HandleFunc("/api/v1/asset-references", func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer service-token" {
			t.Fatalf("authorization = %q", request.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(response).Encode(map[string]any{"code": "ok", "data": map[string]any{}})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewHTTP(Config{
		BaseURL: server.URL, TokenURL: server.URL + "/oauth2/token", ClientID: "gallery-asset-svc",
		ClientSecret: "secret", Scope: "asset:sign", SiteKey: "gallery-main", HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.RegisterSubmission(context.Background(), "guest-resource-token", "asset-1", "submission-1", "Title"); err != nil {
		t.Fatal(err)
	}
}

func TestPrepareSubmissionUsesSignedThumbnailAndReturnsImmutableFacts(t *testing.T) {
	tokenCalls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(response http.ResponseWriter, _ *http.Request) {
		tokenCalls++
		_ = json.NewEncoder(response).Encode(map[string]any{"access_token": "service-token", "expires_in": 600})
	})
	mux.HandleFunc("/api/v1/assets/asset-1/sign", func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("preset") != "thumbnail" || request.Header.Get("Authorization") != "Bearer service-token" {
			t.Fatalf("request = %s, auth = %q", request.URL.String(), request.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(response).Encode(map[string]any{"code": "ok", "data": map[string]any{
			"url": "http://asset.test/api/v1/assets/blob/signed", "contentHash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"mime": "image/png", "width": 1600, "height": 900,
		}})
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	client, err := NewHTTP(Config{BaseURL: server.URL, TokenURL: server.URL + "/oauth2/token", ClientID: "gallery-asset-svc", ClientSecret: "secret", SiteKey: "gallery-main", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	first, err := client.PrepareSubmission(context.Background(), "asset-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.PrepareSubmission(context.Background(), "asset-1"); err != nil {
		t.Fatal(err)
	}
	if first.PreviewURL != "http://asset.test/api/v1/assets/blob/signed" || first.Width != 1600 || first.Height != 900 || tokenCalls != 1 {
		t.Fatalf("facts = %#v, token calls = %d", first, tokenCalls)
	}
}

func TestPublishImageRegistersPublicRenditionWithServiceCredential(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(response http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(response).Encode(map[string]any{"access_token": "service-token", "expires_in": 600})
	})
	mux.HandleFunc("/api/v1/assets/asset-1/publications", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.Header.Get("Authorization") != "Bearer service-token" {
			t.Fatalf("request = %s, auth = %q", request.Method, request.Header.Get("Authorization"))
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["siteKey"] != "gallery-main" || body["refType"] != "gallery-public-image" || body["refId"] != "image-1" {
			t.Fatalf("body = %#v", body)
		}
		_ = json.NewEncoder(response).Encode(map[string]any{"code": "ok", "data": map[string]any{"published": true}})
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	client, err := NewHTTP(Config{BaseURL: server.URL, TokenURL: server.URL + "/oauth2/token", ClientID: "gallery-asset-svc", ClientSecret: "secret", SiteKey: "gallery-main", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.PublishImage(context.Background(), "asset-1", "image-1", "Approved"); err != nil {
		t.Fatal(err)
	}
}
