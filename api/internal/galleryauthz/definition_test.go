package galleryauthz

import (
	"slices"
	"testing"

	"github.com/yueli-official/foundation/go/authorization"
)

func TestDefinitionCompiles(t *testing.T) {
	catalog, err := authorization.Compile(Definition())
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	if catalog.Consumer() != "gallery" {
		t.Fatalf("consumer = %q", catalog.Consumer())
	}
	if catalog.Version() != 4 {
		t.Fatalf("version = %d, want 4", catalog.Version())
	}
}

func TestRegistrationContentOperatorStartsDisabled(t *testing.T) {
	definition := Definition()
	if len(definition.Automatic) != 1 {
		t.Fatalf("automatic rules = %d, want 1", len(definition.Automatic))
	}
	if definition.Automatic[0].Enabled {
		t.Fatal("registration content operator rule must be disabled by default")
	}
}

func TestContentOperatorExcludesProtectedGovernanceCapabilities(t *testing.T) {
	definition := Definition()
	index := slices.IndexFunc(definition.Roles, func(role authorization.RoleDefinition) bool {
		return role.Key == RoleContentOperator
	})
	if index < 0 {
		t.Fatal("content operator role is missing")
	}
	capabilities := definition.Roles[index].Capabilities
	for _, forbidden := range []authorization.CapabilityKey{
		authorization.CapabilityManage,
		CapabilityClassificationGovern,
		CapabilityCaseResolve,
		CapabilityDiscoveryManage,
		CapabilityAssetSettingsManage,
	} {
		if slices.Contains(capabilities, forbidden) {
			t.Fatalf("content operator unexpectedly owns %q", forbidden)
		}
	}
	for _, required := range []authorization.CapabilityKey{
		CapabilityImageUpdate,
		CapabilitySubmissionReview,
		CapabilityCollectionManage,
		CapabilityClassificationProposalReview,
		CapabilityCaseRead,
		CapabilityCommentRead,
		CapabilityCommentModerate,
		CapabilityCommentDelete,
	} {
		if !slices.Contains(capabilities, required) {
			t.Fatalf("content operator is missing %q", required)
		}
	}
}

func TestDefinitionDeclaresProtectedDiscoveryManagement(t *testing.T) {
	definition := Definition()
	index := slices.IndexFunc(definition.Capabilities, func(capability authorization.CapabilityDefinition) bool {
		return capability.Key == CapabilityDiscoveryManage
	})
	if index < 0 {
		t.Fatalf("definition is missing %q", CapabilityDiscoveryManage)
	}
	capability := definition.Capabilities[index]
	if capability.Binding != authorization.BindingProtectedOnly {
		t.Fatalf("%q binding = %q, want %q", CapabilityDiscoveryManage, capability.Binding, authorization.BindingProtectedOnly)
	}
	if capability.Risk != authorization.RiskHigh {
		t.Fatalf("%q risk = %q, want %q", CapabilityDiscoveryManage, capability.Risk, authorization.RiskHigh)
	}

	administratorIndex := slices.IndexFunc(definition.Roles, func(role authorization.RoleDefinition) bool {
		return role.Key == RoleAdministrator
	})
	if administratorIndex < 0 {
		t.Fatal("administrator role is missing")
	}
	if !slices.Contains(definition.Roles[administratorIndex].Capabilities, CapabilityDiscoveryManage) {
		t.Fatalf("administrator role is missing %q", CapabilityDiscoveryManage)
	}
}
