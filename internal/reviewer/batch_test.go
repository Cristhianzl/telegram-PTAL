package reviewer

import (
	"strings"
	"testing"
)

func refs(specs ...*PullRequestRef) []*PullRequestRef { return specs }

func pr(number int, author string, draft bool) *PullRequestRef {
	return &PullRequestRef{Number: number, Author: author, IsDraft: draft, Title: "t"}
}

// Drafts are work in progress; reviewing one is usually noise for its author.
func TestDraftsAreSkippedByDefault(t *testing.T) {
	queue, skipped := planBatch(refs(pr(1, "alice", false), pr(2, "bob", true)), BatchOptions{})

	if len(queue) != 1 || queue[0].Number != 1 {
		t.Errorf("queue = %v, want only the non-draft", numbers(queue))
	}
	if skipped[2] != "draft" {
		t.Errorf("the draft should be skipped with a reason, got %q", skipped[2])
	}
}

func TestDraftsCanBeIncluded(t *testing.T) {
	queue, _ := planBatch(refs(pr(1, "alice", true)), BatchOptions{IncludeDrafts: true})

	if len(queue) != 1 {
		t.Error("IncludeDrafts should let a draft through")
	}
}

// A batch spends minutes and subscription budget per pull request, so an
// unbounded run against a busy repository would occupy the machine for hours.
func TestLimitCapsTheRun(t *testing.T) {
	var all []*PullRequestRef
	for i := 1; i <= 10; i++ {
		all = append(all, pr(i, "alice", false))
	}

	queue, skipped := planBatch(all, BatchOptions{Limit: 3})

	if len(queue) != 3 {
		t.Errorf("queued %d, want the limit of 3", len(queue))
	}
	if len(skipped) != 7 {
		t.Errorf("skipped %d, want the remaining 7", len(skipped))
	}
	// The reason has to name the limit, or it reads as an unexplained gap.
	for _, why := range skipped {
		if !strings.Contains(why, "limit") {
			t.Errorf("skip reason should mention the limit: %q", why)
		}
	}
}

func TestUnsetLimitUsesTheDefault(t *testing.T) {
	var all []*PullRequestRef
	for i := 1; i <= DefaultBatchLimit+5; i++ {
		all = append(all, pr(i, "alice", false))
	}

	queue, _ := planBatch(all, BatchOptions{})

	if len(queue) != DefaultBatchLimit {
		t.Errorf("queued %d, want the default limit of %d", len(queue), DefaultBatchLimit)
	}
}

func TestSkipAuthorsKeepsBotsOut(t *testing.T) {
	queue, skipped := planBatch(
		refs(pr(1, "alice", false), pr(2, "app/dependabot", false)),
		BatchOptions{SkipAuthors: []string{"app/dependabot"}},
	)

	if len(queue) != 1 || queue[0].Number != 1 {
		t.Errorf("queue = %v", numbers(queue))
	}
	if skipped[2] == "" {
		t.Error("the skipped author needs a reason")
	}
}

// The summary is what someone reads on their phone; it must account for every
// pull request, or a silent drop looks like the batch lost one.
func TestSummaryAccountsForEveryOutcome(t *testing.T) {
	b := &BatchResult{
		Repo: "acme/api",
		Reviewed: []*Result{
			{Verdict: VerdictApproveAndMerge, Merged: true},
			{Verdict: VerdictApprove},
			{Verdict: VerdictRequestChanges},
			{Verdict: VerdictComment},
		},
		Skipped: map[int]string{9: "draft"},
		Failed:  map[int]string{8: "timed out"},
	}

	approved, merged, changes, commented := b.Counts()
	if approved != 2 || merged != 1 || changes != 1 || commented != 1 {
		t.Errorf("counts = approved:%d merged:%d changes:%d commented:%d", approved, merged, changes, commented)
	}

	s := b.Summary()
	for _, want := range []string{"merged", "approved", "changes requested", "commented", "skipped", "failed"} {
		if !strings.Contains(s, want) {
			t.Errorf("summary omits %q:\n%s", want, s)
		}
	}
}

// A merged pull request must not also be counted as merely approved, or the
// totals add up to more than were reviewed.
func TestMergedIsNotDoubleCounted(t *testing.T) {
	b := &BatchResult{Reviewed: []*Result{{Verdict: VerdictApproveAndMerge, Merged: true}}}

	if !strings.Contains(b.Summary(), "merged:") {
		t.Error("a merge should be reported")
	}
	if strings.Contains(b.Summary(), "approved:") {
		t.Errorf("a merged pull request should not also be listed as approved:\n%s", b.Summary())
	}
}

func numbers(prs []*PullRequestRef) []int {
	out := make([]int, len(prs))
	for i, p := range prs {
		out[i] = p.Number
	}
	return out
}
