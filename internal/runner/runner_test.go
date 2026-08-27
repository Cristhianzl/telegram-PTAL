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
