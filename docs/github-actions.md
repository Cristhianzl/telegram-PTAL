# Running PTAL on GitHub Actions

Your laptop only alerts you while it is awake. If you would rather run nothing
at all, the same binary works as a scheduled workflow.

## Trade-offs first

This is not the recommended path, and it is worth knowing why before setting it
up.

- **Scheduled workflows are not punctual.** GitHub queues them, and at busy
  times a `*/5` schedule can fire 5 to 20 minutes late. The minimum interval is
  5 minutes regardless.
- **The schedule stops after 60 days without commits** on a public repository.
  GitHub disables it silently.
- **State has to live somewhere.** Without it, every run would either re-alert
  everything or nothing. The cache below is best-effort; a private Gist is more
  reliable if you find the cache being evicted.

For alerts within a minute of the event, run the daemon on a machine that stays
on — your own, or a small VPS.

## Setup

Add three repository secrets under **Settings → Secrets and variables →
Actions**:

| Secret | Value |
|---|---|
| `PTAL_GH_TOKEN` | A classic PAT with `repo` and `notifications`. |
| `PTAL_TG_TOKEN` | Your Telegram bot token. |
| `PTAL_TG_CHAT` | The chat ID that `ptal setup` discovered. |

> Do not reuse `GITHUB_TOKEN`. The automatic token is scoped to the repository
> running the workflow and cannot see your pull requests elsewhere.

Then add `.github/workflows/ptal.yml`:

```yaml
name: PTAL

on:
  schedule:
    - cron: "*/10 * * * *"
  workflow_dispatch:

permissions:
  contents: read

concurrency:
  # Two runs at once would both read the same state and alert twice.
  group: ptal
  cancel-in-progress: false

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7

      - uses: actions/setup-go@v7
        with:
          go-version: "1.24"
          cache: true

      - name: Restore state
        uses: actions/cache/restore@v4
        with:
          path: state.json
          key: ptal-state-${{ github.run_id }}
          restore-keys: ptal-state-

      - run: go run ./cmd/ptal once
        env:
          GH_PAT_TOKEN: ${{ secrets.PTAL_GH_TOKEN }}
          TELEGRAM_BOT_TOKEN: ${{ secrets.PTAL_TG_TOKEN }}
          TELEGRAM_CHAT_ID: ${{ secrets.PTAL_TG_CHAT }}
          MAX_AGE_DAYS: "7"
          MUTE_EVENTS: ""

      - name: Save state
        if: always()
        uses: actions/cache/save@v4
        with:
          path: state.json
          key: ptal-state-${{ github.run_id }}
```

Environment variables override the `.env` file, so every setting from
`.env.example` can be passed this way.

## Verifying it

Trigger it by hand first, from the **Actions** tab → **PTAL** → **Run
workflow**. The first run is silent by design — it records what already exists
without announcing it. The second run onward will alert on changes.

If nothing arrives, check the run log: `ptal once` prints what it found and
which mode it used.
