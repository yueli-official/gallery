// Package galleryerr 声明 Gallery 不可变的公共 Problem 错误合同。
package galleryerr

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/yueli-official/foundation/go/problem"
)

const (
	CodeNotFound                 = "gallery.not_found"
	CodeGone                     = "gallery.gone"
	CodeNotInitialized           = "gallery.not_initialized"
	CodeForbidden                = "gallery.forbidden"
	CodeConflict                 = "gallery.conflict"
	CodeInvalidState             = "gallery.invalid_state"
	CodeUpstreamFailed           = "gallery.upstream_failed"
	CodeRateLimited              = "gallery.rate_limited"
	CodeChallengeRequired        = "gallery.challenge_required"
	CodeAbuseUnavailable         = "gallery.abuse_unavailable"
	CodeAbuseReplay              = "gallery.abuse_attempt_replayed"
	CodeAuthorizationUnavailable = "gallery.authorization_unavailable"
)

var (
	DescriptorRateLimited = descriptor("common.rate_limited", http.StatusTooManyRequests)
	DescriptorValidation  = descriptor("common.validation_failed", http.StatusBadRequest)
	DescriptorInternal    = descriptor("common.internal", http.StatusInternalServerError)

	descriptors = map[string]problem.Descriptor{
		CodeNotFound:                 descriptor(CodeNotFound, http.StatusNotFound),
		CodeGone:                     descriptor(CodeGone, http.StatusGone),
		CodeNotInitialized:           descriptor(CodeNotInitialized, http.StatusServiceUnavailable),
		CodeForbidden:                descriptor(CodeForbidden, http.StatusForbidden),
		CodeConflict:                 descriptor(CodeConflict, http.StatusConflict),
		CodeInvalidState:             descriptor(CodeInvalidState, http.StatusConflict),
		CodeUpstreamFailed:           descriptor(CodeUpstreamFailed, http.StatusBadGateway),
		CodeRateLimited:              descriptor(CodeRateLimited, http.StatusTooManyRequests),
		CodeChallengeRequired:        descriptor(CodeChallengeRequired, http.StatusForbidden),
		CodeAbuseUnavailable:         descriptor(CodeAbuseUnavailable, http.StatusServiceUnavailable),
		CodeAbuseReplay:              descriptor(CodeAbuseReplay, http.StatusConflict),
		CodeAuthorizationUnavailable: descriptor(CodeAuthorizationUnavailable, http.StatusServiceUnavailable),
	}
)

func descriptor(code string, status int) problem.Descriptor {
	return problem.MustDescriptor(
		problem.MustKind(code, status),
		"https://errors.yueli.dev/problems/"+code,
	)
}

func DescriptorForCode(code string) (problem.Descriptor, bool) {
	value, ok := descriptors[code]
	return value, ok
}

type CatalogEntry struct {
	Code   string `json:"code"`
	Status int    `json:"status"`
}

func Catalog() []CatalogEntry {
	result := make([]CatalogEntry, 0, len(descriptors))
	for code, value := range descriptors {
		result = append(result, CatalogEntry{Code: code, Status: value.Kind().Status()})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Code < result[j].Code })
	return result
}

func mapped(code string, params problem.Parameters) error {
	value, ok := DescriptorForCode(code)
	if !ok {
		return fmt.Errorf("gallery public error code is not declared: %s", code)
	}
	result, err := problem.NewError(value, params)
	if err != nil {
		return fmt.Errorf("gallery public error %s: %w", code, err)
	}
	return result
}

func NotFound(resource, id string) error {
	return mapped(CodeNotFound, map[string]any{"resource": resource, "id": id})
}

func Forbidden() error {
	return mapped(CodeForbidden, nil)
}

func AuthorizationUnavailable() error {
	return mapped(CodeAuthorizationUnavailable, nil)
}

func Conflict(resource string) error {
	return mapped(CodeConflict, map[string]any{"resource": resource})
}

func Gone(resource, id string) error {
	return mapped(CodeGone, map[string]any{"resource": resource, "id": id})
}

func Validation(field, detail string) error {
	pointer := "/" + strings.ReplaceAll(strings.ReplaceAll(field, "~", "~0"), "/", "~1")
	violation := problem.Violation{
		Pointer: pointer, Code: "validation.invalid",
		Params: problem.Parameters{"detail": detail},
	}
	result, err := problem.NewError(DescriptorValidation, nil, violation)
	if err != nil {
		return fmt.Errorf("gallery validation error: %w", err)
	}
	return result
}

func InvalidState(resource, state string) error {
	return mapped(CodeInvalidState, map[string]any{"resource": resource, "state": state})
}

func UpstreamFailed(code string) error {
	return mapped(CodeUpstreamFailed, map[string]any{"upstreamCode": code})
}

func NotInitialized(resource string) error {
	return mapped(CodeNotInitialized, map[string]any{"resource": resource})
}

func RateLimited() error {
	return mapped(CodeRateLimited, nil)
}

func ChallengeRequired(attemptID string) error {
	return mapped(CodeChallengeRequired, map[string]any{
		"attemptId": attemptID, "challenge": "turnstile",
	})
}

func AbuseUnavailable() error {
	return mapped(CodeAbuseUnavailable, nil)
}

func AbuseAttemptReplayed() error {
	return mapped(CodeAbuseReplay, nil)
}
