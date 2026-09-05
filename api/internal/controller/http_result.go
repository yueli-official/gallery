package controller

import (
	"context"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/yueli-official/foundation/go/problem"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/model"
)

func writeSuccess(ctx context.Context, status int, location string) {
	if r := ghttp.RequestFromCtx(ctx); r != nil {
		if location != "" {
			r.Response.Header().Set("Location", location)
		}
		r.Response.WriteHeader(status)
	}
}
func boundedPage(page int) int {
	if page < 1 {
		return 1
	}
	if page > 100000 {
		return 100000
	}
	return page
}
func boundedSize(size, fallback int) int {
	if size < 1 {
		return fallback
	}
	if size > 100 {
		return 100
	}
	return size
}

func batchFailure(ctx context.Context, cause error) *problem.Problem {
	trace := "gallery-batch"
	if r := ghttp.RequestFromCtx(ctx); r != nil {
		if value := r.Response.Header().Get("X-Trace-Id"); value != "" {
			trace = value
		} else if value := r.Header.Get("X-Trace-Id"); value != "" {
			trace = value
		}
	}
	value, ok, err := problem.FromError(cause, trace)
	if err != nil || !ok {
		fallback, _ := problem.NewError(galleryerr.DescriptorInternal, nil)
		value, _, err = problem.FromError(fallback, trace)
		if err != nil {
			value, _, _ = problem.FromError(fallback, "gallery-batch")
		}
	}
	return &value
}
func imageBatch(ctx context.Context, items []model.BulkImageActionResult) []model.BulkImageActionResult {
	for i := range items {
		if !items[i].Success {
			items[i].Failure = batchFailure(ctx, items[i].Cause)
		}
	}
	return items
}
func submissionBatch(ctx context.Context, items []model.BulkSubmissionReviewResult) []model.BulkSubmissionReviewResult {
	for i := range items {
		if !items[i].Success {
			items[i].Failure = batchFailure(ctx, items[i].Cause)
		}
	}
	return items
}

func itemsOrEmpty[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}
