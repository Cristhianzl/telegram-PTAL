package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Cristhianzl/telegram-PTAL/internal/githubapi"
	"github.com/Cristhianzl/telegram-PTAL/internal/reviewer"
	"github.com/Cristhianzl/telegram-PTAL/internal/runner"
)

const reviewAllUsage = `ptal review-all - review every open pull request in a repository

USAGE
  ptal review-all <repo> [instructions] [--dry-run] [--drafts]

EXAMPLES
  ptal review-all acme/api
  ptal review-all api --dry-run
  ptal review-all api if nothing blocks, approve and merge it
  ptal review-all api only flag blockers, this is a release branch

Everything after the repository name is passed to the review, so the
instruction is what decides whether a pull request is approved, blocked, or
merged. There is no policy baked in.

Each review takes several minutes and they run one at a time, so a repository
with ten open pull requests takes the better part of an hour. REVIEW_BATCH_LIMIT
caps how many one run will cover.

--dry-run reviews nothing; it prints what would be reviewed and what would be
skipped.
`

func cmdReviewAll(args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		fmt.Print(reviewAllUsage)
		return nil
	}

	dryRun, drafts := false, false
	var rest []string
	for _, a := range args {
		switch a {
		case "--dry-run", "-n":
			dryRun = true
		case "--drafts":
			drafts = true
		default:
			rest = append(rest, a)
		}
	}
	if len(rest) == 0 {
		return fmt.Errorf("usage: ptal review-all <repo> [instructions]")
	}
	name, instructions := rest[0], strings.Join(rest[1:], " ")

	cfg, state, err := prepareAllowNoChat()
	if err != nil {
		return err
	}
	repo, err := githubapi.ResolveRepo(name, cfg.WatchRepos)
	if err != nil {
		return err
	}
	if !dryRun && !cfg.ReviewEnabledFor(repo) {
		return fmt.Errorf("reviewing is not enabled for %s\n"+
			"  Add it:  ptal config review-repos %s\n"+
			"  Or see what it would do:  ptal review-all %s --dry-run",
			repo, repo, name)
	}

	// Ctrl+C stops after the review in flight, rather than losing it.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	r := runner.New(cfg, state, newQuietLogger())
	r.Source().EnsureCredential()
	if err := r.Source().Resolve(ctx); err != nil {
		return err
	}

	prs, err := r.Source().RepoPullRequests(ctx, repo, "", 100)
	if err != nil {
		return err
	}
	refs := make([]*reviewer.PullRequestRef, 0, len(prs))
	for _, pr := range prs {
		refs = append(refs, &reviewer.PullRequestRef{
			Number: pr.Number, Title: pr.Title, Author: pr.Author, IsDraft: pr.IsDraft,
		})
	}

	opts := reviewer.BatchOptions{
		Limit:         cfg.ReviewBatchLimit,
		IncludeDrafts: drafts,
		SkipAuthors:   cfg.IgnoreAuthors,
	}
	queued, skipped := reviewer.PlanSize(refs, opts)

	fmt.Printf("%s · %d open · %d to review", repo, len(refs), queued)
	if skipped > 0 {
		fmt.Printf(" · %d skipped", skipped)
	}
	fmt.Println()
	if instructions != "" {
		fmt.Printf("instructions: %s\n", instructions)
	}

	if dryRun {
		return printBatchPlan(refs, opts)
	}
	if queued == 0 {
		return nil
	}
	fmt.Printf("\nabout %d minutes, one at a time. Ctrl+C stops after the current one.\n\n", queued*4)

	rules := cfg.ReviewRulesDir
	if rules == "" {
		rules = reviewer.DefaultRulesDir(cfg.Dir())
	}
	rv := reviewer.New(reviewer.Options{
		RulesDir:     rules,
		Timeout:      cfg.ReviewTimeout,
		Model:        cfg.ReviewModel,
		Instructions: reviewer.CombineInstructions(cfg.ReviewInstructions, instructions),
	})

	last := ""
	result, err := rv.ReviewAll(ctx, repo, refs, opts,
		func(done, total int, pr *reviewer.PullRequestRef, stage string) {
			line := fmt.Sprintf("[%d/%d] #%d %s", done+1, total, pr.Number, truncate(pr.Title, 46))
			if line != last {
				fmt.Println(line)
				last = line
			}
			fmt.Printf("        %s…\n", stage)
		})
	if err != nil {
		return err
	}

	fmt.Printf("\n%s\n", strings.Repeat("─", 72))
	for _, res := range result.Reviewed {
		mark := map[bool]string{true: "merged", false: string(res.Verdict)}[res.Merged]
		fmt.Printf("  #%-6d %-18s %s\n", res.Number, mark, res.CommentURL)
		if res.Verdict.Merges() && !res.Merged {
			fmt.Printf("          not merged: %s\n", res.MergeSkipped)
		}
	}
	for number, why := range result.Failed {
		fmt.Printf("  #%-6d %-18s %s\n", number, "failed", truncate(why, 60))
	}
	fmt.Printf("\n%s", result.Summary())
	return nil
}

func printBatchPlan(refs []*reviewer.PullRequestRef, opts reviewer.BatchOptions) error {
	queue, skipped := reviewer.Plan(refs, opts)

	fmt.Println("\nwould review:")
	for _, pr := range queue {
		fmt.Printf("  #%-6d %-50s @%s\n", pr.Number, truncate(pr.Title, 48), pr.Author)
	}
	if len(skipped) > 0 {
		fmt.Println("\nwould skip:")
		for _, pr := range refs {
			if why, ok := skipped[pr.Number]; ok {
				fmt.Printf("  #%-6d %-50s %s\n", pr.Number, truncate(pr.Title, 48), why)
			}
		}
	}
	fmt.Println("\nnothing was reviewed (--dry-run)")
	return nil
}

// newQuietLogger keeps the runner's own logging out of the command's output,
// which is already reporting progress line by line.
func newQuietLogger() *log.Logger { return log.New(io.Discard, "", 0) }

