package runtime

import (
	"context"
	"os"
	"testing"

	"github.com/gogf/gf/v2/os/gcfg"
)

func TestEnvironmentAdapterOverridesScalarAndNestedConfig(t *testing.T) {
	base, err := gcfg.NewAdapterContent(`
database:
  default:
    host: file-db
gallery:
  authorization:
    bootstrapAdministratorSubs: [file-admin]
`)
	if err != nil {
		t.Fatal(err)
	}
	config := gcfg.NewWithAdapter(&environmentAdapter{base: base})
	t.Setenv("GF_DATABASE_DEFAULT_HOST", "container-db")
	t.Setenv("GF_GALLERY_AUTHORIZATION_BOOTSTRAPADMINISTRATORSUBS", `["container-admin"]`)

	if host := config.MustGet(context.Background(), "database.default.host").String(); host != "container-db" {
		t.Fatalf("scalar host = %q", host)
	}
	subs := config.MustGet(
		context.Background(),
		"gallery.authorization.bootstrapAdministratorSubs",
	).Strings()
	if len(subs) != 1 || subs[0] != "container-admin" {
		t.Fatalf("nested administrator subjects = %#v", subs)
	}
}

func TestApplyDatabaseURLProjectsOneAuthoritativeConnection(t *testing.T) {
	t.Setenv(
		"GALLERY_DATABASE_URL",
		"postgres://gallery:p%40ss@database.internal:5440/gallery_main?sslmode=require",
	)
	for _, key := range []string{
		"GF_DATABASE_DEFAULT_TYPE", "GF_DATABASE_DEFAULT_HOST", "GF_DATABASE_DEFAULT_PORT",
		"GF_DATABASE_DEFAULT_USER", "GF_DATABASE_DEFAULT_PASS", "GF_DATABASE_DEFAULT_NAME",
		"GF_DATABASE_DEFAULT_EXTRA",
	} {
		t.Setenv(key, "")
	}
	if err := applyDatabaseURL(); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"GF_DATABASE_DEFAULT_TYPE":  "pgsql",
		"GF_DATABASE_DEFAULT_HOST":  "database.internal",
		"GF_DATABASE_DEFAULT_PORT":  "5440",
		"GF_DATABASE_DEFAULT_USER":  "gallery",
		"GF_DATABASE_DEFAULT_PASS":  "p@ss",
		"GF_DATABASE_DEFAULT_NAME":  "gallery_main",
		"GF_DATABASE_DEFAULT_EXTRA": "sslmode=require",
	}
	for key, expected := range want {
		if got := os.Getenv(key); got != expected {
			t.Errorf("%s = %q, want %q", key, got, expected)
		}
	}
}

func TestApplyDatabaseURLRejectsMissingDatabase(t *testing.T) {
	t.Setenv("GALLERY_DATABASE_URL", "postgres://gallery:secret@database.internal")
	if err := applyDatabaseURL(); err == nil {
		t.Fatal("applyDatabaseURL() accepted a URL without a database name")
	}
}
