---
description: Check whether docs and configs reflect the real state of the code
allowed-tools: Read, Glob, Grep, Bash(git status:*)
---

# /sync

Cross-check `CLAUDE.md`, `.claude/rules/*` and `<roadmap file(s)>` against the real state.

## Check

- `CLAUDE.md` reflects the documented stack, versions, and ports
- All rules referenced in `CLAUDE.md` exist in `.claude/rules/`
- All commands in `.claude/commands/` have a `description` in the frontmatter
- Tasks marked ✅ in the roadmap have corresponding code
- Versions in the dependency manifests and lockfiles match what is documented

## Report

List of inconsistencies with the file path. For each one, suggest the fix.

If everything is OK, report only the green summary.
