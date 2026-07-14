package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"platform/products/gallery/api/internal/model"
)

type GetMyCreatorReq struct {
	g.Meta `path:"/api/v1/gallery/me/creator" method:"GET" tags:"Gallery creator" summary:"Get the current account creator profile"`
}
type GetMyCreatorRes struct {
	Creator *model.CreatorProfile `json:"creator"`
}

type RequestCreatorReq struct {
	g.Meta          `path:"/api/v1/gallery/me/creator" method:"POST" tags:"Gallery creator" summary:"Apply for a gallery creator profile"`
	Handle          string `json:"handle" v:"required"`
	DisplayName     string `json:"displayName" v:"required"`
	ApplicationNote string `json:"applicationNote"`
}
type RequestCreatorRes struct {
	Creator model.CreatorProfile `json:"creator"`
}

type ListMyArtworksReq struct {
	g.Meta `path:"/api/v1/gallery/me/artworks" method:"GET" tags:"Gallery studio" summary:"List the current creator artworks"`
}
type ListMyArtworksRes struct {
	Artworks []model.StudioArtwork `json:"artworks"`
}

type CreateArtworkReq struct {
	g.Meta `path:"/api/v1/gallery/me/artworks" method:"POST" tags:"Gallery studio" summary:"Create an empty artwork draft"`
}
type CreateArtworkRes struct {
	Artwork model.StudioArtwork `json:"artwork"`
}

type GetMyArtworkReq struct {
	g.Meta    `path:"/api/v1/gallery/me/artworks/{artworkId}" method:"GET" tags:"Gallery studio" summary:"Get one creator-owned artwork"`
	ArtworkID string `p:"artworkId" v:"required"`
}
type GetMyArtworkRes struct {
	Artwork model.StudioArtwork `json:"artwork"`
}

type SaveArtworkReq struct {
	g.Meta    `path:"/api/v1/gallery/me/artworks/{artworkId}" method:"PATCH" tags:"Gallery studio" summary:"Save an artwork draft and its facet selections"`
	ArtworkID string `p:"artworkId" v:"required"`
	model.ArtworkDraftInput
}
type SaveArtworkRes struct {
	Artwork model.StudioArtwork `json:"artwork"`
}

type AddArtworkAssetReq struct {
	g.Meta    `path:"/api/v1/gallery/me/artworks/{artworkId}/assets" method:"POST" tags:"Gallery studio" summary:"Link an uploaded asset to an artwork"`
	ArtworkID string `p:"artworkId" v:"required"`
	model.AssetInput
}
type AddArtworkAssetRes struct {
	Asset model.Asset `json:"asset"`
}

type RemoveArtworkAssetReq struct {
	g.Meta         `path:"/api/v1/gallery/me/artworks/{artworkId}/assets/{artworkAssetId}" method:"DELETE" tags:"Gallery studio" summary:"Remove an artwork image link"`
	ArtworkID      string `p:"artworkId" v:"required"`
	ArtworkAssetID string `p:"artworkAssetId" v:"required"`
}
type RemoveArtworkAssetRes struct {
	Removed bool `json:"removed"`
}

type SubmitArtworkReq struct {
	g.Meta    `path:"/api/v1/gallery/me/artworks/{artworkId}/submit" method:"POST" tags:"Gallery studio" summary:"Submit a complete artwork for review"`
	ArtworkID string `p:"artworkId" v:"required"`
}
type SubmitArtworkRes struct {
	Artwork model.StudioArtwork `json:"artwork"`
}

type ListCreatorApplicationsReq struct {
	g.Meta `path:"/api/v1/gallery/admin/creators" method:"GET" tags:"Gallery admin" summary:"List creator applications"`
}
type ListCreatorApplicationsRes struct {
	Creators []model.CreatorProfile `json:"creators"`
}

type ReviewCreatorReq struct {
	g.Meta    `path:"/api/v1/gallery/admin/creators/{creatorId}/review" method:"POST" tags:"Gallery admin" summary:"Approve or reject a creator application"`
	CreatorID string `p:"creatorId" v:"required"`
	Decision  string `json:"decision" v:"required|in:approve,reject"`
	Note      string `json:"note"`
}
type ReviewCreatorRes struct {
	Creator model.CreatorProfile `json:"creator"`
}

type ListArtworkReviewsReq struct {
	g.Meta `path:"/api/v1/gallery/admin/reviews" method:"GET" tags:"Gallery admin" summary:"List artworks in a review state"`
	Status string `p:"status" d:"pending_review"`
}
type ListArtworkReviewsRes struct {
	Artworks []model.StudioArtwork `json:"artworks"`
}

type ReviewArtworkReq struct {
	g.Meta    `path:"/api/v1/gallery/admin/reviews/{artworkId}" method:"POST" tags:"Gallery admin" summary:"Approve or reject a submitted artwork"`
	ArtworkID string `p:"artworkId" v:"required"`
	Decision  string `json:"decision" v:"required|in:approve,reject"`
	Note      string `json:"note"`
}
type ReviewArtworkRes struct {
	Artwork model.StudioArtwork `json:"artwork"`
}
