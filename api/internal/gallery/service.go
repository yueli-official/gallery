package gallery

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"

	"platform/gokit/facet"
	"platform/products/gallery/api/internal/collection"
	"platform/products/gallery/api/internal/galleryerr"
	"platform/products/gallery/api/internal/model"
)

type Store interface {
	SiteSettings(context.Context) (*model.SiteSettings, error)
	Facets(context.Context) ([]facet.Facet, error)
	FacetValues(context.Context) ([]facet.Value, error)
	RandomCandidates(context.Context, int) ([]model.ImageCard, error)
	ListImages(context.Context, model.ImageQuery) ([]model.ImageCard, int, error)
	Image(context.Context, string, string) (*model.ImageDetail, error)
	HasTombstone(context.Context, string) (bool, error)
	PublicCollections(context.Context) ([]model.Collection, error)
	EditorialCollections(context.Context) ([]model.Collection, error)
	PublicCollection(context.Context, string, int, int) (*model.CollectionDetail, error)
	CreateEditorialCollection(context.Context, string, model.EditorialCollectionInput) (*model.Collection, error)
	Ranking(context.Context, string, string, int) (*model.Ranking, error)
	CreateSubmission(context.Context, model.Subject, model.SubmissionInput, string) (*model.Submission, error)
	MySubmissions(context.Context, model.Subject, int, int) ([]model.Submission, int, error)
	WithdrawSubmission(context.Context, model.Subject, string) (*model.Submission, error)
	FailSubmission(context.Context, string, string) error
	CollectionDetail(context.Context, string, int, int) (*model.CollectionDetail, error)
	CreateCase(context.Context, model.Subject, string, model.CaseInput) (*model.Case, error)
	AdminOverview(context.Context) (*model.AdminOverview, error)
	ReviewQueue(context.Context, int, int) ([]model.Submission, int, error)
	ReviewSubmission(context.Context, string, string, model.SubmissionReviewInput) (*model.Submission, error)
	HideImage(context.Context, string, string, string) error
	Cases(context.Context, string, int, int) ([]model.Case, int, error)
	ResolveCase(context.Context, string, string, model.CaseResolutionInput) (*model.Case, error)
	RecordEvent(context.Context, model.Subject, string, model.EventInput) error
}

type Service struct {
	store       Store
	collections *collection.Service
	assets      AssetReferencePort
	clock       func() time.Time
}

type AssetReferencePort interface {
	RegisterSubmission(context.Context, string, string, string, string) error
	UnregisterSubmission(context.Context, string, string, string) error
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
	facets, err := s.store.Facets(ctx)
	if err != nil {
		return nil, err
	}
	values, err := s.store.FacetValues(ctx)
	if err != nil {
		return nil, err
	}
	normalizeCards(images)
	if facets == nil {
		facets = []facet.Facet{}
	}
	if values == nil {
		values = []facet.Value{}
	}
	return &model.Discovery{Site: *settings, Seed: seed, Images: images, Facets: facets, FacetValues: values}, nil
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
	for index, rawID := range query.FacetIDs {
		id, err := DatabaseID(rawID)
		if err != nil {
			return nil, galleryerr.Validation("facets", "facet values must be UUIDs or compact UUIDs")
		}
		query.FacetIDs[index] = id
	}
	items, total, err := s.store.ListImages(ctx, query)
	if err != nil {
		return nil, err
	}
	normalizeCards(items)
	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(query.PageSize)))
	}
	return &model.ImagePage{Items: nonNilCards(items), Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: pages}, nil
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

func (s *Service) Collections(ctx context.Context) ([]model.Collection, error) {
	values, err := s.store.PublicCollections(ctx)
	if values == nil {
		values = []model.Collection{}
	}
	for index := range values {
		values[index].ID = PublicID(values[index].ID)
		values[index].CoverImageID = PublicID(values[index].CoverImageID)
	}
	return values, err
}

func (s *Service) AdminCollections(ctx context.Context) ([]model.Collection, error) {
	values, err := s.store.EditorialCollections(ctx)
	if values == nil {
		values = []model.Collection{}
	}
	for index := range values {
		values[index].ID = PublicID(values[index].ID)
		values[index].CoverImageID = PublicID(values[index].CoverImageID)
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
	value.ID = PublicID(value.ID)
	value.CoverImageID = PublicID(value.CoverImageID)
	normalizeCards(value.Images)
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
	input.TopicID = strings.TrimSpace(input.TopicID)
	if input.Title == "" {
		return nil, galleryerr.Validation("title", "title is required")
	}
	if _, err := uuid.Parse(input.AssetID); err != nil {
		return nil, galleryerr.Validation("assetId", "assetId must be a UUID")
	}
	if _, err := uuid.Parse(input.TopicID); err != nil {
		return nil, galleryerr.Validation("topicId", "topicId must be a UUID")
	}
	assignments, err := s.normalizeSubmissionFacets(ctx, input.TopicID, input.Facets)
	if err != nil {
		return nil, err
	}
	input.Assignments = assignments
	input.NormalizedTags, err = normalizeSubmissionTags(input.Tags)
	if err != nil {
		return nil, err
	}
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

func (s *Service) normalizeSubmissionFacets(ctx context.Context, topicID string, selections []facet.Selection) ([]facet.Assignment, error) {
	definitions, err := s.store.Facets(ctx)
	if err != nil {
		return nil, err
	}
	values, err := s.store.FacetValues(ctx)
	if err != nil {
		return nil, err
	}
	var topicFacetID string
	for _, definition := range definitions {
		if definition.Slug == "topic" {
			topicFacetID = definition.ID
			break
		}
	}
	if topicFacetID == "" {
		return nil, galleryerr.NotInitialized("topic_facet")
	}
	selections = append(append([]facet.Selection{}, selections...), facet.Selection{
		FacetID: topicFacetID, ValueIDs: []string{topicID},
	})
	catalog, err := facet.NewCatalog(definitions, values)
	if err != nil {
		return nil, galleryerr.NotInitialized("facet_catalog")
	}
	assignments, err := catalog.NormalizeSelections(selections, facet.PhasePublish)
	if err != nil {
		return nil, galleryerr.Validation("facets", err.Error())
	}
	return assignments, nil
}

func normalizeSubmissionTags(raw []string) ([]model.TagInput, error) {
	const maxTags = 12
	seen := make(map[string]struct{}, len(raw))
	result := make([]model.TagInput, 0, len(raw))
	for _, value := range raw {
		name := strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
		if name == "" {
			continue
		}
		if len([]rune(name)) > 40 {
			return nil, galleryerr.Validation("tags", "each tag must be at most 40 characters")
		}
		slug := tagSlug(name)
		if slug == "" {
			return nil, galleryerr.Validation("tags", "tags must contain a letter or number")
		}
		if _, exists := seen[slug]; exists {
			continue
		}
		seen[slug] = struct{}{}
		result = append(result, model.TagInput{Slug: slug, Name: name})
		if len(result) > maxTags {
			return nil, galleryerr.Validation("tags", "at most 12 tags are allowed")
		}
	}
	return result, nil
}

func tagSlug(value string) string {
	var builder strings.Builder
	separator := false
	for _, character := range strings.ToLower(value) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			if separator && builder.Len() > 0 {
				builder.WriteByte('-')
			}
			builder.WriteRune(character)
			separator = false
			continue
		}
		separator = true
	}
	return builder.String()
}

func (s *Service) MySubmissions(ctx context.Context, subject model.Subject, page, size int) ([]model.Submission, int, error) {
	if err := normalizeSubject(&subject); err != nil {
		return nil, 0, err
	}
	values, total, err := s.store.MySubmissions(ctx, subject, bounded(page, 1, 100000, 1), bounded(size, 10, 60, 20))
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

func (s *Service) Favorites(ctx context.Context, userID string) (*model.CollectionDetail, error) {
	if s.collections == nil {
		return nil, galleryerr.NotInitialized("collection")
	}
	value, err := s.collections.EnsureFavorites(ctx, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	detail, err := s.store.CollectionDetail(ctx, value.ID, 1, 60)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		detail = &model.CollectionDetail{Collection: *value, Images: []model.ImageCard{}}
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

func (s *Service) AdminOverview(ctx context.Context) (*model.AdminOverview, error) {
	return s.store.AdminOverview(ctx)
}

func (s *Service) ReviewQueue(ctx context.Context, page, size int) ([]model.Submission, int, error) {
	values, total, err := s.store.ReviewQueue(ctx, bounded(page, 1, 100000, 1), bounded(size, 10, 60, 20))
	if values == nil {
		values = []model.Submission{}
	}
	for index := range values {
		normalizeSubmission(&values[index])
	}
	return values, total, err
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
	normalizeSubmission(value)
	return value, nil
}

func (s *Service) HideImage(ctx context.Context, operator, rawID, reason string) error {
	id, err := DatabaseID(rawID)
	if err != nil {
		return galleryerr.NotFound("image", rawID)
	}
	return s.store.HideImage(ctx, strings.TrimSpace(operator), id, strings.TrimSpace(reason))
}

func (s *Service) Cases(ctx context.Context, status string, page, size int) ([]model.Case, int, error) {
	status = defaultString(strings.TrimSpace(status), "open")
	if !oneOf(status, "open", "reviewing", "resolved", "dismissed") {
		return nil, 0, galleryerr.Validation("status", "unsupported case status")
	}
	values, total, err := s.store.Cases(ctx, status, bounded(page, 1, 100000, 1), bounded(size, 10, 60, 20))
	if values == nil {
		values = []model.Case{}
	}
	for index := range values {
		values[index].ID = PublicID(values[index].ID)
		values[index].ImageID = PublicID(values[index].ImageID)
		values[index].SubmissionID = PublicID(values[index].SubmissionID)
	}
	return values, total, err
}

func (s *Service) ResolveCase(ctx context.Context, operator, rawID string, input model.CaseResolutionInput) (*model.Case, error) {
	id, err := DatabaseID(rawID)
	if err != nil {
		return nil, galleryerr.NotFound("case", rawID)
	}
	input.Status = strings.TrimSpace(input.Status)
	input.Note = strings.TrimSpace(input.Note)
	if !oneOf(input.Status, "reviewing", "resolved", "dismissed") {
		return nil, galleryerr.Validation("status", "case status must be reviewing, resolved or dismissed")
	}
	if oneOf(input.Status, "resolved", "dismissed") && input.Note == "" {
		return nil, galleryerr.Validation("note", "a resolution note is required")
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
	lastTopic := ""
	for len(values) > 0 && len(result) < limit {
		pick := 0
		if values[0].card.TopicSlug == lastTopic {
			for index := 1; index < len(values); index++ {
				if values[index].card.TopicSlug != lastTopic {
					pick = index
					break
				}
			}
		}
		result = append(result, values[pick].card)
		lastTopic = values[pick].card.TopicSlug
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
		value.Facets = []facet.Assignment{}
	}
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
