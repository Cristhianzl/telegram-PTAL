package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Cristhianzl/telegram-PTAL/internal/config"
	"github.com/Cristhianzl/telegram-PTAL/internal/service"
	"github.com/Cristhianzl/telegram-PTAL/internal/store"
)

// collectFailures captures what reportDaemonHealth flags as a problem.
func collectFailures(t *testing.T, st *store.State) []string {
	t.Helper()
	if err := st.Save(); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{StatePath: st.Path()}
	var failures []string
	reportDaemonHealth(cfg, service.Info{Running: true}, func(format string, args ...any) {
		failures = append(failures, format)
	})
	return failures
}

func newDaemonState(t *testing.T) *store.State {
	t.Helper()
	s, err := store.Load(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	s.FirstRunDone = true
	s.DaemonStartedAt = time.Now().Add(-time.Hour).UTC()
	s.LastSuccessAt = time.Now().Add(-time.Minute).UTC()
	return s
}

// The failure this exists to catch: doctor runs in a shell where the keyring
// is unlocked and reports everything green, while the daemon has been
// searching anonymously since boot and cannot see any private repository.
func TestDoctorFlagsADaemonWithNoCredential(t *testing.T) {
	st := newDaemonState(t)
	st.DaemonMode = "public"
	st.DaemonAuthenticated = false

	failures := collectFailures(t, st)

	if len(failures) == 0 {
		t.Fatal("a daemon with no credential must be reported as a problem")
	}
	joined := strings.Join(failures, "\n")
	if !strings.Contains(joined, "NO GitHub credential") {
		t.Errorf("the message should name the problem plainly: %s", joined)
	}
	// Saying what is broken without saying what to do is what made the
	// previous version useless.
	if !strings.Contains(joined, "ptal restart") {
		t.Errorf("the message must say how to fix it: %s", joined)
	}
	if !strings.Contains(joined, "private repositories are invisible") {
		t.Errorf("the message should explain the consequence: %s", joined)
	}
}

// Reduced mode with a working credential is a note, not a failure: it still
// delivers alerts, just without CI state.
func TestDoctorTreatsReducedModeAsANote(t *testing.T) {
	st := newDaemonState(t)
	st.DaemonMode = "public"
	st.DaemonAuthenticated = true

	if failures := collectFailures(t, st); len(failures) != 0 {
		t.Errorf("authenticated public mode should not be a failure: %v", failures)
	}
}

func TestDoctorSaysNothingWhenHealthy(t *testing.T) {
	st := newDaemonState(t)
	st.DaemonMode = "rich"
	st.DaemonAuthenticated = true

	if failures := collectFailures(t, st); len(failures) != 0 {
		t.Errorf("a healthy daemon should raise nothing: %v", failures)
	}
}

// A daemon that has not finished a cycle since starting is stuck on something
// only the log explains.
func TestDoctorFlagsADaemonThatNeverCompletedACycle(t *testing.T) {
	st := newDaemonState(t)
	st.DaemonMode = "rich"
	st.DaemonAuthenticated = true
	st.DaemonStartedAt = time.Now().Add(-time.Hour).UTC()
	st.LastSuccessAt = st.DaemonStartedAt.Add(-time.Hour) // before it started

	failures := collectFailures(t, st)

	if len(failures) == 0 {
		t.Fatal("a daemon stuck since startup must be reported")
	}
	if !strings.Contains(strings.Join(failures, "\n"), "journalctl") {
		t.Error("the message should point at the log")
	}
}

// With the service stopped there is no daemon to report on, and stale state
// would produce a misleading warning.
func TestDoctorSaysNothingWhenTheServiceIsStopped(t *testing.T) {
	st := newDaemonState(t)
	st.DaemonMode = "public"
	st.DaemonAuthenticated = false
	if err := st.Save(); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{StatePath: st.Path()}
	var failures []string
	reportDaemonHealth(cfg, service.Info{Running: false}, func(f string, a ...any) {
		failures = append(failures, f)
	})

	if len(failures) != 0 {
		t.Errorf("a stopped service should not be diagnosed: %v", failures)
	}
}
