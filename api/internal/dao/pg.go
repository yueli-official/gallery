package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/lib/pq"

	"platform/gokit/facet"
	"platform/products/gallery/api/internal/collection"
	"platform/products/gallery/api/internal/galleryerr"
	"platform/products/gallery/api/internal/model"
)

type PG struct{ db gdb.DB }

func NewPG(db gdb.DB) *PG { return &PG{db: db} }

const eligibleImage = `
i.processing_state = 'ready'
AND i.review_state IN ('not_required', 'approved')
AND i.publication_state = 'published'
AND i.safety_state = 'safe'
AND i.public_rendition_ready`

const imageCardSelect = `
SELECT i.id, i.asset_id, i.title, i.alt_text, i.width, i.height, i.dominant_color, i.published_at,
       COALESCE(topic.name, '') AS topic, COALESCE(topic.slug, '') AS topic_slug,
       COALESCE(metric.view_count, 0)::bigint AS view_count,
       COALESCE(metric.favorite_count, 0)::bigint AS favorite_count
FROM gallery_images i
LEFT JOIN LATERAL (
    SELECT v.name, v.slug
    FROM gallery_image_facet_assignments a
    JOIN gallery_facets f ON f.id = a.facet_id AND f.slug = 'topic'
    JOIN gallery_facet_values v ON v.id = a.value_id
    WHERE a.image_id = i.id
    LIMIT 1
) topic ON TRUE
LEFT JOIN LATERAL (
    SELECT COALESCE(SUM(m.qualified_views), 0) AS view_count,
           COALESCE(SUM(m.favorites), 0) AS favorite_count
    FROM gallery_image_metrics_daily m
    WHERE m.image_id = i.id
) metric ON TRUE`

func (p *PG) SiteSettings(ctx context.Context) (*model.SiteSettings, error) {
	var value *model.SiteSettings
	if err := p.db.Model("gallery_site_settings").Ctx(ctx).Order("site_key ASC").Limit(1).Scan(&value); err != nil {
		return nil, gerror.Wrap(err, "query gallery site settings")
	}
	return value, nil
}

func (p *PG) Facets(ctx context.Context) ([]facet.Facet, error) {
	const query = `
SELECT id::text AS id, slug, name, description, selection_mode,
       required_on_submit AS required_on_publish, filterable, status, sort_order
FROM gallery_facets
ORDER BY sort_order, name`
	var values []facet.Facet
	if err := p.db.Ctx(ctx).Raw(query).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query gallery facets")
	}
	return values, nil
}

func (p *PG) FacetValues(ctx context.Context) ([]facet.Value, error) {
	const query = `
SELECT v.id::text AS id, v.facet_id::text AS facet_id, COALESCE(v.parent_id::text, '') AS parent_id,
       v.slug, v.name, v.description, v.status, v.sort_order,
       COUNT(DISTINCT i.id)::int AS object_count
FROM gallery_facet_values v
LEFT JOIN gallery_image_facet_assignments a ON a.facet_id = v.facet_id AND a.value_id = v.id
LEFT JOIN gallery_images i ON i.id = a.image_id AND ` + eligibleImage + `
GROUP BY v.id, v.facet_id, v.parent_id, v.slug, v.name, v.description, v.status, v.sort_order
ORDER BY v.facet_id, v.sort_order, v.name`
	var values []facet.Value
	if err := p.db.Ctx(ctx).Raw(query).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query gallery facet values")
	}
	return values, nil
}

func (p *PG) RandomCandidates(ctx context.Context, limit int) ([]model.ImageCard, error) {
	query := imageCardSelect + ` WHERE ` + eligibleImage + ` ORDER BY i.published_at DESC, i.id DESC LIMIT ?`
	return p.cards(ctx, query, limit)
}

func (p *PG) ListImages(ctx context.Context, input model.ImageQuery) ([]model.ImageCard, int, error) {
	where := []string{eligibleImage}
	args := []any{}
	if input.Search != "" {
		where = append(where, `(i.title ILIKE ? OR i.description ILIKE ?)`)
		term := "%" + input.Search + "%"
		args = append(args, term, term)
	}
	if input.Tag != "" {
		where = append(where, `EXISTS (
            SELECT 1 FROM gallery_image_tags it
            JOIN gallery_tags t ON t.id = it.tag_id
            WHERE it.image_id = i.id AND t.slug = ?
        )`)
		args = append(args, input.Tag)
	}
	if len(input.FacetIDs) > 0 {
		where = append(where, `EXISTS (
            SELECT 1
            FROM gallery_image_facet_assignments filter_assignment
            WHERE filter_assignment.image_id = i.id
              AND filter_assignment.value_id = ANY(?::uuid[])
            GROUP BY filter_assignment.image_id
            HAVING COUNT(DISTINCT filter_assignment.facet_id) = (
                SELECT COUNT(DISTINCT selected_value.facet_id)
                FROM gallery_facet_values selected_value
                WHERE selected_value.id = ANY(?::uuid[])
            )
        )`)
		args = append(args, pq.Array(input.FacetIDs), pq.Array(input.FacetIDs))
	}
	clause := " WHERE " + strings.Join(where, " AND ")
	countQuery := `SELECT COUNT(*) FROM gallery_images i` + clause
	count, err := p.db.GetValue(ctx, countQuery, args...)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "count public gallery images")
	}
	orders := map[string]string{
		"newest": "i.published_at DESC, i.id DESC", "oldest": "i.published_at ASC, i.id ASC",
		"title_asc": "LOWER(i.title) ASC, i.id ASC", "title_desc": "LOWER(i.title) DESC, i.id DESC",
	}
	query := imageCardSelect + clause + " ORDER BY " + orders[input.Sort] + " LIMIT ? OFFSET ?"
	queryArgs := append(append([]any{}, args...), input.PageSize, (input.Page-1)*input.PageSize)
	values, err := p.cards(ctx, query, queryArgs...)
	return values, count.Int(), err
}

func (p *PG) Image(ctx context.Context, id, userID string) (*model.ImageDetail, error) {
	query := `
SELECT card.*, outer_image.description, COALESCE(outer_image.source_url, '') AS source_url,
       outer_image.focus_x, outer_image.focus_y,
       (? <> '' AND EXISTS (
           SELECT 1 FROM gallery_collections c
           JOIN gallery_collection_members member ON member.collection_id = c.id
           WHERE c.kind = 'gallery.favorites' AND c.owner_id = ? AND member.image_id = outer_image.id
       )) AS favorited
FROM gallery_images outer_image
JOIN LATERAL (` + imageCardSelect + ` WHERE i.id = outer_image.id) card ON TRUE
WHERE outer_image.id = ? AND ` + strings.ReplaceAll(eligibleImage, "i.", "outer_image.")
	var value *model.ImageDetail
	if err := p.db.Ctx(ctx).Raw(query, userID, userID, id).Scan(&value); err != nil {
		return nil, gerror.Wrap(err, "query public gallery image")
	}
	if value == nil {
		return nil, nil
	}
	var tags []struct {
		Name string `orm:"name"`
	}
	if err := p.db.Ctx(ctx).Raw(`
SELECT t.name FROM gallery_tags t
JOIN gallery_image_tags it ON it.tag_id = t.id
WHERE it.image_id = ? ORDER BY t.name`, id).Scan(&tags); err != nil {
		return nil, gerror.Wrap(err, "query gallery image tags")
	}
	value.Tags = make([]string, 0, len(tags))
	for _, tag := range tags {
		value.Tags = append(value.Tags, tag.Name)
	}
	if err := p.db.Model("gallery_image_facet_assignments").Ctx(ctx).
		Fields("facet_id::text AS facet_id", "value_id::text AS value_id").Where("image_id", id).Scan(&value.Facets); err != nil {
		return nil, gerror.Wrap(err, "query gallery image facets")
	}
	return value, nil
}

func (p *PG) HasTombstone(ctx context.Context, id string) (bool, error) {
	count, err := p.db.Model("gallery_image_tombstones").Ctx(ctx).Where("image_id", id).Count()
	return count > 0, err
}

func (p *PG) PublicCollections(ctx context.Context) ([]model.Collection, error) {
	const query = `
SELECT c.id::text AS id, c.kind, c.resource_kind, c.owner_kind, c.visibility, c.name, c.description,
       c.version, c.created_at, c.updated_at, e.slug, COALESCE(e.cover_image_id::text, '') AS cover_image_id,
       COUNT(i.id)::int AS item_count
FROM gallery_collections c
JOIN gallery_collection_editorial e ON e.collection_id = c.id
LEFT JOIN gallery_collection_members member ON member.collection_id = c.id
LEFT JOIN gallery_images i ON i.id = member.image_id AND ` + eligibleImage + `
WHERE c.kind = 'gallery.editorial' AND c.visibility = 'public'
GROUP BY c.id, e.slug, e.cover_image_id
ORDER BY c.updated_at DESC, c.id DESC`
	var values []model.Collection
	if err := p.db.Ctx(ctx).Raw(query).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query public gallery collections")
	}
	return values, nil
}

func (p *PG) EditorialCollections(ctx context.Context) ([]model.Collection, error) {
	const query = `
SELECT c.id::text AS id, c.kind, c.resource_kind, c.owner_kind, c.visibility, c.name, c.description,
       c.version, c.created_at, c.updated_at, e.slug, COALESCE(e.cover_image_id::text, '') AS cover_image_id,
       COUNT(member.image_id)::int AS item_count
FROM gallery_collections c
JOIN gallery_collection_editorial e ON e.collection_id = c.id
LEFT JOIN gallery_collection_members member ON member.collection_id = c.id
WHERE c.kind = 'gallery.editorial'
GROUP BY c.id, e.slug, e.cover_image_id
ORDER BY c.updated_at DESC, c.id DESC`
	var values []model.Collection
	if err := p.db.Ctx(ctx).Raw(query).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query editorial gallery collections")
	}
	return values, nil
}

func (p *PG) PublicCollection(ctx context.Context, slug string, page, size int) (*model.CollectionDetail, error) {
	const query = `
SELECT c.id::text AS id, c.kind, c.resource_kind, c.owner_kind, c.visibility, c.name, c.description,
       c.version, c.created_at, c.updated_at, e.slug, COALESCE(e.cover_image_id::text, '') AS cover_image_id,
       (SELECT COUNT(*) FROM gallery_collection_members count_member WHERE count_member.collection_id = c.id)::int AS item_count
FROM gallery_collections c
JOIN gallery_collection_editorial e ON e.collection_id = c.id
WHERE e.slug = ? AND c.visibility = 'public'`
	var value *model.Collection
	if err := p.db.Ctx(ctx).Raw(query, slug).Scan(&value); err != nil {
		return nil, gerror.Wrap(err, "query public gallery collection")
	}
	if value == nil {
		return nil, nil
	}
	images, err := p.collectionImages(ctx, value.ID, page, size)
	if err != nil {
		return nil, err
	}
	return &model.CollectionDetail{Collection: *value, Images: images}, nil
}

func (p *PG) CreateEditorialCollection(ctx context.Context, operator string, input model.EditorialCollectionInput) (*model.Collection, error) {
	var value *model.Collection
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		record, err := tx.GetOne(`
INSERT INTO gallery_collections (kind, resource_kind, owner_kind, owner_id, visibility, name, description)
VALUES ('gallery.editorial', 'gallery.image', 'site', 'gallery', ?, ?, ?)
RETURNING id::text AS id, kind, resource_kind, owner_kind, owner_id, visibility, name, description, version, created_at, updated_at`, input.Visibility, input.Name, input.Description)
		if err != nil {
			return gerror.Wrap(err, "create editorial collection")
		}
		value, err = recordAs[model.Collection](record)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO gallery_collection_editorial (collection_id, slug) VALUES (?::uuid, ?)`, value.ID, input.Slug); err != nil {
			var postgresError *pq.Error
			if errors.As(err, &postgresError) && postgresError.Code == "23505" {
				return galleryerr.Conflict("collection_slug")
			}
			return gerror.Wrap(err, "create editorial collection extension")
		}
		value.Slug = input.Slug
		_ = operator
		return nil
	})
	return value, err
}

func (p *PG) Ranking(ctx context.Context, kind, window string, limit int) (*model.Ranking, error) {
	cutoff := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Now().UTC()
	switch window {
	case "24h":
		cutoff = now.Add(-24 * time.Hour)
	case "7d":
		cutoff = now.AddDate(0, 0, -7)
	case "30d":
		cutoff = now.AddDate(0, 0, -30)
	}
	order := map[string]string{
		"trending":       "(window_metric.view_count + window_metric.favorite_count * 4) DESC",
		"most_viewed":    "window_metric.view_count DESC",
		"most_favorited": "window_metric.favorite_count DESC",
	}[kind]
	query := imageCardSelect + `
LEFT JOIN LATERAL (
    SELECT COALESCE(SUM(m.qualified_views), 0)::bigint AS view_count,
           COALESCE(SUM(m.favorites), 0)::bigint AS favorite_count
    FROM gallery_image_metrics_daily m
    WHERE m.image_id = i.id AND m.metric_date >= ?::date
) window_metric ON TRUE
WHERE ` + eligibleImage + `
ORDER BY ` + order + `, i.published_at DESC, i.id DESC
LIMIT ?`
	values, err := p.cards(ctx, query, cutoff, limit)
	if err != nil {
		return nil, err
	}
	return &model.Ranking{Kind: kind, Window: window, Images: values}, nil
}

func (p *PG) CreateSubmission(ctx context.Context, subject model.Subject, input model.SubmissionInput, review string) (*model.Submission, error) {
	const query = `
INSERT INTO gallery_submissions (
    subject_kind, subject_id, asset_id, title, description, source_url, alt_text, topic_value_id, review_state
)
SELECT ?, ?, ?::uuid, ?, ?, NULLIF(?, ''), ?, v.id, ?
FROM gallery_facet_values v
JOIN gallery_facets f ON f.id = v.facet_id
WHERE v.id = ?::uuid AND f.slug = 'topic' AND v.status = 'active'
RETURNING id::text AS id, subject_kind, subject_id, asset_id::text AS asset_id,
          COALESCE(image_id::text, '') AS image_id, title, description, COALESCE(source_url, '') AS source_url,
          alt_text, topic_value_id::text AS topic_value_id, processing_state, review_state, safety_state,
          outcome, failure_code, review_note, created_at, updated_at`
	var value *model.Submission
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		record, err := tx.Ctx(ctx).GetOne(query, subject.Kind, subject.ID, input.AssetID, input.Title, input.Description,
			input.SourceURL, input.AltText, review, input.TopicID)
		if err != nil {
			return err
		}
		if len(record) == 0 {
			return galleryerr.Validation("topicId", "topicId must reference an active topic")
		}
		value, err = recordAs[model.Submission](record)
		if err != nil {
			return err
		}
		for _, assignment := range input.Assignments {
			if _, err := tx.Exec(`INSERT INTO gallery_submission_facet_assignments (submission_id, facet_id, value_id)
VALUES (?::uuid, ?::uuid, ?::uuid)`, value.ID, assignment.FacetID, assignment.ValueID); err != nil {
				return gerror.Wrap(err, "assign gallery submission facet")
			}
		}
		for _, tag := range input.NormalizedTags {
			tagRecord, err := tx.GetOne(`INSERT INTO gallery_tags (slug, name) VALUES (?, ?)
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name RETURNING id::text AS id`, tag.Slug, tag.Name)
			if err != nil {
				return gerror.Wrap(err, "upsert gallery submission tag")
			}
			if _, err := tx.Exec(`INSERT INTO gallery_submission_tags (submission_id, tag_id) VALUES (?::uuid, ?::uuid)`, value.ID, tagRecord["id"].String()); err != nil {
				return gerror.Wrap(err, "assign gallery submission tag")
			}
		}
		return nil
	})
	if err != nil {
		var postgresError *pq.Error
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return nil, galleryerr.Conflict("submission")
		}
		return nil, gerror.Wrap(err, "create gallery submission")
	}
	return value, nil
}

func (p *PG) MySubmissions(ctx context.Context, subject model.Subject, page, size int) ([]model.Submission, int, error) {
	count, err := p.db.GetValue(ctx, `SELECT COUNT(*) FROM gallery_submissions WHERE subject_kind = ? AND subject_id = ?`, subject.Kind, subject.ID)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "count owned gallery submissions")
	}
	const query = `
SELECT id::text AS id, subject_kind, subject_id, asset_id::text AS asset_id,
       COALESCE(image_id::text, '') AS image_id, title, description, COALESCE(source_url, '') AS source_url,
       alt_text, topic_value_id::text AS topic_value_id, processing_state, review_state, safety_state,
       outcome, failure_code, review_note, created_at, updated_at
FROM gallery_submissions
WHERE subject_kind = ? AND subject_id = ?
ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?`
	var values []model.Submission
	if err := p.db.Ctx(ctx).Raw(query, subject.Kind, subject.ID, size, (page-1)*size).Scan(&values); err != nil {
		return nil, 0, gerror.Wrap(err, "query owned gallery submissions")
	}
	return values, count.Int(), nil
}

func (p *PG) WithdrawSubmission(ctx context.Context, subject model.Subject, id string) (*model.Submission, error) {
	var value *model.Submission
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		record, err := tx.GetOne(`
UPDATE gallery_submissions
SET outcome = 'withdrawn', updated_at = NOW()
WHERE id = ?::uuid AND subject_kind = ? AND subject_id = ?
  AND outcome IN ('pending', 'published')
RETURNING id::text AS id, subject_kind, subject_id, asset_id::text AS asset_id,
          COALESCE(image_id::text, '') AS image_id, title, description, COALESCE(source_url, '') AS source_url,
          alt_text, topic_value_id::text AS topic_value_id, processing_state, review_state, safety_state,
          outcome, failure_code, review_note, created_at, updated_at`, id, subject.Kind, subject.ID)
		if err != nil {
			return gerror.Wrap(err, "withdraw gallery submission")
		}
		value, err = recordAs[model.Submission](record)
		if err != nil || value == nil || value.ImageID == "" {
			return err
		}
		_, err = tx.Exec(`UPDATE gallery_images SET publication_state = 'hidden', hidden_at = NOW(), updated_at = NOW() WHERE id = ?::uuid AND publication_state = 'published'`, value.ImageID)
		return gerror.Wrap(err, "hide withdrawn gallery image")
	})
	return value, err
}

func (p *PG) FailSubmission(ctx context.Context, id, code string) error {
	_, err := p.db.Exec(ctx, `
UPDATE gallery_submissions
SET processing_state = 'failed', outcome = 'failed', failure_code = ?, updated_at = NOW()
WHERE id = ?::uuid AND outcome = 'pending'`, code, id)
	return gerror.Wrap(err, "fail gallery submission")
}

func (p *PG) CreateCase(ctx context.Context, subject model.Subject, imageID string, input model.CaseInput) (*model.Case, error) {
	const query = `
INSERT INTO gallery_cases (image_id, kind, reporter_kind, reporter_id, reason, description, proposed_source_url)
SELECT i.id, ?, NULLIF(?, ''), ?, ?, ?, NULLIF(?, '')
FROM gallery_images i WHERE i.id = ?::uuid AND ` + eligibleImage + `
RETURNING id::text AS id, COALESCE(image_id::text, '') AS image_id,
          COALESCE(submission_id::text, '') AS submission_id, kind, status, reason, description,
          COALESCE(proposed_source_url, '') AS proposed_source_url, resolution_note, created_at, updated_at`
	record, err := p.db.GetOne(ctx, query, input.Kind, subject.Kind, subject.ID, input.Reason, input.Description, input.ProposedSourceURL, imageID)
	if err != nil {
		return nil, gerror.Wrap(err, "create gallery case")
	}
	if len(record) == 0 {
		return nil, galleryerr.NotFound("image", imageID)
	}
	return recordAs[model.Case](record)
}

func (p *PG) AdminOverview(ctx context.Context) (*model.AdminOverview, error) {
	const query = `
SELECT
    (SELECT COUNT(*) FROM gallery_submissions WHERE review_state = 'pending' AND outcome = 'pending')::int AS pending_submissions,
    (SELECT COUNT(*) FROM gallery_cases WHERE status IN ('open', 'reviewing'))::int AS open_cases,
    (SELECT COUNT(*) FROM gallery_images i WHERE ` + eligibleImage + `)::int AS published_images,
    (SELECT COUNT(*) FROM gallery_submissions WHERE processing_state = 'failed')::int AS failed_processing`
	var value *model.AdminOverview
	if err := p.db.Ctx(ctx).Raw(query).Scan(&value); err != nil {
		return nil, gerror.Wrap(err, "query gallery admin overview")
	}
	return value, nil
}

func (p *PG) ReviewQueue(ctx context.Context, page, size int) ([]model.Submission, int, error) {
	count, err := p.db.GetValue(ctx, `SELECT COUNT(*) FROM gallery_submissions WHERE review_state = 'pending' AND outcome = 'pending'`)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "count gallery review queue")
	}
	const query = `
SELECT id::text AS id, subject_kind, subject_id, asset_id::text AS asset_id,
       COALESCE(image_id::text, '') AS image_id, title, description, COALESCE(source_url, '') AS source_url,
       alt_text, topic_value_id::text AS topic_value_id, processing_state, review_state, safety_state,
       outcome, failure_code, review_note, created_at, updated_at
FROM gallery_submissions
WHERE review_state = 'pending' AND outcome = 'pending'
ORDER BY created_at ASC, id ASC LIMIT ? OFFSET ?`
	var values []model.Submission
	if err := p.db.Ctx(ctx).Raw(query, size, (page-1)*size).Scan(&values); err != nil {
		return nil, 0, gerror.Wrap(err, "query gallery review queue")
	}
	return values, count.Int(), nil
}

func (p *PG) ReviewSubmission(ctx context.Context, operator, id string, input model.SubmissionReviewInput) (*model.Submission, error) {
	var value *model.Submission
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		var current *struct {
			model.Submission
			Width                int    `orm:"width"`
			Height               int    `orm:"height"`
			DominantColor        string `orm:"dominant_color"`
			PublicRenditionReady bool   `orm:"public_rendition_ready"`
		}
		if err := tx.Raw(`
SELECT id::text AS id, subject_kind, subject_id, asset_id::text AS asset_id,
       COALESCE(image_id::text, '') AS image_id, title, description, COALESCE(source_url, '') AS source_url,
       alt_text, topic_value_id::text AS topic_value_id, processing_state, review_state, safety_state,
       outcome, failure_code, review_note, created_at, updated_at, width, height, dominant_color, public_rendition_ready
FROM gallery_submissions WHERE id = ?::uuid FOR UPDATE`, id).Scan(&current); err != nil {
			return gerror.Wrap(err, "lock gallery submission review")
		}
		if current == nil || current.ReviewState != "pending" || current.Outcome != "pending" {
			return nil
		}
		if input.Decision == "reject" {
			record, err := tx.GetOne(`
UPDATE gallery_submissions
SET review_state = 'rejected', outcome = 'rejected', review_note = ?, reviewed_by = ?, reviewed_at = NOW(), updated_at = NOW()
WHERE id = ?::uuid
RETURNING id::text AS id, subject_kind, subject_id, asset_id::text AS asset_id,
          COALESCE(image_id::text, '') AS image_id, title, description, COALESCE(source_url, '') AS source_url,
          alt_text, topic_value_id::text AS topic_value_id, processing_state, review_state, safety_state,
          outcome, failure_code, review_note, created_at, updated_at`, input.Note, operator, id)
			if err != nil {
				return gerror.Wrap(err, "reject gallery submission")
			}
			value, err = recordAs[model.Submission](record)
			return err
		}
		if current.ProcessingState != "ready" || !current.PublicRenditionReady || current.Width <= 0 || current.Height <= 0 || !oneOf(current.SafetyState, "safe", "uncertain") {
			return nil
		}
		duplicateValue, err := tx.GetValue(`
SELECT public_image.id::text
FROM gallery_submissions source
JOIN gallery_images public_image ON public_image.exact_sha256 = source.exact_sha256
WHERE source.id = ?::uuid AND source.exact_sha256 IS NOT NULL AND `+strings.ReplaceAll(eligibleImage, "i.", "public_image.")+`
LIMIT 1`, id)
		if err != nil {
			return gerror.Wrap(err, "find exact gallery duplicate")
		}
		duplicateID := duplicateValue.String()
		imageID := duplicateID
		outcome := "duplicate"
		if imageID == "" {
			record, err := tx.GetOne(`
INSERT INTO gallery_images (
    asset_id, origin_submission_id, title, description, source_url, alt_text, width, height,
    dominant_color, processing_state, review_state, publication_state, safety_state,
    public_rendition_ready, exact_sha256, pdq_hash, published_at
)
SELECT asset_id, id, title, description, source_url, alt_text, width, height,
       dominant_color, 'ready', 'approved', 'published', 'safe', TRUE, exact_sha256, pdq_hash, NOW()
FROM gallery_submissions WHERE id = ?::uuid
RETURNING id::text AS id`, id)
			if err != nil {
				return gerror.Wrap(err, "publish approved gallery submission")
			}
			imageID = record["id"].String()
			outcome = "published"
			if _, err := tx.Exec(`INSERT INTO gallery_image_facet_assignments (image_id, facet_id, value_id)
SELECT ?::uuid, facet_id, value_id FROM gallery_submission_facet_assignments WHERE submission_id = ?::uuid`, imageID, id); err != nil {
				return gerror.Wrap(err, "copy approved gallery facets")
			}
			if _, err := tx.Exec(`INSERT INTO gallery_image_tags (image_id, tag_id)
SELECT ?::uuid, tag_id FROM gallery_submission_tags WHERE submission_id = ?::uuid`, imageID, id); err != nil {
				return gerror.Wrap(err, "copy approved gallery tags")
			}
		}
		record, err := tx.GetOne(`
UPDATE gallery_submissions
SET review_state = 'approved', safety_state = 'safe', outcome = ?, image_id = ?::uuid,
    review_note = ?, reviewed_by = ?, reviewed_at = NOW(), updated_at = NOW()
WHERE id = ?::uuid
RETURNING id::text AS id, subject_kind, subject_id, asset_id::text AS asset_id,
          COALESCE(image_id::text, '') AS image_id, title, description, COALESCE(source_url, '') AS source_url,
          alt_text, topic_value_id::text AS topic_value_id, processing_state, review_state, safety_state,
          outcome, failure_code, review_note, created_at, updated_at`, outcome, imageID, input.Note, operator, id)
		if err != nil {
			return gerror.Wrap(err, "complete gallery submission review")
		}
		value, err = recordAs[model.Submission](record)
		return err
	})
	return value, err
}

func (p *PG) HideImage(ctx context.Context, operator, id, reason string) error {
	result, err := p.db.Exec(ctx, `
WITH hidden AS (
    UPDATE gallery_images
    SET publication_state = 'hidden', hidden_at = NOW(), updated_at = NOW()
    WHERE id = ?::uuid AND publication_state = 'published'
    RETURNING id
)
INSERT INTO gallery_cases (image_id, kind, status, reporter_kind, reporter_id, reason, operator_sub)
SELECT id, 'takedown', 'resolved', 'operator', ?, ?, ? FROM hidden`, id, operator, reason, operator)
	if err != nil {
		return gerror.Wrap(err, "hide public gallery image")
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return gerror.Wrap(err, "read gallery hide result")
	}
	if rows == 0 {
		return galleryerr.InvalidState("image", "not_published")
	}
	return nil
}

func (p *PG) Cases(ctx context.Context, status string, page, size int) ([]model.Case, int, error) {
	count, err := p.db.GetValue(ctx, `SELECT COUNT(*) FROM gallery_cases WHERE status = ?`, status)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "count gallery cases")
	}
	const query = `
SELECT id::text AS id, COALESCE(image_id::text, '') AS image_id,
       COALESCE(submission_id::text, '') AS submission_id, kind, status, reason, description,
       COALESCE(proposed_source_url, '') AS proposed_source_url, resolution_note, created_at, updated_at
FROM gallery_cases WHERE status = ?
ORDER BY created_at ASC, id ASC LIMIT ? OFFSET ?`
	var values []model.Case
	if err := p.db.Ctx(ctx).Raw(query, status, size, (page-1)*size).Scan(&values); err != nil {
		return nil, 0, gerror.Wrap(err, "query gallery cases")
	}
	return values, count.Int(), nil
}

func (p *PG) ResolveCase(ctx context.Context, operator, id string, input model.CaseResolutionInput) (*model.Case, error) {
	resolvedAt := "NULL"
	if input.Status == "resolved" || input.Status == "dismissed" {
		resolvedAt = "NOW()"
	}
	query := `
UPDATE gallery_cases
SET status = ?, operator_sub = ?, resolution_note = ?, resolved_at = ` + resolvedAt + `, updated_at = NOW()
WHERE id = ?::uuid AND status IN ('open', 'reviewing')
RETURNING id::text AS id, COALESCE(image_id::text, '') AS image_id,
          COALESCE(submission_id::text, '') AS submission_id, kind, status, reason, description,
          COALESCE(proposed_source_url, '') AS proposed_source_url, resolution_note, created_at, updated_at`
	record, err := p.db.GetOne(ctx, query, input.Status, operator, input.Note, id)
	if err != nil {
		return nil, gerror.Wrap(err, "resolve gallery case")
	}
	return recordAs[model.Case](record)
}

func (p *PG) RecordEvent(ctx context.Context, subject model.Subject, imageID string, input model.EventInput) error {
	result, err := p.db.Exec(ctx, `
INSERT INTO gallery_image_events (image_id, subject_key, session_key, event_type)
SELECT i.id, ?, ?, ? FROM gallery_images i WHERE i.id = ?::uuid AND `+eligibleImage,
		subject.ID, input.SessionKey, input.Type, imageID)
	if err != nil {
		return gerror.Wrap(err, "record gallery image event")
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return gerror.Wrap(err, "read gallery event result")
	}
	if rows == 0 {
		return galleryerr.NotFound("image", imageID)
	}
	return nil
}

func (p *PG) FindSingleton(ctx context.Context, kind, ownerID string) (*model.Collection, error) {
	return p.collection(ctx, `c.kind = ? AND c.owner_id = ?`, kind, ownerID)
}

func (p *PG) CreateCollection(ctx context.Context, input collection.CreateInput) (*model.Collection, error) {
	const query = `
INSERT INTO gallery_collections (kind, resource_kind, owner_kind, owner_id, visibility, name, description)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (owner_id) WHERE kind = 'gallery.favorites'
DO UPDATE SET updated_at = gallery_collections.updated_at
RETURNING id::text AS id, kind, resource_kind, owner_kind, owner_id, visibility, name, description, version, created_at, updated_at`
	record, err := p.db.GetOne(ctx, query, input.Kind, input.ResourceKind, input.OwnerKind, input.OwnerID, input.Visibility, input.Name, input.Description)
	if err != nil {
		return nil, gerror.Wrap(err, "create gallery collection")
	}
	return recordAs[model.Collection](record)
}

func (p *PG) OwnedCollection(ctx context.Context, id, ownerID string) (*model.Collection, error) {
	return p.collection(ctx, `c.id = ?::uuid AND c.owner_id = ?`, id, ownerID)
}

func (p *PG) MutateMembers(ctx context.Context, id string, expectedVersion int64, add, remove []string) (*model.Collection, error) {
	var value *model.Collection
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		locked, err := tx.GetOne(`SELECT version FROM gallery_collections WHERE id = ?::uuid FOR UPDATE`, id)
		if err != nil {
			return gerror.Wrap(err, "lock gallery collection")
		}
		if len(locked) == 0 {
			return galleryerr.NotFound("collection", id)
		}
		if locked["version"].Int64() != expectedVersion {
			return galleryerr.Conflict("collection_version")
		}
		for _, imageID := range add {
			if _, err := tx.Exec(`INSERT INTO gallery_collection_members (collection_id, image_id) VALUES (?::uuid, ?::uuid) ON CONFLICT DO NOTHING`, id, imageID); err != nil {
				return gerror.Wrap(err, "add gallery collection member")
			}
		}
		for _, imageID := range remove {
			if _, err := tx.Exec(`DELETE FROM gallery_collection_members WHERE collection_id = ?::uuid AND image_id = ?::uuid`, id, imageID); err != nil {
				return gerror.Wrap(err, "remove gallery collection member")
			}
		}
		record, err := tx.GetOne(`
UPDATE gallery_collections SET version = version + 1, updated_at = NOW()
WHERE id = ?::uuid
RETURNING id::text AS id, kind, resource_kind, owner_kind, owner_id, visibility, name, description, version, created_at, updated_at`, id)
		if err != nil {
			return gerror.Wrap(err, "bump gallery collection version")
		}
		value, err = recordAs[model.Collection](record)
		return err
	})
	return value, err
}

func (p *PG) Collectable(ctx context.Context, ids []string) (map[string]bool, error) {
	result := make(map[string]bool, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var rows []struct {
		ID string `orm:"id"`
	}
	query := `SELECT i.id::text AS id FROM gallery_images i WHERE i.id = ANY(?::uuid[]) AND ` + eligibleImage
	if err := p.db.Ctx(ctx).Raw(query, pq.Array(ids)).Scan(&rows); err != nil {
		return nil, gerror.Wrap(err, "validate collectable gallery images")
	}
	for _, row := range rows {
		result[row.ID] = true
	}
	return result, nil
}

func (p *PG) CollectionDetail(ctx context.Context, id string, page, size int) (*model.CollectionDetail, error) {
	value, err := p.collection(ctx, `c.id = ?::uuid`, id)
	if err != nil || value == nil {
		return nil, err
	}
	images, err := p.collectionImages(ctx, id, page, size)
	if err != nil {
		return nil, err
	}
	return &model.CollectionDetail{Collection: *value, Images: images}, nil
}

func (p *PG) collection(ctx context.Context, where string, args ...any) (*model.Collection, error) {
	query := `
SELECT c.id::text AS id, c.kind, c.resource_kind, c.owner_kind, c.owner_id, c.visibility,
       c.name, c.description, c.version, c.created_at, c.updated_at,
       COALESCE(e.slug, '') AS slug, COALESCE(e.cover_image_id::text, '') AS cover_image_id,
       (SELECT COUNT(*) FROM gallery_collection_members member WHERE member.collection_id = c.id)::int AS item_count
FROM gallery_collections c
LEFT JOIN gallery_collection_editorial e ON e.collection_id = c.id
WHERE ` + where
	var value *model.Collection
	if err := p.db.Ctx(ctx).Raw(query, args...).Scan(&value); err != nil {
		return nil, gerror.Wrap(err, "query gallery collection")
	}
	return value, nil
}

func (p *PG) collectionImages(ctx context.Context, collectionID string, page, size int) ([]model.ImageCard, error) {
	query := imageCardSelect + `
JOIN gallery_collection_members member ON member.image_id = i.id
WHERE member.collection_id = ?::uuid AND ` + eligibleImage + `
ORDER BY member.manual_position ASC NULLS LAST, member.added_at DESC, i.id DESC
LIMIT ? OFFSET ?`
	return p.cards(ctx, query, collectionID, size, (page-1)*size)
}

func (p *PG) cards(ctx context.Context, query string, args ...any) ([]model.ImageCard, error) {
	var values []model.ImageCard
	if err := p.db.Ctx(ctx).Raw(query, args...).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query gallery image cards")
	}
	return values, nil
}

func recordAs[T any](record gdb.Record) (*T, error) {
	if len(record) == 0 {
		return nil, nil
	}
	var value T
	if err := record.Struct(&value); err != nil {
		return nil, gerror.Wrap(err, fmt.Sprintf("map %T database record", value))
	}
	return &value, nil
}

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}
