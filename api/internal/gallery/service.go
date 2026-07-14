package gallery

import (
	"context"
	"regexp"
	"strings"

	"platform/gokit/facet"
	"platform/products/gallery/api/internal/galleryerr"
	"platform/products/gallery/api/internal/model"
)

var creatorHandlePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{2,31}$`)

type Store interface {
	SiteSettings(context.Context) (*model.SiteSettings, error)
	Featured(context.Context, int) ([]model.ArtworkCard, error)
	Latest(context.Context, int) ([]model.ArtworkCard, error)
	Facets(context.Context) ([]facet.Facet, error)
	FacetValues(context.Context) ([]facet.Value, error)
	Artwork(context.Context, string) (*model.ArtworkDetail, error)
}

type WorkflowStore interface {
	CreatorBySubject(context.Context, string) (*model.CreatorProfile, error)
	CreateCreator(context.Context, model.CreatorRequest) (*model.CreatorProfile, error)
	CreatorArtworks(context.Context, string) ([]model.StudioArtwork, error)
	CreateArtwork(context.Context, string) (*model.StudioArtwork, error)
	CreatorArtwork(context.Context, string, string) (*model.StudioArtwork, error)
	SaveArtwork(context.Context, string, model.ArtworkDraftInput, []facet.Assignment) (*model.StudioArtwork, error)
	AddArtworkAsset(context.Context, string, model.AssetInput) (*model.Asset, error)
	RemoveArtworkAsset(context.Context, string, string) (*model.Asset, error)
	ArtworkAssetCount(context.Context, string) (int, error)
	ArtworkAssignments(context.Context, string) ([]facet.Assignment, error)
	SetArtworkStatus(context.Context, string, string, string, string) (*model.StudioArtwork, error)
	CreatorApplications(context.Context) ([]model.CreatorProfile, error)
	SetCreatorStatus(context.Context, string, string, string, string) (*model.CreatorProfile, error)
	ReviewQueue(context.Context, string) ([]model.StudioArtwork, error)
}

type CreatorStore interface {
	PublicCreator(context.Context, string) (*model.PublicCreator, error)
	PublishedByCreator(context.Context, string, int) ([]model.ArtworkCard, error)
}

type Service struct {
	store    Store
	workflow WorkflowStore
}

func New(store Store) *Service {
	service := &Service{store: store}
	if workflow, ok := store.(WorkflowStore); ok {
		service.workflow = workflow
	}
	return service
}

func (s *Service) Discovery(ctx context.Context) (*model.Discovery, error) {
	settings, err := s.store.SiteSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, galleryerr.NotInitialized("site_settings")
	}
	featured, err := s.store.Featured(ctx, 6)
	if err != nil {
		return nil, err
	}
	latest, err := s.store.Latest(ctx, 24)
	if err != nil {
		return nil, err
	}
	facets, err := s.store.Facets(ctx)
	if err != nil {
		return nil, err
	}
	values, err := s.store.FacetValues(ctx)
	if err != nil {
		return nil, err
	}
	if featured == nil {
		featured = []model.ArtworkCard{}
	}
	if latest == nil {
		latest = []model.ArtworkCard{}
	}
	catalog, err := facet.NewCatalog(facets, values)
	if err != nil {
		return nil, err
	}
	return &model.Discovery{
		Site: *settings, Featured: featured, Latest: latest,
		Facets: catalog.Facets(), FacetValues: catalog.Values(),
	}, nil
}

func (s *Service) MyCreator(ctx context.Context, subject string) (*model.CreatorProfile, error) {
	if s.workflow == nil {
		return nil, galleryerr.NotInitialized("workflow")
	}
	return s.workflow.CreatorBySubject(ctx, strings.TrimSpace(subject))
}

func (s *Service) RequestCreator(ctx context.Context, subject string, input model.CreatorRequest) (*model.CreatorProfile, error) {
	if s.workflow == nil {
		return nil, galleryerr.NotInitialized("workflow")
	}
	input.AccountSub = strings.TrimSpace(subject)
	input.Handle = strings.ToLower(strings.TrimSpace(input.Handle))
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.ApplicationNote = strings.TrimSpace(input.ApplicationNote)
	if input.AccountSub == "" {
		return nil, galleryerr.Forbidden()
	}
	if input.Handle == "" || input.DisplayName == "" {
		return nil, galleryerr.Validation("creator", "handle and displayName are required")
	}
	if !creatorHandlePattern.MatchString(input.Handle) {
		return nil, galleryerr.Validation("handle", "handle must be 3-32 lowercase letters, numbers or hyphens")
	}
	if existing, err := s.workflow.CreatorBySubject(ctx, input.AccountSub); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, galleryerr.Conflict("creator")
	}
	return s.workflow.CreateCreator(ctx, input)
}

func (s *Service) MyArtworks(ctx context.Context, subject string) ([]model.StudioArtwork, error) {
	creator, err := s.activeCreator(ctx, subject)
	if err != nil {
		return nil, err
	}
	values, err := s.workflow.CreatorArtworks(ctx, creator.ID)
	if values == nil {
		values = []model.StudioArtwork{}
	}
	return values, err
}

func (s *Service) CreateDraft(ctx context.Context, subject string) (*model.StudioArtwork, error) {
	creator, err := s.activeCreator(ctx, subject)
	if err != nil {
		return nil, err
	}
	return s.workflow.CreateArtwork(ctx, creator.ID)
}

func (s *Service) MyArtwork(ctx context.Context, subject, artworkID string) (*model.StudioArtwork, error) {
	creator, err := s.activeCreator(ctx, subject)
	if err != nil {
		return nil, err
	}
	value, err := s.workflow.CreatorArtwork(ctx, creator.ID, strings.TrimSpace(artworkID))
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.NotFound("artwork", strings.TrimSpace(artworkID))
	}
	return value, nil
}

func (s *Service) SaveDraft(ctx context.Context, subject, artworkID string, input model.ArtworkDraftInput) (*model.StudioArtwork, error) {
	artwork, err := s.MyArtwork(ctx, subject, artworkID)
	if err != nil {
		return nil, err
	}
	if artwork.Status != "draft" && artwork.Status != "rejected" {
		return nil, galleryerr.InvalidState("artwork", artwork.Status)
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Visibility = defaultValue(input.Visibility, "public")
	input.ContentRating = defaultValue(input.ContentRating, "general")
	input.AIUsage = defaultValue(input.AIUsage, "none")
	input.AITrainingPermission = defaultValue(input.AITrainingPermission, "unspecified")
	input.RightsBasis = defaultValue(input.RightsBasis, "original")
	input.License = defaultValue(input.License, "all_rights_reserved")
	if !oneOf(input.Visibility, "public", "unlisted", "private") {
		return nil, galleryerr.Validation("visibility", "unsupported visibility")
	}
	if !oneOf(input.ContentRating, "general", "sensitive", "adult") {
		return nil, galleryerr.Validation("contentRating", "unsupported content rating")
	}
	if !oneOf(input.AIUsage, "none", "assistive", "mostly_generated") {
		return nil, galleryerr.Validation("aiUsage", "unsupported AI usage")
	}
	if !oneOf(input.AITrainingPermission, "unspecified", "allow", "disallow") {
		return nil, galleryerr.Validation("aiTrainingPermission", "unsupported AI training permission")
	}
	if !oneOf(input.RightsBasis, "original", "authorized_repost", "public_domain", "licensed_material") {
		return nil, galleryerr.Validation("rightsBasis", "unsupported rights basis")
	}
	catalog, err := s.facetCatalog(ctx)
	if err != nil {
		return nil, err
	}
	assignments, err := catalog.NormalizeSelections(input.FacetSelections, facet.PhaseDraft)
	if err != nil {
		return nil, galleryerr.Validation("facetSelections", err.Error())
	}
	value, err := s.workflow.SaveArtwork(ctx, artwork.ID, input, assignments)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.InvalidState("artwork", "not_editable")
	}
	return value, nil
}

func (s *Service) AddAsset(ctx context.Context, subject, artworkID string, input model.AssetInput) (*model.Asset, error) {
	artwork, err := s.MyArtwork(ctx, subject, artworkID)
	if err != nil {
		return nil, err
	}
	if artwork.Status != "draft" && artwork.Status != "rejected" {
		return nil, galleryerr.InvalidState("artwork", artwork.Status)
	}
	input.AssetID = strings.TrimSpace(input.AssetID)
	input.AltText = strings.TrimSpace(input.AltText)
	if input.AssetID == "" {
		return nil, galleryerr.Validation("assetId", "assetId is required")
	}
	if input.Width < 0 || input.Height < 0 {
		return nil, galleryerr.Validation("dimensions", "image dimensions cannot be negative")
	}
	value, err := s.workflow.AddArtworkAsset(ctx, artwork.ID, input)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.InvalidState("artwork", "not_editable")
	}
	return value, nil
}

func (s *Service) RemoveAsset(ctx context.Context, subject, artworkID, artworkAssetID string) (*model.Asset, error) {
	artwork, err := s.MyArtwork(ctx, subject, artworkID)
	if err != nil {
		return nil, err
	}
	if artwork.Status != "draft" && artwork.Status != "rejected" {
		return nil, galleryerr.InvalidState("artwork", artwork.Status)
	}
	value, err := s.workflow.RemoveArtworkAsset(ctx, artwork.ID, strings.TrimSpace(artworkAssetID))
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.NotFound("artwork_asset", strings.TrimSpace(artworkAssetID))
	}
	return value, nil
}

func (s *Service) SubmitArtwork(ctx context.Context, subject, artworkID string) (*model.StudioArtwork, error) {
	artwork, err := s.MyArtwork(ctx, subject, artworkID)
	if err != nil {
		return nil, err
	}
	if artwork.Status != "draft" && artwork.Status != "rejected" {
		return nil, galleryerr.InvalidState("artwork", artwork.Status)
	}
	if strings.TrimSpace(artwork.Title) == "" {
		return nil, galleryerr.Validation("title", "title is required before submission")
	}
	assetCount, err := s.workflow.ArtworkAssetCount(ctx, artwork.ID)
	if err != nil {
		return nil, err
	}
	if assetCount == 0 {
		return nil, galleryerr.Validation("assets", "at least one image is required before submission")
	}
	assignments, err := s.workflow.ArtworkAssignments(ctx, artwork.ID)
	if err != nil {
		return nil, err
	}
	selections := selectionsFromAssignments(assignments)
	catalog, err := s.facetCatalog(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := catalog.NormalizeSelections(selections, facet.PhasePublish); err != nil {
		return nil, galleryerr.Validation("facetSelections", err.Error())
	}
	value, err := s.workflow.SetArtworkStatus(ctx, artwork.ID, "pending_review", "", "")
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.InvalidState("artwork", "not_submittable")
	}
	return value, nil
}

func (s *Service) CreatorApplications(ctx context.Context) ([]model.CreatorProfile, error) {
	values, err := s.workflow.CreatorApplications(ctx)
	if values == nil {
		values = []model.CreatorProfile{}
	}
	return values, err
}

func (s *Service) ReviewCreator(ctx context.Context, operator, creatorID, decision, note string) (*model.CreatorProfile, error) {
	status := "active"
	if decision == "reject" {
		status = "rejected"
	} else if decision != "approve" {
		return nil, galleryerr.Validation("decision", "decision must be approve or reject")
	}
	value, err := s.workflow.SetCreatorStatus(ctx, strings.TrimSpace(creatorID), status, strings.TrimSpace(operator), strings.TrimSpace(note))
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.InvalidState("creator", "not_pending")
	}
	return value, nil
}

func (s *Service) ReviewQueue(ctx context.Context, status string) ([]model.StudioArtwork, error) {
	values, err := s.workflow.ReviewQueue(ctx, defaultValue(status, "pending_review"))
	if values == nil {
		values = []model.StudioArtwork{}
	}
	return values, err
}

func (s *Service) ReviewArtwork(ctx context.Context, operator, artworkID string, input model.ReviewInput) (*model.StudioArtwork, error) {
	artwork, err := s.workflow.CreatorArtwork(ctx, "", strings.TrimSpace(artworkID))
	if err != nil {
		return nil, err
	}
	if artwork == nil {
		return nil, galleryerr.NotFound("artwork", strings.TrimSpace(artworkID))
	}
	if artwork.Status != "pending_review" {
		return nil, galleryerr.InvalidState("artwork", artwork.Status)
	}
	status := "published"
	if input.Decision == "reject" {
		if strings.TrimSpace(input.Note) == "" {
			return nil, galleryerr.Validation("note", "a rejection note is required")
		}
		status = "rejected"
	} else if input.Decision != "approve" {
		return nil, galleryerr.Validation("decision", "decision must be approve or reject")
	}
	value, err := s.workflow.SetArtworkStatus(ctx, artwork.ID, status, strings.TrimSpace(operator), strings.TrimSpace(input.Note))
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.InvalidState("artwork", "not_pending_review")
	}
	return value, nil
}

func (s *Service) activeCreator(ctx context.Context, subject string) (*model.CreatorProfile, error) {
	if s.workflow == nil {
		return nil, galleryerr.NotInitialized("workflow")
	}
	creator, err := s.workflow.CreatorBySubject(ctx, strings.TrimSpace(subject))
	if err != nil {
		return nil, err
	}
	if creator == nil || creator.Status != "active" {
		return nil, galleryerr.CreatorNotActive()
	}
	return creator, nil
}

func (s *Service) facetCatalog(ctx context.Context) (*facet.Catalog, error) {
	definitions, err := s.store.Facets(ctx)
	if err != nil {
		return nil, err
	}
	values, err := s.store.FacetValues(ctx)
	if err != nil {
		return nil, err
	}
	return facet.NewCatalog(definitions, values)
}

func defaultValue(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
}

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func selectionsFromAssignments(assignments []facet.Assignment) []facet.Selection {
	byFacet := map[string][]string{}
	order := []string{}
	for _, assignment := range assignments {
		if _, ok := byFacet[assignment.FacetID]; !ok {
			order = append(order, assignment.FacetID)
		}
		byFacet[assignment.FacetID] = append(byFacet[assignment.FacetID], assignment.ValueID)
	}
	values := make([]facet.Selection, 0, len(order))
	for _, facetID := range order {
		values = append(values, facet.Selection{FacetID: facetID, ValueIDs: byFacet[facetID]})
	}
	return values
}

func (s *Service) Artwork(ctx context.Context, id string) (*model.ArtworkDetail, error) {
	id = strings.TrimSpace(id)
	value, err := s.store.Artwork(ctx, id)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, galleryerr.NotFound("artwork", id)
	}
	if value.Tags == nil {
		value.Tags = []string{}
	}
	if value.Assets == nil {
		value.Assets = []model.Asset{}
	}
	return value, nil
}

func (s *Service) CreatorPage(ctx context.Context, handle string) (*model.CreatorPage, error) {
	store, ok := s.store.(CreatorStore)
	if !ok {
		return nil, galleryerr.NotInitialized("creator_page")
	}
	handle = strings.ToLower(strings.TrimSpace(handle))
	creator, err := store.PublicCreator(ctx, handle)
	if err != nil {
		return nil, err
	}
	if creator == nil {
		return nil, galleryerr.NotFound("creator", handle)
	}
	artworks, err := store.PublishedByCreator(ctx, creator.ID, 60)
	if err != nil {
		return nil, err
	}
	if artworks == nil {
		artworks = []model.ArtworkCard{}
	}
	return &model.CreatorPage{Creator: *creator, Artworks: artworks}, nil
}
