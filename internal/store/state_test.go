package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Cristhianzl/telegram-PTAL/internal/engine"
	"github.com/Cristhianzl/telegram-PTAL/internal/githubapi"
)

func tempState(t *testing.T) *State {
	t.Helper()
	s, err := Load(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestMissingFileStartsEmpty(t *testing.T) {
	s := tempState(t)

	if s.FirstRunDone {
		t.Error("a fresh state must not claim it already synced")
	}
	if len(s.PRs) != 0 {
		t.Error("a fresh state should come back empty")
	}
}

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s, _ := Load(path)

	s.Viewer = "alice"
	s.FirstRunDone = true
	s.PRs["a"] = &engine.Tracked{
		PullRequest: githubapi.PullRequest{ID: "a", Repo: "acme/app", Number: 1},
		FirstSeenAt: time.Now().UTC(),
	}
	s.MarkSeen("fp-1")
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	again, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if again.Viewer != "alice" || !again.FirstRunDone {
		t.Error("the markers did not survive the disk round trip")
	}
	if len(again.PRs) != 1 || again.PRs["a"].Repo != "acme/app" {
		t.Errorf("the pull requests did not survive: %+v", again.PRs)
	}
	if !again.Seen("fp-1") {
		t.Error("fingerprints did not survive - events would be resent")
	}
}

// A corrupt state must not stop the daemon from starting: the worst that
// happens by starting over is one silent sync.
func TestCorruptFileDoesNotBlockStartup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := os.WriteFile(path, []byte("{{{ this is not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	s, err := Load(path)
	if err != nil {
		t.Fatalf("a corrupt file should be tolerated, got error: %v", err)
	}
	if s.FirstRunDone {
		t.Error("after corruption the daemon must treat it as a first run")
	}
}

func TestSaveIsAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	s, _ := Load(path)
	s.Viewer = "alice"

	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	// No temporary file may remain after a successful write.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("a temporary file was left behind: %s", e.Name())
		}
	}
	if st, _ := os.Stat(path); st.Mode().Perm() != 0o600 {
		t.Error("state should be private to the user")
	}
}

func TestOldFingerprintsArePruned(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s, _ := Load(path)

	s.Fingerprints["antigo"] = time.Now().Add(-60 * 24 * time.Hour)
	s.Fingerprints["recente"] = time.Now()
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	again, _ := Load(path)
	if again.Seen("antigo") {
		t.Error("the old fingerprint should have been pruned")
	}
	if !again.Seen("recente") {
		t.Error("a recent fingerprint was pruned by mistake")
	}
}

func TestSentLastHourCountsOnlyRecent(t *testing.T) {
	s := tempState(t)
	s.Sent = []time.Time{
		time.Now().Add(-2 * time.Hour),
		time.Now().Add(-30 * time.Minute),
		time.Now(),
	}

	if got := s.SentLastHour(); got != 2 {
		t.Errorf("messages in the last hour = %d, want 2", got)
	}
}
