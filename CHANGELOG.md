# Changelog

All notable changes to this project are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

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

[Unreleased]: https://github.com/Cristhianzl/telegram-PTAL/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Cristhianzl/telegram-PTAL/releases/tag/v0.1.0
