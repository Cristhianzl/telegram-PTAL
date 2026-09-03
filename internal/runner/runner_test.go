package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Cristhianzl/telegram-PTAL/internal/config"
	"github.com/Cristhianzl/telegram-PTAL/internal/githubapi"
	"github.com/Cristhianzl/telegram-PTAL/internal/store"
)

// fakeTelegram records the messages it would receive, for the test to inspect.
type fakeTelegram struct {
	mu   sync.Mutex
	sent []string
	srv  *httptest.Server
}

func newFakeTelegram() *fakeTelegram {
	f := &fakeTelegram{}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Text string `json:"text"`
		}
		_ = json.Unmarshal(body, &payload)

		if strings.HasSuffix(r.URL.Path, "/sendMessage") {
			f.mu.Lock()
			f.sent = append(f.sent, payload.Text)
			f.mu.Unlock()
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ok":true,"result":{"message_id":1,"chat":{"id":123}}}`)
	}))
	return f
}

func (f *fakeTelegram) messages() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.sent...)
}

// fakeGitHub serves a sequence of REST search responses, one per cycle.
func newFakeGitHub(pages *[]string, idx *int, mu *sync.Mutex) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")

		// Only the review search returns content; the others come back empty.
		if !strings.Contains(r.URL.Query().Get("q"), "review-requested") {
			fmt.Fprint(w, `{"total_count":0,"items":[]}`)
			return
		}
		if *idx < len(*pages) {
			fmt.Fprint(w, (*pages)[*idx])
		} else {
			fmt.Fprint(w, `{"total_count":0,"items":[]}`)
		}
	}))
}

func item(number int, title string, comments int) string {
	return fmt.Sprintf(`{
      "number": %d, "title": %q,
      "html_url": "https://github.com/acme/app/pull/%d",
      "repository_url": "https://api.github.com/repos/acme/app",
      "node_id": "PR_%d", "draft": false, "comments": %d,
      "created_at": "2026-08-20T10:00:00Z",
      "updated_at": "2026-08-26T18:30:00Z",
      "user": {"login": "bob"}, "assignees": []
    }`, number, title, number, number, comments)
}

func search(items ...string) string {
	return fmt.Sprintf(`{"total_count":%d,"items":[%s]}`, len(items), strings.Join(items, ","))
}

func newTestRunner(t *testing.T, ghURL, tgURL string) (*Runner, *store.State) {
	t.Helper()
	cfg := &config.Config{
		Login:         "alice",
		TelegramToken: "fake",
		TelegramChat:  "123",
		MaxPerHour:    100,
		IgnoreDrafts:  true,
		StatePath:     filepath.Join(t.TempDir(), "state.json"),
		PublicOnly:    true,
	}
	state, err := store.Load(cfg.StatePath)
	if err != nil {
		t.Fatal(err)
	}
	r := New(cfg, state, log.New(io.Discard, "", 0))
	r.source.PublicClientForTest().SetEndpoint(ghURL)
	r.tg.SetBaseURL(tgURL)
	return r, state
}

// The full cycle: a silent first run, then one new pull request produces
// exactly one message, and repeating the cycle produces none.
func TestFullCycleFirstRunThenAlertThenQuiet(t *testing.T) {
	var mu sync.Mutex
	idx := 0
	pages := []string{
		search(item(1, "Existing pull request", 0)),
		search(item(1, "Existing pull request", 0), item(2, "New pull request awaiting review", 0)),
		search(item(1, "Existing pull request", 0), item(2, "New pull request awaiting review", 0)),
	}
	gh := newFakeGitHub(&pages, &idx, &mu)
	defer gh.Close()
	tg := newFakeTelegram()
	defer tg.srv.Close()

	r, _ := newTestRunner(t, gh.URL, tg.srv.URL)
	ctx := context.Background()

	// Cycle 1: first sync, nothing should be sent.
	if _, sent, err := r.Once(ctx); err != nil || sent != 0 {
		t.Fatalf("first run: sent=%d err=%v (should be silent)", sent, err)
	}

	// Cycle 2: a new pull request arrives asking for review.
	mu.Lock()
	idx = 1
	mu.Unlock()
	if _, sent, err := r.Once(ctx); err != nil || sent != 1 {
		t.Fatalf("second cycle: sent=%d err=%v (expected 1)", sent, err)
	}

	msgs := tg.messages()
	if len(msgs) != 1 {
		t.Fatalf("messages = %d, want 1: %v", len(msgs), msgs)
	}
	if !strings.Contains(msgs[0], "Review requested") {
		t.Errorf("wrong message:\n%s", msgs[0])
	}
	if !strings.Contains(msgs[0], "acme/app#2") {
		t.Errorf("the message does not identify the pull request:\n%s", msgs[0])
	}

	// Cycle 3: nothing changed - the alert must not repeat.
	mu.Lock()
	idx = 2
	mu.Unlock()
	if _, sent, err := r.Once(ctx); err != nil || sent != 0 {
		t.Fatalf("third cycle: sent=%d err=%v (nothing changed, should not repeat)", sent, err)
	}
	if len(tg.messages()) != 1 {
		t.Errorf("the alert was resent: %v", tg.messages())
	}
}

// Restarting the daemon rereads state from disk and resends nothing.
func TestRestartDoesNotResendAlerts(t *testing.T) {
	var mu sync.Mutex
	idx := 0
	pages := []string{
		search(item(1, "First", 0)),
		search(item(1, "First", 0), item(2, "Second", 0)),
	}
	gh := newFakeGitHub(&pages, &idx, &mu)
	defer gh.Close()
	tg := newFakeTelegram()
	defer tg.srv.Close()

	r, state := newTestRunner(t, gh.URL, tg.srv.URL)
	ctx := context.Background()

	r.Once(ctx)
	mu.Lock()
	idx = 1
	mu.Unlock()
	r.Once(ctx)

	if len(tg.messages()) != 1 {
		t.Fatalf("setup failed: %v", tg.messages())
	}

	// Simulate the restart: reload state from the same file.
	reloaded, err := store.Load(state.Path())
	if err != nil {
		t.Fatal(err)
	}
	cfg := r.cfg
	r2 := New(cfg, reloaded, log.New(io.Discard, "", 0))
	r2.source.PublicClientForTest().SetEndpoint(gh.URL)
	r2.tg.SetBaseURL(tg.srv.URL)

	if _, sent, err := r2.Once(ctx); err != nil || sent != 0 {
		t.Errorf("after restart: sent=%d err=%v (should resend nothing)", sent, err)
	}
}

// New comments on a known pull request produce a grouped notice.
func TestNewCommentGeneratesActivityAlert(t *testing.T) {
	var mu sync.Mutex
	idx := 0
	pages := []string{
		search(item(1, "A pull request", 2)),
		search(item(1, "A pull request", 7)),
	}
	gh := newFakeGitHub(&pages, &idx, &mu)
	defer gh.Close()
	tg := newFakeTelegram()
	defer tg.srv.Close()

	r, _ := newTestRunner(t, gh.URL, tg.srv.URL)
	ctx := context.Background()

	r.Once(ctx)
	mu.Lock()
	idx = 1
	mu.Unlock()
	r.Once(ctx)

	msgs := tg.messages()
	if len(msgs) != 1 {
		t.Fatalf("messages = %d: %v", len(msgs), msgs)
	}
	if !strings.Contains(msgs[0], "5 new comments") {
		t.Errorf("expected the comment count:\n%s", msgs[0])
	}
}

// The review button must only appear where reviewing is enabled: tapping it
// runs an agent over that branch.
func TestReviewButtonOnlyOnAllowedRepos(t *testing.T) {
	cfg := &config.Config{ReviewRepos: []string{"acme/api", "trusted-org"}}
	r := &Runner{cfg: cfg}

	cases := []struct {
		repo string
		want bool
	}{
		{"acme/api", true},          // exact
		{"trusted-org/anything", true}, // whole organization
		{"stranger/repo", false},
		{"acme/other", false},
	}
	for _, c := range cases {
		_, got := r.reviewButton(&githubapi.PullRequest{Repo: c.repo, Number: 1})
		if got != c.want {
			t.Errorf("%s: button offered = %v, want %v", c.repo, got, c.want)
		}
	}
}

// Telegram caps callback_data at 64 bytes. A truncated payload would resolve
// to the wrong pull request, so the button is dropped instead.
func TestReviewButtonDroppedWhenDataWouldOverflow(t *testing.T) {
	long := "some-very-long-organization-name/an-equally-long-repository-name-here"
	cfg := &config.Config{ReviewRepos: []string{long}}
	r := &Runner{cfg: cfg}

	if _, ok := r.reviewButton(&githubapi.PullRequest{Repo: long, Number: 123456}); ok {
		t.Error("a button whose data exceeds 64 bytes must not be offered")
	}
}

func TestReviewCallbackRoundTrips(t *testing.T) {
	cfg := &config.Config{ReviewRepos: []string{"acme/api"}}
	r := &Runner{cfg: cfg}

	btn, ok := r.reviewButton(&githubapi.PullRequest{Repo: "acme/api", Number: 412})
	if !ok {
		t.Fatal("the button should be offered")
	}
	if len(btn.Data) > maxCallbackData {
		t.Errorf("callback data is %d bytes, over the limit", len(btn.Data))
	}

	repo, number, ok := parseReviewCallback(btn.Data)
	if !ok || repo != "acme/api" || number != 412 {
		t.Errorf("round trip gave (%q, %d, %v)", repo, number, ok)
	}
}

func TestMalformedCallbackIsRejected(t *testing.T) {
	for _, data := range []string{"", "rv:", "rv:acme/api", "rv:acme/api:zero", "other:acme/api:1"} {
		if _, _, ok := parseReviewCallback(data); ok {
			t.Errorf("%q should not parse as a review callback", data)
		}
	}
}

// One review at a time: each drives a Claude session for minutes, and a burst
// of taps must not start several against the same subscription.
func TestOnlyOneReviewRunsAtATime(t *testing.T) {
	l := newReviewLimiter()

	if !l.acquire("acme/api#1") {
		t.Fatal("the first review should start")
	}
	if l.acquire("acme/api#2") {
		t.Error("a second, different review must wait")
	}

	l.release("acme/api#1")
	if !l.acquire("acme/api#2") {
		t.Error("releasing should let the next one through")
	}
}

// The menu and /help come from one list, so what Telegram offers cannot
// drift from what the bot actually answers.
func TestCommandMenuMatchesWhatIsAnswered(t *testing.T) {
	// Every command the handler accepts, from the switch in commands.go.
	answered := map[string]bool{
		"prs": true, "repo": true, "review": true, "status": true,
		"pause": true, "resume": true, "clear": true, "help": true,
	}

	for _, c := range menuCommands() {
		if !answered[c.Command] {
			t.Errorf("/%s is offered in the menu but not handled", c.Command)
		}
		if c.Description == "" {
			t.Errorf("/%s has no description; Telegram requires one", c.Command)
		}
		// Telegram rejects names that are not lowercase letters, digits
		// and underscores.
		for _, r := range c.Command {
			if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '_' {
				t.Errorf("/%s has a character Telegram will reject: %q", c.Command, r)
			}
		}
	}
}

// The menu is a picker, not documentation: argument forms and aliases would
// clutter it without adding anything.
func TestMenuHasNoDuplicatesOrAliases(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range menuCommands() {
		if seen[c.Command] {
			t.Errorf("/%s appears twice in the menu", c.Command)
		}
		seen[c.Command] = true
	}
	if seen["repo"] {
		t.Error("/repo is an alias of /prs and should stay out of the menu")
	}
}

// /help must document every command, including the ones kept out of the menu.
func TestHelpDocumentsEveryCommand(t *testing.T) {
	help := telegramHelp()

	for _, name := range []string{"prs", "repo", "review", "status", "pause", "resume", "clear", "help"} {
		if !strings.Contains(help, "/"+name) {
			t.Errorf("/help does not mention /%s", name)
		}
	}
}

// Staying silent on a non-command was a real failure: someone typed
// "Ptal doctor" twice, got nothing back, and concluded the diagnosis had run
// and found nothing — when no command had executed at all.
func TestPlainTextGetsAnAnswer(t *testing.T) {
	cases := []struct {
		text     string
		wantAny  []string
		wantNone bool
	}{
		// A machine-only command typed into the chat.
		{"Ptal doctor", []string{"machine", "/status"}, false},
		{"ptal config", []string{"machine"}, false},
		{"ptal stop", []string{"/pause"}, false},
		// A real command missing its slash.
		{"prs", []string{"/prs", "slash"}, false},
		{"ptal status", []string{"/status"}, false},
		// Ordinary conversation stays ignored: answering everything would
		// make the bot a nuisance.
		{"thanks!", nil, true},
		{"can you look at the deploy tomorrow morning please", nil, true},
	}

	for _, c := range cases {
		t.Run(c.text, func(t *testing.T) {
			reply := plainTextReply(c.text)

			if c.wantNone {
				if reply != "" {
					t.Errorf("should have stayed silent, replied: %s", reply)
				}
				return
			}
			if reply == "" {
				t.Fatalf("should have answered %q", c.text)
			}
			for _, want := range c.wantAny {
				if !strings.Contains(reply, want) {
					t.Errorf("reply should mention %q: %s", want, reply)
				}
			}
		})
	}
}

// Every machine-only command needs a hint naming the Telegram equivalent or
// saying plainly that there is none. A hint that just repeats the problem
// leaves the person where they started.
func TestEveryCliOnlyHintPointsSomewhere(t *testing.T) {
	for name, hint := range cliOnlyCommands {
		if hint == "" {
			t.Errorf("%s has no hint", name)
		}
		if !strings.Contains(hint, "/") && !strings.Contains(hint, "machine") {
			t.Errorf("%s: hint should name a command here or say it runs on the machine: %q", name, hint)
		}
	}
}
