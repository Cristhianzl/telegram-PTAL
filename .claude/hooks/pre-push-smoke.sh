#!/usr/bin/env bash
# Fast smoke test before `git push` — catch obvious lint/test errors before
# spending CI quota. NOT the full validation (that is the /check command). The
# goal here is to be fast (seconds), not exhaustive.
#
# Triggered by a PreToolUse hook matching `git push`. Reads the event JSON from
# stdin. Exit 2 = block the push; exit 0 = allow.
#
# Only runs checks for areas with committed-but-not-yet-pushed changes.

set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT" || exit 0

# Skip special cases (tag push, branch delete, --force already blocked by deny)
input="$(cat)"
cmd="$(printf '%s' "$input" | jq -r '.tool_input.command // empty' 2>/dev/null)"
case "$cmd" in
  *--force*|*-f\ *|*--delete*|*--tags*) exit 0 ;;
esac

# Files changed vs origin (what this push would carry). Fallback: last commit.
upstream="$(git rev-parse --abbrev-ref --symbolic-full-name '@{u}' 2>/dev/null || echo '')"
if [ -n "$upstream" ]; then
  changed="$(git diff --name-only "$upstream"...HEAD 2>/dev/null)"
else
  changed="$(git diff --name-only HEAD~1...HEAD 2>/dev/null)"
fi
[ -z "$changed" ] && exit 0

fail() { echo "BLOCKED (pre-push smoke): $1" >&2; echo "   Run /check to see the detail. Push aborted." >&2; exit 2; }

# Project-specific lint/test on changed areas — fill in for your stack. Examples:
#   if printf '%s' "$changed" | grep -q '\.py$'; then ruff check . -q || fail "ruff"; fi
#   if printf '%s' "$changed" | grep -q '^web/'; then ( cd web && npm run lint --silent ) || fail "lint (web)"; fi
# With nothing configured below, the push is allowed.

exit 0
