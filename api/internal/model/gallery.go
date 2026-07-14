package model

import "github.com/gogf/gf/v2/os/gtime"

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

type Category struct {
	ID           string `json:"id" orm:"id"`
	Slug         string `json:"slug" orm:"slug"`
	Name         string `json:"name" orm:"name"`
	Description  string `json:"description" orm:"description"`
	ArtworkCount int    `json:"artworkCount" orm:"artwork_count"`
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
	Site       SiteSettings  `json:"site"`
	Featured   []ArtworkCard `json:"featured"`
	Latest     []ArtworkCard `json:"latest"`
	Categories []Category    `json:"categories"`
}
