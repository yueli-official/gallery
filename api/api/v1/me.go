package v1

import "github.com/gogf/gf/v2/frame/g"

type GetGalleryMeReq struct {
	g.Meta `path:"/api/v1/me" method:"get" tags:"gallery" summary:"Get my Gallery access"`
}

type GetGalleryMeRes struct {
	IsAdministrator bool     `json:"isAdministrator"`
	Roles           []string `json:"roles"`
	Capabilities    []string `json:"capabilities"`
}
