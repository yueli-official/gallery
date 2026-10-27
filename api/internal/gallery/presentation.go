package gallery

import (
	"context"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/model"
	"regexp"
	"strings"
)

func normalizeAppearance(value *model.GalleryAppearance) {
	value.Sidebar.ImageID = PublicID(value.Sidebar.ImageID)
	if value.Classification == nil {
		value.Classification = []model.ClassificationPresentation{}
	}
	for i := range value.Classification {
		item := &value.Classification[i]
		item.NodeID = PublicID(item.NodeID)
		item.ImageID = PublicID(item.ImageID)
	}
}

func (s *Service) validateAppearance(ctx context.Context, value *model.GalleryAppearance) error {
	value.Sidebar.Delivery = nil
	for i := range value.Classification {
		value.Classification[i].Delivery = nil
	}
	if !oneOf(value.Sidebar.Mode, "plain", "image") || value.Sidebar.Overlay < 0.55 || value.Sidebar.Overlay > 0.95 {
		return galleryerr.Validation("appearance/sidebar", "invalid navigation background")
	}
	var err error
	if value.Sidebar.Mode == "image" {
		value.Sidebar.ImageID, err = DatabaseID(value.Sidebar.ImageID)
		if err != nil {
			return galleryerr.Validation("appearance/sidebar/imageId", "select a published image")
		}
	} else {
		value.Sidebar.ImageID = ""
	}
	if len(value.Classification) > 2000 {
		return galleryerr.Validation("appearance/classification", "too many presentation entries")
	}
	snapshot, err := s.store.ClassificationSnapshot(ctx)
	if err != nil {
		return err
	}
	identities := map[string]bool{}
	for _, item := range snapshot.Categories {
		identities["category:"+string(item.ID)] = true
	}
	for _, item := range snapshot.FacetValues {
		identities["facet_value:"+string(item.ID)] = true
	}
	for _, item := range snapshot.Facets {
		identities["facet:"+string(item.ID)] = true
	}
	seen := map[string]bool{}
	for i := range value.Classification {
		item := &value.Classification[i]
		item.NodeID, err = DatabaseID(item.NodeID)
		if err != nil {
			return galleryerr.Validation("appearance/classification", "invalid classification identity")
		}
		key := item.Kind + ":" + item.NodeID
		if (!identities[key] && item.Kind != "tag") || seen[key] || !oneOf(item.Mode, "icon", "image") {
			return galleryerr.Validation("appearance/classification", "invalid or duplicate presentation")
		}
		seen[key] = true
		if err := s.validateClassificationVisual(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

var tablerIconName = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (s *Service) validateClassificationVisual(ctx context.Context, value *model.ClassificationPresentation) error {
	if value == nil {
		return nil
	}
	value.Delivery = nil
	value.Icon = strings.TrimPrefix(strings.TrimSpace(value.Icon), "i-tabler-")
	if len(value.Icon) > 100 || !tablerIconName.MatchString(value.Icon) || !oneOf(value.Mode, "icon", "image") {
		return galleryerr.Validation("visual", "select a valid Tabler icon and display mode")
	}
	var err error
	if value.ImageID != "" {
		value.ImageID, err = DatabaseID(value.ImageID)
		if err != nil {
			return galleryerr.Validation("visual/imageId", "select a published image")
		}
	}
	if value.AssetID != "" {
		value.AssetID, err = DatabaseID(value.AssetID)
		if err != nil {
			return galleryerr.Validation("visual/assetId", "select a valid image")
		}
		if s.assets == nil || s.media == nil {
			return galleryerr.NotInitialized("asset")
		}
		// The Asset authorization boundary checks the caller's namespace/profile.
		if _, err := s.media.AuthorizeMedia(ctx, []string{value.AssetID}); err != nil {
			return err
		}
		facts, err := s.assets.PrepareSubmission(ctx, value.AssetID)
		if err != nil {
			return err
		}
		if value.ImageID == "" && (facts.Width != 512 || facts.Height != 512 || facts.Mime != "image/webp") {
			return galleryerr.Validation("visual/assetId", "cover must be 512 × 512 pixels in WebP format")
		}
	}
	if value.Mode == "image" && value.AssetID == "" && value.ImageID == "" {
		return galleryerr.Validation("visual/assetId", "select a cover image")
	}
	return nil
}

func (s *Service) classificationVisuals(ctx context.Context) (map[string]*model.ClassificationPresentation, error) {
	settings, err := s.store.SiteSettings(ctx)
	if err != nil {
		return nil, err
	}
	result := map[string]*model.ClassificationPresentation{}
	if settings == nil {
		return result, nil
	}
	normalizeAppearance(&settings.Appearance)
	if err := s.prepareAppearance(ctx, &settings.Appearance); err != nil {
		return nil, err
	}
	for i := range settings.Appearance.Classification {
		item := &settings.Appearance.Classification[i]
		result[item.Kind+":"+item.NodeID] = item
	}
	return result, nil
}

func (s *Service) SavedFilters(ctx context.Context, owner string) ([]model.SavedFilter, error) {
	items, err := s.store.SavedFilters(ctx, owner)
	for i := range items {
		items[i].ID = PublicID(items[i].ID)
	}
	return items, err
}

func (s *Service) SaveFilter(ctx context.Context, owner string, input model.SavedFilterInput) (*model.SavedFilter, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len([]rune(input.Name)) > 60 {
		return nil, galleryerr.Validation("name", "enter a name up to 60 characters")
	}
	q := input.Query
	if len(q.Search)+len(q.Categories)+len(q.Facets)+len(q.Tag) > 8192 {
		return nil, galleryerr.Validation("query", "filter is too large")
	}
	// Reuse the real public query validation; stale/invalid references cannot be
	// saved as apparently executable filters.
	_, err := s.Images(ctx, model.ImageQuery{Search: q.Search, CategoryRefs: csvRefs(q.Categories), FacetRefs: csvRefs(q.Facets), Tag: q.Tag, MatchAll: q.MatchAll, Orientation: q.Orientation, MinLongEdge: q.MinLongEdge})
	if err != nil {
		return nil, err
	}
	item, err := s.store.SaveFilter(ctx, owner, input)
	if err == nil {
		item.ID = PublicID(item.ID)
	}
	return item, err
}

func csvRefs(raw string) []string {
	refs := []string{}
	for _, value := range strings.Split(raw, ",") {
		if value = strings.TrimSpace(value); value != "" {
			refs = append(refs, value)
		}
	}
	return refs
}

func (s *Service) DeleteSavedFilter(ctx context.Context, owner, rawID string) error {
	id, err := DatabaseID(rawID)
	if err != nil {
		return galleryerr.NotFound("saved_filter", rawID)
	}
	return s.store.DeleteSavedFilter(ctx, owner, id)
}

func (s *Service) DownloadImage(ctx context.Context, rawID string) (*model.ImageDownload, error) {
	image, err := s.Image(ctx, rawID, "")
	if err != nil {
		return nil, err
	}
	if s.assets == nil {
		return nil, galleryerr.NotInitialized("asset")
	}
	// Eligibility is checked before signing. The short-lived URL is never
	// persisted into the Gallery image or presentation configuration.
	return s.assets.SignOriginal(ctx, image.AssetID)
}
