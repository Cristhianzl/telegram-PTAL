---
description: What this configuration does and how to drive it
---

# /help

Print this, adapted to what the user asked:

**This `.claude/` reviews pull requests. It does nothing else** — it cannot edit files, commit,
push, or post to GitHub. Those are blocked by `settings.json` and `hooks/block-mutations.sh`.

| Command | What it does |
|---------|--------------|
| `/review [base]` | Reviews the working tree, the staged diff, or the branch against `base`. |
| `/review-pr <n>` | Reviews GitHub PR number `n` via `gh`. |
| `/dual-review` | Two independent reviewers on the same diff, then converge. Both must approve. |
| `/help` | This message. |

**The ruleset** lives in `skills/reviewing-code/`: `SKILL.md` is the workflow and the severity
model; `references/` holds the depth — `security-checks.md`, `structural-checks.md`,
`correctness-checks.md`, `grep-recipes.md`, `output-format.md`, `checklist.md`.

**To teach it your project:** drop a dated note in `skills/reviewing-code/learnings/`. A learning
overrides the defaults — the agent reads it first. That's where project-specific severity
adjustments, banned patterns, and recurring violations belong.

**The output is a chat message** you copy and paste into GitHub yourself. That is deliberate: a
review posted by a bot gets skimmed, and an agent that can post can also post something wrong.
