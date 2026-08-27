# Svelte / SvelteKit (framework notes)

How the agnostic principles map to Svelte. Load only if the project uses Svelte. (Syntax differs between Svelte 4 reactive statements and Svelte 5 runes — follow the project's version.)

## Components & reuse

- One component per file; extract shared logic into modules/stores or (Svelte 5) reusable rune-based functions, not copy-paste.
- Use **slots** (named/scoped) for composition; replace boolean-mode props with explicit variant components.
- Type props (`export let` with TS / `$props()` in runes); no `any`. Keyed `{#each items as item (item.id)}` — always key by id, never index when items reorder.

## State & data

- **Derive** with `$:` reactive declarations (Svelte 4) or `$derived` (Svelte 5) — don't duplicate state and sync it manually.
- Shared state via **stores** (or `$state`/context in runes) exposed through a small interface; colocate local state, lift only when shared.
- Keep fetching out of leaf components: use SvelteKit `load` functions (server/universal) and pass data via `data`; cancel stale requests with `AbortController`.
- Put navigable state in the URL via `$page`/`goto` and query params.

## Performance

- Route-level code splitting is automatic in SvelteKit; lazy-load heavy components with dynamic `import()`. Avoid barrel/whole-library imports.
- Use keyed each-blocks and avoid unnecessary reactivity over large structures.

## Security

- **Never `{@html ...}` with unsanitized input** (XSS) — sanitize first or render as text.
- Validate untrusted URLs before binding to `href`/`src`.
