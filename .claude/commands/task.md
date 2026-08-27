---
description: Mark a task as done in the roadmap
argument-hint: ["task name or snippet"]
allowed-tools: Read, Edit, Glob, Grep
---

# /task

Argument: `$ARGUMENTS` (task name or snippet).

- Without argument: list 🟡 (in progress) tasks from all `<roadmap file(s)>` and ask which to mark
- With argument: locate the task via Grep and mark it as ✅ (replace `⬜` or `🟡` → `✅`)

## Confirm

```text
✅ <task>
Roadmap: <file>
Phase progress: <X/N>

Next suggested: <task>  → /next
```
