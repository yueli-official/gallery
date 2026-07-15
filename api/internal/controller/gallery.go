package controller

import (
	"context"
	"strings"

	v1 "platform/products/gallery/api/api/v1"
	galleryservice "platform/products/gallery/api/internal/gallery"
	"platform/products/gallery/api/internal/model"
)

type Public struct{ service *galleryservice.Service }

func NewPublic(service *galleryservice.Service) *Public { return &Public{service: service} }

func (c *Public) GetDiscovery(ctx context.Context, req *v1.GetDiscoveryReq) (*v1.GetDiscoveryRes, error) {
	value, err := c.service.Discovery(ctx, req.Seed)
	if err != nil {
		return nil, err
	}
	return &v1.GetDiscoveryRes{
		Site: value.Site, Seed: value.Seed, Images: value.Images,
		Categories: value.Categories, Facets: value.Facets,
	}, nil
}

func (c *Public) ListImages(ctx context.Context, req *v1.ListImagesReq) (*v1.ListImagesRes, error) {
	page, err := c.service.Images(ctx, model.ImageQuery{
		Search: req.Search, Sort: req.Sort, Page: req.Page, PageSize: req.Size,
		CategoryRefs: splitCSV(req.Categories), FacetRefs: splitCSV(req.Facets), Tag: req.Tag,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListImagesRes{ImagePage: *page}, nil
}

func (c *Public) GetImage(ctx context.Context, req *v1.GetImageReq) (*v1.GetImageRes, error) {
	subject, _ := optionalSubject(ctx)
	value, err := c.service.Image(ctx, req.ImageID, subject.ID)
	if err != nil {
		return nil, err
	}
	return &v1.GetImageRes{Image: *value}, nil
}

func (c *Public) ListCollections(ctx context.Context, _ *v1.ListCollectionsReq) (*v1.ListCollectionsRes, error) {
	values, err := c.service.Collections(ctx)
	return &v1.ListCollectionsRes{Collections: values}, err
}

func (c *Public) GetCollection(ctx context.Context, req *v1.GetCollectionReq) (*v1.GetCollectionRes, error) {
	value, err := c.service.Collection(ctx, req.Slug, req.Page, req.Size)
	if err != nil {
		return nil, err
	}
	return &v1.GetCollectionRes{Collection: *value}, nil
}

func (c *Public) GetRankings(ctx context.Context, req *v1.GetRankingsReq) (*v1.GetRankingsRes, error) {
	value, err := c.service.Ranking(ctx, req.Kind, req.Window)
	if err != nil {
		return nil, err
	}
	return &v1.GetRankingsRes{Ranking: *value}, nil
}

func (c *Public) CreateCase(ctx context.Context, req *v1.CreateCaseReq) (*v1.CreateCaseRes, error) {
	subject, _ := optionalSubject(ctx)
	value, err := c.service.Report(ctx, subject, req.ImageID, req.CaseInput)
	if err != nil {
		return nil, err
	}
	return &v1.CreateCaseRes{Case: *value}, nil
}

func (c *Public) TrackImageEvent(ctx context.Context, req *v1.TrackImageEventReq) (*v1.TrackImageEventRes, error) {
	subject, _ := optionalSubject(ctx)
	if err := c.service.TrackEvent(ctx, subject, req.ImageID, req.EventInput); err != nil {
		return nil, err
	}
	return &v1.TrackImageEventRes{Recorded: true}, nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	return result
}
