package assetreferences

import (
	"context"
	"database/sql"
	"github.com/yueli-official/asset/referencesync"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDeletedPublicationRetriesWithoutRevokingHiddenImage(t *testing.T) {
	dsn := os.Getenv("REFERENCE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set REFERENCE_TEST_DATABASE_URL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`CREATE TEMP TABLE gallery_images(id text,asset_id text,publication_state text,deleted_at timestamptz);INSERT INTO gallery_images VALUES ('deleted','shared','deleted',now()),('hidden','shared','hidden',NULL)`)
	if err != nil {
		t.Fatal(err)
	}
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		q := r.URL.Query()
		if r.Method != "DELETE" || q.Get("refId") != "deleted" || q.Get("assetId") != "shared" || q.Get("siteKey") != "gallery" || q.Get("refType") != "asset-rendition-publication" {
			t.Errorf("unexpected revoke: %s", r.URL)
		}
		if attempts == 1 {
			w.WriteHeader(503)
		} else {
			w.WriteHeader(204)
		}
	}))
	defer server.Close()
	client := &referencesync.Client{BaseURL: server.URL, Token: func(context.Context) (string, error) { return "fixture", nil }}
	if err = RemovedImages(context.Background(), db, client, "gallery"); err == nil {
		t.Fatal("failure ignored")
	}
	if err = RemovedImages(context.Background(), db, client, "gallery"); err != nil {
		t.Fatal(err)
	}
	if attempts != 2 {
		t.Fatalf("attempts=%d", attempts)
	}
}
