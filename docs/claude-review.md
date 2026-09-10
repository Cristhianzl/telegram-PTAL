# Reviewing pull requests with Claude Code

PTAL can review a pull request for you: tap a button in Telegram, and a few
minutes later the review is posted on GitHub as a comment, with the pull
request approved or changes requested.

It uses the Claude Code CLI already installed and logged in on your machine —
your subscription, not an API key.

## What actually happens

```
Telegram              ptal                          GitHub
   │                    │                              │
   │  tap 🤖 Review     │                              │
   ├───────────────────>│                              │
   │                    │  gh repo clone + pr checkout │
   │                    ├─────────────────────────────>│
   │  "checking out…"   │                              │
   │<───────────────────┤                              │
   │                    │  claude -p  (read-only)      │
   │                    │  ─────────────────┐          │
   │  "reviewing…"      │                   │          │
   │<───────────────────┤  <────────────────┘          │
   │                    │                              │
   │                    │  gh pr review --approve      │
   │                    ├─────────────────────────────>│
   │  ✅ Approved       │                              │
   │<───────────────────┤                              │
```

The checkout is deleted afterwards.

## The division of labour, and why

**Claude reads and reasons. PTAL writes to GitHub.**

Claude runs with an explicit read-only tool list — `Read`, `Grep`, `Glob`, and
a handful of specific `git`/`gh` read subcommands. No `Write`, no `Edit`, no
general `Bash`. It produces the review as text and nothing else.

PTAL then posts the comment and sets the verdict with `gh`.

This is not ceremony. Running an agent with write access over a branch someone
else wrote means their `Makefile`, git hook, or build script executes on your
machine. Restricting the tool list removes that entire class of risk, which is
what makes it safe to review a pull request from a contributor you do not know.

It also matches how the review rules were already written: the `reviewing-code`
skill is explicitly forbidden from running `gh pr review` or `gh pr comment`,
because the human posts the review. Here PTAL is that human's hands.

## Setup

### 1. Claude Code, logged in

```bash
claude --version     # any recent version
```

Log in once interactively if you have not. PTAL uses whatever authentication
the CLI already has — a Max or Pro subscription works, and no
`ANTHROPIC_API_KEY` is needed. If that variable is set in the daemon's
environment, the CLI will bill the API instead; unset it if you do not want
that.

### 2. Review rules

Point PTAL at a directory containing a `.claude/` folder:

```bash
ptal config review-rules-dir ~/.config/ptal/reviewer-rules
```

Whatever lives there — `CLAUDE.md`, `rules/`, `skills/`, `commands/` — is
copied into the checkout, so the review follows your conventions rather than
generic advice. The rules in the pull request itself are replaced by yours; a
branch cannot talk the reviewer into going easy on it.

If you have no rules yet, any `.claude/` directory works, including an empty
one. The review is better with real conventions in it.

### 3. Choose the repositories

```bash
ptal config review-repos acme/api,my-org
```

**Nothing is reviewable until it is on this list.** A bare organization name
enables every repository under it.

Think about this one. Reviewing your own project is very different from
reviewing a public repository where anyone can open a pull request. The
read-only tool list is a strong boundary, but the honest framing is: enabling a
repository means agreeing that its branches get read by an agent on your
machine.

## Using it

**From Telegram** — review requests in enabled repositories arrive with a
`🤖 Review` button. There is also:

```
/review acme/api 412
```

**From the terminal**

```bash
ptal review acme/api 412
ptal review api 412 --dry-run     # print it, publish nothing
```

`--dry-run` ignores the allowlist, because it cannot touch GitHub. It is the
right way to see what a review looks like before enabling a repository.

## Reviewing a whole repository

`review-all` runs the same review over every open pull request:

```bash
ptal review-all acme/api
ptal review-all api --dry-run          # print the plan, review nothing
```

In Telegram:

```
/review-all api
/review-all api if nothing blocks, approve and merge it
```

Everything after the repository name is passed to each review, exactly as with
a single one. **There is no policy baked in** — whether a pull request ends up
approved, blocked, or merged is decided by the instruction you give and the
rules in your `.claude/`.

For a merge to happen, the instruction has to ask for it and the review has to
agree. Claude signals that with a fourth verdict, `approve_and_merge`, and PTAL
performs the merge — the read-only boundary is unchanged.

### What it refuses to merge

A merge is the one action here that cannot be taken back, so PTAL checks
before performing one, rather than leaving it to `gh` to fail:

| Refused when | Why |
|---|---|
| The pull request is closed or already merged | Nothing to do |
| It is a draft | Work in progress |
| It has conflicts | The merge would be wrong even if it succeeded |
| Branch protection blocks it | Required checks or reviews are unmet |
| The branch is behind | It needs updating first |

Each refusal is reported per pull request, so a merge that did not happen is
never silent.

### Pacing

Reviews run **one at a time**. Each drives a Claude Code session, and several
at once would compete for the same subscription and could exhaust it partway
through — leaving a repository half reviewed, which is worse than a slower run
that finishes.

Ten open pull requests is therefore the better part of an hour. `Ctrl+C` stops
after the review in flight rather than losing it, and `REVIEW_BATCH_LIMIT`
(default 20) caps a single run.

Drafts are skipped unless you pass `--drafts`, and authors in `IGNORE_AUTHORS`
are left alone, so a batch does not spend its budget on dependency bumps.

## Steering a review

Anything passed with `-m` is added to the prompt, between the diff command and
the output rules:

```bash
ptal review api 412 -m "focus on the auth changes"
ptal review api 412 -m "this is a hotfix — only flag blockers"
ptal review api 412 -m "the author is new to Go, explain the why"
```

In Telegram, everything after the number:

```
/review api 412 focus on the auth changes
```

`REVIEW_INSTRUCTIONS` adds the same text to every review, and per-review
instructions are appended after it rather than replacing it:

```bash
ptal config review-instructions "Always check for hardcoded credentials."
```

### Instructions change the verdict, and the verdict is an action

This is worth seeing before you rely on it. The same pull request, reviewed
twice:

| | Default | `-m "only report Blockers"` |
|---|---|---|
| Output | 7 findings across 3 severities | one line |
| Verdict | `request_changes` | **`approve`** |

Both are correct — the second was asked to ignore everything below Blocker,
and found no Blockers. But the verdict is what PTAL executes against GitHub,
so an instruction that narrows the review also narrows what can block a merge.

Instructions are placed **before** the output rules on purpose, so they cannot
redefine the verdict format and break publishing. They can still change which
verdict is reached. Use `--dry-run` when trying out a new phrasing.

## The verdict

The review ends with a machine-readable line:

```
VERDICT: approve | comment | request_changes | approve_and_merge
```

which becomes `gh pr review --approve`, `--comment` or `--request-changes`,
followed by `gh pr merge --squash` for the last one.

If that line is missing — the model forgot it, the output was cut — the verdict
falls back to `comment`. A parsing slip can never approve or block a pull
request on its own.

GitHub refuses to let you approve your own pull request. When that happens the
review is posted as a plain comment instead of being lost.

## Settings

| Variable | What it does |
|---|---|
| `REVIEW_REPOS` | Where reviewing is offered. Empty disables it entirely. |
| `REVIEW_RULES_DIR` | Directory holding `.claude/`. Defaults to `reviewer-rules/` beside your config. |
| `REVIEW_MODEL` | Model override, e.g. `opus`. Defaults to whatever the CLI uses. |
| `REVIEW_TIMEOUT` | Cap on one review. Default `15m`. |
| `REVIEW_BATCH_LIMIT` | Most pull requests one `review-all` covers. Default `20`. |

## What it costs

Time, and your Claude subscription's usage. A real review of a medium pull
request takes **4 to 8 minutes** — it reads the diff, opens the files around
it, and thinks. One review runs at a time; tapping the button again while one
is in flight is refused rather than queued, so a burst of taps cannot drain
your limits.

## When it goes wrong

**"Reviewing is not enabled for this repository"** — add it to `REVIEW_REPOS`.

**"review rules not found"** — `REVIEW_RULES_DIR` does not point at a directory
containing `.claude/`.

**The review is generic** — your rules were not picked up. Check that
`.claude/` sits directly inside `REVIEW_RULES_DIR`, and remember that a fresh
checkout is an untrusted workspace: settings-based permissions are ignored
there, which is why the tool list is passed on the command line.

**It timed out** — large pull requests take longer than the 15 minute default.
Raise `REVIEW_TIMEOUT`.

**Nothing was posted** — the review still ran; the failure is reported in
Telegram along with the findings, so the work is not lost.
