# Typography & UX writing

Polish that reads as professional comes from restraint, consistent tokens, and clear words.

## Typography

- Body `line-height` 1.5–1.75; line length ~65–75 chars (~35–60 on mobile); base font size 16px (also avoids mobile auto-zoom on inputs).
- Use a small, consistent type scale (e.g. 12/14/16/18/24/32) and a weight hierarchy (headings 600–700, body 400, labels 500) — from tokens.
- `font-variant-numeric: tabular-nums` for numbers in columns, prices, and timers to prevent shifting.
- `text-wrap: balance`/`pretty` on headings to avoid widows.

## Visual polish (reads as "made by a pro")

- **Semantic color tokens**, never raw hex in components. Dark mode overrides semantic tokens with desaturated/lighter tonal variants — never just inverts colors.
- Consistent icons: one style per hierarchy level (filled vs outline), consistent stroke width and sizes defined as tokens. **No emoji as structural icons** (font-dependent, inconsistent, untokenizable).
- One primary action per screen; secondary actions are visually subordinate.
- Establish hierarchy with size, spacing, and contrast — not color alone. Restraint with animation: excessive motion reads as AI-generated.

## UX writing

- **Write from the user's side of the screen.** Name things by what the person controls ("Notifications", not "Webhook config").
- **Active voice; say exactly what happens.** "Save changes", not "Submit". An action keeps the same name through the whole flow.
- **Errors give direction, not apologies.** Say what went wrong and the next step; never vague ("Something went wrong").
- **Empty states are an opportunity:** explain what goes here and offer the first action.

## Typographic details

- `…` (ellipsis) not `...`; curly quotes; non-breaking spaces in unbreakable pairs (`10 MB`, `⌘ K`).
- Truncate or clamp long content deliberately (`text-overflow`/line-clamp) rather than letting it overflow.
- Use locale-aware formatting for dates, numbers, and currency (`Intl.*`), never hardcoded formats; mark brand/code tokens `translate="no"`.
