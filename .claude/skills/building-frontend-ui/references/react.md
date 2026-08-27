# React / Next (framework notes)

How the agnostic principles map to React. Load only if the project uses React.

## Components & reuse

- One component per file; extract reusable logic into **custom hooks**, not copy-paste.
- **Never define a component inside another** — it remounts every render. Define at module scope.
- Replace boolean-mode props with explicit variant components; use children/composition over `renderX` props (keep render props only to pass data back, e.g. list item + index).
- Type props/state with interfaces/types; no `any`. Stable `key` from a stable id, never the array index when items reorder.

## State & data

- **Derive during render** (`const total = items.reduce(...)`) — don't store derived state and sync it with `useEffect`. Effects are for synchronizing with *external* systems, not for computing values.
- Decouple state: expose state + actions through a context **interface**; the provider is the only thing that knows it's `useState` vs Zustand vs server. Lift state only when shared.
- Keep fetching out of leaves: use a data layer (Server Components / route loaders / React Query / SWR) and pass data down. Dedupe with the library; guard stale responses with `AbortController`.
- Reflect navigable state in the URL (search params / `nuqs`-style), not just `useState`.

## Performance

- Avoid barrel imports (`import { X } from 'big-lib'`) — import the specific path; barrels can load thousands of modules.
- `React.lazy` / `next/dynamic` for heavy widgets; `memo` + stable callbacks/identities to avoid needless re-renders; lazy `useState` initializer for expensive init.
- Controlled inputs must be cheap per keystroke; prefer uncontrolled when you don't need every change.

## Security & hydration

- Never `dangerouslySetInnerHTML` with unsanitized input.
- Inputs with `value` need `onChange` (or use `defaultValue`); guard date/time/random rendering against hydration mismatch; use `suppressHydrationWarning` only where truly unavoidable.
