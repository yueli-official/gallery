package main

import (
	"strings"
	"testing"
)

func TestEmbeddedSeedOwnsOnlyGalleryTables(t *testing.T) {
	body, err := seedFiles.ReadFile("sql/content.sql")
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(body))
	if strings.Contains(text, "insert into assets") {
		t.Fatal("development content crosses the Asset database boundary")
	}
	if !strings.Contains(text, "gallery_") {
		t.Fatal("development content contains no Gallery-owned table")
	}
}
