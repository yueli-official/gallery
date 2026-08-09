package dao

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/lib/pq"

	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/model"
)

func (p *PG) ClassificationTags(ctx context.Context, cursor model.ClassificationTagCursor, limit int) ([]model.ClassificationTag, bool, error) {
	where := ""
	args := make([]any, 0, 4)
	if cursor.ID != "" {
		where = `WHERE (LOWER(tag.current_name), tag.id) > (LOWER(?), ?::uuid)`
		args = append(args, cursor.Name, cursor.ID)
	}
	args = append(args, limit+1)
	query := `
SELECT tag.id::text AS id, tag.current_name AS name, tag.current_slug AS slug, tag.status,
       COALESCE(tag.replacement_id::text, '') AS replacement_id,
       ((SELECT COUNT(*) FROM gallery_submission_tag_assignments assignment WHERE assignment.tag_id = tag.id) +
        (SELECT COUNT(*) FROM gallery_image_tag_assignments assignment WHERE assignment.tag_id = tag.id))::bigint AS assignment_count,
       (SELECT COUNT(*) FROM gallery_tag_lookup_entries entry WHERE entry.target_tag_id = tag.id AND entry.kind = 'alias')::bigint AS alias_count
FROM gallery_tags tag
` + where + `
ORDER BY LOWER(tag.current_name), tag.id
LIMIT ?`
	var values []model.ClassificationTag
	if err := p.db.Ctx(ctx).Raw(query, args...).Scan(&values); err != nil {
		return nil, false, gerror.Wrap(err, "query gallery classification tags")
	}
	hasMore := len(values) > limit
	if hasMore {
		values = values[:limit]
	}
	return values, hasMore, nil
}

func (p *PG) ClassificationTagProposals(ctx context.Context, status string, page, size int) ([]model.ClassificationTagProposal, int, error) {
	count, err := p.db.GetValue(ctx, `SELECT COUNT(*) FROM gallery_tag_proposals WHERE status = ?`, status)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "count gallery tag proposals")
	}
	var values []model.ClassificationTagProposal
	if err := p.db.Ctx(ctx).Raw(`
SELECT id::text AS id, submission_id::text AS submission_id, input_value, lookup_key, status,
       COALESCE(resolved_tag_id::text, '') AS resolved_tag_id, created_at, reviewed_at
FROM gallery_tag_proposals
WHERE status = ?
ORDER BY created_at, id
LIMIT ? OFFSET ?`, status, size, (page-1)*size).Scan(&values); err != nil {
		return nil, 0, gerror.Wrap(err, "query gallery tag proposals")
	}
	return values, count.Int(), nil
}

func (p *PG) ReviewClassificationTagProposal(ctx context.Context, operator, id string, input model.ClassificationTagProposalReviewInput) (*model.ClassificationTagProposal, bool, error) {
	var value *model.ClassificationTagProposal
	catalogChanged := false
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		catalog, err := tx.GetOne(`
SELECT id::text AS id, revision
FROM gallery_classification_catalogs
WHERE catalog_key = 'gallery'
FOR UPDATE`)
		if err != nil {
			return gerror.Wrap(err, "lock gallery classification catalog for tag proposal")
		}
		if len(catalog) == 0 {
			return galleryerr.NotInitialized("classification_catalog")
		}
		proposal, err := tx.GetOne(`
SELECT id::text AS id, submission_id::text AS submission_id, input_value, lookup_key, status,
       COALESCE(resolved_tag_id::text, '') AS resolved_tag_id, created_at, reviewed_at
FROM gallery_tag_proposals
WHERE id = ?::uuid
FOR UPDATE`, id)
		if err != nil {
			return gerror.Wrap(err, "lock gallery tag proposal")
		}
		if len(proposal) == 0 {
			return nil
		}
		if proposal["status"].String() != "pending" {
			return galleryerr.InvalidState("tag_proposal", proposal["status"].String())
		}
		if input.Decision == "reject" {
			record, err := tx.GetOne(`
UPDATE gallery_tag_proposals
SET status = 'rejected', reviewed_by = ?, reviewed_at = NOW(), updated_at = NOW()
WHERE id = ?::uuid
RETURNING id::text AS id, submission_id::text AS submission_id, input_value, lookup_key, status,
          COALESCE(resolved_tag_id::text, '') AS resolved_tag_id, created_at, reviewed_at`, operator, id)
			if err != nil {
				return gerror.Wrap(err, "reject gallery tag proposal")
			}
			value, err = recordAs[model.ClassificationTagProposal](record)
			return err
		}

		lookupKey := proposal["lookup_key"].String()
		targetID := ""
		existing, err := tx.GetOne(`
SELECT target.id::text AS id, target.status
FROM gallery_tag_lookup_entries entry
JOIN gallery_tags target ON target.id = entry.target_tag_id
WHERE entry.catalog_id = ?::uuid AND entry.lookup_key = ?`, catalog["id"].String(), lookupKey)
		if err != nil {
			return gerror.Wrap(err, "resolve latest gallery tag proposal lookup")
		}
		if len(existing) != 0 {
			if existing["status"].String() != "active" {
				return galleryerr.InvalidState("tag", existing["status"].String())
			}
			targetID = existing["id"].String()
			if input.TargetTagID != "" && input.TargetTagID != targetID {
				return galleryerr.Conflict("tag_lookup")
			}
		} else if input.TargetTagID != "" {
			target, err := tx.GetOne(`
SELECT id::text AS id
FROM gallery_tags
WHERE id = ?::uuid AND catalog_id = ?::uuid AND status = 'active'`, input.TargetTagID, catalog["id"].String())
			if err != nil {
				return gerror.Wrap(err, "query gallery tag proposal target")
			}
			if len(target) == 0 {
				return galleryerr.InvalidState("tag", "not_active")
			}
			targetID = target["id"].String()
			if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_tag_lookup_entries (catalog_id, lookup_key, target_tag_id, kind, display_value)
VALUES (?::uuid, ?, ?::uuid, 'alias', ?)`, catalog["id"].String(), lookupKey, targetID, proposal["input_value"].String()); err != nil {
				return gerror.Wrap(err, "create gallery tag alias from proposal")
			}
			catalogChanged = true
		} else {
			tag, err := tx.GetOne(`
INSERT INTO gallery_tags (id, catalog_id, current_name, current_slug)
VALUES (?::uuid, ?::uuid, ?, ?)
RETURNING id::text AS id`, newIdentifier(), catalog["id"].String(), proposal["input_value"].String(), lookupKey)
			if err != nil {
				return gerror.Wrap(err, "create canonical gallery tag from proposal")
			}
			targetID = tag["id"].String()
			if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_tag_lookup_entries (catalog_id, lookup_key, target_tag_id, kind, display_value)
VALUES (?::uuid, ?, ?::uuid, 'canonical', ?)`, catalog["id"].String(), lookupKey, targetID, proposal["input_value"].String()); err != nil {
				return gerror.Wrap(err, "register canonical gallery tag lookup")
			}
			catalogChanged = true
		}

		if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_submission_tag_assignments (submission_id, tag_id)
VALUES (?::uuid, ?::uuid)
ON CONFLICT DO NOTHING`, proposal["submission_id"].String(), targetID); err != nil {
			return gerror.Wrap(err, "assign approved gallery submission tag")
		}
		if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_image_tag_assignments (image_id, tag_id)
SELECT image_id, ?::uuid
FROM gallery_submissions
WHERE id = ?::uuid AND image_id IS NOT NULL
ON CONFLICT DO NOTHING`, targetID, proposal["submission_id"].String()); err != nil {
			return gerror.Wrap(err, "assign approved gallery image tag")
		}
		record, err := tx.GetOne(`
UPDATE gallery_tag_proposals
SET status = 'approved', resolved_tag_id = ?::uuid, reviewed_by = ?, reviewed_at = NOW(), updated_at = NOW()
WHERE id = ?::uuid
RETURNING id::text AS id, submission_id::text AS submission_id, input_value, lookup_key, status,
          resolved_tag_id::text AS resolved_tag_id, created_at, reviewed_at`, targetID, operator, id)
		if err != nil {
			return gerror.Wrap(err, "approve gallery tag proposal")
		}
		value, err = recordAs[model.ClassificationTagProposal](record)
		if err != nil {
			return err
		}
		if !catalogChanged {
			return nil
		}
		revision, err := tx.GetOne(`
UPDATE gallery_classification_catalogs
SET revision = revision + 1, updated_at = NOW()
WHERE id = ?::uuid
RETURNING revision`, catalog["id"].String())
		if err != nil {
			return gerror.Wrap(err, "advance gallery tag catalog revision")
		}
		payload, err := json.Marshal(map[string]string{
			"operator":   operator,
			"proposalId": id,
			"tagId":      targetID,
		})
		if err != nil {
			return gerror.Wrap(err, "encode gallery tag proposal outbox payload")
		}
		if err := p.enqueueClassificationRefresh(
			ctx, tx, catalog["id"].String(), revision["revision"].Uint64(),
			"classification.tag_proposal.approved", payload,
		); err != nil {
			return gerror.Wrap(err, "enqueue gallery classification refresh")
		}
		return nil
	})
	if err != nil {
		var postgresError *pq.Error
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return nil, false, galleryerr.Conflict("tag_lookup")
		}
		return nil, false, err
	}
	return value, catalogChanged, nil
}
