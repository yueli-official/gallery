package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"platform/gokit/facet"
	"platform/products/gallery/api/internal/model"
)

type GetDiscoveryReq struct {
	g.Meta `path:"/api/v1/gallery/discovery" method:"GET" tags:"Gallery" summary:"Get curated and latest gallery discovery content"`
}

type GetDiscoveryRes struct {
	Site        model.SiteSettings  `json:"site"`
	Featured    []model.ArtworkCard `json:"featured"`
	Latest      []model.ArtworkCard `json:"latest"`
	Facets      []facet.Facet       `json:"facets"`
	FacetValues []facet.Value       `json:"facetValues"`
}

type GetArtworkReq struct {
	g.Meta    `path:"/api/v1/gallery/artworks/{artworkId}" method:"GET" tags:"Gallery" summary:"Get one published gallery artwork"`
	ArtworkID string `p:"artworkId" v:"required"`
}

type GetArtworkRes struct {
	Artwork model.ArtworkDetail `json:"artwork"`
}

type GetCreatorReq struct {
	g.Meta `path:"/api/v1/gallery/creators/{handle}" method:"GET" tags:"Gallery" summary:"Get one public creator and published artworks"`
	Handle string `p:"handle" v:"required"`
}

type GetCreatorRes struct {
	Creator  model.PublicCreator `json:"creator"`
	Artworks []model.ArtworkCard `json:"artworks"`
}
