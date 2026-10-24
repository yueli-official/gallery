package v1

import "github.com/gogf/gf/v2/frame/g"

type GetInitialAdministratorClaimStatusReq struct {
	g.Meta `path:"/api/v1/authorization/setup" method:"GET" tags:"Authorization" summary:"Read initial administrator claim status"`
}
type GetInitialAdministratorClaimStatusRes struct {
	Claimed  bool `json:"claimed"`
	CanClaim bool `json:"canClaim"`
}
type ClaimInitialAdministratorReq struct {
	g.Meta `path:"/api/v1/authorization/setup/claim" method:"POST" tags:"Authorization" summary:"Explicitly claim the initial site administrator"`
}
type ClaimInitialAdministratorRes struct {
	Claimed bool `json:"claimed"`
	Created bool `json:"created"`
}
