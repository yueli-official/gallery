package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"platform/products/gallery/api/internal/model"
)

type CreateSubmissionReq struct {
	g.Meta `path:"/api/v1/gallery/submissions" method:"POST" tags:"Gallery submissions" summary:"Submit one uploaded image"`
	model.SubmissionInput
}
type CreateSubmissionRes struct {
	Submission model.Submission `json:"submission"`
}

type ListMySubmissionsReq struct {
	g.Meta `path:"/api/v1/gallery/me/submissions" method:"GET" tags:"Gallery submissions" summary:"List submissions owned by the current user or guest"`
	Page   int `p:"page" d:"1"`
	Size   int `p:"size" d:"20"`
}
type ListMySubmissionsRes struct {
	Submissions []model.Submission `json:"submissions"`
	Total       int                `json:"total"`
}

type WithdrawSubmissionReq struct {
	g.Meta       `path:"/api/v1/gallery/me/submissions/{submissionId}/withdraw" method:"POST" tags:"Gallery submissions" summary:"Withdraw an owned submission and hide its published image"`
	SubmissionID string `p:"submissionId" v:"required"`
}
type WithdrawSubmissionRes struct {
	Submission model.Submission `json:"submission"`
}

type GetFavoritesReq struct {
	g.Meta `path:"/api/v1/gallery/me/favorites" method:"GET" tags:"Gallery favorites" summary:"Get the current user's private singleton favorites collection"`
}
type GetFavoritesRes struct {
	Collection model.CollectionDetail `json:"collection"`
}

type AddFavoriteReq struct {
	g.Meta  `path:"/api/v1/gallery/me/favorites/{imageId}" method:"PUT" tags:"Gallery favorites" summary:"Idempotently add an image to favorites"`
	ImageID string `p:"imageId" v:"required"`
	Version int64  `json:"version"`
}
type AddFavoriteRes struct {
	Collection model.Collection `json:"collection"`
}

type RemoveFavoriteReq struct {
	g.Meta  `path:"/api/v1/gallery/me/favorites/{imageId}" method:"DELETE" tags:"Gallery favorites" summary:"Idempotently remove an image from favorites"`
	ImageID string `p:"imageId" v:"required"`
	Version int64  `p:"version"`
}
type RemoveFavoriteRes struct {
	Collection model.Collection `json:"collection"`
}

type GetAdminOverviewReq struct {
	g.Meta `path:"/api/v1/gallery/admin/overview" method:"GET" tags:"Gallery admin" summary:"Get actionable Gallery operations counts"`
}
type GetAdminOverviewRes struct {
	Overview model.AdminOverview `json:"overview"`
}

type ListSubmissionReviewsReq struct {
	g.Meta `path:"/api/v1/gallery/admin/submissions" method:"GET" tags:"Gallery admin" summary:"List submissions requiring operator review"`
	Page   int `p:"page" d:"1"`
	Size   int `p:"size" d:"20"`
}
type ListSubmissionReviewsRes struct {
	Submissions []model.Submission `json:"submissions"`
	Total       int                `json:"total"`
}

type ReviewSubmissionReq struct {
	g.Meta       `path:"/api/v1/gallery/admin/submissions/{submissionId}/review" method:"POST" tags:"Gallery admin" summary:"Approve or reject a ready submission"`
	SubmissionID string `p:"submissionId" v:"required"`
	model.SubmissionReviewInput
}
type ReviewSubmissionRes struct {
	Submission model.Submission `json:"submission"`
}

type HideImageReq struct {
	g.Meta  `path:"/api/v1/gallery/admin/images/{imageId}/hide" method:"POST" tags:"Gallery admin" summary:"Hide a published image"`
	ImageID string `p:"imageId" v:"required"`
	Reason  string `json:"reason" v:"required"`
}
type HideImageRes struct {
	Hidden bool `json:"hidden"`
}

type ListCasesReq struct {
	g.Meta `path:"/api/v1/gallery/admin/cases" method:"GET" tags:"Gallery admin" summary:"List Gallery moderation and correction cases"`
	Status string `p:"status" d:"open"`
	Page   int    `p:"page" d:"1"`
	Size   int    `p:"size" d:"20"`
}
type ListCasesRes struct {
	Cases []model.Case `json:"cases"`
	Total int          `json:"total"`
}

type ResolveCaseReq struct {
	g.Meta `path:"/api/v1/gallery/admin/cases/{caseId}/resolve" method:"POST" tags:"Gallery admin" summary:"Move a Gallery case to reviewing, resolved or dismissed"`
	CaseID string `p:"caseId" v:"required"`
	model.CaseResolutionInput
}
type ResolveCaseRes struct {
	Case model.Case `json:"case"`
}

type CreateEditorialCollectionReq struct {
	g.Meta `path:"/api/v1/gallery/admin/collections" method:"POST" tags:"Gallery admin" summary:"Create an editorial image collection"`
	model.EditorialCollectionInput
}
type CreateEditorialCollectionRes struct {
	Collection model.Collection `json:"collection"`
}

type ListEditorialCollectionsReq struct {
	g.Meta `path:"/api/v1/gallery/admin/collections" method:"GET" tags:"Gallery admin" summary:"List private and public editorial collections"`
}
type ListEditorialCollectionsRes struct {
	Collections []model.Collection `json:"collections"`
}

type MutateEditorialMembersReq struct {
	g.Meta       `path:"/api/v1/gallery/admin/collections/{collectionId}/members" method:"POST" tags:"Gallery admin" summary:"Atomically add and remove editorial collection images"`
	CollectionID string `p:"collectionId" v:"required"`
	model.MemberMutationInput
}
type MutateEditorialMembersRes struct {
	Collection model.Collection `json:"collection"`
}
