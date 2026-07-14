package controller

import (
	"context"

	v1 "platform/products/gallery/api/api/v1"
	"platform/products/gallery/api/internal/assetclient"
	galleryservice "platform/products/gallery/api/internal/gallery"
	"platform/products/gallery/api/internal/model"
)

type Workflow struct {
	service *galleryservice.Service
	assets  assetclient.Client
}

func NewWorkflow(service *galleryservice.Service, assets assetclient.Client) *Workflow {
	return &Workflow{service: service, assets: assets}
}

func (c *Workflow) GetMyCreator(ctx context.Context, _ *v1.GetMyCreatorReq) (*v1.GetMyCreatorRes, error) {
	sub, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	creator, err := c.service.MyCreator(ctx, sub)
	return &v1.GetMyCreatorRes{Creator: creator}, err
}

func (c *Workflow) RequestCreator(ctx context.Context, req *v1.RequestCreatorReq) (*v1.RequestCreatorRes, error) {
	sub, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	creator, err := c.service.RequestCreator(ctx, sub, model.CreatorRequest{Handle: req.Handle, DisplayName: req.DisplayName, ApplicationNote: req.ApplicationNote})
	if err != nil {
		return nil, err
	}
	return &v1.RequestCreatorRes{Creator: *creator}, nil
}

func (c *Workflow) ListMyArtworks(ctx context.Context, _ *v1.ListMyArtworksReq) (*v1.ListMyArtworksRes, error) {
	sub, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	values, err := c.service.MyArtworks(ctx, sub)
	return &v1.ListMyArtworksRes{Artworks: values}, err
}

func (c *Workflow) CreateArtwork(ctx context.Context, _ *v1.CreateArtworkReq) (*v1.CreateArtworkRes, error) {
	sub, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.CreateDraft(ctx, sub)
	if err != nil {
		return nil, err
	}
	return &v1.CreateArtworkRes{Artwork: *value}, nil
}

func (c *Workflow) GetMyArtwork(ctx context.Context, req *v1.GetMyArtworkReq) (*v1.GetMyArtworkRes, error) {
	sub, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.MyArtwork(ctx, sub, req.ArtworkID)
	if err != nil {
		return nil, err
	}
	return &v1.GetMyArtworkRes{Artwork: *value}, nil
}

func (c *Workflow) SaveArtwork(ctx context.Context, req *v1.SaveArtworkReq) (*v1.SaveArtworkRes, error) {
	sub, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.SaveDraft(ctx, sub, req.ArtworkID, req.ArtworkDraftInput)
	if err != nil {
		return nil, err
	}
	return &v1.SaveArtworkRes{Artwork: *value}, nil
}

func (c *Workflow) AddArtworkAsset(ctx context.Context, req *v1.AddArtworkAssetReq) (*v1.AddArtworkAssetRes, error) {
	sub, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.AddAsset(ctx, sub, req.ArtworkID, req.AssetInput)
	if err != nil {
		return nil, err
	}
	if c.assets != nil {
		reference := assetclient.ReferenceInput{
			AssetID: value.AssetID, RefID: req.ArtworkID, Label: req.AltText,
			URL: "/artworks/" + req.ArtworkID,
		}
		if err := c.assets.RegisterReference(ctx, bearerOf(ctx), reference); err != nil {
			_, _ = c.service.RemoveAsset(ctx, sub, req.ArtworkID, value.ID)
			return nil, err
		}
	}
	return &v1.AddArtworkAssetRes{Asset: *value}, nil
}

func (c *Workflow) RemoveArtworkAsset(ctx context.Context, req *v1.RemoveArtworkAssetReq) (*v1.RemoveArtworkAssetRes, error) {
	sub, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.RemoveAsset(ctx, sub, req.ArtworkID, req.ArtworkAssetID)
	if err != nil {
		return nil, err
	}
	if c.assets != nil {
		_ = c.assets.UnregisterReference(ctx, bearerOf(ctx), assetclient.ReferenceInput{AssetID: value.AssetID, RefID: req.ArtworkID})
	}
	return &v1.RemoveArtworkAssetRes{Removed: true}, nil
}

func (c *Workflow) SubmitArtwork(ctx context.Context, req *v1.SubmitArtworkReq) (*v1.SubmitArtworkRes, error) {
	sub, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.SubmitArtwork(ctx, sub, req.ArtworkID)
	if err != nil {
		return nil, err
	}
	return &v1.SubmitArtworkRes{Artwork: *value}, nil
}

type Admin struct{ service *galleryservice.Service }

func NewAdmin(service *galleryservice.Service) *Admin { return &Admin{service: service} }

func (c *Admin) ListCreatorApplications(ctx context.Context, _ *v1.ListCreatorApplicationsReq) (*v1.ListCreatorApplicationsRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	values, err := c.service.CreatorApplications(ctx)
	return &v1.ListCreatorApplicationsRes{Creators: values}, err
}

func (c *Admin) ReviewCreator(ctx context.Context, req *v1.ReviewCreatorReq) (*v1.ReviewCreatorRes, error) {
	operator, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.ReviewCreator(ctx, operator, req.CreatorID, req.Decision, req.Note)
	if err != nil {
		return nil, err
	}
	return &v1.ReviewCreatorRes{Creator: *value}, nil
}

func (c *Admin) ListArtworkReviews(ctx context.Context, req *v1.ListArtworkReviewsReq) (*v1.ListArtworkReviewsRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	values, err := c.service.ReviewQueue(ctx, req.Status)
	return &v1.ListArtworkReviewsRes{Artworks: values}, err
}

func (c *Admin) ReviewArtwork(ctx context.Context, req *v1.ReviewArtworkReq) (*v1.ReviewArtworkRes, error) {
	operator, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.ReviewArtwork(ctx, operator, req.ArtworkID, model.ReviewInput{Decision: req.Decision, Note: req.Note})
	if err != nil {
		return nil, err
	}
	return &v1.ReviewArtworkRes{Artwork: *value}, nil
}
