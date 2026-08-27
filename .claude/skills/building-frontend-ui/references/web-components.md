# Web Components (framework notes)

How the agnostic principles map to standards-based custom elements (with or without Lit/Stencil). Load only if the project uses them.

## Components & reuse

- One custom element per file; compose with **`<slot>`** (default/named) rather than configuration flags; replace boolean-mode attributes with explicit variants.
- Keep logic in the element/controller; share via base classes or reactive controllers (Lit) instead of duplication.
- Type the public API — reflected attributes ↔ properties, and a typed event map; no `any`. Render lists with stable keys (Lit's `repeat` with a key fn), not positional reuse.

## State & data

- **Derive** in getters / `willUpdate` rather than storing duplicated state.
- Keep fetching out of presentational elements; pass data in via properties and emit changes **up via `CustomEvent`** (one-way data down, events up). Shared app state lives in a store/controller the element consumes, not inside leaf elements.
- Put navigable state in the URL (History API / router) rather than internal element state.

## Performance & encapsulation

- Use Shadow DOM for style/markup encapsulation; lazy-`define()` heavy elements (dynamic `import()`); avoid barrel imports.
- Batch DOM updates (Lit does this); don't read layout during render.

## Accessibility & security

- Shadow DOM can hide semantics: set ARIA on the host, use `delegatesFocus`, and ensure focus order works across shadow boundaries; reflect state to ARIA attributes.
- **Never assign untrusted `innerHTML`** — sanitize, or use templated text binding.
