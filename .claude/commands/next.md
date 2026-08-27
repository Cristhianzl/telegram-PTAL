---
description: Pick the next roadmap task and prepare the environment
allowed-tools: Read, Glob, Grep, Bash(git checkout:*), Bash(git branch:*)
---

# /next

- Read `<roadmap file(s)>` (and specifics if needed)
- Identify the current phase (first one with unfinished items)
- Select the first pending ⬜ item respecting dependencies

## Show

```text
Next task
Phase:   <phase name>
Area:    <area>
Task:    <description>

Likely files:
- <path 1>
- <path 2>
```

Suggest a branch (`feat/...`, `fix/...`) and ask:

- Start now (create branch and initial structure)
- See details (`/roadmap`)
- Choose another
