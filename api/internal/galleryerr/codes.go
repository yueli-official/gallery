package galleryerr

import (
	"net/http"
	"strings"

	"github.com/yueli-official/foundation/go/problem"
	"platform/gokit/errs"
)

var (
	CodeNotFound          = errs.Register("gallery.not_found", http.StatusNotFound)
	CodeGone              = errs.Register("gallery.gone", 410)
	CodeNotInitialized    = errs.Register("gallery.not_initialized", 503)
	CodeForbidden         = errs.Register("gallery.forbidden", http.StatusForbidden)
	CodeConflict          = errs.Register("gallery.conflict", http.StatusConflict)
	CodeInvalidState      = errs.Register("gallery.invalid_state", http.StatusConflict)
	CodeUpstreamFailed    = errs.Register("gallery.upstream_failed", http.StatusBadGateway)
	CodeRateLimited       = errs.Register("gallery.rate_limited", http.StatusTooManyRequests)
	CodeChallengeRequired = errs.Register("gallery.challenge_required", http.StatusForbidden)
	CodeAbuseUnavailable  = errs.Register("gallery.abuse_unavailable", http.StatusServiceUnavailable)
	CodeAbuseReplay       = errs.Register("gallery.abuse_attempt_replayed", http.StatusConflict)
)

func NotFound(resource, id string) *errs.Coded {
	return errs.New(CodeNotFound, "gallery resource not found", map[string]any{"resource": resource, "id": id})
}

func Forbidden() *errs.Coded {
	return errs.New(CodeForbidden, "gallery operation is forbidden", nil)
}

func Conflict(resource string) *errs.Coded {
	return errs.New(CodeConflict, "gallery resource conflicts with an existing record", map[string]any{"resource": resource})
}

func Gone(resource, id string) *errs.Coded {
	return errs.New(CodeGone, "gallery resource was permanently removed", map[string]any{"resource": resource, "id": id})
}

func Validation(field, detail string) *errs.Coded {
	return errs.New(errs.CommonValidationFailed, "validation failed", map[string]any{
		"details": []problem.Violation{{Pointer: "/" + strings.ReplaceAll(strings.ReplaceAll(field, "~", "~0"), "/", "~1"), Code: "validation.invalid", Params: problem.Parameters{"detail": detail}}},
	})
}

func InvalidState(resource, state string) *errs.Coded {
	return errs.New(CodeInvalidState, "gallery resource is not in an allowed state", map[string]any{"resource": resource, "state": state})
}

func UpstreamFailed(code string) *errs.Coded {
	return errs.New(CodeUpstreamFailed, "gallery asset operation failed", map[string]any{"upstreamCode": code})
}

func NotInitialized(resource string) *errs.Coded {
	return errs.New(CodeNotInitialized, "gallery site configuration is not initialized", map[string]any{"resource": resource})
}

func RateLimited() *errs.Coded {
	return errs.New(CodeRateLimited, "too many submissions — please try again later", nil)
}

func ChallengeRequired(attemptID string) *errs.Coded {
	return errs.New(CodeChallengeRequired, "additional verification required", map[string]any{
		"attemptId": attemptID, "challenge": "turnstile",
	})
}

func AbuseUnavailable() *errs.Coded {
	return errs.New(CodeAbuseUnavailable, "submission admission is temporarily unavailable", nil)
}

func AbuseAttemptReplayed() *errs.Coded {
	return errs.New(CodeAbuseReplay, "submission attempt was already admitted", nil)
}
