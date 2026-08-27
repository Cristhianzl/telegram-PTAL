# Applying security fixes without breaking the app

Security applied carelessly breaks production. A fix should change the **attack surface** and nothing a legitimate user experiences. Follow this on **every** fix.

## The protocol

1. **Never on `main` / prod.** One branch per fix (`sec/<slug>`).
2. **One security change per commit.** If something breaks, you know exactly which change did it.
3. **Backup before any DB change.** RLS, migrations, and schema changes can block legitimate queries.
4. **Test the legitimate flow before and after.** Functional behavior must be identical — only the attack surface changes. Write or run a test that proves the legit path still works.
5. **Report-only rollout for the risky ones.** CSP and rate-limit have report/observe modes — run them there first, watch, then enforce. Use a feature flag where possible.
6. **Password-hash migration is transparent, never bulk.** Re-hash each user's password on their **next successful login** — zero downtime, no forced reset for anyone.
7. **Staging before prod** — especially RLS, CORS, CSP, and rate-limit, the ones that most often break legitimate flow.
8. **Fail closed.** If a permission check errors, deny — never default to allow.

## The golden rule

> If a fix changes the **result a legitimate user sees** (beyond blocking what should be blocked), you probably introduced a regression. Stop, find why the rule caught the legitimate case, tighten the rule (don't loosen security), and retest.

## The ones that most often break legitimate flow

- **RLS** — the #1 cause of "the data disappeared / the app broke" in this kind of stack. Enable it **table by table**; test with two different users and confirm each sees only their own rows.
- **CORS** — a wrong origin allow-list blocks your own frontend.
- **CSP** — blocks inline scripts and CDNs. Report-only first, always.
- **Rate limit** — can catch a legitimate burst. Observe first, then enforce.

## The per-fix loop

Pick one item → branch → (backup if it touches the DB) → make the **minimal** fix → run the legitimate flow it affects → confirm the malicious case is now blocked → report-only if it's CSP/rate-limit → staging → prod → note what changed. If a legitimate user loses access at any step: **revert**, understand why the rule caught the legit case, adjust, retest.
