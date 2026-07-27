package model

import (
	"time"

	"github.com/gogf/gf/v2/os/gtime"
)

type SiteSettings struct {
	Name                string `json:"name" orm:"name"`
	Title               string `json:"title" orm:"title"`
	Description         string `json:"description" orm:"description"`
	SearchPlaceholder   string `json:"searchPlaceholder" orm:"search_placeholder"`
	FooterTagline       string `json:"footerTagline" orm:"footer_tagline"`
	RandomBatchSize     int    `json:"randomBatchSize" orm:"random_batch_size"`
	RandomCandidateSize int    `json:"randomCandidateSize" orm:"random_candidate_size"`
}

type Metrics struct {
	Views     int64 `json:"views" orm:"view_count"`
	Favorites int64 `json:"favorites" orm:"favorite_count"`
}

type ImageCard struct {
	ID                  string      `json:"id" orm:"id"`
	AssetID             string      `json:"assetId" orm:"asset_id"`
	Title               string      `json:"title" orm:"title"`
	AltText             string      `json:"altText" orm:"alt_text"`
	Width               int         `json:"width" orm:"width"`
	Height              int         `json:"height" orm:"height"`
	DominantColor       string      `json:"dominantColor" orm:"dominant_color"`
	PrimaryCategory     string      `json:"primaryCategory" orm:"primary_category"`
	PrimaryCategorySlug string      `json:"primaryCategorySlug" orm:"primary_category_slug"`
	PublishedAt         *gtime.Time `json:"publishedAt" orm:"published_at"`
	Metrics             Metrics     `json:"metrics" orm:"-"`
	ViewCount           int64       `json:"-" orm:"view_count"`
	FavoriteCount       int64       `json:"-" orm:"favorite_count"`
}

type ImagePublicationCandidate struct {
	ID      string `orm:"id"`
	AssetID string `orm:"asset_id"`
	Title   string `orm:"title"`
}

type ImageDetail struct {
	ImageCard
	Description string                 `json:"description" orm:"description"`
	SourceURL   string                 `json:"sourceUrl" orm:"source_url"`
	FocusX      float64                `json:"focusX" orm:"focus_x"`
	FocusY      float64                `json:"focusY" orm:"focus_y"`
	Tags        []string               `json:"tags" orm:"-"`
	Facets      []FacetValueAssignment `json:"facets" orm:"-"`
	Favorited   bool                   `json:"favorited" orm:"-"`
}

type RelatedImageReason struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

type RelatedImage struct {
	ImageCard
	Reasons []RelatedImageReason `json:"reasons" orm:"-"`
}

type ImageQuery struct {
	Search       string
	Sort         string
	Page         int
	PageSize     int
	CategoryRefs []string
	FacetRefs    []string
	Tag          string
}

type ImagePage struct {
	Items       []ImageCard                `json:"items"`
	Page        int                        `json:"page"`
	PageSize    int                        `json:"pageSize"`
	Total       int                        `json:"total"`
	TotalPages  int                        `json:"totalPages"`
	Diagnostics []ClassificationDiagnostic `json:"diagnostics"`
	Categories  []ClassificationNode       `json:"categories"`
	Facets      []ClassificationFacet      `json:"facets"`
}

type AdminImage struct {
	ImageCard
	PrimaryCategoryID    string                 `json:"primaryCategoryId" orm:"primary_category_id"`
	Description          string                 `json:"description" orm:"description"`
	SourceURL            string                 `json:"sourceUrl" orm:"source_url"`
	ProcessingState      string                 `json:"processingState" orm:"processing_state"`
	ReviewState          string                 `json:"reviewState" orm:"review_state"`
	PublicationState     string                 `json:"publicationState" orm:"publication_state"`
	SafetyState          string                 `json:"safetyState" orm:"safety_state"`
	PublicRenditionReady bool                   `json:"publicRenditionReady" orm:"public_rendition_ready"`
	Facets               []FacetValueAssignment `json:"facets" orm:"-"`
	Tags                 []AdminImageTag        `json:"tags" orm:"-"`
	CreatedAt            *gtime.Time            `json:"createdAt" orm:"created_at"`
	UpdatedAt            *time.Time             `json:"updatedAt" orm:"updated_at"`
}

type AdminImageTag struct {
	ID   string `json:"id" orm:"id"`
	Name string `json:"name" orm:"name"`
}

type AdminImageQuery struct {
	Search           string
	Sort             string
	Page             int
	PageSize         int
	ProcessingState  string
	ReviewState      string
	PublicationState string
	SafetyState      string
	CategoryID       string
	FacetValueID     string
}

type AdminImagePage struct {
	Items      []AdminImage   `json:"items"`
	Page       int            `json:"page"`
	PageSize   int            `json:"pageSize"`
	Total      int            `json:"total"`
	TotalPages int            `json:"totalPages"`
	Counts     map[string]int `json:"counts"`
}

type AdminImageUpdateInput struct {
	ExpectedUpdatedAt string                         `json:"expectedUpdatedAt"`
	Title             string                         `json:"title"`
	Description       string                         `json:"description"`
	AltText           string                         `json:"altText"`
	SourceURL         string                         `json:"sourceUrl"`
	Classification    *AdminImageClassificationInput `json:"classification,omitempty"`
}

type AdminImageClassificationInput struct {
	PrimaryCategoryID string   `json:"primaryCategoryId"`
	FacetValueIDs     []string `json:"facetValueIds"`
	TagIDs            []string `json:"tagIds"`
}

type BulkImageHideInput struct {
	ImageIDs []string `json:"imageIds"`
	Reason   string   `json:"reason"`
}

type BulkImageActionInput struct {
	ImageIDs          []string `json:"imageIds"`
	Action            string   `json:"action"`
	Reason            string   `json:"reason"`
	PrimaryCategoryID string   `json:"primaryCategoryId"`
}

type BulkImageActionResult struct {
	ImageID string `json:"imageId"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ClassificationDiagnostic struct {
	Code      string            `json:"code"`
	Path      []string          `json:"path"`
	Reference string            `json:"reference,omitempty"`
	Params    map[string]string `json:"params,omitempty"`
}

type Discovery struct {
	Site       SiteSettings          `json:"site"`
	Seed       string                `json:"seed"`
	Images     []ImageCard           `json:"images"`
	Categories []ClassificationNode  `json:"categories"`
	Facets     []ClassificationFacet `json:"facets"`
}

type SubmissionOptions struct {
	Categories []ClassificationNode  `json:"categories"`
	Facets     []ClassificationFacet `json:"facets"`
}

type ClassificationNode struct {
	ID       string `json:"id"`
	ParentID string `json:"parentId,omitempty"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Count    int64  `json:"count"`
	Selected bool   `json:"selected"`
}

type ClassificationFacet struct {
	ID     string               `json:"id"`
	Slug   string               `json:"slug"`
	Name   string               `json:"name"`
	Values []ClassificationNode `json:"values"`
}

type Collection struct {
	ID             string      `json:"id" orm:"id"`
	Kind           string      `json:"kind" orm:"kind"`
	ResourceKind   string      `json:"resourceKind" orm:"resource_kind"`
	OwnerKind      string      `json:"ownerKind" orm:"owner_kind"`
	OwnerID        string      `json:"-" orm:"owner_id"`
	Visibility     string      `json:"visibility" orm:"visibility"`
	Name           string      `json:"name" orm:"name"`
	Description    string      `json:"description" orm:"description"`
	Version        int64       `json:"version" orm:"version"`
	Slug           string      `json:"slug,omitempty" orm:"slug"`
	CoverImageID   string      `json:"coverImageId,omitempty" orm:"cover_image_id"`
	CoverAssetID   string      `json:"coverAssetId,omitempty" orm:"cover_asset_id"`
	CoverAltText   string      `json:"coverAltText,omitempty" orm:"cover_alt_text"`
	CoverWidth     int         `json:"coverWidth,omitempty" orm:"cover_width"`
	CoverHeight    int         `json:"coverHeight,omitempty" orm:"cover_height"`
	CoverColor     string      `json:"coverColor,omitempty" orm:"cover_color"`
	SEOTitle       string      `json:"seoTitle,omitempty" orm:"seo_title"`
	SEODescription string      `json:"seoDescription,omitempty" orm:"seo_description"`
	ItemCount      int         `json:"itemCount" orm:"item_count"`
	CreatedAt      *gtime.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt      *gtime.Time `json:"updatedAt" orm:"updated_at"`
}

type CollectionDetail struct {
	Collection
	Images     []ImageCard `json:"images"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

type EditorialCollectionInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
	Visibility  string `json:"visibility"`
}

type EditorialCollectionUpdateInput struct {
	Version        int64  `json:"version"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Slug           string `json:"slug"`
	Visibility     string `json:"visibility"`
	CoverImageID   string `json:"coverImageId"`
	SEOTitle       string `json:"seoTitle"`
	SEODescription string `json:"seoDescription"`
}

type EditorialCollectionOrderInput struct {
	Version  int64    `json:"version"`
	ImageIDs []string `json:"imageIds"`
}

type MemberMutationInput struct {
	Version int64    `json:"version"`
	Add     []string `json:"add"`
	Remove  []string `json:"remove"`
}

type Subject struct {
	Kind     string
	ID       string
	Verified bool
	Bearer   string
}

type SubmissionInput struct {
	AssetID           string              `json:"assetId"`
	Title             string              `json:"title"`
	Description       string              `json:"description"`
	SourceURL         string              `json:"sourceUrl"`
	AltText           string              `json:"altText"`
	CategoryIDs       []string            `json:"categoryIds"`
	PrimaryCategoryID string              `json:"primaryCategoryId"`
	Tags              []string            `json:"tags"`
	Facets            []FacetSelection    `json:"facets"`
	Classification    ClassificationWrite `json:"-"`
}

type MySubmissionQuery struct {
	Page            int
	PageSize        int
	Outcome         string
	ProcessingState string
	ReviewState     string
}

type FacetSelection struct {
	FacetID  string   `json:"facetId"`
	ValueIDs []string `json:"valueIds"`
}

type CategoryAssignment struct {
	CategoryID string
}

type FacetValueAssignment struct {
	FacetID string `json:"facetId" orm:"facet_id"`
	ValueID string `json:"valueId" orm:"value_id"`
}

type TagAssignment struct {
	TagID string
}

type TagProposalInput struct {
	LookupKey    string
	DisplayValue string
}

type TagCreationInput struct {
	LookupKey    string
	DisplayValue string
}

type ClassificationWrite struct {
	CatalogRevision   uint64
	Categories        []CategoryAssignment
	PrimaryCategoryID string
	Facets            []FacetValueAssignment
	Tags              []TagAssignment
	TagProposals      []TagProposalInput
	TagCreations      []TagCreationInput
}

type ClassificationChildMove struct {
	ChildID  string `json:"childId"`
	ParentID string `json:"parentId"`
}

type ClassificationGovernanceCommand struct {
	Operation        string                    `json:"operation"`
	Kind             string                    `json:"kind"`
	ID               string                    `json:"id"`
	TargetID         string                    `json:"targetId"`
	ParentID         string                    `json:"parentId"`
	Status           string                    `json:"status"`
	ChildPlan        []ClassificationChildMove `json:"childPlan"`
	DeleteAllRelated bool                      `json:"deleteAllRelated"`
}

type ClassificationGovernancePreviewInput struct {
	Command ClassificationGovernanceCommand `json:"command"`
}

type ClassificationGovernanceExecuteInput struct {
	Command                 ClassificationGovernanceCommand `json:"command"`
	ExpectedCatalogRevision uint64                          `json:"expectedCatalogRevision"`
	ExpectedRequestToken    string                          `json:"expectedRequestToken"`
	ExpectedImpactToken     string                          `json:"expectedImpactToken"`
}

type ClassificationGovernanceDiagnostic struct {
	Code      string            `json:"code"`
	Path      []string          `json:"path"`
	Reference string            `json:"reference"`
	Params    map[string]string `json:"params"`
}

type ClassificationGovernanceStep struct {
	Kind          string `json:"kind"`
	IdentityKind  string `json:"identityKind"`
	SourceID      string `json:"sourceId"`
	TargetID      string `json:"targetId"`
	ParentID      string `json:"parentId"`
	Status        string `json:"status"`
	AffectedCount int64  `json:"affectedCount"`
}

type ClassificationGovernancePlan struct {
	ExpectedCatalogRevision uint64                         `json:"expectedCatalogRevision"`
	ExpectedRequestToken    string                         `json:"expectedRequestToken"`
	ExpectedImpactToken     string                         `json:"expectedImpactToken"`
	Steps                   []ClassificationGovernanceStep `json:"steps"`
}

type ClassificationGovernancePreview struct {
	CatalogRevision uint64                               `json:"catalogRevision"`
	Outcome         string                               `json:"outcome"`
	Diagnostics     []ClassificationGovernanceDiagnostic `json:"diagnostics"`
	Plan            ClassificationGovernancePlan         `json:"plan"`
}

type ClassificationGovernanceExecution struct {
	Applied         bool   `json:"applied"`
	CatalogRevision uint64 `json:"catalogRevision"`
}

type ClassificationCatalogNode struct {
	ID                string `json:"id"`
	ParentID          string `json:"parentId"`
	Slug              string `json:"slug"`
	Name              string `json:"name"`
	Status            string `json:"status"`
	EditorialPosition *int   `json:"editorialPosition"`
	ReplacementID     string `json:"replacementId"`
}

type ClassificationCatalogFacet struct {
	ID                string                      `json:"id"`
	Slug              string                      `json:"slug"`
	Name              string                      `json:"name"`
	Status            string                      `json:"status"`
	EditorialPosition *int                        `json:"editorialPosition"`
	ReplacementID     string                      `json:"replacementId"`
	Values            []ClassificationCatalogNode `json:"values"`
}

type ClassificationCatalog struct {
	Revision   uint64                       `json:"revision"`
	Categories []ClassificationCatalogNode  `json:"categories"`
	Facets     []ClassificationCatalogFacet `json:"facets"`
}

type ClassificationTagCursor struct {
	Name string
	ID   string
}

type ClassificationTag struct {
	ID              string `json:"id" orm:"id"`
	Name            string `json:"name" orm:"name"`
	Slug            string `json:"slug" orm:"slug"`
	Status          string `json:"status" orm:"status"`
	ReplacementID   string `json:"replacementId" orm:"replacement_id"`
	AssignmentCount int64  `json:"assignmentCount" orm:"assignment_count"`
	AliasCount      int64  `json:"aliasCount" orm:"alias_count"`
}

type ClassificationTagPage struct {
	Items      []ClassificationTag `json:"items"`
	NextCursor string              `json:"nextCursor"`
}

type ClassificationTagProposal struct {
	ID            string      `json:"id" orm:"id"`
	SubmissionID  string      `json:"submissionId" orm:"submission_id"`
	InputValue    string      `json:"inputValue" orm:"input_value"`
	LookupKey     string      `json:"lookupKey" orm:"lookup_key"`
	Status        string      `json:"status" orm:"status"`
	ResolvedTagID string      `json:"resolvedTagId" orm:"resolved_tag_id"`
	CreatedAt     *gtime.Time `json:"createdAt" orm:"created_at"`
	ReviewedAt    *gtime.Time `json:"reviewedAt" orm:"reviewed_at"`
}

type ClassificationTagProposalReviewInput struct {
	Decision    string `json:"decision"`
	TargetTagID string `json:"targetTagId"`
}

type Submission struct {
	ID                string      `json:"id" orm:"id"`
	SubjectKind       string      `json:"-" orm:"subject_kind"`
	SubjectID         string      `json:"-" orm:"subject_id"`
	AssetID           string      `json:"assetId" orm:"asset_id"`
	ImageID           string      `json:"imageId" orm:"image_id"`
	Title             string      `json:"title" orm:"title"`
	Description       string      `json:"description" orm:"description"`
	SourceURL         string      `json:"sourceUrl" orm:"source_url"`
	AltText           string      `json:"altText" orm:"alt_text"`
	PrimaryCategoryID string      `json:"primaryCategoryId" orm:"primary_category_id"`
	ProcessingState   string      `json:"processingState" orm:"processing_state"`
	ReviewState       string      `json:"reviewState" orm:"review_state"`
	SafetyState       string      `json:"safetyState" orm:"safety_state"`
	Outcome           string      `json:"outcome" orm:"outcome"`
	FailureCode       string      `json:"failureCode" orm:"failure_code"`
	ReviewNote        string      `json:"reviewNote" orm:"review_note"`
	CreatedAt         *gtime.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt         *gtime.Time `json:"updatedAt" orm:"updated_at"`
}

type SubmissionAssetFacts struct {
	PreviewURL  string
	ContentHash string
	Mime        string
	Width       int
	Height      int
}

type AdminSubmissionQuery struct {
	Search          string
	Sort            string
	Page            int
	PageSize        int
	ProcessingState string
	ReviewState     string
	SafetyState     string
	Outcome         string
}

type AdminSubmissionPage struct {
	Items      []Submission `json:"items"`
	Page       int          `json:"page"`
	PageSize   int          `json:"pageSize"`
	Total      int          `json:"total"`
	TotalPages int          `json:"totalPages"`
}

type SubmissionReviewInput struct {
	Decision string `json:"decision"`
	Note     string `json:"note"`
}

type BulkSubmissionReviewInput struct {
	SubmissionIDs []string `json:"submissionIds"`
	Decision      string   `json:"decision"`
	Note          string   `json:"note"`
}

type BulkSubmissionReviewResult struct {
	SubmissionID string `json:"submissionId"`
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
}

type CaseInput struct {
	Kind              string `json:"kind"`
	Reason            string `json:"reason"`
	Description       string `json:"description"`
	ProposedSourceURL string `json:"proposedSourceUrl"`
}

type Case struct {
	ID                string      `json:"id" orm:"id"`
	ImageID           string      `json:"imageId" orm:"image_id"`
	SubmissionID      string      `json:"submissionId" orm:"submission_id"`
	Kind              string      `json:"kind" orm:"kind"`
	Status            string      `json:"status" orm:"status"`
	Reason            string      `json:"reason" orm:"reason"`
	Description       string      `json:"description" orm:"description"`
	ProposedSourceURL string      `json:"proposedSourceUrl" orm:"proposed_source_url"`
	ResolutionNote    string      `json:"resolutionNote" orm:"resolution_note"`
	CreatedAt         *gtime.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt         *time.Time  `json:"updatedAt" orm:"updated_at"`
}

type AdminCaseQuery struct {
	Search   string
	Sort     string
	Page     int
	PageSize int
	Status   string
	Kind     string
}

type AdminCasePage struct {
	Items      []Case `json:"items"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	Total      int    `json:"total"`
	TotalPages int    `json:"totalPages"`
}

type CaseResolutionInput struct {
	ExpectedUpdatedAt string `json:"expectedUpdatedAt"`
	Status            string `json:"status"`
	Note              string `json:"note"`
}

type EventInput struct {
	Type       string `json:"type"`
	SessionKey string `json:"sessionKey"`
}

type Ranking struct {
	Kind      string      `json:"kind"`
	Window    string      `json:"window"`
	Generated *gtime.Time `json:"generatedAt"`
	Images    []ImageCard `json:"images"`
}

type AdminOverview struct {
	PendingSubmissions int `json:"pendingSubmissions" orm:"pending_submissions"`
	OpenCases          int `json:"openCases" orm:"open_cases"`
	PublishedImages    int `json:"publishedImages" orm:"published_images"`
	FailedProcessing   int `json:"failedProcessing" orm:"failed_processing"`
}
