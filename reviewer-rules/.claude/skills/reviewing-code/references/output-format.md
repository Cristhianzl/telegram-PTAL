# Output format — GitHub comment rules

The review is posted as a single GitHub PR comment. The rules below prevent the most common broken-output failures. Follow them strictly.

---

## Hard rules — never do these

| Anti-pattern                                                                 | Why it fails                                                                                  |
|------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------|
| `#1`, `#2`, `#3` as finding labels                                            | GitHub auto-links `#N` to PR/issue `N` in the repo, creating false references to unrelated PRs |
| `GH-N`, `gh-N`, or bare issue/PR numbers                                     | Same auto-link problem                                                                        |
| `@username` mentions                                                          | Sends an unwanted notification                                                                |
| `Fixes #N`, `Closes #N`, `Resolves #N`                                       | Auto-closes issues on merge                                                                   |
| Pasting full file contents                                                    | Comment becomes unreadable; GitHub truncates                                                  |
| Including the full checklist verbatim                                         | The checklist is an internal tool; the output is the **findings** the checklist produced     |
| Absolute local paths (`/Users/...`, `/home/...`)                              | Leaks local environment                                                                       |
| Emojis in section headers (when user requests "clean" or "professional")      | Adds visual noise; default emojis OK in severity badges                                       |

---

## Required label format

Use one scheme consistently within a single review. **Never `#N`.**

| Scheme                           | Example                         | When to use                                       |
|----------------------------------|---------------------------------|---------------------------------------------------|
| Severity-prefixed (**preferred**)| `B1`, `B2`, `I1`, `I2`, `R1`     | Default. B = Blocker, I = Important, R = Recommended, N = Nice-to-have |
| Plain numeric                    | `1.`, `2.`, `3.`                | Short reviews with < 5 findings, no severity grouping |
| Finding-prefixed                 | `F1`, `F2`, `F3`                | Mixed-severity flat list                          |

When cross-referencing a finding elsewhere in the document, use the same label (e.g., "Test suggestions post-split (**B1**)") — never `(#1)`.

---

## Required structure (skeleton)

```markdown
## Code Review Summary

<2-4 sentence verdict: ship / changes-requested / blocked, and the headline reason>

**Verdict:** <Approve | Approve with comments | Request changes | Block>
**Findings:** <N blockers, M important, K recommended>

---

## ⛔ Blockers (resolve before merge)

### B1 — <Short title>
**File:** `path/to/file.ts:42-58`
**Issue:** <1-3 sentence problem statement>
**Why it matters:** <impact / blast radius>
**Suggested fix:** <concrete action>

<details>
<summary>Code reference</summary>

```ts
// quote only the 3-10 most relevant lines
```
</details>

### B2 — <Short title>
...

---

## ⚠️ Important (preferably this PR)

### I1 — <Short title>
...

---

## 💡 Recommended (can ship as a follow-up)

### R1 — <Short title>
...

---

## ✅ Action checklist for the author

**Blockers (resolve before merge):**
- [ ] **B1** — <one-line restatement>
- [ ] **B2** — <one-line restatement>

**Important (preferably this PR):**
- [ ] **I1** — <one-line restatement>

**Recommended (can ship as a follow-up PR):**
- [ ] **R1** — <one-line restatement>

**Test suggestions (post-B1 split):**
- [ ] <test 1>
- [ ] <test 2>
```

---

## Length and density

- **Target length:** 300–800 lines of markdown for a typical PR.
- **Hard cap:** 1500 lines. GitHub truncates very long comments.
- **One finding = one section.** Don't merge unrelated issues into a single bullet.
- **Quote sparingly.** 3–10 lines per `<details>` block. If the reader needs more, they open the file.
- **Use `<details>` / `<summary>`** for: code quotes longer than 5 lines, long rationale, grep output, command transcripts. Keeps the comment scannable.
- **Use code fences with language tags** (\`\`\`ts, \`\`\`py, \`\`\`bash) for syntax highlighting.

---

## Link format

| Format                                              | Use                                                                 |
|-----------------------------------------------------|---------------------------------------------------------------------|
| `path/to/file.ts:42` (repo-relative, plain)         | Default. GitHub renders as plain text; humans copy-navigate.        |
| `[file.ts:42](path/to/file.ts#L42)`                  | Only when posting to a known repo URL where line anchors resolve.   |
| `/Users/<you>/...` (absolute local path)             | Forbidden — leaks local environment.                                |
| `#42` referring to a line number                     | Forbidden — GitHub will think it's PR #42.                          |

---

## Copy-paste safety check (before posting)

Before declaring the review done, verify the rendered output:

- [ ] No `#N` patterns anywhere except inside fenced code blocks.
- [ ] No `@username` mentions (unless explicitly requested).
- [ ] No `Fixes/Closes/Resolves #N` keywords.
- [ ] No absolute local paths.
- [ ] All finding labels follow the chosen scheme consistently (no mixing `#1` with `B1`).
- [ ] Total length under the cap; long quotes are inside `<details>`.
- [ ] Renders cleanly in GitHub's markdown preview (no broken tables, unclosed code fences, stray HTML).
- [ ] The action checklist at the bottom matches the findings above one-for-one.

---

## How to raise a security finding

When you find a security issue, the finding must include:

1. **What the assumption is** — "This assumes the body is the signed content."
2. **Why it's wrong or unverified** — "Mercado Pago signs a specific constructed string, not the raw body."
3. **What the blast radius is** — "Anyone can forge a payment webhook and credit arbitrary accounts."
4. **What the fix is** — "Read the official Mercado Pago notification docs and implement the `x-signature` verification exactly."

## How to raise a comprehension finding

When a block fails the comprehension audit:

1. **Quote the block** (or link to `file:line`).
2. **Ask the author to explain** the *why*, not the *what*: "Why this dispatch table instead of the existing strategy pattern in `payment_strategies.py`?"
3. **Block the PR** until either: (a) the author defends the choice and captures the rationale durably (test name, `Why:` comment, PR `## Design Decisions`, or ADR), or (b) the author rewrites the block with full understanding.

This is not gatekeeping. This is preventing comprehension debt from being merged.

## How to raise an AI-untrusted-code finding

```
Location: <file:line>
Block: <quote or summary>
Concern: This block is in a high-risk path (<auth | payments | signatures | ...>) and was AI-generated. The following assumption needs verification against <official documentation source>:

  Assumption in the code: "<quote or paraphrase>"
  Where to verify: <link to provider docs or reference implementation>
  Risk if assumption is wrong: <data exposure | privilege escalation | forged events | ...>

Required action: Verify against the authoritative source, then either confirm in a `Why:` comment or refactor.
```
