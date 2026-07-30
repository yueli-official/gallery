package galleryerr_test

import (
	"net/http"
	"testing"

	"github.com/yueli-official/foundation/go/problem"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
)

func TestCodesRegistered(t *testing.T) {
	cases := map[string]int{
		galleryerr.CodeNotFound:                 http.StatusNotFound,
		galleryerr.CodeGone:                     http.StatusGone,
		galleryerr.CodeNotInitialized:           http.StatusServiceUnavailable,
		galleryerr.CodeForbidden:                http.StatusForbidden,
		galleryerr.CodeConflict:                 http.StatusConflict,
		galleryerr.CodeInvalidState:             http.StatusConflict,
		galleryerr.CodeUpstreamFailed:           http.StatusBadGateway,
		galleryerr.CodeRateLimited:              http.StatusTooManyRequests,
		galleryerr.CodeChallengeRequired:        http.StatusForbidden,
		galleryerr.CodeAbuseUnavailable:         http.StatusServiceUnavailable,
		galleryerr.CodeAbuseReplay:              http.StatusConflict,
		galleryerr.CodeAuthorizationUnavailable: http.StatusServiceUnavailable,
	}
	for code, want := range cases {
		descriptor, ok := galleryerr.DescriptorForCode(code)
		if !ok {
			t.Errorf("DescriptorForCode(%q) is missing", code)
			continue
		}
		if got := descriptor.Kind().Status(); got != want {
			t.Errorf("Status(%q) = %d, want %d", code, got, want)
		}
	}
}

func TestConstructorsCarryCode(t *testing.T) {
	cases := map[string]error{
		galleryerr.CodeNotFound:                 galleryerr.NotFound("image", "id"),
		galleryerr.CodeGone:                     galleryerr.Gone("image", "id"),
		galleryerr.CodeNotInitialized:           galleryerr.NotInitialized("store"),
		galleryerr.CodeForbidden:                galleryerr.Forbidden(),
		galleryerr.CodeConflict:                 galleryerr.Conflict("image"),
		galleryerr.CodeInvalidState:             galleryerr.InvalidState("image", "draft"),
		galleryerr.CodeUpstreamFailed:           galleryerr.UpstreamFailed("asset.unavailable"),
		galleryerr.CodeRateLimited:              galleryerr.RateLimited(),
		galleryerr.CodeChallengeRequired:        galleryerr.ChallengeRequired("attempt"),
		galleryerr.CodeAbuseUnavailable:         galleryerr.AbuseUnavailable(),
		galleryerr.CodeAbuseReplay:              galleryerr.AbuseAttemptReplayed(),
		galleryerr.CodeAuthorizationUnavailable: galleryerr.AuthorizationUnavailable(),
	}
	for want, err := range cases {
		value, ok, resolveErr := problem.FromError(err, "gallery-error-inspection")
		if resolveErr != nil || !ok || value.Code != want {
			t.Errorf("FromError(%s) = %#v, %v, %v", want, value, ok, resolveErr)
		}
	}
}
