package controller

import (
	"context"
	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
	v1 "github.com/yueli-official/gallery/api/api/v1"
	"github.com/yueli-official/gallery/api/internal/galleryauthz"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
)

type AuthorizationSetup struct{ service *galleryauthz.Service }

func NewAuthorizationSetup(service *galleryauthz.Service) *AuthorizationSetup {
	return &AuthorizationSetup{service: service}
}
func (controller *AuthorizationSetup) GetInitialAdministratorClaimStatus(ctx context.Context, _ *v1.GetInitialAdministratorClaimStatusReq) (*v1.GetInitialAdministratorClaimStatusRes, error) {
	if controller.service == nil {
		return nil, galleryerr.AuthorizationUnavailable()
	}
	status, err := controller.service.Runtime().AdministratorClaimStatus(ctx)
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.GetInitialAdministratorClaimStatusRes{Claimed: status.Claimed, CanClaim: !status.Claimed}, nil
}
func (*Authorization) ClaimInitialAdministrator(ctx context.Context, _ *v1.ClaimInitialAdministratorReq) (*v1.ClaimInitialAdministratorRes, error) {
	if _, err := requiredUser(ctx); err != nil {
		return nil, err
	}
	principal, _ := foundationauth.FromContext(ctx)
	if principal.IsPersonalToken() {
		return nil, galleryerr.Forbidden()
	}
	service := authorizationService(ctx)
	if service == nil {
		return nil, galleryerr.AuthorizationUnavailable()
	}
	result, err := service.Runtime().ClaimInitialAdministrator(ctx, authorization.ClaimInitialAdministratorCommand{Actor: service.Subject(ctx)})
	if authorization.Is(err, authorization.ErrorConflict) {
		return nil, galleryerr.Conflict("administrator_claim")
	}
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	return &v1.ClaimInitialAdministratorRes{Claimed: result.Status.Claimed, Created: result.Created}, nil
}
