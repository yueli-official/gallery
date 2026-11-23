package gallery

import (
	"context"
	"fmt"

	"github.com/yueli-official/foundation/go/identifier"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/model"
)

type ImagePublicKeyStore interface {
	ImageIDByPublicKey(context.Context, string) (string, error)
}

// Public page keys have one canonical format; UUID and Base64 entity IDs are
// intentionally not interpreted as page addresses.
func (s *Service) ImageByPublicKey(ctx context.Context, key, userID string) (*model.ImageDetail, error) {
	if _, err := identifier.CompactURLV1.Parse(key); err != nil {
		return nil, galleryerr.NotFound("image", key)
	}
	store, ok := s.store.(ImagePublicKeyStore)
	if !ok {
		return nil, fmt.Errorf("image public key store is not configured")
	}
	id, err := store.ImageIDByPublicKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if id == "" {
		return nil, galleryerr.NotFound("image", key)
	}
	return s.Image(ctx, id, userID)
}
