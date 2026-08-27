---
description: Review the current diff (working tree, staged, or a branch against its base)
allowed-tools: Read, Grep, Glob, Bash(git status:*), Bash(git diff:*), Bash(git log:*), Bash(git show:*), Bash(git merge-base:*), Bash(git rev-parse:*)
---

# /review

Review the current changes and output a review comment in the chat.

## Pick the diff

Use the first that produces changes, and say which one you used:

1. `$ARGUMENTS` if the user named a base (`/review main`, `/review origin/develop`) →
   `git diff $(git merge-base HEAD <base>)...HEAD`
2. Staged changes → `git diff --cached`
3. Working tree → `git diff HEAD`
4. The branch against its upstream → `git diff @{u}...HEAD`

If none produce changes, say so and stop. Do not review a clean tree.

## Review

Follow `skills/reviewing-code/SKILL.md` end to end: read the diff completely first, then apply the
lenses in order — security, comprehension, structural, platform, correctness, tests. Read the
`references/` that match what the diff touches, and any relevant `learnings/`.

Read the full file around a changed hunk when the diff alone can't tell you whether a finding is
real. A finding you can't anchor to `file:line` is not ready to report.

## Output

Render per `skills/reviewing-code/references/output-format.md` and run the copy-paste safety check
before printing. Print it in the chat — never write it to a file, never post it.
