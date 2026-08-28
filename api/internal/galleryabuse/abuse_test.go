package galleryabuse

import (
	"context"
	"testing"

	"github.com/yueli-official/foundation/go/abuse"
)

func TestGuestAndMemberSubmissionBudgetsAreIndependent(t *testing.T) {
	module, err := abuse.NewMemory(
		abuse.MustCompile(Definition(Policy{})),
		abuse.MemoryOptions{Secret: []byte("gallery-abuse-test-secret-at-least-32-bytes")},
	)
	if err != nil {
		t.Fatal(err)
	}
	actions, err := Bind(module)
	if err != nil {
		t.Fatal(err)
	}
	network, err := NetworkPrefix("192.0.2.30")
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 6; index++ {
		admission, err := actions.Guest.Admit(context.Background(), abuse.Input{
			ID:      abuse.AttemptID("guest-" + string(rune('a'+index))),
			Signals: abuse.Signals{Network: network, Actor: "guest:visitor-1"},
		})
		if err != nil {
			t.Fatal(err)
		}
		want := abuse.DispositionAllow
		if index == 5 {
			want = abuse.DispositionReject
		}
		if admission.Disposition != want {
			t.Fatalf("guest attempt %d: got %q, want %q", index+1, admission.Disposition, want)
		}
	}
	member, err := actions.Member.Admit(context.Background(), abuse.Input{
		ID:      "member-a",
		Signals: abuse.Signals{Network: network, Actor: "user:visitor-1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if member.Disposition != abuse.DispositionAllow {
		t.Fatalf("member action must have an independent budget, got %q", member.Disposition)
	}
}

func TestDefinitionIncludesIndependentCommentBudgets(t *testing.T) {
	catalog := abuse.MustCompile(Definition(Policy{}))
	module, err := abuse.NewMemory(catalog, abuse.MemoryOptions{Secret: []byte("gallery-comment-test-secret-at-least-32-bytes")})
	if err != nil {
		t.Fatal(err)
	}
	actions, err := Bind(module)
	if err != nil {
		t.Fatal(err)
	}
	if actions.GuestComment == nil || actions.MemberComment == nil {
		t.Fatal("comment abuse actions are missing")
	}
}
