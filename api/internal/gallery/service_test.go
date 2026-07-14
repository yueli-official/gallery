package gallery

import (
	"context"
	"errors"
	"testing"

	"platform/gokit/errs"
	"platform/products/gallery/api/internal/model"
)

type fakeStore struct {
	settings   *model.SiteSettings
	featured   []model.ArtworkCard
	latest     []model.ArtworkCard
	categories []model.Category
	artwork    *model.ArtworkDetail
}

func (f fakeStore) SiteSettings(context.Context) (*model.SiteSettings, error) {
	return f.settings, nil
}

func (f fakeStore) Featured(context.Context, int) ([]model.ArtworkCard, error) {
	return f.featured, nil
}

func (f fakeStore) Latest(context.Context, int) ([]model.ArtworkCard, error) {
	return f.latest, nil
}

func (f fakeStore) Categories(context.Context) ([]model.Category, error) {
	return f.categories, nil
}

func (f fakeStore) Artwork(context.Context, string) (*model.ArtworkDetail, error) {
	return f.artwork, nil
}

func TestDiscoveryRequiresSiteSettings(t *testing.T) {
	_, err := New(fakeStore{}).Discovery(context.Background())
	var coded *errs.Coded
	if !errors.As(err, &coded) || coded.Code != "gallery.not_initialized" {
		t.Fatalf("expected gallery.not_initialized, got %v", err)
	}
}

func TestDiscoveryUsesEmptyArraysInsteadOfNull(t *testing.T) {
	discovery, err := New(fakeStore{settings: &model.SiteSettings{Name: "PLN Gallery"}}).Discovery(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if discovery.Featured == nil || discovery.Latest == nil || discovery.Categories == nil {
		t.Fatalf("discovery collections must serialize as arrays: %#v", discovery)
	}
}

func TestDiscoveryKeepsEditorialAndLatestSeparate(t *testing.T) {
	settings := &model.SiteSettings{Name: "PLN Gallery"}
	featured := []model.ArtworkCard{{ID: "featured"}}
	latest := []model.ArtworkCard{{ID: "latest"}}
	categories := []model.Category{{ID: "illustration"}}

	discovery, err := New(fakeStore{
		settings: settings, featured: featured, latest: latest, categories: categories,
	}).Discovery(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if discovery.Site.Name != settings.Name || discovery.Featured[0].ID != "featured" || discovery.Latest[0].ID != "latest" || discovery.Categories[0].ID != "illustration" {
		t.Fatalf("unexpected discovery response: %#v", discovery)
	}
}

func TestArtworkReturnsNamespacedNotFound(t *testing.T) {
	_, err := New(fakeStore{}).Artwork(context.Background(), " missing ")
	var coded *errs.Coded
	if !errors.As(err, &coded) || coded.Code != "gallery.not_found" || coded.Params["id"] != "missing" {
		t.Fatalf("expected gallery.not_found for trimmed id, got %v", err)
	}
}
