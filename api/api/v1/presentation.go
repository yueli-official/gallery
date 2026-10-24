package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/yueli-official/gallery/api/internal/model"
)

type ListSavedFiltersReq struct {
	g.Meta `path:"/api/v1/gallery/me/saved-filters" method:"GET" tags:"Gallery filters" summary:"List the current user's saved image filters"`
}
type ListSavedFiltersRes struct {
	Items []model.SavedFilter `json:"items"`
}
type SaveFilterReq struct {
	g.Meta `path:"/api/v1/gallery/me/saved-filters" method:"POST" tags:"Gallery filters" summary:"Save a validated image filter for the current user"`
	model.SavedFilterInput
}
type SaveFilterRes struct {
	g.Meta `status:"201"`
	Filter model.SavedFilter `json:"filter"`
}
type DeleteSavedFilterReq struct {
	g.Meta   `path:"/api/v1/gallery/me/saved-filters/{filterId}" method:"DELETE" tags:"Gallery filters" summary:"Delete the current user's saved filter"`
	FilterID string `p:"filterId" v:"required"`
}
type DeleteSavedFilterRes struct {
	g.Meta `status:"204"`
}
type DownloadImageReq struct {
	g.Meta  `path:"/api/v1/gallery/images/{imageId}/download" method:"GET" tags:"Gallery images" summary:"Create a download URL for an eligible image's original file"`
	ImageID string `p:"imageId" v:"required"`
}
type DownloadImageRes struct{ model.ImageDownload }
