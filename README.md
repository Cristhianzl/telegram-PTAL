# PTAL

[![CI](https://github.com/Cristhianzl/telegram-PTAL/actions/workflows/ci.yml/badge.svg)](https://github.com/Cristhianzl/telegram-PTAL/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Cristhianzl/telegram-PTAL.svg)](https://pkg.go.dev/github.com/Cristhianzl/telegram-PTAL)
[![Go Report Card](https://goreportcard.com/badge/github.com/Cristhianzl/telegram-PTAL)](https://goreportcard.com/report/github.com/Cristhianzl/telegram-PTAL)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/Cristhianzl/telegram-PTAL?sort=semver)](https://github.com/Cristhianzl/telegram-PTAL/releases)

**P**lease **T**ake **A** **L**ook — the four letters developers leave on a pull
request when they need your eyes on it.

Your GitHub pull requests on Telegram. Runs in the background, starts with your
computer, and tells you when something needs you.

It watches **only what is tied to your user** — review requests, your own pull
requests, what was assigned to you, and where you were mentioned. It is not a
repository bot: it never tells you about pull requests that are not yours.

```
👀 Review requested

octocat/hello-world#412
Add streaming support to the chat endpoint
@alice · +120 −8 · ❌ CI

opened 2 days ago
```

---

## Requirements

- **Linux or macOS.** No administrator privileges required. Windows is not
  supported; use WSL2, where the Linux path works unchanged.
- **A GitHub credential** — the [GitHub CLI](https://cli.github.com) if you
  already use it, or a personal access token.
- **A Telegram bot token** — see step 1 below.

Nothing else. The binary has no runtime dependencies.

---

## Setup

### 1. Create a Telegram bot

Open [@BotFather](https://t.me/BotFather) in Telegram and send:

```
/newbot
```

Pick a display name and a username ending in `bot`. BotFather replies with a
token like `1234567890:ABC-DEF1234ghIkl-zyx57W2v1u123ew11`. Keep it — anyone
holding it controls your bot.

### 2. Install

```bash
git clone https://github.com/Cristhianzl/telegram-PTAL
cd ptal
make build
```

Or download a binary for your platform from the
[releases page](https://github.com/Cristhianzl/telegram-PTAL/releases).

### 3. Configure

Copy the template and fill in what you need:

```bash
cp .env.example .env
```

At minimum you need a Telegram bot token and a GitHub credential.

### 4. Connect Telegram

```bash
make setup
```

It validates the bot token, prints a link to your bot, waits for you to send
`/start`, and writes the chat ID into `.env` for you. No hunting for chat IDs.

### 5. Verify, then install

```bash
make doctor    # checks token, scopes, connectivity, and what is still missing
make install   # registers the service to start with your computer
```

---

## Configuration

| Variable | Required | What it is |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | yes | The token from BotFather. |
| `TELEGRAM_CHAT_ID` | — | Filled in by `make setup`. |
| `GH_PAT_TOKEN` | see below | GitHub personal access token. |
| `USE_GH_CLI` | no | Use the GitHub CLI's token instead of `GH_PAT_TOKEN`. |
| `GITHUB_LOGIN` | in public mode | Your GitHub username. Discovered from the token when there is one. |
| `WATCH_REPOS` | no | Limit to `owner/repo` or a bare org name, comma-separated. Empty watches everything. |
| `MAX_AGE_DAYS` | no | Ignore pull requests idle for more than N days. Default `14`, `0` disables. |
| `INCLUDE_TEAM_REVIEWS` | no | Include reviews requested from your team. Default `false`. |
| `MUTE_EVENTS` | no | Alert types never to send. See `ptal events`. |
| `ALERT_ON` | no | Send only these alert types. Empty means all. |
| `IGNORE_DRAFTS` | no | Default `true`. |
| `IGNORE_AUTHORS` | no | Comma-separated logins to skip. Defaults to common bots. |
| `POLL_INTERVAL` | no | Default `2m`. Minimum `30s`. |
| `QUIET_HOURS` | no | e.g. `23:00-08:00`. Alerts arrive without a sound. |
| `MAX_PER_HOUR` | no | Message ceiling per hour. Default `30`. |
| `PUBLIC_ONLY` | no | Skip GraphQL and use public search only. |

---

## Choosing a credential

There are three paths. PTAL tries them in order and falls through on its
own.

### The GitHub CLI (easiest)

If you already have [`gh`](https://cli.github.com) installed and authenticated,
you do not need to create a token at all:

```bash
gh auth login              # if you have not already
echo "USE_GH_CLI=true" >> .env
```

The CLI's token is OAuth, carries the `repo` scope, and — unlike fine-grained
tokens — passes organization policies that would reject other token types. It
stays in your system keyring rather than in `.env`.

> On a headless server the keyring may be locked and `gh auth token` fails. Use
> a token in `.env` there.

### A personal access token

Create a **classic** token with the `repo` and `notifications` scopes:

<https://github.com/settings/tokens/new?scopes=repo,notifications>

If your repositories belong to an organization with SSO, you must click
**Configure SSO → Authorize** on the token after creating it. Without that, the
organization's repositories vanish from results with no explicit error.

> **Fine-grained tokens (`github_pat_…`) have two real limitations:** they
> cannot reach the notifications API, and some organizations reject them by
> enterprise policy. When that happens PTAL detects it and switches to
> public search on its own — you keep receiving pull requests from public
> repositories, but without CI state, approvals, or private repositories. Run
> `make doctor` to see which mode you are in.

### No credential at all

For public repositories only, set `GITHUB_LOGIN` to your username and skip the
token. Anonymous search allows 10 requests per minute, which is enough at the
default two-minute interval.

---

## Reducing noise

Two settings matter more than all the others.

**`MAX_AGE_DAYS`** drops pull requests with no recent activity. Ones abandoned
long ago are still technically open and drown out the few that actually need
you. Filtering happens inside the search, so stale results never even transfer.

**`INCLUDE_TEAM_REVIEWS`** stays `false` by default, and this is deliberate.
GitHub treats "someone asked for your review" and "someone asked your team for a
review" with the same search qualifier. On a large team the second category
dominates: on one real account the broad qualifier returned 61 open pull
requests, of which exactly **1** was a request addressed to that person by name.
PTAL uses `user-review-requested` instead.

A third lever is turning off whole categories of alert. If CI noise is not
useful to you, for instance:

```bash
ptal config mute-events checks_failed,checks_fixed
```

All together, from the terminal:

```bash
ptal config max-age-days 5
ptal config include-team-reviews false
ptal config ignore-authors app/dependabot,app/renovate
ptal config mute-events checks_failed,checks_fixed
```

---

## Commands

```
ptal setup       Connect Telegram and discover your chat
ptal config      Read and change settings
ptal events      List alert types and which are on
ptal doctor      Diagnose token, chat, connectivity and service
ptal once        Run a single cycle and print what it found
ptal panel       Send a panel with the current state to Telegram
ptal run         Keep running (what the service executes)
ptal install     Register to start with the system
ptal uninstall   Remove the registration
ptal status      Show the service and the last sync
ptal version     Print the version
```

### `ptal config` — change settings from the terminal

No need to edit `.env` by hand. Keys are case-insensitive and accept hyphens,
so `poll-interval` and `POLL_INTERVAL` are the same setting.

```bash
ptal config                              # list everything with current values
ptal config poll-interval                # read one value
ptal config poll-interval 5m             # change it
```

Values are validated before they are written, so a typo is an error now rather
than a daemon that fails to start later:

```bash
$ ptal config poll-interval 5
error: POLL_INTERVAL: not a duration: "5" (try 2m, 30s, 1h)
```

When the service is running it is restarted for you, because configuration is
read once at startup — otherwise the change would silently do nothing.

```bash
$ ptal config poll-interval 5m
✓ poll-interval: 2m0s → 5m
✓ service restarted
```

Common ones:

```bash
ptal config poll-interval 5m                        # how often to check GitHub
ptal config max-age-days 3                          # only recently active PRs
ptal config quiet-hours 23:00-08:00                 # deliver silently at night
ptal config watch-repos octocat/hello-world,acme    # limit to repos or orgs
ptal config max-per-hour 10                         # hard ceiling on messages
```

### `ptal events` — choose which alerts you get

```bash
$ ptal events

Alert types:

  ● on   review_requested     Someone asked you to review a pull request
  ● on   mentioned            Someone @-mentioned you
  ● on   assigned             A pull request was assigned to you
  ● on   changes_requested    A reviewer requested changes on your pull request
  ● on   approved             Your pull request was approved
  ○ off  checks_failed        CI failed on your pull request
  ○ off  checks_fixed         CI recovered on your pull request
  ● on   conflict             Your pull request developed a merge conflict
  ● on   new_activity         New comments or reviews on a pull request of yours
  ● on   ready_for_review     A draft you are reviewing became ready
  ● on   gone                 A pull request was closed or merged
```

Two ways to control it, and they compose:

```bash
# Turn off the ones that annoy you, keep the rest
ptal config mute-events checks_failed,checks_fixed

# Or name only what you want, and nothing else arrives
ptal config alert-on review_requested,mentioned

# Back to everything
ptal config mute-events ""
```

`mute-events` always wins over `alert-on`. Unknown names are rejected — a typo
in a mute list would otherwise leave you believing an alert was off.

`ptal once` works before Telegram is configured, so you can check the search
first without connecting the bot.

---

## Keeping it running

| System | Mechanism | Needs administrator |
|---|---|---|
| Linux | user-mode systemd + `loginctl enable-linger` | no |
| macOS | LaunchAgent with `RunAtLoad` and `KeepAlive/NetworkState` | no |

`ptal install` detects the system and writes the right unit. On macOS,
`KeepAlive/NetworkState` means the daemon only runs when the network is up. On
Linux, `enable-linger` is what keeps the service alive after logout — without
it, PTAL would stop silently.

### Running it somewhere else

Your laptop only alerts you while it is awake. Two alternatives, both using the
same binary:

**A container**, for a VPS or a home server:

```bash
docker run -d --restart unless-stopped \
  -e TELEGRAM_BOT_TOKEN -e TELEGRAM_CHAT_ID -e GH_PAT_TOKEN \
  -v ptal-state:/data ghcr.io/cristhianzl/telegram-ptal
```

**GitHub Actions**, if you would rather run nothing at all — see
[`docs/github-actions.md`](docs/github-actions.md). Note that scheduled
workflows are not punctual: they can lag 5 to 20 minutes at busy times, and
GitHub disables the schedule after 60 days without commits.

---

## What becomes a message

| Event | When | Urgency |
|---|---|---|
| Review requested | the pull request enters your review bucket | now |
| Changes requested | on one of your pull requests | now |
| CI broke | on one of your pull requests | now |
| Assigned / mentioned | — | now |
| Pull request approved | on one of your pull requests | batched |
| Merge conflict | on one of your pull requests | batched |
| New comments | on any of your pull requests | batched |
| Ready for review | waiting on your review | batched |
| Closed or merged | confirmed across two cycles | batched |

Four guards keep this from becoming spam: every event carries a unique key and
is never sent twice; non-urgent events are grouped; quiet hours deliver without
a sound; and an hourly ceiling protects against floods.

**The first run is silent** — it only records what already exists. You do not
get fifty messages on install.

---

## How it works

Two read paths, with automatic degradation.

The preferred path is a **single GraphQL query** covering every bucket at once,
carrying CI state, review decision and merge status. It costs about 5 points of
the 5,000 available per hour, so a two-minute interval uses roughly 3% of the
budget.

When an organization policy rejects the token, PTAL falls back to **REST
search**. It sees fewer fields, but keeps working. The token is still sent on
this path even in reduced mode, because it raises the request budget from ten
per minute to thirty.

State lives in a single JSON file written atomically. It holds the last picture
of your pull requests, the fingerprints of events already sent, and a few health
markers. Losing it costs one silent resync, nothing more.

A pull request that disappears from search is not immediately reported as
closed: GitHub's search index is eventually consistent, so absence must be
confirmed across two consecutive cycles.

---

## Development

```bash
make test     # all tests
make check    # vet plus tests with the race detector
make dist     # binaries for Linux and macOS
```

No external dependencies — the standard library only.

---

## Troubleshooting

**No messages arrive** — run `make doctor`. It checks the token, its scopes,
the chat, connectivity, and the service, and says what is missing.

**"The organization rejects this token"** — your token is fine-grained and an
enterprise policy blocks it. Use `gh` or a classic PAT with SSO authorized.

**Search returns nothing** — `WATCH_REPOS` or `MAX_AGE_DAYS` may be filtering
everything out. `make doctor` prints both when the result is empty.

**The service stopped after logout (Linux)** — `loginctl enable-linger $USER`.
`ptal install` does this, but it can fail silently on some systems.

**Too many messages** — lower `MAX_AGE_DAYS`, confirm `INCLUDE_TEAM_REVIEWS` is
`false`, and add noisy bots to `IGNORE_AUTHORS`.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Security issues go to
[SECURITY.md](SECURITY.md), not to a public issue.

## License

[MIT](LICENSE) — do whatever you like, keep the copyright notice, no warranty.
