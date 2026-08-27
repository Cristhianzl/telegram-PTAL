---
description: Start a session with project context
allowed-tools: Read, Glob, Bash(git status:*), Bash(git log:*), Bash(git branch:*)
---

# /init

Generate a short session briefing.

## Collect

1. Current branch and last commit (`git status -sb`, `git log -1 --oneline`)
2. Modified files (short, no `-uall`)
3. Current phase in `<roadmap file>` (first one with unfinished items)
4. Next 3 pending tasks of the phase

## Report

```text
Session
Branch:  <branch>
Status:  <N modified | clean>
Last:    <hash> <msg>
Phase:   <name> (<status>)

Next tasks:
1. ...
2. ...
3. ...
```

Suggest `/check`, `/next`, `/done` at the end.
