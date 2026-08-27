package engine

import (
	"time"

	"github.com/cristhianzl/pullalerts/internal/githubapi"
)

// Tracked is a pull request kept between cycles, with the little extra
// metadata the engine needs to remember.
type Tracked struct {
	githubapi.PullRequest

	// MissingCount counts consecutive cycles where the pull request did not
	// appear in the search. GitHub's index is eventually consistent, so a
	// single absence does not mean the pull request was closed.
	MissingCount int `json:"missing_count"`
	// FirstSeenAt allows warning about pull requests idle for a long time.
	FirstSeenAt time.Time `json:"first_seen_at"`
}

// Options tunes what counts as an event.
type Options struct {
	// IgnoreDrafts suppresses review requests on draft pull requests.
	IgnoreDrafts bool
	// FirstRun populates the state without emitting anything, avoiding the
	// initial flood of dozens of messages at minute zero.
	FirstRun bool
	// ModeChanged means the previous picture came from a different data
	// source. Public search carries no review count, CI state or review
	// decision; switching to GraphQL fills all of them at once, and each
	// jump would look like a real change. The switching cycle only
	// re-synchronizes.
	ModeChanged bool
}

// missingThreshold is how many consecutive cycles a pull request must be
// absent from the search before we treat it as genuinely closed or merged.
const missingThreshold = 2

// Diff compares the previous state with the new picture and returns the
// detected events along with the updated state.
func Diff(prev map[string]*Tracked, snap *githubapi.Snapshot, opts Options) ([]Event, map[string]*Tracked) {
	now := snap.FetchedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}

	next := make(map[string]*Tracked, len(snap.PRs))
	var events []Event

	silent := opts.FirstRun || opts.ModeChanged

	emit := func(kind Kind, pr *githubapi.PullRequest, urgency Urgency, detail string) {
		if silent {
			return
		}
		events = append(events, Event{
			Kind: kind, PR: pr, Urgency: urgency, Detail: detail, At: now,
		})
	}

	for id, pr := range snap.PRs {
		old, existed := prev[id]

		tracked := &Tracked{PullRequest: *pr, FirstSeenAt: now}
		if existed {
			tracked.FirstSeenAt = old.FirstSeenAt
		}
		next[id] = tracked

		if !existed {
			// The pull request just entered your radar.
			switch {
			case pr.InBucket(githubapi.BucketReviewRequested) && !(opts.IgnoreDrafts && pr.IsDraft):
				emit(KindReviewRequested, pr, UrgencyNow, "")
			case pr.InBucket(githubapi.BucketAssigned):
				emit(KindAssigned, pr, UrgencyNow, "")
			case pr.InBucket(githubapi.BucketMentioned):
				emit(KindMentioned, pr, UrgencyNow, "")
			}
			// Pull requests you opened yourself raise no entry alert: you
			// just opened them.
			continue
		}

		// Entered a bucket it was not in before.
		if pr.InBucket(githubapi.BucketReviewRequested) &&
			!old.InBucket(githubapi.BucketReviewRequested) &&
			!(opts.IgnoreDrafts && pr.IsDraft) {
			emit(KindReviewRequested, pr, UrgencyNow, "")
		}
		if pr.InBucket(githubapi.BucketAssigned) && !old.InBucket(githubapi.BucketAssigned) {
			emit(KindAssigned, pr, UrgencyNow, "")
		}
		if pr.InBucket(githubapi.BucketMentioned) && !old.InBucket(githubapi.BucketMentioned) {
			emit(KindMentioned, pr, UrgencyNow, "")
		}

		// Left draft state and is now waiting on you.
		if old.IsDraft && !pr.IsDraft && pr.InBucket(githubapi.BucketReviewRequested) {
			emit(KindReadyForReview, pr, UrgencyBatch, "")
		}

		mine := pr.InBucket(githubapi.BucketAuthored)

		// Review decision changed. Only interesting on your own pull
		// requests; on someone else's it is noise.
		//
		// An empty previous value means the field was never observed — when
		// switching from public search to GraphQL, for instance, they all
		// fill at once. That is a first reading, not a change, and must not
		// turn into dozens of alerts.
		if mine && old.ReviewDecision != "" && pr.ReviewDecision != old.ReviewDecision {
			switch pr.ReviewDecision {
			case "APPROVED":
				emit(KindApproved, pr, UrgencyBatch, "")
			case "CHANGES_REQUESTED":
				emit(KindChangesRequested, pr, UrgencyNow, "")
			}
		}

		// CI state, with the same caveat about the first observation.
		if mine && old.Checks != "" && pr.Checks != old.Checks {
			switch pr.Checks {
			case "FAILURE", "ERROR":
				emit(KindChecksFailed, pr, UrgencyNow, "")
			case "SUCCESS":
				if old.Checks == "FAILURE" || old.Checks == "ERROR" {
					emit(KindChecksFixed, pr, UrgencyBatch, "")
				}
			}
		}

		// A conflict appeared.
		if mine && old.Mergeable != "" && pr.Mergeable == "CONFLICTING" &&
			old.Mergeable != "CONFLICTING" {
			emit(KindConflict, pr, UrgencyBatch, "")
		}

		// New comments or reviews. The counter only grows, so comparing is
		// safe even if GitHub reorders the list.
		if newer := (pr.Comments + pr.Reviews) - (old.Comments + old.Reviews); newer > 0 {
			detail := "1 new comment"
			if newer > 1 {
				detail = plural(newer, "new comment", "new comments")
			}
			emit(KindNewActivity, pr, UrgencyBatch, detail)
		}
	}

	// Pull requests that disappeared from the search.
	for id, old := range prev {
		if _, still := snap.PRs[id]; still {
			continue
		}
		old.MissingCount++
		if old.MissingCount < missingThreshold {
			// Could still be search index lag: hold it one more cycle.
			next[id] = old
			continue
		}
		pr := old.PullRequest
		emit(KindGone, &pr, UrgencyBatch, "")
	}

	return events, next
}

func plural(n int, one, many string) string {
	word := many
	if n == 1 {
		word = one
	}
	return itoa(n) + " " + word
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
