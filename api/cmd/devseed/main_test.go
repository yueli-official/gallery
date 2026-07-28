package main

import (
	"strings"
	"testing"
)

func TestEmbeddedSeedOwnsOnlyGalleryTables(t *testing.T) {
	for _, path := range []string{"sql/site.sql", "sql/content.sql"} {
		body, err := seedFiles.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := strings.ToLower(string(body))
		if strings.Contains(text, "insert into assets") {
			t.Fatalf("%s crosses the Asset database boundary", path)
		}
		if !strings.Contains(text, "gallery_") {
			t.Fatalf("%s contains no Gallery-owned table", path)
		}
	}
}
