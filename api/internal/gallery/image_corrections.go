package gallery

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/model"
)

func (s *Service) validateMetadataCorrection(ctx context.Context, input *model.ImageMetadataCorrectionInput) error {
	if input == nil {
		return galleryerr.Validation("proposedMetadata", "metadata changes are required")
	}
	input.ExpectedImageUpdatedAt = strings.TrimSpace(input.ExpectedImageUpdatedAt)
	if _, err := time.Parse(time.RFC3339Nano, input.ExpectedImageUpdatedAt); err != nil {
		return galleryerr.Validation("proposedMetadata/expectedImageUpdatedAt", "a current image version is required")
	}
	if input.Title == nil && input.Description == nil && input.AltText == nil && input.SourceURL == nil && input.Classification == nil {
		return galleryerr.Validation("proposedMetadata", "choose at least one change")
	}
	for _, field := range []struct {
		name     string
		value    *string
		max      int
		required bool
	}{{"title", input.Title, 160, true}, {"description", input.Description, 10000, false}, {"altText", input.AltText, 2000, true}, {"sourceUrl", input.SourceURL, 2000, false}} {
		if field.value == nil {
			continue
		}
		*field.value = strings.TrimSpace(*field.value)
		if (field.required && *field.value == "") || len([]rune(*field.value)) > field.max || strings.ContainsRune(*field.value, 0) {
			return galleryerr.Validation("proposedMetadata/"+field.name, "invalid proposed field")
		}
	}
	if input.SourceURL != nil && *input.SourceURL != "" {
		parsed, err := url.ParseRequestURI(*input.SourceURL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return galleryerr.Validation("proposedMetadata/sourceUrl", "an absolute HTTP or HTTPS URL is required")
		}
	}
	if proposed := input.Classification; proposed != nil {
		if len(proposed.FacetValueIDs) > 64 {
			return galleryerr.Validation("proposedMetadata/classification/facetValueIds", "too many selections")
		}
		category, err := DatabaseID(proposed.PrimaryCategoryID)
		if err != nil {
			return galleryerr.Validation("proposedMetadata/classification/primaryCategoryId", "choose an active primary category")
		}
		values := make([]string, len(proposed.FacetValueIDs))
		for i, id := range proposed.FacetValueIDs {
			values[i], err = DatabaseID(id)
			if err != nil {
				return galleryerr.Validation("proposedMetadata/classification/facetValueIds", "invalid facet value")
			}
		}
		selection := &model.AdminImageClassificationInput{PrimaryCategoryID: category, FacetValueIDs: values}
		if err := s.classifyImageEdit(ctx, selection); err != nil {
			return err
		}
		proposed.PrimaryCategoryID, proposed.FacetValueIDs = selection.PrimaryCategoryID, selection.FacetValueIDs
	}
	return nil
}

func normalizeCase(value *model.Case) error {
	if value == nil {
		return nil
	}
	value.ID, value.ImageID, value.SubmissionID = PublicID(value.ID), PublicID(value.ImageID), PublicID(value.SubmissionID)
	if value.ProposedMetadataJSON != "" {
		var proposal model.ImageMetadataCorrectionInput
		if err := json.Unmarshal([]byte(value.ProposedMetadataJSON), &proposal); err != nil {
			return err
		}
		if proposal.Classification != nil {
			proposal.Classification.PrimaryCategoryID = PublicID(proposal.Classification.PrimaryCategoryID)
			for i, id := range proposal.Classification.FacetValueIDs {
				proposal.Classification.FacetValueIDs[i] = PublicID(id)
			}
		}
		value.ProposedMetadata = &proposal
	}
	return nil
}

func (s *Service) applyMetadataCorrection(ctx context.Context, operator, caseID string, input model.CaseResolutionInput) (*model.Case, error) {
	if input.Status != "resolved" || strings.TrimSpace(operator) == "" {
		return nil, galleryerr.Validation("applyMetadata", "metadata acceptance must resolve the case")
	}
	item, err := s.store.Case(ctx, caseID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, galleryerr.NotFound("case", PublicID(caseID))
	}
	expectedCase, _ := time.Parse(time.RFC3339Nano, input.ExpectedUpdatedAt)
	if item.UpdatedAt == nil || !item.UpdatedAt.Equal(expectedCase) || !oneOf(item.Status, "open", "reviewing") {
		return nil, galleryerr.Conflict("case_version")
	}
	if item.Kind != "metadata_correction" {
		return nil, galleryerr.Validation("applyMetadata", "this case has no metadata changes")
	}
	var proposal model.ImageMetadataCorrectionInput
	if err := json.Unmarshal([]byte(item.ProposedMetadataJSON), &proposal); err != nil {
		return nil, err
	}
	if err := s.validateMetadataCorrection(ctx, &proposal); err != nil {
		return nil, err
	}
	current, err := s.store.AdminImage(ctx, item.ImageID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, galleryerr.NotFound("image", PublicID(item.ImageID))
	}
	expectedImage, _ := time.Parse(time.RFC3339Nano, proposal.ExpectedImageUpdatedAt)
	if current.UpdatedAt == nil || !current.UpdatedAt.Equal(expectedImage) {
		return nil, galleryerr.Conflict("image_version")
	}
	update := model.AdminImageUpdateInput{
		Operator: operator, ExpectedUpdatedAt: proposal.ExpectedImageUpdatedAt,
		Title: current.Title, AltText: current.AltText, Description: current.Description, SourceURL: current.SourceURL,
		MetadataCaseID: caseID, MetadataCaseVersion: input.ExpectedUpdatedAt, MetadataCaseNote: input.Note,
	}
	if proposal.Title != nil {
		update.Title = *proposal.Title
	}
	if proposal.AltText != nil {
		update.AltText = *proposal.AltText
	}
	if proposal.Description != nil {
		update.Description = *proposal.Description
	}
	if proposal.SourceURL != nil {
		update.SourceURL = *proposal.SourceURL
	}
	if proposal.Classification != nil {
		update.Classification = &model.AdminImageClassificationInput{PrimaryCategoryID: proposal.Classification.PrimaryCategoryID, FacetValueIDs: proposal.Classification.FacetValueIDs}
		// Tags have their own Classification proposal review. Preserve the current
		// assignments when accepting a category/facet correction.
		for _, tag := range current.Tags {
			update.Classification.TagIDs = append(update.Classification.TagIDs, tag.ID)
		}
	}
	if err := s.normalizeAdminImageUpdate(ctx, &update); err != nil {
		return nil, err
	}
	if _, err := s.store.UpdateAdminImage(ctx, item.ImageID, update); err != nil {
		return nil, err
	}
	value, err := s.store.Case(ctx, caseID)
	if err != nil {
		return nil, err
	}
	if err := normalizeCase(value); err != nil {
		return nil, err
	}
	return value, nil
}

func (s *Service) normalizeAdminImageUpdate(ctx context.Context, input *model.AdminImageUpdateInput) error {
	var err error
	input.Title, input.Description = strings.TrimSpace(input.Title), strings.TrimSpace(input.Description)
	input.AltText, input.SourceURL = strings.TrimSpace(input.AltText), strings.TrimSpace(input.SourceURL)
	input.ExpectedUpdatedAt = strings.TrimSpace(input.ExpectedUpdatedAt)
	if input.Title == "" || len([]rune(input.Title)) > 160 {
		return galleryerr.Validation("title", "title is required and must be at most 160 characters")
	}
	if input.AltText == "" {
		return galleryerr.Validation("altText", "alt text is required")
	}
	if _, parseErr := time.Parse(time.RFC3339Nano, input.ExpectedUpdatedAt); parseErr != nil {
		return galleryerr.Validation("expectedUpdatedAt", "a current RFC3339 timestamp is required")
	}
	if input.SourceURL != "" {
		parsed, parseErr := url.ParseRequestURI(input.SourceURL)
		if parseErr != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return galleryerr.Validation("sourceUrl", "source URL must be HTTP or HTTPS")
		}
	}
	if input.Classification != nil {
		classificationInput := input.Classification
		classificationInput.PrimaryCategoryID, err = DatabaseID(strings.TrimSpace(classificationInput.PrimaryCategoryID))
		if err != nil {
			return galleryerr.Validation("primaryCategoryId", "an active primary category is required")
		}
		for index, rawValueID := range classificationInput.FacetValueIDs {
			classificationInput.FacetValueIDs[index], err = DatabaseID(strings.TrimSpace(rawValueID))
			if err != nil {
				return galleryerr.Validation("facetValueIds", "facet values must be UUIDs or compact UUIDs")
			}
		}
		for index, rawTagID := range classificationInput.TagIDs {
			classificationInput.TagIDs[index], err = DatabaseID(strings.TrimSpace(rawTagID))
			if err != nil {
				return galleryerr.Validation("tagIds", "tags must be UUIDs or compact UUIDs")
			}
		}
		if err := s.classifyImageEdit(ctx, classificationInput); err != nil {
			return err
		}
	}
	return nil
}
