package controller

import (
	"context"

	v1 "platform/products/gallery/api/api/v1"
	"platform/products/gallery/api/internal/galleryerr"
)

type Me struct{}

func NewMe() *Me { return &Me{} }

func (*Me) GetGalleryMe(ctx context.Context, _ *v1.GetGalleryMeReq) (*v1.GetGalleryMeRes, error) {
	service := authorizationService(ctx)
	if service == nil {
		return nil, galleryerr.AuthorizationUnavailable()
	}
	access, err := service.EffectiveAccess(ctx)
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	roles := make([]string, len(access.Grants))
	for index, grant := range access.Grants {
		roles[index] = string(grant.Role)
	}
	capabilities := make([]string, len(access.Capabilities))
	for index, capability := range access.Capabilities {
		capabilities[index] = string(capability)
	}
	return &v1.GetGalleryMeRes{
		IsAdministrator: isAdmin(ctx), Roles: roles, Capabilities: capabilities,
	}, nil
}
