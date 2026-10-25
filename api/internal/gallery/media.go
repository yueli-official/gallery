package gallery

import (
	"context"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/model"
)

type MediaAuthorizationPort interface {
	AuthorizeMedia(context.Context, []string) (map[string]*model.MediaDelivery, error)
}

func (s *Service) SetMediaAuthorizationPort(port MediaAuthorizationPort) { s.media = port }

func (s *Service) SignedMediaDelivery() bool { return s.media != nil }

type mediaTarget struct {
	assetID string
	output  **model.MediaDelivery
}

// Public delivery and signed delivery are explicit deployment modes. A signing
// failure in signed mode fails the projection; it never emits an unsigned URL.
func (s *Service) prepareMedia(ctx context.Context, targets []mediaTarget) error {
	if s.media == nil {
		return nil
	}
	ids := []string{}
	seen := map[string]bool{}
	for _, target := range targets {
		if target.assetID != "" && !seen[target.assetID] {
			ids = append(ids, target.assetID)
			seen[target.assetID] = true
		}
	}
	if len(ids) == 0 {
		return nil
	}
	grants, err := s.media.AuthorizeMedia(ctx, ids)
	if err != nil {
		return err
	}
	for _, target := range targets {
		if target.assetID == "" {
			continue
		}
		value := grants[target.assetID]
		if value == nil || value.Mode != "signed" {
			return galleryerr.UpstreamFailed("asset.invalid_media_authorization")
		}
		*target.output = value
	}
	return nil
}

func (s *Service) prepareCards(ctx context.Context, cards []model.ImageCard) error {
	targets := make([]mediaTarget, 0, len(cards))
	for i := range cards {
		targets = append(targets, mediaTarget{cards[i].AssetID, &cards[i].Delivery})
	}
	return s.prepareMedia(ctx, targets)
}

func (s *Service) prepareCollections(ctx context.Context, values []model.Collection) error {
	targets := make([]mediaTarget, 0, len(values))
	for i := range values {
		targets = append(targets, mediaTarget{values[i].CoverAssetID, &values[i].CoverDelivery})
	}
	return s.prepareMedia(ctx, targets)
}

func (s *Service) prepareAppearance(ctx context.Context, value *model.GalleryAppearance) error {
	targets := []mediaTarget{{value.Sidebar.AssetID, &value.Sidebar.Delivery}}
	for i := range value.Classification {
		item := &value.Classification[i]
		targets = append(targets, mediaTarget{item.AssetID, &item.Delivery})
	}
	return s.prepareMedia(ctx, targets)
}
