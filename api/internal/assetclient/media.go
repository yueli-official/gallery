package assetclient

import (
	"context"
	"github.com/yueli-official/asset/delivery"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/model"
)

var renditionKeys = []string{"thumbnail", "grid-sm", "grid-lg", "masonry-sm", "masonry-lg", "preview", "display", "og"}

func (c *HTTP) ServiceToken(ctx context.Context) (string, error) { return c.tokens.Token(ctx) }

func (c *HTTP) AuthorizeMedia(ctx context.Context, assetIDs []string) (map[string]*model.MediaDelivery, error) {
	client := delivery.Client{BaseURL: c.baseURL, Token: c.ServiceToken, HTTP: c.http}
	result := map[string]*model.MediaDelivery{}
	// One request authorizes all eight slots for up to 24 distinct assets.
	for offset := 0; offset < len(assetIDs); offset += 24 {
		group := assetIDs[offset:min(offset+24, len(assetIDs))]
		selections := make([]delivery.Selection, 0, len(group)*8)
		for _, id := range group {
			for _, preset := range renditionKeys {
				item, err := delivery.SelectAsset(id, preset, "webp", 20261025)
				if err != nil {
					return nil, galleryerr.UpstreamFailed("asset.invalid_media_authorization")
				}
				selections = append(selections, item)
			}
		}
		grants, err := client.Authorize(ctx, selections)
		if err != nil {
			return nil, galleryerr.UpstreamFailed("asset.media_authorization_unavailable")
		}
		for index, id := range group {
			value := &model.MediaDelivery{Mode: "signed", Renditions: map[string]string{}}
			for slot, preset := range renditionKeys {
				grant := grants[index*8+slot]
				value.Renditions[preset] = grant.URL
				value.ExpiresAt = grant.ExpiresAt
			}
			result[id] = value
		}
	}
	return result, nil
}
