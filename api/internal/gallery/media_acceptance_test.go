package gallery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yueli-official/gallery/api/internal/model"
)

type mediaAcceptancePort struct {
	calls [][]string
	err   error
}

func (p *mediaAcceptancePort) AuthorizeMedia(_ context.Context, ids []string) (map[string]*model.MediaDelivery, error) {
	p.calls = append(p.calls, append([]string(nil), ids...))
	if p.err != nil {
		return nil, p.err
	}
	result := make(map[string]*model.MediaDelivery)
	for _, id := range ids {
		result[id] = &model.MediaDelivery{Mode: "signed", ExpiresAt: time.Now().Add(5 * time.Minute).UTC().Format(time.RFC3339), Renditions: map[string]string{"display": "https://asset.example/protected/" + id}}
	}
	return result, nil
}

// The product service must project authorizations through its detail, related
// and collection flows, and fail the entire response when authorization fails.
func TestSignedMediaAcrossProductResponses(t *testing.T) {
	ctx := context.Background()
	store := &fakeStore{
		image: &model.ImageDetail{ImageCard: model.ImageCard{ID: testCategoryID, AssetID: "asset-a"}},
		relatedImages: []model.RelatedImage{
			{ImageCard: model.ImageCard{ID: testFacetID, AssetID: "asset-b"}},
			{ImageCard: model.ImageCard{ID: testValueID, AssetID: "asset-b"}},
		},
		collectionDetail: &model.CollectionDetail{
			Collection: model.Collection{ID: testCategoryID, Kind: "gallery.editorial", OwnerID: "gallery", CoverAssetID: "asset-c"},
			Images:     []model.ImageCard{{ID: testFacetID, AssetID: "asset-b"}},
		},
	}
	service := New(store)
	port := &mediaAcceptancePort{}
	service.SetMediaAuthorizationPort(port)
	detail, err := service.Image(ctx, PublicID(testCategoryID), "")
	if err != nil || detail.Delivery == nil {
		t.Fatalf("detail: %v", err)
	}
	related, err := service.RelatedImages(ctx, PublicID(testCategoryID), 12)
	if err != nil || len(related) != 2 || related[0].Delivery == nil || related[1].Delivery == nil {
		t.Fatalf("related: %v", err)
	}
	if len(port.calls) != 2 || len(port.calls[1]) != 1 {
		t.Fatalf("related assets were not deduplicated: %v", port.calls)
	}
	collection, err := service.AdminCollection(ctx, PublicID(testCategoryID), 1, 12)
	if err != nil || collection.CoverDelivery == nil || collection.Images[0].Delivery == nil {
		t.Fatalf("collection: %v", err)
	}
	port.err = errors.New("signer unavailable")
	if value, err := service.Image(ctx, PublicID(testCategoryID), ""); err == nil || value != nil {
		t.Fatal("detail succeeded without a valid media authorization")
	}
	if value, err := service.RelatedImages(ctx, PublicID(testCategoryID), 12); err == nil || value != nil {
		t.Fatal("related response succeeded without a valid media authorization")
	}
	if value, err := service.AdminCollection(ctx, PublicID(testCategoryID), 1, 12); err == nil || value != nil {
		t.Fatal("collection succeeded without a valid media authorization")
	}
}
