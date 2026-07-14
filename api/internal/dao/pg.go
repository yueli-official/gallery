package dao

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"platform/gokit/facet"
	"platform/products/gallery/api/internal/galleryerr"
	"platform/products/gallery/api/internal/model"
)

type PG struct {
	db gdb.DB
}

func (p *PG) CreatorBySubject(ctx context.Context, subject string) (*model.CreatorProfile, error) {
	var value *model.CreatorProfile
	if err := p.db.Model("gallery_creator_profiles").Ctx(ctx).Where("account_sub", subject).Scan(&value); err != nil {
		return nil, gerror.Wrap(err, "query gallery creator by subject")
	}
	return value, nil
}

func (p *PG) CreateCreator(ctx context.Context, input model.CreatorRequest) (*model.CreatorProfile, error) {
	id := "creator-" + uuid.NewString()
	const query = `
INSERT INTO gallery_creator_profiles (id, account_sub, handle, display_name, application_note)
VALUES (?, ?, ?, ?, ?)
RETURNING id, account_sub, handle, display_name, bio, cover_asset_id, status,
          application_note, review_note, created_at, updated_at`
	record, err := p.db.GetOne(ctx, query, id, input.AccountSub, input.Handle, input.DisplayName, input.ApplicationNote)
	if err != nil {
		var postgresError *pq.Error
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return nil, galleryerr.Conflict("creator")
		}
		return nil, gerror.Wrap(err, "create gallery creator request")
	}
	return recordAs[model.CreatorProfile](record)
}

func (p *PG) CreatorArtworks(ctx context.Context, creatorID string) ([]model.StudioArtwork, error) {
	const query = `
SELECT a.id, a.creator_id, a.title, a.description, a.status, a.visibility,
       a.content_rating, a.ai_usage, a.ai_training_permission, a.rights_basis,
       a.license, a.review_note, a.created_at, a.updated_at, a.published_at
FROM gallery_artworks a
WHERE a.creator_id = ?
ORDER BY a.updated_at DESC, a.id DESC`
	var values []model.StudioArtwork
	if err := p.db.Ctx(ctx).Raw(query, creatorID).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query creator artworks")
	}
	for index := range values {
		if err := p.hydrateStudioArtwork(ctx, &values[index]); err != nil {
			return nil, err
		}
	}
	return values, nil
}

func (p *PG) CreateArtwork(ctx context.Context, creatorID string) (*model.StudioArtwork, error) {
	id := "artwork-" + uuid.NewString()
	const query = `
INSERT INTO gallery_artworks (id, creator_id, title)
VALUES (?, ?, '')
RETURNING id, creator_id, title, description, status, visibility, content_rating,
          ai_usage, ai_training_permission, rights_basis, license, review_note,
          created_at, updated_at, published_at`
	record, err := p.db.GetOne(ctx, query, id, creatorID)
	if err != nil {
		return nil, gerror.Wrap(err, "create gallery artwork draft")
	}
	value, err := recordAs[model.StudioArtwork](record)
	if err != nil {
		return nil, err
	}
	initializeStudioCollections(value)
	return value, nil
}

func (p *PG) CreatorArtwork(ctx context.Context, creatorID, artworkID string) (*model.StudioArtwork, error) {
	query := `
SELECT a.id, a.creator_id, a.title, a.description, a.status, a.visibility,
       a.content_rating, a.ai_usage, a.ai_training_permission, a.rights_basis,
       a.license, a.review_note, a.created_at, a.updated_at, a.published_at,
       c.id AS creator_id, c.handle AS creator_handle, c.display_name AS creator_name
FROM gallery_artworks a
JOIN gallery_creator_profiles c ON c.id = a.creator_id
WHERE a.id = ?`
	args := []any{artworkID}
	if creatorID != "" {
		query += " AND a.creator_id = ?"
		args = append(args, creatorID)
	}
	var row struct {
		model.StudioArtwork
		CreatorHandle string `orm:"creator_handle"`
		CreatorName   string `orm:"creator_name"`
	}
	if err := p.db.Ctx(ctx).Raw(query, args...).Scan(&row); err != nil {
		return nil, gerror.Wrap(err, "query creator artwork")
	}
	if row.ID == "" {
		return nil, nil
	}
	row.Creator = model.CreatorProfile{ID: row.CreatorID, Handle: row.CreatorHandle, DisplayName: row.CreatorName}
	if err := p.hydrateStudioArtwork(ctx, &row.StudioArtwork); err != nil {
		return nil, err
	}
	return &row.StudioArtwork, nil
}

func (p *PG) SaveArtwork(ctx context.Context, artworkID string, input model.ArtworkDraftInput, assignments []facet.Assignment) (*model.StudioArtwork, error) {
	saved := false
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		const update = `
UPDATE gallery_artworks
SET title = ?, description = ?, visibility = ?, content_rating = ?, ai_usage = ?,
    ai_training_permission = ?, rights_basis = ?, license = ?, updated_at = NOW()
WHERE id = ? AND status IN ('draft', 'rejected')`
		result, err := tx.Exec(update, input.Title, input.Description, input.Visibility, input.ContentRating,
			input.AIUsage, input.AITrainingPermission, input.RightsBasis, input.License, artworkID)
		if err != nil {
			return gerror.Wrap(err, "update gallery artwork draft")
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return gerror.Wrap(err, "read gallery artwork update result")
		}
		if rows == 0 {
			return nil
		}
		saved = true
		if _, err := tx.Exec("DELETE FROM gallery_artwork_facet_assignments WHERE artwork_id = ?", artworkID); err != nil {
			return gerror.Wrap(err, "clear gallery artwork facets")
		}
		for _, assignment := range assignments {
			if _, err := tx.Exec(`INSERT INTO gallery_artwork_facet_assignments (artwork_id, facet_id, value_id) VALUES (?, ?, ?)`, artworkID, assignment.FacetID, assignment.ValueID); err != nil {
				return gerror.Wrap(err, "assign gallery artwork facet")
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !saved {
		return nil, nil
	}
	return p.CreatorArtwork(ctx, "", artworkID)
}

func (p *PG) AddArtworkAsset(ctx context.Context, artworkID string, input model.AssetInput) (*model.Asset, error) {
	id := "artwork-asset-" + uuid.NewString()
	base := "/asset-api/api/v1/assets/" + input.AssetID + "/image/"
	const query = `
INSERT INTO gallery_artwork_assets (
    id, artwork_id, asset_id, sort_order, width, height, format, alt_text,
    placeholder_url, thumbnail_url, card_url, detail_url, original_url
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, asset_id, sort_order, width, height, alt_text, placeholder_url,
          thumbnail_url, card_url, detail_url, original_url`
	var value *model.Asset
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		locked, err := tx.GetOne(`SELECT id FROM gallery_artworks WHERE id = ? AND status IN ('draft', 'rejected') FOR UPDATE`, artworkID)
		if err != nil {
			return gerror.Wrap(err, "lock gallery artwork for asset link")
		}
		if len(locked) == 0 {
			return nil
		}
		position, err := tx.GetValue(`SELECT COALESCE(MAX(sort_order), -1) + 1 FROM gallery_artwork_assets WHERE artwork_id = ?`, artworkID)
		if err != nil {
			return gerror.Wrap(err, "allocate gallery artwork asset position")
		}
		record, err := tx.GetOne(query, id, artworkID, input.AssetID, position.Int(), input.Width, input.Height,
			input.Format, input.AltText,
			base+"@32x32_mode=cover_type=webp_q=30.webp",
			base+"@320x320_mode=cover_type=webp_q=78.webp",
			base+"@960x1200_mode=fit_type=webp_q=82.webp",
			base+"@1800x1800_mode=fit_type=webp_q=88.webp",
			base+"@2400x2400_mode=fit_type=webp_q=92.webp")
		if err != nil {
			return gerror.Wrap(err, "link gallery artwork asset")
		}
		value, err = recordAs[model.Asset](record)
		if err != nil {
			return err
		}
		if position.Int() == 0 {
			if _, err := tx.Exec(`UPDATE gallery_artworks SET cover_asset_id = ?, updated_at = NOW() WHERE id = ?`, input.AssetID, artworkID); err != nil {
				return gerror.Wrap(err, "set gallery artwork cover")
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (p *PG) RemoveArtworkAsset(ctx context.Context, artworkID, artworkAssetID string) (*model.Asset, error) {
	var value *model.Asset
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		locked, err := tx.GetOne(`SELECT id FROM gallery_artworks WHERE id = ? AND status IN ('draft', 'rejected') FOR UPDATE`, artworkID)
		if err != nil {
			return gerror.Wrap(err, "lock gallery artwork for asset removal")
		}
		if len(locked) == 0 {
			return nil
		}
		record, err := tx.GetOne(`DELETE FROM gallery_artwork_assets WHERE artwork_id = ? AND id = ? RETURNING id, asset_id, sort_order, width, height, alt_text, placeholder_url, thumbnail_url, card_url, detail_url, original_url`, artworkID, artworkAssetID)
		if err != nil {
			return gerror.Wrap(err, "remove gallery artwork asset")
		}
		value, err = recordAs[model.Asset](record)
		if err != nil || value == nil {
			return err
		}
		if _, err := tx.Exec(`
WITH ordered AS (
    SELECT id, ROW_NUMBER() OVER (ORDER BY sort_order, id) - 1 AS next_order
    FROM gallery_artwork_assets WHERE artwork_id = ?
)
UPDATE gallery_artwork_assets asset
SET sort_order = ordered.next_order
FROM ordered
WHERE asset.id = ordered.id`, artworkID); err != nil {
			return gerror.Wrap(err, "compact gallery artwork asset positions")
		}
		if _, err := tx.Exec(`
UPDATE gallery_artworks a
SET cover_asset_id = COALESCE((SELECT aa.asset_id FROM gallery_artwork_assets aa WHERE aa.artwork_id = a.id ORDER BY aa.sort_order LIMIT 1), ''),
    updated_at = NOW()
WHERE a.id = ?`, artworkID); err != nil {
			return gerror.Wrap(err, "refresh gallery artwork cover")
		}
		return nil
	})
	return value, err
}

func (p *PG) ArtworkAssetCount(ctx context.Context, artworkID string) (int, error) {
	count, err := p.db.Model("gallery_artwork_assets").Ctx(ctx).Where("artwork_id", artworkID).Count()
	if err != nil {
		return 0, gerror.Wrap(err, "count gallery artwork assets")
	}
	return count, nil
}

func (p *PG) ArtworkAssignments(ctx context.Context, artworkID string) ([]facet.Assignment, error) {
	var values []facet.Assignment
	if err := p.db.Model("gallery_artwork_facet_assignments").Ctx(ctx).
		Fields("facet_id", "value_id").Where("artwork_id", artworkID).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query gallery artwork facet assignments")
	}
	return values, nil
}

func (p *PG) SetArtworkStatus(ctx context.Context, artworkID, status, operator, note string) (*model.StudioArtwork, error) {
	var query string
	var args []any
	if status == "pending_review" {
		query = `
UPDATE gallery_artworks
SET status = 'pending_review', review_note = '', reviewed_by = '', reviewed_at = NULL, updated_at = NOW()
WHERE id = ? AND status IN ('draft', 'rejected')
  AND BTRIM(title) <> ''
  AND EXISTS (SELECT 1 FROM gallery_artwork_assets asset WHERE asset.artwork_id = gallery_artworks.id)
  AND NOT EXISTS (
      SELECT 1 FROM gallery_facets facet
      WHERE facet.status = 'active' AND facet.required_on_publish
        AND NOT EXISTS (
            SELECT 1 FROM gallery_artwork_facet_assignments assignment
            WHERE assignment.artwork_id = gallery_artworks.id AND assignment.facet_id = facet.id
        )
  )
RETURNING id`
		args = []any{artworkID}
	} else {
		publishedExpr := "published_at"
		if status == "published" {
			publishedExpr = "COALESCE(published_at, NOW())"
		}
		query = `
UPDATE gallery_artworks
SET status = ?, review_note = ?, reviewed_by = ?, reviewed_at = NOW(),
    published_at = ` + publishedExpr + `, updated_at = NOW()
WHERE id = ? AND status = 'pending_review'
RETURNING id`
		args = []any{status, note, operator, artworkID}
	}
	record, err := p.db.GetOne(ctx, query, args...)
	if err != nil {
		return nil, gerror.Wrap(err, "change gallery artwork status")
	}
	if len(record) == 0 {
		return nil, nil
	}
	return p.CreatorArtwork(ctx, "", artworkID)
}

func (p *PG) CreatorApplications(ctx context.Context) ([]model.CreatorProfile, error) {
	var values []model.CreatorProfile
	if err := p.db.Model("gallery_creator_profiles").Ctx(ctx).
		WhereIn("status", []string{"pending", "rejected"}).Order("created_at DESC").Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query gallery creator applications")
	}
	return values, nil
}

func (p *PG) SetCreatorStatus(ctx context.Context, creatorID, status, operator, note string) (*model.CreatorProfile, error) {
	const query = `
UPDATE gallery_creator_profiles
SET status = ?, review_note = ?, reviewed_by = ?, reviewed_at = NOW(), updated_at = NOW()
WHERE id = ? AND status IN ('pending', 'rejected')
RETURNING id, account_sub, handle, display_name, bio, cover_asset_id, status,
          application_note, review_note, created_at, updated_at`
	record, err := p.db.GetOne(ctx, query, status, note, operator, creatorID)
	if err != nil {
		return nil, gerror.Wrap(err, "review gallery creator application")
	}
	return recordAs[model.CreatorProfile](record)
}

func (p *PG) ReviewQueue(ctx context.Context, status string) ([]model.StudioArtwork, error) {
	const query = `
SELECT a.id, a.creator_id, a.title, a.description, a.status, a.visibility,
       a.content_rating, a.ai_usage, a.ai_training_permission, a.rights_basis,
       a.license, a.review_note, a.created_at, a.updated_at, a.published_at
FROM gallery_artworks a
WHERE a.status = ?
ORDER BY a.updated_at ASC, a.id ASC`
	var values []model.StudioArtwork
	if err := p.db.Ctx(ctx).Raw(query, status).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query gallery artwork review queue")
	}
	for index := range values {
		if err := p.hydrateStudioArtwork(ctx, &values[index]); err != nil {
			return nil, err
		}
	}
	return values, nil
}

func (p *PG) hydrateStudioArtwork(ctx context.Context, value *model.StudioArtwork) error {
	if value == nil {
		return nil
	}
	if value.Creator.ID == "" && value.CreatorID != "" {
		if err := p.db.Model("gallery_creator_profiles").Ctx(ctx).
			Fields("id", "handle", "display_name", "bio", "cover_asset_id", "status", "created_at", "updated_at").
			Where("id", value.CreatorID).Scan(&value.Creator); err != nil {
			return gerror.Wrap(err, "query studio artwork creator")
		}
	}
	if err := p.db.Model("gallery_artwork_assets").Ctx(ctx).
		Fields("id", "asset_id", "sort_order", "width", "height", "alt_text", "placeholder_url", "thumbnail_url", "card_url", "detail_url", "original_url").
		Where("artwork_id", value.ID).Order("sort_order ASC").Scan(&value.Assets); err != nil {
		return gerror.Wrap(err, "query studio artwork assets")
	}
	assignments, err := p.ArtworkAssignments(ctx, value.ID)
	if err != nil {
		return err
	}
	value.FacetAssignments = assignments
	initializeStudioCollections(value)
	return nil
}

func initializeStudioCollections(value *model.StudioArtwork) {
	if value.Assets == nil {
		value.Assets = []model.Asset{}
	}
	if value.FacetAssignments == nil {
		value.FacetAssignments = []facet.Assignment{}
	}
}

func recordAs[T any](record gdb.Record) (*T, error) {
	if len(record) == 0 {
		return nil, nil
	}
	var value T
	if err := record.Struct(&value); err != nil {
		return nil, gerror.Wrap(err, "map gallery database record")
	}
	return &value, nil
}

func NewPG(db gdb.DB) *PG {
	return &PG{db: db}
}

func (p *PG) SiteSettings(ctx context.Context) (*model.SiteSettings, error) {
	var value *model.SiteSettings
	if err := p.db.Model("gallery_site_settings").Ctx(ctx).Where("id", 1).Scan(&value); err != nil {
		return nil, gerror.Wrap(err, "query gallery site settings")
	}
	return value, nil
}

const artworkCardSelect = `
SELECT a.id, a.title, a.description, a.content_rating, a.ai_usage, a.published_at,
       a.license, a.rights_basis, a.ai_training_permission,
       c.id AS creator_id, c.handle AS creator_handle, c.display_name AS creator_name,
       COALESCE(media.card_url, '') AS cover_url,
       COALESCE(media.placeholder_url, '') AS placeholder_url,
       COALESCE(media.width, 0) AS width,
       COALESCE(media.height, 0) AS height
FROM gallery_artworks a
JOIN gallery_creator_profiles c ON c.id = a.creator_id
LEFT JOIN LATERAL (
    SELECT aa.card_url, aa.placeholder_url, aa.width, aa.height
    FROM gallery_artwork_assets aa
    WHERE aa.artwork_id = a.id
    ORDER BY CASE WHEN aa.asset_id = a.cover_asset_id THEN 0 ELSE 1 END, aa.sort_order
    LIMIT 1
) media ON TRUE`

func (p *PG) Featured(ctx context.Context, limit int) ([]model.ArtworkCard, error) {
	query := artworkCardSelect + `
JOIN gallery_editorial_features f ON f.artwork_id = a.id
WHERE a.status = 'published' AND a.visibility = 'public'
  AND f.placement = 'home'
  AND f.starts_at <= NOW()
  AND (f.ends_at IS NULL OR f.ends_at > NOW())
ORDER BY f.sort_order, f.created_at DESC
LIMIT ?`
	var values []model.ArtworkCard
	if err := p.db.Ctx(ctx).Raw(query, limit).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query featured gallery artworks")
	}
	hydrateCreators(values)
	return values, nil
}

func (p *PG) Latest(ctx context.Context, limit int) ([]model.ArtworkCard, error) {
	query := artworkCardSelect + `
WHERE a.status = 'published' AND a.visibility = 'public'
ORDER BY a.published_at DESC, a.id DESC
LIMIT ?`
	var values []model.ArtworkCard
	if err := p.db.Ctx(ctx).Raw(query, limit).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query latest gallery artworks")
	}
	hydrateCreators(values)
	return values, nil
}

func (p *PG) PublicCreator(ctx context.Context, handle string) (*model.PublicCreator, error) {
	var value *model.PublicCreator
	if err := p.db.Model("gallery_creator_profiles").Ctx(ctx).
		Fields("id", "handle", "display_name", "bio", "cover_asset_id", "created_at").
		Where("handle", handle).Where("status", "active").Scan(&value); err != nil {
		return nil, gerror.Wrap(err, "query public gallery creator")
	}
	return value, nil
}

func (p *PG) PublishedByCreator(ctx context.Context, creatorID string, limit int) ([]model.ArtworkCard, error) {
	query := artworkCardSelect + `
WHERE a.creator_id = ? AND a.status = 'published' AND a.visibility = 'public'
ORDER BY a.published_at DESC, a.id DESC
LIMIT ?`
	var values []model.ArtworkCard
	if err := p.db.Ctx(ctx).Raw(query, creatorID, limit).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query published gallery artworks by creator")
	}
	hydrateCreators(values)
	return values, nil
}

func (p *PG) Facets(ctx context.Context) ([]facet.Facet, error) {
	var values []facet.Facet
	if err := p.db.Model("gallery_facets").Ctx(ctx).
		Fields("id", "slug", "name", "description", "selection_mode", "required_on_publish", "filterable", "status", "sort_order").
		Order("sort_order ASC, name ASC").Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query gallery facets")
	}
	return values, nil
}

func (p *PG) FacetValues(ctx context.Context) ([]facet.Value, error) {
	const query = `
SELECT c.id, c.facet_id, COALESCE(c.parent_id, '') AS parent_id,
       c.slug, c.name, c.description, c.status, c.sort_order,
       COUNT(a.id)::int AS object_count
FROM gallery_facet_values c
JOIN gallery_facets f ON f.id = c.facet_id
LEFT JOIN gallery_artwork_facet_assignments ac ON ac.facet_id = c.facet_id AND ac.value_id = c.id
LEFT JOIN gallery_artworks a ON a.id = ac.artwork_id
  AND a.status = 'published' AND a.visibility = 'public'
GROUP BY c.id, c.facet_id, c.parent_id, c.slug, c.name, c.description, c.status, c.sort_order, f.sort_order
ORDER BY f.sort_order, c.sort_order, c.name`
	var values []facet.Value
	if err := p.db.Ctx(ctx).Raw(query).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query gallery facet values")
	}
	return values, nil
}

func (p *PG) Artwork(ctx context.Context, id string) (*model.ArtworkDetail, error) {
	query := artworkCardSelect + `
WHERE a.id = ? AND a.status = 'published' AND a.visibility = 'public'`
	var value *model.ArtworkDetail
	if err := p.db.Ctx(ctx).Raw(query, id).Scan(&value); err != nil {
		return nil, gerror.Wrap(err, "query gallery artwork")
	}
	if value == nil {
		return nil, nil
	}
	value.Creator = model.Creator{ID: value.CreatorID, Handle: value.CreatorHandle, DisplayName: value.CreatorName}
	if err := p.db.Model("gallery_artwork_assets").Ctx(ctx).
		Fields("id", "asset_id", "sort_order", "width", "height", "alt_text", "placeholder_url", "thumbnail_url", "card_url", "detail_url").
		Where("artwork_id", id).Order("sort_order ASC").Scan(&value.Assets); err != nil {
		return nil, gerror.Wrap(err, "query gallery artwork assets")
	}
	const tagQuery = `
SELECT t.name
FROM gallery_tags t
JOIN gallery_artwork_tags at ON at.tag_id = t.id
WHERE at.artwork_id = ?
ORDER BY t.name`
	if err := p.db.Ctx(ctx).Raw(tagQuery, id).Scan(&value.Tags); err != nil {
		return nil, gerror.Wrap(err, "query gallery artwork tags")
	}
	return value, nil
}

func hydrateCreators(values []model.ArtworkCard) {
	for index := range values {
		values[index].Creator = model.Creator{
			ID:          values[index].CreatorID,
			Handle:      values[index].CreatorHandle,
			DisplayName: values[index].CreatorName,
		}
	}
}
