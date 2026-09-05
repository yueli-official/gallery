package controller

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/model"
	"strings"
	"testing"
)

func TestBatchFailureDoesNotSerializeCause(t *testing.T) {
	values := imageBatch(context.Background(), []model.BulkImageActionResult{{ImageID: "one", Success: true}, {ImageID: "two", Cause: errors.New("SQL password=secret")}, {ImageID: "three", Cause: galleryerr.NotFound("image", "three")}})
	if values[1].Failure == nil || values[1].Failure.Status != 500 || values[2].Failure.Code != "gallery.not_found" {
		t.Fatalf("unexpected projection: %#v", values)
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "secret") || strings.Contains(string(encoded), `"error"`) {
		t.Fatalf("raw error exposed: %s", encoded)
	}
	if values[0].Failure != nil {
		t.Fatal("success has failure")
	}
}
func TestPaginationWireUsesSizeAndDerivedPageCount(t *testing.T) {
	encoded, err := json.Marshal(model.ImagePage{Items: []model.ImageCard{}, Page: 2, PageSize: 20, Total: 21, TotalPages: 2})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "pageSize") || strings.Contains(string(encoded), "totalPages") || !strings.Contains(string(encoded), `"size":20`) {
		t.Fatalf("legacy wire: %s", encoded)
	}
}
