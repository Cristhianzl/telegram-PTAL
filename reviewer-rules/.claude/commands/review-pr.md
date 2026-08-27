---
description: Review a GitHub pull request by number
allowed-tools: Read, Grep, Glob, Bash(gh pr view:*), Bash(gh pr diff:*), Bash(gh pr checks:*), Bash(git fetch:*), Bash(git diff:*), Bash(git log:*), Bash(git show:*)
---

# /review-pr

Review the pull request in `$ARGUMENTS` (a PR number, or a full PR URL).

## Gather

- `gh pr view <n>` — title, description, author, state, linked issues.
- `gh pr diff <n>` — the diff.
- `gh pr checks <n>` — CI state. Red CI is context for the review, not a finding by itself.

If `$ARGUMENTS` is empty, ask which PR. Do not guess.

## Review

Follow `skills/reviewing-code/SKILL.md`. Compare the PR description against the diff first — a
description that claims something the diff doesn't do is a finding on its own.

Read the surrounding file, not just the hunk, whenever the diff is not enough to decide whether a
finding is real. `gh pr diff` shows the change; it does not show what the change breaks.

## Output

Render per `skills/reviewing-code/references/output-format.md`, run the copy-paste safety check,
and print it in the chat. **Never** run `gh pr review` or `gh pr comment` — the human posts it.
