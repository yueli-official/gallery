package controller

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/google/uuid"
	"github.com/yueli-official/foundation/go/abuse"
	foundationauth "github.com/yueli-official/foundation/go/auth"
	v1 "platform/products/gallery/api/api/v1"
	galleryservice "platform/products/gallery/api/internal/gallery"
	"platform/products/gallery/api/internal/galleryerr"
	"platform/products/gallery/api/internal/model"
)

type Workflow struct{ service *galleryservice.Service }

func NewWorkflow(service *galleryservice.Service) *Workflow { return &Workflow{service: service} }

func (c *Workflow) CreateSubmission(ctx context.Context, req *v1.CreateSubmissionReq) (*v1.CreateSubmissionRes, error) {
	subject, err := requiredSubject(ctx)
	if err != nil {
		return nil, err
	}
	attemptID := strings.TrimSpace(req.AbuseAttemptID)
	if attemptID == "" {
		attemptID = uuid.NewString()
	}
	request := ghttp.RequestFromCtx(ctx)
	if request == nil {
		return nil, galleryerr.AbuseUnavailable()
	}
	admission, err := c.service.AdmitSubmission(
		ctx, subject, request.GetClientIp(), attemptID, req.ChallengeProof,
	)
	if err != nil {
		if abuse.IsKind(err, abuse.ErrorConflict) {
			return nil, galleryerr.AbuseAttemptReplayed()
		}
		return nil, galleryerr.AbuseUnavailable()
	}
	switch admission.Disposition {
	case abuse.DispositionAllow:
		if admission.Replay {
			return nil, galleryerr.AbuseAttemptReplayed()
		}
	case abuse.DispositionChallenge:
		return nil, galleryerr.ChallengeRequired(attemptID)
	default:
		return nil, galleryerr.RateLimited()
	}
	value, err := c.service.Submit(ctx, subject, req.SubmissionInput)
	if err != nil {
		return nil, err
	}
	return &v1.CreateSubmissionRes{Submission: *value}, nil
}

func (c *Workflow) ClaimGuestSubmissions(ctx context.Context, _ *v1.ClaimGuestSubmissionsReq) (*v1.ClaimGuestSubmissionsRes, error) {
	principal, ok := foundationauth.FromContext(ctx)
	if !ok || principal == nil || !principal.HasScope("guest:claim") {
		return nil, galleryerr.Forbidden()
	}
	guestClaim, _ := principal.Claim("guest_subject")
	guestSubject := valueString(guestClaim)
	subjectKind, _ := principal.Claim("subject_kind")
	if guestSubject == "" || valueString(subjectKind) != "user" {
		return nil, galleryerr.Forbidden()
	}
	claimed, err := c.service.ClaimGuestSubmissions(ctx, guestSubject, principal.Subject)
	if err != nil {
		return nil, err
	}
	return &v1.ClaimGuestSubmissionsRes{Claimed: claimed}, nil
}

func (c *Workflow) ListMySubmissions(ctx context.Context, req *v1.ListMySubmissionsReq) (*v1.ListMySubmissionsRes, error) {
	subject, err := requiredSubject(ctx)
	if err != nil {
		return nil, err
	}
	query := galleryservice.MySubmissionQuery(req.Page, req.Size, req.Outcome, req.ProcessingState, req.ReviewState)
	values, total, err := c.service.MySubmissions(ctx, subject, query)
	pages := 0
	if total > 0 {
		pages = (total + query.PageSize - 1) / query.PageSize
	}
	return &v1.ListMySubmissionsRes{
		Submissions: values, Total: total, Page: query.Page, PageSize: query.PageSize, TotalPages: pages,
	}, err
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

func (c *Workflow) GetFavorites(ctx context.Context, req *v1.GetFavoritesReq) (*v1.GetFavoritesRes, error) {
	subject, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	value, err := c.service.Favorites(ctx, subject.ID, req.Page, req.Size, req.Sort)
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

func (c *Admin) GetClassificationCatalog(ctx context.Context, _ *v1.GetClassificationCatalogReq) (*v1.GetClassificationCatalogRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	catalog, err := c.service.ClassificationCatalog(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetClassificationCatalogRes{Catalog: *catalog}, nil
}

func (c *Admin) ListClassificationTags(ctx context.Context, req *v1.ListClassificationTagsReq) (*v1.ListClassificationTagsRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	page, err := c.service.ClassificationTags(ctx, req.Cursor, req.Size)
	if err != nil {
		return nil, err
	}
	return &v1.ListClassificationTagsRes{Page: *page}, nil
}

func (c *Admin) ListClassificationTagProposals(ctx context.Context, req *v1.ListClassificationTagProposalsReq) (*v1.ListClassificationTagProposalsRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	proposals, total, err := c.service.ClassificationTagProposals(ctx, req.Status, req.Page, req.Size)
	return &v1.ListClassificationTagProposalsRes{Proposals: proposals, Total: total}, err
}

func (c *Admin) ReviewClassificationTagProposal(ctx context.Context, req *v1.ReviewClassificationTagProposalReq) (*v1.ReviewClassificationTagProposalRes, error) {
	operator, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	proposal, err := c.service.ReviewClassificationTagProposal(ctx, operator, req.ProposalID, req.ClassificationTagProposalReviewInput)
	if err != nil {
		return nil, err
	}
	return &v1.ReviewClassificationTagProposalRes{Proposal: *proposal}, nil
}

func (c *Admin) PreviewClassificationGovernance(ctx context.Context, req *v1.PreviewClassificationGovernanceReq) (*v1.PreviewClassificationGovernanceRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	preview, err := c.service.PreviewClassificationGovernance(ctx, req.ClassificationGovernancePreviewInput)
	if err != nil {
		return nil, err
	}
	return &v1.PreviewClassificationGovernanceRes{Preview: *preview}, nil
}

func (c *Admin) ExecuteClassificationGovernance(ctx context.Context, req *v1.ExecuteClassificationGovernanceReq) (*v1.ExecuteClassificationGovernanceRes, error) {
	operator, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	execution, err := c.service.ExecuteClassificationGovernance(ctx, operator, req.ClassificationGovernanceExecuteInput)
	if err != nil {
		return nil, err
	}
	return &v1.ExecuteClassificationGovernanceRes{Execution: *execution}, nil
}

func (c *Admin) ListSubmissionReviews(ctx context.Context, req *v1.ListSubmissionReviewsReq) (*v1.ListSubmissionReviewsRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	page, err := c.service.ReviewQueue(ctx, model.AdminSubmissionQuery{
		Search: req.Q, Sort: req.Sort, Page: req.Page, PageSize: req.Size,
		ProcessingState: req.ProcessingState, ReviewState: req.ReviewState,
		SafetyState: req.SafetyState, Outcome: req.Outcome,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListSubmissionReviewsRes{AdminSubmissionPage: *page}, nil
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

func (c *Admin) BulkReviewSubmissions(ctx context.Context, req *v1.BulkReviewSubmissionsReq) (*v1.BulkReviewSubmissionsRes, error) {
	operator, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	results, err := c.service.BulkReviewSubmissions(ctx, operator, req.BulkSubmissionReviewInput)
	if err != nil {
		return nil, err
	}
	return &v1.BulkReviewSubmissionsRes{Results: results}, nil
}

func (c *Admin) PreviewSubmission(ctx context.Context, req *v1.PreviewSubmissionReq) (*v1.PreviewSubmissionRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	previewURL, err := c.service.SubmissionPreviewURL(ctx, req.SubmissionID)
	if err != nil {
		return nil, err
	}
	return &v1.PreviewSubmissionRes{URL: previewURL}, nil
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

func (c *Admin) ListAdminImages(ctx context.Context, req *v1.ListAdminImagesReq) (*v1.ListAdminImagesRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	page, err := c.service.AdminImages(ctx, model.AdminImageQuery{
		Search: req.Q, Sort: req.Sort, Page: req.Page, PageSize: req.Size,
		ProcessingState: req.ProcessingState, ReviewState: req.ReviewState,
		PublicationState: req.PublicationState, SafetyState: req.SafetyState,
		CategoryID: req.CategoryID, FacetValueID: req.FacetValueID,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListAdminImagesRes{AdminImagePage: *page}, nil
}

func (c *Admin) GetAdminImage(ctx context.Context, req *v1.GetAdminImageReq) (*v1.GetAdminImageRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	value, err := c.service.AdminImage(ctx, req.ImageID)
	if err != nil {
		return nil, err
	}
	return &v1.GetAdminImageRes{Image: *value}, nil
}

func (c *Admin) UpdateAdminImage(ctx context.Context, req *v1.UpdateAdminImageReq) (*v1.UpdateAdminImageRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	value, err := c.service.UpdateAdminImage(ctx, req.ImageID, req.AdminImageUpdateInput)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAdminImageRes{Image: *value}, nil
}

func (c *Admin) BulkHideImages(ctx context.Context, req *v1.BulkHideImagesReq) (*v1.BulkHideImagesRes, error) {
	operator, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.BulkHideImagesRes{Results: c.service.BulkHideImages(ctx, operator, req.BulkImageHideInput)}, nil
}

func (c *Admin) BulkImages(ctx context.Context, req *v1.BulkImagesReq) (*v1.BulkImagesRes, error) {
	operator, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.BulkImagesRes{Results: c.service.BulkImages(ctx, operator, req.BulkImageActionInput)}, nil
}

func (c *Admin) ListCases(ctx context.Context, req *v1.ListCasesReq) (*v1.ListCasesRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	page, err := c.service.Cases(ctx, model.AdminCaseQuery{
		Search: req.Q, Sort: req.Sort, Status: req.Status, Kind: req.Kind, Page: req.Page, PageSize: req.Size,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListCasesRes{AdminCasePage: *page}, nil
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

func (c *Admin) GetEditorialCollection(ctx context.Context, req *v1.GetEditorialCollectionReq) (*v1.GetEditorialCollectionRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	value, err := c.service.AdminCollection(ctx, req.CollectionID, req.Page, req.Size)
	if err != nil {
		return nil, err
	}
	return &v1.GetEditorialCollectionRes{Collection: *value}, nil
}

func (c *Admin) UpdateEditorialCollection(ctx context.Context, req *v1.UpdateEditorialCollectionReq) (*v1.UpdateEditorialCollectionRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	value, err := c.service.UpdateEditorialCollection(ctx, req.CollectionID, req.EditorialCollectionUpdateInput)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateEditorialCollectionRes{Collection: *value}, nil
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

func (c *Admin) ReorderEditorialMembers(ctx context.Context, req *v1.ReorderEditorialMembersReq) (*v1.ReorderEditorialMembersRes, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	value, err := c.service.ReorderEditorialMembers(ctx, req.CollectionID, req.EditorialCollectionOrderInput)
	if err != nil {
		return nil, err
	}
	return &v1.ReorderEditorialMembersRes{Collection: *value}, nil
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
