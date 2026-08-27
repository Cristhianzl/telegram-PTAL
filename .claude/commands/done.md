---
description: Finalize changes — quick review + check + commit
argument-hint: ["optional commit message"]
allowed-tools: Read, Grep, Bash(git:*)
---

# /done

Provided message: `$ARGUMENTS`.

Chains `/review` (quick) → `/check` → `/commit`. Stops immediately on any failure.

## Quick review (over the diff)

Look for:

- `print()` or `console.log()` in production code
- Hardcoded credentials (tokens, passwords, URLs with auth)
- `.env*` files in the diff
- `dangerouslySetInnerHTML` without sanitizing
- Python functions without type hints
- SQL with f-string instead of bind params

If you find something critical, **STOP** and report.

## Check

Run lint and tests only on the areas with changes (see `/check`).

## Commit

Follow `/commit`. Does not run `git push` (use `/push`).
