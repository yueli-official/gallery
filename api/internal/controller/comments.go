package controller

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/yueli-official/foundation/go/identifier"
	v1 "github.com/yueli-official/gallery/api/api/v1"
	"github.com/yueli-official/gallery/api/internal/galleryauthz"
	"github.com/yueli-official/gallery/api/internal/gallerycomments"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
)

type PublicComments struct{ module *gallerycomments.Module }

func NewPublicComments(module *gallerycomments.Module) *PublicComments {
	return &PublicComments{module: module}
}

func (controller *PublicComments) ListImageComments(ctx context.Context, req *v1.ListImageCommentsReq) (*v1.ListImageCommentsRes, error) {
	order := strings.ToLower(strings.TrimSpace(req.SortOrder))
	if order == "" {
		order = "asc"
	}
	if order != "asc" && order != "desc" {
		return nil, galleryerr.Validation("sortOrder", "unsupported public comment sort order")
	}
	page, err := controller.module.ListPublic(ctx, req.ImageID, order == "asc", req.Page, req.Size)
	if err != nil {
		return nil, err
	}
	return &v1.ListImageCommentsRes{Items: publicCommentViews(page.Items), Total: page.Total, Page: page.Page, Size: page.Size}, nil
}

func (controller *PublicComments) CreateImageComment(ctx context.Context, req *v1.CreateImageCommentReq) (*v1.CreateImageCommentRes, error) {
	subject, _ := optionalSubject(ctx)
	actor := gallerycomments.Actor{}
	if subject.Kind == "user" {
		actor.UserKey = subject.ID
	}
	request := ghttp.RequestFromCtx(ctx)
	if request != nil {
		actor.IP = request.GetClientIp()
		actor.Agent = request.Request.UserAgent()
	}
	attemptID := strings.TrimSpace(req.AbuseAttemptID)
	if attemptID == "" {
		attemptID = identifier.MustNew().String()
	}
	comment, pending, err := controller.module.Create(ctx, actor, req.ImageID, gallerycomments.CreateInput{
		Content: req.Content, ParentID: req.ParentID, AuthorName: req.AuthorName,
		AuthorEmail: req.AuthorEmail, AbuseAttemptID: attemptID, ChallengeProof: req.ChallengeProof,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateImageCommentRes{Comment: publicCommentView(comment), Pending: pending}, nil
}

type Comments struct{ module *gallerycomments.Module }

func NewComments(module *gallerycomments.Module) *Comments { return &Comments{module: module} }

func (controller *Comments) ListAdminComments(ctx context.Context, req *v1.ListAdminCommentsReq) (*v1.ListAdminCommentsRes, error) {
	if _, err := requireCapability(ctx, galleryauthz.CapabilityCommentRead); err != nil {
		return nil, err
	}
	if req.SortBy != "" && req.SortBy != "createdAt" {
		return nil, galleryerr.Validation("sortBy", "unsupported comment sort field")
	}
	if req.SortOrder != "" && req.SortOrder != "asc" && req.SortOrder != "desc" {
		return nil, galleryerr.Validation("sortOrder", "unsupported comment sort order")
	}
	page, err := controller.module.ListAdmin(ctx, gallerycomments.AdminQuery{
		Status: gallerycomments.Status(strings.TrimSpace(req.Status)), Keyword: req.Q,
		Ascending: req.SortOrder == "asc", Page: req.Page, Size: req.Size,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListAdminCommentsRes{Items: adminCommentViews(page.Items), Total: page.Total, Page: page.Page, Size: page.Size}, nil
}

func (controller *Comments) ModerateComment(ctx context.Context, req *v1.ModerateCommentReq) (*v1.ModerateCommentRes, error) {
	if _, err := requireCapability(ctx, galleryauthz.CapabilityCommentModerate); err != nil {
		return nil, err
	}
	comment, err := controller.module.SetStatus(ctx, req.CommentID, gallerycomments.Status(req.Status))
	if err != nil {
		return nil, err
	}
	return &v1.ModerateCommentRes{Comment: adminCommentView(gallerycomments.AdminComment{Comment: comment})}, nil
}

func (controller *Comments) DeleteComment(ctx context.Context, req *v1.DeleteCommentReq) (*v1.DeleteCommentRes, error) {
	if _, err := requireCapability(ctx, galleryauthz.CapabilityCommentDelete); err != nil {
		return nil, err
	}
	if err := controller.module.Delete(ctx, req.CommentID); err != nil {
		return nil, err
	}
	return &v1.DeleteCommentRes{Deleted: true}, nil
}

func publicCommentViews(threads []gallerycomments.Thread) []v1.CommentView {
	result := make([]v1.CommentView, 0, len(threads))
	for _, thread := range threads {
		view := publicCommentView(thread.Comment)
		for _, reply := range thread.Replies {
			view.Replies = append(view.Replies, publicCommentView(reply))
		}
		result = append(result, view)
	}
	return result
}

func publicCommentView(comment gallerycomments.Comment) v1.CommentView {
	return v1.CommentView{
		ID: comment.ID, ParentID: comment.ParentID, AuthorName: comment.AuthorName,
		AvatarURL: commentAvatarURL(comment.AvatarMediaKey), IsAnonymous: comment.UserKey == "",
		Content: comment.Content, CreatedAt: comment.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func adminCommentViews(comments []gallerycomments.AdminComment) []v1.AdminCommentView {
	result := make([]v1.AdminCommentView, 0, len(comments))
	for _, comment := range comments {
		result = append(result, adminCommentView(comment))
	}
	return result
}

func adminCommentView(item gallerycomments.AdminComment) v1.AdminCommentView {
	comment := item.Comment
	return v1.AdminCommentView{
		ID: comment.ID, ImageID: comment.ImageID, ImageTitle: item.ImageTitle,
		ParentID: comment.ParentID, AuthorName: comment.AuthorName,
		AvatarURL: commentAvatarURL(comment.AvatarMediaKey), AuthorEmail: comment.AuthorEmail,
		UserKey: comment.UserKey, Content: comment.Content, Status: string(comment.Status),
		IP: comment.IP, CreatedAt: comment.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func commentAvatarURL(mediaKey string) string {
	if strings.TrimSpace(mediaKey) == "" {
		return ""
	}
	return "/media/" + url.PathEscape(mediaKey) + "?format=webp&name=thumbnail&v=1"
}
