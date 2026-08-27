# Forms

Forms are the highest-friction surface for real users. Get the native behaviors right before adding anything custom.

## Inputs

- Every input has an associated `<label>` (or `aria-label`). Placeholders are not labels.
- Use the correct `type` (`email`, `tel`, `url`, `number`, `password`, `search`, `date`) and `inputmode` so mobile keyboards and validation are right.
- Set a meaningful `name` and `autocomplete` (e.g. `email`, `current-password`, `one-time-code`, `street-address`) so password managers and autofill work.
- Disable `spellcheck` on emails, usernames, codes, and other non-prose fields.
- **Never block paste** (`onPaste` + `preventDefault`) — it breaks password managers and OTP flows.

## Validation & errors

- Validate at the right time (on blur / on submit, not aggressively on every keystroke); show errors **inline next to the field**.
- Announce errors to assistive tech (`role="alert"` / `aria-live`); on submit, move focus to the first invalid field; for long forms add an error summary that links to fields.
- Error messages say what's wrong and how to fix it — not "Invalid input."

## Submission & feedback

- The submit button stays enabled until the request actually starts; show a spinner/disabled state **during** the request to prevent double-submit.
- Reflect success and failure clearly; don't silently swallow either.
- Warn before navigating away with unsaved changes (`beforeunload` or a router guard).

## Layout & content

- Show example/expected format where helpful; placeholders end with `…` and illustrate the pattern.
- Inputs handle short, average, and very long values without breaking layout.
- Group related fields; keep the primary action visually dominant and singular.
