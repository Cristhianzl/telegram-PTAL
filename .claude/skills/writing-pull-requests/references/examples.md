# PR examples: wrong vs. right

Distilled from common mistakes. Each block shows what LLMs typically produce, why it fails, and the corrected version.

---

## Titles

### Too vague

❌ `fix: Fix bug`
✅ `fix(cart): Resolve null pointer when applying coupon to empty cart`

❌ `feat: Add improvements`
✅ `feat(search): Add fuzzy matching for product names`

The test: a reviewer reading only the title should know which bug, which feature, which area.

### Multi-purpose (split the PR)

❌ `feat(auth): Add OAuth and refactor session storage and update docs`
✅ Three PRs:
  - `feat(auth): Add Google OAuth login`
  - `refactor(auth): Move session storage to Redis`
  - `docs(auth): Document OAuth setup`

If the title needs "and", you have two PRs.

### Wrong tense / casing

❌ `feat(api): Added pagination to /users endpoint`
✅ `feat(api): Add pagination to /users endpoint`

❌ `fix(ui): resolve modal flicker`
✅ `fix(ui): Resolve modal flicker`

Imperative mood. Capitalize the first letter after the colon.

### Implementation noise in title

❌ `refactor(cache): Replace HashMap<String, CacheEntry> with ConcurrentHashMap`
✅ `refactor(cache): Make cache reads thread-safe`

The title says what changes for the reader, not how the code changes.

### Inventing a scope

❌ `feat(everywhere): Bump dependencies`
✅ `chore: Bump dependencies to latest patch versions`

Omit scope when the change is genuinely repo-wide.

---

## Descriptions

### Restating the diff

**Diff:** added a `debounce()` call around the loading state setter in `Button.tsx`.

❌
```
## Changes
- Modified Button.tsx
- Added import for debounce from lodash
- Wrapped setLoadingState in debounce with 100ms wait
- Updated the onClick handler to use the new debounced function
```

✅
```
## Changes
- Debounce loading state updates to prevent flicker on rapid clicks
```

The reviewer can read the diff. The description explains the **why** and the **shape**, not the line-by-line.

### Burying the why

❌
```
## Objective
Update the user service.
```

✅
```
## Objective
Stop logging full request bodies on `/users/:id` — bodies contained PII and were retained 30 days in Loki.
```

State the user-visible or business outcome. "Update X" tells you nothing.

### Over-long

❌ A description that walks through every file, every method, every config value. 600 words.

✅ ≤150 words. One sentence Objective. Up to 5 bullet Changes. Notes only for edge cases that require reviewer attention (breaking change, rollback path, follow-up ticket).

### Missing edge case / migration callout

When the change has a migration, flag rollout, or breaking change, **the Notes section is mandatory**, even if Objective and Changes are obvious.

✅
```
## Notes
- Requires migration `0042_add_user_status_column` to run before deploy.
- Old clients (<v2.4) will see `status=null` until they refresh — acceptable per @product.
```

---

## Output channel

❌ Creating `PR_DESCRIPTION.md` in the repo and committing it.
✅ Printing the `TITLE / COMMIT MESSAGE / DESCRIPTION` block directly in the chat for the human to paste.

❌ Running `git commit -m "$(...)"` as part of the PR generation.
✅ Stopping after printing. The human commits.

---

## Karpathy-style summary

| Anti-pattern                            | Why it fails                              | Fix                                    |
|-----------------------------------------|-------------------------------------------|----------------------------------------|
| Vague title ("Fix bug")                 | Reviewer can't tell what's in the PR      | Name the symptom and the area          |
| Multi-purpose title with "and"          | Hides scope, blocks rollback              | Split into two PRs                     |
| Description restates the diff           | Wastes reviewer time, adds no signal      | Describe shape and intent, not lines   |
| Title says how, not what                | Couples title to implementation           | Title is the outcome                   |
| Writes `PR_DESCRIPTION.md`              | Pollutes the repo, requires cleanup       | Output goes in chat                    |
| AI runs `git commit`                    | Removes human review                      | Only the human commits                 |
