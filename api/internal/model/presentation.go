package model

type SidebarAppearance struct {
	Mode    string  `json:"mode" orm:"sidebar_mode"`
	ImageID string  `json:"imageId" orm:"sidebar_image_id"`
	AssetID string  `json:"assetId" orm:"asset_id"`
	Overlay float64 `json:"overlay" orm:"sidebar_overlay"`
}

type ClassificationPresentation struct {
	Kind    string `json:"kind" orm:"kind"`
	NodeID  string `json:"nodeId" orm:"node_id"`
	Mode    string `json:"mode" orm:"mode"`
	Icon    string `json:"icon" orm:"icon"`
	ImageID string `json:"imageId" orm:"image_id"`
	AssetID string `json:"assetId" orm:"asset_id"`
}

type GalleryAppearance struct {
	Sidebar        SidebarAppearance            `json:"sidebar"`
	Classification []ClassificationPresentation `json:"classification"`
}

type SavedFilterQuery struct {
	Search      string `json:"q"`
	Categories  string `json:"categories"`
	Facets      string `json:"facets"`
	Tag         string `json:"tag"`
	MatchAll    bool   `json:"matchAll"`
	Orientation string `json:"orientation"`
	MinLongEdge int    `json:"minLongEdge"`
}

type SavedFilter struct {
	ID    string           `json:"id" orm:"id"`
	Name  string           `json:"name" orm:"name"`
	Query SavedFilterQuery `json:"query" orm:"-"`
}

type SavedFilterInput struct {
	Name  string           `json:"name"`
	Query SavedFilterQuery `json:"query"`
}

type ImageDownload struct {
	URL       string `json:"url"`
	Mime      string `json:"mime"`
	SizeBytes int64  `json:"sizeBytes"`
}
