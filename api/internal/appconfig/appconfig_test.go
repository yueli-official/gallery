package appconfig

import "testing"

func TestIdentityBaseFallsBackToJWKSOrigin(t *testing.T) {
	if got := identityBaseFromJWKS("https://identity-internal:8443/oauth2/jwks.json"); got != "https://identity-internal:8443" {
		t.Fatalf("identity base = %q", got)
	}
	if got := identityBaseFromJWKS("not a URL"); got != "http://localhost:8081" {
		t.Fatalf("invalid fallback = %q", got)
	}
}
