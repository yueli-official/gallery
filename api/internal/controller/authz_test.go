package controller

import (
	"context"
	"testing"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
	"platform/products/gallery/api/internal/galleryauthz"
)

func newGalleryAuthorization(t *testing.T) (*galleryauthz.Service, authorization.SubjectRef) {
	t.Helper()
	administrator := authorization.SubjectRef{Kind: authorization.SubjectUser, ID: "gallery-administrator"}
	runtime, err := authorization.NewMemory(
		authorization.MustCompile(galleryauthz.Definition()),
		authorization.MemoryOptions{
			RootScopeID:       galleryauthz.RootScopeID,
			ProtectedSubjects: []authorization.SubjectRef{administrator},
			Predicates:        galleryauthz.PredicateEvaluators(),
		},
	)
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	return galleryauthz.New(runtime), administrator
}

func authorizationContext(subject authorization.SubjectRef, roles []string, service *galleryauthz.Service) context.Context {
	ctx := foundationauth.NewContext(context.Background(), &foundationauth.Principal{
		Subject: subject.ID,
		Roles:   roles,
	})
	return context.WithValue(ctx, authorizationContextKey{}, service)
}

func TestGalleryAdministratorComesFromLocalAuthorization(t *testing.T) {
	service, administrator := newGalleryAuthorization(t)
	if _, err := requireAdmin(authorizationContext(administrator, nil, service)); err != nil {
		t.Fatalf("gallery administrator rejected: %v", err)
	}

	identityAdmin := authorization.SubjectRef{Kind: authorization.SubjectUser, ID: "identity-admin"}
	if _, err := requireAdmin(authorizationContext(identityAdmin, []string{"admin"}, service)); err == nil {
		t.Fatal("Identity admin unexpectedly became Gallery administrator")
	}
}

func TestContentOperatorGetsDailyCapabilitiesButNotProtectedGovernance(t *testing.T) {
	service, administrator := newGalleryAuthorization(t)
	operator := authorization.SubjectRef{Kind: authorization.SubjectUser, ID: "content-operator"}
	if _, err := service.Runtime().Grant(context.Background(), authorization.GrantCommand{
		Actor: administrator, Target: operator, Role: galleryauthz.RoleContentOperator,
		ScopeID: galleryauthz.RootScopeID, Source: authorization.GrantSourceDirect,
	}); err != nil {
		t.Fatalf("Grant() error = %v", err)
	}
	ctx := authorizationContext(operator, nil, service)
	for _, capability := range []authorization.CapabilityKey{
		galleryauthz.CapabilityImageUpdate,
		galleryauthz.CapabilitySubmissionReview,
		galleryauthz.CapabilityCollectionManage,
		galleryauthz.CapabilityClassificationProposalReview,
	} {
		if _, err := requireCapability(ctx, capability); err != nil {
			t.Fatalf("content operator rejected for %q: %v", capability, err)
		}
	}
	for _, capability := range []authorization.CapabilityKey{
		galleryauthz.CapabilityClassificationGovern,
		galleryauthz.CapabilityCaseResolve,
		galleryauthz.CapabilityAssetSettingsManage,
	} {
		if _, err := requireCapability(ctx, capability); err == nil {
			t.Fatalf("content operator unexpectedly allowed for %q", capability)
		}
	}
}
