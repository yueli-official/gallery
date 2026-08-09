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

func TestEmbeddedSeedUsesCompactSharedAccountKey(t *testing.T) {
	body, err := seedFiles.ReadFile("sql/content.sql")
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(body))
	publicUserKey := "testa123"
	for _, table := range []string{"gallery_cases", "gallery_collections", "gallery_submissions"} {
		start := strings.Index(text, "delete from "+table)
		if start < 0 {
			t.Fatalf("development seed does not clear %s", table)
		}
		end := strings.Index(text[start:], ";")
		if end < 0 {
			t.Fatalf("development seed has an unterminated %s cleanup", table)
		}
		statement := text[start : start+end]
		if !strings.Contains(statement, publicUserKey) {
			t.Fatalf("development seed does not reconcile the compact public identity in %s", table)
		}
	}
}
