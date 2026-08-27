package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// QuietHours is a range of the day where messages arrive without a sound.
// It crosses midnight naturally ("23:00-08:00").
type QuietHours struct {
	start, end int // minutos desde a meia-noite
	valid      bool
}

// ParseQuietHours understands "23:00-08:00". An empty or invalid string
// disables quiet hours rather than failing: this is a preference, not a
// credential, and is not worth bringing the daemon down for.
func ParseQuietHours(s string) QuietHours {
	s = strings.TrimSpace(s)
	if s == "" {
		return QuietHours{}
	}
	from, to, ok := strings.Cut(s, "-")
	if !ok {
		return QuietHours{}
	}
	start, err1 := parseClock(from)
	end, err2 := parseClock(to)
	if err1 != nil || err2 != nil {
		return QuietHours{}
	}
	return QuietHours{start: start, end: end, valid: true}
}

// Active reports whether the given time falls inside the quiet range.
func (q QuietHours) Active(t time.Time) bool {
	if !q.valid {
		return false
	}
	m := t.Hour()*60 + t.Minute()
	if q.start <= q.end {
		return m >= q.start && m < q.end
	}
	// Range crossing midnight.
	return m >= q.start || m < q.end
}

func parseClock(s string) (int, error) {
	h, m, ok := strings.Cut(strings.TrimSpace(s), ":")
	if !ok {
		return 0, fmt.Errorf("invalid format")
	}
	hh, err := strconv.Atoi(h)
	if err != nil || hh < 0 || hh > 23 {
		return 0, fmt.Errorf("invalid hour")
	}
	mm, err := strconv.Atoi(m)
	if err != nil || mm < 0 || mm > 59 {
		return 0, fmt.Errorf("invalid minute")
	}
	return hh*60 + mm, nil
}

// Seen is the minimal interface the gate needs from persisted state.
type Seen interface {
	Seen(fingerprint string) bool
	MarkSeen(fingerprint string)
	SentLastHour() int
}

// Batch is a set of events ready to become one message.
type Batch struct {
	Events []Event
	// Silent asks for delivery without a sound.
	Silent bool
	// Dropped counts events suppressed by the hourly ceiling.
	Dropped int
}

// Empty reports whether there is nothing to send.
func (b Batch) Empty() bool { return len(b.Events) == 0 && b.Dropped == 0 }

// GateOptions configures the anti-spam guards.
type GateOptions struct {
	Quiet      QuietHours
	MaxPerHour int
	Now        time.Time
}

// Gate applies, in order: fingerprint deduplication, the hourly message
// ceiling, and the split between what interrupts now and what waits.
//
// It returns the urgent batch and the grouped batch. Events already sent at
// any point are dropped silently — that is what makes it safe to restart the
// daemon or run an extra sync.
func Gate(events []Event, state Seen, opts GateOptions) (urgent, batched Batch) {
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	silent := opts.Quiet.Active(now)

	budget := opts.MaxPerHour - state.SentLastHour()
	if opts.MaxPerHour <= 0 {
		budget = len(events)
	}

	for _, e := range events {
		fp := e.Fingerprint()
		if state.Seen(fp) {
			continue
		}
		state.MarkSeen(fp)

		if budget <= 0 {
			if e.Urgency == UrgencyNow {
				urgent.Dropped++
			} else {
				batched.Dropped++
			}
			continue
		}
		budget--

		if e.Urgency == UrgencyNow {
			urgent.Events = append(urgent.Events, e)
		} else {
			batched.Events = append(batched.Events, e)
		}
	}

	// During quiet hours nothing makes a sound; urgent events still arrive,
	// just silently.
	urgent.Silent = silent
	batched.Silent = silent
	return urgent, batched
}
