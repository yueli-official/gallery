package controller

import (
	"context"
	"strings"

	v1 "platform/products/gallery/api/api/v1"
	galleryservice "platform/products/gallery/api/internal/gallery"
)

type Workflow struct{ service *galleryservice.Service }

func NewWorkflow(service *galleryservice.Service) *Workflow { return &Workflow{service: service} }

func (c *Workflow) CreateSubmission(ctx context.Context, req *v1.CreateSubmissionReq) (*v1.CreateSubmissionRes, error) {
	subject, err := requiredSubject(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.Submit(ctx, subject, req.SubmissionInput)
	if err != nil {
		return nil, err
	}
	return &v1.CreateSubmissionRes{Submission: *value}, nil
}

func (c *Workflow) ListMySubmissions(ctx context.Context, req *v1.ListMySubmissionsReq) (*v1.ListMySubmissionsRes, error) {
	subject, err := requiredSubject(ctx)
	if err != nil {
		return nil, err
	}
	values, total, err := c.service.MySubmissions(ctx, subject, req.Page, req.Size)
	return &v1.ListMySubmissionsRes{Submissions: values, Total: total}, err
}

func (c *Workflow) WithdrawSubmission(ctx context.Context, req *v1.WithdrawSubmissionReq) (*v1.WithdrawSubmissionRes, error) {
	subject, err := requiredSubject(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.Withdraw(ctx, subject, req.SubmissionID)
	if err != nil {
		return nil, err
	}
	return &v1.WithdrawSubmissionRes{Submission: *value}, nil
}

func (c *Workflow) GetFavorites(ctx context.Context, _ *v1.GetFavoritesReq) (*v1.GetFavoritesRes, error) {
	subject, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.Favorites(ctx, subject.ID)
	if err != nil {
		return nil, err
	}
	return &v1.GetFavoritesRes{Collection: *value}, nil
}

func (c *Workflow) AddFavorite(ctx context.Context, req *v1.AddFavoriteReq) (*v1.AddFavoriteRes, error) {
	subject, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.SetFavorite(ctx, subject.ID, req.ImageID, true, req.Version)
	if err != nil {
		return nil, err
	}
	return &v1.AddFavoriteRes{Collection: *value}, nil
}

func (c *Workflow) RemoveFavorite(ctx context.Context, req *v1.RemoveFavoriteReq) (*v1.RemoveFavoriteRes, error) {
	subject, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.SetFavorite(ctx, subject.ID, req.ImageID, false, req.Version)
	if err != nil {
		return nil, err
	}
	return &v1.RemoveFavoriteRes{Collection: *value}, nil
}

type Admin struct{ service *galleryservice.Service }

func NewAdmin(service *galleryservice.Service) *Admin { return &Admin{service: service} }

func (c *Admin) GetAdminOverview(ctx context.Context, _ *v1.GetAdminOverviewReq) (*v1.GetAdminOverviewRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	value, err := c.service.AdminOverview(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetAdminOverviewRes{Overview: *value}, nil
}

func (c *Admin) ListSubmissionReviews(ctx context.Context, req *v1.ListSubmissionReviewsReq) (*v1.ListSubmissionReviewsRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	values, total, err := c.service.ReviewQueue(ctx, req.Page, req.Size)
	return &v1.ListSubmissionReviewsRes{Submissions: values, Total: total}, err
}

func (c *Admin) ReviewSubmission(ctx context.Context, req *v1.ReviewSubmissionReq) (*v1.ReviewSubmissionRes, error) {
	operator, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.ReviewSubmission(ctx, operator, req.SubmissionID, req.SubmissionReviewInput)
	if err != nil {
		return nil, err
	}
	return &v1.ReviewSubmissionRes{Submission: *value}, nil
}

func (c *Admin) HideImage(ctx context.Context, req *v1.HideImageReq) (*v1.HideImageRes, error) {
	operator, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.service.HideImage(ctx, operator, req.ImageID, req.Reason); err != nil {
		return nil, err
	}
	return &v1.HideImageRes{Hidden: true}, nil
}

func (c *Admin) ListCases(ctx context.Context, req *v1.ListCasesReq) (*v1.ListCasesRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	values, total, err := c.service.Cases(ctx, req.Status, req.Page, req.Size)
	return &v1.ListCasesRes{Cases: values, Total: total}, err
}

func (c *Admin) ResolveCase(ctx context.Context, req *v1.ResolveCaseReq) (*v1.ResolveCaseRes, error) {
	operator, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.ResolveCase(ctx, operator, req.CaseID, req.CaseResolutionInput)
	if err != nil {
		return nil, err
	}
	return &v1.ResolveCaseRes{Case: *value}, nil
}

func (c *Admin) CreateEditorialCollection(ctx context.Context, req *v1.CreateEditorialCollectionReq) (*v1.CreateEditorialCollectionRes, error) {
	operator, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.CreateEditorialCollection(ctx, operator, req.EditorialCollectionInput)
	if err != nil {
		return nil, err
	}
	return &v1.CreateEditorialCollectionRes{Collection: *value}, nil
}

func (c *Admin) ListEditorialCollections(ctx context.Context, _ *v1.ListEditorialCollectionsReq) (*v1.ListEditorialCollectionsRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	values, err := c.service.AdminCollections(ctx)
	return &v1.ListEditorialCollectionsRes{Collections: values}, err
}

func (c *Admin) MutateEditorialMembers(ctx context.Context, req *v1.MutateEditorialMembersReq) (*v1.MutateEditorialMembersRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	value, err := c.service.MutateEditorialMembers(ctx, req.CollectionID, req.MemberMutationInput)
	if err != nil {
		return nil, err
	}
	return &v1.MutateEditorialMembersRes{Collection: *value}, nil
}

func claimBool(value any) bool {
	if parsed, ok := value.(bool); ok {
		return parsed
	}
	return strings.EqualFold(strings.TrimSpace(valueString(value)), "true")
}

func valueString(value any) string {
	if value == nil {
		return ""
	}
	if parsed, ok := value.(string); ok {
		return parsed
	}
	return ""
}
