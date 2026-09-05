package v1

import "github.com/gogf/gf/v2/frame/g"

type CommentView struct {
	ID          string        `json:"id"`
	ParentID    string        `json:"parentId,omitempty"`
	AuthorName  string        `json:"authorName"`
	AvatarURL   string        `json:"avatarUrl,omitempty"`
	IsAnonymous bool          `json:"isAnonymous"`
	Content     string        `json:"content"`
	CreatedAt   string        `json:"createdAt"`
	Replies     []CommentView `json:"replies,omitempty"`
}

type AdminCommentView struct {
	ID          string `json:"id"`
	ImageID     string `json:"imageId"`
	ImageTitle  string `json:"imageTitle"`
	ParentID    string `json:"parentId,omitempty"`
	AuthorName  string `json:"authorName"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
	AuthorEmail string `json:"authorEmail,omitempty"`
	UserKey     string `json:"userKey,omitempty"`
	Content     string `json:"content"`
	Status      string `json:"status"`
	IP          string `json:"ip,omitempty"`
	CreatedAt   string `json:"createdAt"`
}

type ListImageCommentsReq struct {
	g.Meta    `path:"/api/v1/gallery/images/{imageId}/comments" method:"GET" tags:"Gallery comments" summary:"List approved comments for a public image"`
	ImageID   string `p:"imageId" v:"required"`
	SortOrder string `p:"sortOrder" d:"asc"`
	Page      int    `p:"page" d:"1"`
	Size      int    `p:"size" d:"20"`
}
type ListImageCommentsRes struct {
	Items []CommentView `json:"items"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
}

type CreateImageCommentReq struct {
	g.Meta         `path:"/api/v1/gallery/images/{imageId}/comments" method:"POST" tags:"Gallery comments" summary:"Comment on a public image"`
	ImageID        string `p:"imageId" v:"required"`
	Content        string `json:"content" v:"required"`
	ParentID       string `json:"parentId"`
	AuthorName     string `json:"authorName"`
	AuthorEmail    string `json:"authorEmail"`
	AbuseAttemptID string `json:"abuseAttemptId,omitempty"`
	ChallengeProof string `json:"challengeProof,omitempty"`
}
type CreateImageCommentRes struct {
	g.Meta  `status:"201"`
	Comment CommentView `json:"comment"`
	Pending bool        `json:"pending"`
}

type ListAdminCommentsReq struct {
	g.Meta    `path:"/api/v1/gallery/admin/comments" method:"GET" tags:"Gallery admin" summary:"List image comments for moderation"`
	Status    string `p:"status"`
	Q         string `p:"q"`
	SortBy    string `p:"sortBy" d:"createdAt"`
	SortOrder string `p:"sortOrder" d:"desc"`
	Page      int    `p:"page" d:"1"`
	Size      int    `p:"size" d:"20"`
}
type ListAdminCommentsRes struct {
	Items []AdminCommentView `json:"items"`
	Total int                `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
}

type ModerateCommentReq struct {
	g.Meta    `path:"/api/v1/gallery/admin/comments/{commentId}" method:"PATCH" tags:"Gallery admin" summary:"Moderate an image comment"`
	CommentID string `p:"commentId" v:"required"`
	Status    string `json:"status" v:"required"`
}
type ModerateCommentRes struct {
	Comment AdminCommentView `json:"comment"`
}

type DeleteCommentReq struct {
	g.Meta    `path:"/api/v1/gallery/admin/comments/{commentId}" method:"DELETE" tags:"Gallery admin" summary:"Delete an image comment and its replies"`
	CommentID string `p:"commentId" v:"required"`
}
type DeleteCommentRes struct {
	g.Meta `status:"204"`
}
