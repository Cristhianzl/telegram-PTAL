# Contributing

Thanks for considering a contribution. This is a small project — the bar is
that changes are tested and that the tool never becomes noisier by accident.

## Getting set up

```bash
git clone https://github.com/Cristhianzl/telegram-PTAL
cd ptal
cp .env.example .env
make doctor    # tells you exactly what is still missing
```

You need Go 1.24 or newer. No other dependencies: the project uses only the
standard library, and that is deliberate — it keeps `go build` offline and the
binary free of transitive risk.

## Before opening a pull request

```bash
make check     # vet plus tests with the race detector
```

CI runs the same gates on Linux and macOS.

## What a change needs

- **A test.** Every bug fix starts with a test that fails before the fix. Every
  new behavior gets one covering the success and the failure path.
- **English.** Code, comments, commit messages, and documentation, regardless
  of the language of the discussion.
- **Comments that say why, not what.** If a line needs a comment to explain
  what it does, rename something instead. Comments earn their place by
  recording a reason that is not visible in the code — a GitHub API quirk, a
  rate limit, a failure mode observed in production.
- **Conventional commits.** `feat:`, `fix:`, `refactor:`, `docs:`, `test:`,
  `chore:`. Subject in the imperative, under 50 characters, no trailing period.
- **No personal data.** No real usernames, organizations, repository names, or
  tokens in code, tests, or documentation. Use `octocat/hello-world`, `acme`,
  `@alice`.

## Changes that affect how much it talks

The whole value of this tool is that people keep it installed. Anything that
can increase message volume needs care:

- Say in the pull request description how many more messages a typical user
  would receive, and why that is worth it.
- Adding an event kind means adding it to `engine.AllKinds` with a `Summary()`,
  and to the table in the README. Without the first, the event exists but
  cannot be muted — and a test enforces this.
- Anything that could fire on a state that was merely *observed for the first
  time*, rather than genuinely changed, needs a test. This has caused a real
  bug before: switching data sources filled several fields at once, and every
  one of them looked like a change.

## Changes that touch the GitHub API

- `internal/githubapi/repo.go` is the one place that queries without a user
  filter, and it must stay on-demand only. Wiring it into the alerting loop
  would turn a personal tool into a repository firehose.
- Search qualifiers are load-bearing. `review-requested:` versus
  `user-review-requested:` was a 61-to-1 difference on a real account.
- Rate limits differ per path: 5,000 points/hour for GraphQL, 30 requests/minute
  for authenticated REST search, 10/minute anonymous. State the budget impact
  of a new request in the description.
- The search index is eventually consistent. Never treat one absence as proof.

## Project layout

| Path | Responsibility |
|---|---|
| `cmd/ptal/main.go` | Command dispatch and usage. |
| `cmd/ptal/commands.go` | The commands themselves. |
| `cmd/ptal/config_cmd.go` | `ptal config`. |
| `cmd/ptal/events_cmd.go` | `ptal events`. |
| `cmd/ptal/service_cmd.go` | `start`, `stop`, `restart`, `pause`, `resume`. |
| `internal/runner/commands.go` | The Telegram command listener. |
| `internal/runner/review.go` | The review button and its callback. |
| `internal/reviewer` | Checkout, Claude invocation, verdict parsing, publishing. |
| `internal/config` | Configuration loading, validation, credential discovery. `settings.go` is the single table that makes an option settable, listable and documented at once. |
| `internal/githubapi` | GraphQL client, REST search client, and the fallback between them. |
| `internal/engine` | Snapshot diffing, event rules, anti-spam guards, alert-type filtering. |
| `internal/store` | State persisted between runs. |
| `internal/telegram` | Bot API client and message rendering. |
| `internal/service` | Registering the daemon with systemd and launchd. |
| `internal/runner` | The cycle that ties everything together. |

Keep files under 500 lines and one responsibility each. If a file starts
needing "and" to describe it, split it.

## Changes to the reviewer

`internal/reviewer` runs an agent over branches other people wrote. Two rules
are load-bearing and a test enforces each:

- **The tool list stays read-only.** No `Write`, no `Edit`, no unrestricted
  `Bash`. Widening it means a pull request from a stranger can execute on the
  reviewer's machine.
- **Claude never mutates GitHub.** Posting the comment and setting the verdict
  happens in `publish.go`, in code. An agent deciding for itself what to run
  against the API is not auditable.

A missing verdict must keep defaulting to `comment`. Approving or blocking on
a parsing slip is worse than saying nothing.

## Security

If you believe you have found a vulnerability, do not open a public issue — see
[SECURITY.md](SECURITY.md).
