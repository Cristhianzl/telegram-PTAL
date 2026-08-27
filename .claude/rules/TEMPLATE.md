---
description: <one line — what stack/area this rule governs>
globs: "<glob, e.g. **/*.<ext>>"
---

# <Language / Stack>

> Prerequisite: `CLAUDE.md` (general rules). Depth: see `skills/`.

The baseline in `CLAUDE.md` and the `skills/` are language-agnostic. This folder is
where you add the conventions that are specific to one stack and not already covered
by the baseline. Copy this file to `rules/<stack>.md`, set `description` and `globs`,
and keep it short — point to the skills for depth.

What belongs here:

- Type/typing conventions specific to the language
- Idiomatic error handling and logging
- Project/module structure for this stack
- Stack-specific anti-patterns to avoid

Add one such file per stack your project actually uses; delete this template (or keep
it as a reference). With no per-stack files, the baseline alone still applies.
