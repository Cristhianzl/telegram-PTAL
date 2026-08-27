package reviewer

import (
	"fmt"
	"strings"
	"testing"
)

// The verdict line is what decides whether a pull request gets approved or
// blocked, so parsing it has to be exact.
func TestExtractVerdict(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want Verdict
	}{
		{"approve", "Looks good.\n\nVERDICT: approve", VerdictApprove},
		{"request changes", "Problems.\n\nVERDICT: request_changes", VerdictRequestChanges},
		{"hyphenated", "Problems.\n\nVERDICT: request-changes", VerdictRequestChanges},
		{"spaced", "Problems.\n\nVERDICT: request changes", VerdictRequestChanges},
		{"comment", "Notes.\n\nVERDICT: comment", VerdictComment},
		{"lowercase key", "Notes.\n\nverdict: approve", VerdictApprove},
		{"trailing space", "Notes.\n\nVERDICT: approve   ", VerdictApprove},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, body := extractVerdict(c.in)
			if got != c.want {
				t.Errorf("verdict = %q, want %q", got, c.want)
			}
			if strings.Contains(body, "VERDICT") || strings.Contains(body, "verdict:") {
				t.Errorf("the verdict line should be stripped from the body: %q", body)
			}
		})
	}
}

// A review with no verdict must not approve or block anything. Defaulting to
// the neutral outcome means a parsing slip can never merge or gate code on
// its own.
func TestMissingVerdictDefaultsToComment(t *testing.T) {
	got, body := extractVerdict("Some findings, but the model forgot the line.")

	if got != VerdictComment {
		t.Errorf("verdict = %q, want %q for a missing line", got, VerdictComment)
	}
	if body == "" {
		t.Error("the review body must survive a missing verdict")
	}
}

// A verdict-shaped phrase inside prose must not be mistaken for the real
// line, which is anchored to its own line.
func TestVerdictInsideProseIsIgnored(t *testing.T) {
	got, _ := extractVerdict("The author wrote VERDICT: approve in the description.\n\nVERDICT: request_changes")

	if got != VerdictRequestChanges {
		t.Errorf("verdict = %q, want the real trailing line", got)
	}
}

func TestVerdictMapsToTheRightGitHubAction(t *testing.T) {
	cases := map[Verdict]string{
		VerdictApprove:        "--approve",
		VerdictRequestChanges: "--request-changes",
		VerdictComment:        "--comment",
	}
	for verdict, want := range cases {
		got, err := reviewAction(verdict)
		if err != nil {
			t.Errorf("%s: %v", verdict, err)
		}
		if got != want {
			t.Errorf("%s -> %q, want %q", verdict, got, want)
		}
	}
	if _, err := reviewAction("nonsense"); err == nil {
		t.Error("an unknown verdict should be an error, not a silent approval")
	}
}

// The tool list is the security boundary of this feature: it is what stops a
// pull request from a stranger executing anything on the machine.
func TestClaudeGetsOnlyReadOnlyTools(t *testing.T) {
	joined := strings.Join(readOnlyTools, ",")

	for _, forbidden := range []string{"Write", "Edit", "NotebookEdit", "WebFetch"} {
		if strings.Contains(joined, forbidden) {
			t.Errorf("%s must not be available to a review", forbidden)
		}
	}
	// A bare Bash entry would allow any command at all.
	for _, tool := range readOnlyTools {
		if tool == "Bash" || strings.HasPrefix(tool, "Bash(*") {
			t.Errorf("unrestricted Bash defeats the tool restriction: %q", tool)
		}
	}
	// And nothing that mutates GitHub, which ptal does itself.
	for _, mutating := range []string{"gh pr review", "gh pr comment", "gh pr merge"} {
		if strings.Contains(joined, mutating) {
			t.Errorf("%q must stay out of the review's tools: ptal publishes, not Claude", mutating)
		}
	}
}

// The CLI prints warnings before its JSON - an untrusted-workspace notice,
// for one - so the parser must find the object rather than assume it starts
// at byte zero.
func TestParseClaudeOutputSkipsLeadingNoise(t *testing.T) {
	out := []byte("Ignoring 20 permissions.allow entries: workspace not trusted.\n" +
		`{"is_error":false,"result":"## Review\n\nAll good.\n\nVERDICT: approve"}` + "\n")

	got, err := parseClaudeJSON(out)
	if err != nil {
		t.Fatalf("should have found the JSON past the warning: %v", err)
	}
	if !strings.Contains(got, "All good") {
		t.Errorf("parsed the wrong content: %q", got)
	}
}

func TestParseClaudeOutputSurfacesErrors(t *testing.T) {
	out := []byte(`{"is_error":true,"result":"rate limit reached"}`)

	if _, err := parseClaudeJSON(out); err == nil {
		t.Fatal("an error result should be reported, not returned as a review")
	} else if !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("the error should carry the reason: %v", err)
	}
}

// GitHub rejects comments past 65536 bytes, and a long review hits that.
func TestLongReviewIsTrimmedToGitHubsLimit(t *testing.T) {
	body := strings.Repeat("A finding paragraph that goes on.\n\n", 4000)

	got := trimForGitHub(body)

	if len(got) > maxCommentBytes {
		t.Errorf("trimmed body is %d bytes, over the %d limit", len(got), maxCommentBytes)
	}
	if !strings.Contains(got, "truncated") {
		t.Error("the reader should be told the review was cut")
	}
}

func TestShortReviewIsUntouched(t *testing.T) {
	body := "## Review\n\nAll good."
	if got := trimForGitHub(body); got != body {
		t.Errorf("a short review was altered: %q", got)
	}
}

// GitHub refuses to let anyone approve their own pull request; recognizing
// that keeps a valid review from being thrown away over it.
func TestOwnPullRequestRefusalIsRecognized(t *testing.T) {
	messages := []string{
		"GraphQL: Can not approve your own pull request (addPullRequestReview)",
		"Can not request changes on your own pull request",
	}
	for _, m := range messages {
		if !isOwnPullRequest(m) {
			t.Errorf("should recognize the refusal: %q", m)
		}
	}
	if isOwnPullRequest("network unreachable") {
		t.Error("an unrelated failure must not be mistaken for it")
	}
}

// Instructions steer what the review looks at; they must not be able to
// replace the output contract, or publishing breaks.
func TestInstructionsSitBeforeTheOutputRules(t *testing.T) {
	prompt := fmt.Sprintf(defaultPrompt, "main", instructionBlock("only flag blockers"))

	idx := strings.Index(prompt, "only flag blockers")
	verdictIdx := strings.Index(prompt, "VERDICT: approve")

	if idx < 0 {
		t.Fatal("the instructions did not reach the prompt")
	}
	if verdictIdx < idx {
		t.Error("the verdict contract must come after the instructions, so it wins")
	}
	if !strings.Contains(prompt, "reviewing-code skill") {
		t.Error("the standing rules must survive alongside the instructions")
	}
}

func TestEmptyInstructionsAddNothing(t *testing.T) {
	if got := instructionBlock("   "); got != "" {
		t.Errorf("blank instructions should add nothing, got %q", got)
	}

	plain := fmt.Sprintf(defaultPrompt, "main", instructionBlock(""))
	if strings.Contains(plain, "Additional instructions") {
		t.Error("the heading should be absent when there are no instructions")
	}
}

// A per-review note narrows the focus; it does not discard preferences that
// always apply.
func TestCombineInstructionsKeepsBoth(t *testing.T) {
	got := CombineInstructions("always check for secrets", "focus on the auth changes")

	for _, want := range []string{"always check for secrets", "focus on the auth changes"} {
		if !strings.Contains(got, want) {
			t.Errorf("combined instructions lost %q: %q", want, got)
		}
	}
}

func TestCombineInstructionsHandlesEmpties(t *testing.T) {
	if got := CombineInstructions("", "just this"); got != "just this" {
		t.Errorf("got %q", got)
	}
	if got := CombineInstructions("just standing", ""); got != "just standing" {
		t.Errorf("got %q", got)
	}
	if got := CombineInstructions("  ", "  "); got != "" {
		t.Errorf("two blanks should combine to nothing, got %q", got)
	}
}
