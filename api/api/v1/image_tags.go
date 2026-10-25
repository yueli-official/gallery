package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/yueli-official/gallery/api/internal/model"
)

type TagSuggestionsReq struct {
	g.Meta `path:"/api/v1/gallery/tag-suggestions" method:"GET" tags:"Gallery tags" summary:"Suggest active canonical tags, including unused preset tags"`
	Query  string `p:"q" v:"length:0,64"`
}
type TagSuggestionsRes struct {
	Items []model.TagCandidate `json:"items"`
}

type ProposeImageTagsReq struct {
	g.Meta  `path:"/api/v1/gallery/images/{imageId}/tag-proposals" method:"POST" tags:"Gallery tags" summary:"Suggest tags for a public image; changes require moderator approval"`
	ImageID string   `p:"imageId" v:"required"`
	Tags    []string `json:"tags" v:"required"`
}
type ProposeImageTagsRes struct {
	Items []model.ClassificationTagProposal `json:"items"`
}
