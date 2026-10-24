package controller

import (
	"context"
	v1 "github.com/yueli-official/gallery/api/api/v1"
)

func (c *Workflow) ListSavedFilters(ctx context.Context, _ *v1.ListSavedFiltersReq) (*v1.ListSavedFiltersRes, error) {
	subject, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.service.SavedFilters(ctx, subject.ID)
	if err != nil {
		return nil, err
	}
	return &v1.ListSavedFiltersRes{Items: items}, nil
}
func (c *Workflow) SaveFilter(ctx context.Context, req *v1.SaveFilterReq) (*v1.SaveFilterRes, error) {
	subject, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.service.SaveFilter(ctx, subject.ID, req.SavedFilterInput)
	if err != nil {
		return nil, err
	}
	writeSuccess(ctx, 201, "")
	return &v1.SaveFilterRes{Filter: *item}, nil
}
func (c *Workflow) DeleteSavedFilter(ctx context.Context, req *v1.DeleteSavedFilterReq) (*v1.DeleteSavedFilterRes, error) {
	subject, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.service.DeleteSavedFilter(ctx, subject.ID, req.FilterID); err != nil {
		return nil, err
	}
	writeSuccess(ctx, 204, "")
	return &v1.DeleteSavedFilterRes{}, nil
}
func (c *Public) DownloadImage(ctx context.Context, req *v1.DownloadImageReq) (*v1.DownloadImageRes, error) {
	value, err := c.service.DownloadImage(ctx, req.ImageID)
	if err != nil {
		return nil, err
	}
	return &v1.DownloadImageRes{ImageDownload: *value}, nil
}
