---
description: Definition-of-done gate — build, types, lint, tests, a secrets/debug scan, and a diff self-review before calling work done or opening a PR
---

Run the project's checks in order and report a short verdict. This is a gate you **invoke** (not an automatic block) — run it before declaring a task done or opening a PR.

Use the project's actual commands (e.g. `uv run …` where the repo requires it). Run in this order; stop at the first hard failure and report it:

1. **Build** — the project builds/compiles cleanly.
2. **Types** — type checker passes (0 errors).
3. **Lint / format** — linter and formatter clean.
4. **Tests** — full suite green; branch coverage meets the gate.
5. **Secrets / debug scan** — no secrets or credentials, no debug `print()` / `console.log`, no `.env` in the diff.
6. **Diff self-review** — re-read the whole diff: every change is intended, no leftover scaffolding or stray `TODO`, docs updated if behavior changed (the `Stop` doc-sync check helps here).
7. **Runtime validation (when a server / user-provided cURL applies)** — prove it against the live system per `skills/validating-in-reality`: run the acceptance request, check DB state, E2E the UI flow. Static gates green + runtime proof = done.

Report a one-line **PASS/FAIL per step** and an overall verdict. On any FAIL, name the step and the fix — don't claim done. Never run git.
