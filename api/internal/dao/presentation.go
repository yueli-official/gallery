package dao

import (
	"context"
	"encoding/json"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/model"
)

func (p *PG) readAppearance(ctx context.Context) (model.GalleryAppearance, error) {
	value := model.GalleryAppearance{Classification: []model.ClassificationPresentation{}}
	err := p.db.Ctx(ctx).Raw(`SELECT s.sidebar_mode, COALESCE(s.sidebar_image_id::text, '') AS sidebar_image_id,
        s.sidebar_overlay, COALESCE(i.asset_id::text, '') AS asset_id
        FROM gallery_site_settings s LEFT JOIN gallery_images i ON i.id=s.sidebar_image_id WHERE s.site_key='gallery'`).Scan(&value.Sidebar)
	if err != nil {
		return value, err
	}
	err = p.db.Ctx(ctx).Raw(`SELECT p.kind, p.node_id::text, p.mode, p.icon,
        COALESCE(p.image_id::text, '') AS image_id, COALESCE(i.asset_id::text, '') AS asset_id
        FROM gallery_classification_presentation p LEFT JOIN gallery_images i ON i.id=p.image_id
        ORDER BY p.kind,p.node_id`).Scan(&value.Classification)
	return value, err
}

func writeAppearance(ctx context.Context, tx gdb.TX, appearance model.GalleryAppearance) error {
	// Lock chosen images while checking eligibility so a concurrent takedown cannot
	// leave a newly configured public background pointing at a hidden image.
	imageIDs := []string{appearance.Sidebar.ImageID}
	for _, item := range appearance.Classification {
		imageIDs = append(imageIDs, item.ImageID)
	}
	for _, id := range imageIDs {
		if id == "" {
			continue
		}
		row, err := tx.GetOne(`SELECT i.id FROM gallery_images i WHERE i.id=?::uuid AND `+eligibleImage+` FOR SHARE`, id)
		if err != nil {
			return err
		}
		if len(row) == 0 {
			return galleryerr.Validation("appearance", "select an eligible published image")
		}
	}
	if _, err := tx.Exec(`UPDATE gallery_site_settings SET sidebar_mode=?, sidebar_image_id=NULLIF(?, '')::uuid,
        sidebar_overlay=? WHERE site_key='gallery'`, appearance.Sidebar.Mode, appearance.Sidebar.ImageID, appearance.Sidebar.Overlay); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM gallery_classification_presentation`); err != nil {
		return err
	}
	for _, item := range appearance.Classification {
		if _, err := tx.Exec(`INSERT INTO gallery_classification_presentation(kind,node_id,mode,icon,image_id)
            VALUES(?,?::uuid,?,?,NULLIF(?, '')::uuid)`, item.Kind, item.NodeID, item.Mode, item.Icon, item.ImageID); err != nil {
			return err
		}
	}
	return nil
}

const presentationImageUnused = `NOT EXISTS (SELECT 1 FROM gallery_site_settings s WHERE s.sidebar_image_id=gallery_images.id)
    AND NOT EXISTS (SELECT 1 FROM gallery_classification_presentation p WHERE p.image_id=gallery_images.id)`

func lockUnusedPresentationImage(tx gdb.TX, id string) error {
	row, err := tx.GetOne(`SELECT id FROM gallery_images WHERE id=?::uuid FOR UPDATE`, id)
	if err != nil {
		return err
	}
	if len(row) == 0 {
		return galleryerr.NotFound("image", id)
	}
	unused, err := tx.GetValue(`SELECT `+presentationImageUnused+` FROM gallery_images WHERE id=?::uuid`, id)
	if err != nil {
		return err
	}
	if !unused.Bool() {
		return galleryerr.ImagePresentationInUse()
	}
	return nil
}

func (p *PG) SavedFilters(ctx context.Context, owner string) ([]model.SavedFilter, error) {
	rows, err := p.db.GetAll(ctx, `SELECT id::text,name,query::text FROM gallery_saved_filters WHERE owner_id=? ORDER BY created_at DESC,id LIMIT 100`, owner)
	if err != nil {
		return nil, err
	}
	items := make([]model.SavedFilter, 0, len(rows))
	for _, row := range rows {
		item := model.SavedFilter{ID: row["id"].String(), Name: row["name"].String()}
		if err := json.Unmarshal([]byte(row["query"].String()), &item.Query); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (p *PG) SaveFilter(ctx context.Context, owner string, input model.SavedFilterInput) (*model.SavedFilter, error) {
	raw, err := json.Marshal(input.Query)
	if err != nil {
		return nil, err
	}
	var item *model.SavedFilter
	err = p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtext(?))`, "gallery:saved-filters:"+owner); err != nil {
			return err
		}
		count, err := tx.GetValue(`SELECT COUNT(*) FROM gallery_saved_filters WHERE owner_id=?`, owner)
		if err != nil {
			return err
		}
		if count.Int() >= 100 {
			return galleryerr.Validation("name", "at most 100 saved filters")
		}
		existing, err := tx.GetValue(`SELECT EXISTS(SELECT 1 FROM gallery_saved_filters WHERE owner_id=? AND name=?)`, owner, input.Name)
		if err != nil {
			return err
		}
		if existing.Bool() {
			return galleryerr.Conflict("saved_filter_name")
		}
		id := newIdentifier()
		if _, err := tx.Exec(`INSERT INTO gallery_saved_filters(id,owner_id,name,query) VALUES(?::uuid,?,?,?::jsonb)`, id, owner, input.Name, string(raw)); err != nil {
			return err
		}
		item = &model.SavedFilter{ID: id, Name: input.Name, Query: input.Query}
		return nil
	})
	return item, err
}

func (p *PG) DeleteSavedFilter(ctx context.Context, owner, id string) error {
	result, err := p.db.Exec(ctx, `DELETE FROM gallery_saved_filters WHERE id=?::uuid AND owner_id=?`, id, owner)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return galleryerr.NotFound("saved_filter", id)
	}
	return nil
}
