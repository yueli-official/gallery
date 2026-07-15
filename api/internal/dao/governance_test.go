package dao

import (
	"strings"
	"testing"

	"platform/gokit/classification"
)

func TestClassificationGovernanceImpactTokenIsCanonicalAndSensitive(t *testing.T) {
	requests := []classification.ImpactRequest{
		{Kind: classification.GovernCategory, ID: "category-1"},
		{Kind: classification.GovernTag, ID: "tag-1"},
	}
	left := []classification.ReferenceImpact{
		{Kind: classification.GovernTag, ID: "tag-1", Exists: true, Status: classification.StatusActive, AliasCount: 2},
		{Kind: classification.GovernCategory, ID: "category-1", Exists: true, Status: classification.StatusActive, DirectChildIDs: []string{"child-b", "child-a"}, AssignmentCount: 4},
	}
	right := []classification.ReferenceImpact{
		{Kind: classification.GovernCategory, ID: "category-1", Exists: true, Status: classification.StatusActive, DirectChildIDs: []string{"child-a", "child-b"}, AssignmentCount: 4},
		{Kind: classification.GovernTag, ID: "tag-1", Exists: true, Status: classification.StatusActive, AliasCount: 2},
	}
	leftToken := classificationGovernanceImpactToken(requests, left)
	rightToken := classificationGovernanceImpactToken(requests, right)
	if leftToken == "" || leftToken != rightToken {
		t.Fatalf("canonical tokens differ: %q != %q", leftToken, rightToken)
	}
	right[0].AssignmentCount++
	if changed := classificationGovernanceImpactToken(requests, right); changed == leftToken {
		t.Fatal("changing a requested impact must change the token")
	}
}

func TestClassificationStatusUpdateRecordsFirstActivationForStructuredIdentities(t *testing.T) {
	query, args, err := classificationStatusUpdateStatement(classification.GovernStep{
		IdentityKind: classification.GovernCategory,
		SourceID:     "category-1",
		Status:       classification.StatusActive,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(query, "first_activated_at = COALESCE(first_activated_at, NOW())") {
		t.Fatalf("structured identity update does not record first activation: %q", query)
	}
	if len(args) != 2 || args[0] != classification.StatusActive || args[1] != "category-1" {
		t.Fatalf("args = %#v", args)
	}

	tagQuery, tagArgs, err := classificationStatusUpdateStatement(classification.GovernStep{
		IdentityKind: classification.GovernTag,
		SourceID:     "tag-1",
		Status:       classification.StatusInactive,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(tagQuery, "first_activated_at") {
		t.Fatalf("tag update references a field tags do not own: %q", tagQuery)
	}
	if len(tagArgs) != 2 || tagArgs[0] != classification.StatusInactive || tagArgs[1] != "tag-1" {
		t.Fatalf("tag args = %#v", tagArgs)
	}
}
