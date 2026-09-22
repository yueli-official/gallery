package galleryauthz

import (
	"context"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
)

type Runtime interface {
	authorization.Authorizer
	authorization.AccessReader
	authorization.RoleManager
	authorization.RoleReader
	authorization.GrantManager
	authorization.GrantReader
	authorization.WorkflowManager
	authorization.WorkflowReader
	authorization.PolicyManager
	authorization.PolicyReader
	authorization.Reconciler
}

type Service struct {
	runtime Runtime
}

func New(runtime Runtime) *Service {
	return &Service{runtime: runtime}
}

func (service *Service) Runtime() Runtime {
	if service == nil {
		return nil
	}
	return service.runtime
}

func (service *Service) Subject(ctx context.Context) authorization.SubjectRef {
	principal, ok := foundationauth.FromContext(ctx)
	if !ok {
		return authorization.SubjectRef{Kind: authorization.SubjectAnonymous}
	}
	if principal.SubjectKind == foundationauth.SubjectUser && principal.Subject != "" {
		return authorization.SubjectRef{Kind: authorization.SubjectUser, ID: principal.Subject}
	}
	if principal.SubjectKind == foundationauth.SubjectClient && principal.ClientID != "" {
		return authorization.SubjectRef{Kind: authorization.SubjectService, ID: principal.ClientID}
	}
	subjectKind, _ := principal.Claim("subject_kind")
	if subjectKind == "user" && principal.Subject != "" {
		return authorization.SubjectRef{Kind: authorization.SubjectUser, ID: principal.Subject}
	}
	if subjectKind == "client" && principal.ClientID != "" {
		return authorization.SubjectRef{Kind: authorization.SubjectService, ID: principal.ClientID}
	}
	return authorization.SubjectRef{Kind: authorization.SubjectAnonymous}
}

func (service *Service) ReconcileSubject(ctx context.Context) error {
	if service == nil || service.runtime == nil {
		return unavailable("runtime")
	}
	subject := service.Subject(ctx)
	if subject.Kind != authorization.SubjectUser || subject.ID == "" {
		return nil
	}
	preview, err := service.runtime.PreviewReconcileSubject(ctx, authorization.ReconcileSubjectCommand{
		Subject: subject,
	})
	if err != nil || preview.Created == 0 {
		return err
	}
	_, err = service.runtime.ReconcileSubject(ctx, authorization.ReconcileSubjectCommand{Subject: subject})
	return err
}

func (service *Service) Decide(
	ctx context.Context,
	capability authorization.CapabilityKey,
) (authorization.Decision, error) {
	if service == nil || service.runtime == nil {
		return authorization.Decision{}, unavailable("runtime")
	}
	if principal, ok := foundationauth.FromContext(ctx); ok && principal.IsPersonalToken() &&
		!foundationauth.AllowsPersonalCapability(ctx, string(capability)) {
		return authorization.Decision{Allowed: false}, nil
	}
	if err := service.ReconcileSubject(ctx); err != nil {
		return authorization.Decision{}, err
	}
	return service.runtime.Decide(ctx, authorization.DecisionRequest{
		Subject: service.Subject(ctx), Capability: capability, ScopeID: RootScopeID,
	})
}

func (service *Service) CheckCapability(ctx context.Context, capability authorization.CapabilityKey) (bool, error) {
	decision, err := service.Decide(ctx, capability)
	return decision.Allowed, err
}

func (service *Service) EffectiveAccess(ctx context.Context) (authorization.EffectiveAccess, error) {
	if service == nil || service.runtime == nil {
		return authorization.EffectiveAccess{}, unavailable("runtime")
	}
	if err := service.ReconcileSubject(ctx); err != nil {
		return authorization.EffectiveAccess{}, err
	}
	return service.runtime.EffectiveAccess(ctx, authorization.EffectiveAccessQuery{
		Subject: service.Subject(ctx), ScopeID: RootScopeID,
	})
}

func (service *Service) IsAdministrator(ctx context.Context) bool {
	decision, err := service.Decide(ctx, authorization.CapabilityManage)
	return err == nil && decision.Allowed
}

func unavailable(field string) *authorization.Error {
	return &authorization.Error{
		Kind: authorization.ErrorUnavailable, Field: field, Message: "is not configured",
	}
}
