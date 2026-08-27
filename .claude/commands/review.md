---
description: Review the current diff against the project rules
allowed-tools: Read, Grep, Bash(git diff:*), Bash(git status:*)
---

# /review

Review `git diff HEAD` against the rules in `.claude/rules/`.

## Checklist

- Compliance with the rule of the edited area (`.claude/rules/<area>.md`)
- `CLAUDE.md`: type hints, no `print`/`console.log`, no credentials, no duplication
- Security: parameterized SQL, no raw `dangerouslySetInnerHTML`, validated inputs
- Auth: new private endpoints have an auth dependency when needed
- Tests: new code has a corresponding test

## Report

List of issues by severity (critical / warning / suggestion). For each, cite the file:line.

At the end, say whether it is OK for `/done` or whether it needs fixing first.
