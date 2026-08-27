---
description: Create a commit with validation (lint + tests) — optionally pass the message
argument-hint: ["optional commit message"]
allowed-tools: Bash(git:*)
---

# /commit

Provided message: `$ARGUMENTS` (use it directly if not empty).

Flow:

- Detect modified areas (`git diff --name-only`)
- Run lint on the affected areas; if it fails, try `--fix`; if it still fails, **STOP**
- Run the relevant tests (see `/test`); if they fail, **STOP** with an error
- `git add` only the modified files (never a blind `git add -A`)
- Message in English, format `type: description` (max 50 chars, imperative, no trailing period)
- Types: feat, fix, refactor, docs, test, chore, ui
- Confirm with the user before committing (unless a message was already passed)

## Rules

- **Never add a `Co-Authored-By` (or any AI-attribution) trailer** to the message
- Never commit `.env` or credentials
- No `--no-verify` unless the user asks
- No `--amend` unless the user asks (create a new commit)
