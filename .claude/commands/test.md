---
description: Run tests related to the modified files
allowed-tools: Bash(git diff:*)
---

# /test

Steps:

- List changes: `git diff --name-only HEAD`
- Map each changed source file to its test by the project's convention
- Run only the relevant tests
- If any fail, show the error and suggest a fix
- If there is no test for a modified file, warn about it

## Commands

```bash
<test command> <paths or pattern>
```
