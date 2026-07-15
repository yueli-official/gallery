package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"platform/gokit/facet"
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
	ID            string      `json:"id" orm:"id"`
	AssetID       string      `json:"assetId" orm:"asset_id"`
	Title         string      `json:"title" orm:"title"`
	AltText       string      `json:"altText" orm:"alt_text"`
	Width         int         `json:"width" orm:"width"`
	Height        int         `json:"height" orm:"height"`
	DominantColor string      `json:"dominantColor" orm:"dominant_color"`
	Topic         string      `json:"topic" orm:"topic"`
	TopicSlug     string      `json:"topicSlug" orm:"topic_slug"`
	PublishedAt   *gtime.Time `json:"publishedAt" orm:"published_at"`
	Metrics       Metrics     `json:"metrics" orm:"-"`
	ViewCount     int64       `json:"-" orm:"view_count"`
	FavoriteCount int64       `json:"-" orm:"favorite_count"`
}

type ImageDetail struct {
	ImageCard
	Description string             `json:"description" orm:"description"`
	SourceURL   string             `json:"sourceUrl" orm:"source_url"`
	FocusX      float64            `json:"focusX" orm:"focus_x"`
	FocusY      float64            `json:"focusY" orm:"focus_y"`
	Tags        []string           `json:"tags" orm:"-"`
	Facets      []facet.Assignment `json:"facets" orm:"-"`
	Favorited   bool               `json:"favorited" orm:"-"`
}

type ImageQuery struct {
	Search   string
	Sort     string
	Page     int
	PageSize int
	FacetIDs []string
	Tag      string
}

type ImagePage struct {
	Items      []ImageCard `json:"items"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	Total      int         `json:"total"`
	TotalPages int         `json:"totalPages"`
}

type Discovery struct {
	Site        SiteSettings  `json:"site"`
	Seed        string        `json:"seed"`
	Images      []ImageCard   `json:"images"`
	Facets      []facet.Facet `json:"facets"`
	FacetValues []facet.Value `json:"facetValues"`
}

type Collection struct {
	ID           string      `json:"id" orm:"id"`
	Kind         string      `json:"kind" orm:"kind"`
	ResourceKind string      `json:"resourceKind" orm:"resource_kind"`
	OwnerKind    string      `json:"ownerKind" orm:"owner_kind"`
	OwnerID      string      `json:"-" orm:"owner_id"`
	Visibility   string      `json:"visibility" orm:"visibility"`
	Name         string      `json:"name" orm:"name"`
	Description  string      `json:"description" orm:"description"`
	Version      int64       `json:"version" orm:"version"`
	Slug         string      `json:"slug,omitempty" orm:"slug"`
	CoverImageID string      `json:"coverImageId,omitempty" orm:"cover_image_id"`
	ItemCount    int         `json:"itemCount" orm:"item_count"`
	CreatedAt    *gtime.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt    *gtime.Time `json:"updatedAt" orm:"updated_at"`
}

type CollectionDetail struct {
	Collection
	Images []ImageCard `json:"images"`
}

type EditorialCollectionInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
	Visibility  string `json:"visibility"`
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
	AssetID        string             `json:"assetId"`
	Title          string             `json:"title"`
	Description    string             `json:"description"`
	SourceURL      string             `json:"sourceUrl"`
	AltText        string             `json:"altText"`
	TopicID        string             `json:"topicId"`
	Tags           []string           `json:"tags"`
	Facets         []facet.Selection  `json:"facets"`
	Assignments    []facet.Assignment `json:"-"`
	NormalizedTags []TagInput         `json:"-"`
}

// TagInput is the normalized write shape used between the application and
// persistence layers. Public clients continue to submit plain tag names.
type TagInput struct {
	Slug string
	Name string
}

type Submission struct {
	ID              string      `json:"id" orm:"id"`
	SubjectKind     string      `json:"-" orm:"subject_kind"`
	SubjectID       string      `json:"-" orm:"subject_id"`
	AssetID         string      `json:"assetId" orm:"asset_id"`
	ImageID         string      `json:"imageId" orm:"image_id"`
	Title           string      `json:"title" orm:"title"`
	Description     string      `json:"description" orm:"description"`
	SourceURL       string      `json:"sourceUrl" orm:"source_url"`
	AltText         string      `json:"altText" orm:"alt_text"`
	TopicID         string      `json:"topicId" orm:"topic_value_id"`
	ProcessingState string      `json:"processingState" orm:"processing_state"`
	ReviewState     string      `json:"reviewState" orm:"review_state"`
	SafetyState     string      `json:"safetyState" orm:"safety_state"`
	Outcome         string      `json:"outcome" orm:"outcome"`
	FailureCode     string      `json:"failureCode" orm:"failure_code"`
	ReviewNote      string      `json:"reviewNote" orm:"review_note"`
	CreatedAt       *gtime.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt       *gtime.Time `json:"updatedAt" orm:"updated_at"`
}

type SubmissionReviewInput struct {
	Decision string `json:"decision"`
	Note     string `json:"note"`
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
	UpdatedAt         *gtime.Time `json:"updatedAt" orm:"updated_at"`
}

type CaseResolutionInput struct {
	Status string `json:"status"`
	Note   string `json:"note"`
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
