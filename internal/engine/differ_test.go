package engine

import (
	"testing"
	"time"

	"github.com/cristhianzl/pullalerts/internal/githubapi"
)

func pr(id string, buckets ...githubapi.Bucket) *githubapi.PullRequest {
	return &githubapi.PullRequest{
		ID: id, Repo: "acme/app", Number: 1, Title: "Fix thing",
		URL: "https://github.com/acme/app/pull/1", Buckets: buckets,
	}
}

func snapshot(prs ...*githubapi.PullRequest) *githubapi.Snapshot {
	s := &githubapi.Snapshot{PRs: map[string]*githubapi.PullRequest{}, FetchedAt: time.Now()}
	for _, p := range prs {
		s.PRs[p.ID] = p
	}
	return s
}

func kinds(events []Event) map[Kind]int {
	out := map[Kind]int{}
	for _, e := range events {
		out[e.Kind]++
	}
	return out
}

// The first run must be silent, or the user receives dozens of messages at
// minute zero and uninstalls.
func TestFirstRunEmitsNothing(t *testing.T) {
	snap := snapshot(pr("a", githubapi.BucketReviewRequested), pr("b", githubapi.BucketAuthored))

	events, next := Diff(nil, snap, Options{FirstRun: true})

	if len(events) != 0 {
		t.Fatalf("first run should be silent, produced %d events", len(events))
	}
	if len(next) != 2 {
		t.Fatalf("state should record 2 pull requests, recorded %d", len(next))
	}
}

func TestNewReviewRequestIsUrgent(t *testing.T) {
	prev := map[string]*Tracked{}
	snap := snapshot(pr("a", githubapi.BucketReviewRequested))

	events, _ := Diff(prev, snap, Options{})

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Kind != KindReviewRequested {
		t.Errorf("kind = %s, want %s", events[0].Kind, KindReviewRequested)
	}
	if events[0].Urgency != UrgencyNow {
		t.Error("a review request should interrupt immediately")
	}
}

// A pull request you opened yourself must not raise an entry alert: you
// just opened it.
func TestOwnPullRequestDoesNotAlertOnEntry(t *testing.T) {
	events, _ := Diff(map[string]*Tracked{}, snapshot(pr("a", githubapi.BucketAuthored)), Options{})

	if len(events) != 0 {
		t.Fatalf("your own pull request should not alert on entry, produced %v", kinds(events))
	}
}

func TestDraftSuppressesReviewRequest(t *testing.T) {
	p := pr("a", githubapi.BucketReviewRequested)
	p.IsDraft = true

	events, _ := Diff(map[string]*Tracked{}, snapshot(p), Options{IgnoreDrafts: true})

	if len(events) != 0 {
		t.Fatalf("a draft should not request review, produced %v", kinds(events))
	}
}

func TestStateTransitionsOnOwnPR(t *testing.T) {
	old := pr("a", githubapi.BucketAuthored)
	old.Checks = "SUCCESS"
	old.ReviewDecision = "REVIEW_REQUIRED"
	old.Mergeable = "MERGEABLE"
	prev := map[string]*Tracked{"a": {PullRequest: *old}}

	now := pr("a", githubapi.BucketAuthored)
	now.Checks = "FAILURE"
	now.ReviewDecision = "CHANGES_REQUESTED"
	now.Mergeable = "CONFLICTING"

	events, _ := Diff(prev, snapshot(now), Options{})
	got := kinds(events)

	for _, want := range []Kind{KindChecksFailed, KindChangesRequested, KindConflict} {
		if got[want] != 1 {
			t.Errorf("missing event %s (got %v)", want, got)
		}
	}
}

// GitHub's search index is eventually consistent: a single absence cannot be
// treated as "pull request closed".
func TestGoneRequiresTwoConsecutiveAbsences(t *testing.T) {
	prev := map[string]*Tracked{"a": {PullRequest: *pr("a", githubapi.BucketAuthored)}}

	events, next := Diff(prev, snapshot(), Options{})
	if len(events) != 0 {
		t.Fatalf("a single absence should emit nothing, got %v", kinds(events))
	}
	if _, kept := next["a"]; !kept {
		t.Fatal("the pull request should stay in state awaiting confirmation")
	}

	events, next = Diff(next, snapshot(), Options{})
	if kinds(events)[KindGone] != 1 {
		t.Errorf("the second absence should emit %s, got %v", KindGone, kinds(events))
	}
	if _, kept := next["a"]; kept {
		t.Error("the pull request should leave state once the absence is confirmed")
	}
}

func TestNewCommentsDetected(t *testing.T) {
	old := pr("a", githubapi.BucketAuthored)
	old.Comments = 2
	prev := map[string]*Tracked{"a": {PullRequest: *old}}

	now := pr("a", githubapi.BucketAuthored)
	now.Comments = 5

	events, _ := Diff(prev, snapshot(now), Options{})

	if kinds(events)[KindNewActivity] != 1 {
		t.Fatalf("expected new activity, got %v", kinds(events))
	}
	if events[0].Detail != "3 new comments" {
		t.Errorf("detail = %q", events[0].Detail)
	}
}

// Switching from public search to GraphQL fills fields like reviewDecision
// and statusCheckRollup all at once. That is the first reading of the field,
// not a change - treating it as one would fire dozens of false alerts at the
// exact moment the mode switches.
func TestFirstObservationOfAFieldIsNotAChange(t *testing.T) {
	// State from public mode: no CI, no decision, no mergeable.
	old := pr("a", githubapi.BucketAuthored)
	prev := map[string]*Tracked{"a": {PullRequest: *old}}

	// The same pull request read through GraphQL, every field filled.
	now := pr("a", githubapi.BucketAuthored)
	now.ReviewDecision = "APPROVED"
	now.Checks = "FAILURE"
	now.Mergeable = "CONFLICTING"

	events, _ := Diff(prev, snapshot(now), Options{})

	if len(events) != 0 {
		t.Fatalf("the first reading of a field should not alert, produced %v", kinds(events))
	}
}

// Once the field is known, the next change alerts normally.
func TestChangeAfterFirstObservationStillAlerts(t *testing.T) {
	old := pr("a", githubapi.BucketAuthored)
	old.Checks = "SUCCESS"
	prev := map[string]*Tracked{"a": {PullRequest: *old}}

	now := pr("a", githubapi.BucketAuthored)
	now.Checks = "FAILURE"

	events, _ := Diff(prev, snapshot(now), Options{})

	if kinds(events)[KindChecksFailed] != 1 {
		t.Errorf("a real change should alert, got %v", kinds(events))
	}
}

// Switching data source re-syncs silently. Public search carries no review
// count and no CI state; moving to GraphQL fills them all at once, and each
// jump would look like a real change.
func TestModeChangeResyncsSilently(t *testing.T) {
	old := pr("a", githubapi.BucketAuthored)
	old.Comments = 2 // public search carried comments, but reviews = 0
	prev := map[string]*Tracked{"a": {PullRequest: *old}}

	now := pr("a", githubapi.BucketAuthored)
	now.Comments = 2
	now.Reviews = 47 // GraphQL fills the field for the first time
	now.Checks = "FAILURE"
	now.ReviewDecision = "CHANGES_REQUESTED"

	events, next := Diff(prev, snapshot(now), Options{ModeChanged: true})

	if len(events) != 0 {
		t.Fatalf("switching source should not alert, produced %v", kinds(events))
	}
	// State must absorb the new values, or the next cycle would repeat the
	// exact same jump.
	if next["a"].Reviews != 47 || next["a"].Checks != "FAILURE" {
		t.Error("state should have absorbed the new fields")
	}
}
