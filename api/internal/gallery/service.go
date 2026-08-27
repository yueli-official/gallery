package gallery

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yueli-official/foundation/go/abuse"

	"github.com/yueli-official/foundation/go/classification"
	"github.com/yueli-official/foundation/go/identifier"
	"github.com/yueli-official/gallery/api/internal/collection"
	"github.com/yueli-official/gallery/api/internal/galleryabuse"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/model"
)

type Store interface {
	SiteSettings(context.Context) (*model.SiteSettings, error)
	UpdateSiteSettings(context.Context, model.SiteSettingsUpdateInput) (*model.SiteSettings, error)
	ClassificationRevision(context.Context) (uint64, error)
	ClassificationSnapshot(context.Context) (classification.Snapshot, error)
	ClassificationTagMatches(context.Context, []classification.TagLookupRequest) ([]classification.TagMatch, string, error)
	ClassificationCandidateCounts(context.Context, model.ImageQuery, []classification.CandidateCountGroupRequest) ([]classification.CandidateCountGroup, string, error)
	ClassificationGovernanceImpacts(context.Context, []classification.ImpactRequest) ([]classification.ReferenceImpact, string, error)
	ExecuteClassificationGovernance(context.Context, string, classification.GovernFactRequest, classification.GovernancePlan) (uint64, error)
	ClassificationTags(context.Context, model.ClassificationTagCursor, int) ([]model.ClassificationTag, bool, error)
	ClassificationTagProposals(context.Context, string, int, int) ([]model.ClassificationTagProposal, int, error)
	ReviewClassificationTagProposal(context.Context, string, string, model.ClassificationTagProposalReviewInput) (*model.ClassificationTagProposal, bool, error)
	RandomCandidates(context.Context, int) ([]model.ImageCard, error)
	ListImages(context.Context, model.ImageQuery, classification.FilterPlan) ([]model.ImageCard, int, error)
	Image(context.Context, string, string) (*model.ImageDetail, error)
	RelatedImages(context.Context, string, int) ([]model.RelatedImage, error)
	HasTombstone(context.Context, string) (bool, error)
	PublicCollections(context.Context) ([]model.Collection, error)
	EditorialCollections(context.Context) ([]model.Collection, error)
	PublicCollection(context.Context, string, int, int) (*model.CollectionDetail, error)
	CreateEditorialCollection(context.Context, string, model.EditorialCollectionInput) (*model.Collection, error)
	UpdateEditorialCollection(context.Context, string, model.EditorialCollectionUpdateInput) (*model.Collection, error)
	ReorderEditorialMembers(context.Context, string, int64, []string) (*model.Collection, error)
	Ranking(context.Context, string, string, int) (*model.Ranking, error)
	CreateSubmission(context.Context, model.Subject, model.SubmissionInput, string) (*model.Submission, error)
	MySubmissions(context.Context, model.Subject, model.MySubmissionQuery) ([]model.Submission, int, error)
	WithdrawSubmission(context.Context, model.Subject, string) (*model.Submission, error)
	FailSubmission(context.Context, string, string) error
	CollectionDetail(context.Context, string, int, int) (*model.CollectionDetail, error)
	FavoritesDetail(context.Context, string, int, int, string) (*model.CollectionDetail, error)
	CreateCase(context.Context, model.Subject, string, model.CaseInput) (*model.Case, error)
	AdminOverview(context.Context, int) (*model.AdminOverview, error)
	AdminImages(context.Context, model.AdminImageQuery) ([]model.AdminImage, int, error)
	AdminImage(context.Context, string) (*model.AdminImage, error)
	AdminImageCounts(context.Context) (map[string]int, error)
	UpdateAdminImage(context.Context, string, model.AdminImageUpdateInput) (*model.AdminImage, error)
	SetImagePrimaryCategory(context.Context, string, string) error
	ReviewQueue(context.Context, model.AdminSubmissionQuery) ([]model.Submission, int, error)
	ReviewSubmission(context.Context, string, string, model.SubmissionReviewInput) (*model.Submission, error)
	MarkImagePublicRenditionReady(context.Context, string) error
	HideImage(context.Context, string, string, string) error
	Cases(context.Context, model.AdminCaseQuery) ([]model.Case, int, error)
	ResolveCase(context.Context, string, string, model.CaseResolutionInput) (*model.Case, error)
	RecordEvent(context.Context, model.Subject, string, model.EventInput) error
}

type Service struct {
	store            Store
	collections      *collection.Service
	assets           AssetReferencePort
	clock            func() time.Time
	catalogMu        sync.Mutex
	catalog          *classification.Catalog
	catalogRevision  uint64
	catalogCheckedAt time.Time
	abuse            galleryabuse.Actions
}

type AssetReferencePort interface {
	RegisterSubmission(context.Context, string, string, string, string) error
	UnregisterSubmission(context.Context, string, string, string) error
	PrepareSubmission(context.Context, string) (model.SubmissionAssetFacts, error)
	PublishImage(context.Context, string, string, string) error
}

type SubmissionProcessingStore interface {
	ClaimSubmissionProcessing(context.Context, int) ([]model.Submission, error)
	CompleteSubmissionProcessing(context.Context, string, model.SubmissionAssetFacts) error
	SubmissionAssetID(context.Context, string) (string, error)
}

type PublicationReconciliationStore interface {
	PublishedImageCandidates(context.Context, string, int) ([]model.ImagePublicationCandidate, error)
}

type GuestClaimStore interface {
	ClaimGuestSubmissions(context.Context, string, string) (int64, error)
}

func New(store Store) *Service {
	service := &Service{store: store, clock: time.Now}
	collectionStore, storeOK := any(store).(collection.Store)
	resources, resourceOK := any(store).(collection.ResourcePort)
	if storeOK && resourceOK {
		service.collections, _ = collection.New(collectionStore, resources, collection.GalleryKinds()...)
	}
	return service
}

func (s *Service) SetAssetReferencePort(port AssetReferencePort) { s.assets = port }

func (s *Service) SetAbuse(module abuse.Module) error {
	actions, err := galleryabuse.Bind(module)
	if err != nil {
		return err
	}
	s.abuse = actions
	return nil
}

func (s *Service) AdmitSubmission(
	ctx context.Context,
	subject model.Subject,
	ip string,
	attemptID string,
	proof string,
) (abuse.Admission, error) {
	if err := normalizeSubject(&subject); err != nil {
		return abuse.Admission{}, err
	}
	action := s.abuse.Member
	if subject.Kind == "guest" {
		action = s.abuse.Guest
	}
	if action == nil {
		return abuse.Admission{Disposition: abuse.DispositionAllow}, nil
	}
	network, err := galleryabuse.NetworkPrefix(ip)
	if err != nil {
		return abuse.Admission{}, err
	}
	input := abuse.Input{
		ID: abuse.AttemptID(attemptID),
		Signals: abuse.Signals{
			Network: network,
			Actor:   subject.Kind + ":" + subject.ID,
		},
	}
	if proof = strings.TrimSpace(proof); proof != "" {
		input.Proof = &abuse.Proof{Kind: "turnstile", Token: proof}
	}
	return action.Admit(ctx, input)
}

func (s *Service) ProcessQueuedSubmissions(ctx context.Context, limit int) (int, error) {
	store, ok := s.store.(SubmissionProcessingStore)
	if !ok || s.assets == nil {
		return 0, galleryerr.NotInitialized("submission_processor")
	}
	limit = bounded(limit, 1, 32, 8)
	claimed, err := store.ClaimSubmissionProcessing(ctx, limit)
	if err != nil {
		return 0, err
	}
	for _, submission := range claimed {
		facts, prepareErr := s.assets.PrepareSubmission(ctx, submission.AssetID)
		if prepareErr != nil {
			_ = s.store.FailSubmission(ctx, submission.ID, "asset_processing_failed")
			continue
		}
		if err := store.CompleteSubmissionProcessing(ctx, submission.ID, facts); err != nil {
			return len(claimed), err
		}
	}
	return len(claimed), nil
}

// ReconcilePublishedImages backfills the Asset publication reference for
// images approved before public rendition publication became explicit. The
// cursor keeps the sweep bounded and the Asset operation is idempotent.
func (s *Service) ReconcilePublishedImages(ctx context.Context, pageSize int) (int, error) {
	store, ok := s.store.(PublicationReconciliationStore)
	if !ok || s.assets == nil {
		return 0, galleryerr.NotInitialized("publication_reconciler")
	}
	if pageSize <= 0 || pageSize > 500 {
		pageSize = 100
	}
	processed := 0
	afterID := ""
	for {
		items, err := store.PublishedImageCandidates(ctx, afterID, pageSize)
		if err != nil {
			return processed, err
		}
		for _, item := range items {
			if err := s.assets.PublishImage(ctx, item.AssetID, item.ID, item.Title); err != nil {
				return processed, err
			}
			if err := s.store.MarkImagePublicRenditionReady(ctx, item.ID); err != nil {
				return processed, err
			}
			processed++
			afterID = item.ID
		}
		if len(items) < pageSize {
			return processed, nil
		}
	}
}

func (s *Service) SubmissionPreviewURL(ctx context.Context, rawID string) (string, error) {
	id, err := DatabaseID(rawID)
	if err != nil {
		return "", galleryerr.NotFound("submission", rawID)
	}
	store, ok := s.store.(SubmissionProcessingStore)
	if !ok || s.assets == nil {
		return "", galleryerr.NotInitialized("submission_preview")
	}
	assetID, err := store.SubmissionAssetID(ctx, id)
	if err != nil {
		return "", err
	}
	if assetID == "" {
		return "", galleryerr.NotFound("submission", rawID)
	}
	facts, err := s.assets.PrepareSubmission(ctx, assetID)
	if err != nil {
		return "", err
	}
	return facts.PreviewURL, nil
}

func (s *Service) ClaimGuestSubmissions(ctx context.Context, guestSubject, userID string) (int64, error) {
	guestSubject = strings.TrimSpace(guestSubject)
	userID = strings.TrimSpace(userID)
	if guestSubject == "" || userID == "" {
		return 0, galleryerr.Forbidden()
	}
	store, ok := s.store.(GuestClaimStore)
	if !ok {
		return 0, galleryerr.NotInitialized("guest claim store")
	}
	return store.ClaimGuestSubmissions(ctx, guestSubject, userID)
}

func (s *Service) Discovery(ctx context.Context, seed string) (*model.Discovery, error) {
	if s.store == nil {
		return nil, galleryerr.NotInitialized("store")
	}
	settings, err := s.store.SiteSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, galleryerr.NotInitialized("site_settings")
	}
	seed = strings.TrimSpace(seed)
	if seed == "" {
		seed = s.clock().UTC().Format("2006-01-02")
	}
	candidates, err := s.store.RandomCandidates(ctx, bounded(settings.RandomCandidateSize, 40, 2000, 240))
	if err != nil {
		return nil, err
	}
	images := seededDiverse(seed, candidates, bounded(settings.RandomBatchSize, 12, 80, 30))
	catalog, err := s.classificationCatalog(ctx)
	if err != nil {
		return nil, err
	}
	preparation := catalog.Discover(classification.DiscoverRequest{PolicyKey: "gallery.image.public", CandidateProjection: classification.CandidateProjectionAvailable})
	factRequest := preparation.FactRequest()
	countGroups, freshnessToken, err := s.store.ClassificationCandidateCounts(ctx, model.ImageQuery{}, factRequest.CountGroups)
	if err != nil {
		return nil, err
	}
	discovery := preparation.Complete(classification.DiscoverFacts{
		CatalogRevision: factRequest.CatalogRevision,
		RequestToken:    factRequest.RequestToken,
		FreshnessToken:  freshnessToken,
		CountGroups:     countGroups,
	})
	if discovery.Outcome != classification.OutcomeAccepted {
		return nil, classificationValidation(discovery.Diagnostics)
	}
	normalizeCards(images)
	return &model.Discovery{
		Site:       *settings,
		Seed:       seed,
		Images:     images,
		Categories: categoryCandidates(discovery.Candidates.Categories),
		Facets:     facetCandidates(discovery.Candidates.Facets),
	}, nil
}

func (s *Service) SiteSettings(ctx context.Context) (*model.SiteSettings, error) {
	if s.store == nil {
		return nil, galleryerr.NotInitialized("store")
	}
	settings, err := s.store.SiteSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, galleryerr.NotInitialized("site_settings")
	}
	return settings, nil
}

func (s *Service) UpdateSiteSettings(ctx context.Context, input model.SiteSettingsUpdateInput) (*model.SiteSettings, error) {
	if s.store == nil {
		return nil, galleryerr.NotInitialized("store")
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.SearchPlaceholder = strings.TrimSpace(input.SearchPlaceholder)
	input.FooterTagline = strings.TrimSpace(input.FooterTagline)
	if input.Name == "" || len([]rune(input.Name)) > 80 {
		return nil, galleryerr.Validation("name", "name is required and must be at most 80 characters")
	}
	if input.Title == "" || len([]rune(input.Title)) > 120 {
		return nil, galleryerr.Validation("title", "title is required and must be at most 120 characters")
	}
	if len([]rune(input.Description)) > 320 {
		return nil, galleryerr.Validation("description", "description must be at most 320 characters")
	}
	if input.SearchPlaceholder == "" || len([]rune(input.SearchPlaceholder)) > 120 {
		return nil, galleryerr.Validation("searchPlaceholder", "search placeholder is required and must be at most 120 characters")
	}
	if len([]rune(input.FooterTagline)) > 240 {
		return nil, galleryerr.Validation("footerTagline", "footer tagline must be at most 240 characters")
	}
	if input.RandomCandidateSize < 40 || input.RandomCandidateSize > 2000 {
		return nil, galleryerr.Validation("randomCandidateSize", "random candidate size must be between 40 and 2000")
	}

	expected := map[string]struct{}{
		"random":      {},
		"collections": {},
		"latest":      {},
		"trending":    {},
	}
	seenPositions := make(map[int]struct{}, len(expected))
	for index := range input.HomeSections {
		section := &input.HomeSections[index]
		section.Key = strings.TrimSpace(section.Key)
		section.Title = strings.TrimSpace(section.Title)
		section.Description = strings.TrimSpace(section.Description)
		section.ActionLabel = strings.TrimSpace(section.ActionLabel)
		if _, ok := expected[section.Key]; !ok {
			return nil, galleryerr.Validation("homeSections", "unsupported or duplicate home section")
		}
		delete(expected, section.Key)
		if section.Position < 0 || section.Position > 3 {
			return nil, galleryerr.Validation("homeSections", "home section position must be between 0 and 3")
		}
		if _, exists := seenPositions[section.Position]; exists {
			return nil, galleryerr.Validation("homeSections", "home section positions must be unique")
		}
		seenPositions[section.Position] = struct{}{}
		if section.Title == "" || len([]rune(section.Title)) > 80 {
			return nil, galleryerr.Validation("homeSections", "section title is required and must be at most 80 characters")
		}
		if len([]rune(section.Description)) > 240 || len([]rune(section.ActionLabel)) > 40 {
			return nil, galleryerr.Validation("homeSections", "section description or action label is too long")
		}
		minimum, maximum := 1, 24
		if section.Key == "random" {
			minimum, maximum = 12, 60
		}
		if section.Key == "collections" {
			maximum = 12
		}
		if section.ItemLimit < minimum || section.ItemLimit > maximum {
			return nil, galleryerr.Validation("homeSections", "section item limit is outside the supported range")
		}
	}
	if len(input.HomeSections) != 4 || len(expected) != 0 {
		return nil, galleryerr.Validation("homeSections", "all four home sections are required")
	}
	return s.store.UpdateSiteSettings(ctx, input)
}

func (s *Service) SubmissionOptions(ctx context.Context) (*model.SubmissionOptions, error) {
	catalog, err := s.classificationCatalog(ctx)
	if err != nil {
		return nil, err
	}
	preparation := catalog.Discover(classification.DiscoverRequest{
		PolicyKey:           "gallery.image.public",
		CandidateProjection: classification.CandidateProjectionActive,
	})
	factRequest := preparation.FactRequest()
	result := preparation.Complete(classification.DiscoverFacts{
		CatalogRevision: factRequest.CatalogRevision,
		RequestToken:    factRequest.RequestToken,
	})
	if result.Outcome != classification.OutcomeAccepted {
		return nil, classificationValidation(result.Diagnostics)
	}
	return &model.SubmissionOptions{
		Categories: categoryCandidates(result.Candidates.Categories),
		Facets:     facetCandidates(result.Candidates.Facets),
	}, nil
}

func (s *Service) Images(ctx context.Context, query model.ImageQuery) (*model.ImagePage, error) {
	query.Search = strings.TrimSpace(query.Search)
	query.Tag = strings.TrimSpace(query.Tag)
	query.Page = bounded(query.Page, 1, 100000, 1)
	query.PageSize = bounded(query.PageSize, 12, 60, 24)
	query.Sort = defaultString(strings.TrimSpace(query.Sort), "newest")
	if !oneOf(query.Sort, "newest", "oldest", "title_asc", "title_desc") {
		return nil, galleryerr.Validation("sort", "unsupported sort")
	}
	discoverRequest, err := publicDiscoverRequest(query)
	if err != nil {
		return nil, err
	}
	catalog, err := s.classificationCatalog(ctx)
	if err != nil {
		return nil, err
	}
	discoverRequest.CandidateProjection = classification.CandidateProjectionAvailable
	preparation := catalog.Discover(discoverRequest)
	factRequest := preparation.FactRequest()
	countGroups, freshnessToken, err := s.store.ClassificationCandidateCounts(ctx, query, factRequest.CountGroups)
	if err != nil {
		return nil, err
	}
	discovery := preparation.Complete(classification.DiscoverFacts{
		CatalogRevision: factRequest.CatalogRevision,
		RequestToken:    factRequest.RequestToken,
		FreshnessToken:  freshnessToken,
		CountGroups:     countGroups,
	})
	diagnostics := classificationDiagnostics(discovery.Diagnostics)
	if discovery.Outcome == classification.OutcomeRejected {
		return nil, classificationValidation(discovery.Diagnostics)
	}
	if discovery.Outcome == classification.OutcomeNonExecutable {
		return &model.ImagePage{
			Items: []model.ImageCard{}, Page: query.Page, PageSize: query.PageSize,
			Diagnostics: diagnostics,
			Categories:  categoryCandidates(discovery.Candidates.Categories), Facets: facetCandidates(discovery.Candidates.Facets),
		}, nil
	}
	items, total, err := s.store.ListImages(ctx, query, discovery.FilterPlan)
	if err != nil {
		return nil, err
	}
	normalizeCards(items)
	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(query.PageSize)))
	}
	return &model.ImagePage{
		Items: nonNilCards(items), Page: query.Page, PageSize: query.PageSize,
		Total: total, TotalPages: pages, Diagnostics: diagnostics,
		Categories: categoryCandidates(discovery.Candidates.Categories), Facets: facetCandidates(discovery.Candidates.Facets),
	}, nil
}

func publicDiscoverRequest(query model.ImageQuery) (classification.DiscoverRequest, error) {
	request := classification.DiscoverRequest{PolicyKey: "gallery.image.public"}
	for _, raw := range query.CategoryRefs {
		request.Categories = append(request.Categories, classificationReference(raw))
	}
	byFacet := make(map[string][]classification.Reference)
	for _, raw := range query.FacetRefs {
		facet, value, found := strings.Cut(strings.TrimSpace(raw), ":")
		if !found || strings.TrimSpace(facet) == "" || strings.TrimSpace(value) == "" {
			return classification.DiscoverRequest{}, galleryerr.Validation("facets", "facet filters must use facet:value")
		}
		facet = strings.TrimSpace(facet)
		byFacet[facet] = append(byFacet[facet], classificationReference(value))
	}
	facetKeys := make([]string, 0, len(byFacet))
	for facet := range byFacet {
		facetKeys = append(facetKeys, facet)
	}
	sort.Strings(facetKeys)
	for _, facet := range facetKeys {
		request.Facets = append(request.Facets, classification.FacetFilter{
			Facet: classificationReference(facet), Values: byFacet[facet],
		})
	}
	return request, nil
}

func classificationReference(raw string) classification.Reference {
	value := strings.TrimSpace(raw)
	if id, err := DatabaseID(value); err == nil {
		return classification.Reference{Kind: classification.ReferenceByID, Value: id}
	}
	return classification.Reference{Kind: classification.ReferenceBySlug, Value: value}
}

func classificationDiagnostics(values []classification.Diagnostic) []model.ClassificationDiagnostic {
	result := make([]model.ClassificationDiagnostic, 0, len(values))
	for _, value := range values {
		result = append(result, model.ClassificationDiagnostic{
			Code: string(value.Code), Path: append([]string(nil), value.Path...),
			Reference: value.Reference, Params: value.Params,
		})
	}
	return result
}

func (s *Service) Image(ctx context.Context, rawID, userID string) (*model.ImageDetail, error) {
	id, err := DatabaseID(rawID)
	if err != nil {
		return nil, galleryerr.NotFound("image", strings.TrimSpace(rawID))
	}
	value, err := s.store.Image(ctx, id, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	if value == nil {
		gone, tombstoneErr := s.store.HasTombstone(ctx, id)
		if tombstoneErr != nil {
			return nil, tombstoneErr
		}
		if gone {
			return nil, galleryerr.Gone("image", PublicID(id))
		}
		return nil, galleryerr.NotFound("image", PublicID(id))
	}
	normalizeDetail(value)
	return value, nil
}

func (s *Service) RelatedImages(ctx context.Context, rawID string, limit int) ([]model.RelatedImage, error) {
	id, err := DatabaseID(rawID)
	if err != nil {
		return nil, galleryerr.NotFound("image", strings.TrimSpace(rawID))
	}
	_, err = s.Image(ctx, rawID, "")
	if err != nil {
		return nil, err
	}
	values, err := s.store.RelatedImages(ctx, id, bounded(limit, 1, 24, 8))
	if err != nil {
		return nil, err
	}
	if values == nil {
		values = []model.RelatedImage{}
	}
	for index := range values {
		values[index].ID = PublicID(values[index].ID)
		values[index].Metrics = model.Metrics{
			Views: values[index].ViewCount, Favorites: values[index].FavoriteCount,
		}
		if values[index].Reasons == nil {
			values[index].Reasons = []model.RelatedImageReason{}
		}
	}
	return values, nil
}

func (s *Service) Collections(ctx context.Context) ([]model.Collection, error) {
	values, err := s.store.PublicCollections(ctx)
	if values == nil {
		values = []model.Collection{}
	}
	for index := range values {
		normalizeCollection(&values[index])
	}
	return values, err
}

func (s *Service) AdminCollections(ctx context.Context) ([]model.Collection, error) {
	values, err := s.store.EditorialCollections(ctx)
	if values == nil {
		values = []model.Collection{}
	}
	for index := range values {
		normalizeCollection(&values[index])
	}
	return values, err
}

func (s *Service) Collection(ctx context.Context, slug string, page, size int) (*model.CollectionDetail, error) {
	value, err := s.store.PublicCollection(ctx, strings.TrimSpace(slug), bounded(page, 1, 100000, 1), bounded(size, 12, 60, 24))
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.NotFound("collection", strings.TrimSpace(slug))
	}
	normalizeCollection(&value.Collection)
	normalizeCards(value.Images)
	value.Page = bounded(page, 1, 100000, 1)
	value.PageSize = bounded(size, 12, 60, 24)
	if value.ItemCount > 0 {
		value.TotalPages = int(math.Ceil(float64(value.ItemCount) / float64(value.PageSize)))
	}
	return value, nil
}

func (s *Service) AdminCollection(ctx context.Context, rawID string, page, size int) (*model.CollectionDetail, error) {
	id, err := DatabaseID(rawID)
	if err != nil {
		return nil, galleryerr.NotFound("collection", strings.TrimSpace(rawID))
	}
	page = bounded(page, 1, 100000, 1)
	size = bounded(size, 12, 60, 24)
	value, err := s.store.CollectionDetail(ctx, id, page, size)
	if err != nil {
		return nil, err
	}
	if value == nil || value.Kind != collection.KindEditorial || value.OwnerID != "gallery" {
		return nil, galleryerr.NotFound("collection", rawID)
	}
	normalizeCollection(&value.Collection)
	normalizeCards(value.Images)
	value.Page, value.PageSize = page, size
	if value.ItemCount > 0 {
		value.TotalPages = int(math.Ceil(float64(value.ItemCount) / float64(size)))
	}
	return value, nil
}

func (s *Service) CreateEditorialCollection(ctx context.Context, operator string, input model.EditorialCollectionInput) (*model.Collection, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	input.Visibility = defaultString(strings.TrimSpace(input.Visibility), "private")
	if input.Name == "" {
		return nil, galleryerr.Validation("name", "name is required")
	}
	if input.Slug == "" || strings.ContainsAny(input.Slug, " /?#") {
		return nil, galleryerr.Validation("slug", "slug must be a URL-safe segment")
	}
	if !oneOf(input.Visibility, "private", "public") {
		return nil, galleryerr.Validation("visibility", "visibility must be private or public")
	}
	value, err := s.store.CreateEditorialCollection(ctx, strings.TrimSpace(operator), input)
	if value != nil {
		value.ID = PublicID(value.ID)
	}
	return value, err
}

func (s *Service) UpdateEditorialCollection(ctx context.Context, rawID string, input model.EditorialCollectionUpdateInput) (*model.Collection, error) {
	id, err := DatabaseID(rawID)
	if err != nil {
		return nil, galleryerr.NotFound("collection", rawID)
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	input.Visibility = strings.TrimSpace(input.Visibility)
	input.SEOTitle = strings.TrimSpace(input.SEOTitle)
	input.SEODescription = strings.TrimSpace(input.SEODescription)
	if input.Name == "" || input.Slug == "" || strings.ContainsAny(input.Slug, " /?#") {
		return nil, galleryerr.Validation("collection", "name and a URL-safe slug are required")
	}
	if !oneOf(input.Visibility, "private", "public") {
		return nil, galleryerr.Validation("visibility", "visibility must be private or public")
	}
	if input.CoverImageID != "" {
		input.CoverImageID, err = DatabaseID(input.CoverImageID)
		if err != nil {
			return nil, galleryerr.Validation("coverImageId", "cover image ID must be a UUID or compact UUID")
		}
	}
	value, err := s.store.UpdateEditorialCollection(ctx, id, input)
	if value != nil {
		normalizeCollection(value)
	}
	return value, err
}

func (s *Service) ReorderEditorialMembers(ctx context.Context, rawID string, input model.EditorialCollectionOrderInput) (*model.Collection, error) {
	id, err := DatabaseID(rawID)
	if err != nil {
		return nil, galleryerr.NotFound("collection", rawID)
	}
	seen := make(map[string]struct{}, len(input.ImageIDs))
	for index, rawImageID := range input.ImageIDs {
		imageID, parseErr := DatabaseID(rawImageID)
		if parseErr != nil {
			return nil, galleryerr.Validation("imageIds", "image IDs must be UUIDs or compact UUIDs")
		}
		if _, exists := seen[imageID]; exists {
			return nil, galleryerr.Validation("imageIds", "image IDs must be unique")
		}
		seen[imageID] = struct{}{}
		input.ImageIDs[index] = imageID
	}
	value, err := s.store.ReorderEditorialMembers(ctx, id, input.Version, input.ImageIDs)
	if value != nil {
		normalizeCollection(value)
	}
	return value, err
}

func (s *Service) MutateEditorialMembers(ctx context.Context, rawCollectionID string, input model.MemberMutationInput) (*model.Collection, error) {
	if s.collections == nil {
		return nil, galleryerr.NotInitialized("collection")
	}
	collectionID, err := DatabaseID(rawCollectionID)
	if err != nil {
		return nil, galleryerr.NotFound("collection", rawCollectionID)
	}
	for index, rawID := range input.Add {
		id, parseErr := DatabaseID(rawID)
		if parseErr != nil {
			return nil, galleryerr.Validation("add", "image IDs must be UUIDs or compact UUIDs")
		}
		input.Add[index] = id
	}
	for index, rawID := range input.Remove {
		id, parseErr := DatabaseID(rawID)
		if parseErr != nil {
			return nil, galleryerr.Validation("remove", "image IDs must be UUIDs or compact UUIDs")
		}
		input.Remove[index] = id
	}
	value, err := s.collections.Mutate(ctx, "gallery", collectionID, input.Version, input.Add, input.Remove)
	if value != nil {
		value.ID = PublicID(value.ID)
	}
	return value, err
}

func (s *Service) Ranking(ctx context.Context, kind, window string) (*model.Ranking, error) {
	kind = defaultString(strings.TrimSpace(kind), "trending")
	window = defaultString(strings.TrimSpace(window), "7d")
	if !oneOf(kind, "trending", "most_viewed", "most_favorited") || !oneOf(window, "24h", "7d", "30d", "all") {
		return nil, galleryerr.Validation("ranking", "unsupported ranking kind or window")
	}
	value, err := s.store.Ranking(ctx, kind, window, 60)
	if err != nil {
		return nil, err
	}
	if value == nil {
		value = &model.Ranking{Kind: kind, Window: window, Images: []model.ImageCard{}}
	}
	value.Kind, value.Window = kind, window
	value.Images = nonNilCards(value.Images)
	normalizeCards(value.Images)
	return value, nil
}

func (s *Service) Submit(ctx context.Context, subject model.Subject, input model.SubmissionInput) (*model.Submission, error) {
	if err := normalizeSubject(&subject); err != nil {
		return nil, err
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.SourceURL = strings.TrimSpace(input.SourceURL)
	input.AltText = defaultString(strings.TrimSpace(input.AltText), input.Title)
	input.PrimaryCategoryID = strings.TrimSpace(input.PrimaryCategoryID)
	if input.Title == "" {
		return nil, galleryerr.Validation("title", "title is required")
	}
	if _, err := identifier.Parse(input.AssetID); err != nil {
		return nil, galleryerr.Validation("assetId", "assetId must be a UUID")
	}
	for index, rawID := range input.CategoryIDs {
		id, parseErr := DatabaseID(rawID)
		if parseErr != nil {
			return nil, galleryerr.Validation("categoryIds", "category IDs must be UUIDs or compact UUIDs")
		}
		input.CategoryIDs[index] = id
	}
	if input.PrimaryCategoryID != "" {
		id, parseErr := DatabaseID(input.PrimaryCategoryID)
		if parseErr != nil {
			return nil, galleryerr.Validation("primaryCategoryId", "primary category must be a UUID or compact UUID")
		}
		input.PrimaryCategoryID = id
	}
	for selectionIndex := range input.Facets {
		facetID, parseErr := DatabaseID(input.Facets[selectionIndex].FacetID)
		if parseErr != nil {
			return nil, galleryerr.Validation("facets", "facet IDs must be UUIDs or compact UUIDs")
		}
		input.Facets[selectionIndex].FacetID = facetID
		for valueIndex, rawID := range input.Facets[selectionIndex].ValueIDs {
			valueID, valueErr := DatabaseID(rawID)
			if valueErr != nil {
				return nil, galleryerr.Validation("facets", "facet value IDs must be UUIDs or compact UUIDs")
			}
			input.Facets[selectionIndex].ValueIDs[valueIndex] = valueID
		}
	}
	classified, err := s.classifySubmission(ctx, input)
	if err != nil {
		return nil, err
	}
	input.Classification = classified
	if input.SourceURL != "" {
		parsed, err := url.ParseRequestURI(input.SourceURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return nil, galleryerr.Validation("sourceUrl", "sourceUrl must be an absolute HTTP or HTTPS URL")
		}
	}
	review := "pending"
	if subject.Kind == "operator" || (subject.Kind == "user" && subject.Verified) {
		review = "not_required"
	}
	value, err := s.store.CreateSubmission(ctx, subject, input, review)
	if err != nil {
		return nil, err
	}
	if s.assets != nil {
		if err := s.assets.RegisterSubmission(ctx, subject.Bearer, value.AssetID, value.ID, value.Title); err != nil {
			_ = s.store.FailSubmission(ctx, value.ID, "asset_reference_failed")
			return nil, err
		}
	}
	normalizeSubmission(value)
	return value, nil
}

func (s *Service) classifySubmission(ctx context.Context, input model.SubmissionInput) (model.ClassificationWrite, error) {
	catalog, err := s.classificationCatalog(ctx)
	if err != nil {
		return model.ClassificationWrite{}, err
	}
	facets := make([]classification.FacetSelection, 0, len(input.Facets))
	for _, selection := range input.Facets {
		facets = append(facets, classification.FacetSelection{
			FacetID:  selection.FacetID,
			ValueIDs: append([]string(nil), selection.ValueIDs...),
		})
	}
	preparation := catalog.Classify(classification.ClassifyRequest{
		PolicyKey:         "gallery.image.public",
		CategoryIDs:       append([]string(nil), input.CategoryIDs...),
		PrimaryCategoryID: input.PrimaryCategoryID,
		Facets:            facets,
		Tags:              append([]string(nil), input.Tags...),
	})
	factRequest := preparation.FactRequest()
	matches, freshnessToken, err := s.store.ClassificationTagMatches(ctx, factRequest.TagLookups)
	if err != nil {
		return model.ClassificationWrite{}, err
	}
	result := preparation.Complete(classification.ClassifyFacts{
		CatalogRevision: factRequest.CatalogRevision,
		RequestToken:    factRequest.RequestToken,
		FreshnessToken:  freshnessToken,
		TagMatches:      matches,
	})
	if result.Outcome != classification.OutcomeAccepted {
		return model.ClassificationWrite{}, classificationValidation(result.Diagnostics)
	}
	write := model.ClassificationWrite{
		CatalogRevision:   result.CatalogRevision,
		PrimaryCategoryID: result.Assignments.PrimaryCategoryID,
		Categories:        make([]model.CategoryAssignment, 0, len(result.Assignments.Categories)),
		Facets:            make([]model.FacetValueAssignment, 0, len(result.Assignments.Facets)),
		Tags:              make([]model.TagAssignment, 0, len(result.Assignments.Tags)),
		TagProposals:      make([]model.TagProposalInput, 0, len(result.TagProposals)),
		TagCreations:      make([]model.TagCreationInput, 0, len(result.TagCreations)),
	}
	for _, assignment := range result.Assignments.Categories {
		write.Categories = append(write.Categories, model.CategoryAssignment{CategoryID: assignment.CategoryID})
	}
	for _, assignment := range result.Assignments.Facets {
		write.Facets = append(write.Facets, model.FacetValueAssignment{FacetID: assignment.FacetID, ValueID: assignment.ValueID})
	}
	for _, assignment := range result.Assignments.Tags {
		write.Tags = append(write.Tags, model.TagAssignment{TagID: assignment.TagID})
	}
	for _, proposal := range result.TagProposals {
		write.TagProposals = append(write.TagProposals, model.TagProposalInput{
			LookupKey: proposal.LookupKey, DisplayValue: proposal.DisplayValue,
		})
	}
	for _, creation := range result.TagCreations {
		write.TagCreations = append(write.TagCreations, model.TagCreationInput{
			LookupKey: creation.LookupKey, DisplayValue: creation.DisplayValue,
		})
	}
	return write, nil
}

func (s *Service) classificationCatalog(ctx context.Context) (*classification.Catalog, error) {
	s.catalogMu.Lock()
	defer s.catalogMu.Unlock()

	now := s.clock()
	if s.catalog != nil && now.Sub(s.catalogCheckedAt) < 30*time.Second {
		return s.catalog, nil
	}
	revision, err := s.store.ClassificationRevision(ctx)
	if err != nil {
		return nil, err
	}
	if s.catalog != nil && revision == s.catalogRevision {
		s.catalogCheckedAt = now
		return s.catalog, nil
	}
	snapshot, err := s.store.ClassificationSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	if snapshot.Revision != revision {
		return nil, galleryerr.NotInitialized("classification_revision")
	}
	compiled := classification.Compile(snapshot)
	if compiled.Outcome != classification.OutcomeAccepted || compiled.Catalog == nil {
		return nil, galleryerr.NotInitialized("classification_catalog")
	}
	s.catalog = compiled.Catalog
	s.catalogRevision = compiled.CatalogRevision
	s.catalogCheckedAt = now
	return s.catalog, nil
}

func (s *Service) PreviewClassificationGovernance(ctx context.Context, input model.ClassificationGovernancePreviewInput) (*model.ClassificationGovernancePreview, error) {
	_, result, err := s.prepareClassificationGovernance(ctx, input.Command)
	if err != nil {
		return nil, err
	}
	return governancePreview(result), nil
}

func (s *Service) ClassificationCatalog(ctx context.Context) (*model.ClassificationCatalog, error) {
	snapshot, err := s.store.ClassificationSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	result := &model.ClassificationCatalog{
		Revision: snapshot.Revision, Categories: make([]model.ClassificationCatalogNode, 0, len(snapshot.Categories)),
		Facets: make([]model.ClassificationCatalogFacet, 0, len(snapshot.Facets)),
	}
	for _, category := range snapshot.Categories {
		result.Categories = append(result.Categories, classificationCatalogNode(
			category.ID, category.ParentID, category.Slug, category.Name, category.Status,
			category.EditorialPosition, category.ReplacementID,
		))
	}
	valuesByFacet := make(map[string][]model.ClassificationCatalogNode, len(snapshot.Facets))
	for _, value := range snapshot.FacetValues {
		valuesByFacet[value.FacetID] = append(valuesByFacet[value.FacetID], classificationCatalogNode(
			value.ID, value.ParentID, value.Slug, value.Name, value.Status,
			value.EditorialPosition, value.ReplacementID,
		))
	}
	for _, facet := range snapshot.Facets {
		values := valuesByFacet[facet.ID]
		if values == nil {
			values = []model.ClassificationCatalogNode{}
		}
		result.Facets = append(result.Facets, model.ClassificationCatalogFacet{
			ID: facet.ID, Slug: facet.Slug, Name: facet.Name, Status: string(facet.Status),
			EditorialPosition: facet.EditorialPosition, ReplacementID: facet.ReplacementID, Values: values,
		})
	}
	return result, nil
}

func (s *Service) ClassificationTags(ctx context.Context, cursor string, size int) (*model.ClassificationTagPage, error) {
	parsed, err := decodeClassificationTagCursor(cursor)
	if err != nil {
		return nil, galleryerr.Validation("cursor", "cursor is invalid")
	}
	size = bounded(size, 1, 100, 50)
	items, hasMore, err := s.store.ClassificationTags(ctx, parsed, size)
	if err != nil {
		return nil, err
	}
	page := &model.ClassificationTagPage{Items: items}
	if page.Items == nil {
		page.Items = []model.ClassificationTag{}
	}
	if hasMore && len(items) > 0 {
		page.NextCursor, err = encodeClassificationTagCursor(model.ClassificationTagCursor{
			Name: items[len(items)-1].Name,
			ID:   items[len(items)-1].ID,
		})
		if err != nil {
			return nil, err
		}
	}
	for index := range page.Items {
		page.Items[index].ID = PublicID(page.Items[index].ID)
		page.Items[index].ReplacementID = PublicID(page.Items[index].ReplacementID)
	}
	return page, nil
}

func (s *Service) ClassificationTagProposals(ctx context.Context, status string, page, size int) ([]model.ClassificationTagProposal, int, error) {
	status = defaultString(strings.TrimSpace(status), "pending")
	if !oneOf(status, "pending", "approved", "rejected") {
		return nil, 0, galleryerr.Validation("status", "unsupported tag proposal status")
	}
	page = bounded(page, 1, 100000, 1)
	size = bounded(size, 1, 100, 30)
	values, total, err := s.store.ClassificationTagProposals(ctx, status, page, size)
	if values == nil {
		values = []model.ClassificationTagProposal{}
	}
	return values, total, err
}

func (s *Service) ReviewClassificationTagProposal(ctx context.Context, operator, rawID string, input model.ClassificationTagProposalReviewInput) (*model.ClassificationTagProposal, error) {
	operator = strings.TrimSpace(operator)
	if operator == "" {
		return nil, galleryerr.Forbidden()
	}
	id, err := DatabaseID(rawID)
	if err != nil {
		return nil, galleryerr.Validation("proposalId", "proposal ID must be a UUID or compact UUID")
	}
	input.Decision = strings.TrimSpace(input.Decision)
	if !oneOf(input.Decision, "approve", "reject") {
		return nil, galleryerr.Validation("decision", "decision must be approve or reject")
	}
	if input.TargetTagID = strings.TrimSpace(input.TargetTagID); input.TargetTagID != "" {
		input.TargetTagID, err = DatabaseID(input.TargetTagID)
		if err != nil {
			return nil, galleryerr.Validation("targetTagId", "target tag ID must be a UUID or compact UUID")
		}
	}
	value, catalogChanged, err := s.store.ReviewClassificationTagProposal(ctx, operator, id, input)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.NotFound("tag_proposal", rawID)
	}
	if catalogChanged {
		s.invalidateClassificationCatalog()
	}
	return value, nil
}

type classificationTagCursorEnvelope struct {
	Name string `json:"n"`
	ID   string `json:"i"`
}

func decodeClassificationTagCursor(raw string) (model.ClassificationTagCursor, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return model.ClassificationTagCursor{}, nil
	}
	if len(raw) > 2048 {
		return model.ClassificationTagCursor{}, fmt.Errorf("classification tag cursor is too long")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return model.ClassificationTagCursor{}, err
	}
	var envelope classificationTagCursorEnvelope
	if err := json.Unmarshal(decoded, &envelope); err != nil {
		return model.ClassificationTagCursor{}, err
	}
	if strings.TrimSpace(envelope.Name) == "" {
		return model.ClassificationTagCursor{}, fmt.Errorf("classification tag cursor name is empty")
	}
	id, err := DatabaseID(envelope.ID)
	if err != nil {
		return model.ClassificationTagCursor{}, err
	}
	return model.ClassificationTagCursor{Name: envelope.Name, ID: id}, nil
}

func encodeClassificationTagCursor(cursor model.ClassificationTagCursor) (string, error) {
	encoded, err := json.Marshal(classificationTagCursorEnvelope{Name: cursor.Name, ID: cursor.ID})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

func classificationCatalogNode(id, parentID, slug, name string, status classification.Status, position *int, replacementID string) model.ClassificationCatalogNode {
	return model.ClassificationCatalogNode{
		ID: id, ParentID: parentID, Slug: slug, Name: name, Status: string(status),
		EditorialPosition: position, ReplacementID: replacementID,
	}
}

func (s *Service) ExecuteClassificationGovernance(ctx context.Context, operator string, input model.ClassificationGovernanceExecuteInput) (*model.ClassificationGovernanceExecution, error) {
	operator = strings.TrimSpace(operator)
	if operator == "" {
		return nil, galleryerr.Forbidden()
	}
	factRequest, result, err := s.prepareClassificationGovernance(ctx, input.Command)
	if err != nil {
		return nil, err
	}
	if result.Outcome != classification.OutcomePlanned {
		return nil, classificationValidation(result.Diagnostics)
	}
	if input.ExpectedCatalogRevision == 0 || strings.TrimSpace(input.ExpectedRequestToken) == "" ||
		(len(factRequest.Impacts) != 0 && strings.TrimSpace(input.ExpectedImpactToken) == "") {
		return nil, galleryerr.Validation("governance", "a governance preview envelope is required")
	}
	if input.ExpectedCatalogRevision != result.Plan.ExpectedCatalogRevision ||
		strings.TrimSpace(input.ExpectedRequestToken) != result.Plan.ExpectedRequestToken ||
		strings.TrimSpace(input.ExpectedImpactToken) != result.Plan.ExpectedImpactToken {
		return nil, galleryerr.Conflict("classification_governance")
	}
	revision, err := s.store.ExecuteClassificationGovernance(ctx, operator, factRequest, result.Plan)
	if err != nil {
		return nil, err
	}
	s.invalidateClassificationCatalog()
	return &model.ClassificationGovernanceExecution{Applied: true, CatalogRevision: revision}, nil
}

func (s *Service) prepareClassificationGovernance(ctx context.Context, command model.ClassificationGovernanceCommand) (classification.GovernFactRequest, classification.GovernResult, error) {
	normalized, err := normalizeGovernanceCommand(command)
	if err != nil {
		return classification.GovernFactRequest{}, classification.GovernResult{}, err
	}
	catalog, err := s.classificationCatalog(ctx)
	if err != nil {
		return classification.GovernFactRequest{}, classification.GovernResult{}, err
	}
	preparation := catalog.Govern(classification.GovernRequest{
		PolicyKey: "gallery.image.public",
		Command:   normalized,
	})
	factRequest := preparation.FactRequest()
	impacts, freshnessToken, err := s.store.ClassificationGovernanceImpacts(ctx, factRequest.Impacts)
	if err != nil {
		return classification.GovernFactRequest{}, classification.GovernResult{}, err
	}
	result := preparation.Complete(classification.GovernFacts{
		CatalogRevision: factRequest.CatalogRevision,
		RequestToken:    factRequest.RequestToken,
		FreshnessToken:  freshnessToken,
		Impacts:         impacts,
	})
	return factRequest, result, nil
}

func normalizeGovernanceCommand(input model.ClassificationGovernanceCommand) (classification.GovernCommand, error) {
	command := classification.GovernCommand{
		Operation:        classification.GovernOperation(strings.TrimSpace(input.Operation)),
		Kind:             classification.GovernIdentityKind(strings.TrimSpace(input.Kind)),
		Status:           classification.Status(strings.TrimSpace(input.Status)),
		DeleteAllRelated: input.DeleteAllRelated,
	}
	var err error
	if command.ID, err = governanceDatabaseID("id", input.ID, false); err != nil {
		return classification.GovernCommand{}, err
	}
	if command.TargetID, err = governanceDatabaseID("targetId", input.TargetID, true); err != nil {
		return classification.GovernCommand{}, err
	}
	if command.ParentID, err = governanceDatabaseID("parentId", input.ParentID, true); err != nil {
		return classification.GovernCommand{}, err
	}
	for _, move := range input.ChildPlan {
		childID, childErr := governanceDatabaseID("childPlan.childId", move.ChildID, false)
		if childErr != nil {
			return classification.GovernCommand{}, childErr
		}
		parentID, parentErr := governanceDatabaseID("childPlan.parentId", move.ParentID, true)
		if parentErr != nil {
			return classification.GovernCommand{}, parentErr
		}
		command.ChildPlan = append(command.ChildPlan, classification.ChildMove{ChildID: childID, ParentID: parentID})
	}
	return command, nil
}

func governanceDatabaseID(field, raw string, optional bool) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" && optional {
		return "", nil
	}
	id, err := DatabaseID(raw)
	if err != nil {
		return "", galleryerr.Validation(field, field+" must be a UUID or compact UUID")
	}
	return id, nil
}

func governancePreview(result classification.GovernResult) *model.ClassificationGovernancePreview {
	preview := &model.ClassificationGovernancePreview{
		CatalogRevision: result.CatalogRevision,
		Outcome:         string(result.Outcome),
		Diagnostics:     make([]model.ClassificationGovernanceDiagnostic, 0, len(result.Diagnostics)),
		Plan: model.ClassificationGovernancePlan{
			ExpectedCatalogRevision: result.Plan.ExpectedCatalogRevision,
			ExpectedRequestToken:    result.Plan.ExpectedRequestToken,
			ExpectedImpactToken:     result.Plan.ExpectedImpactToken,
			Steps:                   make([]model.ClassificationGovernanceStep, 0, len(result.Plan.Steps)),
		},
	}
	for _, diagnostic := range result.Diagnostics {
		preview.Diagnostics = append(preview.Diagnostics, model.ClassificationGovernanceDiagnostic{
			Code: string(diagnostic.Code), Path: append([]string(nil), diagnostic.Path...),
			Reference: diagnostic.Reference, Params: diagnostic.Params,
		})
	}
	for _, step := range result.Plan.Steps {
		preview.Plan.Steps = append(preview.Plan.Steps, model.ClassificationGovernanceStep{
			Kind: string(step.Kind), IdentityKind: string(step.IdentityKind), SourceID: step.SourceID,
			TargetID: step.TargetID, ParentID: step.ParentID, Status: string(step.Status), AffectedCount: step.AffectedCount,
		})
	}
	return preview
}

func (s *Service) invalidateClassificationCatalog() {
	s.catalogMu.Lock()
	defer s.catalogMu.Unlock()
	s.catalog = nil
	s.catalogRevision = 0
	s.catalogCheckedAt = time.Time{}
}

func (s *Service) MarkClassificationCatalogStale() {
	s.catalogMu.Lock()
	defer s.catalogMu.Unlock()
	s.catalogCheckedAt = time.Time{}
}

func (s *Service) RefreshClassificationCatalog(ctx context.Context) error {
	s.MarkClassificationCatalogStale()
	_, err := s.classificationCatalog(ctx)
	return err
}

func classificationValidation(diagnostics []classification.Diagnostic) error {
	field := "classification"
	if len(diagnostics) == 0 {
		return galleryerr.Validation(field, "classification was rejected")
	}
	if len(diagnostics[0].Path) > 0 {
		switch diagnostics[0].Path[0] {
		case "categoryIds", "categories":
			field = "categoryIds"
		case "primaryCategoryId":
			field = "primaryCategoryId"
		case "facets":
			field = "facets"
		case "tags":
			field = "tags"
		}
	}
	return galleryerr.Validation(field, string(diagnostics[0].Code))
}

func categoryCandidates(values []classification.CandidateNode) []model.ClassificationNode {
	result := make([]model.ClassificationNode, 0, len(values))
	for _, value := range values {
		result = append(result, model.ClassificationNode{
			ID: PublicID(value.ID), ParentID: PublicID(value.ParentID), Slug: value.Slug, Name: value.Name,
			Count: value.Count, Selected: value.Selected,
		})
	}
	return result
}

func facetCandidates(values []classification.CandidateFacet) []model.ClassificationFacet {
	result := make([]model.ClassificationFacet, 0, len(values))
	for _, value := range values {
		result = append(result, model.ClassificationFacet{
			ID: PublicID(value.ID), Slug: value.Slug, Name: value.Name, Values: categoryCandidates(value.Values),
		})
	}
	return result
}

func MySubmissionQuery(page, size int, outcome, processingState, reviewState string) model.MySubmissionQuery {
	return model.MySubmissionQuery{
		Page: bounded(page, 1, 100000, 1), PageSize: bounded(size, 10, 60, 20),
		Outcome: strings.TrimSpace(outcome), ProcessingState: strings.TrimSpace(processingState), ReviewState: strings.TrimSpace(reviewState),
	}
}

func (s *Service) MySubmissions(ctx context.Context, subject model.Subject, query model.MySubmissionQuery) ([]model.Submission, int, error) {
	if err := normalizeSubject(&subject); err != nil {
		return nil, 0, err
	}
	query = MySubmissionQuery(query.Page, query.PageSize, query.Outcome, query.ProcessingState, query.ReviewState)
	if query.Outcome != "" && !oneOf(query.Outcome, "pending", "published", "duplicate", "rejected", "withdrawn", "failed") {
		return nil, 0, galleryerr.Validation("outcome", "unsupported submission outcome")
	}
	if query.ProcessingState != "" && !oneOf(query.ProcessingState, "queued", "processing", "ready", "failed") {
		return nil, 0, galleryerr.Validation("processingState", "unsupported processing state")
	}
	if query.ReviewState != "" && !oneOf(query.ReviewState, "not_required", "pending", "approved", "rejected") {
		return nil, 0, galleryerr.Validation("reviewState", "unsupported review state")
	}
	values, total, err := s.store.MySubmissions(ctx, subject, query)
	if values == nil {
		values = []model.Submission{}
	}
	for index := range values {
		normalizeSubmission(&values[index])
	}
	return values, total, err
}

func (s *Service) Withdraw(ctx context.Context, subject model.Subject, rawID string) (*model.Submission, error) {
	if err := normalizeSubject(&subject); err != nil {
		return nil, err
	}
	id, err := DatabaseID(rawID)
	if err != nil {
		return nil, galleryerr.NotFound("submission", rawID)
	}
	value, err := s.store.WithdrawSubmission(ctx, subject, id)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.InvalidState("submission", "not_withdrawable")
	}
	if s.assets != nil {
		_ = s.assets.UnregisterSubmission(ctx, subject.Bearer, value.AssetID, id)
	}
	normalizeSubmission(value)
	return value, nil
}

func (s *Service) Favorites(ctx context.Context, userID string, page, size int, order string) (*model.CollectionDetail, error) {
	if s.collections == nil {
		return nil, galleryerr.NotInitialized("collection")
	}
	value, err := s.collections.EnsureFavorites(ctx, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	page, size, order = bounded(page, 1, 100000, 1), bounded(size, 12, 60, 24), strings.TrimSpace(order)
	if order == "" {
		order = "newest"
	}
	if !oneOf(order, "newest", "oldest", "title_asc", "title_desc") {
		return nil, galleryerr.Validation("sort", "unsupported favorites sort")
	}
	detail, err := s.store.FavoritesDetail(ctx, value.ID, page, size, order)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		detail = &model.CollectionDetail{Collection: *value, Images: []model.ImageCard{}}
	}
	detail.Page, detail.PageSize = page, size
	if detail.ItemCount > 0 {
		detail.TotalPages = int(math.Ceil(float64(detail.ItemCount) / float64(size)))
	}
	detail.ID = PublicID(detail.ID)
	normalizeCards(detail.Images)
	return detail, nil
}

func (s *Service) SetFavorite(ctx context.Context, userID, rawImageID string, favorite bool, expectedVersion int64) (*model.Collection, error) {
	if s.collections == nil {
		return nil, galleryerr.NotInitialized("collection")
	}
	imageID, err := DatabaseID(rawImageID)
	if err != nil {
		return nil, galleryerr.NotFound("image", rawImageID)
	}
	favorites, err := s.collections.EnsureFavorites(ctx, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	if expectedVersion == 0 {
		expectedVersion = favorites.Version
	}
	var add, remove []string
	if favorite {
		add = []string{imageID}
	} else {
		remove = []string{imageID}
	}
	value, err := s.collections.Mutate(ctx, userID, favorites.ID, expectedVersion, add, remove)
	if err == nil {
		eventType := "favorite"
		if !favorite {
			eventType = "unfavorite"
		}
		_ = s.store.RecordEvent(ctx, model.Subject{Kind: "user", ID: userID}, imageID, model.EventInput{Type: eventType})
	}
	if value != nil {
		value.ID = PublicID(value.ID)
	}
	return value, err
}

func (s *Service) Report(ctx context.Context, subject model.Subject, rawImageID string, input model.CaseInput) (*model.Case, error) {
	if subject.ID != "" {
		if err := normalizeSubject(&subject); err != nil {
			return nil, err
		}
	}
	imageID, err := DatabaseID(rawImageID)
	if err != nil {
		return nil, galleryerr.NotFound("image", rawImageID)
	}
	input.Kind = defaultString(strings.TrimSpace(input.Kind), "report")
	if !oneOf(input.Kind, "report", "source_correction") {
		return nil, galleryerr.Validation("kind", "public cases must be reports or source corrections")
	}
	input.Reason = strings.TrimSpace(input.Reason)
	input.Description = strings.TrimSpace(input.Description)
	input.ProposedSourceURL = strings.TrimSpace(input.ProposedSourceURL)
	if input.Kind == "report" && input.Reason == "" {
		return nil, galleryerr.Validation("reason", "reason is required")
	}
	if input.Kind == "source_correction" {
		parsed, parseErr := url.ParseRequestURI(input.ProposedSourceURL)
		if parseErr != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return nil, galleryerr.Validation("proposedSourceUrl", "an absolute HTTP or HTTPS URL is required")
		}
	}
	value, err := s.store.CreateCase(ctx, subject, imageID, input)
	if err == nil && input.Kind == "report" {
		_ = s.store.RecordEvent(ctx, subject, imageID, model.EventInput{Type: "report"})
	}
	if value != nil {
		value.ID, value.ImageID = PublicID(value.ID), PublicID(value.ImageID)
	}
	return value, err
}

func (s *Service) AdminOverview(ctx context.Context, days int) (*model.AdminOverview, error) {
	days = bounded(days, 7, 30, 14)
	value, err := s.store.AdminOverview(ctx, days)
	if err != nil {
		return nil, err
	}
	value.Days = days
	if value.Series == nil {
		value.Series = []model.AdminTrafficPoint{}
	}
	if value.TopImages == nil {
		value.TopImages = []model.AdminTrafficImage{}
	}
	for index := range value.TopImages {
		value.TopImages[index].ID = PublicID(value.TopImages[index].ID)
	}
	return value, nil
}

func (s *Service) AdminImages(ctx context.Context, query model.AdminImageQuery) (*model.AdminImagePage, error) {
	query.Search = strings.TrimSpace(query.Search)
	query.SortBy = defaultString(strings.TrimSpace(query.SortBy), "createdAt")
	query.SortOrder = defaultString(strings.TrimSpace(query.SortOrder), "desc")
	query.Page, query.PageSize = bounded(query.Page, 1, 100000, 1), bounded(query.PageSize, 12, 60, 24)
	if !oneOf(query.SortBy, "createdAt", "updatedAt", "title", "views") {
		return nil, galleryerr.Validation("sortBy", "unsupported admin image sort field")
	}
	if !oneOf(query.SortOrder, "asc", "desc") {
		return nil, galleryerr.Validation("sortOrder", "sort order must be asc or desc")
	}
	query.ProcessingState = strings.TrimSpace(query.ProcessingState)
	query.ReviewState = strings.TrimSpace(query.ReviewState)
	query.PublicationState = strings.TrimSpace(query.PublicationState)
	query.SafetyState = strings.TrimSpace(query.SafetyState)
	var err error
	if query.CategoryID != "" {
		query.CategoryID, err = DatabaseID(query.CategoryID)
		if err != nil {
			return nil, galleryerr.Validation("categoryId", "category must be a UUID or compact UUID")
		}
	}
	if query.FacetValueID != "" {
		query.FacetValueID, err = DatabaseID(query.FacetValueID)
		if err != nil {
			return nil, galleryerr.Validation("facetValueId", "facet value must be a UUID or compact UUID")
		}
	}
	if query.ProcessingState != "" && !oneOf(query.ProcessingState, "queued", "processing", "ready", "failed") {
		return nil, galleryerr.Validation("processingState", "unsupported processing state")
	}
	if query.ReviewState != "" && !oneOf(query.ReviewState, "not_required", "pending", "approved", "rejected") {
		return nil, galleryerr.Validation("reviewState", "unsupported review state")
	}
	if query.PublicationState != "" && !oneOf(query.PublicationState, "draft", "published", "hidden", "deleted") {
		return nil, galleryerr.Validation("publicationState", "unsupported publication state")
	}
	if query.SafetyState != "" && !oneOf(query.SafetyState, "pending", "safe", "uncertain", "blocked", "unavailable") {
		return nil, galleryerr.Validation("safetyState", "unsupported safety state")
	}
	values, total, err := s.store.AdminImages(ctx, query)
	if err != nil {
		return nil, err
	}
	if values == nil {
		values = []model.AdminImage{}
	}
	for index := range values {
		normalizeAdminImage(&values[index])
	}
	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(query.PageSize)))
	}
	counts, err := s.store.AdminImageCounts(ctx)
	if err != nil {
		return nil, err
	}
	if counts == nil {
		counts = map[string]int{}
	}
	return &model.AdminImagePage{Items: values, Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: pages, Counts: counts}, nil
}

func (s *Service) AdminImage(ctx context.Context, rawID string) (*model.AdminImage, error) {
	id, err := DatabaseID(rawID)
	if err != nil {
		return nil, galleryerr.NotFound("image", rawID)
	}
	value, err := s.store.AdminImage(ctx, id)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.NotFound("image", rawID)
	}
	normalizeAdminImage(value)
	return value, nil
}

func (s *Service) UpdateAdminImage(ctx context.Context, rawID string, input model.AdminImageUpdateInput) (*model.AdminImage, error) {
	id, err := DatabaseID(rawID)
	if err != nil {
		return nil, galleryerr.NotFound("image", rawID)
	}
	input.Title, input.Description = strings.TrimSpace(input.Title), strings.TrimSpace(input.Description)
	input.AltText, input.SourceURL = strings.TrimSpace(input.AltText), strings.TrimSpace(input.SourceURL)
	input.ExpectedUpdatedAt = strings.TrimSpace(input.ExpectedUpdatedAt)
	if input.Title == "" || len([]rune(input.Title)) > 160 {
		return nil, galleryerr.Validation("title", "title is required and must be at most 160 characters")
	}
	if input.AltText == "" {
		return nil, galleryerr.Validation("altText", "alt text is required")
	}
	if _, parseErr := time.Parse(time.RFC3339Nano, input.ExpectedUpdatedAt); parseErr != nil {
		return nil, galleryerr.Validation("expectedUpdatedAt", "a current RFC3339 timestamp is required")
	}
	if input.SourceURL != "" {
		parsed, parseErr := url.ParseRequestURI(input.SourceURL)
		if parseErr != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return nil, galleryerr.Validation("sourceUrl", "source URL must be HTTP or HTTPS")
		}
	}
	if input.Classification != nil {
		classificationInput := input.Classification
		classificationInput.PrimaryCategoryID, err = DatabaseID(strings.TrimSpace(classificationInput.PrimaryCategoryID))
		if err != nil {
			return nil, galleryerr.Validation("primaryCategoryId", "an active primary category is required")
		}
		for index, rawValueID := range classificationInput.FacetValueIDs {
			classificationInput.FacetValueIDs[index], err = DatabaseID(strings.TrimSpace(rawValueID))
			if err != nil {
				return nil, galleryerr.Validation("facetValueIds", "facet values must be UUIDs or compact UUIDs")
			}
		}
		for index, rawTagID := range classificationInput.TagIDs {
			classificationInput.TagIDs[index], err = DatabaseID(strings.TrimSpace(rawTagID))
			if err != nil {
				return nil, galleryerr.Validation("tagIds", "tags must be UUIDs or compact UUIDs")
			}
		}
	}
	value, err := s.store.UpdateAdminImage(ctx, id, input)
	if value != nil {
		normalizeAdminImage(value)
	}
	return value, err
}

func normalizeAdminImage(value *model.AdminImage) {
	value.ID = PublicID(value.ID)
	value.PrimaryCategoryID = PublicID(value.PrimaryCategoryID)
	value.Metrics = model.Metrics{Views: value.ViewCount, Favorites: value.FavoriteCount}
	for index := range value.Facets {
		value.Facets[index].FacetID = PublicID(value.Facets[index].FacetID)
		value.Facets[index].ValueID = PublicID(value.Facets[index].ValueID)
	}
	for index := range value.Tags {
		value.Tags[index].ID = PublicID(value.Tags[index].ID)
	}
}

func (s *Service) BulkHideImages(ctx context.Context, operator string, input model.BulkImageHideInput) []model.BulkImageActionResult {
	reason := strings.TrimSpace(input.Reason)
	results := make([]model.BulkImageActionResult, 0, len(input.ImageIDs))
	seen := map[string]struct{}{}
	for _, rawID := range input.ImageIDs {
		publicID := strings.TrimSpace(rawID)
		if _, duplicate := seen[publicID]; duplicate {
			continue
		}
		seen[publicID] = struct{}{}
		result := model.BulkImageActionResult{ImageID: publicID}
		if reason == "" {
			result.Error = "reason_required"
		} else if err := s.HideImage(ctx, operator, publicID, reason); err != nil {
			result.Error = "not_hidden"
		} else {
			result.Success = true
		}
		results = append(results, result)
	}
	return results
}

func (s *Service) BulkImages(ctx context.Context, operator string, input model.BulkImageActionInput) []model.BulkImageActionResult {
	input.Action = strings.TrimSpace(input.Action)
	input.Reason = strings.TrimSpace(input.Reason)
	categoryID := ""
	categoryInvalid := false
	if input.Action == "set_primary_category" {
		var err error
		categoryID, err = DatabaseID(input.PrimaryCategoryID)
		categoryInvalid = err != nil
	}
	results := make([]model.BulkImageActionResult, 0, min(len(input.ImageIDs), 60))
	seen := map[string]struct{}{}
	for _, rawID := range input.ImageIDs {
		if len(results) >= 60 {
			break
		}
		publicID := strings.TrimSpace(rawID)
		if _, duplicate := seen[publicID]; duplicate {
			continue
		}
		seen[publicID] = struct{}{}
		result := model.BulkImageActionResult{ImageID: publicID}
		switch {
		case input.Action == "hide" && input.Reason == "":
			result.Error = "reason_required"
		case input.Action == "hide":
			if err := s.HideImage(ctx, operator, publicID, input.Reason); err != nil {
				result.Error = "not_hidden"
			} else {
				result.Success = true
			}
		case input.Action == "set_primary_category" && categoryInvalid:
			result.Error = "category_required"
		case input.Action == "set_primary_category":
			imageID, err := DatabaseID(publicID)
			if err != nil {
				result.Error = "image_not_found"
			} else if err := s.store.SetImagePrimaryCategory(ctx, imageID, categoryID); err != nil {
				result.Error = "category_not_set"
			} else {
				result.Success = true
			}
		default:
			result.Error = "unsupported_action"
		}
		results = append(results, result)
	}
	return results
}

func (s *Service) ReviewQueue(ctx context.Context, query model.AdminSubmissionQuery) (*model.AdminSubmissionPage, error) {
	query.Search = strings.TrimSpace(query.Search)
	query.SortBy = defaultString(strings.TrimSpace(query.SortBy), "createdAt")
	query.SortOrder = defaultString(strings.TrimSpace(query.SortOrder), "asc")
	query.Page, query.PageSize = bounded(query.Page, 1, 100000, 1), bounded(query.PageSize, 10, 60, 20)
	if !oneOf(query.SortBy, "createdAt", "updatedAt", "title") {
		return nil, galleryerr.Validation("sortBy", "unsupported submission sort field")
	}
	if !oneOf(query.SortOrder, "asc", "desc") {
		return nil, galleryerr.Validation("sortOrder", "sort order must be asc or desc")
	}
	query.ProcessingState = strings.TrimSpace(query.ProcessingState)
	query.ReviewState = strings.TrimSpace(query.ReviewState)
	query.SafetyState = strings.TrimSpace(query.SafetyState)
	query.Outcome = strings.TrimSpace(query.Outcome)
	if query.ProcessingState != "" && !oneOf(query.ProcessingState, "queued", "processing", "ready", "failed") {
		return nil, galleryerr.Validation("processingState", "unsupported processing state")
	}
	if query.ReviewState != "" && !oneOf(query.ReviewState, "not_required", "pending", "approved", "rejected") {
		return nil, galleryerr.Validation("reviewState", "unsupported review state")
	}
	if query.SafetyState != "" && !oneOf(query.SafetyState, "pending", "safe", "uncertain", "blocked", "unavailable") {
		return nil, galleryerr.Validation("safetyState", "unsupported safety state")
	}
	if query.Outcome != "" && !oneOf(query.Outcome, "pending", "published", "duplicate", "rejected", "withdrawn", "failed") {
		return nil, galleryerr.Validation("outcome", "unsupported submission outcome")
	}
	values, total, err := s.store.ReviewQueue(ctx, query)
	if values == nil {
		values = []model.Submission{}
	}
	for index := range values {
		normalizeSubmission(&values[index])
	}
	if err != nil {
		return nil, err
	}
	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(query.PageSize)))
	}
	return &model.AdminSubmissionPage{Items: values, Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: pages}, nil
}

func (s *Service) ReviewSubmission(ctx context.Context, operator, rawID string, input model.SubmissionReviewInput) (*model.Submission, error) {
	id, err := DatabaseID(rawID)
	if err != nil {
		return nil, galleryerr.NotFound("submission", rawID)
	}
	input.Decision = strings.TrimSpace(input.Decision)
	input.Note = strings.TrimSpace(input.Note)
	if !oneOf(input.Decision, "approve", "reject") {
		return nil, galleryerr.Validation("decision", "decision must be approve or reject")
	}
	if input.Decision == "reject" && input.Note == "" {
		return nil, galleryerr.Validation("note", "a rejection note is required")
	}
	value, err := s.store.ReviewSubmission(ctx, strings.TrimSpace(operator), id, input)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.InvalidState("submission", "not_reviewable")
	}
	if input.Decision == "approve" && value.Outcome == "published" && value.ImageID != "" && s.assets != nil {
		if err := s.assets.PublishImage(ctx, value.AssetID, value.ImageID, value.Title); err != nil {
			return nil, err
		}
		if err := s.store.MarkImagePublicRenditionReady(ctx, value.ImageID); err != nil {
			return nil, err
		}
	}
	normalizeSubmission(value)
	return value, nil
}

func (s *Service) BulkReviewSubmissions(ctx context.Context, operator string, input model.BulkSubmissionReviewInput) ([]model.BulkSubmissionReviewResult, error) {
	if len(input.SubmissionIDs) == 0 {
		return nil, galleryerr.Validation("submissionIds", "at least one submission is required")
	}
	if len(input.SubmissionIDs) > 60 {
		return nil, galleryerr.Validation("submissionIds", "at most 60 submissions can be reviewed at once")
	}
	input.Decision = strings.TrimSpace(input.Decision)
	input.Note = strings.TrimSpace(input.Note)
	if !oneOf(input.Decision, "approve", "reject") {
		return nil, galleryerr.Validation("decision", "decision must be approve or reject")
	}
	if input.Decision == "reject" && input.Note == "" {
		return nil, galleryerr.Validation("note", "a rejection note is required")
	}

	results := make([]model.BulkSubmissionReviewResult, 0, len(input.SubmissionIDs))
	seen := make(map[string]struct{}, len(input.SubmissionIDs))
	for _, rawID := range input.SubmissionIDs {
		rawID = strings.TrimSpace(rawID)
		if rawID == "" {
			continue
		}
		if _, exists := seen[rawID]; exists {
			continue
		}
		seen[rawID] = struct{}{}
		result := model.BulkSubmissionReviewResult{SubmissionID: rawID}
		if _, err := s.ReviewSubmission(ctx, operator, rawID, model.SubmissionReviewInput{
			Decision: input.Decision,
			Note:     input.Note,
		}); err != nil {
			result.Error = err.Error()
		} else {
			result.Success = true
		}
		results = append(results, result)
	}
	if len(results) == 0 {
		return nil, galleryerr.Validation("submissionIds", "at least one valid submission is required")
	}
	return results, nil
}

func (s *Service) HideImage(ctx context.Context, operator, rawID, reason string) error {
	id, err := DatabaseID(rawID)
	if err != nil {
		return galleryerr.NotFound("image", rawID)
	}
	return s.store.HideImage(ctx, strings.TrimSpace(operator), id, strings.TrimSpace(reason))
}

func (s *Service) Cases(ctx context.Context, query model.AdminCaseQuery) (*model.AdminCasePage, error) {
	query.Search = strings.TrimSpace(query.Search)
	query.SortBy = defaultString(strings.TrimSpace(query.SortBy), "createdAt")
	query.SortOrder = defaultString(strings.TrimSpace(query.SortOrder), "asc")
	query.Status = defaultString(strings.TrimSpace(query.Status), "open")
	query.Kind = strings.TrimSpace(query.Kind)
	query.Page, query.PageSize = bounded(query.Page, 1, 100000, 1), bounded(query.PageSize, 10, 60, 20)
	if !oneOf(query.SortBy, "createdAt", "updatedAt", "kind", "status") {
		return nil, galleryerr.Validation("sortBy", "unsupported case sort field")
	}
	if !oneOf(query.SortOrder, "asc", "desc") {
		return nil, galleryerr.Validation("sortOrder", "sort order must be asc or desc")
	}
	if query.Status != "all" && !oneOf(query.Status, "open", "reviewing", "resolved", "dismissed") {
		return nil, galleryerr.Validation("status", "unsupported case status")
	}
	if query.Kind != "" && !oneOf(query.Kind, "report", "source_correction", "safety_uncertain", "near_duplicate", "takedown") {
		return nil, galleryerr.Validation("kind", "unsupported case kind")
	}
	values, total, err := s.store.Cases(ctx, query)
	if values == nil {
		values = []model.Case{}
	}
	for index := range values {
		values[index].ID = PublicID(values[index].ID)
		values[index].ImageID = PublicID(values[index].ImageID)
		values[index].SubmissionID = PublicID(values[index].SubmissionID)
	}
	if err != nil {
		return nil, err
	}
	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(query.PageSize)))
	}
	return &model.AdminCasePage{Items: values, Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: pages}, nil
}

func (s *Service) ResolveCase(ctx context.Context, operator, rawID string, input model.CaseResolutionInput) (*model.Case, error) {
	id, err := DatabaseID(rawID)
	if err != nil {
		return nil, galleryerr.NotFound("case", rawID)
	}
	input.Status = strings.TrimSpace(input.Status)
	input.Note = strings.TrimSpace(input.Note)
	input.ExpectedUpdatedAt = strings.TrimSpace(input.ExpectedUpdatedAt)
	if !oneOf(input.Status, "reviewing", "resolved", "dismissed") {
		return nil, galleryerr.Validation("status", "case status must be reviewing, resolved or dismissed")
	}
	if oneOf(input.Status, "resolved", "dismissed") && input.Note == "" {
		return nil, galleryerr.Validation("note", "a resolution note is required")
	}
	if _, parseErr := time.Parse(time.RFC3339Nano, input.ExpectedUpdatedAt); parseErr != nil {
		return nil, galleryerr.Validation("expectedUpdatedAt", "a current RFC3339 timestamp is required")
	}
	value, err := s.store.ResolveCase(ctx, strings.TrimSpace(operator), id, input)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.InvalidState("case", "not_open")
	}
	value.ID, value.ImageID, value.SubmissionID = PublicID(value.ID), PublicID(value.ImageID), PublicID(value.SubmissionID)
	return value, nil
}

func (s *Service) TrackEvent(ctx context.Context, subject model.Subject, rawImageID string, input model.EventInput) error {
	imageID, err := DatabaseID(rawImageID)
	if err != nil {
		return galleryerr.NotFound("image", rawImageID)
	}
	input.Type = strings.TrimSpace(input.Type)
	input.SessionKey = strings.TrimSpace(input.SessionKey)
	if !oneOf(input.Type, "grid_exposure", "qualified_view", "share") {
		return galleryerr.Validation("type", "unsupported public event type")
	}
	if len(input.SessionKey) > 160 {
		return galleryerr.Validation("sessionKey", "session key is too long")
	}
	return s.store.RecordEvent(ctx, subject, imageID, input)
}

func seededDiverse(seed string, candidates []model.ImageCard, limit int) []model.ImageCard {
	type scored struct {
		card  model.ImageCard
		score uint64
	}
	values := make([]scored, 0, len(candidates))
	for _, candidate := range candidates {
		digest := sha256.Sum256([]byte(seed + "\x00" + candidate.ID))
		values = append(values, scored{card: candidate, score: binary.BigEndian.Uint64(digest[:8])})
	}
	sort.SliceStable(values, func(i, j int) bool { return values[i].score < values[j].score })
	result := make([]model.ImageCard, 0, min(limit, len(values)))
	lastCategory := ""
	for len(values) > 0 && len(result) < limit {
		pick := 0
		if values[0].card.PrimaryCategorySlug == lastCategory {
			for index := 1; index < len(values); index++ {
				if values[index].card.PrimaryCategorySlug != lastCategory {
					pick = index
					break
				}
			}
		}
		result = append(result, values[pick].card)
		lastCategory = values[pick].card.PrimaryCategorySlug
		values = append(values[:pick], values[pick+1:]...)
	}
	return result
}

func normalizeCards(values []model.ImageCard) {
	for index := range values {
		values[index].ID = PublicID(values[index].ID)
		values[index].Metrics = model.Metrics{Views: values[index].ViewCount, Favorites: values[index].FavoriteCount}
	}
}

func normalizeDetail(value *model.ImageDetail) {
	if value == nil {
		return
	}
	value.ID = PublicID(value.ID)
	value.Metrics = model.Metrics{Views: value.ViewCount, Favorites: value.FavoriteCount}
	if value.Tags == nil {
		value.Tags = []string{}
	}
	if value.Facets == nil {
		value.Facets = []model.FacetValueAssignment{}
	}
}

func normalizeCollection(value *model.Collection) {
	if value == nil {
		return
	}
	value.ID = PublicID(value.ID)
	value.CoverImageID = PublicID(value.CoverImageID)
}

func normalizeSubmission(value *model.Submission) {
	if value == nil {
		return
	}
	value.ID = PublicID(value.ID)
	value.ImageID = PublicID(value.ImageID)
}

func normalizeSubject(subject *model.Subject) error {
	if subject == nil {
		return galleryerr.Forbidden()
	}
	subject.Kind = defaultString(strings.TrimSpace(subject.Kind), "user")
	subject.ID = strings.TrimSpace(subject.ID)
	if subject.ID == "" || !oneOf(subject.Kind, "user", "guest", "operator") {
		return galleryerr.Forbidden()
	}
	return nil
}

func nonNilCards(values []model.ImageCard) []model.ImageCard {
	if values == nil {
		return []model.ImageCard{}
	}
	return values
}

func bounded(value, low, high, fallback int) int {
	if value == 0 {
		return fallback
	}
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func (s *Service) String() string {
	return fmt.Sprintf("GalleryService(store=%T)", s.store)
}
