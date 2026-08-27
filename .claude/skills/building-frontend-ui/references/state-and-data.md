# State & data

Where state lives decides how testable, reusable, and bug-prone the UI is. These principles are framework-neutral; see `references/<framework>.md` for the local idiom (hooks, stores, signals, loaders).

## Own state at the right level

- **Keep data fetching out of leaf components.** Fetch in a loader/provider/store/route boundary and pass data down, or expose it through a context interface the UI consumes. The component that renders a value shouldn't be the one that knows how it's fetched.
- **Decouple state from UI.** The provider is the only place that knows *how* state is managed (local state vs store vs server sync); the UI depends on an interface (values + actions + meta). This is dependency injection — it maps to React context, Vue provide/inject, Svelte stores, etc.
- **Colocate by default; lift only when shared.** Move state up only when two siblings genuinely need it. Premature lifting creates prop-drilling and re-render churn.

## Derive, don't duplicate

- Compute values from existing state during render (`const total = items.reduce(...)`) instead of storing a second copy and syncing it. Duplicated state is the most common source of "the UI is out of sync" bugs.
- Don't run effects/watchers to copy one piece of state into another — derive it.

## The URL is state

- Filters, search, tabs, pagination, sort, and expanded panels belong in query params: shareable, deep-linkable, and they survive reload and back/forward.
- Treat the URL as the source of truth for navigable state; local state is for the ephemeral (open menu, hover, in-progress input).

## Fetching discipline

- **Parallelize independent requests** (`Promise.all`) — never serialize calls that don't depend on each other (waterfalls).
- Dedupe and cache identical in-flight requests; revalidate intentionally.
- Guard against **race conditions**: when inputs change mid-flight, ignore stale responses (cancel, or check a request token / `AbortController`).
- Consider optimistic updates only with a defined rollback path.

## Always handle the three non-happy states

Every data-backed view handles **loading**, **error**, and **empty** explicitly — not just the success path. Loading > ~1s shows a skeleton; errors give a recovery action; empty states give direction (what to do next), never a blank screen.
