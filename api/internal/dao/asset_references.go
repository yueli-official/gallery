package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/yueli-official/asset/referencesync"
	"github.com/yueli-official/gallery/api/internal/assetreferences"
	"github.com/yueli-official/gallery/api/internal/model"
	"time"
)

func (p *PG) WithAssetReferences(module *referencesync.Module, origins ...string) (*PG, error) {
	source, err := assetreferences.ObjectSource(origins...)
	if err != nil {
		return nil, err
	}
	if module == nil {
		return nil, fmt.Errorf("gallery reference module is required")
	}
	module.Source = source
	p.references = module
	p.mediaOrigins = append([]string(nil), origins...)
	return p, nil
}
func (p *PG) createReferences(ctx context.Context, sets []referencesync.ObjectSet, action func(context.Context, gdb.TX) error) error {
	if p.references == nil {
		return fmt.Errorf("gallery reference module is required for writes")
	}
	return p.references.Transaction(ctx, "service:gallery-reference-writer", sets, func(ctx context.Context, enqueue func(*sql.Tx) error) error {
		return p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			if err := action(ctx, tx); err != nil {
				return err
			}
			return enqueue(tx.GetSqlTX())
		})
	})
}
func (p *PG) changeReferences(ctx context.Context, keys []referencesync.OwnerKey, project func([]referencesync.ObjectSet) ([]referencesync.ObjectSet, error), action func(context.Context, gdb.TX) error) error {
	if p.references == nil {
		return fmt.Errorf("gallery reference module is required for writes")
	}
	return p.references.Change(ctx, "service:gallery-reference-writer", keys, project, func(ctx context.Context, verify, enqueue func(*sql.Tx) error) error {
		return p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			if err := verify(tx.GetSqlTX()); err != nil {
				return err
			}
			if err := action(ctx, tx); err != nil {
				return err
			}
			return enqueue(tx.GetSqlTX())
		})
	})
}
func (p *PG) referenceSet(kind, id, label, route, asset string) referencesync.ObjectSet {
	return referencesync.ObjectSet{RefType: kind, RefID: id, Version: 1, Label: label, URL: referencesync.ReferenceURL(route, p.mediaOrigins...), Assets: (referencesync.ObjectFacts{AssetIDs: []string{asset}}).Assets(p.mediaOrigins...)}
}
func unchangedReferences(sets []referencesync.ObjectSet) ([]referencesync.ObjectSet, error) {
	return sets, nil
}
func (p *PG) clearReferencesTx(ctx context.Context, tx gdb.TX, kind, id string) error {
	if p.references == nil {
		return fmt.Errorf("gallery reference module is required for writes")
	}
	_, err := p.references.SaveTx(ctx, tx.GetSqlTX(), p.referenceSet(kind, id, "", "", ""))
	return err
}
func (p *PG) SynchronizeImagePublication(ctx context.Context, item model.ImagePublicationCandidate, port assetreferences.PublicationPort) error {
	db, err := p.db.Master()
	if err != nil {
		return err
	}
	return assetreferences.SynchronizePublication(ctx, db, port, item.ID)
}

type reviewReferencePlan struct {
	sets                            []referencesync.ObjectSet
	updatedAt                       time.Time
	imageID, imageAsset, imageTitle string
	newImage                        bool
}

func (p *PG) reviewReferences(ctx context.Context, id string, plan reviewReferencePlan, action func(context.Context, gdb.TX) error) error {
	if plan.imageID == "" {
		return p.changeReferences(ctx, []referencesync.OwnerKey{{RefType: "gallery-submission-image", RefID: id}}, unchangedReferences, action)
	}
	return p.createReferences(ctx, plan.sets, action)
}

func (p *PG) planReviewReferences(ctx context.Context, id, decision string) (reviewReferencePlan, error) {
	plan := reviewReferencePlan{}
	row, err := p.db.GetOne(ctx, `SELECT s.asset_id::text,s.title,s.updated_at,s.outcome,s.review_state,s.processing_state,s.public_rendition_ready,s.width,s.height,s.safety_state,
 COALESCE((SELECT i.id::text FROM gallery_images i WHERE i.exact_sha256=s.exact_sha256 LIMIT 1),'') AS duplicate_id
 FROM gallery_submissions s WHERE s.id=?::uuid`, id)
	if err != nil {
		return plan, err
	}
	if len(row) == 0 {
		return plan, referencesync.ErrOwnerNotFound
	}
	plan.updatedAt = row["updated_at"].Time()
	asset := row["asset_id"].String()
	if row["outcome"].String() == "withdrawn" {
		asset = ""
	}
	plan.sets = []referencesync.ObjectSet{p.referenceSet("gallery-submission-image", id, row["title"].String(), "/submissions/"+id, asset)}
	if decision != "approve" || row["review_state"].String() != "pending" || row["outcome"].String() != "pending" || row["processing_state"].String() != "ready" || !row["public_rendition_ready"].Bool() || row["width"].Int() <= 0 || row["height"].Int() <= 0 || !oneOf(row["safety_state"].String(), "safe", "uncertain") {
		return plan, nil
	}
	plan.imageID = row["duplicate_id"].String()
	plan.imageAsset = asset
	plan.imageTitle = row["title"].String()
	if plan.imageID == "" {
		plan.imageID = newIdentifier()
		plan.newImage = true
	} else {
		image, err := p.db.GetOne(ctx, `SELECT asset_id::text,title,deleted_at IS NOT NULL OR publication_state='deleted' AS deleted FROM gallery_images WHERE id=?::uuid`, plan.imageID)
		if err != nil {
			return plan, err
		}
		if len(image) == 0 || image["deleted"].Bool() {
			return plan, referencesync.ErrOwnerChanged
		}
		plan.imageAsset = image["asset_id"].String()
		plan.imageTitle = image["title"].String()
	}
	plan.sets = append(plan.sets, p.referenceSet("gallery-image", plan.imageID, plan.imageTitle, "/images/"+plan.imageID, plan.imageAsset))
	return plan, nil
}

func (p *PG) visualReference(ctx context.Context, kind, id, label string, visual *model.ClassificationPresentation) (referencesync.ObjectSet, error) {
	asset := ""
	if visual == nil {
		row, err := p.db.GetValue(ctx, `SELECT COALESCE(asset_id::text,'') FROM gallery_classification_presentation WHERE kind=? AND node_id=?::uuid`, kind, id)
		if err != nil {
			return referencesync.ObjectSet{}, err
		}
		asset = row.String()
	} else {
		asset = visual.AssetID
		if visual.ImageID != "" {
			row, err := p.db.GetValue(ctx, `SELECT asset_id::text FROM gallery_images WHERE id=?::uuid`, visual.ImageID)
			if err != nil {
				return referencesync.ObjectSet{}, err
			}
			if row.String() == "" {
				return referencesync.ObjectSet{}, referencesync.ErrOwnerNotFound
			}
			asset = row.String()
		}
	}
	return p.referenceSet("gallery-classification-cover", kind+":"+id, label, "/manage/classification", asset), nil
}
func (p *PG) classificationReferences(ctx context.Context, kind, id, label string, visual *model.ClassificationPresentation, action func(context.Context, gdb.TX) error) error {
	// All cover writes share the catalog fence, including settings replacement.
	revision, err := p.ClassificationRevision(ctx)
	if err != nil {
		return err
	}
	set, err := p.visualReference(ctx, kind, id, label, visual)
	if err != nil {
		return err
	}
	return p.catalogReferences(ctx, revision, []referencesync.ObjectSet{set}, action)
}
func (p *PG) catalogReferences(ctx context.Context, revision uint64, sets []referencesync.ObjectSet, action func(context.Context, gdb.TX) error) error {
	return p.createReferences(ctx, sets, func(ctx context.Context, tx gdb.TX) error {
		current, err := tx.GetValue(`SELECT revision FROM gallery_classification_catalogs WHERE catalog_key='gallery' FOR UPDATE`)
		if err != nil {
			return err
		}
		if current.Uint64() != revision {
			return referencesync.ErrOwnerChanged
		}
		return action(ctx, tx)
	})
}
func (p *PG) siteReferences(ctx context.Context, input model.SiteSettingsUpdateInput, action func(context.Context, gdb.TX) error) error {
	if input.Appearance == nil {
		return p.changeReferences(ctx, []referencesync.OwnerKey{{RefType: "gallery-site-artwork", RefID: "site"}}, func(sets []referencesync.ObjectSet) ([]referencesync.ObjectSet, error) {
			sets[0].Label = input.Name
			return sets, nil
		}, action)
	}
	revision, err := p.ClassificationRevision(ctx)
	if err != nil {
		return err
	}
	stamp, err := p.db.GetValue(ctx, `SELECT updated_at FROM gallery_site_settings WHERE site_key='gallery'`)
	if err != nil {
		return err
	}
	sidebarAsset := ""
	if input.Appearance.Sidebar.ImageID != "" {
		asset, err := p.db.GetValue(ctx, `SELECT asset_id::text FROM gallery_images WHERE id=?::uuid`, input.Appearance.Sidebar.ImageID)
		if err != nil {
			return err
		}
		if asset.String() == "" {
			return referencesync.ErrOwnerNotFound
		}
		sidebarAsset = asset.String()
	}
	sets := []referencesync.ObjectSet{p.referenceSet("gallery-site-artwork", "site", input.Name, "/manage/site", sidebarAsset)}
	if input.Appearance.Classification != nil {
		previous, err := p.db.GetAll(ctx, `SELECT kind,node_id::text FROM gallery_classification_presentation`)
		if err != nil {
			return err
		}
		replacements := map[string]referencesync.ObjectSet{}
		for _, row := range previous {
			key := row["kind"].String() + ":" + row["node_id"].String()
			replacements[key] = p.referenceSet("gallery-classification-cover", key, "", "/manage/classification", "")
		}
		for _, visual := range input.Appearance.Classification {
			table := map[string]string{"category": "gallery_categories", "facet": "gallery_facets", "facet_value": "gallery_facet_values", "tag": "gallery_tags"}[visual.Kind]
			if table == "" {
				return referencesync.ErrInvalidQuery
			}
			nameColumn := "name"
			if visual.Kind == "tag" {
				nameColumn = "current_name"
			}
			label, err := p.db.GetValue(ctx, `SELECT `+nameColumn+` FROM `+table+` WHERE id=?::uuid`, visual.NodeID)
			if err != nil {
				return err
			}
			set, err := p.visualReference(ctx, visual.Kind, visual.NodeID, label.String(), &visual)
			if err != nil {
				return err
			}
			replacements[set.RefID] = set
		}
		for _, set := range replacements {
			sets = append(sets, set)
		}
	}
	return p.catalogReferences(ctx, revision, sets, func(ctx context.Context, tx gdb.TX) error {
		current, err := tx.GetValue(`SELECT updated_at FROM gallery_site_settings WHERE site_key='gallery' FOR UPDATE`)
		if err != nil {
			return err
		}
		if !current.Time().Equal(stamp.Time()) {
			return referencesync.ErrOwnerChanged
		}
		if err = action(ctx, tx); err != nil {
			return err
		}
		if input.Appearance.Classification != nil {
			_, err = tx.Exec(`UPDATE gallery_classification_catalogs SET revision=revision+1,updated_at=NOW() WHERE catalog_key='gallery'`)
		}
		return err
	})
}
