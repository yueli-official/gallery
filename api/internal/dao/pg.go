package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"platform/products/gallery/api/internal/model"
)

type PG struct {
	db gdb.DB
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

func (p *PG) Facets(ctx context.Context) ([]model.Facet, error) {
	var values []model.Facet
	if err := p.db.Model("gallery_facets").Ctx(ctx).
		Fields("id", "slug", "name", "description", "selection_mode").
		Order("sort_order ASC, name ASC").Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query gallery facets")
	}
	return values, nil
}

func (p *PG) Categories(ctx context.Context) ([]model.Category, error) {
	const query = `
SELECT c.id, c.facet_id, COALESCE(c.parent_id, '') AS parent_id,
       c.slug, c.name, c.description, COUNT(a.id)::int AS artwork_count
FROM gallery_categories c
JOIN gallery_facets f ON f.id = c.facet_id
LEFT JOIN gallery_artwork_categories ac ON ac.category_id = c.id
LEFT JOIN gallery_artworks a ON a.id = ac.artwork_id
  AND a.status = 'published' AND a.visibility = 'public'
GROUP BY c.id, c.facet_id, c.parent_id, c.slug, c.name, c.description, c.sort_order, f.sort_order
ORDER BY f.sort_order, c.sort_order, c.name`
	var values []model.Category
	if err := p.db.Ctx(ctx).Raw(query).Scan(&values); err != nil {
		return nil, gerror.Wrap(err, "query gallery categories")
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
