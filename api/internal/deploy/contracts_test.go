package deploy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDeploymentRequirementsStayConsumerOrientedAndLocked(t *testing.T) {
	root := repositoryRoot(t)
	var requirements struct {
		SchemaVersion int `json:"schemaVersion"`
		Requires      map[string]struct {
			Contract string `json:"contract"`
			Profile  string `json:"profile"`
		} `json:"requires"`
	}
	readJSON(t, filepath.Join(root, "deploy", "contracts", "requirements.json"), &requirements)
	if requirements.SchemaVersion != 1 {
		t.Fatalf("requirements schema version = %d", requirements.SchemaVersion)
	}
	if len(requirements.Requires) != 2 {
		t.Fatalf("requirements = %#v", requirements.Requires)
	}
	if requirements.Requires["auth.oidc"].Contract != "1.0" {
		t.Fatalf("auth requirement = %#v", requirements.Requires["auth.oidc"])
	}
	media := requirements.Requires["media.library"]
	if media.Contract != "1.0" || media.Profile != "gallery-submission" {
		t.Fatalf("media requirement = %#v", media)
	}
	for key := range requirements.Requires {
		if strings.HasPrefix(key, "identity.") || strings.HasPrefix(key, "asset.") {
			t.Fatalf("consumer requirement names a producer: %s", key)
		}
	}

	var lock struct {
		MinimumComposeVersion string `json:"minimumComposeVersion"`
		Bindings              map[string]struct {
			Provider           string `json:"provider"`
			ProviderCapability string `json:"providerCapability"`
		} `json:"bindings"`
		Dependencies map[string]struct {
			Revision string `json:"revision"`
		} `json:"dependencies"`
	}
	readJSON(t, filepath.Join(root, "deploy", "deployment.lock.json"), &lock)
	if lock.MinimumComposeVersion != "2.20.0" {
		t.Fatalf("minimum Compose version = %q", lock.MinimumComposeVersion)
	}
	if lock.Bindings["auth.oidc"].Provider != "identity" ||
		lock.Bindings["media.library"].Provider != "asset" {
		t.Fatalf("deployment bindings = %#v", lock.Bindings)
	}
	for provider, expected := range map[string]string{
		"identity": "9d367d3a10d8362b5ec7d8232a77050184980277",
		"asset":    "371647602911f8f268e741ec9cf88600018c1f6d",
	} {
		if lock.Dependencies[provider].Revision != expected {
			t.Fatalf("%s revision = %q", provider, lock.Dependencies[provider].Revision)
		}
	}
}

func TestComposeEntrypointsHaveDistinctTopologies(t *testing.T) {
	root := repositoryRoot(t)
	cases := map[string][]string{
		"compose.yaml":        {"includes/identity/compose.yaml", "includes/asset/compose.yaml", "gallery.standalone.yaml"},
		"compose.attach.yaml": {"compose.gallery.yaml"},
		"compose.hybrid.yaml": {"includes/asset/compose.yaml", "gallery.hybrid.yaml"},
	}
	for name, fragments := range cases {
		body, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range fragments {
			if !strings.Contains(string(body), fragment) {
				t.Errorf("%s does not include %s", name, fragment)
			}
		}
	}
	attach, err := os.ReadFile(filepath.Join(root, "compose.attach.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(attach), "includes/identity") ||
		strings.Contains(string(attach), "includes/asset") {
		t.Fatal("attach topology unexpectedly manages an external provider")
	}
}

func TestManagedAssetDoesNotIncludeRemovedMalwareScanner(t *testing.T) {
	root := repositoryRoot(t)
	for _, relativePath := range []string{
		filepath.Join("deploy", "deployment.lock.json"),
		filepath.Join("deploy", "includes", "asset", "compose.yaml"),
		filepath.Join("deploy", "includes", "asset", "config.yaml"),
	} {
		body, err := os.ReadFile(filepath.Join(root, relativePath))
		if err != nil {
			t.Fatal(err)
		}
		lower := strings.ToLower(string(body))
		for _, forbidden := range []string{"clamav", "clamd", "scanner"} {
			if strings.Contains(lower, forbidden) {
				t.Errorf("%s still contains removed %q dependency", relativePath, forbidden)
			}
		}
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve deployment test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func readJSON(t *testing.T, path string, target any) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}
