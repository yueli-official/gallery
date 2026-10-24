package gallery

import (
	"context"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/model"
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
	seen := map[string]bool{}
	for i := range value.Classification {
		item := &value.Classification[i]
		item.NodeID, err = DatabaseID(item.NodeID)
		if err != nil {
			return galleryerr.Validation("appearance/classification", "invalid classification identity")
		}
		key := item.Kind + ":" + item.NodeID
		if !identities[key] || seen[key] || !oneOf(item.Mode, "icon", "image") {
			return galleryerr.Validation("appearance/classification", "invalid or duplicate presentation")
		}
		seen[key] = true
		if !oneOf(item.Icon, "photo", "device-desktop", "camera", "palette", "user", "mountain", "building", "device-gamepad-2", "book", "sparkles", "category", "device-mobile", "leaf", "heart", "tag", "square") {
			return galleryerr.Validation("appearance/classification", "unsupported icon")
		}
		if item.Mode == "image" {
			item.ImageID, err = DatabaseID(item.ImageID)
			if err != nil {
				return galleryerr.Validation("appearance/classification", "select a published image")
			}
		} else {
			item.ImageID = ""
		}
	}
	return nil
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
