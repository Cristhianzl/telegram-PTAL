# Changelog

All notable changes to this project are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.2.2] - 2026-09-02

### Fixed

- A daemon that starts with no credential at all now retries, instead of
  searching anonymously for the rest of the session. The earlier fix only
  covered degrading *into* reduced mode; one that started there never tried
  again, so every private repository stayed invisible until a manual restart.
- `ptal doctor` reports on the running daemon rather than only on itself. It
  runs from a shell, where the keyring is unlocked and a token is available,
  and was reporting everything green while the service had been running blind
  since boot. `ptal status` shows the daemon's mode and credential too.
- Starting without a credential is logged as a warning rather than passing
  silently.

## [0.2.1] - 2026-08-31

### Added

- The bot registers its commands with Telegram, so they appear in the chat's
  `/` menu instead of only being findable by typing `/help` first. `/commands`
  and `/start` answer with the list too.

### Fixed

- Falling back to public search is no longer permanent. A daemon starting at
  boot finds the system keyring locked, cannot read the GitHub CLI's token,
  falls back to whatever is in `.env`, and can be rejected by an organization
  policy — after which it stayed in reduced mode until someone restarted it by
  hand. The authenticated path is now retried every 10 minutes, re-reading the
  CLI token, so an unlock minutes after boot is picked up on its own.
- A search naming a repository the credential cannot see now says so. GitHub
  answers `422 Validation Failed`, which sent people hunting for a syntax
  error that was not there.

## [0.2.0] - 2026-08-27

### Added

- Review a pull request with Claude Code: a button on the alert checks the
  branch out, runs the CLI headlessly against your own `.claude/` rules, and
  posts the review to GitHub with the pull request approved or changes
  requested. Uses the CLI's existing subscription, not an API key.
  Also available as `ptal review <repo> <number>` and `/review` in Telegram.
- Reviewing is off until repositories are named in `REVIEW_REPOS`.
- `ptal review ... -m "..."` and `/review <repo> <n> <instructions>` steer a
  single review; `REVIEW_INSTRUCTIONS` steers every one. Instructions sit
  before the output rules so they cannot break the verdict contract.
- `ptal start`, `stop`, `restart`, `pause` and `resume`. Stopping the daemon
  previously meant reaching for `systemctl` or `launchctl` directly.
- Pausing suppresses delivery without stopping the cycle, so resuming does not
  produce a flood of everything that changed while you were away.
- Telegram commands: `/prs`, `/status`, `/clear`, `/pause`, `/resume`, `/help`.
  Only the configured chat is obeyed.
- `/clear` deletes the messages the bot has sent, and reports how many were
  actually removed — Telegram refuses anything older than 48 hours.
- `/prs <repo>` and `ptal repo <name>` list every open pull request in a
  repository, whoever they belong to. A bare name is matched against
  `WATCH_REPOS`, so the owner does not have to be remembered. A second
  argument narrows it to one author, with `me` as a shorthand for your own
  login. On demand only: it never feeds the alerting loop.

### Removed

- `BATCH_WINDOW`, which was configurable and documented but never read by
  anything. Grouping happens by urgency within a cycle, not on a timer. A
  setting that silently does nothing is worse than no setting.

### Fixed

- Commands that take no arguments now reject extra ones. `ptal install --help`
  used to install the service instead of printing help.

## [0.1.0] - 2026-08-27

First release.

### Added

- Watches the pull requests tied to your GitHub user: review requests, your own
  pull requests, assignments, mentions, and ones you already reviewed.
- Telegram alerts, with a pinned panel that is rewritten each cycle rather than
  posting a new message every time.
- Automatic chat discovery: `ptal setup` waits for you to message the bot and
  writes the chat ID into `.env` itself.
- Credential discovery from the GitHub CLI, a personal access token, or nothing
  at all for public repositories.
- Two read paths with automatic degradation: authenticated GraphQL for CI
  state, review decisions and merge status; REST search when an organization
  policy rejects the token.
- Starts with the system through systemd or launchd, neither of which requires
  administrator privileges. Windows is not supported; WSL2 works.
- Anti-spam guards: per-event fingerprints, grouping of non-urgent events,
  quiet hours, and an hourly ceiling.
- `ptal doctor`, which checks the token, its scopes, the chat, connectivity and
  the service, and explains what is missing.
- `ptal config` reads and changes settings from the terminal, with validation
  before anything is written and an automatic service restart so the change
  takes effect.
- `ptal events` lists every alert type, what triggers it, and whether it is on.
  `MUTE_EVENTS` and `ALERT_ON` control which ones are delivered.
- One-line install script for macOS and Linux, with checksum verification.
- A container image, and a documented path for running on GitHub Actions.

### Notes

- Review requests use `user-review-requested`, matching only requests addressed
  to you by name. `INCLUDE_TEAM_REVIEWS=true` restores the broader behavior.
- `MAX_AGE_DAYS` defaults to 14. Pull requests abandoned long ago stay
  technically open and would otherwise drown out the rest.
- The first run is silent, recording what already exists without announcing it.

[Unreleased]: https://github.com/Cristhianzl/telegram-PTAL/compare/v0.2.2...HEAD
[0.2.2]: https://github.com/Cristhianzl/telegram-PTAL/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/Cristhianzl/telegram-PTAL/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/Cristhianzl/telegram-PTAL/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/Cristhianzl/telegram-PTAL/releases/tag/v0.1.0
