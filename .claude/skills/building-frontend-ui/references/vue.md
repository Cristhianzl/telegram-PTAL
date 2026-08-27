# Vue (framework notes)

How the agnostic principles map to Vue 3 (Composition API). Load only if the project uses Vue.

## Components & reuse

- One SFC per component; extract reusable logic into **composables** (`useX()`), not duplicated blocks.
- Prefer **slots** (default/named/scoped) for composition over passing render functions; use scoped slots when the child must hand data back.
- Replace boolean-mode props with explicit variant components.
- Type props/emits with `defineProps<T>()` / `defineEmits<T>()`; no `any`. Stable `:key` from an id, not the index when items reorder.

## State & data

- **Derive with `computed`** — don't mirror state into a `ref` and keep it in sync with a `watch`. Watchers are for side effects, not derivation.
- Shared state via `provide`/`inject` against an interface, or a **Pinia** store; the store/provider is the only place that knows how state is managed. Colocate local state; lift only when shared.
- Keep fetching out of leaf components (route-level data, a store action, or a query library); dedupe and cancel stale requests (`AbortController`).
- Put navigable state (filters/tabs/pagination) in the route query via `vue-router`.

## Performance

- `defineAsyncComponent` / route-level lazy loading for heavy views; avoid whole-library/barrel imports.
- Use `v-once`/`v-memo` for static or rarely-changing subtrees; keep `computed` cheap; avoid heavy work in templates.

## Security

- **Never `v-html` with unsanitized input** (XSS) — sanitize first, or render as text.
- Don't bind untrusted URLs to `href`/`src` without validation.
