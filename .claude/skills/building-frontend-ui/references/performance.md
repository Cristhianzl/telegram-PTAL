# Performance

Test on realistic data and mid-range devices, not the happy path. These are framework-neutral; the local idiom for code-splitting/memoization is in `references/<framework>.md`.

## Rendering & layout

- **Virtualize long lists** (> ~50 items) or use `content-visibility: auto` — never render thousands of nodes.
- **Reserve space to prevent layout shift (CLS):** images/embeds need `width`/`height` or `aspect-ratio`; reserve space for async content and skeletons.
- **Don't read layout during render** (`getBoundingClientRect`, `offsetHeight`, `scrollTop`) and don't interleave reads and writes — batch DOM reads, then writes.
- Keep per-frame work under ~16ms (60fps). Target input latency < ~100ms; `debounce`/`throttle` high-frequency handlers; prefer passive scroll listeners.

## Assets

- Images: modern formats (WebP/AVIF), `srcset`/`sizes` for responsive delivery, `loading="lazy"` below the fold.
- Fonts: `font-display: swap` (or `optional`) to avoid invisible text; `preload` critical fonts; `preconnect` to asset/CDN origins.
- Ship only what's needed: **avoid barrel/whole-library imports** (they pull thousands of modules); import the specific symbol. Code-split routes and heavy widgets; defer non-critical third-party scripts until after first interaction; preload on hover/focus intent.

## Animation

- Animate **only `transform` and `opacity`** (compositor-friendly); never animate `width`/`height`/`top`/`left` in hot paths.
- **Never `transition: all`** — list the properties. Keep micro-interactions 150–300ms; exits ~60–70% of the enter duration; make animations interruptible and honor `prefers-reduced-motion`.

## Data & computation

- Parallelize independent I/O; avoid request waterfalls.
- For hot loops over collections: build a `Map`/`Set` for O(1) lookups instead of repeated `.find`/`.includes`; hoist invariant work (regex, property access) out of the loop; early-return.
- Derive instead of recomputing across the tree; give children stable identities so they don't re-render needlessly.
