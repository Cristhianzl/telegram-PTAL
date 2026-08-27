---
description: Send commits to the remote
allowed-tools: Bash(git:*)
---

# /push

- `git status -sb` to see the branch and state
- `git log @{u}..HEAD --oneline` to list the commits to send (if no upstream, use `git log -3 --oneline`)
- If there are no new commits, report it and stop
- Show a summary (branch, N commits) and confirm
- Push: `git push -u origin $(git branch --show-current)` (`-u` the first time)
- On conflict: do not force; suggest `git pull --rebase` and stop
- Never `--force` unless the user explicitly asks
- Never `--force` on `main` under any circumstance
