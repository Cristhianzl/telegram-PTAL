package reviewer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// maxCommentBytes is GitHub's limit for a single comment body.
const maxCommentBytes = 65536

// publish posts the review to GitHub and sets the verdict.
//
// This runs here rather than inside Claude on purpose. The review skill is
// explicitly forbidden from performing GitHub mutations, and keeping the side
// effect in code means every write is one auditable call rather than whatever
// an agent decided to run.
func (r *Reviewer) publish(ctx context.Context, repo string, number int, result *Result) (string, error) {
	body := trimForGitHub(result.Body)

	// The body goes through a file: a long review as an argument runs into
	// the argument length limit, and shell-quoting Markdown is a trap.
	f, err := os.CreateTemp("", "ptal-review-*.md")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(body); err != nil {
		f.Close()
		return "", err
	}
	f.Close()

	action, err := reviewAction(result.Verdict)
	if err != nil {
		return "", err
	}

	out, err := runCmd(ctx, "", "gh", "pr", "review", fmt.Sprint(number),
		"--repo", repo, action, "--body-file", f.Name())
	if err == nil {
		return commentURLFrom(out, repo, number), nil
	}

	// GitHub refuses to let you approve or request changes on your own pull
	// request. Falling back to a plain comment keeps the review from being
	// lost over a rule that has nothing to do with its content.
	if isOwnPullRequest(out) {
		out, cErr := runCmd(ctx, "", "gh", "pr", "comment", fmt.Sprint(number),
			"--repo", repo, "--body-file", f.Name())
		if cErr != nil {
			return "", fmt.Errorf("%w: %s", cErr, truncate(out, 300))
		}
		return commentURLFrom(out, repo, number), nil
	}
	return "", fmt.Errorf("%w: %s", err, truncate(out, 300))
}

func reviewAction(v Verdict) (string, error) {
	switch v {
	case VerdictApprove, VerdictApproveAndMerge:
		return "--approve", nil
	case VerdictRequestChanges:
		return "--request-changes", nil
	case VerdictComment:
		return "--comment", nil
	}
	return "", fmt.Errorf("unknown verdict %q", v)
}

// isOwnPullRequest recognizes GitHub's refusal to review your own work.
func isOwnPullRequest(out string) bool {
	low := strings.ToLower(out)
	return strings.Contains(low, "can not approve your own") ||
		strings.Contains(low, "cannot approve your own") ||
		strings.Contains(low, "can not request changes on your own") ||
		strings.Contains(low, "review your own")
}

// trimForGitHub keeps the comment inside GitHub's size limit, cutting at a
// line boundary so the Markdown does not end mid-structure.
func trimForGitHub(body string) string {
	if len(body) <= maxCommentBytes {
		return body
	}
	const notice = "\n\n---\n\n*Review truncated: it exceeded GitHub's comment size limit.*"
	cut := body[:maxCommentBytes-len(notice)]
	if idx := strings.LastIndex(cut, "\n\n"); idx > len(cut)/2 {
		cut = cut[:idx]
	}
	return cut + notice
}

// commentURLFrom picks the URL out of gh's output, falling back to the pull
// request itself so the caller always has somewhere to point.
func commentURLFrom(out, repo string, number int) string {
	for _, field := range strings.Fields(out) {
		if strings.HasPrefix(field, "https://") {
			return field
		}
	}
	return fmt.Sprintf("https://github.com/%s/pull/%d", repo, number)
}

// merge completes a pull request after the review has been posted.
//
// It runs only when the verdict asked for it, and it refuses cases GitHub
// would either reject or silently do the wrong thing with. Those checks are
// here rather than left to `gh` because a merge is the one action in this
// tool that cannot be taken back.
func (r *Reviewer) merge(ctx context.Context, repo string, number int) (bool, string) {
	out, err := runCmd(ctx, "", "gh", "pr", "view", fmt.Sprint(number), "--repo", repo,
		"--json", "isDraft,mergeable,state,mergeStateStatus")
	if err != nil {
		return false, "could not read the pull request: " + truncate(out, 120)
	}

	var pr struct {
		IsDraft          bool   `json:"isDraft"`
		Mergeable        string `json:"mergeable"`
		State            string `json:"state"`
		MergeStateStatus string `json:"mergeStateStatus"`
	}
	if err := json.Unmarshal([]byte(out), &pr); err != nil {
		return false, "could not read the pull request state"
	}

	switch {
	case pr.State != "OPEN":
		return false, "the pull request is " + strings.ToLower(pr.State)
	case pr.IsDraft:
		return false, "it is a draft"
	case pr.Mergeable == "CONFLICTING":
		return false, "it has merge conflicts"
	case pr.MergeStateStatus == "BLOCKED":
		return false, "the branch protection rules block it"
	case pr.MergeStateStatus == "BEHIND":
		return false, "the branch is behind and needs updating"
	}

	if out, err := runCmd(ctx, "", "gh", "pr", "merge", fmt.Sprint(number),
		"--repo", repo, "--squash", "--delete-branch=false"); err != nil {
		return false, truncate(out, 160)
	}
	return true, ""
}

// DefaultRulesDir is where PTAL looks for review rules when none is set:
// alongside the configuration, so the rules travel with it.
func DefaultRulesDir(configDir string) string {
	return filepath.Join(configDir, "reviewer-rules")
}
