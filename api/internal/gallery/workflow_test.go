package gallery

import (
	"context"
	"errors"
	"testing"

	"platform/gokit/errs"
	"platform/gokit/facet"
	"platform/products/gallery/api/internal/model"
)

type workflowStore struct {
	fakeStore
	creator     *model.CreatorProfile
	artwork     *model.StudioArtwork
	assets      int
	assignments []facet.Assignment
	savedStatus string
}

func (s *workflowStore) CreatorBySubject(context.Context, string) (*model.CreatorProfile, error) {
	return s.creator, nil
}

func (s *workflowStore) CreateCreator(context.Context, model.CreatorRequest) (*model.CreatorProfile, error) {
	return s.creator, nil
}

func (s *workflowStore) CreatorArtworks(context.Context, string) ([]model.StudioArtwork, error) {
	return nil, nil
}

func (s *workflowStore) CreateArtwork(context.Context, string) (*model.StudioArtwork, error) {
	return s.artwork, nil
}

func (s *workflowStore) CreatorArtwork(context.Context, string, string) (*model.StudioArtwork, error) {
	return s.artwork, nil
}

func (s *workflowStore) SaveArtwork(context.Context, string, model.ArtworkDraftInput, []facet.Assignment) (*model.StudioArtwork, error) {
	return s.artwork, nil
}

func (s *workflowStore) AddArtworkAsset(context.Context, string, model.AssetInput) (*model.Asset, error) {
	return nil, nil
}

func (s *workflowStore) RemoveArtworkAsset(context.Context, string, string) (*model.Asset, error) {
	return nil, nil
}

func (s *workflowStore) ArtworkAssetCount(context.Context, string) (int, error) {
	return s.assets, nil
}

func (s *workflowStore) ArtworkAssignments(context.Context, string) ([]facet.Assignment, error) {
	return s.assignments, nil
}

func (s *workflowStore) SetArtworkStatus(_ context.Context, _ string, status string, _ string, _ string) (*model.StudioArtwork, error) {
	s.savedStatus = status
	if s.artwork != nil {
		s.artwork.Status = status
	}
	return s.artwork, nil
}

func (s *workflowStore) CreatorApplications(context.Context) ([]model.CreatorProfile, error) {
	return nil, nil
}

func (s *workflowStore) SetCreatorStatus(context.Context, string, string, string, string) (*model.CreatorProfile, error) {
	return s.creator, nil
}

func (s *workflowStore) ReviewQueue(context.Context, string) ([]model.StudioArtwork, error) {
	return nil, nil
}

func TestCreateDraftRequiresActiveCreator(t *testing.T) {
	store := &workflowStore{creator: &model.CreatorProfile{ID: "creator-1", Status: "pending"}}
	_, err := New(store).CreateDraft(context.Background(), "account-1")
	assertCode(t, err, "gallery.creator_not_active")
}

func TestCreatorRequestRejectsInvalidHandle(t *testing.T) {
	store := &workflowStore{}
	_, err := New(store).RequestCreator(context.Background(), "account-1", model.CreatorRequest{Handle: "Not valid", DisplayName: "创作者"})
	assertCode(t, err, "common.validation_failed")
}

func TestDraftCanBeSavedBeforePublishRequirementsAreComplete(t *testing.T) {
	store := &workflowStore{
		creator: &model.CreatorProfile{ID: "creator-1", Status: "active"},
		artwork: &model.StudioArtwork{ID: "artwork-1", CreatorID: "creator-1", Status: "draft"},
		fakeStore: fakeStore{
			facets: []facet.Facet{{ID: "medium", Slug: "medium", Name: "媒介", SelectionMode: facet.SelectionSingle, RequiredOnPublish: true, Status: facet.StatusActive}},
			values: []facet.Value{{ID: "illustration", FacetID: "medium", Slug: "illustration", Name: "插画", Status: facet.StatusActive}},
		},
	}
	_, err := New(store).SaveDraft(context.Background(), "account-1", "artwork-1", model.ArtworkDraftInput{})
	if err != nil {
		t.Fatalf("draft must allow incomplete publish fields: %v", err)
	}
}

func TestDraftRejectsUnsupportedControlledValues(t *testing.T) {
	store := &workflowStore{
		creator: &model.CreatorProfile{ID: "creator-1", Status: "active"},
		artwork: &model.StudioArtwork{ID: "artwork-1", CreatorID: "creator-1", Status: "draft"},
	}
	_, err := New(store).SaveDraft(context.Background(), "account-1", "artwork-1", model.ArtworkDraftInput{Visibility: "everyone"})
	assertCode(t, err, "common.validation_failed")
}

func TestSubmitRequiresTitleAssetAndRequiredFacets(t *testing.T) {
	base := workflowStore{
		creator: &model.CreatorProfile{ID: "creator-1", Status: "active"},
		artwork: &model.StudioArtwork{ID: "artwork-1", CreatorID: "creator-1", Status: "draft"},
		fakeStore: fakeStore{
			facets: []facet.Facet{{ID: "medium", Slug: "medium", Name: "媒介", SelectionMode: facet.SelectionSingle, RequiredOnPublish: true, Status: facet.StatusActive}},
			values: []facet.Value{{ID: "illustration", FacetID: "medium", Slug: "illustration", Name: "插画", Status: facet.StatusActive}},
		},
	}

	_, err := New(&base).SubmitArtwork(context.Background(), "account-1", "artwork-1")
	assertCode(t, err, "common.validation_failed")

	base.artwork.Title = "雨夜"
	_, err = New(&base).SubmitArtwork(context.Background(), "account-1", "artwork-1")
	assertCode(t, err, "common.validation_failed")

	base.assets = 1
	_, err = New(&base).SubmitArtwork(context.Background(), "account-1", "artwork-1")
	assertCode(t, err, "common.validation_failed")

	base.assignments = []facet.Assignment{{FacetID: "medium", ValueID: "illustration"}}
	_, err = New(&base).SubmitArtwork(context.Background(), "account-1", "artwork-1")
	if err != nil {
		t.Fatal(err)
	}
	if base.savedStatus != "pending_review" {
		t.Fatalf("expected pending_review, got %q", base.savedStatus)
	}
}

func TestReviewOnlyAcceptsPendingArtwork(t *testing.T) {
	store := &workflowStore{artwork: &model.StudioArtwork{ID: "artwork-1", Status: "draft"}}
	_, err := New(store).ReviewArtwork(context.Background(), "operator-1", "artwork-1", model.ReviewInput{Decision: "approve"})
	assertCode(t, err, "gallery.invalid_state")

	store.artwork.Status = "pending_review"
	_, err = New(store).ReviewArtwork(context.Background(), "operator-1", "artwork-1", model.ReviewInput{Decision: "approve"})
	if err != nil {
		t.Fatal(err)
	}
	if store.savedStatus != "published" {
		t.Fatalf("expected published, got %q", store.savedStatus)
	}
}

func assertCode(t *testing.T, err error, want string) {
	t.Helper()
	var coded *errs.Coded
	if !errors.As(err, &coded) || coded.Code != want {
		t.Fatalf("expected %s, got %v", want, err)
	}
}
