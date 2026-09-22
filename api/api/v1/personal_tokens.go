package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	foundationauth "github.com/yueli-official/foundation/go/auth"
)

type PersonalPermissionsReq struct {
	g.Meta  `path:"/api/v1/internal/personal-token/permissions" method:"get" tags:"authorization" summary:"Read current user capabilities for the Identity token catalog"`
	UserKey string `json:"userKey" in:"query" v:"required|length:1,200"`
}

type PersonalPermissionsRes = foundationauth.PersonalPermissions

type PersonalMediaAuthorizationReq struct {
	g.Meta `path:"/api/v1/personal-token/media-authorization" method:"post" tags:"authorization" summary:"Authorize constrained Gallery submission upload for the current personal token"`
}

type PersonalMediaAuthorizationRes struct {
	UserKey   string   `json:"userKey"`
	Scopes    []string `json:"scopes"`
	ExpiresAt string   `json:"expiresAt,omitempty"`
}
