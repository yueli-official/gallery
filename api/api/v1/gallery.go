package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"platform/gokit/facet"
	"platform/products/gallery/api/internal/model"
)

type GetDiscoveryReq struct {
	g.Meta `path:"/api/v1/gallery/discovery" method:"GET" tags:"Gallery discovery" summary:"Get a stable seeded random image batch"`
	Seed   string `p:"seed"`
}
type GetDiscoveryRes struct {
	Site        model.SiteSettings `json:"site"`
	Seed        string             `json:"seed"`
	Images      []model.ImageCard  `json:"images"`
	Facets      []facet.Facet      `json:"facets"`
	FacetValues []facet.Value      `json:"facetValues"`
}

type ListImagesReq struct {
	g.Meta `path:"/api/v1/gallery/images" method:"GET" tags:"Gallery images" summary:"List eligible public images with page-based filtering"`
	Search string `p:"q"`
	Sort   string `p:"sort" d:"newest"`
	Page   int    `p:"page" d:"1"`
	Size   int    `p:"size" d:"24"`
	Facets string `p:"facets"`
	Tag    string `p:"tag"`
}
type ListImagesRes struct{ model.ImagePage }

type GetImageReq struct {
	g.Meta  `path:"/api/v1/gallery/images/{imageId}" method:"GET" tags:"Gallery images" summary:"Get one eligible public image"`
	ImageID string `p:"imageId" v:"required"`
}
type GetImageRes struct {
	Image model.ImageDetail `json:"image"`
}

type ListCollectionsReq struct {
	g.Meta `path:"/api/v1/gallery/collections" method:"GET" tags:"Gallery collections" summary:"List public editorial collections"`
}
type ListCollectionsRes struct {
	Collections []model.Collection `json:"collections"`
}

type GetCollectionReq struct {
	g.Meta `path:"/api/v1/gallery/collections/{slug}" method:"GET" tags:"Gallery collections" summary:"Get a public editorial collection"`
	Slug   string `p:"slug" v:"required"`
	Page   int    `p:"page" d:"1"`
	Size   int    `p:"size" d:"24"`
}
type GetCollectionRes struct {
	Collection model.CollectionDetail `json:"collection"`
}

type GetRankingsReq struct {
	g.Meta `path:"/api/v1/gallery/rankings" method:"GET" tags:"Gallery rankings" summary:"Get a cached or aggregated public image ranking"`
	Kind   string `p:"kind" d:"trending"`
	Window string `p:"window" d:"7d"`
}
type GetRankingsRes struct {
	Ranking model.Ranking `json:"ranking"`
}

type CreateCaseReq struct {
	g.Meta  `path:"/api/v1/gallery/images/{imageId}/cases" method:"POST" tags:"Gallery cases" summary:"Report an image or suggest a source correction"`
	ImageID string `p:"imageId" v:"required"`
	model.CaseInput
}
type CreateCaseRes struct {
	Case model.Case `json:"case"`
}

type TrackImageEventReq struct {
	g.Meta  `path:"/api/v1/gallery/images/{imageId}/events" method:"POST" tags:"Gallery metrics" summary:"Record a qualified public image interaction"`
	ImageID string `p:"imageId" v:"required"`
	model.EventInput
}
type TrackImageEventRes struct {
	Recorded bool `json:"recorded"`
}
