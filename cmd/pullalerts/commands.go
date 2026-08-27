package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cristhianzl/pullalerts/internal/config"
	"github.com/cristhianzl/pullalerts/internal/githubapi"
	"github.com/cristhianzl/pullalerts/internal/runner"
	"github.com/cristhianzl/pullalerts/internal/service"
	"github.com/cristhianzl/pullalerts/internal/store"
	"github.com/cristhianzl/pullalerts/internal/telegram"
)

// cmdSetup connects Telegram without making the user hunt for a chat ID.
func cmdSetup(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.TelegramToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is not in .env\n" +
			"  Create a bot by messaging @BotFather on Telegram and paste the token here.")
	}

	tg := telegram.New(cfg.TelegramToken)
	bot, err := tg.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("invalid Telegram token: %w", err)
	}
	fmt.Printf("✓ Bot @%s connected\n", bot.Username)

	if cfg.TelegramChat != "" {
		fmt.Printf("✓ Chat already configured (%s)\n", cfg.TelegramChat)
		return sendTest(ctx, tg, cfg.TelegramChat)
	}

	fmt.Printf("\n  Open https://t.me/%s and send /start to the bot.\n", bot.Username)
	fmt.Print("  Waiting")

	chatID, err := waitForStart(ctx, tg)
	if err != nil {
		return err
	}
	fmt.Println()

	if err := cfg.SaveChatID(chatID); err != nil {
		return fmt.Errorf("writing the chat into .env: %w", err)
	}
	fmt.Printf("✓ Chat connected and saved to .env (%s)\n", chatID)

	return sendTest(ctx, tg, chatID)
}

// waitForStart long-polls until someone messages the bot.
func waitForStart(ctx context.Context, tg *telegram.Client) (string, error) {
	deadline := time.Now().Add(5 * time.Minute)
	var offset int64

	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		updates, err := tg.GetUpdates(ctx, offset, 25)
		if err != nil {
			return "", err
		}
		for _, u := range updates {
			offset = u.UpdateID + 1
			if u.Message != nil && u.Message.Chat.ID != 0 {
				return strconv.FormatInt(u.Message.Chat.ID, 10), nil
			}
		}
		fmt.Print(".")
	}
	return "", fmt.Errorf("nobody messaged the bot within 5 minutes; run `pullalerts setup` again")
}

func sendTest(ctx context.Context, tg *telegram.Client, chatID string) error {
	text := "✅ <b>PullAlerts connected</b>\n\n" +
		"From now on I will tell you here about the pull requests tied to you.\n\n" +
		"<i>Next step: <code>pullalerts install</code> so I start with your computer.</i>"
	if _, err := tg.Send(ctx, chatID, text, telegram.SendOptions{}); err != nil {
		return fmt.Errorf("could not send the test message: %w", err)
	}
	fmt.Println("✓ Test message sent - check Telegram")
	return nil
}

// cmdDoctor checks everything that usually goes wrong, in the order it matters.
func cmdDoctor(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	fmt.Printf("pullalerts %s\n", version)
	if cfg.SourcePath != "" {
		fmt.Printf("configuration: %s\n", cfg.SourcePath)
	} else {
		fmt.Println("configuration: no .env found")
	}
	fmt.Printf("state:         %s\n\n", cfg.StatePath)

	problems := 0
	fail := func(format string, args ...any) {
		problems++
		fmt.Printf("✗ "+format+"\n", args...)
	}

	if cfg.GitHubToken == "" {
		if cfg.Login == "" {
			fail("No GH_PAT_TOKEN and no GITHUB_LOGIN: I do not know whose pull requests to fetch.")
		} else {
			fmt.Printf("• No token - using public search as @%s\n", cfg.Login)
		}
	} else {
		fmt.Printf("• GitHub token: %s · source: %s\n", describeToken(cfg.GitHubToken), cfg.TokenSource)
		checkGitHub(ctx, cfg, fail)
	}

	if cfg.TelegramToken == "" {
		fail("TELEGRAM_BOT_TOKEN is missing - create a bot with @BotFather.")
	} else {
		bot, err := telegram.New(cfg.TelegramToken).GetMe(ctx)
		if err != nil {
			fail("Invalid Telegram token: %v", err)
		} else {
			fmt.Printf("✓ Telegram: bot @%s\n", bot.Username)
			if cfg.TelegramChat == "" {
				fail("TELEGRAM_CHAT_ID is missing - run `pullalerts setup`.")
			} else {
				fmt.Printf("✓ Destination chat: %s\n", cfg.TelegramChat)
			}
		}
	}

	info := service.Status()
	switch {
	case info.Running:
		fmt.Printf("✓ Service running (%s)\n", info.Manager)
	case info.Installed:
		fmt.Printf("• Service installed but stopped (%s)\n", info.Manager)
	default:
		fmt.Printf("• Service not installed - run `pullalerts install`\n")
	}

	if st, err := store.Load(cfg.StatePath); err == nil && st.FirstRunDone {
		fmt.Printf("✓ Last sync: %s (%d pull requests known)\n",
			st.LastSuccessAt.Local().Format("2006-01-02 15:04"), len(st.PRs))
		if st.LastError != "" {
			fmt.Printf("• Last error: %s\n", st.LastError)
		}
	}

	fmt.Println()
	if problems > 0 {
		return fmt.Errorf("%d problem(s) found", problems)
	}
	fmt.Println("All good.")
	return nil
}

// checkGitHub exercises the token for real and turns known failures into
// actionable instructions instead of forwarding the raw API message.
func checkGitHub(ctx context.Context, cfg *config.Config, fail func(string, ...any)) {
	client := githubapi.New(cfg.GitHubToken)
	client.Repos = cfg.WatchRepos
	client.IgnoreAuthors = cfg.IgnoreAuthors
	client.MaxAgeDays = cfg.MaxAgeDays
	client.IncludeTeamReviews = cfg.IncludeTeamReviews

	login, err := client.Viewer(ctx)
	if err != nil {
		var authErr *githubapi.AuthError
		if errors.As(err, &authErr) {
			fail("GitHub rejected the token (%s).\n"+
				"  Generate a classic PAT at:\n"+
				"  https://github.com/settings/tokens/new?scopes=repo,notifications", authErr.Msg)
			return
		}
		fail("Could not reach GitHub: %v", err)
		return
	}
	fmt.Printf("✓ GitHub: authenticated as @%s\n", login)

	// Search is what actually matters: a token can authenticate and still be
	// blocked by an organization policy.
	snap, err := client.Fetch(ctx)
	if err != nil {
		var policyErr *githubapi.PolicyError
		if errors.As(err, &policyErr) {
			fmt.Printf("• The organization rejects this token:\n    %s\n", policyErr.Msg)
			fmt.Println("  -> PullAlerts will fall back to public search automatically.")
			fmt.Println("  -> Private repositories and CI state become unavailable.")
			fmt.Println("  -> For full mode, use a classic PAT with the organization's SSO authorized.")
			return
		}
		fail("Search failed: %v", err)
		return
	}

	counts := snap.Counts()
	total := 0
	for _, n := range counts {
		total += n
	}
	if total == 0 {
		fmt.Println("• Search worked but returned no pull requests.")
		if len(cfg.WatchRepos) > 0 {
			fmt.Printf("  WATCH_REPOS is limiting it to: %s\n", strings.Join(cfg.WatchRepos, ", "))
		}
		if cfg.MaxAgeDays > 0 {
			fmt.Printf("  MAX_AGE_DAYS is limiting it to the last %d days.\n", cfg.MaxAgeDays)
		}
		return
	}
	fmt.Printf("✓ Search working · %d pull requests tied to you:\n", len(snap.PRs))
	for _, b := range githubapi.AllBuckets {
		if n := counts[b]; n > 0 {
			fmt.Printf("    %-24s %d\n", b.Label(), n)
		}
	}
}

// describeToken classifies the token without ever printing its contents.
func describeToken(token string) string {
	switch {
	case strings.HasPrefix(token, "github_pat_"):
		return "fine-grained (note: cannot reach /notifications and may be " +
			"blocked by an organization policy)"
	case strings.HasPrefix(token, "ghp_"):
		return "classic"
	case strings.HasPrefix(token, "gho_"), strings.HasPrefix(token, "ghu_"):
		return "OAuth"
	}
	return "unrecognized format"
}

func cmdOnce(ctx context.Context) error {
	// `once` runs even without Telegram configured, listing what it found.
	// That makes it possible to verify the search before connecting the bot.
	cfg, state, err := prepareAllowNoChat()
	if err != nil {
		return err
	}
	if cfg.TelegramChat == "" {
		fmt.Println("• Telegram not configured - listing only (run `pullalerts setup` to receive alerts)")
	}
	r := runner.New(cfg, state, log.New(os.Stdout, "", log.Ltime))

	snap, sent, err := r.Once(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("\nmode: %s\n", r.Source().Mode().Description())
	fmt.Printf("pull requests tied to @%s: %d · messages sent: %d\n\n", snap.Viewer, len(snap.PRs), sent)
	for _, b := range githubapi.AllBuckets {
		prs := snap.InBucket(b)
		if len(prs) == 0 {
			continue
		}
		fmt.Printf("%s (%d)\n", b.Label(), len(prs))
		for _, pr := range prs {
			flags := pr.ChecksSymbol()
			if pr.IsDraft {
				flags += " draft"
			}
			fmt.Printf("   %-30s %s %s\n", pr.Slug(), truncate(pr.Title, 50), flags)
		}
		fmt.Println()
	}
	return nil
}

// cmdPanel sends the current picture to Telegram without waiting for a change.
func cmdPanel(ctx context.Context) error {
	cfg, state, err := prepare()
	if err != nil {
		return err
	}
	r := runner.New(cfg, state, log.New(os.Stdout, "", log.Ltime))

	snap, _, err := r.Once(ctx)
	if err != nil {
		return err
	}
	if err := r.Panel(ctx, snap, true); err != nil {
		return err
	}
	fmt.Printf("✓ Panel sent and pinned · %d pull requests\n", len(snap.PRs))
	return nil
}

func cmdRun(ctx context.Context) error {
	cfg, state, err := prepare()
	if err != nil {
		return err
	}
	return runner.New(cfg, state, log.New(os.Stdout, "", log.LstdFlags)).Run(ctx)
}

func cmdInstall() error {
	if err := service.Install(); err != nil {
		return err
	}
	info := service.Status()
	fmt.Printf("✓ Installed under %s\n", info.Manager)
	if info.UnitPath != "" {
		fmt.Printf("  %s\n", info.UnitPath)
	}
	fmt.Println("✓ It will start on its own when the computer boots")
	return nil
}

func cmdUninstall() error {
	if err := service.Uninstall(); err != nil {
		return err
	}
	fmt.Println("✓ Service removed")
	return nil
}

func cmdStatus() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	info := service.Status()
	fmt.Printf("service:      %s\n", statusWord(info))
	fmt.Printf("manager:      %s\n", info.Manager)

	st, err := store.Load(cfg.StatePath)
	if err != nil {
		return err
	}
	if !st.FirstRunDone {
		fmt.Println("sync:         never run")
		return nil
	}
	fmt.Printf("user:         @%s\n", st.Viewer)
	fmt.Printf("known PRs:    %d\n", len(st.PRs))
	fmt.Printf("last sync:    %s\n", st.LastSuccessAt.Local().Format("2006-01-02 15:04:05"))
	fmt.Printf("sent (1h):    %d\n", st.SentLastHour())
	if st.LastError != "" {
		fmt.Printf("last error:   %s\n", st.LastError)
	}
	return nil
}

func statusWord(i service.Info) string {
	switch {
	case i.Running:
		return "running"
	case i.Installed:
		return "installed, stopped"
	}
	return "not installed"
}

// prepareAllowNoChat loads what is needed to search, even with no Telegram
// destination configured.
func prepareAllowNoChat() (*config.Config, *store.State, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}
	if cfg.GitHubToken == "" && cfg.Login == "" {
		return nil, nil, fmt.Errorf("set GH_PAT_TOKEN or GITHUB_LOGIN in .env")
	}
	state, err := store.Load(cfg.StatePath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading state: %w", err)
	}
	return cfg, state, nil
}

// prepare loads configuration and state, with errors that say what to do next.
func prepare() (*config.Config, *store.State, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}
	state, err := store.Load(cfg.StatePath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading state: %w", err)
	}
	return cfg, state, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return strings.TrimSpace(s[:n-1]) + "…"
}
