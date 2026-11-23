// Command import-pln performs an explicit, server-operator migration. It uses
// Gallery's classification governance and submission services; Asset ingestion
// is performed separately through the registered service upload API.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/yueli-official/asset/referencesync"
	"github.com/yueli-official/gallery/api/internal/assetreferences"
	"os"
	"strings"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/yueli-official/foundation/go/work"
	workpostgres "github.com/yueli-official/foundation/go/work/postgres"
	"github.com/yueli-official/gallery/api/internal/appconfig"
	"github.com/yueli-official/gallery/api/internal/assetclient"
	"github.com/yueli-official/gallery/api/internal/dao"
	galleryservice "github.com/yueli-official/gallery/api/internal/gallery"
	"github.com/yueli-official/gallery/api/internal/model"
	"github.com/yueli-official/gallery/api/internal/runtime"
)

const operator = "migration:pln:20261025"

type importedAsset struct {
	SourceID   int    `json:"sourceId"`
	Filename   string `json:"filename"`
	SHA256     string `json:"sha256"`
	AssetID    string `json:"assetId"`
	OldAssetID string `json:"oldAssetId,omitempty"`
}

func main() {
	action := flag.String("action", "inspect", "inspect, catalog, submit, or publish")
	manifest := flag.String("manifest", "", "JSON array of finalized Asset imports")
	flag.Parse()
	if err := run(*action, *manifest); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(action, manifest string) error {
	if err := runtime.EnableEnvironmentConfig(); err != nil {
		return err
	}
	ctx := context.Background()
	db, err := runtime.OpenDefaultPostgres(ctx)
	if err != nil {
		return err
	}
	defer db.Close()
	// One migration operator at a time, across repeated invocations and hosts.
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	var locked bool
	if err = conn.QueryRowContext(ctx, `SELECT pg_try_advisory_lock(hashtext('gallery:import-pln'))`).Scan(&locked); err != nil {
		return err
	}
	if !locked {
		return errors.New("PLN migration is already running")
	}
	defer conn.ExecContext(ctx, `SELECT pg_advisory_unlock(hashtext('gallery:import-pln'))`)
	catalog, err := work.Compile(galleryservice.WorkDefinition())
	if err != nil {
		return err
	}
	adapter, err := workpostgres.New(ctx, catalog, workpostgres.Options{DB: db, InstanceKey: "gallery:" + appconfig.SiteSlug(ctx)})
	if err != nil {
		return err
	}
	module, _, err := assetreferences.New(ctx, db)
	if err != nil {
		return err
	}
	store, err := dao.NewPG(g.DB(), adapter).WithAssetReferences(module, os.Getenv("ASSET_PUBLIC_ORIGIN"))
	if err != nil {
		return err
	}
	service := galleryservice.New(store)
	switch action {
	case "inspect":
		current, err := service.ClassificationCatalog(ctx)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(current)
	case "catalog":
		return configureCatalog(ctx, db, service)
	case "submit":
		return submit(ctx, db, service, manifest)
	case "publish":
		return publish(ctx, db, store, service, manifest)
	case "rebind":
		return rebind(ctx, db, module, service, manifest)
	default:
		return fmt.Errorf("unknown action %q", action)
	}
}

// Rebind the task's existing entries to newly encoded assets. New identities
// keep previously immutable delivery URLs unchanged. Only this operator's rows
// and hash-verified WebP assets are admitted.
func rebind(ctx context.Context, db *sql.DB, module *referencesync.Module, s *galleryservice.Service, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var inputs []importedAsset
	if err = json.Unmarshal(raw, &inputs); err != nil {
		return err
	}
	if len(inputs) == 0 || len(inputs) > 1000 {
		return errors.New("manifest must contain 1–1000 images")
	}
	assets, err := assetPort(ctx)
	if err != nil {
		return err
	}
	var outputs []map[string]any
	for _, input := range inputs {
		if input.OldAssetID == "" || input.OldAssetID == input.AssetID {
			return errors.New("explicit distinct previous asset is required")
		}
		facts, err := assets.PrepareSubmission(ctx, input.AssetID)
		if err != nil {
			return err
		}
		if facts.ContentHash != input.SHA256 || facts.Mime != "image/webp" || facts.Width > 4096 || facts.Height > 4096 {
			return errors.New("normalized media facts mismatch")
		}
		var id, imageID, title, current string
		err = db.QueryRowContext(ctx, `SELECT id::text,COALESCE(image_id::text,''),title,asset_id::text FROM gallery_submissions WHERE subject_kind='operator' AND subject_id=$1 AND asset_id IN ($2::uuid,$3::uuid)`, operator, input.OldAssetID, input.AssetID).Scan(&id, &imageID, &title, &current)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return err
		}
		keys := []referencesync.OwnerKey{{RefType: "gallery-submission-image", RefID: id}}
		if imageID != "" {
			keys = append(keys, referencesync.OwnerKey{RefType: "gallery-image", RefID: imageID})
		}
		err = module.Change(ctx, operator, keys, func(sets []referencesync.ObjectSet) ([]referencesync.ObjectSet, error) {
			for i := range sets {
				if len(sets[i].Assets) == 0 {
					return nil, errors.New("cannot rebind a withdrawn or deleted image")
				}
				sets[i].Assets = []referencesync.ObjectAsset{{AssetID: input.AssetID}}
			}
			return sets, nil
		}, func(ctx context.Context, verify, enqueue func(*sql.Tx) error) error {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return err
			}
			defer tx.Rollback()
			if err = verify(tx); err != nil {
				return err
			}
			result, err := tx.ExecContext(ctx, `UPDATE gallery_submissions SET asset_id=$1::uuid,exact_sha256=decode($2,'hex'),width=$3,height=$4,pdq_hash=NULL,updated_at=NOW() WHERE id=$5::uuid AND subject_kind='operator' AND subject_id=$6 AND asset_id=$7::uuid`, input.AssetID, input.SHA256, facts.Width, facts.Height, id, operator, current)
			if err != nil {
				return err
			}
			affected, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if affected != 1 {
				return referencesync.ErrOwnerChanged
			}
			if imageID != "" {
				result, err = tx.ExecContext(ctx, `UPDATE gallery_images SET asset_id=$1::uuid,exact_sha256=decode($2,'hex'),width=$3,height=$4,pdq_hash=NULL,public_rendition_ready=FALSE,updated_at=NOW() WHERE id=$5::uuid AND origin_submission_id=$6::uuid AND asset_id=$7::uuid AND deleted_at IS NULL AND publication_state<>'deleted'`, input.AssetID, input.SHA256, facts.Width, facts.Height, imageID, id, current)
				if err != nil {
					return err
				}
				affected, err = result.RowsAffected()
				if err != nil {
					return err
				}
				if affected != 1 {
					return referencesync.ErrOwnerChanged
				}
				if current != input.AssetID {
					if _, err = tx.ExecContext(ctx, `INSERT INTO gallery_publication_retirements(image_id,asset_id) VALUES($1::uuid,$2::uuid) ON CONFLICT DO NOTHING`, imageID, current); err != nil {
						return err
					}
				}
			}
			if err = enqueue(tx); err != nil {
				return err
			}
			return tx.Commit()
		})
		if err != nil {
			return err
		}
		outputs = append(outputs, map[string]any{"sourceId": input.SourceID, "assetId": input.AssetID, "submissionId": galleryservice.PublicID(id), "imageId": galleryservice.PublicID(imageID)})
	}
	for _, output := range outputs {
		if err = json.NewEncoder(os.Stdout).Encode(output); err != nil {
			return err
		}
	}
	return nil
}

func govern(ctx context.Context, s *galleryservice.Service, command model.ClassificationGovernanceCommand) error {
	preview, err := s.PreviewClassificationGovernance(ctx, model.ClassificationGovernancePreviewInput{Command: command})
	if err != nil {
		return err
	}
	if preview.Outcome != "planned" {
		data, _ := json.Marshal(preview.Diagnostics)
		return fmt.Errorf("classification governance rejected: %s", data)
	}
	for _, step := range preview.Plan.Steps {
		if (step.Kind == "delete_assignments" || step.Kind == "clear_primary_assignments") && step.AffectedCount > 0 {
			return errors.New("refusing to remove a classification used by existing content")
		}
	}
	_, err = s.ExecuteClassificationGovernance(ctx, operator, model.ClassificationGovernanceExecuteInput{Command: command, ExpectedCatalogRevision: preview.Plan.ExpectedCatalogRevision, ExpectedRequestToken: preview.Plan.ExpectedRequestToken, ExpectedImpactToken: preview.Plan.ExpectedImpactToken})
	return err
}

func ensure(ctx context.Context, s *galleryservice.Service, kind, slug, name, facetID, parentID string) (string, error) {
	current, err := s.ClassificationCatalog(ctx)
	if err != nil {
		return "", err
	}
	var id, status string
	if kind == "category" {
		for _, node := range current.Categories {
			if node.Slug == slug {
				id, status = node.ID, node.Status
			}
		}
	}
	if kind == "facet" {
		for _, node := range current.Facets {
			if node.Slug == slug {
				id, status = node.ID, node.Status
			}
		}
	}
	if kind == "facet_value" {
		for _, facet := range current.Facets {
			if galleryservice.PublicID(facet.ID) == galleryservice.PublicID(facetID) {
				for _, node := range facet.Values {
					if node.Slug == slug {
						if galleryservice.PublicID(node.ParentID) != galleryservice.PublicID(parentID) {
							return "", fmt.Errorf("%s has a different owning work", slug)
						}
						id, status = node.ID, node.Status
					}
				}
			}
		}
	}
	if id == "" {
		result, err := s.CreateClassificationIdentity(ctx, operator, model.ClassificationIdentityCreateInput{Kind: kind, Slug: slug, Name: name, FacetID: facetID, ParentID: parentID})
		if err != nil {
			return "", err
		}
		id, status = result.ID, "draft"
	}
	if status == "draft" {
		if err := govern(ctx, s, model.ClassificationGovernanceCommand{Operation: "set_status", Kind: kind, ID: id, Status: "active"}); err != nil {
			return "", err
		}
	} else if status != "active" {
		return "", fmt.Errorf("%s %s exists with status %s", kind, slug, status)
	}
	return galleryservice.PublicID(id), nil
}

func configureCatalog(ctx context.Context, db *sql.DB, s *galleryservice.Service) error {
	before, err := s.ClassificationCatalog(ctx)
	if err != nil {
		return err
	}
	obsoleteCategories := map[string]bool{"wallpaper": true, "illustration": true, "photography": true}
	obsoleteFacets := map[string]bool{"scene": true, "orientation": true, "style": true, "purpose": true}
	hasObsolete := false
	for _, v := range before.Categories {
		hasObsolete = hasObsolete || obsoleteCategories[v.Slug]
	}
	for _, v := range before.Facets {
		hasObsolete = hasObsolete || obsoleteFacets[v.Slug]
	}
	if hasObsolete {
		var occupied bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gallery_images) OR EXISTS(SELECT 1 FROM gallery_submissions)`).Scan(&occupied); err != nil {
			return err
		}
		if occupied {
			return errors.New("catalog cleanup requires the verified empty destination; existing content must be classified explicitly")
		}
	}
	for _, v := range [][2]string{{"game", "游戏"}, {"anime", "动漫"}, {"original", "原创"}} {
		if _, err := ensure(ctx, s, "category", v[0], v[1], "", ""); err != nil {
			return err
		}
	}
	workID, err := ensure(ctx, s, "facet", "work", "作品与角色", "", "")
	if err != nil {
		return err
	}
	workValueID, err := ensure(ctx, s, "facet_value", "blue-archive", "碧蓝档案", workID, "")
	if err != nil {
		return err
	}
	if _, err = ensure(ctx, s, "facet_value", "plana-blue-archive", "普拉娜", workID, workValueID); err != nil {
		return err
	}
	for _, v := range before.Categories {
		if obsoleteCategories[v.Slug] {
			if err := govern(ctx, s, model.ClassificationGovernanceCommand{Operation: "delete", Kind: "category", ID: v.ID, DeleteAllRelated: true}); err != nil {
				return err
			}
		}
	}
	for _, v := range before.Facets {
		if obsoleteFacets[v.Slug] {
			if err := govern(ctx, s, model.ClassificationGovernanceCommand{Operation: "delete", Kind: "facet", ID: v.ID, DeleteAllRelated: true}); err != nil {
				return err
			}
		}
	}
	current, err := s.ClassificationCatalog(ctx)
	if err != nil {
		return err
	}
	settings, err := s.SiteSettings(ctx)
	if err != nil {
		return err
	}
	for _, node := range current.Categories {
		icon := map[string]string{"game": "device-gamepad-2", "anime": "book", "original": "palette"}[node.Slug]
		if icon == "" {
			continue
		}
		found := false
		for _, rule := range settings.Appearance.Classification {
			if rule.Kind == "category" && galleryservice.PublicID(rule.NodeID) == galleryservice.PublicID(node.ID) {
				found = true
			}
		}
		if !found {
			settings.Appearance.Classification = append(settings.Appearance.Classification, model.ClassificationPresentation{Kind: "category", NodeID: node.ID, Mode: "icon", Icon: icon})
		}
	}
	_, err = s.UpdateSiteSettings(ctx, model.SiteSettingsUpdateInput{Appearance: &settings.Appearance, Name: settings.Name, Title: settings.Title, Description: settings.Description, SearchPlaceholder: settings.SearchPlaceholder, FooterTagline: settings.FooterTagline, RandomCandidateSize: settings.RandomCandidateSize, HomeSections: settings.HomeSections})
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(current)
}

func submit(ctx context.Context, db *sql.DB, s *galleryservice.Service, path string) error {
	if path == "" {
		return errors.New("an explicit finalized-asset manifest is required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var inputs []importedAsset
	if err = json.Unmarshal(raw, &inputs); err != nil {
		return err
	}
	if len(inputs) == 0 || len(inputs) > 1000 {
		return errors.New("manifest must contain 1–1000 images")
	}
	current, err := s.ClassificationCatalog(ctx)
	if err != nil {
		return err
	}
	var categoryID, workID, workValueID, characterValueID string
	for _, c := range current.Categories {
		if c.Slug == "game" && c.Status == "active" {
			categoryID = c.ID
		}
	}
	for _, f := range current.Facets {
		if f.Slug == "work" && f.Status == "active" {
			workID = f.ID
			for _, v := range f.Values {
				if v.Slug == "blue-archive" && v.Status == "active" {
					workValueID = v.ID
				}
				if v.Slug == "plana-blue-archive" && v.Status == "active" {
					characterValueID = v.ID
				}
			}
		}
	}
	if categoryID == "" || workID == "" || workValueID == "" || characterValueID == "" {
		return errors.New("configure the game / Blue Archive / Plana catalog before importing")
	}
	assets, err := assetPort(ctx)
	if err != nil {
		return err
	}
	for _, input := range inputs {
		if input.SourceID <= 0 || len(input.SHA256) != 64 || strings.Contains(strings.ToLower(input.Filename), "preview") || strings.Contains(strings.ToLower(input.Filename), "thumbnail") {
			return errors.New("invalid original image manifest entry")
		}
		var id string
		err = db.QueryRowContext(ctx, `SELECT id::text FROM gallery_submissions WHERE subject_kind='operator' AND subject_id=$1 AND asset_id=$2::uuid ORDER BY created_at LIMIT 1`, operator, input.AssetID).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			callCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			facts, factsErr := assets.PrepareSubmission(callCtx, input.AssetID)
			cancel()
			if factsErr != nil {
				return fmt.Errorf("PLN %d media validation: %w", input.SourceID, factsErr)
			}
			if facts.ContentHash != input.SHA256 {
				return fmt.Errorf("PLN %d original digest mismatch", input.SourceID)
			}
			value, err := s.Submit(ctx, model.Subject{Kind: "operator", ID: operator}, model.SubmissionInput{AssetID: input.AssetID, Title: fmt.Sprintf("普拉娜 · %03d", input.SourceID), Description: fmt.Sprintf("从普拉娜旧图库迁入（PLN %d）。", input.SourceID), AltText: "碧蓝档案中的普拉娜", CategoryIDs: []string{categoryID}, PrimaryCategoryID: categoryID, Facets: []model.FacetSelection{{FacetID: workID, ValueIDs: []string{characterValueID}}}})
			if err != nil {
				return fmt.Errorf("PLN %d submission: %w", input.SourceID, err)
			}
			id = value.ID
		} else if err != nil {
			return err
		}
		if err = json.NewEncoder(os.Stdout).Encode(map[string]any{"sourceId": input.SourceID, "assetId": input.AssetID, "submissionId": id}); err != nil {
			return err
		}
	}
	return nil
}

func assetPort(ctx context.Context) (*assetclient.HTTP, error) {
	cfg := appconfig.LoadAssetClient(ctx)
	return assetclient.NewHTTP(assetclient.Config{BaseURL: cfg.BaseURL, TokenURL: cfg.TokenURL, ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret, Scope: cfg.Scope, SiteKey: appconfig.AssetNamespace(ctx)})
}

// Publishing is a separate, explicit operator action after processing. It
// approves the authorized source collection through the normal review path.
func publish(ctx context.Context, db *sql.DB, store *dao.PG, s *galleryservice.Service, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var inputs []importedAsset
	if err = json.Unmarshal(raw, &inputs); err != nil {
		return err
	}
	if len(inputs) == 0 || len(inputs) > 1000 {
		return errors.New("manifest must contain 1–1000 images")
	}
	assets, err := assetPort(ctx)
	if err != nil {
		return err
	}
	s.SetAssetReferencePort(assets)
	for _, input := range inputs {
		var id, imageID, outcome, processing, title string
		err = db.QueryRowContext(ctx, `SELECT id::text,COALESCE(image_id::text,''),outcome,processing_state,title FROM gallery_submissions WHERE subject_kind='operator' AND subject_id=$1 AND asset_id=$2::uuid ORDER BY created_at LIMIT 1`, operator, input.AssetID).Scan(&id, &imageID, &outcome, &processing, &title)
		if err != nil {
			return fmt.Errorf("PLN %d submission lookup: %w", input.SourceID, err)
		}
		if outcome != "published" {
			if processing != "ready" || outcome != "pending" {
				return fmt.Errorf("PLN %d is not ready for migration approval (%s/%s)", input.SourceID, processing, outcome)
			}
			value, err := s.ReviewSubmission(ctx, operator, id, model.SubmissionReviewInput{Decision: "approve", Note: "用户授权将 PLN 既有有效原图迁入图库。"})
			if err != nil {
				return fmt.Errorf("PLN %d approval: %w", input.SourceID, err)
			}
			if value.Outcome != "published" {
				return fmt.Errorf("PLN %d unexpected outcome %s", input.SourceID, value.Outcome)
			}
			imageID = value.ImageID
		} else {
			if err := store.SynchronizeImagePublication(ctx, model.ImagePublicationCandidate{ID: imageID}, assets); err != nil {
				return err
			}
		}
		if err = json.NewEncoder(os.Stdout).Encode(map[string]any{"sourceId": input.SourceID, "submissionId": id, "imageId": galleryservice.PublicID(imageID)}); err != nil {
			return err
		}
	}
	return nil
}
