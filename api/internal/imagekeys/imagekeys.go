// Package imagekeys owns Gallery's persistent public URL keys. Entity UUIDs
// and their foreign keys remain unchanged.
package imagekeys

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/yueli-official/foundation/go/identifier"
)

func Verify(ctx context.Context, db *sql.DB) error {
	var ready bool
	if err := db.QueryRowContext(ctx, `SELECT COALESCE((SELECT attnotnull FROM pg_attribute WHERE attrelid='gallery_images'::regclass AND attname='public_id' AND NOT attisdropped),FALSE)`).Scan(&ready); err != nil {
		return err
	}
	if !ready {
		return fmt.Errorf("image public keys are not ready; run schema migrations and imagekeys before starting Gallery")
	}
	return nil
}

// Backfill assigns keys and seals the schema in a single transaction. The table
// lock prevents old writers or concurrent importers from inserting unkeyed rows.
func Backfill(ctx context.Context, db *sql.DB) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `LOCK TABLE gallery_images IN ACCESS EXCLUSIVE MODE`); err != nil {
		return 0, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id::text FROM gallery_images WHERE public_id IS NULL ORDER BY id`)
	if err != nil {
		return 0, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return 0, err
	}
	for _, id := range ids {
		_, err = identifier.Allocate(ctx, identifier.CompactURLV1, func(ctx context.Context, key identifier.Key) (identifier.ClaimResult, error) {
			// Reserve through the exact public-key constraint without aborting the
			// transaction on collisions; the existing owning row keeps its UUID.
			result, err := tx.ExecContext(ctx, `UPDATE gallery_images SET public_id=$1 WHERE id=$2::uuid AND NOT EXISTS (SELECT 1 FROM gallery_images WHERE public_id=$1)`, key.String(), id)
			if err != nil {
				return 0, err
			}
			n, err := result.RowsAffected()
			if err != nil {
				return 0, err
			}
			if n == 0 {
				return identifier.Collision, nil
			}
			return identifier.Claimed, nil
		})
		if err != nil {
			return 0, fmt.Errorf("allocate image public key: %w", err)
		}
	}
	if _, err = tx.ExecContext(ctx, `ALTER TABLE gallery_images ALTER COLUMN public_id SET NOT NULL`); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return len(ids), nil
}
