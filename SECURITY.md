# Security Policy

## What this software is

PTAL is a background daemon that holds two credentials — a GitHub token
and a Telegram bot token — and makes outbound requests with them. It listens on
no port, accepts no inbound connection, and executes nothing on your behalf.

The realistic risk is **credential exposure**, not remote compromise.

## Threat model

**Trusted:** the machine's operating system and user account, the GitHub API,
and the Telegram Bot API.

**Not trusted:** anyone else with read access to the machine, and the content
of pull request titles and author names, which are attacker-influenced text
that ends up inside a Telegram message.

**Controls in place:**

| Control | What it stops | What it does not stop |
|---|---|---|
| `.env` written with mode `0600` | Other users on the machine reading your tokens. | Anyone who can already read your home directory as you, or root. |
| State file written with mode `0600` | The same, for cached pull request data. | The same. |
| Tokens never logged, never printed | Credentials leaking into journald, log files, or a pasted `doctor` output. | A token you paste somewhere yourself. |
| HTML escaping on every GitHub-sourced string | A crafted pull request title breaking the message markup or injecting a link. | Nothing else; it is an integrity control for the message. |
| GitHub CLI credential path | The token being stored in a file at all — it stays in the system keyring. | Anything once `gh auth token` can be run as you. |

## Token scope

PTAL only ever **reads**. It never writes to GitHub: no comments, no
reviews, no merges, no repository changes. If you are choosing scopes for a
classic token, `repo` is broader than this tool needs — it is required because
GitHub has no read-only equivalent that also covers private repository search.

The Telegram bot token is more sensitive than it looks: anyone holding it can
read every message sent to your bot and send messages as it. Treat it like a
password.

## Reporting a vulnerability

Do not open a public issue.

Report privately through
[GitHub Security Advisories](https://github.com/Cristhianzl/telegram-PTAL/security/advisories/new),
which lets us discuss and fix the issue before it becomes public.

Please include what an attacker could do, the steps to reproduce, and the
version (`ptal version`). You can expect an initial response within a
week.

## Out of scope

- Anyone with an interactive session as your user account. They can read the
  `.env` directly; nothing this tool does can prevent that.
- The GitHub or Telegram APIs themselves.
- Rate limiting or denial of service against your own daemon.
