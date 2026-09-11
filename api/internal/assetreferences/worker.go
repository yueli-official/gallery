package assetreferences

import (
	"context"
	"database/sql"
	"github.com/yueli-official/asset/referencesync"
	"time"
)

// RemovedImages is separate from business usage: publication is an Asset-owned
// public-rendition capability and must be explicitly revoked after deletion.
func RemovedImages(ctx context.Context, db *sql.DB, client *referencesync.Client, site string) error {
	rows, err := db.QueryContext(ctx, `SELECT id::text,asset_id::text FROM gallery_images WHERE publication_state='deleted' OR deleted_at IS NOT NULL`)
	if err != nil {
		return err
	}
	refs := []referencesync.Reference{}
	for rows.Next() {
		var ref referencesync.Reference
		if err = rows.Scan(&ref.RefID, &ref.AssetID); err != nil {
			rows.Close()
			return err
		}
		refs = append(refs, ref)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, ref := range refs {
		if err = client.Remove(ctx, site, "asset-rendition-publication", ref); err != nil {
			return err
		}
	}
	return nil
}

func Run(ctx context.Context, db *sql.DB, client *referencesync.Client, site string, report func(error), origins ...string) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		cycle, cancel := context.WithTimeout(ctx, 2*time.Minute)
		_, err := referencesync.Reconcile(cycle, db, "gallery:asset-references", Source(origins...), client)
		if err == nil {
			err = RemovedImages(cycle, db, client, site)
		}
		cancel()
		if err != nil {
			report(err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
