package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"platform/gokit/facet"
)

type SiteSettings struct {
	Name              string `json:"name" orm:"name"`
	Title             string `json:"title" orm:"title"`
	Description       string `json:"description" orm:"description"`
	SearchPlaceholder string `json:"searchPlaceholder" orm:"search_placeholder"`
	FooterTagline     string `json:"footerTagline" orm:"footer_tagline"`
}

type Creator struct {
	ID          string `json:"id" orm:"creator_id"`
	Handle      string `json:"handle" orm:"creator_handle"`
	DisplayName string `json:"displayName" orm:"creator_name"`
}

type ArtworkCard struct {
	ID             string      `json:"id" orm:"id"`
	Title          string      `json:"title" orm:"title"`
	Description    string      `json:"description" orm:"description"`
	CoverURL       string      `json:"coverUrl" orm:"cover_url"`
	PlaceholderURL string      `json:"placeholderUrl" orm:"placeholder_url"`
	Width          int         `json:"width" orm:"width"`
	Height         int         `json:"height" orm:"height"`
	ContentRating  string      `json:"contentRating" orm:"content_rating"`
	AIUsage        string      `json:"aiUsage" orm:"ai_usage"`
	PublishedAt    *gtime.Time `json:"publishedAt" orm:"published_at"`
	Creator        Creator     `json:"creator" orm:"-"`
	CreatorID      string      `json:"-" orm:"creator_id"`
	CreatorHandle  string      `json:"-" orm:"creator_handle"`
	CreatorName    string      `json:"-" orm:"creator_name"`
}

type Asset struct {
	ID             string `json:"id" orm:"id"`
	AssetID        string `json:"assetId" orm:"asset_id"`
	SortOrder      int    `json:"sortOrder" orm:"sort_order"`
	Width          int    `json:"width" orm:"width"`
	Height         int    `json:"height" orm:"height"`
	AltText        string `json:"altText" orm:"alt_text"`
	PlaceholderURL string `json:"placeholderUrl" orm:"placeholder_url"`
	ThumbnailURL   string `json:"thumbnailUrl" orm:"thumbnail_url"`
	CardURL        string `json:"cardUrl" orm:"card_url"`
	DetailURL      string `json:"detailUrl" orm:"detail_url"`
	OriginalURL    string `json:"originalUrl" orm:"original_url"`
}

type ArtworkDetail struct {
	ArtworkCard
	License              string   `json:"license" orm:"license"`
	RightsBasis          string   `json:"rightsBasis" orm:"rights_basis"`
	AITrainingPermission string   `json:"aiTrainingPermission" orm:"ai_training_permission"`
	Tags                 []string `json:"tags"`
	Assets               []Asset  `json:"assets"`
}

type Discovery struct {
	Site        SiteSettings  `json:"site"`
	Featured    []ArtworkCard `json:"featured"`
	Latest      []ArtworkCard `json:"latest"`
	Facets      []facet.Facet `json:"facets"`
	FacetValues []facet.Value `json:"facetValues"`
}

type CreatorProfile struct {
	ID              string      `json:"id" orm:"id"`
	AccountSub      string      `json:"-" orm:"account_sub"`
	Handle          string      `json:"handle" orm:"handle"`
	DisplayName     string      `json:"displayName" orm:"display_name"`
	Bio             string      `json:"bio" orm:"bio"`
	CoverAssetID    string      `json:"coverAssetId" orm:"cover_asset_id"`
	Status          string      `json:"status" orm:"status"`
	ApplicationNote string      `json:"applicationNote" orm:"application_note"`
	ReviewNote      string      `json:"reviewNote" orm:"review_note"`
	CreatedAt       *gtime.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt       *gtime.Time `json:"updatedAt" orm:"updated_at"`
}

type PublicCreator struct {
	ID           string      `json:"id" orm:"id"`
	Handle       string      `json:"handle" orm:"handle"`
	DisplayName  string      `json:"displayName" orm:"display_name"`
	Bio          string      `json:"bio" orm:"bio"`
	CoverAssetID string      `json:"coverAssetId" orm:"cover_asset_id"`
	CreatedAt    *gtime.Time `json:"createdAt" orm:"created_at"`
}

type CreatorRequest struct {
	AccountSub      string `json:"-"`
	Handle          string `json:"handle"`
	DisplayName     string `json:"displayName"`
	ApplicationNote string `json:"applicationNote"`
}

type ArtworkDraftInput struct {
	Title                string            `json:"title"`
	Description          string            `json:"description"`
	Visibility           string            `json:"visibility"`
	ContentRating        string            `json:"contentRating"`
	AIUsage              string            `json:"aiUsage"`
	AITrainingPermission string            `json:"aiTrainingPermission"`
	RightsBasis          string            `json:"rightsBasis"`
	License              string            `json:"license"`
	FacetSelections      []facet.Selection `json:"facetSelections"`
}

type AssetInput struct {
	AssetID string `json:"assetId"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Format  string `json:"format"`
	AltText string `json:"altText"`
}

type StudioArtwork struct {
	ID                   string             `json:"id" orm:"id"`
	CreatorID            string             `json:"creatorId" orm:"creator_id"`
	Title                string             `json:"title" orm:"title"`
	Description          string             `json:"description" orm:"description"`
	Status               string             `json:"status" orm:"status"`
	Visibility           string             `json:"visibility" orm:"visibility"`
	ContentRating        string             `json:"contentRating" orm:"content_rating"`
	AIUsage              string             `json:"aiUsage" orm:"ai_usage"`
	AITrainingPermission string             `json:"aiTrainingPermission" orm:"ai_training_permission"`
	RightsBasis          string             `json:"rightsBasis" orm:"rights_basis"`
	License              string             `json:"license" orm:"license"`
	ReviewNote           string             `json:"reviewNote" orm:"review_note"`
	CreatedAt            *gtime.Time        `json:"createdAt" orm:"created_at"`
	UpdatedAt            *gtime.Time        `json:"updatedAt" orm:"updated_at"`
	PublishedAt          *gtime.Time        `json:"publishedAt" orm:"published_at"`
	Creator              CreatorProfile     `json:"creator" orm:"-"`
	Assets               []Asset            `json:"assets" orm:"-"`
	FacetAssignments     []facet.Assignment `json:"facetAssignments" orm:"-"`
}

type ReviewInput struct {
	Decision string `json:"decision"`
	Note     string `json:"note"`
}

type CreatorPage struct {
	Creator  PublicCreator `json:"creator"`
	Artworks []ArtworkCard `json:"artworks"`
}
