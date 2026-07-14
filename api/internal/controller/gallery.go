package controller

import (
	"context"

	v1 "platform/products/gallery/api/api/v1"
	galleryservice "platform/products/gallery/api/internal/gallery"
)

type Public struct {
	service *galleryservice.Service
}

func NewPublic(service *galleryservice.Service) *Public {
	return &Public{service: service}
}

func (c *Public) GetDiscovery(ctx context.Context, _ *v1.GetDiscoveryReq) (*v1.GetDiscoveryRes, error) {
	value, err := c.service.Discovery(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetDiscoveryRes{
		Site: value.Site, Featured: value.Featured, Latest: value.Latest, Categories: value.Categories,
	}, nil
}

func (c *Public) GetArtwork(ctx context.Context, req *v1.GetArtworkReq) (*v1.GetArtworkRes, error) {
	value, err := c.service.Artwork(ctx, req.ArtworkID)
	if err != nil {
		return nil, err
	}
	return &v1.GetArtworkRes{Artwork: *value}, nil
}
