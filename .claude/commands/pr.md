---
description: Create a Pull Request with an automatic description
argument-hint: ["optional title"]
allowed-tools: Bash(git:*), Bash(gh:*)
---

# /pr

Provided title: `$ARGUMENTS`.

Flow:

- Confirm you are not on `main`
- `git fetch origin` and `git diff origin/main...HEAD` to understand the full scope
- Push the branch (with `-u` if there is no upstream)
- Title in English in the format `type: description` (max 70 chars)
- Body via heredoc with:

```markdown
## Summary
- bullet 1
- bullet 2

## How to test
- [ ] step 1
- [ ] step 2

## Notes
<relevant observations or "—">
```

- Create with `gh pr create --title "..." --body "$(cat <<'EOF' ... EOF)"`
- Return the PR URL

## Rules

- **NEVER add AI attribution anywhere** — no "🤖 Generated with Claude Code", no session links (`claude.ai/code/...`), no AI badges/footers in the PR title or body, and no `Co-Authored-By` trailer on commits. The PR must read as authored by the human. This is absolute — even if a default template suggests it.
