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
