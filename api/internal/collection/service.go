package collection

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"platform/products/gallery/api/internal/model"
)

const (
	KindEditorial = "gallery.editorial"
	KindFavorites = "gallery.favorites"
	ImageResource = "gallery.image"
)

type Kind struct {
	Name            string
	ResourceKind    string
	OwnerKind       string
	ForceVisibility string
	Singleton       bool
}

type CreateInput struct {
	Kind         string
	ResourceKind string
	OwnerKind    string
	OwnerID      string
	Visibility   string
	Name         string
	Description  string
}

type Store interface {
	FindSingleton(context.Context, string, string) (*model.Collection, error)
	CreateCollection(context.Context, CreateInput) (*model.Collection, error)
	OwnedCollection(context.Context, string, string) (*model.Collection, error)
	MutateMembers(context.Context, string, int64, []string, []string) (*model.Collection, error)
}

type ResourcePort interface {
	Collectable(context.Context, []string) (map[string]bool, error)
}

type Service struct {
	store     Store
	resources ResourcePort
	kinds     map[string]Kind
}

func New(store Store, resources ResourcePort, kinds ...Kind) (*Service, error) {
	registry := make(map[string]Kind, len(kinds))
	for _, kind := range kinds {
		kind.Name = strings.TrimSpace(kind.Name)
		if !strings.Contains(kind.Name, ".") || kind.ResourceKind == "" || kind.OwnerKind == "" {
			return nil, fmt.Errorf("invalid collection kind registration %q", kind.Name)
		}
		if _, exists := registry[kind.Name]; exists {
			return nil, fmt.Errorf("duplicate collection kind %q", kind.Name)
		}
		registry[kind.Name] = kind
	}
	return &Service{store: store, resources: resources, kinds: registry}, nil
}

func GalleryKinds() []Kind {
	return []Kind{
		{Name: KindEditorial, ResourceKind: ImageResource, OwnerKind: "site"},
		{Name: KindFavorites, ResourceKind: ImageResource, OwnerKind: "user", ForceVisibility: "private", Singleton: true},
	}
}

func (s *Service) EnsureFavorites(ctx context.Context, userID string) (*model.Collection, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("favorites owner is required")
	}
	if current, err := s.store.FindSingleton(ctx, KindFavorites, userID); err != nil || current != nil {
		return current, err
	}
	return s.store.CreateCollection(ctx, CreateInput{
		Kind: KindFavorites, ResourceKind: ImageResource, OwnerKind: "user", OwnerID: userID,
		Visibility: "private", Name: "我的收藏",
	})
}

func (s *Service) Mutate(ctx context.Context, ownerID, collectionID string, expectedVersion int64, add, remove []string) (*model.Collection, error) {
	collection, err := s.store.OwnedCollection(ctx, strings.TrimSpace(collectionID), strings.TrimSpace(ownerID))
	if err != nil || collection == nil {
		return collection, err
	}
	kind, ok := s.kinds[collection.Kind]
	if !ok {
		return nil, fmt.Errorf("collection kind %q is not registered", collection.Kind)
	}
	add = uniqueNonEmpty(add)
	remove = uniqueNonEmpty(remove)
	for _, id := range add {
		if slices.Contains(remove, id) {
			return nil, fmt.Errorf("image %q cannot be added and removed together", id)
		}
	}
	if len(add) > 0 {
		if s.resources == nil {
			return nil, fmt.Errorf("collection resource port is not configured")
		}
		collectable, err := s.resources.Collectable(ctx, add)
		if err != nil {
			return nil, err
		}
		for _, id := range add {
			if !collectable[id] {
				return nil, fmt.Errorf("image %q is not collectable", id)
			}
		}
	}
	if kind.ForceVisibility != "" && collection.Visibility != kind.ForceVisibility {
		return nil, fmt.Errorf("collection kind %q requires %s visibility", kind.Name, kind.ForceVisibility)
	}
	return s.store.MutateMembers(ctx, collection.ID, expectedVersion, add, remove)
}

func uniqueNonEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
