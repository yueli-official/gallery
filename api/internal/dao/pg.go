package dao

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/lib/pq"
	"github.com/yueli-official/foundation/go/webhook"
	"github.com/yueli-official/foundation/go/work"
	workpostgres "github.com/yueli-official/foundation/go/work/postgres"

	"github.com/yueli-official/foundation/go/classification"
	"github.com/yueli-official/gallery/api/internal/collection"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/gallerywebhook"
	"github.com/yueli-official/gallery/api/internal/model"
)

type PG struct {
	db       gdb.DB
	work     *workpostgres.Adapter
	webhooks TransactionalWebhookPublisher
}

type TransactionalWebhookPublisher interface {
	PublishTx(context.Context, *sql.Tx, webhook.EventCommand) (webhook.EventReceipt, error)
}

func (p *PG) SetWebhook(runtime TransactionalWebhookPublisher) {
	p.webhooks = runtime
}

func NewPG(db gdb.DB, adapters ...*workpostgres.Adapter) *PG {
	var adapter *workpostgres.Adapter
	if len(adapters) > 0 {
		adapter = adapters[0]
	}
	return &PG{db: db, work: adapter}
}

func (p *PG) enqueueClassificationRefresh(
	ctx context.Context,
	tx gdb.TX,
	catalogID string,
	revision uint64,
	eventType string,
	payload json.RawMessage,
) error {
	if p.work == nil {
		if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_classification_outbox (event_id, catalog_id, revision, event_type, payload)
VALUES (?::uuid, ?::uuid, ?, ?, ?::jsonb)`, newIdentifier(), catalogID, revision, eventType, string(payload)); err != nil {
			return err
		}
	} else {
		envelope, err := json.Marshal(map[string]any{
			"catalogId": catalogID, "revision": revision, "eventType": eventType, "data": json.RawMessage(payload),
		})
		if err != nil {
			return err
		}
		if _, err = p.work.EnqueueTx(ctx, tx.GetSqlTX(), work.Request{
			Kind: "gallery.classification-refresh", Payload: envelope,
			IdempotencyKey: fmt.Sprintf("gallery.classification:%s:%d", catalogID, revision),
		}); err != nil {
			return err
		}
	}
	if p.webhooks == nil {
		return nil
	}
	event, err := json.Marshal(map[string]any{
		"catalogId": catalogID, "revision": revision,
		"changeType": eventType,
	})
	if err != nil {
		return err
	}
	_, err = p.webhooks.PublishTx(ctx, tx.GetSqlTX(), webhook.EventCommand{
		Type: gallerywebhook.ClassificationRevised, Subject: "classification/" + catalogID,
		Data: event, OccurredAt: time.Now().UTC(),
		IdempotencyKey: fmt.Sprintf("gallery:classification:%s:%d", catalogID, revision),
	})
	return err
}

func (p *PG) ClaimGuestSubmissions(ctx context.Context, guestSubject, userID string) (int64, error) {
	result, err := p.db.Exec(ctx, `UPDATE gallery_submissions
SET subject_kind = 'user', subject_id = ?
WHERE subject_kind = 'guest' AND subject_id = ?`, userID, guestSubject)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const eligibleImage = `
i.processing_state = 'ready'
AND i.review_state IN ('not_required', 'approved')
AND i.publication_state = 'published'
AND i.safety_state = 'safe'
AND i.public_rendition_ready`

const imageCardSelect = `
SELECT i.id, i.asset_id, i.title, i.alt_text, i.width, i.height, i.dominant_color, i.published_at,
       COALESCE(primary_category.id::text, '') AS primary_category_id,
       COALESCE(primary_category.name, '') AS primary_category,
       COALESCE(primary_category.slug, '') AS primary_category_slug,
       COALESCE(metric.view_count, 0)::bigint AS view_count,
       COALESCE(metric.favorite_count, 0)::bigint AS favorite_count
FROM gallery_images i
LEFT JOIN LATERAL (
    SELECT category.id, category.name, category.slug
    FROM gallery_image_primary_categories primary_assignment
    JOIN gallery_categories category ON category.id = primary_assignment.category_id
    WHERE primary_assignment.image_id = i.id
) primary_category ON TRUE
LEFT JOIN LATERAL (
    SELECT COALESCE(SUM(m.qualified_views), 0) AS view_count,
           COALESCE(SUM(m.favorites), 0) AS favorite_count
    FROM gallery_image_metrics_daily m
    WHERE m.image_id = i.id
) metric ON TRUE`

const adminImageSelect = `
SELECT i.id::text AS id, i.asset_id::text AS asset_id, i.title, i.description,
       COALESCE(i.source_url, '') AS source_url, i.alt_text, i.width, i.height, i.dominant_color,
       i.processing_state, i.review_state, i.publication_state, i.safety_state,
       i.public_rendition_ready, i.published_at, i.created_at, i.updated_at,
       COALESCE(primary_category.id::text, '') AS primary_category_id,
       COALESCE(primary_category.name, '') AS primary_category,
       COALESCE(primary_category.slug, '') AS primary_category_slug,
       COALESCE(metric.view_count, 0)::bigint AS view_count,
       COALESCE(metric.favorite_count, 0)::bigint AS favorite_count
FROM gallery_images i
LEFT JOIN LATERAL (
    SELECT category.id, category.name, category.slug
    FROM gallery_image_primary_categories primary_assignment
    JOIN gallery_categories category ON category.id = primary_assignment.category_id
    WHERE primary_assignment.image_id = i.id
) primary_category ON TRUE
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
	if value == nil {
		return nil, nil
	}
	if err := p.db.Model("gallery_home_sections").
		Ctx(ctx).
		Where("site_key", "gallery").
		Order("position ASC").
		Scan(&value.HomeSections); err != nil {
		return nil, gerror.Wrap(err, "query gallery home sections")
	}
	for _, section := range value.HomeSections {
		if section.Key == "random" {
			value.RandomBatchSize = section.ItemLimit
			break
		}
	}
	return value, nil
}

func (p *PG) UpdateSiteSettings(ctx context.Context, input model.SiteSettingsUpdateInput) (*model.SiteSettings, error) {
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		randomBatchSize := 24
		for _, section := range input.HomeSections {
			if section.Key == "random" {
				randomBatchSize = section.ItemLimit
				break
			}
		}
		result, err := tx.Exec(`
UPDATE gallery_site_settings
SET name = ?,
    title = ?,
    description = ?,
    search_placeholder = ?,
    footer_tagline = ?,
    random_batch_size = ?,
    random_candidate_size = ?,
    updated_at = NOW()
WHERE site_key = 'gallery'`,
			input.Name,
			input.Title,
			input.Description,
			input.SearchPlaceholder,
			input.FooterTagline,
			randomBatchSize,
			input.RandomCandidateSize,
		)
		if err != nil {
			return gerror.Wrap(err, "update gallery site settings")
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return gerror.Wrap(err, "count updated gallery site settings")
		}
		if affected == 0 {
			return galleryerr.NotInitialized("site_settings")
		}
		if _, err := tx.Exec(`
UPDATE gallery_home_sections
SET position = position + 10
WHERE site_key = 'gallery'`); err != nil {
			return gerror.Wrap(err, "prepare gallery home section reorder")
		}
		for _, section := range input.HomeSections {
			result, err := tx.Exec(`
UPDATE gallery_home_sections
SET enabled = ?,
    position = ?,
    title = ?,
    description = ?,
    action_label = ?,
    item_limit = ?,
    updated_at = NOW()
WHERE site_key = 'gallery' AND section_key = ?`,
				section.Enabled,
				section.Position,
				section.Title,
				section.Description,
				section.ActionLabel,
				section.ItemLimit,
				section.Key,
			)
			if err != nil {
				return gerror.Wrap(err, "update gallery home section")
			}
			affected, err := result.RowsAffected()
			if err != nil {
				return gerror.Wrap(err, "count updated gallery home section")
			}
			if affected == 0 {
				return galleryerr.NotInitialized("home_section_" + section.Key)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return p.SiteSettings(ctx)
}

func (p *PG) ClassificationRevision(ctx context.Context) (uint64, error) {
	value, err := p.db.GetValue(ctx, `
SELECT revision
FROM gallery_classification_catalogs
WHERE catalog_key = 'gallery'`)
	if err != nil {
		return 0, gerror.Wrap(err, "query gallery classification revision")
	}
	if value.IsEmpty() {
		return 0, galleryerr.NotInitialized("classification_catalog")
	}
	return value.Uint64(), nil
}

func (p *PG) ClassificationSnapshot(ctx context.Context) (classification.Snapshot, error) {
	type catalogRow struct {
		ID       string `orm:"id"`
		Revision uint64 `orm:"revision"`
	}
	type policyRow struct {
		Key            string `orm:"policy_key"`
		SchemaVersion  uint16 `orm:"schema_version"`
		PolicyRevision uint64 `orm:"policy_revision"`
		Document       string `orm:"document"`
	}

	var snapshot classification.Snapshot
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		if _, err := tx.Ctx(ctx).Exec(`SET TRANSACTION ISOLATION LEVEL REPEATABLE READ, READ ONLY`); err != nil {
			return gerror.Wrap(err, "start gallery classification snapshot")
		}

		var catalog *catalogRow
		if err := tx.Raw(`
SELECT id::text AS id, revision
FROM gallery_classification_catalogs
WHERE catalog_key = 'gallery'`).Scan(&catalog); err != nil {
			return gerror.Wrap(err, "query gallery classification catalog")
		}
		if catalog == nil {
			return galleryerr.NotInitialized("classification_catalog")
		}

		snapshot = classification.Snapshot{CatalogID: catalog.ID, Revision: catalog.Revision}
		if err := tx.Raw(`
SELECT id::text AS id, COALESCE(parent_id::text, '') AS parent_id, slug, name, status,
       editorial_position, COALESCE(replacement_id::text, '') AS replacement_id
FROM gallery_categories
WHERE catalog_id = ?::uuid
ORDER BY id`, catalog.ID).Scan(&snapshot.Categories); err != nil {
			return gerror.Wrap(err, "query gallery classification categories")
		}
		if err := tx.Raw(`
SELECT id::text AS id, slug, name, status, editorial_position,
       COALESCE(replacement_id::text, '') AS replacement_id
FROM gallery_facets
WHERE catalog_id = ?::uuid
ORDER BY id`, catalog.ID).Scan(&snapshot.Facets); err != nil {
			return gerror.Wrap(err, "query gallery classification facets")
		}
		if err := tx.Raw(`
SELECT id::text AS id, facet_id::text AS facet_id, COALESCE(parent_id::text, '') AS parent_id,
       slug, name, status, editorial_position, COALESCE(replacement_id::text, '') AS replacement_id
FROM gallery_facet_values
WHERE catalog_id = ?::uuid
ORDER BY facet_id, id`, catalog.ID).Scan(&snapshot.FacetValues); err != nil {
			return gerror.Wrap(err, "query gallery classification facet values")
		}

		var rows []policyRow
		if err := tx.Raw(`
SELECT policy_key, schema_version, policy_revision, document::text AS document
FROM gallery_classification_policy_profiles
WHERE catalog_id = ?::uuid
ORDER BY policy_key`, catalog.ID).Scan(&rows); err != nil {
			return gerror.Wrap(err, "query gallery classification policies")
		}
		for _, row := range rows {
			policy := classification.PolicyProfile{
				Key:            row.Key,
				SchemaVersion:  row.SchemaVersion,
				PolicyRevision: row.PolicyRevision,
			}
			decoder := json.NewDecoder(strings.NewReader(row.Document))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&policy); err != nil {
				return gerror.Wrapf(err, "decode gallery classification policy %q", row.Key)
			}
			if err := ensureJSONDocumentEnd(decoder); err != nil {
				return gerror.Wrapf(err, "decode gallery classification policy %q", row.Key)
			}
			snapshot.Policies = append(snapshot.Policies, policy)
		}
		return nil
	})
	if err != nil {
		return classification.Snapshot{}, err
	}
	return snapshot, nil
}

func (p *PG) ClassificationTagMatches(ctx context.Context, lookups []classification.TagLookupRequest) ([]classification.TagMatch, string, error) {
	if len(lookups) == 0 {
		return []classification.TagMatch{}, "", nil
	}
	keys := make([]string, 0, len(lookups))
	for _, lookup := range lookups {
		keys = append(keys, lookup.LookupKey)
	}
	keyArray, err := pq.StringArray(keys).Value()
	if err != nil {
		return nil, "", gerror.Wrap(err, "encode gallery classification tag lookup keys")
	}
	type matchRow struct {
		LookupKey      string `orm:"lookup_key"`
		Kind           string `orm:"kind"`
		TagID          string `orm:"tag_id"`
		FreshnessToken string `orm:"freshness_token"`
	}
	var rows []matchRow
	if err := p.db.Ctx(ctx).Raw(`
WITH requested AS (
    SELECT lookup_key, ordinal
    FROM unnest(?::text[]) WITH ORDINALITY AS input(lookup_key, ordinal)
), catalog AS (
    SELECT id
    FROM gallery_classification_catalogs
    WHERE catalog_key = 'gallery'
)
SELECT requested.lookup_key,
	   CASE
	       WHEN entry.lookup_key IS NULL THEN 'not_found'
	       WHEN target.status = 'active' THEN entry.kind
	       ELSE 'inactive'
	   END AS kind,
	   COALESCE(target.id::text, '') AS tag_id,
       pg_current_snapshot()::text AS freshness_token
FROM requested
CROSS JOIN catalog
LEFT JOIN gallery_tag_lookup_entries entry
  ON entry.catalog_id = catalog.id AND entry.lookup_key = requested.lookup_key
LEFT JOIN gallery_tags target ON target.id = entry.target_tag_id
ORDER BY requested.ordinal`, keyArray).Scan(&rows); err != nil {
		return nil, "", gerror.Wrap(err, "query gallery classification tag matches")
	}
	if len(rows) != len(keys) {
		return nil, "", galleryerr.NotInitialized("classification_catalog")
	}
	matches := make([]classification.TagMatch, 0, len(rows))
	freshnessToken := ""
	for _, row := range rows {
		matches = append(matches, classification.TagMatch{
			LookupKey: row.LookupKey,
			Kind:      classification.TagMatchKind(row.Kind),
			TagID:     row.TagID,
		})
		freshnessToken = row.FreshnessToken
	}
	return matches, freshnessToken, nil
}

func ensureJSONDocumentEnd(decoder *json.Decoder) error {
	var trailing any
	err := decoder.Decode(&trailing)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("unexpected trailing JSON value")
	}
	return err
}

func (p *PG) RandomCandidates(ctx context.Context, limit int) ([]model.ImageCard, error) {
	query := imageCardSelect + ` WHERE ` + eligibleImage + ` ORDER BY i.published_at DESC, i.id DESC LIMIT ?`
	return p.cards(ctx, query, limit)
}

func (p *PG) ClassificationCandidateCounts(ctx context.Context, input model.ImageQuery, requests []classification.CandidateCountGroupRequest) ([]classification.CandidateCountGroup, string, error) {
	if len(requests) == 0 {
		return []classification.CandidateCountGroup{}, "", nil
	}
	type candidatePayload struct {
		ValueID     string   `json:"value_id"`
		MatchingIDs []string `json:"matching_ids"`
	}
	type countRow struct {
		ValueID string `orm:"value_id"`
		Count   int64  `orm:"object_count"`
	}
	groups := make([]classification.CandidateCountGroup, 0, len(requests))
	freshnessToken := ""
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		if _, err := tx.Ctx(ctx).Exec(`SET TRANSACTION ISOLATION LEVEL REPEATABLE READ, READ ONLY`); err != nil {
			return gerror.Wrap(err, "start gallery candidate count snapshot")
		}
		for _, request := range requests {
			payload := make([]candidatePayload, 0, len(request.Candidates))
			for _, candidate := range request.Candidates {
				payload = append(payload, candidatePayload{ValueID: candidate.ValueID, MatchingIDs: candidate.MatchingIDs})
			}
			encoded, err := json.Marshal(payload)
			if err != nil {
				return gerror.Wrap(err, "encode gallery candidate count buckets")
			}
			where, args := publicImagePredicates(input, request.OtherFilters)
			assignmentPredicate := ""
			switch request.Kind {
			case classification.FilterGroupCategory:
				assignmentPredicate = `EXISTS (
                    SELECT 1
                    FROM gallery_image_category_assignments assignment
                    WHERE assignment.image_id = i.id
                      AND assignment.category_id = ANY(candidate.matching_ids)
                )`
			case classification.FilterGroupFacet:
				assignmentPredicate = `EXISTS (
                    SELECT 1
                    FROM gallery_image_facet_assignments assignment
                    WHERE assignment.image_id = i.id
                      AND assignment.facet_value_id = ANY(candidate.matching_ids)
                )`
			default:
				return fmt.Errorf("unsupported gallery candidate group %q", request.Kind)
			}
			query := `
WITH candidates AS (
    SELECT value_id::uuid AS value_id,
           ARRAY(SELECT item::uuid FROM jsonb_array_elements_text(matching_ids) AS elements(item))::uuid[] AS matching_ids
    FROM jsonb_to_recordset(?::jsonb) AS candidate(value_id text, matching_ids jsonb)
)
SELECT candidate.value_id::text AS value_id,
       (
           SELECT COUNT(DISTINCT i.id)::bigint
           FROM gallery_images i
           WHERE ` + strings.Join(where, " AND ") + `
             AND ` + assignmentPredicate + `
       ) AS object_count
FROM candidates candidate
ORDER BY candidate.value_id`
			queryArgs := append([]any{string(encoded)}, args...)
			var rows []countRow
			if err := tx.Raw(query, queryArgs...).Scan(&rows); err != nil {
				return gerror.Wrap(err, "query gallery contextual candidate counts")
			}
			group := classification.CandidateCountGroup{Kind: request.Kind, OwnerID: request.OwnerID}
			for _, row := range rows {
				group.Counts = append(group.Counts, classification.CandidateCount{ValueID: row.ValueID, Count: row.Count})
			}
			groups = append(groups, group)
		}
		value, err := tx.GetValue(`SELECT pg_current_snapshot()::text`)
		if err != nil {
			return gerror.Wrap(err, "read gallery candidate count snapshot")
		}
		freshnessToken = value.String()
		return nil
	})
	if err != nil {
		return nil, "", err
	}
	return groups, freshnessToken, nil
}

func (p *PG) ListImages(ctx context.Context, input model.ImageQuery, plan classification.FilterPlan) ([]model.ImageCard, int, error) {
	where, args := publicImagePredicates(input, plan)
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

func publicImagePredicates(input model.ImageQuery, plan classification.FilterPlan) ([]string, []any) {
	where := []string{eligibleImage}
	args := make([]any, 0)
	if input.Search != "" {
		where = append(where, `(i.title ILIKE ?
			OR i.description ILIKE ?
			OR i.alt_text ILIKE ?
			OR EXISTS (
				SELECT 1
				FROM gallery_image_tag_assignments search_assignment
				JOIN gallery_tags search_tag ON search_tag.id = search_assignment.tag_id
				WHERE search_assignment.image_id = i.id
				  AND search_tag.status = 'active'
				  AND (search_tag.current_name ILIKE ? OR search_tag.current_slug ILIKE ?)
			))`)
		term := "%" + input.Search + "%"
		args = append(args, term, term, term, term, term)
	}
	if input.Tag != "" {
		where = append(where, `EXISTS (
            SELECT 1 FROM gallery_image_tag_assignments it
            JOIN gallery_tags t ON t.id = it.tag_id
			WHERE it.image_id = i.id AND t.current_slug = ? AND t.status = 'active'
        )`)
		args = append(args, input.Tag)
	}
	for _, group := range plan.Groups {
		switch group.Kind {
		case classification.FilterGroupCategory:
			where = append(where, `EXISTS (
                SELECT 1
                FROM gallery_image_category_assignments category_assignment
                WHERE category_assignment.image_id = i.id
                  AND category_assignment.category_id = ANY(?::uuid[])
            )`)
			args = append(args, uuidArrayLiteral(group.ValueIDs))
		case classification.FilterGroupFacet:
			where = append(where, `EXISTS (
                SELECT 1
                FROM gallery_image_facet_assignments facet_assignment
                JOIN gallery_facet_values facet_value ON facet_value.id = facet_assignment.facet_value_id
                WHERE facet_assignment.image_id = i.id
                  AND facet_value.facet_id = ?::uuid
                  AND facet_assignment.facet_value_id = ANY(?::uuid[])
            )`)
			args = append(args, group.OwnerID, uuidArrayLiteral(group.ValueIDs))
		}
	}
	return where, args
}

func uuidArrayLiteral(values []string) string {
	return "{" + strings.Join(values, ",") + "}"
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
SELECT t.current_name AS name FROM gallery_tags t
JOIN gallery_image_tag_assignments it ON it.tag_id = t.id
WHERE it.image_id = ? ORDER BY t.current_name`, id).Scan(&tags); err != nil {
		return nil, gerror.Wrap(err, "query gallery image tags")
	}
	value.Tags = make([]string, 0, len(tags))
	for _, tag := range tags {
		value.Tags = append(value.Tags, tag.Name)
	}
	if err := p.db.Ctx(ctx).Raw(`
SELECT value.facet_id::text AS facet_id, assignment.facet_value_id::text AS value_id
FROM gallery_image_facet_assignments assignment
JOIN gallery_facet_values value ON value.id = assignment.facet_value_id
WHERE assignment.image_id = ?::uuid
ORDER BY value.facet_id, assignment.facet_value_id`, id).Scan(&value.Facets); err != nil {
		return nil, gerror.Wrap(err, "query gallery image facets")
	}
	return value, nil
}

func (p *PG) RelatedImages(ctx context.Context, id string, limit int) ([]model.RelatedImage, error) {
	type relatedRow struct {
		model.ImageCard
		SamePrimary  bool           `orm:"same_primary"`
		SharedFacets pq.StringArray `orm:"shared_facets"`
		SharedTags   pq.StringArray `orm:"shared_tags"`
	}
	query := `
WITH source AS (
    SELECT source_image.id,
           (SELECT category_id FROM gallery_image_primary_categories WHERE image_id = source_image.id) AS primary_category_id
    FROM gallery_images source_image
    WHERE source_image.id = ?::uuid AND ` + strings.ReplaceAll(eligibleImage, "i.", "source_image.") + `
), candidate_signals AS (
    SELECT candidate.id,
           COALESCE(candidate_primary.category_id = source.primary_category_id, FALSE) AS same_primary,
           COALESCE((
               SELECT ARRAY_AGG(DISTINCT facet_value.name ORDER BY facet_value.name)
               FROM gallery_image_facet_assignments candidate_facet
               JOIN gallery_image_facet_assignments source_facet
                 ON source_facet.image_id = source.id
                AND source_facet.facet_value_id = candidate_facet.facet_value_id
               JOIN gallery_facet_values facet_value ON facet_value.id = candidate_facet.facet_value_id
               WHERE candidate_facet.image_id = candidate.id
           ), ARRAY[]::text[]) AS shared_facets,
           COALESCE((
               SELECT ARRAY_AGG(DISTINCT tag.current_name ORDER BY tag.current_name)
               FROM gallery_image_tag_assignments candidate_tag
               JOIN gallery_image_tag_assignments source_tag
                 ON source_tag.image_id = source.id
                AND source_tag.tag_id = candidate_tag.tag_id
               JOIN gallery_tags tag ON tag.id = candidate_tag.tag_id
               WHERE candidate_tag.image_id = candidate.id
           ), ARRAY[]::text[]) AS shared_tags
    FROM gallery_images candidate
    CROSS JOIN source
    LEFT JOIN gallery_image_primary_categories candidate_primary ON candidate_primary.image_id = candidate.id
    WHERE candidate.id <> source.id AND ` + strings.ReplaceAll(eligibleImage, "i.", "candidate.") + `
), ranked AS (
    SELECT candidate_signals.*,
           (CASE WHEN same_primary THEN 40 ELSE 0 END)
             + CARDINALITY(shared_facets) * 12
             + CARDINALITY(shared_tags) * 6 AS relationship_score
    FROM candidate_signals
)
SELECT card.*, ranked.same_primary, ranked.shared_facets, ranked.shared_tags
FROM ranked
JOIN LATERAL (` + imageCardSelect + ` WHERE i.id = ranked.id) card ON TRUE
ORDER BY ranked.relationship_score DESC, card.published_at DESC, card.id DESC
LIMIT ?`
	var rows []relatedRow
	if err := p.db.Ctx(ctx).Raw(query, id, limit).Scan(&rows); err != nil {
		return nil, gerror.Wrap(err, "query related gallery images")
	}
	values := make([]model.RelatedImage, 0, len(rows))
	for _, row := range rows {
		reasons := make([]model.RelatedImageReason, 0, 3)
		if row.SamePrimary && row.PrimaryCategory != "" {
			reasons = append(reasons, model.RelatedImageReason{Kind: "primary_category", Label: "同属" + row.PrimaryCategory})
		}
		for _, value := range row.SharedFacets {
			if len(reasons) == 3 {
				break
			}
			reasons = append(reasons, model.RelatedImageReason{Kind: "facet", Label: "共享属性：" + value})
		}
		for _, value := range row.SharedTags {
			if len(reasons) == 3 {
				break
			}
			reasons = append(reasons, model.RelatedImageReason{Kind: "tag", Label: "共享标签：" + value})
		}
		values = append(values, model.RelatedImage{ImageCard: row.ImageCard, Reasons: reasons})
	}
	return values, nil
}

func (p *PG) HasTombstone(ctx context.Context, id string) (bool, error) {
	count, err := p.db.Model("gallery_image_tombstones").Ctx(ctx).Where("image_id", id).Count()
	return count > 0, err
}

func (p *PG) PublicCollections(ctx context.Context) ([]model.Collection, error) {
	query := `
SELECT c.id::text AS id, c.kind, c.resource_kind, c.owner_kind, c.visibility, c.name, c.description,
       c.version, c.created_at, c.updated_at, e.slug, e.seo_title, e.seo_description,
       COALESCE(e.cover_image_id::text, '') AS cover_image_id,
       COALESCE(cover_image.asset_id::text, '') AS cover_asset_id,
       COALESCE(cover_image.alt_text, '') AS cover_alt_text,
       COALESCE(cover_image.width, 0) AS cover_width, COALESCE(cover_image.height, 0) AS cover_height,
       COALESCE(cover_image.dominant_color, '') AS cover_color,
       COUNT(i.id)::int AS item_count
FROM gallery_collections c
JOIN gallery_collection_editorial e ON e.collection_id = c.id
LEFT JOIN gallery_images cover_image ON cover_image.id = e.cover_image_id AND ` + strings.ReplaceAll(eligibleImage, "i.", "cover_image.") + `
LEFT JOIN gallery_collection_members member ON member.collection_id = c.id
LEFT JOIN gallery_images i ON i.id = member.image_id AND ` + eligibleImage + `
WHERE c.kind = 'gallery.editorial' AND c.visibility = 'public'
GROUP BY c.id, e.slug, e.cover_image_id, e.seo_title, e.seo_description,
         cover_image.asset_id, cover_image.alt_text, cover_image.width, cover_image.height, cover_image.dominant_color
ORDER BY c.updated_at DESC, c.id DESC`
	var values []model.Collection
	if err := p.db.Ctx(ctx).Raw(query).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query public gallery collections")
	}
	return values, nil
}

func (p *PG) EditorialCollections(ctx context.Context) ([]model.Collection, error) {
	query := `
SELECT c.id::text AS id, c.kind, c.resource_kind, c.owner_kind, c.visibility, c.name, c.description,
       c.version, c.created_at, c.updated_at, e.slug, e.seo_title, e.seo_description,
       COALESCE(e.cover_image_id::text, '') AS cover_image_id,
       COALESCE(cover_image.asset_id::text, '') AS cover_asset_id,
       COALESCE(cover_image.alt_text, '') AS cover_alt_text,
       COALESCE(cover_image.width, 0) AS cover_width, COALESCE(cover_image.height, 0) AS cover_height,
       COALESCE(cover_image.dominant_color, '') AS cover_color,
       COUNT(member.image_id)::int AS item_count
FROM gallery_collections c
JOIN gallery_collection_editorial e ON e.collection_id = c.id
LEFT JOIN gallery_images cover_image ON cover_image.id = e.cover_image_id
LEFT JOIN gallery_collection_members member ON member.collection_id = c.id
WHERE c.kind = 'gallery.editorial'
GROUP BY c.id, e.slug, e.cover_image_id, e.seo_title, e.seo_description,
         cover_image.asset_id, cover_image.alt_text, cover_image.width, cover_image.height, cover_image.dominant_color
ORDER BY c.updated_at DESC, c.id DESC`
	var values []model.Collection
	if err := p.db.Ctx(ctx).Raw(query).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query editorial gallery collections")
	}
	return values, nil
}

func (p *PG) PublicCollection(ctx context.Context, slug string, page, size int) (*model.CollectionDetail, error) {
	query := `
SELECT c.id::text AS id, c.kind, c.resource_kind, c.owner_kind, c.visibility, c.name, c.description,
       c.version, c.created_at, c.updated_at, e.slug, e.seo_title, e.seo_description,
       COALESCE(e.cover_image_id::text, '') AS cover_image_id,
       COALESCE(cover_image.asset_id::text, '') AS cover_asset_id,
       COALESCE(cover_image.alt_text, '') AS cover_alt_text,
       COALESCE(cover_image.width, 0) AS cover_width, COALESCE(cover_image.height, 0) AS cover_height,
       COALESCE(cover_image.dominant_color, '') AS cover_color,
       (SELECT COUNT(*) FROM gallery_collection_members count_member
        JOIN gallery_images count_image ON count_image.id = count_member.image_id
        WHERE count_member.collection_id = c.id AND ` + strings.ReplaceAll(eligibleImage, "i.", "count_image.") + `)::int AS item_count
FROM gallery_collections c
JOIN gallery_collection_editorial e ON e.collection_id = c.id
LEFT JOIN gallery_images cover_image ON cover_image.id = e.cover_image_id AND ` + strings.ReplaceAll(eligibleImage, "i.", "cover_image.") + `
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
INSERT INTO gallery_collections (id, kind, resource_kind, owner_kind, owner_id, visibility, name, description)
VALUES (?::uuid, 'gallery.editorial', 'gallery.image', 'site', 'gallery', ?, ?, ?)
RETURNING id::text AS id, kind, resource_kind, owner_kind, owner_id, visibility, name, description, version, created_at, updated_at`,
			newIdentifier(), input.Visibility, input.Name, input.Description)
		if err != nil {
			return gerror.Wrap(err, "create editorial collection")
		}
		value, err = recordAs[model.Collection](record)
		if err != nil {
			return err
		}
		if _, err := tx.Ctx(ctx).Exec(`INSERT INTO gallery_collection_editorial (collection_id, slug) VALUES (?::uuid, ?)`, value.ID, input.Slug); err != nil {
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

func (p *PG) UpdateEditorialCollection(ctx context.Context, id string, input model.EditorialCollectionUpdateInput) (*model.Collection, error) {
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		locked, err := tx.GetOne(`
SELECT c.version
FROM gallery_collections c
JOIN gallery_collection_editorial editorial ON editorial.collection_id = c.id
WHERE c.id = ?::uuid AND c.kind = 'gallery.editorial' AND c.owner_id = 'gallery'
FOR UPDATE OF c`, id)
		if err != nil {
			return gerror.Wrap(err, "lock editorial collection")
		}
		if len(locked) == 0 {
			return galleryerr.NotFound("collection", id)
		}
		if locked["version"].Int64() != input.Version {
			return galleryerr.Conflict("collection_version")
		}
		if input.CoverImageID != "" {
			cover, err := tx.GetValue(`
SELECT EXISTS (
    SELECT 1 FROM gallery_collection_members member
    JOIN gallery_images i ON i.id = member.image_id
    WHERE member.collection_id = ?::uuid AND member.image_id = ?::uuid AND `+eligibleImage+`
)`, id, input.CoverImageID)
			if err != nil {
				return gerror.Wrap(err, "validate editorial collection cover")
			}
			if !cover.Bool() {
				return galleryerr.Validation("coverImageId", "cover image must be an eligible collection member")
			}
		}
		if _, err := tx.Ctx(ctx).Exec(`
UPDATE gallery_collections
SET name = ?, description = ?, visibility = ?, version = version + 1, updated_at = NOW()
WHERE id = ?::uuid`, input.Name, input.Description, input.Visibility, id); err != nil {
			return gerror.Wrap(err, "update editorial collection")
		}
		coverID := any(nil)
		if input.CoverImageID != "" {
			coverID = input.CoverImageID
		}
		if _, err := tx.Ctx(ctx).Exec(`
UPDATE gallery_collection_editorial
SET slug = ?, cover_image_id = ?::uuid, seo_title = ?, seo_description = ?
WHERE collection_id = ?::uuid`, input.Slug, coverID, input.SEOTitle, input.SEODescription, id); err != nil {
			var postgresError *pq.Error
			if errors.As(err, &postgresError) && postgresError.Code == "23505" {
				return galleryerr.Conflict("collection_slug")
			}
			return gerror.Wrap(err, "update editorial collection extension")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return p.collection(ctx, `c.id = ?::uuid AND c.owner_id = 'gallery'`, id)
}

func (p *PG) ReorderEditorialMembers(ctx context.Context, id string, expectedVersion int64, imageIDs []string) (*model.Collection, error) {
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		locked, err := tx.GetOne(`
SELECT c.version
FROM gallery_collections c
JOIN gallery_collection_editorial editorial ON editorial.collection_id = c.id
WHERE c.id = ?::uuid AND c.kind = 'gallery.editorial' AND c.owner_id = 'gallery'
FOR UPDATE OF c`, id)
		if err != nil {
			return gerror.Wrap(err, "lock editorial collection order")
		}
		if len(locked) == 0 {
			return galleryerr.NotFound("collection", id)
		}
		if locked["version"].Int64() != expectedVersion {
			return galleryerr.Conflict("collection_version")
		}
		var members []struct {
			ID string `orm:"id"`
		}
		if err := tx.Raw(`SELECT image_id::text AS id FROM gallery_collection_members WHERE collection_id = ?::uuid`, id).Scan(&members); err != nil {
			return gerror.Wrap(err, "query editorial collection members for reorder")
		}
		if len(members) != len(imageIDs) {
			return galleryerr.Conflict("collection_members")
		}
		requested := make(map[string]struct{}, len(imageIDs))
		for _, imageID := range imageIDs {
			requested[imageID] = struct{}{}
		}
		for _, member := range members {
			if _, exists := requested[member.ID]; !exists {
				return galleryerr.Conflict("collection_members")
			}
		}
		for position, imageID := range imageIDs {
			if _, err := tx.Ctx(ctx).Exec(`UPDATE gallery_collection_members SET manual_position = ? WHERE collection_id = ?::uuid AND image_id = ?::uuid`, position, id, imageID); err != nil {
				return gerror.Wrap(err, "reorder editorial collection member")
			}
		}
		if _, err := tx.Ctx(ctx).Exec(`UPDATE gallery_collections SET version = version + 1, updated_at = NOW() WHERE id = ?::uuid`, id); err != nil {
			return gerror.Wrap(err, "bump editorial collection order version")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return p.collection(ctx, `c.id = ?::uuid AND c.owner_id = 'gallery'`, id)
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
	snapshot, err := p.currentRankingSnapshot(ctx, kind, window)
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		snapshot, err = p.refreshRankingSnapshot(ctx, kind, window, cutoff, max(limit, 60))
		if err != nil {
			// A snapshot refresh is an optimization boundary. If the writer is
			// temporarily unavailable, serve the latest materialized result instead
			// of turning the public ranking page into a hard failure.
			stale, staleErr := p.latestRankingSnapshot(ctx, kind, window)
			if staleErr != nil || stale == nil {
				return nil, err
			}
			snapshot = stale
		}
	}
	query := imageCardSelect + `
JOIN gallery_ranking_entries ranking_entry ON ranking_entry.image_id = i.id
WHERE ranking_entry.snapshot_id = ?::uuid AND ` + eligibleImage + `
ORDER BY ranking_entry.rank
LIMIT ?`
	values, err := p.cards(ctx, query, snapshot.ID, limit)
	if err != nil {
		return nil, err
	}
	return &model.Ranking{Kind: kind, Window: window, Generated: gtime.NewFromTime(snapshot.GeneratedAt), Images: values}, nil
}

func (p *PG) latestRankingSnapshot(ctx context.Context, kind, window string) (*rankingSnapshot, error) {
	var rows []rankingSnapshot
	if err := p.db.Ctx(ctx).Raw(`
SELECT id::text AS id, generated_at
FROM gallery_ranking_snapshots
WHERE ranking_kind = ? AND window_key = ?
ORDER BY generated_at DESC
LIMIT 1`, kind, window).Scan(&rows); err != nil {
		return nil, gerror.Wrap(err, "query latest gallery ranking snapshot")
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

type rankingSnapshot struct {
	ID          string    `orm:"id"`
	GeneratedAt time.Time `orm:"generated_at"`
}

func (p *PG) currentRankingSnapshot(ctx context.Context, kind, window string) (*rankingSnapshot, error) {
	var rows []rankingSnapshot
	if err := p.db.Ctx(ctx).Raw(`
SELECT id::text AS id, generated_at
FROM gallery_ranking_snapshots
WHERE ranking_kind = ? AND window_key = ? AND expires_at > NOW()
ORDER BY generated_at DESC
LIMIT 1`, kind, window).Scan(&rows); err != nil {
		return nil, gerror.Wrap(err, "query current gallery ranking snapshot")
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func (p *PG) refreshRankingSnapshot(ctx context.Context, kind, window string, cutoff time.Time, limit int) (*rankingSnapshot, error) {
	score := map[string]string{
		"trending":       "COALESCE(SUM(metric.qualified_views), 0) + COALESCE(SUM(metric.favorites), 0) * 4",
		"most_viewed":    "COALESCE(SUM(metric.qualified_views), 0)",
		"most_favorited": "COALESCE(SUM(metric.favorites), 0)",
	}[kind]
	var snapshot *rankingSnapshot
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		if _, err := tx.Ctx(ctx).Exec(`SELECT pg_advisory_xact_lock(hashtextextended(?::text, 0))`, "gallery-ranking:"+kind+":"+window); err != nil {
			return gerror.Wrap(err, "lock gallery ranking refresh")
		}
		var existing []rankingSnapshot
		if err := tx.Raw(`
SELECT id::text AS id, generated_at
FROM gallery_ranking_snapshots
WHERE ranking_kind = ? AND window_key = ? AND expires_at > NOW()
ORDER BY generated_at DESC
LIMIT 1`, kind, window).Scan(&existing); err != nil {
			return gerror.Wrap(err, "recheck gallery ranking snapshot")
		}
		if len(existing) != 0 {
			snapshot = &existing[0]
			return nil
		}
		var created []rankingSnapshot
		if err := tx.Raw(`
INSERT INTO gallery_ranking_snapshots (id, ranking_kind, window_key, generated_at, expires_at)
VALUES (?::uuid, ?, ?, NOW(), NOW() + INTERVAL '5 minutes')
RETURNING id::text AS id, generated_at`, newIdentifier(), kind, window).Scan(&created); err != nil {
			return gerror.Wrap(err, "create gallery ranking snapshot")
		}
		if len(created) != 1 {
			return errors.New("gallery ranking snapshot insert returned no row")
		}
		snapshot = &created[0]
		query := `
WITH scores AS (
    SELECT i.id, i.published_at,
           (` + score + `)::double precision AS score
    FROM gallery_images i
    LEFT JOIN gallery_image_metrics_daily metric
      ON metric.image_id = i.id AND metric.metric_date >= ?::date
    WHERE ` + eligibleImage + `
    GROUP BY i.id, i.published_at
), ranked AS (
    SELECT id, score,
           ROW_NUMBER() OVER (ORDER BY score DESC, published_at DESC, id DESC) AS rank
    FROM scores
)
INSERT INTO gallery_ranking_entries (snapshot_id, image_id, rank, score)
SELECT ?::uuid, id, rank, score
FROM ranked
WHERE rank <= ?`
		if _, err := tx.Ctx(ctx).Exec(query, cutoff, snapshot.ID, limit); err != nil {
			return gerror.Wrap(err, "populate gallery ranking snapshot")
		}
		if _, err := tx.Ctx(ctx).Exec(`DELETE FROM gallery_ranking_snapshots WHERE expires_at < NOW() - INTERVAL '1 day'`); err != nil {
			return gerror.Wrap(err, "prune gallery ranking snapshots")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (p *PG) CreateSubmission(ctx context.Context, subject model.Subject, input model.SubmissionInput, review string) (*model.Submission, error) {
	const query = `
INSERT INTO gallery_submissions (
    id, subject_kind, subject_id, asset_id, title, description, source_url, alt_text, review_state
)
VALUES (?::uuid, ?, ?, ?::uuid, ?, ?, NULLIF(?, ''), ?, ?)
RETURNING id::text AS id, subject_kind, subject_id, asset_id::text AS asset_id,
          COALESCE(image_id::text, '') AS image_id, title, description, COALESCE(source_url, '') AS source_url,
          alt_text, processing_state, review_state, safety_state,
          outcome, failure_code, review_note, created_at, updated_at`
	var value *model.Submission
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		catalog, err := tx.GetOne(`
SELECT revision
FROM gallery_classification_catalogs
WHERE catalog_key = 'gallery'
FOR SHARE`)
		if err != nil {
			return gerror.Wrap(err, "lock gallery classification revision")
		}
		if len(catalog) == 0 || catalog["revision"].Uint64() != input.Classification.CatalogRevision {
			return galleryerr.Conflict("classification_revision")
		}
		if _, err := tx.Ctx(ctx).Exec(`SELECT pg_advisory_xact_lock(?)`, submissionLockKey(subject.Kind, subject.ID, input.AssetID)); err != nil {
			return gerror.Wrap(err, "lock gallery pending submission identity")
		}
		pending, err := tx.GetValue(`
SELECT EXISTS (
    SELECT 1 FROM gallery_submissions
    WHERE subject_kind = ? AND subject_id = ? AND asset_id = ?::uuid AND outcome = 'pending'
)`, subject.Kind, subject.ID, input.AssetID)
		if err != nil {
			return gerror.Wrap(err, "find pending gallery asset submission")
		}
		if pending.Bool() {
			return galleryerr.Conflict("submission")
		}
		if len(input.Classification.TagCreations) != 0 {
			return galleryerr.NotInitialized("classification_tag_creation")
		}
		record, err := tx.GetOne(query, newIdentifier(), subject.Kind, subject.ID, input.AssetID, input.Title, input.Description,
			input.SourceURL, input.AltText, review)
		if err != nil {
			return gerror.Wrap(err, "insert gallery submission")
		}
		value, err = recordAs[model.Submission](record)
		if err != nil {
			return err
		}
		for _, assignment := range input.Classification.Categories {
			if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_submission_category_assignments (submission_id, category_id)
VALUES (?::uuid, ?::uuid)`, value.ID, assignment.CategoryID); err != nil {
				return gerror.Wrap(err, "assign gallery submission category")
			}
		}
		if input.Classification.PrimaryCategoryID != "" {
			if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_submission_primary_categories (submission_id, category_id)
VALUES (?::uuid, ?::uuid)`, value.ID, input.Classification.PrimaryCategoryID); err != nil {
				return gerror.Wrap(err, "assign gallery submission primary category")
			}
			value.PrimaryCategoryID = input.Classification.PrimaryCategoryID
		}
		for _, assignment := range input.Classification.Facets {
			if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_submission_facet_assignments (submission_id, facet_value_id)
VALUES (?::uuid, ?::uuid)`, value.ID, assignment.ValueID); err != nil {
				return gerror.Wrap(err, "assign gallery submission facet")
			}
		}
		for _, tag := range input.Classification.Tags {
			if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_submission_tag_assignments (submission_id, tag_id)
VALUES (?::uuid, ?::uuid)`, value.ID, tag.TagID); err != nil {
				return gerror.Wrap(err, "assign gallery submission tag")
			}
		}
		for _, proposal := range input.Classification.TagProposals {
			if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_tag_proposals (id, submission_id, input_value, lookup_key)
VALUES (?::uuid, ?::uuid, ?, ?)`, newIdentifier(), value.ID, proposal.DisplayValue, proposal.LookupKey); err != nil {
				return gerror.Wrap(err, "create gallery submission tag proposal")
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

func submissionLockKey(subjectKind, subjectID, assetID string) int64 {
	sum := sha256.Sum256([]byte(subjectKind + "\x00" + subjectID + "\x00" + assetID))
	return int64(binary.BigEndian.Uint64(sum[:8]))
}

func (p *PG) MySubmissions(ctx context.Context, subject model.Subject, input model.MySubmissionQuery) ([]model.Submission, int, error) {
	where := []string{"submission.subject_kind = ?", "submission.subject_id = ?"}
	args := []any{subject.Kind, subject.ID}
	if input.Outcome != "" {
		where, args = append(where, "submission.outcome = ?"), append(args, input.Outcome)
	}
	if input.ProcessingState != "" {
		where, args = append(where, "submission.processing_state = ?"), append(args, input.ProcessingState)
	}
	if input.ReviewState != "" {
		where, args = append(where, "submission.review_state = ?"), append(args, input.ReviewState)
	}
	predicate := strings.Join(where, " AND ")
	count, err := p.db.GetValue(ctx, `SELECT COUNT(*) FROM gallery_submissions submission WHERE `+predicate, args...)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "count owned gallery submissions")
	}
	query := `
SELECT submission.id::text AS id, submission.subject_kind, submission.subject_id, submission.asset_id::text AS asset_id,
       COALESCE(submission.image_id::text, '') AS image_id, submission.title, submission.description,
       COALESCE(submission.source_url, '') AS source_url, submission.alt_text,
       COALESCE(primary_category.category_id::text, '') AS primary_category_id,
       submission.processing_state, submission.review_state, submission.safety_state,
       submission.outcome, submission.failure_code, submission.review_note, submission.created_at, submission.updated_at
FROM gallery_submissions submission
LEFT JOIN gallery_submission_primary_categories primary_category ON primary_category.submission_id = submission.id
WHERE ` + predicate + `
ORDER BY submission.created_at DESC, submission.id DESC LIMIT ? OFFSET ?`
	var values []model.Submission
	pageArgs := append(append([]any{}, args...), input.PageSize, (input.Page-1)*input.PageSize)
	if err := p.db.Ctx(ctx).Raw(query, pageArgs...).Scan(&values); err != nil {
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
		  alt_text, COALESCE((SELECT category_id::text FROM gallery_submission_primary_categories
		    WHERE submission_id = gallery_submissions.id), '') AS primary_category_id,
		  processing_state, review_state, safety_state,
          outcome, failure_code, review_note, created_at, updated_at`, id, subject.Kind, subject.ID)
		if err != nil {
			return gerror.Wrap(err, "withdraw gallery submission")
		}
		value, err = recordAs[model.Submission](record)
		if err != nil || value == nil || value.ImageID == "" {
			return err
		}
		_, err = tx.Ctx(ctx).Exec(`UPDATE gallery_images SET publication_state = 'hidden', hidden_at = NOW(), updated_at = NOW() WHERE id = ?::uuid AND publication_state = 'published'`, value.ImageID)
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

func (p *PG) ClaimSubmissionProcessing(ctx context.Context, limit int) ([]model.Submission, error) {
	var values []model.Submission
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		return tx.Ctx(ctx).Raw(`
WITH claimed AS (
    SELECT id
    FROM gallery_submissions
    WHERE outcome = 'pending'
      AND (processing_state = 'queued' OR (processing_state = 'processing' AND updated_at < NOW() - INTERVAL '5 minutes'))
    ORDER BY created_at, id
    FOR UPDATE SKIP LOCKED
    LIMIT ?
)
UPDATE gallery_submissions submission
SET processing_state = 'processing', failure_code = '', updated_at = NOW()
FROM claimed
WHERE submission.id = claimed.id
RETURNING submission.id::text AS id, submission.asset_id::text AS asset_id,
          submission.title, submission.alt_text, submission.processing_state,
          submission.review_state, submission.safety_state, submission.outcome,
          submission.created_at, submission.updated_at`, limit).Scan(&values)
	})
	return values, gerror.Wrap(err, "claim gallery submission processing")
}

func (p *PG) CompleteSubmissionProcessing(ctx context.Context, id string, facts model.SubmissionAssetFacts) error {
	result, err := p.db.Exec(ctx, `
UPDATE gallery_submissions
SET processing_state = 'ready', safety_state = 'uncertain', width = ?, height = ?,
    exact_sha256 = decode(?, 'hex'), public_rendition_ready = TRUE,
    failure_code = '', updated_at = NOW()
WHERE id = ?::uuid AND processing_state = 'processing' AND outcome = 'pending'`, facts.Width, facts.Height, facts.ContentHash, id)
	if err != nil {
		return gerror.Wrap(err, "complete gallery submission processing")
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return gerror.Wrap(err, "read gallery submission processing result")
	}
	if rows == 0 {
		return galleryerr.InvalidState("submission", "not_processing")
	}
	return nil
}

func (p *PG) SubmissionAssetID(ctx context.Context, id string) (string, error) {
	value, err := p.db.GetValue(ctx, `SELECT asset_id::text FROM gallery_submissions WHERE id = ?::uuid`, id)
	if err != nil {
		return "", gerror.Wrap(err, "get gallery submission asset")
	}
	return value.String(), nil
}

func (p *PG) CreateCase(ctx context.Context, subject model.Subject, imageID string, input model.CaseInput) (*model.Case, error) {
	const query = `
INSERT INTO gallery_cases (id, image_id, kind, reporter_kind, reporter_id, reason, description, proposed_source_url)
SELECT ?::uuid, i.id, ?, NULLIF(?, ''), ?, ?, ?, NULLIF(?, '')
FROM gallery_images i WHERE i.id = ?::uuid AND ` + eligibleImage + `
RETURNING id::text AS id, COALESCE(image_id::text, '') AS image_id,
          COALESCE(submission_id::text, '') AS submission_id, kind, status, reason, description,
          COALESCE(proposed_source_url, '') AS proposed_source_url, resolution_note, created_at, updated_at`
	record, err := p.db.GetOne(ctx, query, newIdentifier(), input.Kind, subject.Kind, subject.ID, input.Reason, input.Description, input.ProposedSourceURL, imageID)
	if err != nil {
		return nil, gerror.Wrap(err, "create gallery case")
	}
	if len(record) == 0 {
		return nil, galleryerr.NotFound("image", imageID)
	}
	return recordAs[model.Case](record)
}

func (p *PG) AdminOverview(ctx context.Context, days int) (*model.AdminOverview, error) {
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
	if err := p.db.Ctx(ctx).Raw(`
SELECT
    COALESCE(SUM(qualified_views), 0)::bigint AS all_time_views,
    COALESCE(SUM(qualified_views) FILTER (WHERE metric_date >= CURRENT_DATE - (? - 1) * INTERVAL '1 day'), 0)::bigint AS period_views,
    COALESCE(SUM(qualified_views) FILTER (WHERE metric_date >= CURRENT_DATE - (? * 2 - 1) * INTERVAL '1 day' AND metric_date < CURRENT_DATE - (? - 1) * INTERVAL '1 day'), 0)::bigint AS previous_period_views,
    COALESCE(SUM(favorites) FILTER (WHERE metric_date >= CURRENT_DATE - (? - 1) * INTERVAL '1 day'), 0)::bigint AS period_favorites,
    COALESCE(SUM(favorites) FILTER (WHERE metric_date >= CURRENT_DATE - (? * 2 - 1) * INTERVAL '1 day' AND metric_date < CURRENT_DATE - (? - 1) * INTERVAL '1 day'), 0)::bigint AS previous_period_favorites
FROM gallery_image_metrics_daily`, days, days, days, days, days, days).Scan(value); err != nil {
		return nil, gerror.Wrap(err, "query gallery dashboard totals")
	}
	if err := p.db.Ctx(ctx).Raw(`
SELECT day::date::text AS day,
       COALESCE(SUM(metric.qualified_views), 0)::bigint AS views,
       COALESCE(SUM(metric.favorites), 0)::bigint AS favorites
FROM generate_series(CURRENT_DATE - (? - 1) * INTERVAL '1 day', CURRENT_DATE, INTERVAL '1 day') day
LEFT JOIN gallery_image_metrics_daily metric ON metric.metric_date = day::date
GROUP BY day ORDER BY day`, days).Scan(&value.Series); err != nil {
		return nil, gerror.Wrap(err, "query gallery dashboard series")
	}
	if err := p.db.Ctx(ctx).Raw(`
SELECT image.id::text AS id, image.title,
       COALESCE(SUM(metric.qualified_views), 0)::bigint AS views,
       COALESCE(SUM(metric.favorites), 0)::bigint AS favorites
FROM gallery_images image
JOIN gallery_image_metrics_daily metric ON metric.image_id = image.id
WHERE metric.metric_date >= CURRENT_DATE - (? - 1) * INTERVAL '1 day'
  AND image.processing_state = 'ready'
  AND image.review_state IN ('not_required', 'approved')
  AND image.publication_state = 'published'
  AND image.safety_state = 'safe'
  AND image.public_rendition_ready
GROUP BY image.id, image.title
ORDER BY views DESC, favorites DESC, image.id
LIMIT 8`, days).Scan(&value.TopImages); err != nil {
		return nil, gerror.Wrap(err, "query gallery dashboard top images")
	}
	return value, nil
}

func (p *PG) AdminImages(ctx context.Context, input model.AdminImageQuery) ([]model.AdminImage, int, error) {
	where := []string{"i.review_state IN ('approved', 'not_required')"}
	args := []any{}
	if input.Search != "" {
		where, args = append(where, "(i.title ILIKE ? OR i.description ILIKE ? OR i.alt_text ILIKE ?)"), append(args, "%"+input.Search+"%", "%"+input.Search+"%", "%"+input.Search+"%")
	}
	if input.ProcessingState != "" {
		where, args = append(where, "i.processing_state = ?"), append(args, input.ProcessingState)
	}
	if input.ReviewState != "" {
		where, args = append(where, "i.review_state = ?"), append(args, input.ReviewState)
	}
	if input.PublicationState != "" {
		where, args = append(where, "i.publication_state = ?"), append(args, input.PublicationState)
	}
	if input.SafetyState != "" {
		where, args = append(where, "i.safety_state = ?"), append(args, input.SafetyState)
	}
	if input.CategoryID != "" {
		where, args = append(where, "EXISTS (SELECT 1 FROM gallery_image_category_assignments category_assignment WHERE category_assignment.image_id = i.id AND category_assignment.category_id = ?::uuid)"), append(args, input.CategoryID)
	}
	if input.FacetValueID != "" {
		where, args = append(where, "EXISTS (SELECT 1 FROM gallery_image_facet_assignments facet_assignment WHERE facet_assignment.image_id = i.id AND facet_assignment.facet_value_id = ?::uuid)"), append(args, input.FacetValueID)
	}
	predicate := strings.Join(where, " AND ")
	count, err := p.db.GetValue(ctx, `SELECT COUNT(*) FROM gallery_images i WHERE `+predicate, args...)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "count admin gallery images")
	}
	orders := map[string]map[string]string{
		"createdAt": {"asc": "i.created_at ASC, i.id ASC", "desc": "i.created_at DESC, i.id DESC"},
		"updatedAt": {"asc": "i.updated_at ASC, i.id ASC", "desc": "i.updated_at DESC, i.id DESC"},
		"title":     {"asc": "LOWER(i.title) ASC, i.id ASC", "desc": "LOWER(i.title) DESC, i.id DESC"},
		"views":     {"asc": "view_count ASC, i.id ASC", "desc": "view_count DESC, i.id DESC"},
	}
	query := adminImageSelect + ` WHERE ` + predicate + ` ORDER BY ` + orders[input.SortBy][input.SortOrder] + ` LIMIT ? OFFSET ?`
	pageArgs := append(append([]any{}, args...), input.PageSize, (input.Page-1)*input.PageSize)
	var values []model.AdminImage
	if err := p.db.Ctx(ctx).Raw(query, pageArgs...).Scan(&values); err != nil {
		return nil, 0, gerror.Wrap(err, "query admin gallery images")
	}
	return values, count.Int(), nil
}

func (p *PG) AdminImageCounts(ctx context.Context) (map[string]int, error) {
	type row struct {
		State string `orm:"state"`
		Count int    `orm:"count"`
	}
	var rows []row
	if err := p.db.Ctx(ctx).Raw(`SELECT publication_state AS state, COUNT(*)::int AS count FROM gallery_images WHERE review_state IN ('approved', 'not_required') GROUP BY publication_state`).Scan(&rows); err != nil {
		return nil, gerror.Wrap(err, "count gallery image lifecycle states")
	}
	counts := map[string]int{"all": 0, "draft": 0, "published": 0, "hidden": 0, "deleted": 0}
	for _, item := range rows {
		counts[item.State] = item.Count
		counts["all"] += item.Count
	}
	return counts, nil
}

func (p *PG) AdminImage(ctx context.Context, id string) (*model.AdminImage, error) {
	record, err := p.db.GetOne(ctx, adminImageSelect+` WHERE i.id = ?::uuid AND i.review_state IN ('approved', 'not_required')`, id)
	if err != nil {
		return nil, gerror.Wrap(err, "query admin gallery image")
	}
	value, err := recordAs[model.AdminImage](record)
	if err != nil || value == nil {
		return value, err
	}
	if err := p.db.Ctx(ctx).Raw(`
SELECT facet_value.facet_id::text AS facet_id, assignment.facet_value_id::text AS value_id
FROM gallery_image_facet_assignments assignment
JOIN gallery_facet_values facet_value ON facet_value.id = assignment.facet_value_id
WHERE assignment.image_id = ?::uuid
ORDER BY facet_value.facet_id, assignment.facet_value_id`, id).Scan(&value.Facets); err != nil {
		return nil, gerror.Wrap(err, "query admin gallery image facets")
	}
	if err := p.db.Ctx(ctx).Raw(`
SELECT tag.id::text AS id, tag.current_name AS name
FROM gallery_image_tag_assignments assignment
JOIN gallery_tags tag ON tag.id = assignment.tag_id
WHERE assignment.image_id = ?::uuid
ORDER BY tag.current_name`, id).Scan(&value.Tags); err != nil {
		return nil, gerror.Wrap(err, "query admin gallery image tags")
	}
	return value, nil
}

func (p *PG) SetImagePrimaryCategory(ctx context.Context, imageID, categoryID string) error {
	return p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		image, err := tx.Ctx(ctx).GetOne(`SELECT id FROM gallery_images WHERE id = ?::uuid AND publication_state <> 'deleted' FOR UPDATE`, imageID)
		if err != nil {
			return gerror.Wrap(err, "lock gallery image classification")
		}
		if len(image) == 0 {
			return galleryerr.NotFound("image", imageID)
		}
		category, err := tx.Ctx(ctx).GetValue(`SELECT EXISTS (SELECT 1 FROM gallery_categories WHERE id = ?::uuid AND status = 'active')`, categoryID)
		if err != nil {
			return gerror.Wrap(err, "validate gallery primary category")
		}
		if !category.Bool() {
			return galleryerr.NotFound("category", categoryID)
		}
		if _, err := tx.Ctx(ctx).Exec(`INSERT INTO gallery_image_category_assignments (image_id, category_id) VALUES (?::uuid, ?::uuid) ON CONFLICT DO NOTHING`, imageID, categoryID); err != nil {
			return gerror.Wrap(err, "assign gallery image category")
		}
		if _, err := tx.Ctx(ctx).Exec(`INSERT INTO gallery_image_primary_categories (image_id, category_id) VALUES (?::uuid, ?::uuid) ON CONFLICT (image_id) DO UPDATE SET category_id = EXCLUDED.category_id`, imageID, categoryID); err != nil {
			return gerror.Wrap(err, "set gallery image primary category")
		}
		_, err = tx.Ctx(ctx).Exec(`UPDATE gallery_images SET updated_at = NOW() WHERE id = ?::uuid`, imageID)
		return gerror.Wrap(err, "touch gallery image classification")
	})
}

func (p *PG) UpdateAdminImage(ctx context.Context, id string, input model.AdminImageUpdateInput) (*model.AdminImage, error) {
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		record, err := tx.GetOne(`
UPDATE gallery_images
SET title = ?, description = ?, alt_text = ?, source_url = NULLIF(?, ''), updated_at = NOW()
WHERE id = ?::uuid AND updated_at = ?::timestamptz AND publication_state <> 'deleted'
  AND review_state IN ('approved', 'not_required')
RETURNING id`, input.Title, input.Description, input.AltText, input.SourceURL, id, input.ExpectedUpdatedAt)
		if err != nil {
			return gerror.Wrap(err, "update admin gallery image")
		}
		if len(record) == 0 {
			return galleryerr.Conflict("image_version")
		}
		if input.Classification == nil {
			return nil
		}

		classificationInput := input.Classification
		category, err := tx.GetValue(`SELECT EXISTS (SELECT 1 FROM gallery_categories WHERE id = ?::uuid AND status = 'active')`, classificationInput.PrimaryCategoryID)
		if err != nil {
			return gerror.Wrap(err, "validate gallery image primary category")
		}
		if !category.Bool() {
			return galleryerr.NotFound("category", classificationInput.PrimaryCategoryID)
		}
		if _, err := tx.Ctx(ctx).Exec(`INSERT INTO gallery_image_category_assignments (image_id, category_id) VALUES (?::uuid, ?::uuid) ON CONFLICT DO NOTHING`, id, classificationInput.PrimaryCategoryID); err != nil {
			return gerror.Wrap(err, "assign gallery image primary category")
		}
		if _, err := tx.Ctx(ctx).Exec(`INSERT INTO gallery_image_primary_categories (image_id, category_id) VALUES (?::uuid, ?::uuid) ON CONFLICT (image_id) DO UPDATE SET category_id = EXCLUDED.category_id`, id, classificationInput.PrimaryCategoryID); err != nil {
			return gerror.Wrap(err, "update gallery image primary category")
		}

		facetGroups := map[string]struct{}{}
		for _, valueID := range classificationInput.FacetValueIDs {
			facetID, err := tx.GetValue(`SELECT facet_id::text FROM gallery_facet_values WHERE id = ?::uuid AND status = 'active'`, valueID)
			if err != nil {
				return gerror.Wrap(err, "validate gallery image facet value")
			}
			if facetID.String() == "" {
				return galleryerr.NotFound("facet_value", valueID)
			}
			if _, duplicate := facetGroups[facetID.String()]; duplicate {
				return galleryerr.Validation("facetValueIds", "only one value may be selected for each facet")
			}
			facetGroups[facetID.String()] = struct{}{}
		}
		if _, err := tx.Ctx(ctx).Exec(`DELETE FROM gallery_image_facet_assignments WHERE image_id = ?::uuid`, id); err != nil {
			return gerror.Wrap(err, "replace gallery image facets")
		}
		for _, valueID := range classificationInput.FacetValueIDs {
			if _, err := tx.Ctx(ctx).Exec(`INSERT INTO gallery_image_facet_assignments (image_id, facet_value_id) VALUES (?::uuid, ?::uuid)`, id, valueID); err != nil {
				return gerror.Wrap(err, "assign gallery image facet")
			}
		}

		for _, tagID := range classificationInput.TagIDs {
			tag, err := tx.GetValue(`SELECT EXISTS (SELECT 1 FROM gallery_tags WHERE id = ?::uuid AND status = 'active')`, tagID)
			if err != nil {
				return gerror.Wrap(err, "validate gallery image tag")
			}
			if !tag.Bool() {
				return galleryerr.NotFound("tag", tagID)
			}
		}
		if _, err := tx.Ctx(ctx).Exec(`DELETE FROM gallery_image_tag_assignments WHERE image_id = ?::uuid`, id); err != nil {
			return gerror.Wrap(err, "replace gallery image tags")
		}
		for _, tagID := range classificationInput.TagIDs {
			if _, err := tx.Ctx(ctx).Exec(`INSERT INTO gallery_image_tag_assignments (image_id, tag_id) VALUES (?::uuid, ?::uuid) ON CONFLICT DO NOTHING`, id, tagID); err != nil {
				return gerror.Wrap(err, "assign gallery image tag")
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return p.AdminImage(ctx, id)
}

func (p *PG) ReviewQueue(ctx context.Context, input model.AdminSubmissionQuery) ([]model.Submission, int, error) {
	where := []string{"TRUE"}
	args := []any{}
	if input.Search != "" {
		where, args = append(where, "(submission.title ILIKE ? OR submission.description ILIKE ? OR COALESCE(submission.source_url, '') ILIKE ?)"), append(args, "%"+input.Search+"%", "%"+input.Search+"%", "%"+input.Search+"%")
	}
	if input.ProcessingState != "" {
		where, args = append(where, "submission.processing_state = ?"), append(args, input.ProcessingState)
	}
	if input.ReviewState != "" {
		where, args = append(where, "submission.review_state = ?"), append(args, input.ReviewState)
	}
	if input.SafetyState != "" {
		where, args = append(where, "submission.safety_state = ?"), append(args, input.SafetyState)
	}
	if input.Outcome != "" {
		where, args = append(where, "submission.outcome = ?"), append(args, input.Outcome)
	}
	predicate := strings.Join(where, " AND ")
	count, err := p.db.GetValue(ctx, `SELECT COUNT(*) FROM gallery_submissions submission WHERE `+predicate, args...)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "count gallery submission workbench")
	}
	orders := map[string]map[string]string{
		"createdAt": {"asc": "submission.created_at ASC, submission.id ASC", "desc": "submission.created_at DESC, submission.id DESC"},
		"updatedAt": {"asc": "submission.updated_at ASC, submission.id ASC", "desc": "submission.updated_at DESC, submission.id DESC"},
		"title":     {"asc": "LOWER(submission.title) ASC, submission.id ASC", "desc": "LOWER(submission.title) DESC, submission.id DESC"},
	}
	query := `
SELECT submission.id::text AS id, submission.subject_kind, submission.subject_id, submission.asset_id::text AS asset_id,
       COALESCE(submission.image_id::text, '') AS image_id, submission.title, submission.description,
       COALESCE(submission.source_url, '') AS source_url, submission.alt_text,
       COALESCE(primary_category.category_id::text, '') AS primary_category_id,
       submission.processing_state, submission.review_state, submission.safety_state,
       submission.outcome, submission.failure_code, submission.review_note, submission.created_at, submission.updated_at
FROM gallery_submissions submission
LEFT JOIN gallery_submission_primary_categories primary_category ON primary_category.submission_id = submission.id
WHERE ` + predicate + `
ORDER BY ` + orders[input.SortBy][input.SortOrder] + ` LIMIT ? OFFSET ?`
	pageArgs := append(append([]any{}, args...), input.PageSize, (input.Page-1)*input.PageSize)
	var values []model.Submission
	if err := p.db.Ctx(ctx).Raw(query, pageArgs...).Scan(&values); err != nil {
		return nil, 0, gerror.Wrap(err, "query gallery submission workbench")
	}
	return values, count.Int(), nil
}

func (p *PG) ReviewSubmission(ctx context.Context, operator, id string, input model.SubmissionReviewInput) (*model.Submission, error) {
	var value *model.Submission
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		catalog, err := tx.GetOne(`
SELECT id
FROM gallery_classification_catalogs
WHERE catalog_key = 'gallery'
FOR SHARE`)
		if err != nil {
			return gerror.Wrap(err, "lock gallery classification catalog for review")
		}
		if len(catalog) == 0 {
			return galleryerr.NotInitialized("classification_catalog")
		}
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
	   alt_text, COALESCE((SELECT category_id::text FROM gallery_submission_primary_categories
	     WHERE submission_id = gallery_submissions.id), '') AS primary_category_id,
	   processing_state, review_state, safety_state,
       outcome, failure_code, review_note, created_at, updated_at, width, height, dominant_color, public_rendition_ready
FROM gallery_submissions WHERE id = ?::uuid FOR UPDATE`, id).Scan(&current); err != nil {
			return gerror.Wrap(err, "lock gallery submission review")
		}
		if current == nil {
			return nil
		}
		if current.ReviewState != "pending" || current.Outcome != "pending" {
			if input.Decision == "approve" && current.ReviewState == "approved" && current.Outcome == "published" && current.ImageID != "" {
				value = &current.Submission
			}
			return nil
		}
		if input.Decision == "reject" {
			record, err := tx.GetOne(`
UPDATE gallery_submissions
SET review_state = 'rejected', outcome = 'rejected', review_note = ?, reviewed_by = ?, reviewed_at = NOW(), updated_at = NOW()
WHERE id = ?::uuid
RETURNING id::text AS id, subject_kind, subject_id, asset_id::text AS asset_id,
          COALESCE(image_id::text, '') AS image_id, title, description, COALESCE(source_url, '') AS source_url,
		  alt_text, COALESCE((SELECT category_id::text FROM gallery_submission_primary_categories
		    WHERE submission_id = gallery_submissions.id), '') AS primary_category_id,
		  processing_state, review_state, safety_state,
          outcome, failure_code, review_note, created_at, updated_at`, input.Note, operator, id)
			if err != nil {
				return gerror.Wrap(err, "reject gallery submission")
			}
			value, err = recordAs[model.Submission](record)
			if err != nil {
				return err
			}
			return p.publishSubmissionReviewedTx(ctx, tx, value, input.Decision)
		}
		if current.ProcessingState != "ready" || !current.PublicRenditionReady || current.Width <= 0 || current.Height <= 0 || !oneOf(current.SafetyState, "safe", "uncertain") {
			return nil
		}
		if _, err := tx.Ctx(ctx).Exec(`
SELECT pg_advisory_xact_lock(hashtextextended(encode(exact_sha256, 'hex'), 0))
FROM gallery_submissions WHERE id = ?::uuid`, id); err != nil {
			return gerror.Wrap(err, "lock gallery exact duplicate hash")
		}
		duplicateValue, err := tx.GetValue(`
SELECT public_image.id::text
FROM gallery_submissions source
JOIN gallery_images public_image ON public_image.exact_sha256 = source.exact_sha256
WHERE source.id = ?::uuid AND source.exact_sha256 IS NOT NULL
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
    id, asset_id, origin_submission_id, title, description, source_url, alt_text, width, height,
    dominant_color, processing_state, review_state, publication_state, safety_state,
    public_rendition_ready, exact_sha256, pdq_hash, published_at
)
SELECT ?::uuid, asset_id, id, title, description, source_url, alt_text, width, height,
       dominant_color, 'ready', 'approved', 'published', 'safe', FALSE, exact_sha256, pdq_hash, NOW()
FROM gallery_submissions WHERE id = ?::uuid
RETURNING id::text AS id`, newIdentifier(), id)
			if err != nil {
				return gerror.Wrap(err, "publish approved gallery submission")
			}
			imageID = record["id"].String()
			outcome = "published"
			if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_image_category_assignments (image_id, category_id)
SELECT ?::uuid, category_id
FROM gallery_submission_category_assignments
WHERE submission_id = ?::uuid`, imageID, id); err != nil {
				return gerror.Wrap(err, "copy approved gallery categories")
			}
			if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_image_primary_categories (image_id, category_id)
SELECT ?::uuid, category_id
FROM gallery_submission_primary_categories
WHERE submission_id = ?::uuid`, imageID, id); err != nil {
				return gerror.Wrap(err, "copy approved gallery primary category")
			}
			if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_image_facet_assignments (image_id, facet_value_id)
SELECT ?::uuid, facet_value_id
FROM gallery_submission_facet_assignments
WHERE submission_id = ?::uuid`, imageID, id); err != nil {
				return gerror.Wrap(err, "copy approved gallery facets")
			}
			if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_image_tag_assignments (image_id, tag_id)
SELECT ?::uuid, tag_id
FROM gallery_submission_tag_assignments
WHERE submission_id = ?::uuid`, imageID, id); err != nil {
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
		  alt_text, COALESCE((SELECT category_id::text FROM gallery_submission_primary_categories
		    WHERE submission_id = gallery_submissions.id), '') AS primary_category_id,
		  processing_state, review_state, safety_state,
          outcome, failure_code, review_note, created_at, updated_at`, outcome, imageID, input.Note, operator, id)
		if err != nil {
			return gerror.Wrap(err, "complete gallery submission review")
		}
		value, err = recordAs[model.Submission](record)
		if err != nil {
			return err
		}
		return p.publishSubmissionReviewedTx(ctx, tx, value, input.Decision)
	})
	return value, err
}

func (p *PG) MarkImagePublicRenditionReady(ctx context.Context, imageID string) error {
	result, err := p.db.Exec(ctx, `
UPDATE gallery_images
SET public_rendition_ready = TRUE, updated_at = NOW()
WHERE id = ?::uuid AND publication_state = 'published'`, imageID)
	if err != nil {
		return gerror.Wrap(err, "mark gallery image public rendition ready")
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return gerror.Wrap(err, "read gallery rendition readiness result")
	}
	if rows == 0 {
		return galleryerr.InvalidState("image", "not_published")
	}
	return nil
}

func (p *PG) PublishedImageCandidates(ctx context.Context, afterID string, limit int) ([]model.ImagePublicationCandidate, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	args := []any{}
	predicate := "publication_state = 'published' AND safety_state = 'safe'"
	if afterID != "" {
		predicate += " AND id > ?::uuid"
		args = append(args, afterID)
	}
	args = append(args, limit)
	var values []model.ImagePublicationCandidate
	if err := p.db.Ctx(ctx).Raw(`
SELECT id::text AS id, asset_id::text AS asset_id, title
FROM gallery_images
WHERE `+predicate+`
ORDER BY id
LIMIT ?`, args...).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "list gallery image publication candidates")
	}
	return values, nil
}

func (p *PG) HideImage(ctx context.Context, operator, id, reason string) error {
	result, err := p.db.Exec(ctx, `
WITH hidden AS (
    UPDATE gallery_images
    SET publication_state = 'hidden', hidden_at = NOW(), updated_at = NOW()
    WHERE id = ?::uuid AND publication_state = 'published'
    RETURNING id
)
INSERT INTO gallery_cases (id, image_id, kind, status, reporter_kind, reporter_id, reason, operator_sub)
SELECT ?::uuid, id, 'takedown', 'resolved', 'operator', ?, ?, ? FROM hidden`,
		id, newIdentifier(), operator, reason, operator)
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

func (p *PG) Cases(ctx context.Context, input model.AdminCaseQuery) ([]model.Case, int, error) {
	where := []string{"TRUE"}
	args := []any{}
	if input.Status != "all" {
		where, args = append(where, "status = ?"), append(args, input.Status)
	}
	if input.Kind != "" {
		where, args = append(where, "kind = ?"), append(args, input.Kind)
	}
	if input.Search != "" {
		where, args = append(where, "(reason ILIKE ? OR description ILIKE ? OR resolution_note ILIKE ?)"), append(args, "%"+input.Search+"%", "%"+input.Search+"%", "%"+input.Search+"%")
	}
	predicate := strings.Join(where, " AND ")
	count, err := p.db.GetValue(ctx, `SELECT COUNT(*) FROM gallery_cases WHERE `+predicate, args...)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "count gallery cases")
	}
	orders := map[string]map[string]string{
		"createdAt": {"asc": "created_at ASC, id ASC", "desc": "created_at DESC, id DESC"},
		"updatedAt": {"asc": "updated_at ASC, id ASC", "desc": "updated_at DESC, id DESC"},
		"kind":      {"asc": "kind ASC, id ASC", "desc": "kind DESC, id DESC"},
		"status":    {"asc": "status ASC, id ASC", "desc": "status DESC, id DESC"},
	}
	query := `
SELECT id::text AS id, COALESCE(image_id::text, '') AS image_id,
       COALESCE(submission_id::text, '') AS submission_id, kind, status, reason, description,
       COALESCE(proposed_source_url, '') AS proposed_source_url, resolution_note, created_at, updated_at
FROM gallery_cases WHERE ` + predicate + `
ORDER BY ` + orders[input.SortBy][input.SortOrder] + ` LIMIT ? OFFSET ?`
	pageArgs := append(append([]any{}, args...), input.PageSize, (input.Page-1)*input.PageSize)
	var values []model.Case
	if err := p.db.Ctx(ctx).Raw(query, pageArgs...).Scan(&values); err != nil {
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
WHERE id = ?::uuid AND status IN ('open', 'reviewing') AND updated_at = ?::timestamptz
RETURNING id::text AS id, COALESCE(image_id::text, '') AS image_id,
          COALESCE(submission_id::text, '') AS submission_id, kind, status, reason, description,
          COALESCE(proposed_source_url, '') AS proposed_source_url, resolution_note, created_at, updated_at`
	record, err := p.db.GetOne(ctx, query, input.Status, operator, input.Note, id, input.ExpectedUpdatedAt)
	if err != nil {
		return nil, gerror.Wrap(err, "resolve gallery case")
	}
	value, err := recordAs[model.Case](record)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.Conflict("case_version")
	}
	return value, nil
}

func (p *PG) RecordEvent(ctx context.Context, subject model.Subject, imageID string, input model.EventInput) error {
	now := time.Now().UTC()
	metricDate := now.Format(time.DateOnly)
	dayStart, dayEnd := now.Truncate(24*time.Hour), now.Truncate(24*time.Hour).Add(24*time.Hour)
	subjectKey := strings.TrimSpace(subject.ID)
	if kind := strings.TrimSpace(subject.Kind); kind != "" {
		subjectKey = kind + ":" + subjectKey
	}
	identity := subjectKey
	if input.SessionKey != "" {
		identity = "session:" + input.SessionKey
	}
	return p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		uniqueVisitor := int64(0)
		if input.Type == "qualified_view" && identity != "" {
			if _, err := tx.Ctx(ctx).Exec(`SELECT pg_advisory_xact_lock(hashtextextended(?::text, 0))`, "gallery-visitor:"+metricDate+":"+imageID+":"+identity); err != nil {
				return gerror.Wrap(err, "lock gallery daily visitor")
			}
			seen, err := tx.GetValue(`
SELECT EXISTS (
    SELECT 1
    FROM gallery_image_events
    WHERE image_id = ?::uuid
      AND event_type = 'qualified_view'
      AND occurred_at >= ?::timestamptz
      AND occurred_at < ?::timestamptz
      AND subject_key = ?
      AND session_key = ?
)`, imageID, dayStart, dayEnd, subjectKey, input.SessionKey)
			if err != nil {
				return gerror.Wrap(err, "query gallery daily visitor")
			}
			if !seen.Bool() {
				uniqueVisitor = 1
			}
		}
		result, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_image_events (id, image_id, subject_key, session_key, event_type)
SELECT ?::uuid, i.id, ?, ?, ? FROM gallery_images i WHERE i.id = ?::uuid AND `+eligibleImage,
			newIdentifier(), subjectKey, input.SessionKey, input.Type, imageID)
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
		delta := map[string][6]int64{
			"grid_exposure":  {1, 0, 0, 0, 0, 0},
			"qualified_view": {0, 1, uniqueVisitor, 0, 0, 0},
			"favorite":       {0, 0, 0, 1, 0, 0},
			"unfavorite":     {0, 0, 0, -1, 0, 0},
			"share":          {0, 0, 0, 0, 1, 0},
			"report":         {0, 0, 0, 0, 0, 1},
		}[input.Type]
		if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO gallery_image_metrics_daily (
    image_id, metric_date, exposures, qualified_views, unique_visitors, favorites, shares, reports
)
VALUES (?::uuid, ?::date, ?, ?, ?, ?, ?, ?)
ON CONFLICT (image_id, metric_date) DO UPDATE SET
    exposures = gallery_image_metrics_daily.exposures + EXCLUDED.exposures,
    qualified_views = gallery_image_metrics_daily.qualified_views + EXCLUDED.qualified_views,
    unique_visitors = gallery_image_metrics_daily.unique_visitors + EXCLUDED.unique_visitors,
    favorites = GREATEST(0, gallery_image_metrics_daily.favorites + EXCLUDED.favorites),
    shares = gallery_image_metrics_daily.shares + EXCLUDED.shares,
    reports = gallery_image_metrics_daily.reports + EXCLUDED.reports`,
			imageID, metricDate, delta[0], delta[1], delta[2], delta[3], delta[4], delta[5]); err != nil {
			return gerror.Wrap(err, "aggregate gallery daily metric")
		}
		return nil
	})
}

func (p *PG) FindSingleton(ctx context.Context, kind, ownerID string) (*model.Collection, error) {
	return p.collection(ctx, `c.kind = ? AND c.owner_id = ?`, kind, ownerID)
}

func (p *PG) CreateCollection(ctx context.Context, input collection.CreateInput) (*model.Collection, error) {
	const query = `
INSERT INTO gallery_collections (id, kind, resource_kind, owner_kind, owner_id, visibility, name, description)
VALUES (?::uuid, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (owner_id) WHERE kind = 'gallery.favorites'
DO UPDATE SET updated_at = gallery_collections.updated_at
RETURNING id::text AS id, kind, resource_kind, owner_kind, owner_id, visibility, name, description, version, created_at, updated_at`
	record, err := p.db.GetOne(ctx, query, newIdentifier(), input.Kind, input.ResourceKind, input.OwnerKind, input.OwnerID, input.Visibility, input.Name, input.Description)
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
			if _, err := tx.Ctx(ctx).Exec(`INSERT INTO gallery_collection_members (collection_id, image_id) VALUES (?::uuid, ?::uuid) ON CONFLICT DO NOTHING`, id, imageID); err != nil {
				return gerror.Wrap(err, "add gallery collection member")
			}
		}
		for _, imageID := range remove {
			if _, err := tx.Ctx(ctx).Exec(`DELETE FROM gallery_collection_members WHERE collection_id = ?::uuid AND image_id = ?::uuid`, id, imageID); err != nil {
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
	if err := p.db.Ctx(ctx).Raw(query, uuidArrayLiteral(ids)).Scan(&rows); err != nil {
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

func (p *PG) FavoritesDetail(ctx context.Context, id string, page, size int, order string) (*model.CollectionDetail, error) {
	value, err := p.collection(ctx, `c.id = ?::uuid`, id)
	if err != nil || value == nil {
		return nil, err
	}
	count, err := p.db.GetValue(ctx, `
SELECT COUNT(*) FROM gallery_collection_members member
JOIN gallery_images i ON i.id = member.image_id
WHERE member.collection_id = ?::uuid AND `+eligibleImage, id)
	if err != nil {
		return nil, gerror.Wrap(err, "count eligible favorite images")
	}
	orders := map[string]string{
		"newest": "member.added_at DESC, i.id DESC", "oldest": "member.added_at ASC, i.id ASC",
		"title_asc": "LOWER(i.title) ASC, i.id ASC", "title_desc": "LOWER(i.title) DESC, i.id DESC",
	}
	query := imageCardSelect + `
JOIN gallery_collection_members member ON member.image_id = i.id
WHERE member.collection_id = ?::uuid AND ` + eligibleImage + `
ORDER BY ` + orders[order] + ` LIMIT ? OFFSET ?`
	images, err := p.cards(ctx, query, id, size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	value.ItemCount = count.Int()
	return &model.CollectionDetail{Collection: *value, Images: images}, nil
}

func (p *PG) collection(ctx context.Context, where string, args ...any) (*model.Collection, error) {
	query := `
SELECT c.id::text AS id, c.kind, c.resource_kind, c.owner_kind, c.owner_id, c.visibility,
       c.name, c.description, c.version, c.created_at, c.updated_at,
       COALESCE(e.slug, '') AS slug, COALESCE(e.cover_image_id::text, '') AS cover_image_id,
       COALESCE(e.seo_title, '') AS seo_title, COALESCE(e.seo_description, '') AS seo_description,
       COALESCE(cover_image.asset_id::text, '') AS cover_asset_id,
       COALESCE(cover_image.alt_text, '') AS cover_alt_text,
       COALESCE(cover_image.width, 0) AS cover_width, COALESCE(cover_image.height, 0) AS cover_height,
       COALESCE(cover_image.dominant_color, '') AS cover_color,
       (SELECT COUNT(*) FROM gallery_collection_members member WHERE member.collection_id = c.id)::int AS item_count
FROM gallery_collections c
LEFT JOIN gallery_collection_editorial e ON e.collection_id = c.id
LEFT JOIN gallery_images cover_image ON cover_image.id = e.cover_image_id
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
