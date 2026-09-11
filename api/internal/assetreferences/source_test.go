package assetreferences

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/yueli-official/asset/referencesync"
	"os"
	"testing"
)

// Real PostgreSQL query tests use transaction-local tables only; no production data is touched.
func TestCommittedUsageLifecycle(t *testing.T) {
	dsn := os.Getenv("REFERENCE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set REFERENCE_TEST_DATABASE_URL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	exec := func(query string) {
		t.Helper()
		if _, err := tx.ExecContext(ctx, query); err != nil {
			t.Fatal(err)
		}
	}
	exec(`CREATE TEMP TABLE gallery_images(id text,title text,asset_id text,publication_state text,deleted_at timestamptz) ON COMMIT DROP; CREATE TEMP TABLE gallery_submissions(id text,title text,asset_id text,image_id text,outcome text) ON COMMIT DROP;INSERT INTO gallery_images VALUES ('a','A','KeyA','published',NULL),('b','B','KeyA','hidden',NULL)`)
	read := func(want int) []referencesync.Reference {
		t.Helper()
		snapshots, err := Source()(ctx, tx)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range snapshots {
			if s.RefType == "gallery-image" {
				if len(s.References) != want {
					t.Fatalf("references=%d want=%d", len(s.References), want)
				}
				return s.References
			}
		}
		t.Fatal("snapshot missing")
		return nil
	}
	read(2)
	read(2) // Backfill and repeat keep two independent users of the same asset.
	exec(`UPDATE gallery_images SET publication_state='deleted',deleted_at=now() WHERE id='a'`)
	read(1)
	exec(`UPDATE gallery_images SET publication_state='published',deleted_at=NULL WHERE id='a'`)
	read(2)
	exec(`UPDATE gallery_images SET asset_id='KeyB' WHERE id='a'`)
	for _, ref := range read(2) {
		if ref.RefID == "a" && ref.AssetID != "KeyB" && ref.MediaKey != "KeyB" {
			t.Fatalf("replacement retained old asset: %+v", ref)
		}
	}

	check := func(kind string, want int) {
		t.Helper()
		snapshots, err := Source()(ctx, tx)
		if err != nil {
			t.Fatal(err)
		}
		for _, snapshot := range snapshots {
			if snapshot.RefType == kind {
				if len(snapshot.References) != want {
					t.Fatalf("%s count=%d want=%d", kind, len(snapshot.References), want)
				}
				return
			}
		}
		t.Fatalf("missing %s", kind)
	}

	exec(`INSERT INTO gallery_submissions VALUES ('pending','Pending','PendingAsset',NULL,'pending'),('withdrawn','Withdrawn','Unused',NULL,'withdrawn'),('linked','Linked','KeyA','b','accepted')`)
	check("gallery-submission-image", 2)
	exec(`UPDATE gallery_submissions SET outcome='withdrawn' WHERE id='pending'`)
	check("gallery-submission-image", 1)
	exec(`UPDATE gallery_images SET publication_state='deleted',deleted_at=now() WHERE id='b'`)
	check("gallery-submission-image", 0)
}
