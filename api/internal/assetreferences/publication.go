package assetreferences

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/yueli-official/asset/referencesync"
)

type PublicationPort interface {
	PublishImage(context.Context, string, string, string, string) error
	UnpublishImage(context.Context, string, string) error
}

// Serialize external publication with the authoritative image row. Replays are
// idempotent; a late delivery always observes the current takedown decision.
func SynchronizePublication(ctx context.Context, db *sql.DB, port PublicationPort, id string) error {
	if db == nil || port == nil {
		return fmt.Errorf("gallery publication database and client are required")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var asset, title, state, processing, review, safety, publicKey string
	var ready, deleted bool
	err = tx.QueryRowContext(ctx, `SELECT public_id,asset_id::text,title,publication_state,processing_state,review_state,safety_state,public_rendition_ready,deleted_at IS NOT NULL FROM gallery_images WHERE id=$1::uuid FOR UPDATE`, id).Scan(&publicKey, &asset, &title, &state, &processing, &review, &safety, &ready, &deleted)
	if err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, `SELECT asset_id::text FROM gallery_publication_retirements WHERE image_id=$1::uuid ORDER BY asset_id`, id)
	if err != nil {
		return err
	}
	var retired []string
	for rows.Next() {
		var old string
		if err = rows.Scan(&old); err != nil {
			rows.Close()
			return err
		}
		retired = append(retired, old)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, old := range retired {
		if old != asset {
			if err = port.UnpublishImage(ctx, old, id); err != nil {
				return err
			}
		}
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM gallery_publication_retirements WHERE image_id=$1::uuid`, id); err != nil {
		return err
	}
	switch {
	case deleted || state == "deleted" || state == "hidden":
		if err = port.UnpublishImage(ctx, asset, id); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE gallery_images SET public_rendition_ready=FALSE WHERE id=$1::uuid`, id)
	case state == "published" && !ready:
		if processing != "ready" || (review != "approved" && review != "not_required") || safety != "safe" {
			return fmt.Errorf("gallery image is not publishable")
		}
		if err = port.PublishImage(ctx, asset, id, title, publicKey); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE gallery_images SET public_rendition_ready=TRUE WHERE id=$1::uuid`, id)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

func PublicationHook(db *sql.DB, port PublicationPort) func(context.Context, referencesync.ObjectSet, referencesync.ObjectResult) error {
	return func(ctx context.Context, set referencesync.ObjectSet, _ referencesync.ObjectResult) error {
		if set.RefType != "gallery-image" {
			return nil
		}
		return SynchronizePublication(ctx, db, port, set.RefID)
	}
}
