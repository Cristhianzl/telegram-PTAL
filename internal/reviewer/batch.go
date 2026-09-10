package reviewer

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// BatchOptions configures a review of every open pull request in a repository.
type BatchOptions struct {
	// Limit caps how many pull requests are reviewed in one run.
	//
	// Each review takes minutes, so an unbounded run against a busy
	// repository would occupy the machine for hours and spend a
	// correspondingly large amount of the subscription.
	Limit int
	// IncludeDrafts reviews drafts too. Off by default: a draft is work in
	// progress, and reviewing it is usually noise for its author.
	IncludeDrafts bool
	// SkipAuthors are logins whose pull requests are left alone, so a batch
	// does not spend its budget on bot-generated dependency bumps.
	SkipAuthors []string
}

// DefaultBatchLimit bounds an unqualified run.
const DefaultBatchLimit = 20

// BatchProgress reports movement through the batch, so a run measured in tens
// of minutes does not look like it has hung.
type BatchProgress func(done, total int, current *PullRequestRef, stage string)

// PullRequestRef identifies one pull request in a batch.
type PullRequestRef struct {
	Number  int
	Title   string
	Author  string
	IsDraft bool
}

// BatchResult is the outcome of reviewing a whole repository.
type BatchResult struct {
	Repo     string
	Reviewed []*Result
	// Skipped records pull requests that were not reviewed, and why.
	Skipped  map[int]string
	Failed   map[int]string
	Duration time.Duration
}

// Counts summarizes a batch by what happened to each pull request.
func (b *BatchResult) Counts() (approved, merged, changes, commented int) {
	for _, r := range b.Reviewed {
		if r.Merged {
			merged++
		}
		switch r.Verdict {
		case VerdictApprove, VerdictApproveAndMerge:
			approved++
		case VerdictRequestChanges:
			changes++
		case VerdictComment:
			commented++
		}
	}
	return
}

// ReviewAll reviews every open pull request in a repository, one at a time.
//
// Sequential rather than parallel, deliberately: each review drives a Claude
// Code session, and several at once would compete for the same subscription
// and could exhaust it partway through — leaving a repository half reviewed,
// which is worse than a slower run that finishes.
func (r *Reviewer) ReviewAll(
	ctx context.Context,
	repo string,
	prs []*PullRequestRef,
	opts BatchOptions,
	progress BatchProgress,
) (*BatchResult, error) {
	if progress == nil {
		progress = func(int, int, *PullRequestRef, string) {}
	}
	if opts.Limit <= 0 {
		opts.Limit = DefaultBatchLimit
	}

	started := time.Now()
	out := &BatchResult{
		Repo:    repo,
		Skipped: map[int]string{},
		Failed:  map[int]string{},
	}

	queue, skipped := planBatch(prs, opts)
	out.Skipped = skipped

	for i, pr := range queue {
		// A cancelled run stops cleanly and reports what it managed, rather
		// than losing the reviews already posted.
		if ctx.Err() != nil {
			for _, rest := range queue[i:] {
				out.Skipped[rest.Number] = "the run was stopped"
			}
			break
		}

		result, err := r.Review(ctx, repo, pr.Number, func(stage string) {
			progress(i, len(queue), pr, stage)
		})
		switch {
		case err != nil && result == nil:
			out.Failed[pr.Number] = err.Error()
		case err != nil:
			// The review ran but publishing failed; keep it, the findings
			// still matter.
			out.Failed[pr.Number] = err.Error()
			out.Reviewed = append(out.Reviewed, result)
		default:
			out.Reviewed = append(out.Reviewed, result)
		}
	}

	out.Duration = time.Since(started)
	return out, nil
}

// planBatch decides which pull requests are reviewed and which are left out,
// separated from the run so the decisions can be asserted without driving
// Claude.
func planBatch(prs []*PullRequestRef, opts BatchOptions) ([]*PullRequestRef, map[int]string) {
	if opts.Limit <= 0 {
		opts.Limit = DefaultBatchLimit
	}
	queue := make([]*PullRequestRef, 0, len(prs))
	skipped := map[int]string{}

	for _, pr := range prs {
		if skip, why := shouldSkip(pr, opts); skip {
			skipped[pr.Number] = why
			continue
		}
		if len(queue) >= opts.Limit {
			skipped[pr.Number] = fmt.Sprintf("over the limit of %d for one run", opts.Limit)
			continue
		}
		queue = append(queue, pr)
	}
	return queue, skipped
}

func shouldSkip(pr *PullRequestRef, opts BatchOptions) (bool, string) {
	if pr.IsDraft && !opts.IncludeDrafts {
		return true, "draft"
	}
	for _, author := range opts.SkipAuthors {
		if strings.EqualFold(author, pr.Author) {
			return true, "author is on the skip list"
		}
	}
	return false, ""
}

// Summary renders the batch as a short report.
func (b *BatchResult) Summary() string {
	approved, merged, changes, commented := b.Counts()

	var s strings.Builder
	fmt.Fprintf(&s, "%s · %d reviewed in %s\n",
		b.Repo, len(b.Reviewed), b.Duration.Round(time.Second))

	if merged > 0 {
		fmt.Fprintf(&s, "  merged:           %d\n", merged)
	}
	if approved-merged > 0 {
		fmt.Fprintf(&s, "  approved:         %d\n", approved-merged)
	}
	if changes > 0 {
		fmt.Fprintf(&s, "  changes requested: %d\n", changes)
	}
	if commented > 0 {
		fmt.Fprintf(&s, "  commented:        %d\n", commented)
	}
	if len(b.Skipped) > 0 {
		fmt.Fprintf(&s, "  skipped:          %d\n", len(b.Skipped))
	}
	if len(b.Failed) > 0 {
		fmt.Fprintf(&s, "  failed:           %d\n", len(b.Failed))
	}
	return s.String()
}

// Plan reports which pull requests a batch would review and which it would
// skip, with the reason, without running anything.
func Plan(prs []*PullRequestRef, opts BatchOptions) ([]*PullRequestRef, map[int]string) {
	return planBatch(prs, opts)
}

// PlanSize reports how many pull requests a batch would review and how many
// it would skip, without running anything. It exists so a caller can tell
// someone what is about to happen before it starts.
func PlanSize(prs []*PullRequestRef, opts BatchOptions) (queued, skipped int) {
	q, s := planBatch(prs, opts)
	return len(q), len(s)
}
