package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"platform/products/gallery/api/internal/model"
)

type GetDiscoveryReq struct {
	g.Meta `path:"/api/v1/gallery/discovery" method:"GET" tags:"Gallery" summary:"Get curated and latest gallery discovery content"`
}

type GetDiscoveryRes struct {
	Site       model.SiteSettings  `json:"site"`
	Featured   []model.ArtworkCard `json:"featured"`
	Latest     []model.ArtworkCard `json:"latest"`
	Facets     []model.Facet       `json:"facets"`
	Categories []model.Category    `json:"categories"`
}

type GetArtworkReq struct {
	g.Meta    `path:"/api/v1/gallery/artworks/{artworkId}" method:"GET" tags:"Gallery" summary:"Get one published gallery artwork"`
	ArtworkID string `p:"artworkId" v:"required"`
}

type GetArtworkRes struct {
	Artwork model.ArtworkDetail `json:"artwork"`
}
