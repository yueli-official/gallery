// Package galleryerr 声明 Gallery 不可变的公共 Problem 错误合同。
package galleryerr

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/yueli-official/foundation/go/problem"
)

var (
	DescriptorRateLimited = descriptor("common.rate_limited", http.StatusTooManyRequests)
	DescriptorValidation  = descriptor("common.validation_failed", http.StatusBadRequest)
	DescriptorInternal    = descriptor("common.internal", http.StatusInternalServerError)
)

func mapped(code string, params problem.Parameters) error {
	for key, value := range params {
		text, ok := value.(string)
		limit := 128
		switch key {
		case "resource", "state":
			limit = 64
		case "challenge":
			limit = 32
		}
		if !ok || len(text) > limit || strings.ContainsAny(text, "\r\n") {
			return fmt.Errorf("gallery public parameter %s exceeds its declared budget", key)
		}
	}
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

func Validation(field, _ string) error {
	pointer := "/" + strings.ReplaceAll(strings.ReplaceAll(field, "~", "~0"), "/", "~1")
	violation := problem.Violation{
		Pointer: pointer, Code: "validation.invalid",
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
	if !regexp.MustCompile(`^[a-z][a-z0-9._-]{0,127}$`).MatchString(code) {
		code = "asset.unavailable"
	}
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
