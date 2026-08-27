package engine

import (
	"testing"
	"time"

	"github.com/Cristhianzl/telegram-PTAL/internal/githubapi"
)

// fakeState implements Seen for the gate tests.
type fakeState struct {
	seen map[string]bool
	sent int
}

func newFakeState() *fakeState { return &fakeState{seen: map[string]bool{}} }

func (f *fakeState) Seen(fp string) bool  { return f.seen[fp] }
func (f *fakeState) MarkSeen(fp string)   { f.seen[fp] = true }
func (f *fakeState) SentLastHour() int    { return f.sent }

func event(kind Kind, urgency Urgency, id string) Event {
	return Event{
		Kind:    kind,
		PR:      &githubapi.PullRequest{ID: id, Repo: "acme/app", Number: 1},
		Urgency: urgency,
	}
}

func TestGateSeparatesUrgentFromBatched(t *testing.T) {
	state := newFakeState()
	events := []Event{
		event(KindReviewRequested, UrgencyNow, "a"),
		event(KindApproved, UrgencyBatch, "b"),
		event(KindChecksFailed, UrgencyNow, "c"),
	}

	urgent, batched := Gate(events, state, GateOptions{MaxPerHour: 100})

	if len(urgent.Events) != 2 {
		t.Errorf("urgent = %d, want 2", len(urgent.Events))
	}
	if len(batched.Events) != 1 {
		t.Errorf("batched = %d, want 1", len(batched.Events))
	}
}

// Restarting the daemon or running an extra cycle must not resend anything.
func TestGateNeverRepeatsAnEvent(t *testing.T) {
	state := newFakeState()
	events := []Event{event(KindReviewRequested, UrgencyNow, "a")}

	urgent, _ := Gate(events, state, GateOptions{MaxPerHour: 100})
	if len(urgent.Events) != 1 {
		t.Fatalf("first pass should deliver 1, delivered %d", len(urgent.Events))
	}

	urgent, _ = Gate(events, state, GateOptions{MaxPerHour: 100})
	if len(urgent.Events) != 0 {
		t.Errorf("second pass should deliver 0, delivered %d", len(urgent.Events))
	}
}

func TestGateRespectsHourlyCeiling(t *testing.T) {
	state := newFakeState()
	state.sent = 9 // nine already went out in the last hour

	var events []Event
	for _, id := range []string{"a", "b", "c", "d"} {
		events = append(events, event(KindApproved, UrgencyBatch, id))
	}

	_, batched := Gate(events, state, GateOptions{MaxPerHour: 10})

	if len(batched.Events) != 1 {
		t.Errorf("should deliver only 1 (ceiling 10, 9 already sent), delivered %d", len(batched.Events))
	}
	if batched.Dropped != 3 {
		t.Errorf("dropped = %d, want 3", batched.Dropped)
	}
}

func TestQuietHoursAcrossMidnight(t *testing.T) {
	q := ParseQuietHours("23:00-08:00")

	cases := []struct {
		hour int
		want bool
	}{
		{23, true}, {2, true}, {7, true},
		{8, false}, {12, false}, {22, false},
	}
	for _, c := range cases {
		at := time.Date(2026, 8, 27, c.hour, 30, 0, 0, time.UTC)
		if got := q.Active(at); got != c.want {
			t.Errorf("%02d:30 -> quiet=%v, want %v", c.hour, got, c.want)
		}
	}
}

// A malformed preference must not bring the daemon down.
func TestInvalidQuietHoursDisablesSilently(t *testing.T) {
	for _, s := range []string{"", "lixo", "25:00-08:00", "23:00"} {
		if ParseQuietHours(s).Active(time.Now()) {
			t.Errorf("%q should disable quiet hours", s)
		}
	}
}

func TestQuietHoursMarksBatchSilent(t *testing.T) {
	state := newFakeState()
	night := time.Date(2026, 8, 27, 2, 0, 0, 0, time.UTC)

	urgent, _ := Gate([]Event{event(KindReviewRequested, UrgencyNow, "a")}, state,
		GateOptions{Quiet: ParseQuietHours("23:00-08:00"), MaxPerHour: 100, Now: night})

	if !urgent.Silent {
		t.Error("during quiet hours the message should arrive without a sound")
	}
	if len(urgent.Events) != 1 {
		t.Error("urgent events must still be delivered, just silently")
	}
}
