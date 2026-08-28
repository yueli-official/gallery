package galleryauthz

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/yueli-official/foundation/go/authorization"
)

func TestCatalogV3MigrationMatchesCompiledDefinition(t *testing.T) {
	catalog := authorization.MustCompile(Definition())
	migration, err := os.ReadFile("../../manifest/sql/migrations/0011_authorization_comments_v3.up.sql")
	if err != nil {
		t.Fatalf("read catalog migration: %v", err)
	}
	text := string(migration)
	if !strings.Contains(text, "catalog_version = "+strconv.FormatUint(uint64(catalog.Version()), 10)) {
		t.Fatalf("catalog migration does not install version %d", catalog.Version())
	}
	if !strings.Contains(text, catalog.Digest()) {
		t.Fatalf("catalog migration does not install compiled digest %q", catalog.Digest())
	}
	for _, capability := range []authorization.CapabilityKey{CapabilityCommentRead, CapabilityCommentModerate, CapabilityCommentDelete} {
		if !strings.Contains(text, string(capability)) {
			t.Fatalf("catalog migration does not bind %q", capability)
		}
	}
}
