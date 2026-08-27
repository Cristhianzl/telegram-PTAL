package telegram

import (
	"strings"
	"testing"

	"github.com/cristhianzl/pullalerts/internal/engine"
	"github.com/cristhianzl/pullalerts/internal/githubapi"
)

func samplePR() *githubapi.PullRequest {
	return &githubapi.PullRequest{
		ID: "x", Repo: "acme/app", Number: 42,
		Title: "Fix <script> & \"quotes\" in parser",
		URL:   "https://github.com/acme/app/pull/42",
		Author: "bob", Additions: 120, Deletions: 8, Checks: "FAILURE",
		Buckets: []githubapi.Bucket{githubapi.BucketReviewRequested},
	}
}

// Pull request titles frequently contain <, > and &. Without escaping,
// Telegram rejects the whole message and the alert simply never arrives.
func TestTitlesWithHTMLAreEscaped(t *testing.T) {
	text := RenderEvents([]engine.Event{{
		Kind: engine.KindReviewRequested, PR: samplePR(), Urgency: engine.UrgencyNow,
	}})

	if strings.Contains(text, "<script>") {
		t.Error("the title was not escaped - the message would be rejected")
	}
	if !strings.Contains(text, "&lt;script&gt;") {
		t.Errorf("expected the title escaped, got: %s", text)
	}
	if !strings.Contains(text, "&quot;") && !strings.Contains(text, "&#34;") {
		t.Error("the quotes should have been escaped")
	}
}

func TestSingleEventShowsDetail(t *testing.T) {
	text := RenderEvents([]engine.Event{{
		Kind: engine.KindChecksFailed, PR: samplePR(), Urgency: engine.UrgencyNow,
	}})

	for _, want := range []string{"CI broke", "acme/app#42", "+120 −8", "@bob"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in the message:\n%s", want, text)
		}
	}
}

// Five approvals must become one block, not five notices.
func TestSameKindIsGrouped(t *testing.T) {
	var events []engine.Event
	for i := range 5 {
		pr := samplePR()
		pr.ID = string(rune('a' + i))
		pr.Number = 100 + i
		events = append(events, engine.Event{
			Kind: engine.KindApproved, PR: pr, Urgency: engine.UrgencyBatch,
		})
	}

	text := RenderEvents(events)

	if strings.Count(text, "Pull request approved") != 1 {
		t.Errorf("the title should appear exactly once:\n%s", text)
	}
	if !strings.Contains(text, "· 5") {
		t.Errorf("missing the group count:\n%s", text)
	}
	for i := range 5 {
		if !strings.Contains(text, "acme/app#"+string(rune('1'))+"0"+string(rune('0'+i))) {
			t.Errorf("missing pull request %d:\n%s", 100+i, text)
		}
	}
}

// Telegram rejects messages above 4096 characters. The panel and the daily
// digest cross that limit sooner than it seems.
func TestTruncateRespectsTelegramLimit(t *testing.T) {
	long := strings.Repeat("some line of text\n", 500)

	out := Truncate(long)

	if len(out) > MaxMessageLen {
		t.Errorf("a %d byte message exceeds the %d limit", len(out), MaxMessageLen)
	}
	if !strings.Contains(out, "truncated") {
		t.Error("the cut should be signalled to the reader")
	}
}

func TestTruncateLeavesShortMessagesAlone(t *testing.T) {
	if got := Truncate("short"); got != "short" {
		t.Errorf("a short message was altered: %q", got)
	}
}

func TestPanelListsEveryBucket(t *testing.T) {
	snap := &githubapi.Snapshot{PRs: map[string]*githubapi.PullRequest{}}
	review := samplePR()
	snap.PRs["a"] = review

	mine := samplePR()
	mine.ID, mine.Number = "b", 7
	mine.Buckets = []githubapi.Bucket{githubapi.BucketAuthored}
	snap.PRs["b"] = mine

	text := RenderPanel(snap)

	if !strings.Contains(text, "Waiting for your review") {
		t.Error("missing the review bucket")
	}
	if !strings.Contains(text, "Your open pull requests") {
		t.Error("missing the authored bucket")
	}
}

func TestEmptyPanelSaysSo(t *testing.T) {
	text := RenderPanel(&githubapi.Snapshot{PRs: map[string]*githubapi.PullRequest{}})

	if !strings.Contains(text, "Nothing open") {
		t.Errorf("the empty panel should say so plainly:\n%s", text)
	}
}
