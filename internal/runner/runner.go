// Package runner ties the pieces together: fetch the picture, compare it with
// the previous one, apply the guards, and send what is left.
package runner

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/cristhianzl/pullalerts/internal/config"
	"github.com/cristhianzl/pullalerts/internal/engine"
	"github.com/cristhianzl/pullalerts/internal/githubapi"
	"github.com/cristhianzl/pullalerts/internal/store"
	"github.com/cristhianzl/pullalerts/internal/telegram"
)

// Runner executes synchronization cycles.
type Runner struct {
	cfg    *config.Config
	source *githubapi.Source
	tg     *telegram.Client
	state  *store.State
	log    *log.Logger

	quiet engine.QuietHours
}

// New assembles a fully configured runner.
func New(cfg *config.Config, state *store.State, logger *log.Logger) *Runner {
	// The token goes to the source even when public mode is forced: it still
	// counts for REST search, where it raises the budget from ten to thirty
	// requests per minute.
	src := githubapi.NewSource(cfg.GitHubToken, cfg.Login, cfg.WatchRepos,
		cfg.IgnoreAuthors, cfg.MaxAgeDays, cfg.IncludeTeamReviews)
	if cfg.PublicOnly {
		src.UsePublicOnly(false)
	}
	r := &Runner{
		cfg:    cfg,
		source: src,
		tg:     telegram.New(cfg.TelegramToken),
		state:  state,
		log:    logger,
		quiet:  engine.ParseQuietHours(cfg.QuietHours),
	}

	// Announcing a downgrade on Telegram matters: the user needs to know
	// they stopped receiving CI state and approvals.
	src.OnModeChange = func(from, to githubapi.Mode, reason string) {
		r.log.Printf("mode changed: %s -> %s (%s)", from, to, reason)
		if to == githubapi.ModePublic {
			r.notify(context.Background(),
				"⚠️ <b>Reduced mode</b>\n\nThe organization rejected your token, so I switched to "+
					"public search. Pull requests still arrive, but without CI state, approvals "+
					"or private repositories.\n\n<i>Use a classic PAT with SSO authorized to "+
					"return to full mode.</i>", false)
		}
	}
	return r
}

// Source exposes the data source, for the `doctor` command.
func (r *Runner) Source() *githubapi.Source { return r.source }

// Once runs one complete cycle. It returns the picture obtained and how many
// messages were sent.
func (r *Runner) Once(ctx context.Context) (*githubapi.Snapshot, int, error) {
	if err := r.source.Resolve(ctx); err != nil {
		return nil, 0, fmt.Errorf("identifying your user: %w", err)
	}

	snap, err := r.source.Fetch(ctx)
	if err != nil {
		r.state.LastErrorAt = time.Now().UTC()
		r.state.LastError = err.Error()
		return nil, 0, err
	}

	firstRun := !r.state.FirstRunDone
	mode := string(r.source.Mode())
	modeChanged := r.state.Mode != "" && r.state.Mode != mode

	if modeChanged {
		r.log.Printf("data source changed from %s to %s: re-syncing silently",
			r.state.Mode, mode)
	}

	events, next := engine.Diff(r.state.PRs, snap, engine.Options{
		IgnoreDrafts: r.cfg.IgnoreDrafts,
		FirstRun:     firstRun,
		ModeChanged:  modeChanged,
	})

	urgent, batched := engine.Gate(events, r.state, engine.GateOptions{
		Quiet:      r.quiet,
		MaxPerHour: r.cfg.MaxPerHour,
		Now:        time.Now(),
	})

	sent := 0
	if !firstRun && !modeChanged {
		sent += r.deliver(ctx, urgent)
		sent += r.deliver(ctx, batched)
	} else {
		r.log.Printf("first sync: %d pull requests recorded silently", len(snap.PRs))
	}

	r.state.PRs = next
	r.state.Viewer = snap.Viewer
	r.state.Mode = mode
	r.state.FirstRunDone = true
	r.state.LastSuccessAt = time.Now().UTC()
	r.state.LastError = ""

	if err := r.state.Save(); err != nil {
		r.log.Printf("warning: could not write state: %v", err)
	}
	return snap, sent, nil
}

// deliver sends a batch, if there is anything to send.
func (r *Runner) deliver(ctx context.Context, batch engine.Batch) int {
	if batch.Empty() {
		return 0
	}
	text := telegram.RenderEvents(batch.Events)
	if batch.Dropped > 0 {
		text += fmt.Sprintf("\n\n<i>+%d events suppressed by the hourly limit</i>", batch.Dropped)
	}
	if text == "" {
		return 0
	}

	if r.cfg.TelegramChat == "" {
		// With no destination configured the cycle is still useful: state
		// advances and `once` can show what it found.
		return 0
	}

	opts := telegram.SendOptions{Silent: batch.Silent}
	// A single event gets direct buttons to the pull request.
	if len(batch.Events) == 1 && batch.Events[0].PR != nil {
		pr := batch.Events[0].PR
		opts.Buttons = [][]telegram.Button{{
			{Text: "Open PR", URL: pr.URL},
			{Text: "View diff", URL: pr.URL + "/files"},
		}}
	}

	if _, err := r.tg.Send(ctx, r.cfg.TelegramChat, text, opts); err != nil {
		r.log.Printf("failed to send on Telegram: %v", err)
		return 0
	}
	r.state.RecordSend()
	return 1
}

// Panel sends, or updates, the panel holding the current state of your pull
// requests.
//
// Keeping one pinned message and rewriting it each cycle avoids turning the
// conversation into a log of notices: what matters is the state now, not the
// sequence of changes.
func (r *Runner) Panel(ctx context.Context, snap *githubapi.Snapshot, pin bool) error {
	text := telegram.RenderPanel(snap)

	if r.state.PanelMessageID != 0 {
		err := r.tg.Edit(ctx, r.cfg.TelegramChat, r.state.PanelMessageID, text)
		if err == nil {
			return nil
		}
		// The user may have deleted the message; send a new one rather than
		// failing.
		r.log.Printf("previous panel unavailable, sending a new one: %v", err)
	}

	msg, err := r.tg.Send(ctx, r.cfg.TelegramChat, text, telegram.SendOptions{Silent: true})
	if err != nil {
		return err
	}
	r.state.PanelMessageID = msg.MessageID
	if pin {
		if err := r.tg.Pin(ctx, r.cfg.TelegramChat, msg.MessageID); err != nil {
			r.log.Printf("could not pin the panel: %v", err)
		}
	}
	return r.state.Save()
}

// notify sends a standalone message, ignoring errors: it is operational
// signalling and must not bring the cycle down.
func (r *Runner) notify(ctx context.Context, text string, silent bool) {
	if r.cfg.TelegramChat == "" {
		return
	}
	if _, err := r.tg.Send(ctx, r.cfg.TelegramChat, text, telegram.SendOptions{Silent: silent}); err != nil {
		r.log.Printf("failed to notify on Telegram: %v", err)
	}
}

// Run keeps cycling until the context is cancelled.
//
// Network errors are treated as normal — a daemon running for months will
// lose connectivity many times — with growing backoff and a little jitter so
// attempts from several machines do not line up.
func (r *Runner) Run(ctx context.Context) error {
	r.log.Printf("starting · interval %s · mode %s", r.cfg.PollInterval, r.source.Mode())

	const maxBackoff = 15 * time.Minute
	failures := 0
	warned := false

	for {
		snap, sent, err := r.Once(ctx)

		switch {
		case err == nil:
			if failures > 0 {
				r.log.Printf("connection restored after %d failures", failures)
				if warned {
					r.notify(ctx, "🟢 <b>Syncing again</b>\n\nThe connection to GitHub was restored.", true)
					warned = false
				}
			}
			failures = 0
			r.log.Printf("cycle ok · %d pull requests · %d messages · mode %s", len(snap.PRs), sent, r.source.Mode())

		default:
			var authErr *githubapi.AuthError
			if errors.As(err, &authErr) {
				// An invalid credential is not fixed by retrying.
				r.log.Printf("invalid credential: %v", authErr)
				r.notify(ctx, "🔑 <b>Invalid GitHub token</b>\n\nSyncing is paused. "+
					"Generate a new token and run <code>pullalerts doctor</code>.", false)
				return authErr
			}
			failures++
			r.log.Printf("cycle failed (%d): %v", failures, err)
			// Only bother the user once the failure persists.
			if failures == 5 && !warned {
				r.notify(ctx, "🔌 <b>No contact with GitHub</b>\n\nStill retrying.", true)
				warned = true
			}
		}

		wait := r.cfg.PollInterval
		if failures > 0 {
			wait = backoff(failures, r.cfg.PollInterval, maxBackoff)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(wait):
		}
	}
}

// backoff grows exponentially with up to 20% jitter, so several instances do
// not hit the server at exactly the same moment.
func backoff(failures int, base, max time.Duration) time.Duration {
	d := base << min(failures-1, 6)
	if d > max {
		d = max
	}
	jitter := time.Duration(rand.Int64N(int64(d) / 5))
	return d + jitter
}
