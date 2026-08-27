---
name: validating-in-reality
description: Validate a bugfix or feature against the REAL running system — execute the user's cURL, assert actual responses, check persisted database state, and drive the UI end-to-end — instead of assuming the code change works. Use whenever the user provides a cURL command, a localhost/endpoint URL, or asks to "validate", "test it for real", "check the database", "no assumptions", or when finishing a bugfix/feature in a project with a running server. The user's cURL IS the acceptance test. Not a substitute for writing-tests (automated coverage) or /verify (static gates) — this is live-system proof on top of them.
license: MIT
---

# Validating in reality

A change isn't done because the code looks right — it's done when the **running system proves it**. This skill turns a user-provided cURL (or endpoint, or repro) into the acceptance test, and replaces "should work" with pasted evidence.

## Read first (always)

List `learnings/` and read anything relevant — how this project starts its server, seeds data, connects to the DB, and authenticates test requests lives there and overrides the defaults here.

## The principle: evidence over inference

Reading the diff tells you what the code *should* do. Only the running system tells you what it *does*. When the user supplies a cURL, an endpoint, or a repro, that artifact is the **definition of done** — the task is complete when it behaves correctly against the live system, and you have the output to show it.

## Workflow

1. **Confirm the system is running.** Find the server (the project's run command, `docker compose ps`, the port in the cURL). If nothing is running and you can't start it with the project's documented command, **say so explicitly and stop** — never simulate a validation you didn't run.
2. **Run the acceptance test BEFORE the change** (bug: prove it reproduces — the live RED, mirroring `fixing-bugs`; feature: prove the behavior is missing). Paste the actual response.
3. **Implement** (per `fixing-bugs` / `developing-features` — tests included as usual).
4. **Run the acceptance test AFTER.** Paste the actual response and diff it against the BEFORE. Status code, body, and headers must show the expected change — actually compare them, don't eyeball.
5. **Check persisted state, not just the response.** Query the database (the project's client/ORM shell) and confirm the row/document actually changed — created, updated, deleted — as claimed. An API can return 200 and persist nothing.
6. **Exercise the negative cases**: wrong/missing auth, invalid payload, nonexistent ID. The fix must not have opened a hole the happy-path cURL can't see.
7. **If the flow has a UI, drive it end-to-end** with `skills/playwright-cli`: perform the user action in the browser and verify the visible result — the backend being right doesn't prove the frontend wired it.
8. **Report with evidence.** For each step: the command run and the real output (trimmed to the relevant part). Redact tokens/secrets/PII from anything you paste.

## Rules

- **Never claim "validated" for anything you didn't execute.** Partial validation is fine if labeled ("request + DB checked; UI not exercised — no frontend in this flow").
- The user's cURL is run **as given** (plus auth fixes they approve) — don't silently substitute a different endpoint that happens to pass.
- Keep validation **read-only beyond the tested operation**: don't mutate unrelated data; use test/seed data where the project provides it.
- These checks complement — never replace — automated tests (`writing-tests`) and the static gates (`/verify`).

## Capture a learning

If you discover how this project runs/seeds/authenticates for validation (base URL, test users, DB access command), append a `learnings/YYYY-MM-DD-slug.md` (or use `/learn`) so the next validation is instant.

## See also

- `skills/fixing-bugs` — the test-level RED→GREEN this skill mirrors at runtime level.
- `skills/playwright-cli` — driving the UI end-to-end.
- `/verify` command — static definition-of-done gates (build, types, lint, tests, secrets, diff).
- `hooks/check-real-validation.py` — auto-injects this requirement when your request contains a cURL or local endpoint.
