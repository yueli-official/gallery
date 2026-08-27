package gallery

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/yueli-official/foundation/go/classification"
	"github.com/yueli-official/foundation/go/problem"
	"github.com/yueli-official/gallery/api/internal/collection"
	"github.com/yueli-official/gallery/api/internal/model"
)

const (
	testCategoryID = "019817c8-0000-7000-8300-000000000001"
	testFacetID    = "019817c8-0000-7000-8100-000000000001"
	testValueID    = "019817c8-0000-7000-8200-000000000003"
)

type guestClaimTestStore struct {
	Store
	guest string
	user  string
}

type submissionProcessingTestStore struct {
	*fakeStore
	claimed       []model.Submission
	completedID   string
	completedFact model.SubmissionAssetFacts
	assetID       string
}

type publicationReconciliationStore struct {
	*fakeStore
	candidates []model.ImagePublicationCandidate
}

func (store *publicationReconciliationStore) PublishedImageCandidates(_ context.Context, afterID string, limit int) ([]model.ImagePublicationCandidate, error) {
	start := 0
	if afterID != "" {
		for index, item := range store.candidates {
			if item.ID == afterID {
				start = index + 1
				break
			}
		}
	}
	end := min(start+limit, len(store.candidates))
	return append([]model.ImagePublicationCandidate(nil), store.candidates[start:end]...), nil
}

func (store *submissionProcessingTestStore) ClaimSubmissionProcessing(context.Context, int) ([]model.Submission, error) {
	return append([]model.Submission(nil), store.claimed...), nil
}

func (store *submissionProcessingTestStore) CompleteSubmissionProcessing(_ context.Context, id string, facts model.SubmissionAssetFacts) error {
	store.completedID = id
	store.completedFact = facts
	return nil
}

func (store *submissionProcessingTestStore) SubmissionAssetID(context.Context, string) (string, error) {
	return store.assetID, nil
}

type submissionProcessingAssetPort struct {
	preparedAssetID  string
	publishedAssetID string
	publishedImageID string
	facts            model.SubmissionAssetFacts
}

func (*submissionProcessingAssetPort) RegisterSubmission(context.Context, string, string, string, string) error {
	return nil
}

func (*submissionProcessingAssetPort) UnregisterSubmission(context.Context, string, string, string) error {
	return nil
}

func (port *submissionProcessingAssetPort) PrepareSubmission(_ context.Context, assetID string) (model.SubmissionAssetFacts, error) {
	port.preparedAssetID = assetID
	return port.facts, nil
}

func (port *submissionProcessingAssetPort) PublishImage(_ context.Context, assetID, imageID, _ string) error {
	port.publishedAssetID = assetID
	port.publishedImageID = imageID
	return nil
}

func TestProcessQueuedSubmissionPersistsAssetFactsForReviewAndDedup(t *testing.T) {
	facts := model.SubmissionAssetFacts{
		PreviewURL: "http://asset.test/api/v1/assets/blob/signed", ContentHash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Mime: "image/png", Width: 1600, Height: 900,
	}
	store := &submissionProcessingTestStore{fakeStore: &fakeStore{}, claimed: []model.Submission{{ID: "submission-1", AssetID: "asset-1"}}}
	assets := &submissionProcessingAssetPort{facts: facts}
	service := New(store)
	service.SetAssetReferencePort(assets)
	processed, err := service.ProcessQueuedSubmissions(context.Background(), 8)
	if err != nil {
		t.Fatal(err)
	}
	if processed != 1 || assets.preparedAssetID != "asset-1" || store.completedID != "submission-1" || store.completedFact != facts {
		t.Fatalf("processed = %d, asset = %q, completed = %q %#v", processed, assets.preparedAssetID, store.completedID, store.completedFact)
	}
}

func TestSubmissionPreviewUsesSignedAssetRendition(t *testing.T) {
	store := &submissionProcessingTestStore{fakeStore: &fakeStore{}, assetID: "asset-1"}
	assets := &submissionProcessingAssetPort{facts: model.SubmissionAssetFacts{PreviewURL: "http://asset.test/api/v1/assets/blob/signed"}}
	service := New(store)
	service.SetAssetReferencePort(assets)
	preview, err := service.SubmissionPreviewURL(context.Background(), "019817c8-0000-7000-8300-000000000099")
	if err != nil {
		t.Fatal(err)
	}
	if preview != assets.facts.PreviewURL || assets.preparedAssetID != "asset-1" {
		t.Fatalf("preview = %q, asset = %q", preview, assets.preparedAssetID)
	}
}

func (store *guestClaimTestStore) ClaimGuestSubmissions(_ context.Context, guestSubject, userID string) (int64, error) {
	store.guest = guestSubject
	store.user = userID
	return 3, nil
}

func TestClaimGuestSubmissionsTransfersOwnership(t *testing.T) {
	store := &guestClaimTestStore{}
	service := New(store)
	claimed, err := service.ClaimGuestSubmissions(context.Background(), "guest-1", "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if claimed != 3 || store.guest != "guest-1" || store.user != "user-1" {
		t.Fatalf("claim = %d, guest = %q, user = %q", claimed, store.guest, store.user)
	}
}

type fakeStore struct {
	settings                 *model.SiteSettings
	settingsUpdateSeen       model.SiteSettingsUpdateInput
	candidates               []model.ImageCard
	image                    *model.ImageDetail
	relatedImages            []model.RelatedImage
	relatedLimitSeen         int
	collectionDetail         *model.CollectionDetail
	collectionUpdateSeen     model.EditorialCollectionUpdateInput
	collectionOrderSeen      []string
	tombstone                bool
	submission               *model.Submission
	reviewSeen               string
	submitSeen               model.SubmissionInput
	classificationSnapshot   classification.Snapshot
	classificationCreateSeen model.ClassificationIdentityCreateInput
	classificationUpdateSeen model.ClassificationIdentityUpdateInput
	tagMatches               []classification.TagMatch
	candidateCountCalls      int
	filterPlanSeen           classification.FilterPlan
	governanceImpacts        []classification.ReferenceImpact
	governanceToken          string
	governanceExecuted       bool
	governanceFactRequest    classification.GovernFactRequest
	governancePlan           classification.GovernancePlan
	classificationTags       []model.ClassificationTag
	tagProposals             []model.ClassificationTagProposal
	tagProposalReviewSeen    model.ClassificationTagProposalReviewInput
	mySubmissionsQuerySeen   model.MySubmissionQuery
	adminImages              []model.AdminImage
	adminImageQuerySeen      model.AdminImageQuery
	adminImageUpdateSeen     model.AdminImageUpdateInput
	adminImageDeleteSeen     string
	adminImageCounts         map[string]int
	primaryCategorySeen      string
	adminSubmissions         []model.Submission
	adminSubmissionQuery     model.AdminSubmissionQuery
	adminCases               []model.Case
	adminCaseQuery           model.AdminCaseQuery
	caseResolutionSeen       model.CaseResolutionInput
	publicRenditionReadyID   string
	adminOverview            *model.AdminOverview
	adminOverviewDays        int
}

func (f *fakeStore) SiteSettings(context.Context) (*model.SiteSettings, error) {
	return f.settings, nil
}
func (f *fakeStore) UpdateSiteSettings(_ context.Context, input model.SiteSettingsUpdateInput) (*model.SiteSettings, error) {
	f.settingsUpdateSeen = input
	f.settings = &model.SiteSettings{
		Name:                input.Name,
		Title:               input.Title,
		Description:         input.Description,
		SearchPlaceholder:   input.SearchPlaceholder,
		FooterTagline:       input.FooterTagline,
		RandomCandidateSize: input.RandomCandidateSize,
		HomeSections:        append([]model.HomeSectionSettings(nil), input.HomeSections...),
	}
	for _, section := range input.HomeSections {
		if section.Key == "random" {
			f.settings.RandomBatchSize = section.ItemLimit
			break
		}
	}
	return f.settings, nil
}
func (f *fakeStore) ClassificationRevision(context.Context) (uint64, error) {
	return f.classificationSnapshot.Revision, nil
}
func (f *fakeStore) ClassificationSnapshot(context.Context) (classification.Snapshot, error) {
	return f.classificationSnapshot, nil
}
func (f *fakeStore) CreateClassificationIdentity(_ context.Context, _ string, input model.ClassificationIdentityCreateInput) (*model.ClassificationIdentityCreateResult, error) {
	f.classificationCreateSeen = input
	return &model.ClassificationIdentityCreateResult{ID: testCategoryID, CatalogRevision: 9}, nil
}
func (f *fakeStore) UpdateClassificationIdentity(_ context.Context, _ string, _ string, input model.ClassificationIdentityUpdateInput) (*model.ClassificationIdentityCreateResult, error) {
	f.classificationUpdateSeen = input
	return &model.ClassificationIdentityCreateResult{ID: testCategoryID, CatalogRevision: 10}, nil
}
func (f *fakeStore) ClassificationTagMatches(_ context.Context, _ []classification.TagLookupRequest) ([]classification.TagMatch, string, error) {
	return append([]classification.TagMatch(nil), f.tagMatches...), "tags:test", nil
}
func (f *fakeStore) ClassificationCandidateCounts(_ context.Context, _ model.ImageQuery, requests []classification.CandidateCountGroupRequest) ([]classification.CandidateCountGroup, string, error) {
	f.candidateCountCalls++
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
func (f *fakeStore) ListImages(_ context.Context, query model.ImageQuery, plan classification.FilterPlan) ([]model.ImageCard, int, error) {
	f.filterPlanSeen = plan
	start := (query.Page - 1) * query.PageSize
	if start >= len(f.candidates) {
		return []model.ImageCard{}, len(f.candidates), nil
	}
	end := min(start+query.PageSize, len(f.candidates))
	return f.candidates[start:end], len(f.candidates), nil
}
func (f *fakeStore) Image(context.Context, string, string) (*model.ImageDetail, error) {
	return f.image, nil
}
func (f *fakeStore) RelatedImages(_ context.Context, _ string, limit int) ([]model.RelatedImage, error) {
	f.relatedLimitSeen = limit
	return append([]model.RelatedImage(nil), f.relatedImages...), nil
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
func (f *fakeStore) UpdateEditorialCollection(_ context.Context, _ string, input model.EditorialCollectionUpdateInput) (*model.Collection, error) {
	f.collectionUpdateSeen = input
	return &model.Collection{ID: "019817c8-0000-7000-8000-000000000030"}, nil
}
func (f *fakeStore) ReorderEditorialMembers(_ context.Context, _ string, _ int64, ids []string) (*model.Collection, error) {
	f.collectionOrderSeen = append([]string(nil), ids...)
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
func (f *fakeStore) MySubmissions(_ context.Context, _ model.Subject, query model.MySubmissionQuery) ([]model.Submission, int, error) {
	f.mySubmissionsQuerySeen = query
	return nil, 0, nil
}
func (f *fakeStore) WithdrawSubmission(context.Context, model.Subject, string) (*model.Submission, error) {
	return f.submission, nil
}
func (f *fakeStore) FailSubmission(context.Context, string, string) error { return nil }
func (f *fakeStore) CollectionDetail(context.Context, string, int, int) (*model.CollectionDetail, error) {
	return f.collectionDetail, nil
}
func (f *fakeStore) FavoritesDetail(context.Context, string, int, int, string) (*model.CollectionDetail, error) {
	return f.collectionDetail, nil
}

type fakeCollectionStore struct {
	*fakeStore
	favoritePageSeen int
	favoriteSizeSeen int
	favoriteSortSeen string
}

func (f *fakeCollectionStore) FindSingleton(context.Context, string, string) (*model.Collection, error) {
	return &model.Collection{ID: testCategoryID, Kind: collection.KindFavorites, OwnerKind: "user", OwnerID: "user-1", Version: 3}, nil
}
func (f *fakeCollectionStore) CreateCollection(context.Context, collection.CreateInput) (*model.Collection, error) {
	return nil, nil
}
func (f *fakeCollectionStore) OwnedCollection(context.Context, string, string) (*model.Collection, error) {
	return nil, nil
}
func (f *fakeCollectionStore) MutateMembers(context.Context, string, int64, []string, []string) (*model.Collection, error) {
	return nil, nil
}
func (f *fakeCollectionStore) Collectable(context.Context, []string) (map[string]bool, error) {
	return map[string]bool{}, nil
}
func (f *fakeCollectionStore) FavoritesDetail(_ context.Context, _ string, page, size int, sort string) (*model.CollectionDetail, error) {
	f.favoritePageSeen, f.favoriteSizeSeen, f.favoriteSortSeen = page, size, sort
	return f.collectionDetail, nil
}
func (f *fakeStore) CreateCase(context.Context, model.Subject, string, model.CaseInput) (*model.Case, error) {
	return nil, nil
}
func (f *fakeStore) AdminOverview(_ context.Context, days int) (*model.AdminOverview, error) {
	f.adminOverviewDays = days
	if f.adminOverview != nil {
		return f.adminOverview, nil
	}
	return &model.AdminOverview{}, nil
}
func (f *fakeStore) AdminImages(_ context.Context, query model.AdminImageQuery) ([]model.AdminImage, int, error) {
	f.adminImageQuerySeen = query
	return f.adminImages, len(f.adminImages), nil
}
func (f *fakeStore) AdminImage(context.Context, string) (*model.AdminImage, error) {
	if len(f.adminImages) == 0 {
		return nil, nil
	}
	value := f.adminImages[0]
	return &value, nil
}
func (f *fakeStore) AdminImageCounts(context.Context) (map[string]int, error) {
	return f.adminImageCounts, nil
}
func (f *fakeStore) UpdateAdminImage(_ context.Context, _ string, input model.AdminImageUpdateInput) (*model.AdminImage, error) {
	f.adminImageUpdateSeen = input
	return &model.AdminImage{ImageCard: model.ImageCard{ID: testCategoryID}}, nil
}
func (f *fakeStore) SetImagePrimaryCategory(_ context.Context, _ string, categoryID string) error {
	f.primaryCategorySeen = categoryID
	return nil
}
func (f *fakeStore) ReviewQueue(_ context.Context, query model.AdminSubmissionQuery) ([]model.Submission, int, error) {
	f.adminSubmissionQuery = query
	return f.adminSubmissions, len(f.adminSubmissions), nil
}
func (f *fakeStore) ReviewSubmission(context.Context, string, string, model.SubmissionReviewInput) (*model.Submission, error) {
	return f.submission, nil
}
func (f *fakeStore) MarkImagePublicRenditionReady(_ context.Context, imageID string) error {
	f.publicRenditionReadyID = imageID
	return nil
}
func (f *fakeStore) HideImage(context.Context, string, string, string) error { return nil }
func (f *fakeStore) DeleteAdminImage(_ context.Context, _ string, _ string, expectedUpdatedAt string) error {
	f.adminImageDeleteSeen = expectedUpdatedAt
	return nil
}
func (f *fakeStore) Cases(_ context.Context, query model.AdminCaseQuery) ([]model.Case, int, error) {
	f.adminCaseQuery = query
	return f.adminCases, len(f.adminCases), nil
}
func (f *fakeStore) ResolveCase(_ context.Context, _ string, _ string, input model.CaseResolutionInput) (*model.Case, error) {
	f.caseResolutionSeen = input
	if len(f.adminCases) == 0 {
		return nil, nil
	}
	return &f.adminCases[0], nil
}
func (f *fakeStore) RecordEvent(context.Context, model.Subject, string, model.EventInput) error {
	return nil
}

func TestApprovedSubmissionPublishesAssetBeforeImageBecomesPubliclyEligible(t *testing.T) {
	store := &fakeStore{submission: &model.Submission{
		ID: "019817c8-0000-7000-8300-000000000099", AssetID: "asset-1", ImageID: "image-1",
		Title: "Approved", ReviewState: "approved", Outcome: "published",
	}}
	assets := &submissionProcessingAssetPort{}
	service := New(store)
	service.SetAssetReferencePort(assets)

	value, err := service.ReviewSubmission(context.Background(), "operator-1", store.submission.ID, model.SubmissionReviewInput{Decision: "approve"})
	if err != nil {
		t.Fatal(err)
	}
	if value.ImageID != "image-1" || assets.publishedAssetID != "asset-1" || assets.publishedImageID != "image-1" {
		t.Fatalf("submission = %#v, published asset/image = %q/%q", value, assets.publishedAssetID, assets.publishedImageID)
	}
	if store.publicRenditionReadyID != "image-1" {
		t.Fatalf("ready image = %q", store.publicRenditionReadyID)
	}
}

func TestBulkReviewSubmissionsApprovesValidItemsAndKeepsPerItemFailure(t *testing.T) {
	store := &fakeStore{submission: &model.Submission{
		ID: testCategoryID, ReviewState: "approved", Outcome: "duplicate",
	}}
	results, err := New(store).BulkReviewSubmissions(context.Background(), "operator-1", model.BulkSubmissionReviewInput{
		SubmissionIDs: []string{PublicID(testCategoryID), "not-an-id", PublicID(testCategoryID)},
		Decision:      "approve",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 || !results[0].Success || results[1].Success || results[1].Error == "" {
		t.Fatalf("bulk review results = %#v", results)
	}
	_, err = New(store).BulkReviewSubmissions(context.Background(), "operator-1", model.BulkSubmissionReviewInput{
		SubmissionIDs: []string{PublicID(testCategoryID)}, Decision: "reject",
	})
	assertCode(t, err, "common.validation_failed")
}

func TestReconcilePublishedImagesBackfillsLegacyApprovals(t *testing.T) {
	store := &publicationReconciliationStore{
		fakeStore: &fakeStore{},
		candidates: []model.ImagePublicationCandidate{
			{ID: "image-1", AssetID: "asset-1", Title: "First"},
			{ID: "image-2", AssetID: "asset-2", Title: "Second"},
		},
	}
	assets := &submissionProcessingAssetPort{}
	service := New(store)
	service.SetAssetReferencePort(assets)
	processed, err := service.ReconcilePublishedImages(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if processed != 2 || assets.publishedImageID != "image-2" || store.publicRenditionReadyID != "image-2" {
		t.Fatalf("processed = %d, published = %q, ready = %q", processed, assets.publishedImageID, store.publicRenditionReadyID)
	}
}

func TestDiscoveryRequiresSettings(t *testing.T) {
	_, err := New(&fakeStore{}).Discovery(context.Background(), "seed")
	assertCode(t, err, "gallery.not_initialized")
}

func TestUpdateSiteSettingsPersistsOperatorManagedHomepage(t *testing.T) {
	store := &fakeStore{}
	input := model.SiteSettingsUpdateInput{
		Name:                " 月离图库 ",
		Title:               " 找到值得使用的图片 ",
		Description:         " 首页说明 ",
		SearchPlaceholder:   " 搜索图片、分类或标签 ",
		FooterTagline:       " 页脚说明 ",
		RandomCandidateSize: 240,
		HomeSections: []model.HomeSectionSettings{
			{Key: "random", Enabled: true, Position: 0, Title: "随机看看", ActionLabel: "换一批", ItemLimit: 24},
			{Key: "collections", Enabled: true, Position: 1, Title: "专题", ActionLabel: "查看专题", ItemLimit: 4},
			{Key: "latest", Enabled: true, Position: 2, Title: "最新", ActionLabel: "查看图片", ItemLimit: 8},
			{Key: "trending", Enabled: true, Position: 3, Title: "趋势", ActionLabel: "查看排行", ItemLimit: 8},
		},
	}
	settings, err := New(store).UpdateSiteSettings(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if settings.Name != "月离图库" || settings.RandomBatchSize != 24 {
		t.Fatalf("settings = %#v", settings)
	}
	if store.settingsUpdateSeen.Title != "找到值得使用的图片" {
		t.Fatalf("update input was not normalized: %#v", store.settingsUpdateSeen)
	}
}

func TestUpdateSiteSettingsRequiresCompleteUniqueSectionOrder(t *testing.T) {
	input := model.SiteSettingsUpdateInput{
		Name: "Gallery", Title: "Gallery", SearchPlaceholder: "Search",
		RandomCandidateSize: 240,
		HomeSections: []model.HomeSectionSettings{
			{Key: "random", Position: 0, Title: "Random", ItemLimit: 24},
			{Key: "collections", Position: 0, Title: "Collections", ItemLimit: 4},
		},
	}
	_, err := New(&fakeStore{}).UpdateSiteSettings(context.Background(), input)
	assertCode(t, err, "common.validation_failed")
}

func TestSubmissionOptionsKeepActiveClassificationOnEmptyGallery(t *testing.T) {
	store := validSubmissionStore()
	options, err := New(store).SubmissionOptions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if store.candidateCountCalls != 0 {
		t.Fatalf("submission options requested public candidate counts %d times", store.candidateCountCalls)
	}
	if !reflect.DeepEqual(options.Categories, []model.ClassificationNode{{
		ID: PublicID(testCategoryID), Slug: "wallpaper", Name: "壁纸",
	}}) {
		t.Fatalf("categories = %#v", options.Categories)
	}
	if !reflect.DeepEqual(options.Facets, []model.ClassificationFacet{{
		ID: PublicID(testFacetID), Slug: "scene", Name: "场景",
		Values: []model.ClassificationNode{{ID: PublicID(testValueID), Slug: "landscape", Name: "风景"}},
	}}) {
		t.Fatalf("facets = %#v", options.Facets)
	}
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

func TestRelatedImagesBoundsLimitAndNormalizesCards(t *testing.T) {
	store := &fakeStore{
		image: &model.ImageDetail{ImageCard: model.ImageCard{ID: testCategoryID}},
		relatedImages: []model.RelatedImage{{
			ImageCard: model.ImageCard{
				ID: testFacetID, ViewCount: 1250, FavoriteCount: 23,
			},
		}},
	}
	values, err := New(store).RelatedImages(context.Background(), PublicID(testCategoryID), 100)
	if err != nil {
		t.Fatal(err)
	}
	if store.relatedLimitSeen != 24 {
		t.Fatalf("related image limit must be bounded to 24, got %d", store.relatedLimitSeen)
	}
	if len(values) != 1 || values[0].ID != PublicID(testFacetID) {
		t.Fatalf("related images must expose public ids: %#v", values)
	}
	if values[0].Metrics.Views != 1250 || values[0].Metrics.Favorites != 23 {
		t.Fatalf("related image metrics were not normalized: %#v", values[0].Metrics)
	}
	if values[0].Reasons == nil {
		t.Fatal("related image reasons must serialize as an empty array")
	}
}

func TestAdminCollectionNormalizesPaginationAndPublicIDs(t *testing.T) {
	store := &fakeStore{collectionDetail: &model.CollectionDetail{
		Collection: model.Collection{
			ID: testCategoryID, Kind: "gallery.editorial", OwnerID: "gallery",
			CoverImageID: testFacetID, ItemCount: 25,
		},
		Images: []model.ImageCard{{ID: testValueID}},
	}}
	value, err := New(store).AdminCollection(context.Background(), PublicID(testCategoryID), 1, 12)
	if err != nil {
		t.Fatal(err)
	}
	if value.ID != PublicID(testCategoryID) || value.CoverImageID != PublicID(testFacetID) || value.Images[0].ID != PublicID(testValueID) {
		t.Fatalf("collection ids were not normalized: %#v", value)
	}
	if value.Page != 1 || value.PageSize != 12 || value.TotalPages != 3 {
		t.Fatalf("unexpected collection pagination: %#v", value)
	}
}

func TestAdminImagesNormalizesLifecycleQueryAndMetrics(t *testing.T) {
	store := validSubmissionStore()
	store.adminImages = []model.AdminImage{{ImageCard: model.ImageCard{ID: testCategoryID, ViewCount: 42, FavoriteCount: 7}, PrimaryCategoryID: testFacetID}}
	store.adminImageCounts = map[string]int{"all": 8, "published": 5, "draft": 3}
	page, err := New(store).AdminImages(context.Background(), model.AdminImageQuery{
		Search: "  rain  ", SortBy: "updatedAt", SortOrder: "asc", Page: -1, PageSize: 99,
		ProcessingState: "ready", ReviewState: "approved", PublicationState: "hidden", SafetyState: "safe",
		CategoryID: PublicID(testCategoryID), FacetValueID: PublicID(testValueID),
	})
	if err != nil {
		t.Fatal(err)
	}
	if store.adminImageQuerySeen.Search != "rain" || store.adminImageQuerySeen.Page != 1 || store.adminImageQuerySeen.PageSize != 60 || store.adminImageQuerySeen.CategoryID != testCategoryID || store.adminImageQuerySeen.FacetValueID != testValueID {
		t.Fatalf("admin image query = %#v", store.adminImageQuerySeen)
	}
	if page.Items[0].ID != PublicID(testCategoryID) || page.Items[0].PrimaryCategoryID != PublicID(testFacetID) || page.Items[0].Metrics.Views != 42 || page.Total != 1 || page.TotalPages != 1 || page.Counts["published"] != 5 {
		t.Fatalf("admin image page = %#v", page)
	}
	_, err = New(store).AdminImages(context.Background(), model.AdminImageQuery{SafetyState: "trusted"})
	assertCode(t, err, "common.validation_failed")
}

func TestAdminOverviewBoundsDaysAndNormalizesImageIDs(t *testing.T) {
	store := validSubmissionStore()
	store.adminOverview = &model.AdminOverview{TopImages: []model.AdminTrafficImage{{ID: testCategoryID, Title: "雨夜", Views: 42}}}
	value, err := New(store).AdminOverview(context.Background(), 99)
	if err != nil {
		t.Fatal(err)
	}
	if store.adminOverviewDays != 30 || value.Days != 30 || value.TopImages[0].ID != PublicID(testCategoryID) {
		t.Fatalf("overview = %#v days = %d", value, store.adminOverviewDays)
	}
}

func TestBulkImagesSupportsPrimaryCategoryAndPerItemFailure(t *testing.T) {
	store := validSubmissionStore()
	results := New(store).BulkImages(context.Background(), "operator-1", model.BulkImageActionInput{
		ImageIDs: []string{PublicID(testCategoryID), "not-an-id", PublicID(testCategoryID)},
		Action:   "set_primary_category", PrimaryCategoryID: PublicID(testFacetID),
	})
	if len(results) != 2 || !results[0].Success || results[1].Success || results[1].Error != "image_not_found" || store.primaryCategorySeen != testFacetID {
		t.Fatalf("bulk primary category results = %#v seen = %q", results, store.primaryCategorySeen)
	}
	unsupported := New(store).BulkImages(context.Background(), "operator-1", model.BulkImageActionInput{ImageIDs: []string{PublicID(testCategoryID)}, Action: "delete"})
	if len(unsupported) != 1 || unsupported[0].Error != "unsupported_action" {
		t.Fatalf("unsupported bulk action = %#v", unsupported)
	}
}

func TestUpdateAdminImageValidatesAndNormalizesMetadata(t *testing.T) {
	store := validSubmissionStore()
	service := New(store)
	updatedAt := "2026-07-16T14:00:00.123456Z"
	value, err := service.UpdateAdminImage(context.Background(), PublicID(testCategoryID), model.AdminImageUpdateInput{
		ExpectedUpdatedAt: updatedAt, Title: "  雨夜  ", Description: "  城市  ", AltText: "  雨夜街道  ", SourceURL: " https://example.com/source ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if value.ID != PublicID(testCategoryID) || store.adminImageUpdateSeen.Title != "雨夜" || store.adminImageUpdateSeen.SourceURL != "https://example.com/source" {
		t.Fatalf("admin update = %#v value = %#v", store.adminImageUpdateSeen, value)
	}
	_, err = service.UpdateAdminImage(context.Background(), PublicID(testCategoryID), model.AdminImageUpdateInput{ExpectedUpdatedAt: "stale", Title: "雨夜", AltText: "雨夜"})
	assertCode(t, err, "common.validation_failed")
}

func TestDeleteAdminImageRequiresOperatorAndVersion(t *testing.T) {
	store := validSubmissionStore()
	service := New(store)
	if err := service.DeleteAdminImage(context.Background(), "operator-1", PublicID(testCategoryID), model.AdminImageDeleteInput{
		ExpectedUpdatedAt: "2026-08-28T04:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	if store.adminImageDeleteSeen != "2026-08-28T04:00:00Z" {
		t.Fatalf("delete version = %q", store.adminImageDeleteSeen)
	}
	assertCode(t, service.DeleteAdminImage(context.Background(), "", PublicID(testCategoryID), model.AdminImageDeleteInput{ExpectedUpdatedAt: "2026-08-28T04:00:00Z"}), "gallery.forbidden")
	assertCode(t, service.DeleteAdminImage(context.Background(), "operator-1", PublicID(testCategoryID), model.AdminImageDeleteInput{}), "common.validation_failed")
}

func TestAdminImageUpdatedAtRoundTripsIntoOptimisticUpdate(t *testing.T) {
	updatedAt := time.Date(2026, time.July, 27, 5, 33, 0, 123456000, time.FixedZone("CST", 8*60*60))
	payload, err := json.Marshal(model.AdminImage{UpdatedAt: &updatedAt})
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		UpdatedAt string `json:"updatedAt"`
	}
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatal(err)
	}

	store := validSubmissionStore()
	_, err = New(store).UpdateAdminImage(context.Background(), PublicID(testCategoryID), model.AdminImageUpdateInput{
		ExpectedUpdatedAt: response.UpdatedAt,
		Title:             "雨夜",
		AltText:           "雨夜街道",
	})
	if err != nil {
		t.Fatalf("serialized updatedAt %q must be accepted for optimistic update: %v", response.UpdatedAt, err)
	}
}

func TestAdminCaseUpdatedAtRoundTripsIntoOptimisticUpdate(t *testing.T) {
	updatedAt := time.Date(2026, time.July, 27, 5, 33, 0, 123456000, time.FixedZone("CST", 8*60*60))
	payload, err := json.Marshal(model.Case{UpdatedAt: &updatedAt})
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		UpdatedAt string `json:"updatedAt"`
	}
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatal(err)
	}

	store := &fakeStore{adminCases: []model.Case{{ID: testCategoryID, ImageID: testFacetID, SubmissionID: testValueID}}}
	_, err = New(store).ResolveCase(context.Background(), "operator", PublicID(testCategoryID), model.CaseResolutionInput{
		ExpectedUpdatedAt: response.UpdatedAt,
		Status:            "reviewing",
	})
	if err != nil {
		t.Fatalf("serialized updatedAt %q must be accepted for case update: %v", response.UpdatedAt, err)
	}
}

func TestBulkHideImagesReportsPerItemOutcomes(t *testing.T) {
	store := validSubmissionStore()
	results := New(store).BulkHideImages(context.Background(), "operator-1", model.BulkImageHideInput{
		ImageIDs: []string{PublicID(testCategoryID), "not-an-id", PublicID(testCategoryID)}, Reason: "policy",
	})
	if len(results) != 2 || !results[0].Success || results[1].Success || results[1].Error == "" {
		t.Fatalf("bulk hide results = %#v", results)
	}
}

func TestAdminSubmissionQueueNormalizesFiltersAndPagination(t *testing.T) {
	store := &fakeStore{adminSubmissions: []model.Submission{{ID: testCategoryID, AssetID: testFacetID, ImageID: testValueID}}}
	page, err := New(store).ReviewQueue(context.Background(), model.AdminSubmissionQuery{
		Search: "  rain  ", ProcessingState: "failed", ReviewState: "pending",
		SafetyState: "uncertain", Outcome: "failed",
	})
	if err != nil {
		t.Fatal(err)
	}
	if store.adminSubmissionQuery.Search != "rain" || store.adminSubmissionQuery.SortBy != "createdAt" || store.adminSubmissionQuery.SortOrder != "asc" || store.adminSubmissionQuery.Page != 1 || store.adminSubmissionQuery.PageSize != 20 {
		t.Fatalf("normalized submission query = %#v", store.adminSubmissionQuery)
	}
	if page.Total != 1 || page.TotalPages != 1 || len(page.Items) != 1 || page.Items[0].ID == testCategoryID {
		t.Fatalf("admin submission page = %#v", page)
	}
	if _, err := New(store).ReviewQueue(context.Background(), model.AdminSubmissionQuery{Outcome: "archived"}); err == nil {
		t.Fatal("unsupported admin submission outcome must be rejected")
	}
}

func TestAdminCasesNormalizeQueryAndRequireCurrentVersion(t *testing.T) {
	updatedAt := time.Now().UTC().Format(time.RFC3339Nano)
	store := &fakeStore{adminCases: []model.Case{{ID: testCategoryID, ImageID: testFacetID, SubmissionID: testValueID}}}
	page, err := New(store).Cases(context.Background(), model.AdminCaseQuery{
		Search: "  source  ", Status: "all", Kind: "source_correction", SortBy: "updatedAt", SortOrder: "desc", Page: 2, PageSize: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if store.adminCaseQuery.Search != "source" || store.adminCaseQuery.Page != 2 || page.PageSize != 10 || page.Items[0].ID == testCategoryID {
		t.Fatalf("admin case query/page = query %#v page %#v", store.adminCaseQuery, page)
	}
	value, err := New(store).ResolveCase(context.Background(), " operator ", PublicID(testCategoryID), model.CaseResolutionInput{
		ExpectedUpdatedAt: updatedAt, Status: "resolved", Note: "  source corrected  ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if value.ID == testCategoryID || store.caseResolutionSeen.Note != "source corrected" || store.caseResolutionSeen.ExpectedUpdatedAt != updatedAt {
		t.Fatalf("resolved case = value %#v input %#v", value, store.caseResolutionSeen)
	}
	if _, err := New(store).ResolveCase(context.Background(), "operator", PublicID(testCategoryID), model.CaseResolutionInput{
		ExpectedUpdatedAt: "stale", Status: "reviewing",
	}); err == nil {
		t.Fatal("case resolution without an RFC3339 version must be rejected")
	}
}

func TestMySubmissionsNormalizesAndValidatesFilters(t *testing.T) {
	store := validSubmissionStore()
	_, _, err := New(store).MySubmissions(context.Background(), model.Subject{Kind: "user", ID: "user-1"}, model.MySubmissionQuery{
		Page: -2, PageSize: 999, Outcome: " published ", ProcessingState: "ready", ReviewState: "approved",
	})
	if err != nil {
		t.Fatal(err)
	}
	if store.mySubmissionsQuerySeen.Page != 1 || store.mySubmissionsQuerySeen.PageSize != 60 || store.mySubmissionsQuerySeen.Outcome != "published" {
		t.Fatalf("submission query was not normalized: %#v", store.mySubmissionsQuerySeen)
	}
	_, _, err = New(store).MySubmissions(context.Background(), model.Subject{Kind: "user", ID: "user-1"}, model.MySubmissionQuery{Outcome: "unknown"})
	assertCode(t, err, "common.validation_failed")
}

func TestFavoritesPaginatesSortsAndNormalizesPublicIDs(t *testing.T) {
	store := &fakeCollectionStore{fakeStore: validSubmissionStore()}
	store.collectionDetail = &model.CollectionDetail{
		Collection: model.Collection{ID: testCategoryID, ItemCount: 61},
		Images:     []model.ImageCard{{ID: testFacetID}},
	}
	value, err := New(store).Favorites(context.Background(), "user-1", 2, 24, "title_asc")
	if err != nil {
		t.Fatal(err)
	}
	if store.favoritePageSeen != 2 || store.favoriteSizeSeen != 24 || store.favoriteSortSeen != "title_asc" {
		t.Fatalf("favorites query = page %d size %d sort %q", store.favoritePageSeen, store.favoriteSizeSeen, store.favoriteSortSeen)
	}
	if value.Page != 2 || value.PageSize != 24 || value.TotalPages != 3 || value.Images[0].ID != PublicID(testFacetID) {
		t.Fatalf("favorites page = %#v", value)
	}
	_, err = New(store).Favorites(context.Background(), "user-1", 1, 24, "random")
	assertCode(t, err, "common.validation_failed")
}

func TestEditorialCollectionUpdateAndOrderNormalizeInputs(t *testing.T) {
	store := &fakeStore{}
	service := New(store)
	_, err := service.UpdateEditorialCollection(context.Background(), PublicID(testCategoryID), model.EditorialCollectionUpdateInput{
		Version: 4, Name: "  夜色  ", Slug: " Night-Colors ", Visibility: "public", CoverImageID: PublicID(testFacetID),
	})
	if err != nil {
		t.Fatal(err)
	}
	if store.collectionUpdateSeen.Name != "夜色" || store.collectionUpdateSeen.Slug != "night-colors" || store.collectionUpdateSeen.CoverImageID != testFacetID {
		t.Fatalf("collection update was not normalized: %#v", store.collectionUpdateSeen)
	}
	_, err = service.ReorderEditorialMembers(context.Background(), PublicID(testCategoryID), model.EditorialCollectionOrderInput{
		Version: 5, ImageIDs: []string{PublicID(testFacetID), PublicID(testValueID)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(store.collectionOrderSeen, []string{testFacetID, testValueID}) {
		t.Fatalf("collection order ids were not normalized: %#v", store.collectionOrderSeen)
	}
	_, err = service.ReorderEditorialMembers(context.Background(), PublicID(testCategoryID), model.EditorialCollectionOrderInput{
		Version: 5, ImageIDs: []string{PublicID(testFacetID), PublicID(testFacetID)},
	})
	assertCode(t, err, "common.validation_failed")
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

func TestImagesPaginatesLargeFixture(t *testing.T) {
	store := validSubmissionStore()
	store.candidates = make([]model.ImageCard, 101)
	for index := range store.candidates {
		store.candidates[index] = model.ImageCard{ID: "fixture-" + strconv.Itoa(index+1)}
	}

	page, err := New(store).Images(context.Background(), model.ImageQuery{Page: 3, PageSize: 24})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 24 || page.Total != 101 || page.TotalPages != 5 || page.Page != 3 {
		t.Fatalf("large fixture page = %#v", page)
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

func TestCreateClassificationIdentityNormalizesDraftInput(t *testing.T) {
	store := validSubmissionStore()
	service := New(store)
	if _, err := service.classificationCatalog(context.Background()); err != nil {
		t.Fatal(err)
	}
	created, err := service.CreateClassificationIdentity(context.Background(), " operator-1 ", model.ClassificationIdentityCreateInput{
		Kind: "category", Name: " 风景 ", Slug: " LANDSCAPE ", ParentID: PublicID(testCategoryID),
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != PublicID(testCategoryID) || created.CatalogRevision != 9 {
		t.Fatalf("created = %#v", created)
	}
	if store.classificationCreateSeen.Name != "风景" || store.classificationCreateSeen.Slug != "landscape" || store.classificationCreateSeen.ParentID != testCategoryID {
		t.Fatalf("creation input = %#v", store.classificationCreateSeen)
	}
	if service.catalog != nil || service.catalogRevision != 0 {
		t.Fatal("classification creation must invalidate the local catalog")
	}
}

func TestCreateClassificationIdentityRequiresFacetForValue(t *testing.T) {
	_, err := New(validSubmissionStore()).CreateClassificationIdentity(context.Background(), "operator-1", model.ClassificationIdentityCreateInput{
		Kind: "facet_value", Name: "横图", Slug: "landscape",
	})
	assertCode(t, err, "common.validation_failed")
}

func TestUpdateClassificationIdentityNormalizesTag(t *testing.T) {
	store := validSubmissionStore()
	updated, err := New(store).UpdateClassificationIdentity(context.Background(), "operator-1", PublicID(testCategoryID), model.ClassificationIdentityUpdateInput{
		Kind: "tag", Name: " 角色 ", Slug: " CHARACTER ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != PublicID(testCategoryID) || updated.CatalogRevision != 10 {
		t.Fatalf("updated = %#v", updated)
	}
	if store.classificationUpdateSeen.Name != "角色" || store.classificationUpdateSeen.Slug != "character" {
		t.Fatalf("update input = %#v", store.classificationUpdateSeen)
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

func TestClassificationTagsExposePublicIDsThatMatchAdminImageSelections(t *testing.T) {
	store := validSubmissionStore()
	store.classificationTags = []model.ClassificationTag{{
		ID: testCategoryID, Name: "旅行", Slug: "travel", Status: "active",
		ReplacementID: testFacetID,
	}}

	page, err := New(store).ClassificationTags(context.Background(), "", 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("classification tags = %#v", page.Items)
	}
	if page.Items[0].ID != PublicID(testCategoryID) || page.Items[0].ReplacementID != PublicID(testFacetID) {
		t.Fatalf("classification tag IDs must use the public boundary: %#v", page.Items[0])
	}
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
	value, ok, resolveErr := problem.FromError(err, "gallery-test")
	if resolveErr != nil || !ok || value.Code != want {
		t.Fatalf("expected %s, got %v", want, err)
	}
}
