package galleryerr

import (
	"net/http"

	"platform/gokit/errs"
)

var (
	CodeNotFound       = errs.Register("gallery.not_found", http.StatusNotFound)
	CodeNotInitialized = errs.Register("gallery.not_initialized", 503)
)

func NotFound(resource, id string) *errs.Coded {
	return errs.New(CodeNotFound, "gallery resource not found", map[string]any{"resource": resource, "id": id})
}

func NotInitialized(resource string) *errs.Coded {
	return errs.New(CodeNotInitialized, "gallery site configuration is not initialized", map[string]any{"resource": resource})
}
