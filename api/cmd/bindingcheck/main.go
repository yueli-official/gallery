// Command bindingcheck validates the concrete Identity and Asset bindings
// before Gallery migrations and runtime processes are started.
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yueli-official/gallery/api/internal/deploy"
)

func main() {
	timeout := durationFromEnvironment("GALLERY_BINDING_TIMEOUT", 3*time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := deploy.WaitForBindings(ctx, deploy.BindingConfig{
		IdentityIssuer:       os.Getenv("GALLERY_IDENTITY_ISSUER"),
		IdentityDiscoveryURL: os.Getenv("GALLERY_IDENTITY_DISCOVERY_URL"),
		IdentityJWKSURL:      os.Getenv("GALLERY_IDENTITY_JWKS_URL"),
		AssetBaseURL:         os.Getenv("GALLERY_ASSET_BASE_URL"),
		TokenURL:             os.Getenv("GALLERY_IDENTITY_TOKEN_URL"),
		ClientID:             os.Getenv("GALLERY_ASSET_CLIENT_ID"),
		ClientSecret:         os.Getenv("GALLERY_ASSET_CLIENT_SECRET"),
		Scope:                os.Getenv("GALLERY_ASSET_SCOPE"),
		AllowInsecureIdentityHTTP: boolFromEnvironment(
			"GALLERY_BINDING_ALLOW_INSECURE_IDENTITY_HTTP",
		),
		AllowInsecureAssetHTTP: boolFromEnvironment(
			"GALLERY_BINDING_ALLOW_INSECURE_ASSET_HTTP",
		),
	}, 2*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gallery binding check:", err)
		os.Exit(1)
	}
	fmt.Println("Gallery Identity and Asset bindings are compatible")
}

func durationFromEnvironment(name string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		fmt.Fprintf(os.Stderr, "%s must be a positive duration\n", name)
		os.Exit(2)
	}
	return value
}

func boolFromEnvironment(name string) bool {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return false
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s must be a boolean\n", name)
		os.Exit(2)
	}
	return value
}
