package gallery

import "testing"

func TestPublicIDIsACompactRepresentationOfTheSameUUID(t *testing.T) {
	canonical := "019817c8-0000-7000-8000-000000000001"
	compact := PublicID(canonical)
	if len(compact) != 22 {
		t.Fatalf("expected 22 characters, got %q", compact)
	}
	roundTrip, err := DatabaseID(compact)
	if err != nil || roundTrip != canonical {
		t.Fatalf("compact ID must round trip: id=%q err=%v", roundTrip, err)
	}
}
