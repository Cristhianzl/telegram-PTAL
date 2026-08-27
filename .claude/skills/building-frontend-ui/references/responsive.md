# Responsive & layout

## Mobile-first

- Design and build for the smallest screen first, then scale up. It forces focus on essential content, scales up more easily than down, and tends to perform better.
- **Choose breakpoints by content, not by specific devices** — add a breakpoint where the layout starts to break, not at a phone's width. A common set is ~480 / 768 / 1024 / 1440, but let the content decide.

## Simplify on small screens (don't just shrink)

Re-shape dense UI for small viewports instead of squeezing it:

| Pattern | Mobile | Desktop |
|---------|--------|---------|
| Data table | Stacked cards / key fields | Full table |
| Filters | Modal / bottom sheet / drawer | Persistent sidebar |
| Multi-column form | Single column | Multi-column |
| Primary nav | Hamburger / bottom bar | Inline / sidebar |

## Layout system

- Use a consistent spacing scale (4/8px increments) from tokens, not arbitrary values.
- Constrain reading width on large screens (max-width); don't let body text run edge-to-edge.
- Define a small, named `z-index` scale (e.g. base / dropdown / sticky / overlay / modal / toast) instead of ad-hoc numbers.
- Let flex/grid children shrink (`min-width: 0`) so long content can truncate instead of overflowing.

## Responsive media

- `<img srcset sizes loading="lazy">`; `<video preload="metadata">` with per-breakpoint `<source media=...>`.
- Respect safe areas and touch targets (≥ ~44px); set `touch-action: manipulation`; use `overscroll-behavior: contain` in modals/drawers/sheets.

## Verify

- Check the layout at small / medium / large widths. If the project has visual testing, snapshot each breakpoint (e.g. 375 / 768 / 1440) so regressions are caught.
