package githubapi

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Bucket is the reason a pull request counts as yours. Every bucket is
// centered on the authenticated user — PTAL never watches a whole
// repository, only what connects back to you.
type Bucket string

const (
	// BucketReviewRequested is the real inbox: someone asked for your review.
	BucketReviewRequested Bucket = "review_requested"
	// BucketAuthored are the pull requests you opened.
	BucketAuthored Bucket = "authored"
	// BucketAssigned are pull requests assigned to you, which do not always
	// overlap with review requests.
	BucketAssigned Bucket = "assigned"
	// BucketMentioned are pull requests where someone @-mentioned you.
	BucketMentioned Bucket = "mentioned"
	// BucketReviewed are pull requests you already reviewed that are still open.
	BucketReviewed Bucket = "reviewed"
)

// AllBuckets is ordered by priority: the first bucket a pull request belongs
// to decides how it is presented.
var AllBuckets = []Bucket{
	BucketReviewRequested,
	BucketAuthored,
	BucketAssigned,
	BucketMentioned,
	BucketReviewed,
}

// Label returns the human-readable name of the bucket.
func (b Bucket) Label() string {
	switch b {
	case BucketReviewRequested:
		return "Waiting for your review"
	case BucketAuthored:
		return "Your open pull requests"
	case BucketAssigned:
		return "Assigned to you"
	case BucketMentioned:
		return "You were mentioned"
	case BucketReviewed:
		return "You already reviewed"
	}
	return string(b)
}

// query builds the bucket's search qualifier.
//
// includeTeam controls a distinction GitHub makes that changes everything on
// large teams: `review-requested:` also matches requests sent to your team,
// while `user-review-requested:` matches only when someone asked for you by
// name. On one real account the difference was 61 pull requests versus 1.
func (b Bucket) query(includeTeam bool) string {
	switch b {
	case BucketReviewRequested:
		if includeTeam {
			return "is:open is:pr review-requested:@me archived:false"
		}
		return "is:open is:pr user-review-requested:@me archived:false"
	case BucketAuthored:
		return "is:open is:pr author:@me archived:false"
	case BucketAssigned:
		return "is:open is:pr assignee:@me archived:false"
	case BucketMentioned:
		return "is:open is:pr mentions:@me archived:false"
	case BucketReviewed:
		return "is:open is:pr reviewed-by:@me archived:false"
	}
	return ""
}

// UpdatedSinceQualifier builds the recent-activity search filter.
//
// Filtering at the source rather than discarding afterwards is what keeps the
// result small: pull requests abandoned years ago are still technically open
// and, without this, they drown out the few that actually need you.
func UpdatedSinceQualifier(days int, now time.Time) string {
	if days <= 0 {
		return ""
	}
	return " updated:>=" + now.AddDate(0, 0, -days).Format("2006-01-02")
}

// PullRequest is the normalized shape of a pull request, carrying only what
// the event engine compares between one cycle and the next.
type PullRequest struct {
	ID        string    `json:"id"`
	Repo      string    `json:"repo"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Author    string    `json:"author"`
	IsDraft   bool      `json:"is_draft"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// ReviewDecision is APPROVED, CHANGES_REQUESTED, REVIEW_REQUIRED or empty.
	ReviewDecision string `json:"review_decision"`
	// Mergeable is MERGEABLE, CONFLICTING or UNKNOWN.
	Mergeable string `json:"mergeable"`
	// Checks is SUCCESS, FAILURE, PENDING, ERROR, EXPECTED or empty (no CI).
	Checks string `json:"checks"`

	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
	Comments  int `json:"comments"`
	Reviews   int `json:"reviews"`

	// Buckets are every reason this pull request is yours.
	Buckets []Bucket `json:"buckets"`
}

// Slug identifies the pull request readably, as "owner/repo#123".
func (p PullRequest) Slug() string {
	return fmt.Sprintf("%s#%d", p.Repo, p.Number)
}

// InBucket reports whether the pull request belongs to the given bucket.
func (p PullRequest) InBucket(b Bucket) bool {
	for _, x := range p.Buckets {
		if x == b {
			return true
		}
	}
	return false
}

// PrimaryBucket returns the pull request's highest-priority bucket.
func (p PullRequest) PrimaryBucket() Bucket {
	for _, b := range AllBuckets {
		if p.InBucket(b) {
			return b
		}
	}
	return ""
}

// ChecksSymbol renders the CI state as a short message symbol.
func (p PullRequest) ChecksSymbol() string {
	switch p.Checks {
	case "SUCCESS":
		return "✅ CI"
	case "FAILURE", "ERROR":
		return "❌ CI"
	case "PENDING", "EXPECTED":
		return "🟡 CI"
	}
	return ""
}

// Snapshot is the complete picture of your pull requests at one instant.
type Snapshot struct {
	Viewer    string                  `json:"viewer"`
	PRs       map[string]*PullRequest `json:"prs"`
	FetchedAt time.Time               `json:"fetched_at"`

	RateRemaining int `json:"rate_remaining"`
	RateCost      int `json:"rate_cost"`
}

// Sorted returns the snapshot's pull requests in a stable order: by bucket
// priority and, within a bucket, oldest first — whatever has been waiting
// longest shows up at the top.
func (s *Snapshot) Sorted() []*PullRequest {
	out := make([]*PullRequest, 0, len(s.PRs))
	for _, pr := range s.PRs {
		out = append(out, pr)
	}
	rank := map[Bucket]int{}
	for i, b := range AllBuckets {
		rank[b] = i
	}
	sort.Slice(out, func(i, j int) bool {
		ri, rj := rank[out[i].PrimaryBucket()], rank[out[j].PrimaryBucket()]
		if ri != rj {
			return ri < rj
		}
		if !out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].UpdatedAt.Before(out[j].UpdatedAt)
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// InBucket returns the snapshot's pull requests in the given bucket.
func (s *Snapshot) InBucket(b Bucket) []*PullRequest {
	var out []*PullRequest
	for _, pr := range s.Sorted() {
		if pr.InBucket(b) {
			out = append(out, pr)
		}
	}
	return out
}

// Counts summarizes how many pull requests sit in each bucket.
func (s *Snapshot) Counts() map[Bucket]int {
	out := map[Bucket]int{}
	for _, pr := range s.PRs {
		for _, b := range pr.Buckets {
			out[b]++
		}
	}
	return out
}

// AuthError signals an invalid credential: the daemon must stop and tell the
// user rather than retry forever.
type AuthError struct{ Msg string }

func (e *AuthError) Error() string { return "invalid GitHub credential: " + e.Msg }

// PolicyError means an organization or enterprise policy rejected the token
// even though the resource is publicly readable. It is the signal for
// PTAL to fall back to public search.
type PolicyError struct{ Msg string }

func (e *PolicyError) Error() string {
	return "organization policy rejected the token: " + e.Msg
}

// RateLimitError carries how long to wait before the next attempt.
type RateLimitError struct {
	RetryAfter time.Duration
	Msg        string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("GitHub rate limit: %s (wait %s)", e.Msg, e.RetryAfter)
}

// GraphQLErrors groups the errors the API returns inside a 200 response.
type GraphQLErrors struct{ Messages []string }

func (e *GraphQLErrors) Error() string {
	return "GraphQL error: " + strings.Join(e.Messages, "; ")
}
