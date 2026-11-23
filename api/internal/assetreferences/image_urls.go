package assetreferences

import (
	"context"
	"database/sql"
	"strings"

	"github.com/yueli-official/asset/referencesync"
)

// RefreshImageURLs projects the migrated page addresses through the normal
// reference transaction. Force sync alone would resend the old persisted URL.
func RefreshImageURLs(ctx context.Context, module *referencesync.Module, actor, origin string) (int, error) {
	rows, err := module.DB.QueryContext(ctx, `SELECT i.id::text
FROM gallery_images i JOIN asset_reference_intents r
ON r.instance_key=$1 AND r.ref_type='gallery-image' AND r.ref_id=i.id::text
WHERE COALESCE(r.payload->>'url','')<>$2||'/images/'||i.public_id ORDER BY i.id`, module.Instance, strings.TrimRight(origin, "/"))
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
	for n, id := range ids {
		err = module.Change(ctx, actor, []referencesync.OwnerKey{{RefType: "gallery-image", RefID: id}},
			func(sets []referencesync.ObjectSet) ([]referencesync.ObjectSet, error) { return sets, nil },
			func(ctx context.Context, verify, enqueue func(*sql.Tx) error) error {
				tx, err := module.DB.BeginTx(ctx, nil)
				if err != nil {
					return err
				}
				defer tx.Rollback()
				if err = verify(tx); err != nil {
					return err
				}
				if err = enqueue(tx); err != nil {
					return err
				}
				return tx.Commit()
			})
		if err != nil {
			return n, err
		}
	}
	return len(ids), nil
}
