---
name: building-frontend-ui
description: Build, refactor, and review frontend UI in any framework — components, accessibility, forms, client state and data fetching, responsive layout, performance, and client-side security/PII. Use whenever the task touches the UI layer: creating or changing a component, page, or screen; styling or theming; accessibility/a11y; forms; client state; responsive/mobile layout; or reviewing/refactoring frontend code. Use even when the user doesn't say "frontend" — if they mention a component, button, form, modal, page, screen, layout, CSS/Tailwind, ARIA, focus, or a UI bug, this applies. Not for pure backend/API/database work.
license: MIT
---

# Building Frontend UI

A frontend change is good when it **reuses what already exists**, is **accessible by default**, **handles every state** (loading, error, empty, long content), **performs** under real data, and **leaks no data** to the client. This skill is framework-agnostic: the workflow and rules live here; framework-specific mechanics live in `references/<framework>.md` and are loaded only when relevant.

## Read first (always)

List `learnings/` and read every file relevant to the current task — project conventions, the design system, banned dependencies, and component patterns live there and override the defaults in this file. If a learning conflicts with this skill, **the learning wins** — mention it.

## Step 0 — Audit before you build (do not skip)

Most bad frontend work is *net-new code that should have been reuse*. Before writing anything:

1. **Find the design system.** Grep for the token/theme source — CSS custom properties (`--color-*`, `--space-*`), a `tailwind.config`, a theme file, design-token JSON. Use those tokens; **never hardcode raw hex, px spacing, or font sizes** in a component.
2. **Find existing components and utilities.** Grep/glob for a component that already does this (button, modal, field, table, empty-state) and **reuse or extend it** instead of creating a parallel one. Match the established prop and naming patterns.
3. **Find the i18n system, if any.** Grep for a `locales/` dir, i18n config, or translation calls (`t(...)`, `$t`, `useTranslation`, `FormattedMessage`). If the project has one, **every user-facing string goes through it** — a hardcoded string in a translated app is a bug, and **every new key must be added to all locale files** (a key present in one language and missing in another ships broken UI).
4. **Read the surrounding files** to copy the project's conventions (file layout, styling approach, state approach, test location).

→ verify: you can name the token source, the existing components you're reusing, the pattern you're following — and whether the project has i18n — or you've confirmed none exists.

## Priority order

When concerns compete, resolve in this order. Each row points to its deep-dive reference.

| # | Concern | Why it's ranked here | Reference |
|---|---------|----------------------|-----------|
| 1 | **Accessibility** | A broken experience for keyboard/AT users is a defect, not a polish item | `references/accessibility.md` |
| 2 | **Security & PII (client)** | Leaked tokens/PII and XSS are incidents | `references/` + baseline `CLAUDE.md` |
| 3 | **Reuse & design tokens** | Consistency and maintainability start here (Step 0) | this file |
| 4 | **Component design** | Small, composable, single-responsibility units | this file |
| 5 | **State & data** | Where state lives decides everything downstream | `references/state-and-data.md` |
| 6 | **Forms & feedback** | Highest-friction surface for real users | `references/forms.md` |
| 7 | **Responsive & layout** | Mobile-first; break by content, not device | `references/responsive.md` |
| 8 | **Performance** | Real data and real devices, not the happy path | `references/performance.md` |
| 9 | **Typography & UX writing** | Clarity and polish that read as professional | `references/ux-writing.md` |
| 10 | **Testing** | Behavior, a11y, and visual — see the testing skills | `skills/writing-tests` |

## Component design (agnostic)

- **Small, single responsibility, one component per file.** If you describe it with "and"/"or", split it.
- **Composition over configuration.** Prefer composing smaller pieces (children/slots, compound components) over a component with many flags.
- **No boolean-prop explosion.** Five booleans = 32 states, most invalid. Replace mode flags with **explicit variant components** so invalid combinations can't exist.
- **Children/slots over render callbacks** unless the parent must pass data back (then use a scoped slot / render prop deliberately).
- **Never define a component inside another component** — it remounts every render and breaks identity.
- **Type the public surface.** Props and state are fully typed; no `any`. Stable keys for lists (never the array index when items reorder).
- **Specify a component as Variants × Sizes × States.** Cover every state — default, hover, focus, active, disabled, loading — and resolve overlaps in that priority (disabled > loading > active > focus > hover > default).

## State & data (agnostic)

- **Keep fetching out of leaf components.** Data ownership lives in a provider/store/loader the UI consumes through an interface — agnostic to `useState`, a store, signals, or a server loader.
- **Derive, don't duplicate.** Compute from existing state during render instead of storing a copy and syncing it.
- **Lift state only when shared;** otherwise colocate it next to where it's used.
- **The URL is state.** Filters, tabs, pagination, and expanded panels belong in query params so they're shareable and survive reload.
- **Always handle loading, error, and empty** as first-class states. Details: `references/state-and-data.md`.

## Accessibility floor (every change)

Semantic HTML before ARIA. `<button>` for actions, `<a>` for navigation — never a clickable `<div>`. Visible keyboard focus (never remove the outline without replacing it). Labels on every control; alt text on meaningful images (`alt=""` for decorative); `aria-label` on icon-only buttons. Honor `prefers-reduced-motion`. **Fixability rule:** apply *mechanical* fixes directly; for *contextual/visual* ones (real alt text, copy, color choices) leave a clear `TODO` — **never invent** alt text, labels, or content. Full checklist: `references/accessibility.md`.

## Security & PII on the client

- **No tokens or secrets in `localStorage`/`sessionStorage`** — prefer httpOnly cookies. Store only what the UI needs, versioned, in a `try/catch`.
- **Never render untrusted HTML** (`dangerouslySetInnerHTML`/`innerHTML`) without sanitizing.
- **Minimize PII** sent to the client, serialized into props/markup, or logged to analytics. Treat the client↔server boundary as a trust boundary.
- No `console.log` in production. (These align with, and are partly enforced by, the baseline `CLAUDE.md` and hooks.)

## Workflow

1. **Audit first** (Step 0) — tokens, existing components, conventions.
2. **Design the component tree** — name the pieces, pick variants over flags, decide where state lives.
3. **Build with semantic, accessible markup** and the project's tokens. Handle loading/error/empty/long-content from the start.
4. **Wire state and data** at the right level; keep fetching out of leaves; reflect navigable state in the URL.
5. **Make it responsive** mobile-first; verify the layout at small/medium/large.
6. **Check performance** on realistic data (lists, images, animation).
7. **Test** behavior + accessibility (delegate to `skills/writing-tests`); add a visual check if the project has one.
8. **Capture a learning** if you hit a convention/trap not in this skill.

## Pre-delivery checklist

- [ ] Reused existing components/tokens; no hardcoded hex/spacing/font sizes.
- [ ] Keyboard-operable; visible focus; semantic elements; labels/alt present (or `TODO` for contextual ones).
- [ ] Loading, error, empty, and long-content states handled.
- [ ] No tokens/secrets in web storage; no unsanitized HTML; no PII leaked to client/logs.
- [ ] Typed props/state; small single-responsibility components; stable keys.
- [ ] Responsive at the project's breakpoints; animations honor reduced-motion.
- [ ] If the project has i18n: no hardcoded user-facing strings; every new key present in **all** locale files.
- [ ] Tests for behavior and accessibility; existing tests still pass.

## See also

- `references/accessibility.md` · `references/state-and-data.md` · `references/forms.md` · `references/performance.md` · `references/responsive.md` · `references/ux-writing.md`
- Framework specifics: `references/react.md` · `references/vue.md` · `references/svelte.md` · `references/web-components.md`
- `skills/writing-tests` (test strategy), `skills/developing-features` (general production-code rules, security depth), `skills/reviewing-code` (review output), `rules/frontend.md` (the enforced subset).
