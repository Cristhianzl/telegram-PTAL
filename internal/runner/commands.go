package runner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Cristhianzl/telegram-PTAL/internal/githubapi"
	"github.com/Cristhianzl/telegram-PTAL/internal/telegram"
)

// telegramHelp is what /help answers with.
const telegramHelp = `<b>PTAL commands</b>

/prs - your pull requests right now
/prs &lt;repo&gt; - every open PR in a repository
/status - last sync, mode, rate limit
/clear - delete every message I have sent
/pause 2h - stop alerting for a while
/resume - start alerting again
/help - this

<i>Settings live on the machine running me: `+"`ptal config`"+`</i>`

// ListenCommands answers Telegram commands until the context is cancelled.
//
// It uses long polling rather than a webhook, which is what lets PTAL run
// behind a home NAT with no public address and no open port.
func (r *Runner) ListenCommands(ctx context.Context) {
	if r.cfg.TelegramChat == "" {
		return
	}
	r.log.Printf("listening for Telegram commands")

	for {
		if ctx.Err() != nil {
			return
		}

		updates, err := r.tg.GetUpdates(ctx, r.state.LastUpdateID+1, 30)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			// Long polling fails routinely on flaky networks; a quiet retry
			// is right, and the sync loop already reports real outages.
			select {
			case <-ctx.Done():
				return
			case <-time.After(15 * time.Second):
			}
			continue
		}

		for _, u := range updates {
			r.state.LastUpdateID = u.UpdateID
			if u.Message == nil {
				continue
			}
			// Only the configured chat is obeyed. Anyone can find a bot and
			// message it; without this check they could clear your history.
			if fmt.Sprint(u.Message.Chat.ID) != r.cfg.TelegramChat {
				continue
			}
			r.handleCommand(ctx, strings.TrimSpace(u.Message.Text))
		}
		if len(updates) > 0 {
			if err := r.state.Save(); err != nil {
				r.log.Printf("warning: could not persist the update cursor: %v", err)
			}
		}
	}
}

func (r *Runner) handleCommand(ctx context.Context, text string) {
	if !strings.HasPrefix(text, "/") {
		return
	}
	// Telegram appends @botname in groups.
	command, args, _ := strings.Cut(text, " ")
	command = strings.ToLower(strings.TrimSuffix(command, "@"+r.botUsername))
	args = strings.TrimSpace(args)

	switch command {
	case "/start", "/help":
		r.reply(ctx, telegramHelp)
	case "/prs":
		if args != "" {
			r.replyWithRepo(ctx, args)
		} else {
			r.replyWithPanel(ctx)
		}
	case "/repo":
		r.replyWithRepo(ctx, args)
	case "/status":
		r.replyWithStatus(ctx)
	case "/clear":
		r.clearHistory(ctx)
	case "/pause":
		r.pauseFromTelegram(ctx, args)
	case "/resume":
		r.resumeFromTelegram(ctx)
	default:
		r.reply(ctx, "Unknown command. Try /help")
	}
}

// clearHistory deletes everything the bot has posted.
func (r *Runner) clearHistory(ctx context.Context) {
	ids := r.state.TrackedMessages()
	if len(ids) == 0 {
		r.reply(ctx, "Nothing to clear.")
		return
	}

	deleted, err := r.tg.Delete(ctx, r.cfg.TelegramChat, ids)
	r.state.ForgetMessages()
	if err := r.state.Save(); err != nil {
		r.log.Printf("warning: could not save after clearing: %v", err)
	}

	switch {
	case err != nil && deleted == 0:
		r.reply(ctx, "Could not delete anything. Telegram only lets me remove "+
			"messages I sent in the last 48 hours.")
	case deleted < len(ids):
		// Being precise here matters: claiming a clean sweep when older
		// messages remain would just look broken.
		r.reply(ctx, fmt.Sprintf("Deleted %d of %d. The rest are older than "+
			"48 hours, which Telegram will not let me remove.", deleted, len(ids)))
	default:
		r.reply(ctx, fmt.Sprintf("Deleted %d messages.", deleted))
	}
}

func (r *Runner) replyWithPanel(ctx context.Context) {
	snap, _, err := r.Once(ctx)
	if err != nil {
		r.reply(ctx, "Could not reach GitHub: "+telegram.EscapeHTML(err.Error()))
		return
	}
	r.reply(ctx, telegram.RenderPanel(snap))
}

// replyWithRepo answers with every open pull request in a repository, which
// is a different question from "what needs me" and deliberately ignores the
// user filters.
func (r *Runner) replyWithRepo(ctx context.Context, input string) {
	if input == "" {
		r.reply(ctx, "Which repository? Try <code>/prs owner/name</code>, or a "+
			"short name from the ones you watch.")
		return
	}

	repo, err := githubapi.ResolveRepo(input, r.cfg.WatchRepos)
	if err != nil {
		r.reply(ctx, telegram.EscapeHTML(err.Error()))
		return
	}

	prs, err := r.source.RepoPullRequests(ctx, repo, 50)
	if err != nil {
		r.reply(ctx, "Could not read "+telegram.EscapeHTML(repo)+": "+
			telegram.EscapeHTML(err.Error()))
		return
	}
	r.reply(ctx, telegram.RenderRepoList(repo, prs, 25))
}

func (r *Runner) replyWithStatus(ctx context.Context) {
	var b strings.Builder
	b.WriteString("🤖 <b>PTAL</b>\n\n")
	if r.state.Viewer != "" {
		fmt.Fprintf(&b, "GitHub: <b>@%s</b>\n", telegram.EscapeHTML(r.state.Viewer))
	}
	fmt.Fprintf(&b, "Mode: %s\n", r.source.Mode())
	fmt.Fprintf(&b, "Tracking: %d pull requests\n", len(r.state.PRs))

	if !r.state.LastSuccessAt.IsZero() {
		fmt.Fprintf(&b, "Last sync: %s\n", telegram.HumanAge(r.state.LastSuccessAt))
	}
	if paused, until := r.state.Paused(); paused {
		fmt.Fprintf(&b, "\n⏸ <b>Paused</b> until %s\n", until.Local().Format("15:04"))
	}
	if r.state.LastError != "" {
		fmt.Fprintf(&b, "\n⚠️ %s\n", telegram.EscapeHTML(r.state.LastError))
	}

	counts := map[githubapi.Bucket]int{}
	for _, pr := range r.state.PRs {
		for _, bucket := range pr.Buckets {
			counts[bucket]++
		}
	}
	if len(counts) > 0 {
		b.WriteString("\n")
		for _, bucket := range githubapi.AllBuckets {
			if n := counts[bucket]; n > 0 {
				fmt.Fprintf(&b, "%s: %d\n", telegram.EscapeHTML(bucket.Label()), n)
			}
		}
	}
	r.reply(ctx, b.String())
}

func (r *Runner) pauseFromTelegram(ctx context.Context, args string) {
	if args == "" {
		args = "1h"
	}
	d, err := time.ParseDuration(args)
	if err != nil || d <= 0 {
		r.reply(ctx, "Say how long, like <code>/pause 2h</code> or <code>/pause 30m</code>.")
		return
	}

	until := time.Now().Add(d)
	r.state.Pause(until)
	if err := r.state.Save(); err != nil {
		r.reply(ctx, "Could not save the pause: "+telegram.EscapeHTML(err.Error()))
		return
	}
	r.reply(ctx, fmt.Sprintf("⏸ Paused until %s. Still watching, just quiet.\n\n"+
		"<i>/resume to start again</i>", until.Local().Format("15:04")))
}

func (r *Runner) resumeFromTelegram(ctx context.Context) {
	if paused, _ := r.state.Paused(); !paused {
		r.reply(ctx, "Not paused.")
		return
	}
	r.state.Resume()
	if err := r.state.Save(); err != nil {
		r.reply(ctx, "Could not save: "+telegram.EscapeHTML(err.Error()))
		return
	}
	r.reply(ctx, "▶️ Alerting again.")
}

// reply answers a command. Replies are tracked like any other message so
// /clear removes them too.
func (r *Runner) reply(ctx context.Context, text string) {
	msg, err := r.tg.Send(ctx, r.cfg.TelegramChat, text, telegram.SendOptions{Silent: true})
	if err != nil {
		r.log.Printf("failed to reply on Telegram: %v", err)
		return
	}
	r.state.TrackMessage(msg.MessageID)
}
