// Package reviewer runs a Claude Code review over a pull request and
// publishes the result to GitHub.
//
// The division of labour is deliberate. Claude reads and reasons; it is given
// read-only tools and never touches the network or the GitHub API. Publishing
// the comment and setting the verdict is done here, by code, with `gh`.
//
// That split is what makes it safe to review a pull request from someone you
// do not know. Running an agent with write access over an untrusted branch
// means a malicious Makefile, git hook, or build script executes on your
// machine. Restricting the tools removes that entire class of risk, and the
// review skill was already designed to emit a comment rather than perform one.
package reviewer

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Verdict is the conclusion the review reached.
type Verdict string

const (
	VerdictApprove        Verdict = "approve"
	VerdictComment        Verdict = "comment"
	VerdictRequestChanges Verdict = "request_changes"
)

// Result is everything one review produced.
type Result struct {
	Repo     string
	Number   int
	Body     string
	Verdict  Verdict
	Duration time.Duration
	// Posted records whether the comment reached GitHub.
	Posted bool
	// CommentURL is the published comment, when there is one.
	CommentURL string
}

// Options configures a review run.
type Options struct {
	// RulesDir holds the .claude directory whose rules the review follows.
	RulesDir string
	// WorkRoot is where checkouts are made. Each review gets a fresh
	// subdirectory that is removed afterwards.
	WorkRoot string
	// Timeout bounds one review. A review that hangs must not hold the
	// daemon's worker forever.
	Timeout time.Duration
	// DryRun produces the review without publishing it.
	DryRun bool
	// Model overrides the model Claude uses.
	Model string
	// Prompt replaces the instruction sent to Claude entirely.
	Prompt string
	// Instructions are added to the standard prompt rather than replacing
	// it, so the review still follows the project rules and still emits a
	// verdict line.
	Instructions string
}

// DefaultTimeout is generous: a real review reads a diff and several files.
const DefaultTimeout = 15 * time.Minute

// readOnlyTools is what Claude is allowed to do.
//
// Everything here reads. There is no Write, no Edit, and no general Bash —
// only the specific git and gh subcommands needed to see the change. This is
// the security boundary of the whole feature.
var readOnlyTools = []string{
	"Read", "Grep", "Glob",
	"Bash(git diff:*)",
	"Bash(git log:*)",
	"Bash(git show:*)",
	"Bash(git status:*)",
	"Bash(gh pr view:*)",
	"Bash(gh pr diff:*)",
}

// defaultPrompt asks for the project's own review, in the shape the rules
// already define, plus a machine-readable verdict line.
const defaultPrompt = `Review this pull request.

Follow the review rules in .claude/ exactly: apply the reviewing-code skill,
use its severity labels and its output format, and produce the review as a
single GitHub comment.

The diff under review is the difference between this branch and its merge
base. Start with:

  git diff $(git merge-base HEAD origin/%s)...HEAD
%s
Output ONLY the review comment itself, in Markdown, with no preamble and no
closing remarks. The last line of your output must be exactly one of:

VERDICT: approve
VERDICT: comment
VERDICT: request_changes

Use approve when the change is safe to merge, request_changes when there is at
least one Blocker, and comment when the findings are worth raising but none of
them block.`

// Reviewer runs reviews.
type Reviewer struct {
	opts Options
}

// New builds a Reviewer, filling in the defaults.
func New(opts Options) *Reviewer {
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultTimeout
	}
	if opts.WorkRoot == "" {
		opts.WorkRoot = filepath.Join(os.TempDir(), "ptal-reviews")
	}
	return &Reviewer{opts: opts}
}

// Progress reports what stage a review has reached, so the caller can keep
// the user informed during what is a multi-minute operation.
type Progress func(stage string)

// Review checks out the pull request, runs Claude over it, and publishes the
// result. The returned Result is filled in even when publishing is skipped.
func (r *Reviewer) Review(ctx context.Context, repo string, number int, progress Progress) (*Result, error) {
	if progress == nil {
		progress = func(string) {}
	}
	started := time.Now()

	ctx, cancel := context.WithTimeout(ctx, r.opts.Timeout)
	defer cancel()

	progress("checking out")
	dir, baseBranch, err := r.checkout(ctx, repo, number)
	if err != nil {
		return nil, fmt.Errorf("checking out %s#%d: %w", repo, number, err)
	}
	// The checkout is a full clone of someone else's branch; it must not
	// survive the review.
	defer os.RemoveAll(dir)

	if err := r.installRules(dir); err != nil {
		return nil, fmt.Errorf("installing review rules: %w", err)
	}

	progress("reviewing")
	body, err := r.runClaude(ctx, dir, baseBranch)
	if err != nil {
		return nil, err
	}

	verdict, body := extractVerdict(body)
	result := &Result{
		Repo: repo, Number: number,
		Body: body, Verdict: verdict,
		Duration: time.Since(started),
	}

	if r.opts.DryRun {
		return result, nil
	}

	progress("publishing")
	url, err := r.publish(ctx, repo, number, result)
	if err != nil {
		// The review itself succeeded; failing to publish is worth
		// reporting without throwing the work away.
		return result, fmt.Errorf("publishing the review: %w", err)
	}
	result.Posted = true
	result.CommentURL = url
	return result, nil
}

// checkout clones just enough of the repository to review the branch, and
// returns the directory and the base branch name.
func (r *Reviewer) checkout(ctx context.Context, repo string, number int) (string, string, error) {
	if err := os.MkdirAll(r.opts.WorkRoot, 0o700); err != nil {
		return "", "", err
	}
	dir, err := os.MkdirTemp(r.opts.WorkRoot, fmt.Sprintf("pr-%d-", number))
	if err != nil {
		return "", "", err
	}

	// A treeless clone still carries full history for merge-base while
	// avoiding downloading every blob of a large repository.
	if out, err := runCmd(ctx, "", "gh", "repo", "clone", repo, dir, "--",
		"--filter=blob:none", "--quiet"); err != nil {
		os.RemoveAll(dir)
		return "", "", fmt.Errorf("%w: %s", err, out)
	}

	if out, err := runCmd(ctx, dir, "gh", "pr", "checkout", fmt.Sprint(number)); err != nil {
		os.RemoveAll(dir)
		return "", "", fmt.Errorf("%w: %s", err, out)
	}

	base, err := runCmd(ctx, dir, "gh", "pr", "view", fmt.Sprint(number),
		"--json", "baseRefName", "--jq", ".baseRefName")
	if err != nil {
		base = "main"
	}
	return dir, strings.TrimSpace(base), nil
}

// installRules copies the .claude directory into the checkout so the review
// follows the project's own conventions.
//
// Copying rather than symlinking matters: a symlink pointing outside the
// checkout would let a crafted repository read through it.
func (r *Reviewer) installRules(dir string) error {
	if r.opts.RulesDir == "" {
		return nil
	}
	src := r.opts.RulesDir
	if filepath.Base(src) != ".claude" {
		src = filepath.Join(src, ".claude")
	}
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("rules directory %s: %w", src, err)
	}
	// The pull request may ship its own .claude; the reviewer's rules win.
	dst := filepath.Join(dir, ".claude")
	os.RemoveAll(dst)
	return copyTree(src, dst)
}

// runClaude executes the review headlessly and returns the review text.
func (r *Reviewer) runClaude(ctx context.Context, dir, baseBranch string) (string, error) {
	prompt := r.opts.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf(defaultPrompt, baseBranch, instructionBlock(r.opts.Instructions))
	}

	args := []string{
		"-p", prompt,
		"--output-format", "json",
		// Tools are passed on the command line rather than relied on from
		// settings.json: a fresh checkout is an untrusted workspace, and
		// Claude ignores its permission entries there.
		"--allowed-tools", strings.Join(readOnlyTools, ","),
	}
	if r.opts.Model != "" {
		args = append(args, "--model", r.opts.Model)
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("the review timed out after %s", r.opts.Timeout)
		}
		return "", fmt.Errorf("claude failed: %w: %s", err, truncate(stderr.String(), 300))
	}

	text, err := parseClaudeJSON(stdout.Bytes())
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("the review came back empty")
	}
	return text, nil
}

// claudeResult is the subset of the CLI's JSON output that matters here.
type claudeResult struct {
	Result  string `json:"result"`
	IsError bool   `json:"is_error"`
}

// parseClaudeJSON pulls the result out of the CLI output.
//
// The CLI may print warnings before the JSON - an untrusted workspace notice,
// for one - so the object is located rather than assumed to start at byte
// zero.
func parseClaudeJSON(out []byte) (string, error) {
	var payload claudeResult

	if err := json.Unmarshal(bytes.TrimSpace(out), &payload); err == nil {
		if payload.IsError {
			return "", fmt.Errorf("claude reported an error: %s", truncate(payload.Result, 300))
		}
		return payload.Result, nil
	}

	// Fall back to scanning for the line that holds the object.
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 1<<20), 16<<20)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if !bytes.HasPrefix(line, []byte("{")) {
			continue
		}
		if err := json.Unmarshal(line, &payload); err == nil && payload.Result != "" {
			if payload.IsError {
				return "", fmt.Errorf("claude reported an error: %s", truncate(payload.Result, 300))
			}
			return payload.Result, nil
		}
	}
	return "", fmt.Errorf("could not read Claude's output: %s", truncate(string(out), 300))
}

var verdictLine = regexp.MustCompile(`(?im)^\s*VERDICT:\s*(approve|comment|request[_ -]?changes)\s*$`)

// extractVerdict separates the machine-readable verdict from the comment body.
//
// A missing verdict is treated as "comment" rather than an error: the review
// text is still worth publishing, and defaulting to the neutral outcome means
// a parsing slip can never approve or block a pull request on its own.
func extractVerdict(text string) (Verdict, string) {
	match := verdictLine.FindStringSubmatch(text)
	if match == nil {
		return VerdictComment, strings.TrimSpace(text)
	}

	body := strings.TrimSpace(verdictLine.ReplaceAllString(text, ""))
	switch strings.ToLower(strings.NewReplacer("-", "_", " ", "_").Replace(match[1])) {
	case "approve":
		return VerdictApprove, body
	case "request_changes":
		return VerdictRequestChanges, body
	}
	return VerdictComment, body
}

// CombineInstructions merges the standing instructions with the ones given
// for a single review. Both are kept: a per-review note narrows the focus, it
// does not discard the preferences that always apply.
func CombineInstructions(standing, perReview string) string {
	standing = strings.TrimSpace(standing)
	perReview = strings.TrimSpace(perReview)
	switch {
	case standing == "":
		return perReview
	case perReview == "":
		return standing
	default:
		return standing + "\n\n" + perReview
	}
}

// instructionBlock places the caller's instructions between the diff command
// and the output rules.
//
// The position is deliberate. Putting them last would let an instruction
// override the verdict format and break publishing; putting them first would
// bury them under the standing rules. Between the two, they steer what gets
// looked at while the output contract still has the final word.
func instructionBlock(instructions string) string {
	instructions = strings.TrimSpace(instructions)
	if instructions == "" {
		return ""
	}
	return "\nAdditional instructions for this review:\n\n" + instructions + "\n"
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func runCmd(ctx context.Context, dir, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// copyTree copies a directory, skipping symlinks so nothing can point out of
// the destination.
func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		switch {
		case info.IsDir():
			return os.MkdirAll(target, 0o755)
		case info.Mode()&os.ModeSymlink != 0:
			return nil
		default:
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, data, info.Mode().Perm())
		}
	})
}
