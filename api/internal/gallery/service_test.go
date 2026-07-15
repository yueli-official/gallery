package gallery

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"platform/gokit/classification"
	"platform/gokit/errs"
	"platform/products/gallery/api/internal/model"
)

const (
	testCategoryID = "019817c8-0000-7000-8300-000000000001"
	testFacetID    = "019817c8-0000-7000-8100-000000000001"
	testValueID    = "019817c8-0000-7000-8200-000000000003"
)

type fakeStore struct {
	settings               *model.SiteSettings
	candidates             []model.ImageCard
	image                  *model.ImageDetail
	tombstone              bool
	submission             *model.Submission
	reviewSeen             string
	submitSeen             model.SubmissionInput
	classificationSnapshot classification.Snapshot
	tagMatches             []classification.TagMatch
	filterPlanSeen         classification.FilterPlan
	governanceImpacts      []classification.ReferenceImpact
	governanceToken        string
	governanceExecuted     bool
	governanceFactRequest  classification.GovernFactRequest
	governancePlan         classification.GovernancePlan
	classificationTags     []model.ClassificationTag
	tagProposals           []model.ClassificationTagProposal
	tagProposalReviewSeen  model.ClassificationTagProposalReviewInput
}

func (f *fakeStore) SiteSettings(context.Context) (*model.SiteSettings, error) {
	return f.settings, nil
}
func (f *fakeStore) ClassificationRevision(context.Context) (uint64, error) {
	return f.classificationSnapshot.Revision, nil
}
func (f *fakeStore) ClassificationSnapshot(context.Context) (classification.Snapshot, error) {
	return f.classificationSnapshot, nil
}
func (f *fakeStore) ClassificationTagMatches(_ context.Context, _ []classification.TagLookupRequest) ([]classification.TagMatch, string, error) {
	return append([]classification.TagMatch(nil), f.tagMatches...), "tags:test", nil
}
func (f *fakeStore) ClassificationCandidateCounts(_ context.Context, _ model.ImageQuery, requests []classification.CandidateCountGroupRequest) ([]classification.CandidateCountGroup, string, error) {
	groups := make([]classification.CandidateCountGroup, 0, len(requests))
	for _, request := range requests {
		group := classification.CandidateCountGroup{Kind: request.Kind, OwnerID: request.OwnerID}
		for _, candidate := range request.Candidates {
			group.Counts = append(group.Counts, classification.CandidateCount{ValueID: candidate.ValueID, Count: 1})
		}
		groups = append(groups, group)
	}
	return groups, "counts:test", nil
}
func (f *fakeStore) ClassificationGovernanceImpacts(_ context.Context, requests []classification.ImpactRequest) ([]classification.ReferenceImpact, string, error) {
	if len(requests) == 0 {
		return []classification.ReferenceImpact{}, "", nil
	}
	return append([]classification.ReferenceImpact(nil), f.governanceImpacts...), f.governanceToken, nil
}
func (f *fakeStore) ExecuteClassificationGovernance(_ context.Context, _ string, request classification.GovernFactRequest, plan classification.GovernancePlan) (uint64, error) {
	f.governanceExecuted = true
	f.governanceFactRequest = request
	f.governancePlan = plan
	return plan.ExpectedCatalogRevision + 1, nil
}
func (f *fakeStore) ClassificationTags(_ context.Context, _ model.ClassificationTagCursor, _ int) ([]model.ClassificationTag, bool, error) {
	return append([]model.ClassificationTag(nil), f.classificationTags...), false, nil
}
func (f *fakeStore) ClassificationTagProposals(context.Context, string, int, int) ([]model.ClassificationTagProposal, int, error) {
	return append([]model.ClassificationTagProposal(nil), f.tagProposals...), len(f.tagProposals), nil
}
func (f *fakeStore) ReviewClassificationTagProposal(_ context.Context, _ string, _ string, input model.ClassificationTagProposalReviewInput) (*model.ClassificationTagProposal, bool, error) {
	f.tagProposalReviewSeen = input
	if len(f.tagProposals) == 0 {
		return nil, false, nil
	}
	return &f.tagProposals[0], true, nil
}
func (f *fakeStore) RandomCandidates(context.Context, int) ([]model.ImageCard, error) {
	return f.candidates, nil
}
func (f *fakeStore) ListImages(_ context.Context, _ model.ImageQuery, plan classification.FilterPlan) ([]model.ImageCard, int, error) {
	f.filterPlanSeen = plan
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
	store := validSubmissionStore()
	store.settings = &model.SiteSettings{Name: "Gallery", RandomBatchSize: 4, RandomCandidateSize: 40}
	store.candidates = []model.ImageCard{
		{ID: "019817c8-0000-7000-8000-000000000001", PrimaryCategorySlug: "wallpaper"},
		{ID: "019817c8-0000-7000-8000-000000000002", PrimaryCategorySlug: "wallpaper"},
		{ID: "019817c8-0000-7000-8000-000000000003", PrimaryCategorySlug: "illustration"},
		{ID: "019817c8-0000-7000-8000-000000000004", PrimaryCategorySlug: "photography"},
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
		if index > 0 && first.Images[index].PrimaryCategorySlug == first.Images[index-1].PrimaryCategorySlug {
			t.Fatalf("diversity rerank should avoid adjacent primary categories when possible: %#v", first.Images)
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
	store := validSubmissionStore()
	service := New(store)
	input := validSubmissionInput()
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
	store := validSubmissionStore()
	service := New(store)
	input := validSubmissionInput()
	input.SourceURL = "file:///etc/passwd"
	_, err := service.Submit(context.Background(), model.Subject{Kind: "user", ID: "user-1", Verified: true}, input)
	assertCode(t, err, "common.validation_failed")
	input.SourceURL = ""
	if _, err := service.Submit(context.Background(), model.Subject{Kind: "user", ID: "user-1", Verified: true}, input); err != nil {
		t.Fatalf("empty source must be accepted: %v", err)
	}
}

func TestSubmitClassifiesCategoryAndFacetBeforePersistence(t *testing.T) {
	store := validSubmissionStore()
	input := validSubmissionInput()
	if _, err := New(store).Submit(context.Background(), model.Subject{Kind: "user", ID: "user-1", Verified: true}, input); err != nil {
		t.Fatal(err)
	}
	if len(store.submitSeen.Classification.Categories) != 1 || store.submitSeen.Classification.Categories[0].CategoryID != testCategoryID {
		t.Fatalf("unexpected category assignments: %#v", store.submitSeen.Classification.Categories)
	}
	if store.submitSeen.Classification.PrimaryCategoryID != testCategoryID {
		t.Fatalf("unexpected primary category: %q", store.submitSeen.Classification.PrimaryCategoryID)
	}
	if len(store.submitSeen.Classification.Facets) != 1 || store.submitSeen.Classification.Facets[0].ValueID != testValueID {
		t.Fatalf("unexpected facet assignments: %#v", store.submitSeen.Classification.Facets)
	}
}

func TestImagesExecutesCatalogFilterPlanWithDescendants(t *testing.T) {
	store := validSubmissionStore()
	store.classificationSnapshot.Categories = append(store.classificationSnapshot.Categories,
		classification.Category{
			ID: "019817c8-0000-7000-8300-000000000002", ParentID: testCategoryID,
			Slug: "desktop", Name: "桌面壁纸", Status: classification.StatusActive,
		},
	)
	store.classificationSnapshot.FacetValues = append(store.classificationSnapshot.FacetValues,
		classification.FacetValue{
			ID: "019817c8-0000-7000-8200-000000000004", FacetID: testFacetID, ParentID: testValueID,
			Slug: "mountain", Name: "山景", Status: classification.StatusActive,
		},
	)

	_, err := New(store).Images(context.Background(), model.ImageQuery{
		CategoryRefs: []string{"wallpaper"},
		FacetRefs:    []string{"scene:landscape"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(store.filterPlanSeen.Groups, []classification.FilterGroup{
		{
			Kind: classification.FilterGroupCategory,
			ValueIDs: []string{
				"019817c8-0000-7000-8300-000000000001",
				"019817c8-0000-7000-8300-000000000002",
			},
		},
		{
			Kind:    classification.FilterGroupFacet,
			OwnerID: testFacetID,
			ValueIDs: []string{
				"019817c8-0000-7000-8200-000000000003",
				"019817c8-0000-7000-8200-000000000004",
			},
		},
	}) {
		t.Fatalf("filter plan = %#v", store.filterPlanSeen)
	}
}

func TestGovernancePreviewBuildsPlanFromFreshImpactFacts(t *testing.T) {
	store := validSubmissionStore()
	store.governanceToken = "impact:category"
	store.governanceImpacts = []classification.ReferenceImpact{
		{Kind: classification.GovernCategory, ID: testCategoryID, Exists: true, Status: classification.StatusActive, AssignmentCount: 4, PrimaryCount: 3},
	}
	preview, err := New(store).PreviewClassificationGovernance(context.Background(), model.ClassificationGovernancePreviewInput{
		Command: model.ClassificationGovernanceCommand{
			Operation: "delete", Kind: "category", ID: testCategoryID, DeleteAllRelated: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if preview.Outcome != "planned" || preview.Plan.ExpectedRequestToken == "" || preview.Plan.ExpectedImpactToken != "impact:category" {
		t.Fatalf("preview = %#v", preview)
	}
	if len(preview.Plan.Steps) != 3 || preview.Plan.Steps[0].Kind != "clear_primary_assignments" {
		t.Fatalf("steps = %#v", preview.Plan.Steps)
	}
}

func TestClassificationCatalogIncludesInactiveManagementIdentities(t *testing.T) {
	store := validSubmissionStore()
	store.classificationSnapshot.Categories = append(store.classificationSnapshot.Categories, classification.Category{
		ID: "019817c8-0000-7000-8300-000000000099", Slug: "archive", Name: "归档", Status: classification.StatusInactive,
	})
	catalog, err := New(store).ClassificationCatalog(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Revision != 1 || len(catalog.Categories) != 2 || catalog.Categories[1].Status != "inactive" {
		t.Fatalf("catalog = %#v", catalog)
	}
	if len(catalog.Facets) != 1 || len(catalog.Facets[0].Values) != 1 {
		t.Fatalf("facets = %#v", catalog.Facets)
	}
}

func TestGovernanceExecuteRecomputesPlanAndRejectsStaleEnvelope(t *testing.T) {
	store := validSubmissionStore()
	store.governanceToken = "impact:current"
	store.governanceImpacts = []classification.ReferenceImpact{
		{Kind: classification.GovernCategory, ID: testCategoryID, Exists: true, Status: classification.StatusActive},
	}
	service := New(store)
	_, err := service.ExecuteClassificationGovernance(context.Background(), "operator-1", model.ClassificationGovernanceExecuteInput{
		Command:                 model.ClassificationGovernanceCommand{Operation: "set_status", Kind: "category", ID: testCategoryID, Status: "inactive"},
		ExpectedCatalogRevision: 99,
		ExpectedRequestToken:    "preview-token",
	})
	assertCode(t, err, "gallery.conflict")
	if store.governanceExecuted {
		t.Fatal("stale preview must not reach the transaction executor")
	}
}

func TestGovernanceExecuteUsesCanonicalPlanAndInvalidatesCatalog(t *testing.T) {
	store := validSubmissionStore()
	store.governanceToken = "impact:current"
	store.governanceImpacts = []classification.ReferenceImpact{
		{Kind: classification.GovernCategory, ID: testCategoryID, Exists: true, Status: classification.StatusActive},
	}
	service := New(store)
	if _, err := service.classificationCatalog(context.Background()); err != nil {
		t.Fatal(err)
	}
	command := model.ClassificationGovernanceCommand{Operation: "set_status", Kind: "category", ID: testCategoryID, Status: "inactive"}
	preview, err := service.PreviewClassificationGovernance(context.Background(), model.ClassificationGovernancePreviewInput{Command: command})
	if err != nil {
		t.Fatal(err)
	}
	execution, err := service.ExecuteClassificationGovernance(context.Background(), "operator-1", model.ClassificationGovernanceExecuteInput{
		Command:                 command,
		ExpectedCatalogRevision: 1,
		ExpectedRequestToken:    preview.Plan.ExpectedRequestToken,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !execution.Applied || execution.CatalogRevision != 2 || !store.governanceExecuted {
		t.Fatalf("execution = %#v", execution)
	}
	if len(store.governancePlan.Steps) != 1 || store.governancePlan.Steps[0].Kind != classification.GovernChangeStatus {
		t.Fatalf("plan = %#v", store.governancePlan)
	}
	if service.catalog != nil || service.catalogRevision != 0 {
		t.Fatal("successful governance must invalidate the local catalog immediately")
	}
}

func TestGovernanceExecuteRejectsACommandSwappedAfterPreview(t *testing.T) {
	service := New(validSubmissionStore())
	preview, err := service.PreviewClassificationGovernance(context.Background(), model.ClassificationGovernancePreviewInput{
		Command: model.ClassificationGovernanceCommand{Operation: "set_status", Kind: "category", ID: testCategoryID, Status: "inactive"},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.ExecuteClassificationGovernance(context.Background(), "operator-1", model.ClassificationGovernanceExecuteInput{
		Command:                 model.ClassificationGovernanceCommand{Operation: "reparent", Kind: "category", ID: testCategoryID},
		ExpectedCatalogRevision: preview.Plan.ExpectedCatalogRevision,
		ExpectedRequestToken:    preview.Plan.ExpectedRequestToken,
	})
	assertCode(t, err, "gallery.conflict")
}

func TestClassificationTagsRejectsMalformedCursor(t *testing.T) {
	_, err := New(validSubmissionStore()).ClassificationTags(context.Background(), "not-a-cursor", 20)
	assertCode(t, err, "common.validation_failed")
}

func TestReviewTagProposalNormalizesTargetAndInvalidatesChangedCatalog(t *testing.T) {
	store := validSubmissionStore()
	store.tagProposals = []model.ClassificationTagProposal{{ID: "019817c8-0000-7000-8400-000000000001"}}
	service := New(store)
	if _, err := service.classificationCatalog(context.Background()); err != nil {
		t.Fatal(err)
	}
	target := "019817c8-0000-7000-8500-000000000001"
	_, err := service.ReviewClassificationTagProposal(context.Background(), "operator-1", "019817c8-0000-7000-8400-000000000001", model.ClassificationTagProposalReviewInput{
		Decision: "approve", TargetTagID: target,
	})
	if err != nil {
		t.Fatal(err)
	}
	if store.tagProposalReviewSeen.TargetTagID != target {
		t.Fatalf("review input = %#v", store.tagProposalReviewSeen)
	}
	if service.catalog != nil {
		t.Fatal("catalog-changing tag review must invalidate the catalog")
	}
}

func validSubmissionStore() *fakeStore {
	return &fakeStore{
		submission: &model.Submission{ID: "019817c8-0000-7000-8000-000000000001"},
		classificationSnapshot: classification.Snapshot{
			CatalogID: "gallery",
			Revision:  1,
			Categories: []classification.Category{
				{ID: testCategoryID, Slug: "wallpaper", Name: "壁纸", Status: classification.StatusActive},
			},
			Facets: []classification.Facet{
				{ID: testFacetID, Slug: "scene", Name: "场景", Status: classification.StatusActive},
			},
			FacetValues: []classification.FacetValue{
				{ID: testValueID, FacetID: testFacetID, Slug: "landscape", Name: "风景", Status: classification.StatusActive},
			},
			Policies: []classification.PolicyProfile{
				{
					Key: "gallery.image.public", SchemaVersion: 1, PolicyRevision: 1,
					Category: classification.CategoryPolicy{MinAssignments: 1, MaxAssignments: 3, RequirePrimary: true},
					Facets:   []classification.FacetAssignmentPolicy{{FacetID: testFacetID, MinValues: 1, MaxValues: 2}},
					Tags:     classification.TagAdmissionPolicy{Unknown: classification.UnknownTagPropose},
				},
			},
		},
	}
}

func validSubmissionInput() model.SubmissionInput {
	return model.SubmissionInput{
		AssetID:           "019817c8-0000-7000-8000-000000000010",
		Title:             "雨夜",
		CategoryIDs:       []string{testCategoryID},
		PrimaryCategoryID: testCategoryID,
		Facets:            []model.FacetSelection{{FacetID: testFacetID, ValueIDs: []string{testValueID}}},
	}
}

func TestDefaultDiscoverySeedUsesUTCDate(t *testing.T) {
	store := validSubmissionStore()
	store.settings = &model.SiteSettings{Name: "Gallery"}
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
