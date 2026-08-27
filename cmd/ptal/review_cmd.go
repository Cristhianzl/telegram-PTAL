package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Cristhianzl/telegram-PTAL/internal/config"
	"github.com/Cristhianzl/telegram-PTAL/internal/githubapi"
	"github.com/Cristhianzl/telegram-PTAL/internal/reviewer"
)

const reviewUsage = `ptal review - review a pull request with Claude Code

USAGE
  ptal review <repo> <number> [-m "instructions"] [--dry-run]

EXAMPLES
  ptal review acme/api 412
  ptal review api 412 --dry-run                  # print it, publish nothing
  ptal review api 412 -m "focus on the auth changes"
  ptal review api 412 -m "this is a hotfix, only flag blockers"

Anything passed with -m is added to the standard prompt, so the review still
follows the project rules and still produces a verdict. Use it to steer what
gets looked at, not to replace the rules.

REVIEW_INSTRUCTIONS adds the same thing to every review.

The review runs Claude Code headlessly against a fresh checkout, using the
rules in REVIEW_RULES_DIR. Claude is given read-only tools; publishing the
comment and setting the verdict is done by ptal itself.

Reviewing is only offered in repositories listed in REVIEW_REPOS.
`

func cmdReview(args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		fmt.Print(reviewUsage)
		return nil
	}

	dryRun := false
	instructions := ""
	var positional []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dry-run", "-n":
			dryRun = true
		case "-m", "--message", "--instructions":
			if i+1 >= len(args) {
				return fmt.Errorf("-m needs the instructions after it, in quotes:\n" +
					`  ptal review acme/api 412 -m "focus on the auth changes"`)
			}
			i++
			instructions = args[i]
		default:
			positional = append(positional, args[i])
		}
	}
	if len(positional) < 2 {
		return fmt.Errorf("usage: ptal review <repo> <number> [--dry-run]")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	repo, err := githubapi.ResolveRepo(positional[0], cfg.WatchRepos)
	if err != nil {
		return err
	}
	number, err := strconv.Atoi(positional[1])
	if err != nil {
		return fmt.Errorf("not a pull request number: %q", positional[1])
	}

	// The allowlist is enforced here too, not only where the button is
	// drawn: the command is another way in.
	if !dryRun && !cfg.ReviewEnabledFor(repo) {
		return fmt.Errorf("reviewing is not enabled for %s\n"+
			"  Add it: ptal config review-repos %s\n"+
			"  Or preview without publishing: ptal review %s %d --dry-run",
			repo, repo, positional[0], number)
	}

	rules := cfg.ReviewRulesDir
	if rules == "" {
		rules = reviewer.DefaultRulesDir(cfg.Dir())
	}
	if _, err := os.Stat(rules); err != nil {
		return fmt.Errorf("review rules not found at %s\n"+
			"  Point REVIEW_RULES_DIR at a directory containing .claude/", rules)
	}

	rv := reviewer.New(reviewer.Options{
		RulesDir:     rules,
		Timeout:      cfg.ReviewTimeout,
		Model:        cfg.ReviewModel,
		DryRun:       dryRun,
		Instructions: reviewer.CombineInstructions(cfg.ReviewInstructions, instructions),
	})

	fmt.Printf("Reviewing %s#%d\n", repo, number)
	if instructions != "" {
		fmt.Printf("  with: %s\n", instructions)
	}
	result, err := rv.Review(context.Background(), repo, number, func(stage string) {
		fmt.Printf("  %s…\n", stage)
	})
	if err != nil {
		return err
	}

	fmt.Printf("\n%s\n\n", strings.Repeat("─", 72))
	fmt.Println(result.Body)
	fmt.Printf("\n%s\n", strings.Repeat("─", 72))
	fmt.Printf("verdict:  %s\n", result.Verdict)
	fmt.Printf("took:     %s\n", result.Duration.Round(time.Second))
	if result.Posted {
		fmt.Printf("posted:   %s\n", result.CommentURL)
	} else if dryRun {
		fmt.Println("posted:   no (dry run)")
	}
	return nil
}
