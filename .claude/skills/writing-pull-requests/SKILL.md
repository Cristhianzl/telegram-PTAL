---
name: writing-pull-requests
description: Generate conventional-commit PR titles, commit messages, and concise PR descriptions from staged changes or a branch diff. Use when the user asks to create a PR, write a PR title, generate a commit message, or draft a PR description. Output goes directly to the chat — never to a file.
license: MIT
---

# Writing Pull Requests

Produce a PR title, a commit message, and a PR description from the current diff, using a conventional-commit format. Output is **always printed in chat**, never written to a file.

## When to apply

Apply when the user says any of: "create a PR", "PR title", "PR description", "commit message", "title for the commit", "write the PR".

**Skip / push back** when:
- The diff is empty or untracked-only — ask the user what to compare against.
- The diff bundles multiple unrelated changes — propose splitting first; a PR whose title needs "and" is a smell.
- The user already pasted a title/description — don't rewrite unless asked.

## Read first (always)

Before doing anything else, **list `learnings/` and read every file whose name looks relevant** to the current PR. Project-specific scopes, type aliases, vocabulary, or hard constraints live there and override the defaults in this SKILL.md. If a learning conflicts with this file, **the learning wins** — mention it to the user so they can refine it.

If `learnings/` contains only `README.md`, proceed with the defaults below.

## Workflow

1. **Inspect the diff** → `git status` + `git diff --stat <base>..HEAD` (or staged with `--cached`). Identify files added/modified/deleted and the dominant area.
   → verify: you can name the affected module(s) and the dominant intent (feat/fix/refactor/etc.).

2. **Pick the type** from the table below. If two types tie, the change is mixed — flag it and ask whether to split the PR.
   → verify: one type, defensible from the diff alone.

3. **Pick the scope** from folder names, file paths, or the imports that changed most. Single token, lowercase. Optional — omit if the change is repo-wide.
   → verify: scope appears in the diff paths.

4. **Write the summary** in imperative mood, ≤72 chars total title length, no trailing punctuation, first letter capitalized after the colon.
   → verify: `<type>(<scope>): <Summary>` ≤72 chars.

5. **Write the description**: one-sentence Objective, ≤5 bullet Changes, optional Notes for edge cases. Whole description ≤150 words.
   → verify: each bullet starts with a verb; no implementation noise that a reviewer can read from the diff.

6. **Print the output block** (format below) in chat. **Do not write a file.** **Do not run `git commit`, `git add`, or `git push`** — only the human commits.

7. **Capture a learning (final step, mandatory ask).** Before closing the task, ask yourself: *did I encounter a convention, constraint, or trap that wasn't in this SKILL.md or in `references/`?* If yes, append a `learnings/YYYY-MM-DD-slug.md` per `learnings/README.md`. If no, skip — don't write filler.

## Type table

| Type       | Use when                                  |
|------------|-------------------------------------------|
| `feat`     | New user-visible capability               |
| `fix`      | Corrects a defect in existing behavior    |
| `docs`     | Docs / comments only                      |
| `style`    | Formatting, whitespace, no logic change   |
| `refactor` | Restructure, no behavior change           |
| `perf`     | Performance improvement                   |
| `test`     | Adds or updates tests only                |
| `chore`    | Build, CI, deps, tooling                  |

## Output format (print verbatim in chat)

```
TITLE: <type>(<scope>): <Summary>

COMMIT MESSAGE: <type>(<scope>): <Summary>

DESCRIPTION:
## Objective
<one sentence — business outcome, not implementation>

## Changes
- <verb-led bullet>
- <verb-led bullet>

## Notes
<edge cases, limitations, follow-ups — omit the heading if empty>
```

## Hard rules

- **Never** write a `.md` file or any other file containing the PR output. It goes in the chat.
- **Never** run `git commit`, `git add`, `git push`, or any state-changing git command. Only the human commits.
- **Never add AI attribution** — no "🤖 Generated with Claude Code" footer, no session links, no AI badges in titles/descriptions/commits, no `Co-Authored-By` trailer. Output must read as authored by the human.
- **Never** produce a title needing "and" — that means two PRs.
- **Never** invent scope. If the diff doesn't justify a scope, omit it.

## See also

- `../documenting-features/references/communication.md` — answer-first structure (Minto Pyramid); first line = what it does and why it's safe to merge.
- `references/examples.md` — wrong-vs-right titles and descriptions, distilled from common LLM mistakes.
- `references/templates.md` — full template variants (feature, fix, refactor, revert, breaking change).
- `learnings/` — project-specific conventions accumulated over time. **Read before generating.** **Append a new file when you discover a non-obvious convention** (see `learnings/README.md` for the append protocol).
