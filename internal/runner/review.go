package runner

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Cristhianzl/telegram-PTAL/internal/githubapi"
	"github.com/Cristhianzl/telegram-PTAL/internal/reviewer"
	"github.com/Cristhianzl/telegram-PTAL/internal/telegram"
)

// callbackPrefix marks a button that asks for a review.
const callbackPrefix = "rv:"

// maxCallbackData is Telegram's hard limit on callback_data.
const maxCallbackData = 64

// reviewButton builds the review button for a pull request, or reports that
// there should not be one.
//
// The button is only offered where reviewing is enabled, because tapping it
// runs an agent over that branch. Long repository names can overflow
// Telegram's 64-byte callback limit, and a truncated payload would resolve to
// the wrong pull request - so the button is dropped rather than risk that.
func (r *Runner) reviewButton(pr *githubapi.PullRequest) (telegram.Button, bool) {
	if !r.cfg.ReviewEnabledFor(pr.Repo) {
		return telegram.Button{}, false
	}
	data := fmt.Sprintf("%s%s:%d", callbackPrefix, pr.Repo, pr.Number)
	if len(data) > maxCallbackData {
		return telegram.Button{}, false
	}
	return telegram.Button{Text: "🤖 Review", Data: data}, true
}

// parseReviewCallback reads a repository and number back out of button data.
func parseReviewCallback(data string) (string, int, bool) {
	if !strings.HasPrefix(data, callbackPrefix) {
		return "", 0, false
	}
	rest := strings.TrimPrefix(data, callbackPrefix)
	idx := strings.LastIndex(rest, ":")
	if idx <= 0 {
		return "", 0, false
	}
	number, err := strconv.Atoi(rest[idx+1:])
	if err != nil || number <= 0 {
		return "", 0, false
	}
	return rest[:idx], number, true
}

// reviewLimiter keeps one review running at a time.
//
// A review takes minutes and drives a Claude Code session. Running several at
// once would compete for the same subscription and could exhaust it on a
// burst of button taps.
type reviewLimiter struct {
	mu     sync.Mutex
	active map[string]bool
}

func newReviewLimiter() *reviewLimiter {
	return &reviewLimiter{active: map[string]bool{}}
}

func (l *reviewLimiter) acquire(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.active) > 0 {
		return false
	}
	l.active[key] = true
	return true
}

func (l *reviewLimiter) release(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.active, key)
}

// handleReviewCallback answers a tapped review button.
func (r *Runner) handleReviewCallback(ctx context.Context, q *telegram.CallbackQuery) {
	repo, number, ok := parseReviewCallback(q.Data)
	if !ok {
		_ = r.tg.AnswerCallback(ctx, q.ID, "Unrecognized button")
		return
	}

	// Answering immediately clears Telegram's loading spinner. The work
	// itself takes minutes and continues after this returns.
	if !r.cfg.ReviewEnabledFor(repo) {
		_ = r.tg.AnswerCallback(ctx, q.ID, "Reviewing is not enabled for this repository")
		return
	}

	key := fmt.Sprintf("%s#%d", repo, number)
	if !r.reviews.acquire(key) {
		_ = r.tg.AnswerCallback(ctx, q.ID, "A review is already running")
		return
	}
	_ = r.tg.AnswerCallback(ctx, q.ID, "Starting the review…")

	// The review outlives this update, so it gets its own context rather
	// than the polling one.
	go func() {
		defer r.reviews.release(key)
		r.runReview(context.Background(), repo, number)
	}()
}

// runReview performs the review and reports it in the chat.
func (r *Runner) runReview(ctx context.Context, repo string, number int) {
	rules := r.cfg.ReviewRulesDir
	if rules == "" {
		rules = reviewer.DefaultRulesDir(r.cfg.Dir())
	}

	progressMsg, _ := r.tg.Send(ctx, r.cfg.TelegramChat,
		fmt.Sprintf("🤖 <b>Reviewing %s</b>\n<i>starting…</i>",
			telegram.EscapeHTML(fmt.Sprintf("%s#%d", repo, number))),
		telegram.SendOptions{Silent: true})
	if progressMsg != nil {
		r.state.TrackMessage(progressMsg.MessageID)
	}

	edit := func(stage string) {
		if progressMsg == nil {
			return
		}
		_ = r.tg.Edit(ctx, r.cfg.TelegramChat, progressMsg.MessageID,
			fmt.Sprintf("🤖 <b>Reviewing %s</b>\n<i>%s…</i>",
				telegram.EscapeHTML(fmt.Sprintf("%s#%d", repo, number)),
				telegram.EscapeHTML(stage)))
	}

	rv := reviewer.New(reviewer.Options{
		RulesDir: rules,
		Timeout:  r.cfg.ReviewTimeout,
		Model:    r.cfg.ReviewModel,
	})

	started := time.Now()
	result, err := rv.Review(ctx, repo, number, edit)

	switch {
	case err != nil && result == nil:
		r.log.Printf("review of %s#%d failed: %v", repo, number, err)
		edit(fmt.Sprintf("failed after %s", time.Since(started).Round(time.Second)))
		r.reply(ctx, fmt.Sprintf("❌ <b>Review failed</b>\n%s#%d\n\n<code>%s</code>",
			telegram.EscapeHTML(repo), number, telegram.EscapeHTML(truncate(err.Error(), 400))))

	case err != nil:
		// The review ran but could not be published; the findings are worth
		// showing anyway.
		r.log.Printf("review of %s#%d not published: %v", repo, number, err)
		edit("done, not published")
		r.reply(ctx, fmt.Sprintf("⚠️ <b>Review done, but not posted</b>\n%s\n\n<code>%s</code>\n\n%s",
			telegram.EscapeHTML(key(repo, number)),
			telegram.EscapeHTML(truncate(err.Error(), 200)),
			telegram.EscapeHTML(truncate(result.Body, 2500))))

	default:
		edit(fmt.Sprintf("done in %s", result.Duration.Round(time.Second)))
		r.reply(ctx, renderReviewResult(result))
	}
}

func key(repo string, number int) string { return fmt.Sprintf("%s#%d", repo, number) }

// renderReviewResult summarizes the review for the chat, and links to the
// full comment on GitHub rather than repeating it here.
func renderReviewResult(result *reviewer.Result) string {
	icon, label := "💬", "Commented"
	switch result.Verdict {
	case reviewer.VerdictApprove:
		icon, label = "✅", "Approved"
	case reviewer.VerdictRequestChanges:
		icon, label = "🔁", "Changes requested"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s <b>%s</b>\n", icon, label)
	fmt.Fprintf(&b, "<a href=%q>%s</a>\n\n",
		result.CommentURL, telegram.EscapeHTML(key(result.Repo, result.Number)))

	if summary := firstSection(result.Body); summary != "" {
		fmt.Fprintf(&b, "%s\n\n", telegram.EscapeHTML(summary))
	}
	fmt.Fprintf(&b, "<i>reviewed in %s</i>", result.Duration.Round(time.Second))
	return b.String()
}

// firstSection pulls the summary paragraph out of the review, which is what
// belongs in a phone notification. The full review is on GitHub.
func firstSection(body string) string {
	lines := strings.Split(body, "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Stop at the first finding heading: everything past it is detail.
		if strings.HasPrefix(trimmed, "##") && len(out) > 0 {
			break
		}
		if strings.HasPrefix(trimmed, "#") || trimmed == "---" {
			continue
		}
		if trimmed != "" {
			out = append(out, trimmed)
		}
		if len(out) >= 6 {
			break
		}
	}
	return truncate(strings.Join(out, "\n"), 700)
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
