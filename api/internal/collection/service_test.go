package collection

import (
	"context"
	"testing"

	"github.com/yueli-official/gallery/api/internal/model"
)

type memoryStore struct {
	value   *model.Collection
	add     []string
	remove  []string
	version int64
}

func (m *memoryStore) FindSingleton(context.Context, string, string) (*model.Collection, error) {
	return m.value, nil
}

func (m *memoryStore) CreateCollection(_ context.Context, input CreateInput) (*model.Collection, error) {
	m.value = &model.Collection{ID: "collection-1", Kind: input.Kind, ResourceKind: input.ResourceKind, OwnerKind: input.OwnerKind, OwnerID: input.OwnerID, Visibility: input.Visibility, Name: input.Name, Version: 1}
	return m.value, nil
}

func (m *memoryStore) OwnedCollection(context.Context, string, string) (*model.Collection, error) {
	return m.value, nil
}

func (m *memoryStore) MutateMembers(_ context.Context, _ string, version int64, add, remove []string) (*model.Collection, error) {
	m.version, m.add, m.remove = version, add, remove
	m.value.Version++
	return m.value, nil
}

type memoryResources map[string]bool

func (m memoryResources) Collectable(_ context.Context, ids []string) (map[string]bool, error) {
	result := make(map[string]bool, len(ids))
	for _, id := range ids {
		result[id] = m[id]
	}
	return result, nil
}

func newGalleryService(t *testing.T, store Store, resources ResourcePort) *Service {
	t.Helper()
	service, err := New(store, resources, GalleryKinds()...)
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func TestFavoritesAreALazyPrivateSingleton(t *testing.T) {
	store := &memoryStore{}
	service := newGalleryService(t, store, memoryResources{})

	first, err := service.EnsureFavorites(context.Background(), "user-1")
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.EnsureFavorites(context.Background(), "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || first.Kind != KindFavorites || first.Visibility != "private" || first.Name != "我的收藏" {
		t.Fatalf("unexpected favorites singleton: %#v %#v", first, second)
	}
}

func TestMutateIsIdempotentAtTheBoundaryAndChecksResources(t *testing.T) {
	store := &memoryStore{value: &model.Collection{ID: "collection-1", Kind: KindFavorites, OwnerID: "user-1", Visibility: "private", Version: 4}}
	service := newGalleryService(t, store, memoryResources{"image-1": true})

	value, err := service.Mutate(context.Background(), "user-1", "collection-1", 4, []string{" image-1 ", "image-1"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if value.Version != 5 || store.version != 4 || len(store.add) != 1 || store.add[0] != "image-1" {
		t.Fatalf("unexpected atomic mutation: value=%#v add=%#v", value, store.add)
	}
}

func TestMutateRejectsUncollectableImages(t *testing.T) {
	store := &memoryStore{value: &model.Collection{ID: "collection-1", Kind: KindFavorites, OwnerID: "user-1", Visibility: "private", Version: 1}}
	service := newGalleryService(t, store, memoryResources{})
	if _, err := service.Mutate(context.Background(), "user-1", "collection-1", 1, []string{"hidden-image"}, nil); err == nil {
		t.Fatal("expected uncollectable image to be rejected")
	}
}
