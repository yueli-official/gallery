package assetreferences

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
)

type publicationFixture struct {
	published, revoked int
	fail               bool
}

func (p *publicationFixture) PublishImage(context.Context, string, string, string) error {
	p.published++
	if p.fail {
		return fmt.Errorf("publication temporarily unavailable")
	}
	return nil
}
func (p *publicationFixture) UnpublishImage(context.Context, string, string) error {
	p.revoked++
	if p.fail {
		return fmt.Errorf("revocation temporarily unavailable")
	}
	return nil
}

// Real rows and rollback protect the externally observable publication boundary.
func TestPublicationDeliveryUsesCurrentStateAndRetries(t *testing.T) {
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
	_, err = db.Exec(`CREATE TEMP TABLE gallery_publication_retirements(image_id uuid,asset_id uuid);CREATE TEMP TABLE gallery_images(id uuid PRIMARY KEY,asset_id uuid,title text,publication_state text,processing_state text,review_state text,safety_state text,public_rendition_ready boolean,deleted_at timestamptz);
 INSERT INTO gallery_images VALUES('019c1000-0000-7000-8100-000000000001','019c1000-0000-7000-8100-000000000002','Image','published','ready','approved','safe',false,NULL)`)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	id := "019c1000-0000-7000-8100-000000000001"
	port := &publicationFixture{fail: true}
	if err = SynchronizePublication(ctx, db, port, id); err == nil {
		t.Fatal("failed publish acknowledged")
	}
	var ready bool
	if err = db.QueryRow(`SELECT public_rendition_ready FROM gallery_images`).Scan(&ready); err != nil || ready {
		t.Fatal("failed publish became eligible", err)
	}
	port.fail = false
	if err = SynchronizePublication(ctx, db, port, id); err != nil {
		t.Fatal(err)
	}
	if err = SynchronizePublication(ctx, db, port, id); err != nil {
		t.Fatal(err)
	}
	if port.published != 2 {
		t.Fatal("confirmed publication repeated")
	}
	if _, err = db.Exec(`UPDATE gallery_images SET publication_state='hidden'`); err != nil {
		t.Fatal(err)
	}
	port.fail = true
	if err = SynchronizePublication(ctx, db, port, id); err == nil {
		t.Fatal("failed revocation acknowledged")
	}
	port.fail = false
	if err = SynchronizePublication(ctx, db, port, id); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`UPDATE gallery_images SET publication_state='deleted',deleted_at=NOW()`); err != nil {
		t.Fatal(err)
	}
	if err = SynchronizePublication(ctx, db, port, id); err != nil {
		t.Fatal(err)
	}
	if port.published != 2 || port.revoked != 3 {
		t.Fatalf("late delivery ignored current takedown: %+v", port)
	}
}
