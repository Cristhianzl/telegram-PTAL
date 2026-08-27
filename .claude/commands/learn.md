---
description: Capture a reusable learning from this session — with a dedup + Save/Absorb/Drop gate so the knowledge base stays clean
---

Capture a durable learning from what just happened — but only if it's worth keeping, and place it where it belongs. This upgrades the plain "append a learning" step with a quality gate so `learnings/` doesn't fill with duplicates.

1. **Identify** the single most reusable insight from this session — a convention, constraint, trap, or fix pattern. Skip trivial or one-off fixes.

2. **Check for overlap first** (don't append blindly):
   - Grep the relevant skill's `learnings/` and any project `learnings/` / `MEMORY.md` for the same topic and keywords.
   - Consider whether appending to an existing learning would suffice.

3. **State a verdict, then act:**
   - **Save** — unique and reusable → write `skills/<skill>/learnings/YYYY-MM-DD-slug.md` (frontmatter `description:`; body = the rule + **Why** + **How to apply**).
   - **Improve then Save** — tighten the scope/wording, then save.
   - **Absorb into [X]** — append to the existing learning instead; show the diff.
   - **Drop** — trivial or already covered → say so and stop.

4. Put it in the **most specific skill** it belongs to (a DB trap → `developing-features/learnings/`; a review heuristic → `reviewing-code/learnings/`). One pattern per file; body ≤ ~30 lines.

Confirm with the user before writing the file. Never run git.
