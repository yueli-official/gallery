package assetreferences

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/yueli-official/asset/referencesync"
	"os"
	"testing"
)

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
	_, err = tx.Exec(`CREATE TEMP TABLE gallery_images(id uuid PRIMARY KEY,title text,asset_id uuid,publication_state text,deleted_at timestamptz,updated_at timestamptz DEFAULT NOW());
 INSERT INTO gallery_images(id,title,asset_id,publication_state) VALUES('019c1000-0000-7000-8100-000000000001','Image','019c1000-0000-7000-8100-000000000002','published')`)
	if err != nil {
		t.Fatal(err)
	}
	source, err := ObjectSource()
	if err != nil {
		t.Fatal(err)
	}
	key := referencesync.OwnerKey{RefType: "gallery-image", RefID: "019c1000-0000-7000-8100-000000000001"}
	read := func(want int) {
		t.Helper()
		set, err := source.ReadReferenceObject(ctx, tx, key)
		if err != nil || len(set.Assets) != want {
			t.Fatalf("owner=%+v err=%v", set, err)
		}
	}
	read(1)
	if _, err = tx.Exec(`UPDATE gallery_images SET publication_state='hidden'`); err != nil {
		t.Fatal(err)
	}
	read(1)
	if _, err = tx.Exec(`UPDATE gallery_images SET publication_state='deleted',deleted_at=NOW()`); err != nil {
		t.Fatal(err)
	}
	read(0)
	if _, err = tx.Exec(`UPDATE gallery_images SET publication_state='published',deleted_at=NULL`); err != nil {
		t.Fatal(err)
	}
	read(1)
}
