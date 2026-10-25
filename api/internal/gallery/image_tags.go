package gallery

import (
	"context"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yueli-official/foundation/go/classification"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/model"
)

// Tag lookup preparation belongs to Classification; this flow proposes words,
// rather than changing the image's category/facet assignments.
func (s *Service) tagLookups(ctx context.Context, tags []string) ([]classification.TagLookupRequest, error) {
	catalog, err := s.classificationCatalog(ctx)
	if err != nil {
		return nil, err
	}
	return catalog.Classify(classification.ClassifyRequest{PolicyKey: "gallery.image.public", Tags: tags}).FactRequest().TagLookups, nil
}

func (s *Service) TagSuggestions(ctx context.Context, query string) ([]model.TagCandidate, error) {
	if utf8.RuneCountInString(query) > 64 {
		return nil, galleryerr.Validation("q", "query is too long")
	}
	lookups, err := s.tagLookups(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	key := ""
	if len(lookups) > 0 {
		key = lookups[0].LookupKey
	}
	items, err := s.store.TagSuggestions(ctx, key, 28)
	if items == nil {
		items = []model.TagCandidate{}
	}
	for i := range items {
		items[i].ID = PublicID(items[i].ID)
	}
	return items, err
}

func (s *Service) ProposeImageTags(ctx context.Context, subject model.Subject, rawID string, tags []string) ([]model.ClassificationTagProposal, error) {
	if !oneOf(subject.Kind, "user", "guest") || strings.TrimSpace(subject.ID) == "" {
		return nil, galleryerr.Forbidden()
	}
	id, err := DatabaseID(rawID)
	if err != nil {
		return nil, galleryerr.Validation("imageId", "invalid image ID")
	}
	if len(tags) == 0 || len(tags) > 10 {
		return nil, galleryerr.Validation("tags", "choose one to ten tags")
	}
	for _, tag := range tags {
		if strings.TrimSpace(tag) == "" || utf8.RuneCountInString(tag) > 32 || strings.ContainsFunc(tag, unicode.IsControl) {
			return nil, galleryerr.Validation("tags", "invalid tag")
		}
	}
	lookups, err := s.tagLookups(ctx, tags)
	if err != nil {
		return nil, err
	}
	items, err := s.store.ProposeImageTags(ctx, subject, id, lookups)
	if items == nil {
		items = []model.ClassificationTagProposal{}
	}
	for i := range items {
		publicTagProposal(&items[i])
	}
	return items, err
}

func publicTagProposal(value *model.ClassificationTagProposal) {
	value.ID = PublicID(value.ID)
	value.SubmissionID = PublicID(value.SubmissionID)
	value.ImageID = PublicID(value.ImageID)
	value.ResolvedTagID = PublicID(value.ResolvedTagID)
}
