package runner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Cristhianzl/telegram-PTAL/internal/githubapi"
	"github.com/Cristhianzl/telegram-PTAL/internal/reviewer"
	"github.com/Cristhianzl/telegram-PTAL/internal/telegram"
)

// reviewAllFromTelegram reviews every open pull request in a repository.
//
// Everything after the repository name is passed to the review verbatim, the
// same way /review works — the instruction is what decides whether a pull
// request is approved, blocked, or merged, rather than a policy baked in here.
func (r *Runner) reviewAllFromTelegram(ctx context.Context, args string) {
	name, instructions, _ := strings.Cut(strings.TrimSpace(args), " ")
	if name == "" {
		r.reply(ctx, "Usage: <code>/review-all &lt;repo&gt; [instructions]</code>\n\n"+
			"<i>Example: /review-all boibid if nothing blocks, approve and merge</i>")
		return
	}

	repo, err := githubapi.ResolveRepo(name, r.cfg.WatchRepos)
	if err != nil {
		r.reply(ctx, telegram.EscapeHTML(err.Error()))
		return
	}
	if !r.cfg.ReviewEnabledFor(repo) {
		r.reply(ctx, "Reviewing is not enabled for "+telegram.EscapeHTML(repo)+
			".\n\n<i>Enable it with</i> <code>ptal config review-repos "+
			telegram.EscapeHTML(repo)+"</code>")
		return
	}

	// One batch at a time, for the same reason a single review is: each one
	// drives a Claude session, and two competing would exhaust the
	// subscription partway through both.
	key := "batch:" + repo
	if !r.reviews.acquire(key) {
		r.reply(ctx, "A review is already running.")
		return
	}

	go func() {
		defer r.reviews.release(key)
		r.runBatchReview(context.Background(), repo, strings.TrimSpace(instructions))
	}()
}

// runBatchReview performs the batch and reports it in the chat.
func (r *Runner) runBatchReview(ctx context.Context, repo, instructions string) {
	r.source.EnsureCredential()

	prs, err := r.source.RepoPullRequests(ctx, repo, "", 100)
	if err != nil {
		r.reply(ctx, "Could not list "+telegram.EscapeHTML(repo)+": "+
			telegram.EscapeHTML(err.Error()))
		return
	}

	refs := make([]*reviewer.PullRequestRef, 0, len(prs))
	for _, pr := range prs {
		refs = append(refs, &reviewer.PullRequestRef{
			Number: pr.Number, Title: pr.Title, Author: pr.Author, IsDraft: pr.IsDraft,
		})
	}
	if len(refs) == 0 {
		r.reply(ctx, "📂 <b>"+telegram.EscapeHTML(repo)+"</b>\n\nNo open pull requests.")
		return
	}

	rules := r.cfg.ReviewRulesDir
	if rules == "" {
		rules = reviewer.DefaultRulesDir(r.cfg.Dir())
	}
	rv := reviewer.New(reviewer.Options{
		RulesDir:     rules,
		Timeout:      r.cfg.ReviewTimeout,
		Model:        r.cfg.ReviewModel,
		Instructions: reviewer.CombineInstructions(r.cfg.ReviewInstructions, instructions),
	})

	opts := reviewer.BatchOptions{
		Limit:       r.cfg.ReviewBatchLimit,
		SkipAuthors: r.cfg.IgnoreAuthors,
	}
	queued, _ := reviewer.PlanSize(refs, opts)

	header := fmt.Sprintf("🤖 <b>Reviewing %s</b>\n%d pull requests",
		telegram.EscapeHTML(repo), queued)
	if instructions != "" {
		header += "\n<i>" + telegram.EscapeHTML(instructions) + "</i>"
	}
	// A batch runs for tens of minutes; saying so up front is the difference
	// between waiting and assuming it broke.
	header += fmt.Sprintf("\n\n<i>about %d minutes</i>", queued*4)

	progressMsg, _ := r.tg.Send(ctx, r.cfg.TelegramChat, header, telegram.SendOptions{Silent: true})
	if progressMsg != nil {
		r.state.TrackMessage(progressMsg.MessageID)
	}

	started := time.Now()
	result, err := rv.ReviewAll(ctx, repo, refs, opts,
		func(done, total int, current *reviewer.PullRequestRef, stage string) {
			if progressMsg == nil || current == nil {
				return
			}
			_ = r.tg.Edit(ctx, r.cfg.TelegramChat, progressMsg.MessageID,
				fmt.Sprintf("%s\n\n<b>%d/%d</b> · #%d %s\n<i>%s…</i>",
					header, done+1, total, current.Number,
					telegram.EscapeHTML(truncate(current.Title, 40)),
					telegram.EscapeHTML(stage)))
		})
	if err != nil {
		r.reply(ctx, "❌ <b>Batch failed</b>\n\n<code>"+
			telegram.EscapeHTML(truncate(err.Error(), 300))+"</code>")
		return
	}

	r.log.Printf("batch review of %s finished in %s: %d reviewed",
		repo, time.Since(started).Round(time.Second), len(result.Reviewed))
	r.reply(ctx, renderBatchResult(result))
}

// renderBatchResult reports the batch, one line per pull request.
//
// Each line links to its review, because the summary answers "what happened"
// and the comment on GitHub answers "why" — repeating the findings here would
// produce a message nobody reads on a phone.
func renderBatchResult(b *reviewer.BatchResult) string {
	approved, merged, changes, commented := b.Counts()

	var s strings.Builder
	fmt.Fprintf(&s, "✅ <b>%s reviewed</b>\n", telegram.EscapeHTML(b.Repo))
	fmt.Fprintf(&s, "<i>%d pull requests in %s</i>\n\n",
		len(b.Reviewed), b.Duration.Round(time.Second))

	for _, r := range b.Reviewed {
		icon := verdictIcon(r)
		line := fmt.Sprintf("%s <a href=%q>#%d</a>", icon, r.CommentURL, r.Number)
		if r.Findings.Counted && r.Findings.Blockers > 0 {
			line += fmt.Sprintf(" · %d blocker", r.Findings.Blockers)
			if r.Findings.Blockers > 1 {
				line += "s"
			}
		}
		if r.Verdict.Merges() && !r.Merged {
			line += " · <i>not merged: " + telegram.EscapeHTML(r.MergeSkipped) + "</i>"
		}
		s.WriteString(line + "\n")
	}

	var tally []string
	if merged > 0 {
		tally = append(tally, fmt.Sprintf("%d merged", merged))
	}
	if approved-merged > 0 {
		tally = append(tally, fmt.Sprintf("%d approved", approved-merged))
	}
	if changes > 0 {
		tally = append(tally, fmt.Sprintf("%d changes requested", changes))
	}
	if commented > 0 {
		tally = append(tally, fmt.Sprintf("%d commented", commented))
	}
	if len(tally) > 0 {
		fmt.Fprintf(&s, "\n<b>%s</b>", strings.Join(tally, " · "))
	}

	// Anything not reviewed is stated rather than quietly absent, or the
	// count looks like the batch lost a pull request.
	if len(b.Skipped) > 0 {
		fmt.Fprintf(&s, "\n<i>%d skipped</i>", len(b.Skipped))
	}
	if len(b.Failed) > 0 {
		fmt.Fprintf(&s, "\n⚠️ <i>%d failed</i>", len(b.Failed))
	}
	return s.String()
}

func verdictIcon(r *reviewer.Result) string {
	switch {
	case r.Merged:
		return "🔀"
	case r.Verdict == reviewer.VerdictApprove, r.Verdict == reviewer.VerdictApproveAndMerge:
		return "✅"
	case r.Verdict == reviewer.VerdictRequestChanges:
		return "🔁"
	}
	return "💬"
}
