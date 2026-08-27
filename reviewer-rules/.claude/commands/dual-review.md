---
description: Adversarial dual review — two independent reviewers with no shared context; both must approve
---

# /dual-review

Review the diff in `$ARGUMENTS` (or the current diff, picked as in `/review`) with **two
independent reviewers**, then converge. This kills anchoring bias — neither reviewer is the author,
and neither sees the other's findings until the end.

1. **Run Reviewer A and Reviewer B independently**, as two separate subagents with fresh context,
   launched in parallel. Each applies `skills/reviewing-code/SKILL.md` to the same diff and returns
   severity-labeled findings plus a verdict: approve / approve-with-nits / request-changes. Give
   neither any hint of the other.

2. **Converge.** Merge both reports and de-duplicate. A finding **either** reviewer marks Blocker
   is a Blocker.

3. **Verdict.** Approved only if **both** return approve or approve-with-nits **and** no Blockers
   remain open. Otherwise list exactly what must change.

4. **Up to 2 rounds.** After the author fixes the issues, re-run **both** reviewers fresh. If they
   still haven't converged after round 2, escalate to the human with the disagreement stated plainly.

Output the merged review in the chat per `skills/reviewing-code/references/output-format.md`.
Never post to GitHub, never run git.
