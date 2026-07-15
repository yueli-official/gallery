package gallery

import (
	"context"
	"errors"
	"testing"
	"time"

	"platform/gokit/errs"
	"platform/gokit/facet"
	"platform/products/gallery/api/internal/model"
)

type fakeStore struct {
	settings   *model.SiteSettings
	candidates []model.ImageCard
	image      *model.ImageDetail
	tombstone  bool
	submission *model.Submission
	reviewSeen string
	submitSeen model.SubmissionInput
}

func (f *fakeStore) SiteSettings(context.Context) (*model.SiteSettings, error) {
	return f.settings, nil
}
func (f *fakeStore) Facets(context.Context) ([]facet.Facet, error) {
	return []facet.Facet{{
		ID: "019817c8-0000-7000-8000-000000000021", Slug: "topic", Name: "主题",
		SelectionMode: facet.SelectionSingle, RequiredOnPublish: true, Filterable: true, Status: facet.StatusActive,
	}}, nil
}
func (f *fakeStore) FacetValues(context.Context) ([]facet.Value, error) {
	return []facet.Value{{
		ID: "019817c8-0000-7000-8000-000000000020", FacetID: "019817c8-0000-7000-8000-000000000021",
		Slug: "night", Name: "夜景", Status: facet.StatusActive,
	}}, nil
}
func (f *fakeStore) RandomCandidates(context.Context, int) ([]model.ImageCard, error) {
	return f.candidates, nil
}
func (f *fakeStore) ListImages(context.Context, model.ImageQuery) ([]model.ImageCard, int, error) {
	return f.candidates, len(f.candidates), nil
}
func (f *fakeStore) Image(context.Context, string, string) (*model.ImageDetail, error) {
	return f.image, nil
}
func (f *fakeStore) HasTombstone(context.Context, string) (bool, error) { return f.tombstone, nil }
func (f *fakeStore) PublicCollections(context.Context) ([]model.Collection, error) {
	return nil, nil
}
func (f *fakeStore) EditorialCollections(context.Context) ([]model.Collection, error) {
	return nil, nil
}
func (f *fakeStore) PublicCollection(context.Context, string, int, int) (*model.CollectionDetail, error) {
	return nil, nil
}
func (f *fakeStore) CreateEditorialCollection(context.Context, string, model.EditorialCollectionInput) (*model.Collection, error) {
	return &model.Collection{ID: "019817c8-0000-7000-8000-000000000030"}, nil
}
func (f *fakeStore) Ranking(context.Context, string, string, int) (*model.Ranking, error) {
	return nil, nil
}
func (f *fakeStore) CreateSubmission(_ context.Context, _ model.Subject, input model.SubmissionInput, review string) (*model.Submission, error) {
	f.reviewSeen = review
	f.submitSeen = input
	return f.submission, nil
}
func (f *fakeStore) MySubmissions(context.Context, model.Subject, int, int) ([]model.Submission, int, error) {
	return nil, 0, nil
}
func (f *fakeStore) WithdrawSubmission(context.Context, model.Subject, string) (*model.Submission, error) {
	return f.submission, nil
}
func (f *fakeStore) FailSubmission(context.Context, string, string) error { return nil }
func (f *fakeStore) CollectionDetail(context.Context, string, int, int) (*model.CollectionDetail, error) {
	return nil, nil
}
func (f *fakeStore) CreateCase(context.Context, model.Subject, string, model.CaseInput) (*model.Case, error) {
	return nil, nil
}
func (f *fakeStore) AdminOverview(context.Context) (*model.AdminOverview, error) {
	return &model.AdminOverview{}, nil
}
func (f *fakeStore) ReviewQueue(context.Context, int, int) ([]model.Submission, int, error) {
	return nil, 0, nil
}
func (f *fakeStore) ReviewSubmission(context.Context, string, string, model.SubmissionReviewInput) (*model.Submission, error) {
	return f.submission, nil
}
func (f *fakeStore) HideImage(context.Context, string, string, string) error { return nil }
func (f *fakeStore) Cases(context.Context, string, int, int) ([]model.Case, int, error) {
	return nil, 0, nil
}
func (f *fakeStore) ResolveCase(context.Context, string, string, model.CaseResolutionInput) (*model.Case, error) {
	return nil, nil
}
func (f *fakeStore) RecordEvent(context.Context, model.Subject, string, model.EventInput) error {
	return nil
}

func TestDiscoveryRequiresSettings(t *testing.T) {
	_, err := New(&fakeStore{}).Discovery(context.Background(), "seed")
	assertCode(t, err, "gallery.not_initialized")
}

func TestSeededDiscoveryIsStableAndAvoidsAdjacentTopics(t *testing.T) {
	store := &fakeStore{
		settings: &model.SiteSettings{Name: "Gallery", RandomBatchSize: 4, RandomCandidateSize: 40},
		candidates: []model.ImageCard{
			{ID: "019817c8-0000-7000-8000-000000000001", TopicSlug: "people"},
			{ID: "019817c8-0000-7000-8000-000000000002", TopicSlug: "people"},
			{ID: "019817c8-0000-7000-8000-000000000003", TopicSlug: "landscape"},
			{ID: "019817c8-0000-7000-8000-000000000004", TopicSlug: "animals"},
		},
	}
	first, err := New(store).Discovery(context.Background(), "same-seed")
	if err != nil {
		t.Fatal(err)
	}
	second, err := New(store).Discovery(context.Background(), "same-seed")
	if err != nil {
		t.Fatal(err)
	}
	for index := range first.Images {
		if first.Images[index].ID != second.Images[index].ID {
			t.Fatalf("same seed must be stable: %#v %#v", first.Images, second.Images)
		}
		if index > 0 && first.Images[index].TopicSlug == first.Images[index-1].TopicSlug {
			t.Fatalf("diversity rerank should avoid adjacent topics when possible: %#v", first.Images)
		}
	}
}

func TestImageDistinguishesTombstoneFromHidden(t *testing.T) {
	store := &fakeStore{tombstone: true}
	_, err := New(store).Image(context.Background(), "AZgXyAAAcACAAAAAAAAAAQ", "")
	assertCode(t, err, "gallery.gone")
	store.tombstone = false
	_, err = New(store).Image(context.Background(), "AZgXyAAAcACAAAAAAAAAAQ", "")
	assertCode(t, err, "gallery.not_found")
}

func TestVerifiedUserSkipsManualReviewButGuestDoesNot(t *testing.T) {
	store := &fakeStore{submission: &model.Submission{ID: "019817c8-0000-7000-8000-000000000001"}}
	service := New(store)
	input := model.SubmissionInput{
		AssetID: "019817c8-0000-7000-8000-000000000010", TopicID: "019817c8-0000-7000-8000-000000000020", Title: "雨夜",
	}
	if _, err := service.Submit(context.Background(), model.Subject{Kind: "user", ID: "user-1", Verified: true}, input); err != nil {
		t.Fatal(err)
	}
	if store.reviewSeen != "not_required" {
		t.Fatalf("verified user should skip manual review, got %q", store.reviewSeen)
	}
	if _, err := service.Submit(context.Background(), model.Subject{Kind: "guest", ID: "guest-1"}, input); err != nil {
		t.Fatal(err)
	}
	if store.reviewSeen != "pending" {
		t.Fatalf("guest must require review, got %q", store.reviewSeen)
	}
}

func TestSourceURLIsOptionalButMustBeHTTP(t *testing.T) {
	store := &fakeStore{submission: &model.Submission{ID: "019817c8-0000-7000-8000-000000000001"}}
	service := New(store)
	input := model.SubmissionInput{AssetID: "019817c8-0000-7000-8000-000000000010", TopicID: "019817c8-0000-7000-8000-000000000020", Title: "雨夜", SourceURL: "file:///etc/passwd"}
	_, err := service.Submit(context.Background(), model.Subject{Kind: "user", ID: "user-1", Verified: true}, input)
	assertCode(t, err, "common.validation_failed")
	input.SourceURL = ""
	if _, err := service.Submit(context.Background(), model.Subject{Kind: "user", ID: "user-1", Verified: true}, input); err != nil {
		t.Fatalf("empty source must be accepted: %v", err)
	}
}

func TestSubmitNormalizesTopicAndTagsBeforePersistence(t *testing.T) {
	store := &fakeStore{submission: &model.Submission{ID: "019817c8-0000-7000-8000-000000000001"}}
	input := model.SubmissionInput{
		AssetID: "019817c8-0000-7000-8000-000000000010",
		TopicID: "019817c8-0000-7000-8000-000000000020",
		Title:   "雨夜", Tags: []string{"  Night Sky ", "night   sky", "雨夜"},
	}
	if _, err := New(store).Submit(context.Background(), model.Subject{Kind: "user", ID: "user-1", Verified: true}, input); err != nil {
		t.Fatal(err)
	}
	if len(store.submitSeen.Assignments) != 1 || store.submitSeen.Assignments[0].ValueID != input.TopicID {
		t.Fatalf("unexpected assignments: %#v", store.submitSeen.Assignments)
	}
	if len(store.submitSeen.NormalizedTags) != 2 || store.submitSeen.NormalizedTags[0].Slug != "night-sky" || store.submitSeen.NormalizedTags[1].Slug != "雨夜" {
		t.Fatalf("unexpected normalized tags: %#v", store.submitSeen.NormalizedTags)
	}
}

func TestDefaultDiscoverySeedUsesUTCDate(t *testing.T) {
	store := &fakeStore{settings: &model.SiteSettings{Name: "Gallery"}}
	service := New(store)
	service.clock = func() time.Time { return time.Date(2026, 7, 15, 23, 0, 0, 0, time.FixedZone("CST", 8*3600)) }
	value, err := service.Discovery(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if value.Seed != "2026-07-15" {
		t.Fatalf("unexpected UTC seed %q", value.Seed)
	}
}

func assertCode(t *testing.T, err error, want string) {
	t.Helper()
	var coded *errs.Coded
	if !errors.As(err, &coded) || coded.Code != want {
		t.Fatalf("expected %s, got %v", want, err)
	}
}
