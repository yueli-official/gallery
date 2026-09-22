package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
	v1 "github.com/yueli-official/gallery/api/api/v1"
	"github.com/yueli-official/gallery/api/internal/galleryauthz"
)

func galleryPersonalContext(t *testing.T, user string, capabilities ...string) context.Context {
	t.Helper()
	scopes := make([]string, 0, len(capabilities))
	for _, capability := range capabilities {
		scope, err := foundationauth.PersonalScope("gallery-main-web", capability)
		if err != nil {
			t.Fatal(err)
		}
		scopes = append(scopes, scope)
	}
	endpoint := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		if err := json.NewEncoder(response).Encode(map[string]any{"userKey": user, "scopes": scopes}); err != nil {
			t.Error(err)
		}
	}))
	defer endpoint.Close()
	verifier, err := foundationauth.NewPersonalTokenVerifier(endpoint.URL, "gallery-main-web", nil)
	if err != nil {
		t.Fatal(err)
	}
	principal, err := verifier.Verify(context.Background(), "pat_test")
	if err != nil {
		t.Fatal(err)
	}
	return foundationauth.NewContext(context.Background(), principal)
}

func galleryAuthorization(t *testing.T) (*galleryauthz.Service, *authorization.Memory) {
	t.Helper()
	runtime, err := authorization.NewMemory(authorization.MustCompile(galleryauthz.Definition()), authorization.MemoryOptions{
		RootScopeID:       galleryauthz.RootScopeID,
		ProtectedSubjects: []authorization.SubjectRef{{Kind: authorization.SubjectUser, ID: "admin"}},
		Predicates:        galleryauthz.PredicateEvaluators(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return galleryauthz.New(runtime), runtime
}

func TestPersonalScopeIntersectsCurrentGalleryRights(t *testing.T) {
	service, runtime := galleryAuthorization(t)
	for _, test := range []struct {
		name, user, selected string
		required             authorization.CapabilityKey
		allowed              bool
	}{
		{"ordinary submission", "user", string(galleryauthz.CapabilitySubmissionCreate), galleryauthz.CapabilitySubmissionCreate, true},
		{"admin image update", "admin", string(galleryauthz.CapabilityImageUpdate), galleryauthz.CapabilityImageUpdate, true},
		{"admin unselected", "admin", string(galleryauthz.CapabilityImageRead), galleryauthz.CapabilityImageUpdate, false},
		{"scope cannot grant image update", "user", string(galleryauthz.CapabilityImageUpdate), galleryauthz.CapabilityImageUpdate, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			allowed, err := service.CheckCapability(galleryPersonalContext(t, test.user, test.selected), test.required)
			if err != nil || allowed != test.allowed {
				t.Fatalf("allowed=%v error=%v want=%v", allowed, err, test.allowed)
			}
		})
	}

	admin := authorization.SubjectRef{Kind: authorization.SubjectUser, ID: "admin"}
	grant, err := runtime.Grant(context.Background(), authorization.GrantCommand{
		Actor: admin, Target: authorization.SubjectRef{Kind: authorization.SubjectUser, ID: "temporary"},
		Role: galleryauthz.RoleContentOperator, ScopeID: galleryauthz.RootScopeID, Source: authorization.GrantSourceDirect,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := galleryPersonalContext(t, "temporary", string(galleryauthz.CapabilityImageRead))
	if allowed, err := service.CheckCapability(ctx, galleryauthz.CapabilityImageRead); err != nil || !allowed {
		t.Fatalf("granted operator denied: allowed=%v error=%v", allowed, err)
	}
	if _, err := runtime.Revoke(context.Background(), authorization.RevokeCommand{Actor: admin, GrantID: grant.ID}); err != nil {
		t.Fatal(err)
	}
	if allowed, err := service.CheckCapability(ctx, galleryauthz.CapabilityImageRead); err != nil || allowed {
		t.Fatalf("revoked operator retained access: allowed=%v error=%v", allowed, err)
	}
}

func TestPersonalDirectoryRequiresTrustedIdentityAndCurrentRights(t *testing.T) {
	service, _ := galleryAuthorization(t)
	controller := NewPersonalPermissions("gallery-main-web", service)
	trusted := foundationauth.NewContext(context.Background(), &foundationauth.Principal{
		SubjectKind: foundationauth.SubjectClient, ClientID: "identity-svc", Scopes: []string{foundationauth.PersonalPermissionsScope},
	})
	ordinary, err := controller.GetPersonalPermissions(trusted, &v1.PersonalPermissionsReq{UserKey: "user"})
	if err != nil {
		t.Fatal(err)
	}
	ordinaryKeys := permissionKeys(ordinary.Items)
	for _, key := range []string{
		"media.upload", "gallery.submission.create", "gallery.submission.read_own", "gallery.submission.withdraw",
		"gallery.favorite.read", "gallery.favorite.manage", "gallery.comment.create",
	} {
		if !slices.Contains(ordinaryKeys, key) {
			t.Errorf("ordinary directory missing %q: %v", key, ordinaryKeys)
		}
	}
	if slices.Contains(ordinaryKeys, string(galleryauthz.CapabilityImageRead)) {
		t.Fatalf("ordinary directory exposed operator permission: %v", ordinaryKeys)
	}
	admin, err := controller.GetPersonalPermissions(trusted, &v1.PersonalPermissionsReq{UserKey: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	adminKeys := permissionKeys(admin.Items)
	for _, key := range []string{
		string(galleryauthz.CapabilityImageRead), string(galleryauthz.CapabilitySubmissionReview),
		string(galleryauthz.CapabilityCollectionManage), string(galleryauthz.CapabilityClassificationGovern),
		string(galleryauthz.CapabilityCaseResolve), string(galleryauthz.CapabilityDiscoveryManage),
	} {
		if !slices.Contains(adminKeys, key) {
			t.Errorf("administrator directory missing %q: %v", key, adminKeys)
		}
	}
	for _, protected := range []string{string(galleryauthz.CapabilityAssetSettingsManage)} {
		if slices.Contains(adminKeys, protected) {
			t.Errorf("protected capability exposed: %q", protected)
		}
	}

	untrusted := foundationauth.NewContext(context.Background(), &foundationauth.Principal{SubjectKind: foundationauth.SubjectUser, Subject: "admin"})
	if _, err := controller.GetPersonalPermissions(untrusted, &v1.PersonalPermissionsReq{UserKey: "admin"}); err == nil {
		t.Fatal("untrusted directory request accepted")
	}
}

func TestPersonalMediaAuthorizationRequiresBothScopesAndCurrentRights(t *testing.T) {
	service, _ := galleryAuthorization(t)
	controller := NewPersonalPermissions("gallery-main-web", service)
	ctx := galleryPersonalContext(t, "user", "media.upload", string(galleryauthz.CapabilitySubmissionCreate))
	result, err := controller.AuthorizePersonalMedia(ctx, &v1.PersonalMediaAuthorizationReq{})
	if err != nil {
		t.Fatal(err)
	}
	want, _ := foundationauth.PersonalScope("gallery-main-web", "asset.profile.gallery-submission.upload")
	if result.UserKey != "user" || !slices.Equal(result.Scopes, []string{want}) {
		t.Fatalf("media authorization = %+v", result)
	}
	for _, denied := range []context.Context{
		galleryPersonalContext(t, "user", "media.upload"),
		galleryPersonalContext(t, "user", string(galleryauthz.CapabilitySubmissionCreate)),
	} {
		if _, err := controller.AuthorizePersonalMedia(denied, &v1.PersonalMediaAuthorizationReq{}); err == nil {
			t.Fatal("incomplete media scope accepted")
		}
	}
}

func TestPersonalRoutesAreExplicitAndCapabilityBound(t *testing.T) {
	for _, test := range []struct {
		method, path, capability string
	}{
		{"POST", "/api/v1/gallery/submissions", string(galleryauthz.CapabilitySubmissionCreate)},
		{"GET", "/api/v1/gallery/me/submissions", string(galleryauthz.CapabilitySubmissionReadOwn)},
		{"POST", "/api/v1/gallery/me/submissions/submission/withdraw", string(galleryauthz.CapabilitySubmissionWithdraw)},
		{"PUT", "/api/v1/gallery/me/favorites/image", string(galleryauthz.CapabilityFavoriteManage)},
		{"POST", "/api/v1/gallery/images/image/comments", string(galleryauthz.CapabilityCommentCreate)},
		{"GET", "/api/v1/gallery/admin/images", string(galleryauthz.CapabilityImageRead)},
		{"POST", "/api/v1/gallery/admin/submissions/submission/review", string(galleryauthz.CapabilitySubmissionReview)},
		{"POST", "/api/v1/gallery/admin/collections/collection/members", string(galleryauthz.CapabilityCollectionManage)},
		{"PATCH", "/api/v1/gallery/admin/site-settings", string(galleryauthz.CapabilityDiscoveryManage)},
		{"POST", "/api/v1/gallery/admin/classification/governance/preview", string(galleryauthz.CapabilityClassificationGovern)},
		{"POST", "/api/v1/gallery/admin/cases/case/resolve", string(galleryauthz.CapabilityCaseResolve)},
	} {
		if !allowsPersonalRoute(galleryPersonalContext(t, "admin", test.capability), test.method, test.path) {
			t.Errorf("selected capability denied: %s %s", test.method, test.path)
		}
	}
	all := galleryPersonalContext(t, "admin",
		"media.upload", string(galleryauthz.CapabilitySubmissionCreate), string(galleryauthz.CapabilitySubmissionReadOwn),
		string(galleryauthz.CapabilitySubmissionWithdraw), string(galleryauthz.CapabilityFavoriteRead),
		string(galleryauthz.CapabilityFavoriteManage), string(galleryauthz.CapabilityCommentCreate),
		string(galleryauthz.CapabilityImageRead), string(galleryauthz.CapabilityImageUpdate),
		string(galleryauthz.CapabilityImageHide), string(galleryauthz.CapabilitySubmissionReview),
		string(galleryauthz.CapabilityCollectionManage),
	)
	for _, path := range []string{
		"/api/v1/gallery/guest-claims", "/api/v1/gallery/me", "/api/v1/internal/personal-token/permissions",
		"/api/v1/authorization/manage/console", "/api/v1/gallery/admin/assets",
		"/api/v1/gallery/me/submissions//withdraw", "/api/v1/gallery/admin/images/../hide",
	} {
		for _, method := range []string{"GET", "POST", "PATCH", "DELETE", "PUT"} {
			if allowsPersonalRoute(all, method, path) {
				t.Errorf("unexpected route allowed: %s %s", method, path)
			}
		}
	}
}

func permissionKeys(items []foundationauth.PersonalPermission) []string {
	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, item.Key)
	}
	return keys
}
