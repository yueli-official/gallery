// Package galleryauthz owns Gallery's instance-local authorization declaration.
// Foundation owns execution and persistence; Gallery owns its roles and capabilities.
package galleryauthz

import (
	"context"

	"github.com/yueli-official/foundation/go/authorization"
)

const (
	RootScopeID authorization.ScopeID   = "gallery"
	ScopeSite   authorization.ScopeType = "site"

	RoleAdministrator   authorization.RoleKey = "administrator"
	RoleContentOperator authorization.RoleKey = "content_operator"

	CapabilityPublicRead                   authorization.CapabilityKey = "gallery.public.read"
	CapabilitySubmissionCreate             authorization.CapabilityKey = "gallery.submission.create"
	CapabilitySubmissionReadOwn            authorization.CapabilityKey = "gallery.submission.read_own"
	CapabilitySubmissionWithdraw           authorization.CapabilityKey = "gallery.submission.withdraw"
	CapabilityFavoriteRead                 authorization.CapabilityKey = "gallery.favorite.read"
	CapabilityFavoriteManage               authorization.CapabilityKey = "gallery.favorite.manage"
	CapabilityCommentCreate                authorization.CapabilityKey = "gallery.comment.create"
	CapabilityDashboardRead                authorization.CapabilityKey = "gallery.dashboard.read"
	CapabilityImageRead                    authorization.CapabilityKey = "gallery.image.read"
	CapabilityImageUpdate                  authorization.CapabilityKey = "gallery.image.update"
	CapabilityImageHide                    authorization.CapabilityKey = "gallery.image.hide"
	CapabilitySubmissionRead               authorization.CapabilityKey = "gallery.submission.read"
	CapabilitySubmissionReview             authorization.CapabilityKey = "gallery.submission.review"
	CapabilityCollectionRead               authorization.CapabilityKey = "gallery.collection.read"
	CapabilityCollectionManage             authorization.CapabilityKey = "gallery.collection.manage"
	CapabilityClassificationRead           authorization.CapabilityKey = "gallery.classification.read"
	CapabilityClassificationProposalReview authorization.CapabilityKey = "gallery.classification.proposal_review"
	CapabilityClassificationGovern         authorization.CapabilityKey = "gallery.classification.govern"
	CapabilityCaseRead                     authorization.CapabilityKey = "gallery.case.read"
	CapabilityCaseResolve                  authorization.CapabilityKey = "gallery.case.resolve"
	CapabilityDiscoveryRead                authorization.CapabilityKey = "gallery.discovery.read"
	CapabilityDiscoveryManage              authorization.CapabilityKey = "gallery.discovery.manage"
	CapabilityAssetSettingsManage          authorization.CapabilityKey = "gallery.asset_settings.manage"
	CapabilityCommentRead                  authorization.CapabilityKey = "gallery.comment.read"
	CapabilityCommentModerate              authorization.CapabilityKey = "gallery.comment.moderate"
	CapabilityCommentDelete                authorization.CapabilityKey = "gallery.comment.delete"

	PredicateRegistrationContentOperator    authorization.PredicateKey = "gallery.registration_auto_content_operator"
	TriggerUserRegistered                   authorization.TriggerKey   = "identity.user.registered"
	AutomaticRegistrationContentOperatorKey                            = "gallery.registration_content_operator"
)

func Definition() authorization.Definition {
	operatorCapabilities := []authorization.CapabilityKey{
		CapabilityDashboardRead,
		CapabilityImageRead,
		CapabilityImageUpdate,
		CapabilityImageHide,
		CapabilitySubmissionRead,
		CapabilitySubmissionReview,
		CapabilityCollectionRead,
		CapabilityCollectionManage,
		CapabilityClassificationRead,
		CapabilityClassificationProposalReview,
		CapabilityCaseRead,
		CapabilityDiscoveryRead,
		CapabilityCommentRead,
		CapabilityCommentModerate,
		CapabilityCommentDelete,
	}
	administratorCapabilities := append(
		[]authorization.CapabilityKey{
			authorization.CapabilityManage,
			authorization.CapabilityAuditRead,
		},
		operatorCapabilities...,
	)
	administratorCapabilities = append(
		administratorCapabilities,
		CapabilityClassificationGovern,
		CapabilityCaseResolve,
		CapabilityDiscoveryManage,
		CapabilityAssetSettingsManage,
	)

	return authorization.Definition{
		Consumer: "gallery",
		Version:  4,
		Capabilities: []authorization.CapabilityDefinition{
			{
				Key: CapabilityPublicRead, Version: 1,
				Binding:       authorization.BindingAccessLayerEligible,
				AllowedScopes: []authorization.ScopeType{ScopeSite},
			},
			accessCapability(CapabilitySubmissionCreate),
			accessCapability(CapabilitySubmissionReadOwn),
			accessCapability(CapabilitySubmissionWithdraw),
			accessCapability(CapabilityFavoriteRead),
			accessCapability(CapabilityFavoriteManage),
			accessCapability(CapabilityCommentCreate),
			normalCapability(CapabilityDashboardRead),
			normalCapability(CapabilityImageRead),
			normalCapability(CapabilityImageUpdate),
			normalCapability(CapabilityImageHide),
			normalCapability(CapabilitySubmissionRead),
			normalCapability(CapabilitySubmissionReview),
			normalCapability(CapabilityCollectionRead),
			normalCapability(CapabilityCollectionManage),
			normalCapability(CapabilityClassificationRead),
			normalCapability(CapabilityClassificationProposalReview),
			protectedCapability(CapabilityClassificationGovern),
			normalCapability(CapabilityCaseRead),
			protectedCapability(CapabilityCaseResolve),
			normalCapability(CapabilityDiscoveryRead),
			protectedCapability(CapabilityDiscoveryManage),
			protectedCapability(CapabilityAssetSettingsManage),
			normalCapability(CapabilityCommentRead),
			normalCapability(CapabilityCommentModerate),
			normalCapability(CapabilityCommentDelete),
		},
		Scopes: authorization.ScopeSchema{Types: []authorization.ScopeTypeDefinition{
			{Key: ScopeSite, Root: true},
		}},
		AccessLayers: []authorization.AccessLayerDefinition{
			{
				Key:          authorization.AccessLayerVisitor,
				Capabilities: []authorization.CapabilityKey{CapabilityPublicRead},
			},
			{
				Key: authorization.AccessLayerAuthenticated,
				Capabilities: []authorization.CapabilityKey{
					authorization.CapabilityApplicationCreate,
					authorization.CapabilityApplicationReadOwn,
					authorization.CapabilityApplicationWithdraw,
					authorization.CapabilityInvitationAccept,
					CapabilitySubmissionCreate,
					CapabilitySubmissionReadOwn,
					CapabilitySubmissionWithdraw,
					CapabilityFavoriteRead,
					CapabilityFavoriteManage,
					CapabilityCommentCreate,
				},
			},
		},
		Roles: []authorization.RoleDefinition{
			{
				Key: RoleAdministrator, DisplayName: "管理员", Protected: true,
				Capabilities: administratorCapabilities,
			},
			{
				Key: RoleContentOperator, DisplayName: "内容运营者",
				Capabilities: operatorCapabilities,
				Assignment: authorization.AssignmentPolicy{Sources: []authorization.GrantSource{
					authorization.GrantSourceApplication,
					authorization.GrantSourceInvitation,
					authorization.GrantSourceDirect,
					authorization.GrantSourceAutomatic,
					authorization.GrantSourceGroup,
				}},
			},
		},
		Automatic: []authorization.AutomaticRuleDefinition{{
			Key: AutomaticRegistrationContentOperatorKey, Trigger: TriggerUserRegistered,
			Predicate: PredicateRegistrationContentOperator, Role: RoleContentOperator, Enabled: false,
		}},
	}
}

func accessCapability(key authorization.CapabilityKey) authorization.CapabilityDefinition {
	return authorization.CapabilityDefinition{
		Key: key, Version: 1, Binding: authorization.BindingAccessLayerEligible,
		AllowedScopes: []authorization.ScopeType{ScopeSite},
	}
}

func PredicateEvaluators() map[authorization.PredicateKey]authorization.PredicateEvaluator {
	return map[authorization.PredicateKey]authorization.PredicateEvaluator{
		PredicateRegistrationContentOperator: authorization.PredicateFunc(
			func(_ context.Context, input authorization.PredicateInput) bool {
				return input.Subject.Kind == authorization.SubjectUser
			},
		),
	}
}

func normalCapability(key authorization.CapabilityKey) authorization.CapabilityDefinition {
	return authorization.CapabilityDefinition{
		Key: key, Version: 1, Binding: authorization.BindingNormal,
		AllowedScopes: []authorization.ScopeType{ScopeSite},
		EligibleSubjects: []authorization.SubjectKind{
			authorization.SubjectUser,
			authorization.SubjectService,
		},
		Delegable: true,
	}
}

func protectedCapability(key authorization.CapabilityKey) authorization.CapabilityDefinition {
	return authorization.CapabilityDefinition{
		Key: key, Version: 1, Binding: authorization.BindingProtectedOnly,
		Risk: authorization.RiskHigh, Audit: authorization.AuditFull,
		AllowedScopes: []authorization.ScopeType{ScopeSite},
		EligibleSubjects: []authorization.SubjectKind{
			authorization.SubjectUser,
			authorization.SubjectService,
		},
	}
}
