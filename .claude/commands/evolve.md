---
description: Promote accumulated learnings into a real skill, reference, or command — periodic self-improvement of this config
---

Look at the knowledge that has piled up in `learnings/` and decide what has earned promotion into a first-class artifact. Run this occasionally (e.g. when a skill's `learnings/` is getting crowded), not every session.

1. **Scan** every `skills/*/learnings/*.md` (skip the `README.md` templates) and any `MEMORY.md`.
2. **Cluster** related learnings — the same topic recurring across dates or skills is the signal.
3. **Pick the artifact per cluster:**
   - **3+ related learnings on one topic →** a new `references/<topic>.md` under the owning skill (or a new skill if it's a distinct domain).
   - **A recurring multi-step workflow →** a new `commands/<name>.md`.
   - **A one-off →** leave it as a learning.
4. **Propose before creating.** Show the user what you'd promote, where, and why (cite the source learnings). On approval, write the artifact and record `evolved from:` the source learning files in its frontmatter or a footer.

Don't delete the source learnings unless the user asks. Keep promoted artifacts short and in the repo's style. Never run git.
