package gallery

import (
	"context"
	"strings"

	"platform/gokit/facet"
	"platform/products/gallery/api/internal/galleryerr"
	"platform/products/gallery/api/internal/model"
)

type Store interface {
	SiteSettings(context.Context) (*model.SiteSettings, error)
	Featured(context.Context, int) ([]model.ArtworkCard, error)
	Latest(context.Context, int) ([]model.ArtworkCard, error)
	Facets(context.Context) ([]facet.Facet, error)
	FacetValues(context.Context) ([]facet.Value, error)
	Artwork(context.Context, string) (*model.ArtworkDetail, error)
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) Discovery(ctx context.Context) (*model.Discovery, error) {
	settings, err := s.store.SiteSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, galleryerr.NotInitialized("site_settings")
	}
	featured, err := s.store.Featured(ctx, 6)
	if err != nil {
		return nil, err
	}
	latest, err := s.store.Latest(ctx, 24)
	if err != nil {
		return nil, err
	}
	facets, err := s.store.Facets(ctx)
	if err != nil {
		return nil, err
	}
	values, err := s.store.FacetValues(ctx)
	if err != nil {
		return nil, err
	}
	if featured == nil {
		featured = []model.ArtworkCard{}
	}
	if latest == nil {
		latest = []model.ArtworkCard{}
	}
	catalog, err := facet.NewCatalog(facets, values)
	if err != nil {
		return nil, err
	}
	return &model.Discovery{
		Site: *settings, Featured: featured, Latest: latest,
		Facets: catalog.Facets(), FacetValues: catalog.Values(),
	}, nil
}

func (s *Service) Artwork(ctx context.Context, id string) (*model.ArtworkDetail, error) {
	id = strings.TrimSpace(id)
	value, err := s.store.Artwork(ctx, id)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.NotFound("artwork", id)
	}
	if value.Tags == nil {
		value.Tags = []string{}
	}
	if value.Assets == nil {
		value.Assets = []model.Asset{}
	}
	return value, nil
}
