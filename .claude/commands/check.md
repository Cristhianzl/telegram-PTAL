---
description: Quick health check (git, lint, unit tests)
allowed-tools: Bash(git status:*)
---

# /check

Check only the areas with changes (`git diff --name-only HEAD`).

|Affected area|Command|
|------------|-------|
|changed source|`<lint command>`|
|changed source|`<unit test command>`|

If nothing changed, skip that area.

## Report

```text
Check
Git:    <summarized status -sb>
Lint:   <✅ | ⚠️>
Tests:  <X passed | Y failed>
```

If something fails, show the command to fix it and stop.
