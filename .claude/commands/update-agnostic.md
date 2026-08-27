---
description: Update this project's .claude/ to the latest "agnostic" config from the claude-skills-czl repo, preserving local learnings
allowed-tools: Bash, Read, Write, Edit
---

Update the project's `.claude/` to the latest **agnostic** config from
`https://github.com/Cristhianzl/claude-skills-czl`, **without losing local additions**.

Run these steps carefully and report at the end. Stop and ask if anything is ambiguous.

1. **Pre-checks.**
   - If `.claude` is a **symlink** (`test -L .claude`), it already tracks the kit — tell the user to `git pull` in the kit repo instead, then stop.
   - **Variant guard.** Read the installed variant: `installed=$(cat .claude/.variant 2>/dev/null)`. If it is non-empty and **not `agnostic`** (e.g. `langflow`), **STOP**: this project has the `$installed` config, and `/update-agnostic` would switch variants — overwriting the `$installed`-specific rules, skills, and hooks. Tell the user to run `/update-$installed` to stay on their variant, or to re-run only after **explicitly confirming** they want to switch to `agnostic`.
   - Confirm with the user before overwriting `.claude/`.

2. **Download the latest config** to a temp dir:
   ```bash
   tmp=$(mktemp -d)
   curl -fsSL https://github.com/Cristhianzl/claude-skills-czl/archive/refs/heads/main.tar.gz | tar -xz -C "$tmp"
   new="$tmp/claude-skills-czl-main/configs/agnostic"
   ```
   If `curl`/`tar` fail, report and stop.

3. **Back up the current config:** `cp -R .claude ".claude.bak-$(date +%Y%m%d-%H%M%S)"`.

4. **Preserve local additions** (copy them aside, then restore after the sync):
   - every `learnings/*.md` that is not `README.md` (project-accumulated knowledge),
   - `settings.local.json` if present,
   - any `rules/*.md` or `skills/*` that exist locally but not in `$new` (project-specific).

5. **Apply the update:** replace the tracked parts of `.claude/` (CLAUDE.md, skills, commands, hooks, rules, settings.json) with the contents of `$new`, then restore everything from step 4 on top.

6. **Report:** what was updated, what was preserved, and the backup path. If the project commits `.claude/`, remind the user to review the diff before committing. Clean up `$tmp`.

Never delete the user's `learnings/` content. (Forks: change the repo URL above to your own.)
