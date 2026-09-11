// Package assetreferences owns this consumer's authoritative business usage.
package assetreferences

import (
	"context"
	"database/sql"
	"github.com/yueli-official/asset/referencesync"
)

func Source(origins ...string) func(context.Context, *sql.Tx) ([]referencesync.Snapshot, error) {
	return referencesync.QuerySource([]referencesync.Query{
		{RefType: "gallery-submission-image", Kind: "asset", SQL: `SELECT s.id::text,s.title,'/submissions/'||s.id::text,s.asset_id::text FROM gallery_submissions s LEFT JOIN gallery_images i ON i.id=s.image_id WHERE s.outcome<>'withdrawn' AND (s.image_id IS NULL OR (i.deleted_at IS NULL AND i.publication_state<>'deleted'))`},
		{RefType: "gallery-image", Kind: "asset", SQL: `SELECT id::text,title,'/images/'||id::text,asset_id::text FROM gallery_images WHERE deleted_at IS NULL AND publication_state<>'deleted'`},
	}, origins...)
}
