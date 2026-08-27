---
description: Adversarial dual review — two independent reviewers with no shared context; both must approve before merge
---

Review the current diff with **two independent reviewers** that don't share context, then converge. This catches what a single pass (or a self-review) misses and kills anchoring bias — neither reviewer is the author, neither sees the other's findings until the end.

1. **Run Reviewer A and Reviewer B independently** — as two separate subagents / fresh contexts (launch them in parallel). Each applies the `reviewing-code` skill to the same diff and returns severity-labeled findings (Blocker / Important / Recommended / Nice-to-have) plus a verdict: approve / approve-with-nits / request-changes. Give them no hint of each other.

2. **Converge.** Merge both reports and de-duplicate. A finding that **either** reviewer marks Blocker is a Blocker.

3. **Verdict.** **Approved only if both** reviewers return approve / approve-with-nits **and** there are no open Blockers. Otherwise, list exactly what must change.

4. **Up to 2 rounds.** After the author fixes the issues, re-run **both reviewers fresh** (new context). If still not converged after round 2, escalate to the human with the disagreement.

Output the merged review in chat following `skills/reviewing-code/references/output-format.md`. Never post to GitHub or run git.
