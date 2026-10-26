package controller

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/yueli-official/asset/referencesync"
	referencehttp "github.com/yueli-official/asset/referencesync/goframe"
	"github.com/yueli-official/gallery/api/internal/galleryauthz"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
)

func NewReferenceSync(module *referencesync.Module) *referencehttp.Controller {
	return referencehttp.New(module, func(ctx context.Context) (string, error) {
		return requireCapability(ctx, galleryauthz.CapabilityAssetSettingsManage)
	}, func(ctx context.Context, err error) error {
		switch {
		case errors.Is(err, referencesync.ErrInvalidQuery):
			return galleryerr.Validation("query", "invalid reference query")
		case errors.Is(err, referencesync.ErrOwnerNotFound), errors.Is(err, sql.ErrNoRows):
			return galleryerr.NotFound("reference", "")
		case errors.Is(err, referencesync.ErrOwnerChanged), errors.Is(err, referencesync.ErrNotRetryable):
			return galleryerr.Conflict("reference")
		default:
			g.Log().Error(ctx, "reference management:", err)
			return galleryerr.UpstreamFailed("asset.reference_unavailable")
		}
	})
}
