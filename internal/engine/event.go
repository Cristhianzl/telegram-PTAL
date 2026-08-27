// Package engine compares two consecutive pictures of your pull requests and
// decides what deserves to become a Telegram message.
package engine

import (
	"fmt"
	"time"

	"github.com/cristhianzl/pullalerts/internal/githubapi"
)

// Kind identifies the detected transition.
type Kind string

const (
	KindReviewRequested  Kind = "review_requested"
	KindChangesRequested Kind = "changes_requested"
	KindApproved         Kind = "approved"
	KindChecksFailed     Kind = "checks_failed"
	KindChecksFixed      Kind = "checks_fixed"
	KindConflict         Kind = "conflict"
	KindNewActivity      Kind = "new_activity"
	KindReadyForReview   Kind = "ready_for_review"
	KindAssigned         Kind = "assigned"
	KindMentioned        Kind = "mentioned"
	KindGone             Kind = "gone"
)

// Urgency decides whether the event interrupts you now or waits for the batch.
type Urgency int

const (
	// UrgencyNow sends immediately: something needs you, or something broke.
	UrgencyNow Urgency = iota
	// UrgencyBatch waits for the grouping window.
	UrgencyBatch
)

// Event is a change detected between two cycles.
type Event struct {
	Kind    Kind                    `json:"kind"`
	PR      *githubapi.PullRequest  `json:"pr"`
	Urgency Urgency                 `json:"urgency"`
	Detail  string                  `json:"detail,omitempty"`
	At      time.Time               `json:"at"`
}

// Fingerprint is the event's idempotency key. Restarting the daemon, running
// two syncs, or retrying after a failure never duplicates a message, because
// the key is already recorded in the state.
func (e Event) Fingerprint() string {
	return fmt.Sprintf("%s:%s:%s", e.PR.ID, e.Kind, e.variant())
}

// variant separates repeated occurrences of the same kind that are genuinely
// different events — CI breaking again after being fixed, say, or another
// comment after the previous one.
func (e Event) variant() string {
	switch e.Kind {
	case KindNewActivity:
		return fmt.Sprintf("%d", e.PR.Comments+e.PR.Reviews)
	case KindChecksFailed, KindChecksFixed:
		return e.PR.Checks
	case KindApproved, KindChangesRequested:
		return e.PR.ReviewDecision
	}
	return ""
}

// Title is the message's headline.
func (e Event) Title() string {
	switch e.Kind {
	case KindReviewRequested:
		return "Review requested"
	case KindChangesRequested:
		return "Changes requested"
	case KindApproved:
		return "Pull request approved"
	case KindChecksFailed:
		return "CI broke"
	case KindChecksFixed:
		return "CI is green again"
	case KindConflict:
		return "Merge conflict"
	case KindNewActivity:
		return "New activity"
	case KindReadyForReview:
		return "Ready for review"
	case KindAssigned:
		return "Assigned to you"
	case KindMentioned:
		return "You were mentioned"
	case KindGone:
		return "Closed or merged"
	}
	return string(e.Kind)
}

// Emoji conveys the state at a glance, before the text is read.
func (e Event) Emoji() string {
	switch e.Kind {
	case KindReviewRequested:
		return "👀"
	case KindChangesRequested:
		return "🔁"
	case KindApproved:
		return "✅"
	case KindChecksFailed:
		return "❌"
	case KindChecksFixed:
		return "🟢"
	case KindConflict:
		return "⚠️"
	case KindNewActivity:
		return "💬"
	case KindReadyForReview:
		return "📤"
	case KindAssigned:
		return "📌"
	case KindMentioned:
		return "📣"
	case KindGone:
		return "🏁"
	}
	return "•"
}
