package runner

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Cristhianzl/telegram-PTAL/internal/githubapi"
	"github.com/Cristhianzl/telegram-PTAL/internal/telegram"
)

// telegramCommands is the single list behind both Telegram's native command
// menu and /help. Keeping one source means the menu cannot drift from what
// the bot actually answers.
var telegramCommands = []struct {
	Name, Args, Description string
	// Menu excludes aliases and argument-only forms, which would clutter
	// Telegram's picker without adding anything.
	Menu bool
}{
	{"prs", "", "Your pull requests right now", true},
	{"prs", "&lt;repo&gt;", "Every open PR in a repository", false},
	{"prs", "&lt;repo&gt; me", "Only the ones you opened", false},
	{"repo", "&lt;repo&gt;", "The same as /prs &lt;repo&gt;", false},
	{"review", "&lt;repo&gt; &lt;n&gt; [instructions]", "Review a PR with Claude Code", true},
	{"status", "", "Last sync, mode, what is tracked", true},
	{"pause", "2h", "Stop alerting for a while", true},
	{"resume", "", "Start alerting again", true},
	{"clear", "", "Delete every message I have sent", true},
	{"help", "", "This list", true},
}

// telegramHelp renders the command list for /help.
func telegramHelp() string {
	var b strings.Builder
	b.WriteString("<b>PTAL commands</b>\n\n")
	for _, c := range telegramCommands {
		if c.Args != "" {
			fmt.Fprintf(&b, "/%s %s - %s\n", c.Name, c.Args, c.Description)
		} else {
			fmt.Fprintf(&b, "/%s - %s\n", c.Name, c.Description)
		}
	}
	b.WriteString("\n<i>Settings live on the machine running me: <code>ptal config</code></i>")
	return b.String()
}

// menuCommands is what Telegram shows in its command picker.
func menuCommands() []telegram.Command {
	var out []telegram.Command
	seen := map[string]bool{}
	for _, c := range telegramCommands {
		if !c.Menu || seen[c.Name] {
			continue
		}
		seen[c.Name] = true
		out = append(out, telegram.Command{Command: c.Name, Description: c.Description})
	}
	return out
}

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

			if q := u.CallbackQuery; q != nil {
				if q.Message == nil || fmt.Sprint(q.Message.Chat.ID) != r.cfg.TelegramChat {
					continue
				}
				r.handleReviewCallback(ctx, q)
				continue
			}

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
	// /commands and /start are what people try before /help.
	case "/start", "/help", "/commands", "/cmd":
		r.reply(ctx, telegramHelp())
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
	case "/review":
		r.reviewFromTelegram(ctx, args)
	default:
		r.reply(ctx, "Unknown command. Try /help")
	}
}

// reviewFromTelegram starts a review named by hand, for pull requests that
// did not arrive with a button.
func (r *Runner) reviewFromTelegram(ctx context.Context, args string) {
	fields := strings.Fields(args)
	if len(fields) < 2 {
		r.reply(ctx, "Usage: <code>/review &lt;repo&gt; &lt;number&gt;</code>")
		return
	}
	repo, err := githubapi.ResolveRepo(fields[0], r.cfg.WatchRepos)
	if err != nil {
		r.reply(ctx, telegram.EscapeHTML(err.Error()))
		return
	}
	number, err := strconv.Atoi(fields[1])
	if err != nil {
		r.reply(ctx, "That is not a pull request number.")
		return
	}
	// Everything after the number steers this one review.
	instructions := strings.TrimSpace(strings.Join(fields[2:], " "))
	if !r.cfg.ReviewEnabledFor(repo) {
		r.reply(ctx, "Reviewing is not enabled for "+telegram.EscapeHTML(repo)+
			".\n\n<i>Enable it with</i> <code>ptal config review-repos "+
			telegram.EscapeHTML(repo)+"</code>")
		return
	}

	key := key(repo, number)
	if !r.reviews.acquire(key) {
		r.reply(ctx, "A review is already running.")
		return
	}
	go func() {
		defer r.reviews.release(key)
		r.runReview(context.Background(), repo, number, instructions)
	}()
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

	// A second word narrows to one author: "/prs langflow me".
	name, author, _ := strings.Cut(input, " ")
	author = strings.TrimSpace(author)

	repo, err := githubapi.ResolveRepo(name, r.cfg.WatchRepos)
	if err != nil {
		r.reply(ctx, telegram.EscapeHTML(err.Error()))
		return
	}

	prs, err := r.source.RepoPullRequests(ctx, repo, author, 50)
	if err != nil {
		r.reply(ctx, "Could not read "+telegram.EscapeHTML(repo)+": "+
			telegram.EscapeHTML(err.Error()))
		return
	}
	r.reply(ctx, telegram.RenderRepoList(repo, author, prs, 25))
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
