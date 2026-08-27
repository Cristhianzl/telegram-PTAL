// Package store persists what the daemon must remember between runs: the last
// picture of the pull requests, the keys of events already sent, and a few
// health markers.
//
// The format is a single JSON file written atomically. For one user's data
// volume — dozens of pull requests — that is enough, and it keeps the binary
// free of external dependencies.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Cristhianzl/telegram-PTAL/internal/engine"
)

// schemaVersion allows migrating the file in future versions.
const schemaVersion = 1

// Retention of the history used for deduplication and the hourly ceiling.
const (
	fingerprintTTL = 30 * 24 * time.Hour
	sentWindow     = time.Hour
)

// State is everything that survives between runs.
type State struct {
	Version int    `json:"version"`
	Viewer  string `json:"viewer"`

	// PRs is the last known picture, the basis for the next comparison.
	PRs map[string]*engine.Tracked `json:"prs"`
	// Fingerprints records events already sent, so none is ever repeated.
	Fingerprints map[string]time.Time `json:"fingerprints"`
	// Sent holds the timestamps of recent messages, for the hourly ceiling.
	Sent []time.Time `json:"sent"`

	// FirstRunDone tells "no pull requests" apart from "never synced".
	FirstRunDone bool `json:"first_run_done"`
	// Mode records which source produced the stored picture. Different
	// sources fill different sets of fields, so comparing pictures from
	// different modes produces changes that never happened.
	Mode string `json:"mode,omitempty"`

	LastSuccessAt time.Time `json:"last_success_at"`
	LastErrorAt   time.Time `json:"last_error_at"`
	LastError     string    `json:"last_error,omitempty"`

	// PausedUntil suppresses delivery without stopping the daemon. Keeping
	// the cycle running while paused is the whole point: state stays current,
	// so resuming does not produce a flood of everything that changed.
	PausedUntil time.Time `json:"paused_until,omitempty"`

	// PanelMessageID is the pinned message the daemon edits each cycle.
	PanelMessageID int `json:"panel_message_id,omitempty"`

	// SentMessageIDs are the messages the bot posted, so /clear can remove
	// them. Telegram will not delete anything older than 48 hours, so the
	// list is trimmed to roughly that horizon.
	SentMessageIDs []int `json:"sent_message_ids,omitempty"`

	// LastUpdateID is the getUpdates cursor. Persisting it means a restart
	// does not replay commands that were already handled.
	LastUpdateID int64 `json:"last_update_id,omitempty"`

	path string
}

// Load reads the state from disk. A missing file returns an empty state,
// which is the normal path on the first run.
func Load(path string) (*State, error) {
	s := &State{
		Version:      schemaVersion,
		PRs:          map[string]*engine.Tracked{},
		Fingerprints: map[string]time.Time{},
		path:         path,
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return s, nil
	}

	if err := json.Unmarshal(data, s); err != nil {
		// A corrupt state must not stop the daemon from starting: the worst
		// that happens by starting over is one silent sync.
		s.PRs = map[string]*engine.Tracked{}
		s.Fingerprints = map[string]time.Time{}
		s.FirstRunDone = false
	}
	if s.PRs == nil {
		s.PRs = map[string]*engine.Tracked{}
	}
	if s.Fingerprints == nil {
		s.Fingerprints = map[string]time.Time{}
	}
	s.Version = schemaVersion
	s.path = path
	return s, nil
}

// Save writes atomically: to a temporary file, then rename, so a power loss
// mid-write cannot leave half a JSON document behind.
func (s *State) Save() error {
	if s.path == "" {
		return fmt.Errorf("state has no path configured")
	}
	if dir := filepath.Dir(s.path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}

	s.prune()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, s.path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// prune drops old history so the file does not grow without bound.
func (s *State) prune() {
	cutoff := time.Now().Add(-fingerprintTTL)
	for k, at := range s.Fingerprints {
		if at.Before(cutoff) {
			delete(s.Fingerprints, k)
		}
	}
	s.Sent = recentTimes(s.Sent, sentWindow)
}

// Seen reports whether the event has ever been sent.
func (s *State) Seen(fingerprint string) bool {
	_, ok := s.Fingerprints[fingerprint]
	return ok
}

// MarkSeen records the event as sent.
func (s *State) MarkSeen(fingerprint string) {
	s.Fingerprints[fingerprint] = time.Now().UTC()
}

// maxTrackedMessages bounds the delete list. Telegram refuses to delete
// anything older than 48 hours anyway, so tracking more is pure growth.
const maxTrackedMessages = 500

// RecordSend counts a delivered message, for the hourly ceiling.
func (s *State) RecordSend() {
	s.Sent = append(recentTimes(s.Sent, sentWindow), time.Now().UTC())
}

// TrackMessage remembers a message so /clear can delete it later.
func (s *State) TrackMessage(id int) {
	if id == 0 {
		return
	}
	s.SentMessageIDs = append(s.SentMessageIDs, id)
	if excess := len(s.SentMessageIDs) - maxTrackedMessages; excess > 0 {
		s.SentMessageIDs = s.SentMessageIDs[excess:]
	}
}

// TrackedMessages returns the messages eligible for deletion.
func (s *State) TrackedMessages() []int {
	return append([]int(nil), s.SentMessageIDs...)
}

// ForgetMessages clears the delete list.
func (s *State) ForgetMessages() {
	s.SentMessageIDs = nil
	s.PanelMessageID = 0
}

// SentLastHour reports how many messages went out in the last hour.
func (s *State) SentLastHour() int {
	return len(recentTimes(s.Sent, sentWindow))
}

// Paused reports whether delivery is currently suppressed, and until when.
func (s *State) Paused() (bool, time.Time) {
	if s.PausedUntil.IsZero() || time.Now().After(s.PausedUntil) {
		return false, time.Time{}
	}
	return true, s.PausedUntil
}

// Pause suppresses delivery until the given time.
func (s *State) Pause(until time.Time) { s.PausedUntil = until.UTC() }

// Resume lifts a pause.
func (s *State) Resume() { s.PausedUntil = time.Time{} }

// Path returns where the state is stored.
func (s *State) Path() string { return s.path }

func recentTimes(in []time.Time, window time.Duration) []time.Time {
	cutoff := time.Now().Add(-window)
	out := in[:0:0]
	for _, t := range in {
		if t.After(cutoff) {
			out = append(out, t)
		}
	}
	return out
}
