package controller

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
	v1 "github.com/yueli-official/gallery/api/api/v1"
	"github.com/yueli-official/gallery/api/internal/galleryauthz"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
)

type personalPermission struct {
	foundationauth.PersonalPermission
	Capability authorization.CapabilityKey
}

var personalPermissions = []personalPermission{
	{foundationauth.PersonalPermission{Key: "media.upload", Label: "上传投稿图片", Description: "只上传 Gallery 投稿原图；还需选择创建投稿权限。"}, galleryauthz.CapabilitySubmissionCreate},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilitySubmissionCreate), Label: "创建投稿", Description: "以当前账号提交已经上传的图片，仍受分类、反滥用和处理规则限制。"}, galleryauthz.CapabilitySubmissionCreate},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilitySubmissionReadOwn), Label: "读取自己的投稿", Description: "列出当前账号拥有的投稿和处理状态。"}, galleryauthz.CapabilitySubmissionReadOwn},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilitySubmissionWithdraw), Label: "撤回自己的投稿", Description: "撤回当前账号拥有的投稿，并沿用现有状态限制。"}, galleryauthz.CapabilitySubmissionWithdraw},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityFavoriteRead), Label: "读取自己的收藏", Description: "读取当前账号的私有收藏。"}, galleryauthz.CapabilityFavoriteRead},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityFavoriteManage), Label: "管理自己的收藏", Description: "向当前账号收藏中添加或移除图片。"}, galleryauthz.CapabilityFavoriteManage},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityCommentCreate), Label: "发表评论", Description: "以当前账号在公开图片下发表评论或回复。"}, galleryauthz.CapabilityCommentCreate},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityDashboardRead), Label: "查看运营概览", Description: "读取 Gallery 运营概览；仍需当前账号拥有对应权限。"}, galleryauthz.CapabilityDashboardRead},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityImageRead), Label: "查看图片后台", Description: "查询后台图片和详情；仍需当前账号拥有对应权限。"}, galleryauthz.CapabilityImageRead},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityImageUpdate), Label: "编辑图片", Description: "编辑图片元数据与分类；发布等动作仍需额外当前权限。"}, galleryauthz.CapabilityImageUpdate},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityImageHide), Label: "下架或删除图片", Description: "下架、删除或批量隐藏图片；仍需当前账号拥有对应权限。"}, galleryauthz.CapabilityImageHide},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilitySubmissionRead), Label: "查看投稿审核队列", Description: "读取全站投稿与预览；仍需当前账号拥有对应权限。"}, galleryauthz.CapabilitySubmissionRead},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilitySubmissionReview), Label: "审核投稿", Description: "批准或拒绝投稿；仍需当前账号拥有对应权限。"}, galleryauthz.CapabilitySubmissionReview},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityCollectionRead), Label: "查看专题集合", Description: "读取私有和公开专题；仍需当前账号拥有对应权限。"}, galleryauthz.CapabilityCollectionRead},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityCollectionManage), Label: "管理专题集合", Description: "创建、编辑和编排专题；仍需当前账号拥有对应权限。"}, galleryauthz.CapabilityCollectionManage},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityClassificationRead), Label: "查看分类体系", Description: "读取完整分类与标签目录；仍需当前账号拥有对应权限。"}, galleryauthz.CapabilityClassificationRead},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityClassificationProposalReview), Label: "审核标签提案", Description: "处理投稿产生的标签提案；仍需当前账号拥有对应权限。"}, galleryauthz.CapabilityClassificationProposalReview},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityClassificationGovern), Label: "治理分类体系", Description: "创建、修改和执行分类治理计划；仍需当前 Gallery 管理员权限。"}, galleryauthz.CapabilityClassificationGovern},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityCaseRead), Label: "查看处理单", Description: "读取纠错和举报处理单；仍需当前账号拥有对应权限。"}, galleryauthz.CapabilityCaseRead},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityCaseResolve), Label: "处置处理单", Description: "推进、解决或驳回纠错和举报处理单；仍需当前 Gallery 管理员权限。"}, galleryauthz.CapabilityCaseResolve},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityDiscoveryRead), Label: "查看发现策略", Description: "读取后台站点与发现设置；仍需当前账号拥有对应权限。"}, galleryauthz.CapabilityDiscoveryRead},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityDiscoveryManage), Label: "管理发现与站点设置", Description: "修改站点公开信息与发现策略；仍需当前 Gallery 管理员权限。"}, galleryauthz.CapabilityDiscoveryManage},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityCommentRead), Label: "查看评论后台", Description: "读取评论审核队列；仍需当前账号拥有对应权限。"}, galleryauthz.CapabilityCommentRead},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityCommentModerate), Label: "审核评论", Description: "修改评论审核状态；仍需当前账号拥有对应权限。"}, galleryauthz.CapabilityCommentModerate},
	{foundationauth.PersonalPermission{Key: string(galleryauthz.CapabilityCommentDelete), Label: "删除评论", Description: "删除评论及其回复；仍需当前账号拥有对应权限。"}, galleryauthz.CapabilityCommentDelete},
}

type PersonalPermissions struct {
	site    string
	service *galleryauthz.Service
}

func NewPersonalPermissions(site string, service *galleryauthz.Service) *PersonalPermissions {
	return &PersonalPermissions{site: site, service: service}
}

func (controller *PersonalPermissions) GetPersonalPermissions(ctx context.Context, req *v1.PersonalPermissionsReq) (*v1.PersonalPermissionsRes, error) {
	principal, _ := foundationauth.FromContext(ctx)
	if principal == nil || principal.SubjectKind != foundationauth.SubjectClient || principal.ClientID != "identity-svc" || !principal.HasScope(foundationauth.PersonalPermissionsScope) || controller.site == "" {
		return nil, galleryerr.Forbidden()
	}
	if controller.service == nil {
		return nil, galleryerr.AuthorizationUnavailable()
	}
	userCtx := foundationauth.NewContext(ctx, &foundationauth.Principal{Subject: req.UserKey, SubjectKind: foundationauth.SubjectUser})
	items := make([]foundationauth.PersonalPermission, 0, len(personalPermissions))
	for _, permission := range personalPermissions {
		allowed, err := controller.service.CheckCapability(userCtx, permission.Capability)
		if err != nil {
			return nil, mapAuthorizationError(err)
		}
		if allowed {
			items = append(items, permission.PersonalPermission)
		}
	}
	return &v1.PersonalPermissionsRes{Site: controller.site, UserKey: req.UserKey, Items: items}, nil
}

func (controller *PersonalPermissions) AuthorizePersonalMedia(ctx context.Context, _ *v1.PersonalMediaAuthorizationReq) (*v1.PersonalMediaAuthorizationRes, error) {
	principal, _ := foundationauth.FromContext(ctx)
	if principal == nil || !principal.IsPersonalToken() || !principal.HasScope("media.upload") || controller.site == "" || controller.service == nil {
		return nil, galleryerr.Forbidden()
	}
	allowed, err := controller.service.CheckCapability(ctx, galleryauthz.CapabilitySubmissionCreate)
	if err != nil {
		return nil, mapAuthorizationError(err)
	}
	if !allowed {
		return nil, galleryerr.Forbidden()
	}
	scope, err := foundationauth.PersonalScope(controller.site, "asset.profile.gallery-submission.upload")
	if err != nil {
		return nil, galleryerr.Forbidden()
	}
	result := &v1.PersonalMediaAuthorizationRes{UserKey: principal.Subject, Scopes: []string{scope}}
	if !principal.ExpiresAt.IsZero() {
		result.ExpiresAt = principal.ExpiresAt.Format(time.RFC3339)
	}
	return result, nil
}

type personalRoute struct {
	method       string
	path         string
	capabilities []authorization.CapabilityKey
}

var personalRoutes = []personalRoute{
	{"POST", "/api/v1/gallery/submissions", []authorization.CapabilityKey{galleryauthz.CapabilitySubmissionCreate}},
	{"GET", "/api/v1/gallery/me/submissions", []authorization.CapabilityKey{galleryauthz.CapabilitySubmissionReadOwn}},
	{"POST", "/api/v1/gallery/me/submissions/{submissionId}/withdraw", []authorization.CapabilityKey{galleryauthz.CapabilitySubmissionWithdraw}},
	{"GET", "/api/v1/gallery/me/favorites", []authorization.CapabilityKey{galleryauthz.CapabilityFavoriteRead}},
	{"PUT", "/api/v1/gallery/me/favorites/{imageId}", []authorization.CapabilityKey{galleryauthz.CapabilityFavoriteManage}},
	{"DELETE", "/api/v1/gallery/me/favorites/{imageId}", []authorization.CapabilityKey{galleryauthz.CapabilityFavoriteManage}},
	{"POST", "/api/v1/gallery/images/{imageId}/comments", []authorization.CapabilityKey{galleryauthz.CapabilityCommentCreate}},
	{"POST", "/api/v1/personal-token/media-authorization", nil},
	{"GET", "/api/v1/gallery/admin/overview", []authorization.CapabilityKey{galleryauthz.CapabilityDashboardRead}},
	{"GET", "/api/v1/gallery/admin/site-settings", []authorization.CapabilityKey{galleryauthz.CapabilityDiscoveryRead}},
	{"PATCH", "/api/v1/gallery/admin/site-settings", []authorization.CapabilityKey{galleryauthz.CapabilityDiscoveryManage}},
	{"GET", "/api/v1/gallery/admin/classification", []authorization.CapabilityKey{galleryauthz.CapabilityClassificationRead}},
	{"POST", "/api/v1/gallery/admin/classification/identities", []authorization.CapabilityKey{galleryauthz.CapabilityClassificationGovern}},
	{"PATCH", "/api/v1/gallery/admin/classification/identities/{identityId}", []authorization.CapabilityKey{galleryauthz.CapabilityClassificationGovern}},
	{"GET", "/api/v1/gallery/admin/classification/tags", []authorization.CapabilityKey{galleryauthz.CapabilityClassificationRead}},
	{"GET", "/api/v1/gallery/admin/classification/tag-proposals", []authorization.CapabilityKey{galleryauthz.CapabilityClassificationRead}},
	{"POST", "/api/v1/gallery/admin/classification/tag-proposals/{proposalId}/review", []authorization.CapabilityKey{galleryauthz.CapabilityClassificationProposalReview}},
	{"POST", "/api/v1/gallery/admin/classification/governance/preview", []authorization.CapabilityKey{galleryauthz.CapabilityClassificationGovern}},
	{"POST", "/api/v1/gallery/admin/classification/governance/execute", []authorization.CapabilityKey{galleryauthz.CapabilityClassificationGovern}},
	{"GET", "/api/v1/gallery/admin/submissions", []authorization.CapabilityKey{galleryauthz.CapabilitySubmissionRead}},
	{"GET", "/api/v1/gallery/admin/submissions/{submissionId}/preview", []authorization.CapabilityKey{galleryauthz.CapabilitySubmissionRead}},
	{"POST", "/api/v1/gallery/admin/submissions/{submissionId}/review", []authorization.CapabilityKey{galleryauthz.CapabilitySubmissionReview}},
	{"POST", "/api/v1/gallery/admin/submissions/bulk-review", []authorization.CapabilityKey{galleryauthz.CapabilitySubmissionReview}},
	{"GET", "/api/v1/gallery/admin/images", []authorization.CapabilityKey{galleryauthz.CapabilityImageRead}},
	{"GET", "/api/v1/gallery/admin/images/{imageId}", []authorization.CapabilityKey{galleryauthz.CapabilityImageRead}},
	{"PATCH", "/api/v1/gallery/admin/images/{imageId}", []authorization.CapabilityKey{galleryauthz.CapabilityImageUpdate, galleryauthz.CapabilityImageHide}},
	{"DELETE", "/api/v1/gallery/admin/images/{imageId}", []authorization.CapabilityKey{galleryauthz.CapabilityImageHide}},
	{"POST", "/api/v1/gallery/admin/images/{imageId}/hide", []authorization.CapabilityKey{galleryauthz.CapabilityImageHide}},
	{"POST", "/api/v1/gallery/admin/images/bulk-hide", []authorization.CapabilityKey{galleryauthz.CapabilityImageHide}},
	{"POST", "/api/v1/gallery/admin/images/bulk", []authorization.CapabilityKey{galleryauthz.CapabilityImageUpdate, galleryauthz.CapabilityImageHide}},
	{"GET", "/api/v1/gallery/admin/cases", []authorization.CapabilityKey{galleryauthz.CapabilityCaseRead}},
	{"POST", "/api/v1/gallery/admin/cases/{caseId}/resolve", []authorization.CapabilityKey{galleryauthz.CapabilityCaseResolve}},
	{"GET", "/api/v1/gallery/admin/collections", []authorization.CapabilityKey{galleryauthz.CapabilityCollectionRead}},
	{"GET", "/api/v1/gallery/admin/collections/{collectionId}", []authorization.CapabilityKey{galleryauthz.CapabilityCollectionRead}},
	{"POST", "/api/v1/gallery/admin/collections", []authorization.CapabilityKey{galleryauthz.CapabilityCollectionManage}},
	{"PATCH", "/api/v1/gallery/admin/collections/{collectionId}", []authorization.CapabilityKey{galleryauthz.CapabilityCollectionManage}},
	{"POST", "/api/v1/gallery/admin/collections/{collectionId}/members", []authorization.CapabilityKey{galleryauthz.CapabilityCollectionManage}},
	{"PUT", "/api/v1/gallery/admin/collections/{collectionId}/order", []authorization.CapabilityKey{galleryauthz.CapabilityCollectionManage}},
	{"GET", "/api/v1/gallery/admin/comments", []authorization.CapabilityKey{galleryauthz.CapabilityCommentRead}},
	{"PATCH", "/api/v1/gallery/admin/comments/{commentId}", []authorization.CapabilityKey{galleryauthz.CapabilityCommentModerate}},
	{"DELETE", "/api/v1/gallery/admin/comments/{commentId}", []authorization.CapabilityKey{galleryauthz.CapabilityCommentDelete}},
}

func PersonalTokenRoutes(request *ghttp.Request) {
	principal, _ := foundationauth.FromContext(request.Context())
	if principal != nil && principal.IsPersonalToken() && !allowsPersonalRoute(request.Context(), request.Method, request.URL.Path) {
		request.SetError(galleryerr.Forbidden())
		return
	}
	request.Middleware.Next()
}

func allowsPersonalRoute(ctx context.Context, method, path string) bool {
	if method == "GET" && isPublicGalleryRead(path) {
		return true
	}
	for _, route := range personalRoutes {
		if route.method != method || !matchesPersonalPath(route.path, path) {
			continue
		}
		if len(route.capabilities) == 0 {
			return foundationauth.AllowsPersonalCapability(ctx, "media.upload")
		}
		for _, capability := range route.capabilities {
			if foundationauth.AllowsPersonalCapability(ctx, string(capability)) {
				return true
			}
		}
	}
	return false
}

func matchesPersonalPath(pattern, path string) bool {
	want, got := strings.Split(pattern, "/"), strings.Split(path, "/")
	if len(want) != len(got) {
		return false
	}
	for index, part := range want {
		if strings.HasPrefix(part, "{") {
			if got[index] == "" || got[index] == "." || got[index] == ".." {
				return false
			}
			continue
		}
		if part != got[index] {
			return false
		}
	}
	return true
}

func isPublicGalleryRead(path string) bool {
	for _, pattern := range []string{
		"/api/v1/gallery/discovery", "/api/v1/gallery/site", "/api/v1/gallery/submission-options",
		"/api/v1/gallery/images", "/api/v1/gallery/images/{imageId}", "/api/v1/gallery/images/{imageId}/related",
		"/api/v1/gallery/images/{imageId}/comments", "/api/v1/gallery/collections",
		"/api/v1/gallery/collections/{slug}", "/api/v1/gallery/rankings",
	} {
		if matchesPersonalPath(pattern, path) {
			return true
		}
	}
	return false
}
