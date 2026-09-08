// Package gallerycomments owns public Image discussion and site-wide comment moderation.
package gallerycomments

import (
	"context"
	"strings"
	"time"

	"github.com/yueli-official/foundation/go/identifier"
	galleryservice "github.com/yueli-official/gallery/api/internal/gallery"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusSpam     Status = "spam"
	StatusTrash    Status = "trash"
)

type Actor struct {
	UserKey string
	IP      string
	Agent   string
}

type CreateInput struct {
	Content        string
	ParentID       string
	AuthorName     string
	AuthorEmail    string
	AbuseAttemptID string
	ChallengeProof string
}

type Comment struct {
	ID             string
	ImageID        string
	ParentID       string
	UserKey        string
	AuthorName     string
	AuthorEmail    string
	AvatarMediaKey string
	Content        string
	Status         Status
	IP             string
	UserAgent      string
	CreatedAt      time.Time
	DeletedAt      *time.Time
}

type Thread struct {
	Comment Comment
	Replies []Comment
}

type ImageHead struct {
	ID    string
	Title string
}

type AdminComment struct {
	ImageAssetID     string
	ParentUserKey    string
	ParentAuthorName string
	ParentContent    string
	Comment
	ImageTitle string
}

type PublicPage struct {
	Items []Thread
	Total int
	Page  int
	Size  int
}

type AdminQuery struct {
	Status    Status
	Keyword   string
	Ascending bool
	Page      int
	Size      int
}

type AdminPage struct {
	Items []AdminComment
	Total int
	Page  int
	Size  int
}

// Store is the internal persistence seam. The Module owns validation,
// two-level threading, public-ID normalization and presentation.
type Store interface {
	PublicImage(context.Context, string) (*ImageHead, error)
	Comment(context.Context, string) (*Comment, error)
	InsertComment(context.Context, Comment) error
	PublicComments(context.Context, string, bool, int, int) ([]Comment, int, error)
	Replies(context.Context, []string) (map[string][]Comment, error)
	AdminComments(context.Context, AdminQuery) ([]AdminComment, int, error)
	SetStatus(context.Context, string, Status) (bool, error)
	SoftDelete(context.Context, string) (bool, error)
}

type PublicProfile struct {
	UserKey        string
	DisplayName    string
	AvatarMediaKey string
}

type ProfileResolver interface {
	Resolve(context.Context, []string) map[string]PublicProfile
}

type Guard interface {
	Admit(context.Context, Actor, string, string) error
}

type Module struct {
	store    Store
	profiles ProfileResolver
	guard    Guard
}

func New(store Store, profiles ProfileResolver, guards ...Guard) *Module {
	module := &Module{store: store, profiles: profiles}
	if len(guards) > 0 {
		module.guard = guards[0]
	}
	return module
}

func (module *Module) ListPublic(ctx context.Context, rawImageID string, ascending bool, page, size int) (*PublicPage, error) {
	imageID, err := galleryservice.DatabaseID(rawImageID)
	if err != nil {
		return nil, galleryerr.Validation("imageId", "invalid image identifier")
	}
	page, size = normalizePage(page, size)
	if module.store == nil {
		return nil, galleryerr.NotInitialized("comments")
	}
	if image, err := module.store.PublicImage(ctx, imageID); err != nil {
		return nil, err
	} else if image == nil {
		return nil, galleryerr.NotFound("image", rawImageID)
	}
	tops, total, err := module.store.PublicComments(ctx, imageID, ascending, size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(tops))
	for _, comment := range tops {
		ids = append(ids, comment.ID)
	}
	replies, err := module.store.Replies(ctx, ids)
	if err != nil {
		return nil, err
	}
	profiles := module.resolveProfiles(ctx, tops, replies)
	items := make([]Thread, 0, len(tops))
	for _, top := range tops {
		thread := Thread{Comment: module.present(top, profiles)}
		for _, reply := range replies[top.ID] {
			thread.Replies = append(thread.Replies, module.present(reply, profiles))
		}
		items = append(items, thread)
	}
	return &PublicPage{Items: items, Total: total, Page: page, Size: size}, nil
}

func (module *Module) Create(ctx context.Context, actor Actor, rawImageID string, input CreateInput) (Comment, bool, error) {
	if module.store == nil {
		return Comment{}, false, galleryerr.NotInitialized("comments")
	}
	imageID, err := galleryservice.DatabaseID(rawImageID)
	if err != nil {
		return Comment{}, false, galleryerr.Validation("imageId", "invalid image identifier")
	}
	content := strings.TrimSpace(input.Content)
	if length := len([]rune(content)); length < 1 || length > 5000 {
		return Comment{}, false, galleryerr.Validation("content", "comment must contain 1 to 5000 characters")
	}
	if image, err := module.store.PublicImage(ctx, imageID); err != nil {
		return Comment{}, false, err
	} else if image == nil {
		return Comment{}, false, galleryerr.NotFound("image", rawImageID)
	}
	comment := Comment{
		ID: identifier.MustNew().String(), ImageID: imageID, Content: content,
		UserKey: strings.TrimSpace(actor.UserKey), IP: strings.TrimSpace(actor.IP),
		UserAgent: strings.TrimSpace(actor.Agent), CreatedAt: time.Now().UTC(),
	}
	if comment.UserKey != "" {
		comment.Status = StatusApproved
	} else {
		comment.AuthorName = strings.TrimSpace(input.AuthorName)
		comment.AuthorEmail = strings.TrimSpace(input.AuthorEmail)
		if length := len([]rune(comment.AuthorName)); length < 1 || length > 40 {
			return Comment{}, false, galleryerr.Validation("authorName", "anonymous comments require a name of at most 40 characters")
		}
		if comment.AuthorEmail != "" && !looksLikeEmail(comment.AuthorEmail) {
			return Comment{}, false, galleryerr.Validation("authorEmail", "invalid email")
		}
		comment.Status = StatusPending
	}
	if strings.TrimSpace(input.ParentID) != "" {
		parentID, err := galleryservice.DatabaseID(input.ParentID)
		if err != nil {
			return Comment{}, false, galleryerr.Validation("parentId", "invalid parent comment identifier")
		}
		parent, err := module.store.Comment(ctx, parentID)
		if err != nil {
			return Comment{}, false, err
		}
		if parent == nil || parent.ImageID != imageID {
			return Comment{}, false, galleryerr.Validation("parentId", "parent comment was not found on this image")
		}
		if parent.ParentID != "" {
			comment.ParentID = parent.ParentID
		} else {
			comment.ParentID = parent.ID
		}
	}
	if countLinks(content) > 2 {
		comment.Status = StatusPending
	}
	if module.guard != nil {
		if err := module.guard.Admit(ctx, actor, input.AbuseAttemptID, input.ChallengeProof); err != nil {
			return Comment{}, false, err
		}
	}
	if err := module.store.InsertComment(ctx, comment); err != nil {
		return Comment{}, false, err
	}
	profiles := module.resolveProfiles(ctx, []Comment{comment}, nil)
	comment = module.present(comment, profiles)
	return comment, comment.Status != StatusApproved, nil
}

func (module *Module) ListAdmin(ctx context.Context, query AdminQuery) (*AdminPage, error) {
	if module.store == nil {
		return nil, galleryerr.NotInitialized("comments")
	}
	query.Page, query.Size = normalizePage(query.Page, query.Size)
	query.Keyword = strings.TrimSpace(query.Keyword)
	if query.Status != "" && !validStatus(query.Status) {
		return nil, galleryerr.Validation("status", "invalid comment status")
	}
	items, total, err := module.store.AdminComments(ctx, query)
	if err != nil {
		return nil, err
	}
	comments := make([]Comment, 0, len(items))
	for _, item := range items {
		comments = append(comments, item.Comment)
		comments = append(comments, Comment{UserKey: item.ParentUserKey})
	}
	profiles := module.resolveProfiles(ctx, comments, nil)
	for index := range items {
		items[index].Comment = module.present(items[index].Comment, profiles)
		if items[index].ParentContent != "" {
			parent := module.present(Comment{UserKey: items[index].ParentUserKey, AuthorName: items[index].ParentAuthorName}, profiles)
			items[index].ParentAuthorName = parent.AuthorName
			if items[index].ParentAuthorName == "" {
				items[index].ParentAuthorName = "匿名用户"
			}
		}
	}
	return &AdminPage{Items: items, Total: total, Page: query.Page, Size: query.Size}, nil
}

func (module *Module) SetStatus(ctx context.Context, rawID string, status Status) (Comment, error) {
	if module.store == nil {
		return Comment{}, galleryerr.NotInitialized("comments")
	}
	if status != StatusApproved && status != StatusSpam && status != StatusTrash {
		return Comment{}, galleryerr.Validation("status", "invalid moderation target")
	}
	id, err := galleryservice.DatabaseID(rawID)
	if err != nil {
		return Comment{}, galleryerr.Validation("commentId", "invalid comment identifier")
	}
	changed, err := module.store.SetStatus(ctx, id, status)
	if err != nil {
		return Comment{}, err
	}
	if !changed {
		return Comment{}, galleryerr.NotFound("comment", rawID)
	}
	comment, err := module.store.Comment(ctx, id)
	if err != nil {
		return Comment{}, err
	}
	if comment == nil {
		return Comment{}, galleryerr.NotFound("comment", rawID)
	}
	return module.present(*comment, module.resolveProfiles(ctx, []Comment{*comment}, nil)), nil
}

func (module *Module) Delete(ctx context.Context, rawID string) error {
	if module.store == nil {
		return galleryerr.NotInitialized("comments")
	}
	id, err := galleryservice.DatabaseID(rawID)
	if err != nil {
		return galleryerr.Validation("commentId", "invalid comment identifier")
	}
	deleted, err := module.store.SoftDelete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return galleryerr.NotFound("comment", rawID)
	}
	return nil
}

func (module *Module) present(comment Comment, profiles map[string]PublicProfile) Comment {
	if comment.UserKey != "" {
		comment.AuthorName = comment.UserKey
		if profile, ok := profiles[comment.UserKey]; ok {
			if profile.DisplayName != "" {
				comment.AuthorName = profile.DisplayName
			}
			comment.AvatarMediaKey = profile.AvatarMediaKey
		}
	}
	comment.ID = galleryservice.PublicID(comment.ID)
	comment.ImageID = galleryservice.PublicID(comment.ImageID)
	comment.ParentID = galleryservice.PublicID(comment.ParentID)
	return comment
}

func (module *Module) resolveProfiles(ctx context.Context, comments []Comment, replies map[string][]Comment) map[string]PublicProfile {
	if module.profiles == nil {
		return map[string]PublicProfile{}
	}
	ids := make([]string, 0, len(comments))
	seen := map[string]bool{}
	appendID := func(value string) {
		if value != "" && !seen[value] {
			seen[value] = true
			ids = append(ids, value)
		}
	}
	for _, comment := range comments {
		appendID(comment.UserKey)
	}
	for _, children := range replies {
		for _, comment := range children {
			appendID(comment.UserKey)
		}
	}
	return module.profiles.Resolve(ctx, ids)
}

func normalizePage(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return page, size
}

func validStatus(status Status) bool {
	return status == StatusPending || status == StatusApproved || status == StatusSpam || status == StatusTrash
}

func looksLikeEmail(value string) bool {
	at := strings.IndexByte(value, '@')
	return at > 0 && at < len(value)-1 && strings.Contains(value[at+1:], ".")
}

func countLinks(value string) int {
	return strings.Count(strings.ToLower(value), "http://") + strings.Count(strings.ToLower(value), "https://")
}
