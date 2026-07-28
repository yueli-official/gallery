package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"github.com/yueli-official/gallery/api/internal/model"
)

type CreateSubmissionReq struct {
	g.Meta `path:"/api/v1/gallery/submissions" method:"POST" tags:"Gallery submissions" summary:"Submit one uploaded image"`
	model.SubmissionInput
	AbuseAttemptID string `json:"abuseAttemptId" v:"length:0,128"`
	ChallengeProof string `json:"challengeProof" v:"length:0,4096"`
}
type CreateSubmissionRes struct {
	Submission model.Submission `json:"submission"`
}

type ClaimGuestSubmissionsReq struct {
	g.Meta `path:"/api/v1/gallery/guest-claims" method:"POST" tags:"Gallery submissions" summary:"Transfer guest submissions to the signed-in user"`
}

type ClaimGuestSubmissionsRes struct {
	Claimed int64 `json:"claimed"`
}

type ListMySubmissionsReq struct {
	g.Meta          `path:"/api/v1/gallery/me/submissions" method:"GET" tags:"Gallery submissions" summary:"List submissions owned by the current user or guest"`
	Page            int    `p:"page" d:"1"`
	Size            int    `p:"size" d:"20"`
	Outcome         string `p:"outcome"`
	ProcessingState string `p:"processingState"`
	ReviewState     string `p:"reviewState"`
}
type ListMySubmissionsRes struct {
	Submissions []model.Submission `json:"submissions"`
	Total       int                `json:"total"`
	Page        int                `json:"page"`
	PageSize    int                `json:"pageSize"`
	TotalPages  int                `json:"totalPages"`
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
	Page   int    `p:"page" d:"1"`
	Size   int    `p:"size" d:"24"`
	Sort   string `p:"sort" d:"newest"`
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

type GetAdminSiteSettingsReq struct {
	g.Meta `path:"/api/v1/gallery/admin/site-settings" method:"GET" tags:"Gallery admin" summary:"Get editable Gallery site and homepage settings"`
}
type GetAdminSiteSettingsRes struct {
	Site model.SiteSettings `json:"site"`
}

type UpdateAdminSiteSettingsReq struct {
	g.Meta `path:"/api/v1/gallery/admin/site-settings" method:"PATCH" tags:"Gallery admin" summary:"Update Gallery site and homepage settings"`
	model.SiteSettingsUpdateInput
}
type UpdateAdminSiteSettingsRes struct {
	Site model.SiteSettings `json:"site"`
}

type GetClassificationCatalogReq struct {
	g.Meta `path:"/api/v1/gallery/admin/classification" method:"GET" tags:"Gallery admin" summary:"Get the authoritative classification catalog including inactive identities"`
}
type GetClassificationCatalogRes struct {
	Catalog model.ClassificationCatalog `json:"catalog"`
}

type ListClassificationTagsReq struct {
	g.Meta `path:"/api/v1/gallery/admin/classification/tags" method:"GET" tags:"Gallery admin" summary:"List classification tags with a stable keyset cursor"`
	Cursor string `p:"cursor"`
	Size   int    `p:"size" d:"50"`
}
type ListClassificationTagsRes struct {
	Page model.ClassificationTagPage `json:"page"`
}

type ListClassificationTagProposalsReq struct {
	g.Meta `path:"/api/v1/gallery/admin/classification/tag-proposals" method:"GET" tags:"Gallery admin" summary:"List tag proposals awaiting or completing governance"`
	Status string `p:"status" d:"pending"`
	Page   int    `p:"page" d:"1"`
	Size   int    `p:"size" d:"30"`
}
type ListClassificationTagProposalsRes struct {
	Proposals []model.ClassificationTagProposal `json:"proposals"`
	Total     int                               `json:"total"`
}

type ReviewClassificationTagProposalReq struct {
	g.Meta     `path:"/api/v1/gallery/admin/classification/tag-proposals/{proposalId}/review" method:"POST" tags:"Gallery admin" summary:"Resolve, create, alias or reject a tag proposal"`
	ProposalID string `p:"proposalId" v:"required"`
	model.ClassificationTagProposalReviewInput
}
type ReviewClassificationTagProposalRes struct {
	Proposal model.ClassificationTagProposal `json:"proposal"`
}

type PreviewClassificationGovernanceReq struct {
	g.Meta `path:"/api/v1/gallery/admin/classification/governance/preview" method:"POST" tags:"Gallery admin" summary:"Preview an immutable classification governance plan"`
	model.ClassificationGovernancePreviewInput
}
type PreviewClassificationGovernanceRes struct {
	Preview model.ClassificationGovernancePreview `json:"preview"`
}

type ExecuteClassificationGovernanceReq struct {
	g.Meta `path:"/api/v1/gallery/admin/classification/governance/execute" method:"POST" tags:"Gallery admin" summary:"Execute a freshly previewed classification governance plan"`
	model.ClassificationGovernanceExecuteInput
}
type ExecuteClassificationGovernanceRes struct {
	Execution model.ClassificationGovernanceExecution `json:"execution"`
}

type ListSubmissionReviewsReq struct {
	g.Meta          `path:"/api/v1/gallery/admin/submissions" method:"GET" tags:"Gallery admin" summary:"List Gallery submissions across processing, review, safety and outcome states"`
	Q               string `p:"q"`
	Sort            string `p:"sort" d:"oldest"`
	Page            int    `p:"page" d:"1"`
	Size            int    `p:"size" d:"20"`
	ProcessingState string `p:"processingState"`
	ReviewState     string `p:"reviewState"`
	SafetyState     string `p:"safetyState"`
	Outcome         string `p:"outcome"`
}
type ListSubmissionReviewsRes struct{ model.AdminSubmissionPage }

type ReviewSubmissionReq struct {
	g.Meta       `path:"/api/v1/gallery/admin/submissions/{submissionId}/review" method:"POST" tags:"Gallery admin" summary:"Approve or reject a ready submission"`
	SubmissionID string `p:"submissionId" v:"required"`
	model.SubmissionReviewInput
}
type ReviewSubmissionRes struct {
	Submission model.Submission `json:"submission"`
}

type BulkReviewSubmissionsReq struct {
	g.Meta `path:"/api/v1/gallery/admin/submissions/bulk-review" method:"POST" tags:"Gallery admin" summary:"Approve or reject multiple ready submissions with per-item outcomes"`
	model.BulkSubmissionReviewInput
}
type BulkReviewSubmissionsRes struct {
	Results []model.BulkSubmissionReviewResult `json:"results"`
}

type PreviewSubmissionReq struct {
	g.Meta       `path:"/api/v1/gallery/admin/submissions/{submissionId}/preview" method:"GET" tags:"Gallery admin" summary:"Get a short-lived signed submission thumbnail URL"`
	SubmissionID string `p:"submissionId" v:"required"`
}
type PreviewSubmissionRes struct {
	URL string `json:"url"`
}

type HideImageReq struct {
	g.Meta  `path:"/api/v1/gallery/admin/images/{imageId}/hide" method:"POST" tags:"Gallery admin" summary:"Hide a published image"`
	ImageID string `p:"imageId" v:"required"`
	Reason  string `json:"reason" v:"required"`
}
type HideImageRes struct {
	Hidden bool `json:"hidden"`
}

type ListAdminImagesReq struct {
	g.Meta           `path:"/api/v1/gallery/admin/images" method:"GET" tags:"Gallery admin" summary:"List all Gallery images for lifecycle operations"`
	Q                string `p:"q"`
	Sort             string `p:"sort" d:"newest"`
	Page             int    `p:"page" d:"1"`
	Size             int    `p:"size" d:"24"`
	ProcessingState  string `p:"processingState"`
	ReviewState      string `p:"reviewState"`
	PublicationState string `p:"publicationState"`
	SafetyState      string `p:"safetyState"`
	CategoryID       string `p:"categoryId"`
	FacetValueID     string `p:"facetValueId"`
}
type ListAdminImagesRes struct{ model.AdminImagePage }

type GetAdminImageReq struct {
	g.Meta  `path:"/api/v1/gallery/admin/images/{imageId}" method:"GET" tags:"Gallery admin" summary:"Get a reviewed Gallery image with editable classification"`
	ImageID string `p:"imageId" v:"required"`
}
type GetAdminImageRes struct {
	Image model.AdminImage `json:"image"`
}

type UpdateAdminImageReq struct {
	g.Meta  `path:"/api/v1/gallery/admin/images/{imageId}" method:"PATCH" tags:"Gallery admin" summary:"Optimistically update Gallery image metadata"`
	ImageID string `p:"imageId" v:"required"`
	model.AdminImageUpdateInput
}
type UpdateAdminImageRes struct {
	Image model.AdminImage `json:"image"`
}

type BulkHideImagesReq struct {
	g.Meta `path:"/api/v1/gallery/admin/images/bulk-hide" method:"POST" tags:"Gallery admin" summary:"Hide multiple published images with per-item outcomes"`
	model.BulkImageHideInput
}
type BulkHideImagesRes struct {
	Results []model.BulkImageActionResult `json:"results"`
}

type BulkImagesReq struct {
	g.Meta `path:"/api/v1/gallery/admin/images/bulk" method:"POST" tags:"Gallery admin" summary:"Apply a lifecycle or classification action to multiple images with per-item outcomes"`
	model.BulkImageActionInput
}
type BulkImagesRes struct {
	Results []model.BulkImageActionResult `json:"results"`
}

type ListCasesReq struct {
	g.Meta `path:"/api/v1/gallery/admin/cases" method:"GET" tags:"Gallery admin" summary:"List and filter Gallery moderation and correction cases"`
	Q      string `p:"q"`
	Sort   string `p:"sort" d:"oldest"`
	Status string `p:"status" d:"open"`
	Kind   string `p:"kind"`
	Page   int    `p:"page" d:"1"`
	Size   int    `p:"size" d:"20"`
}
type ListCasesRes struct{ model.AdminCasePage }

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

type GetEditorialCollectionReq struct {
	g.Meta       `path:"/api/v1/gallery/admin/collections/{collectionId}" method:"GET" tags:"Gallery admin" summary:"Get one editorial collection with ordered members"`
	CollectionID string `p:"collectionId" v:"required"`
	Page         int    `p:"page" d:"1"`
	Size         int    `p:"size" d:"60"`
}
type GetEditorialCollectionRes struct {
	Collection model.CollectionDetail `json:"collection"`
}

type UpdateEditorialCollectionReq struct {
	g.Meta       `path:"/api/v1/gallery/admin/collections/{collectionId}" method:"PATCH" tags:"Gallery admin" summary:"Update editorial metadata, SEO, cover and visibility"`
	CollectionID string `p:"collectionId" v:"required"`
	model.EditorialCollectionUpdateInput
}
type UpdateEditorialCollectionRes struct {
	Collection model.Collection `json:"collection"`
}

type MutateEditorialMembersReq struct {
	g.Meta       `path:"/api/v1/gallery/admin/collections/{collectionId}/members" method:"POST" tags:"Gallery admin" summary:"Atomically add and remove editorial collection images"`
	CollectionID string `p:"collectionId" v:"required"`
	model.MemberMutationInput
}
type MutateEditorialMembersRes struct {
	Collection model.Collection `json:"collection"`
}

type ReorderEditorialMembersReq struct {
	g.Meta       `path:"/api/v1/gallery/admin/collections/{collectionId}/order" method:"PUT" tags:"Gallery admin" summary:"Replace the manual order for all editorial collection members"`
	CollectionID string `p:"collectionId" v:"required"`
	model.EditorialCollectionOrderInput
}
type ReorderEditorialMembersRes struct {
	Collection model.Collection `json:"collection"`
}
