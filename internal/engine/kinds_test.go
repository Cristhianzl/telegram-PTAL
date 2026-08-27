package engine

import (
	"strings"
	"testing"
)

// A typo in a mute list must be an error, not a silent no-op. Otherwise
// someone believes they turned an alert off and only finds out they did not
// when it interrupts them.
func TestUnknownEventNameIsRejected(t *testing.T) {
	_, err := ParseKinds("checks_failed,ci_broke")

	if err == nil {
		t.Fatal("an unknown event name should be rejected")
	}
	if !strings.Contains(err.Error(), "ci_broke") {
		t.Errorf("the error should name the offending value: %v", err)
	}
	// Listing the valid names is what makes the error actionable.
	if !strings.Contains(err.Error(), "review_requested") {
		t.Errorf("the error should list the valid events: %v", err)
	}
}

func TestParseKindsAcceptsTheShapesPeopleType(t *testing.T) {
	kinds, err := ParseKinds(" Checks-Failed , review_requested ,, ")
	if err != nil {
		t.Fatalf("should tolerate spacing, case, hyphens and empty entries: %v", err)
	}
	if len(kinds) != 2 {
		t.Fatalf("parsed %d kinds, want 2: %v", len(kinds), kinds)
	}
	if kinds[0] != KindChecksFailed || kinds[1] != KindReviewRequested {
		t.Errorf("parsed the wrong kinds: %v", kinds)
	}
}

func TestEmptyFilterAllowsEverything(t *testing.T) {
	f := NewKindFilter(nil, nil)

	for _, k := range AllKinds {
		if !f.Allows(k) {
			t.Errorf("%s should be allowed by an empty filter", k)
		}
	}
	if muted := f.Muted(); len(muted) != 0 {
		t.Errorf("an empty filter mutes nothing, reported %v", muted)
	}
}

func TestMuteRemovesOnlyTheNamedKinds(t *testing.T) {
	f := NewKindFilter(nil, []Kind{KindChecksFailed, KindChecksFixed})

	if f.Allows(KindChecksFailed) || f.Allows(KindChecksFixed) {
		t.Error("muted kinds must not be allowed")
	}
	if !f.Allows(KindReviewRequested) {
		t.Error("muting CI events must not affect review requests")
	}
	if len(f.Muted()) != 2 {
		t.Errorf("Muted() = %v, want exactly the two CI kinds", f.Muted())
	}
}

func TestAllowListRestrictsToTheNamedKinds(t *testing.T) {
	f := NewKindFilter([]Kind{KindReviewRequested, KindMentioned}, nil)

	if !f.Allows(KindReviewRequested) || !f.Allows(KindMentioned) {
		t.Error("kinds on the allow list must pass")
	}
	for _, k := range []Kind{KindApproved, KindChecksFailed, KindGone} {
		if f.Allows(k) {
			t.Errorf("%s is not on the allow list and must not pass", k)
		}
	}
}

// The two lists compose in one order only, and it has to be the intuitive
// one: naming a kind in both means it stays off.
func TestMuteWinsOverAllow(t *testing.T) {
	f := NewKindFilter(
		[]Kind{KindReviewRequested, KindChecksFailed},
		[]Kind{KindChecksFailed},
	)

	if f.Allows(KindChecksFailed) {
		t.Error("a kind on both lists must stay muted")
	}
	if !f.Allows(KindReviewRequested) {
		t.Error("the rest of the allow list must still pass")
	}
}

// Every event the engine can emit needs an entry, or it becomes impossible to
// turn off — the filter would have no name to match against.
func TestEveryEmittableKindIsListed(t *testing.T) {
	listed := map[Kind]bool{}
	for _, k := range AllKinds {
		listed[k] = true
	}

	emittable := []Kind{
		KindReviewRequested, KindChangesRequested, KindApproved,
		KindChecksFailed, KindChecksFixed, KindConflict, KindNewActivity,
		KindReadyForReview, KindAssigned, KindMentioned, KindGone,
	}
	for _, k := range emittable {
		if !listed[k] {
			t.Errorf("%s can be emitted but is missing from AllKinds, so it cannot be muted", k)
		}
		if k.Summary() == string(k) {
			t.Errorf("%s has no human-readable summary", k)
		}
	}
}
