package controller

import (
	"context"
	v1 "github.com/yueli-official/gallery/api/api/v1"
)

func (c *Public) TagSuggestions(ctx context.Context, req *v1.TagSuggestionsReq) (*v1.TagSuggestionsRes, error) {
	items, err := c.service.TagSuggestions(ctx, req.Query)
	return &v1.TagSuggestionsRes{Items: items}, err
}

func (c *Workflow) ProposeImageTags(ctx context.Context, req *v1.ProposeImageTagsReq) (*v1.ProposeImageTagsRes, error) {
	subject, err := requiredSubject(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.service.ProposeImageTags(ctx, subject, req.ImageID, req.Tags)
	if err != nil {
		return nil, err
	}
	return &v1.ProposeImageTagsRes{Items: items}, err
}
