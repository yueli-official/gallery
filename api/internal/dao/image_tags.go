package dao

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/yueli-official/foundation/go/classification"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/model"
)

func (p *PG) TagSuggestions(ctx context.Context, query string, limit int) ([]model.TagCandidate, error) {
	pattern := "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(query) + "%"
	var items []model.TagCandidate
	err := p.db.Ctx(ctx).Raw(`
SELECT tag.id::text AS id, tag.current_name AS name, tag.current_slug AS slug
FROM gallery_tags tag
JOIN gallery_classification_catalogs catalog ON catalog.id = tag.catalog_id AND catalog.catalog_key = 'gallery'
WHERE tag.status = 'active' AND EXISTS (
 SELECT 1 FROM gallery_tag_lookup_entries lookup
 WHERE lookup.catalog_id = tag.catalog_id AND lookup.target_tag_id = tag.id
   AND lookup.lookup_key ILIKE ? ESCAPE '\'
)
ORDER BY (tag.current_name = ?) DESC, LOWER(tag.current_name), tag.id
LIMIT ?`, pattern, query, limit).Scan(&items)
	if err != nil {
		return nil, gerror.Wrap(err, "suggest gallery canonical tags")
	}
	return items, nil
}

func (p *PG) ProposeImageTags(ctx context.Context, subject model.Subject, imageID string, lookups []classification.TagLookupRequest) ([]model.ClassificationTagProposal, error) {
	items := make([]model.ClassificationTagProposal, 0, len(lookups))
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		// A temporary identity or account shares one budget across all images.
		if _, err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtextextended(?, 0))`, "gallery:tag-contributor:"+subject.Kind+":"+subject.ID); err != nil {
			return err
		}
		image, err := tx.GetOne(`SELECT i.id, i.title FROM gallery_images i WHERE i.id = ?::uuid AND `+eligibleImage+` FOR UPDATE`, imageID)
		if err != nil {
			return gerror.Wrap(err, "lock image for tag contribution")
		}
		if len(image) == 0 {
			return galleryerr.NotFound("image", imageID)
		}
		for _, lookup := range lookups {
			// An inactive canonical word cannot be reintroduced as a fresh proposal.
			inactive, err := tx.GetValue(`SELECT COUNT(*) FROM gallery_tag_lookup_entries lookup
JOIN gallery_tags tag ON tag.id = lookup.target_tag_id
JOIN gallery_classification_catalogs catalog ON catalog.id = lookup.catalog_id
WHERE catalog.catalog_key = 'gallery' AND lookup.lookup_key = ? AND tag.status <> 'active'`, lookup.LookupKey)
			if err != nil {
				return err
			}
			if inactive.Int() > 0 {
				return galleryerr.Validation("tags", "inactive tag")
			}
			record, err := tx.GetOne(`SELECT id::text AS id, image_id::text AS image_id, proposer_id, proposer_kind,
input_value, lookup_key, status, COALESCE(resolved_tag_id::text, '') AS resolved_tag_id, created_at, reviewed_at
FROM gallery_tag_proposals WHERE image_id = ?::uuid AND proposer_kind = ? AND proposer_id = ? AND lookup_key = ?`, imageID, subject.Kind, subject.ID, lookup.LookupKey)
			if err != nil {
				return err
			}
			if len(record) == 0 {
				budget, err := tx.GetOne(`SELECT COUNT(*) FILTER (WHERE status = 'pending') AS pending,
COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '1 minute') AS minute,
COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '24 hours') AS recent
FROM gallery_tag_proposals WHERE proposer_kind = ? AND proposer_id = ? AND image_id IS NOT NULL`, subject.Kind, subject.ID)
				if err != nil {
					return err
				}
				if (subject.Kind == "guest" && budget["minute"].Int() >= 10) ||
					(subject.Kind == "user" && (budget["pending"].Int() >= 50 || budget["recent"].Int() >= 100)) {
					return galleryerr.RateLimited()
				}
				record, err = tx.GetOne(`INSERT INTO gallery_tag_proposals (id, image_id, proposer_kind, proposer_id, input_value, lookup_key)
VALUES (?::uuid, ?::uuid, ?, ?, ?, ?)
RETURNING id::text AS id, image_id::text AS image_id, proposer_kind, proposer_id, input_value, lookup_key, status,
COALESCE(resolved_tag_id::text, '') AS resolved_tag_id, created_at, reviewed_at`, newIdentifier(), imageID, subject.Kind, subject.ID, lookup.DisplayValue, lookup.LookupKey)
				if err != nil {
					return gerror.Wrap(err, "create image tag proposal")
				}
			}
			item, err := recordAs[model.ClassificationTagProposal](record)
			if err != nil {
				return err
			}
			item.ImageTitle = image["title"].String()
			items = append(items, *item)
		}
		return nil
	})
	return items, err
}
