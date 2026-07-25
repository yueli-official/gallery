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
	} {
		if !slices.Contains(capabilities, required) {
			t.Fatalf("content operator is missing %q", required)
		}
	}
}
