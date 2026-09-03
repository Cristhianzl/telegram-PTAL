# Choosing a GitHub credential

PTAL reads GitHub with one of three credentials, tried in this order:

1. The **GitHub CLI**'s token, when `USE_GH_CLI=true` or no token is set
2. **`GH_PAT_TOKEN`** from the configuration file
3. **Nothing** — anonymous search, which sees only public repositories

Which one you want depends on where PTAL runs, and the answer is not always
the CLI.

## The trap: "but I am logged into `gh`"

Being logged in is not the same as the token being *readable*, and the
difference only shows up at boot.

`gh` stores its token in the operating system's credential store — the GNOME
keyring, or the macOS keychain. That store is encrypted, and it is unlocked by
your login. A daemon that starts with the system starts **before** that
happens:

```
07:48  machine boots, ptal.service starts
       gh auth token → fails, the keyring is still locked
       ptal comes up with no credential at all
   ⋮
12:24  you log in, the keyring unlocks
       ptal's next retry reads the token and recovers
```

Everything between those two points is answered anonymously. Private
repositories are invisible, and `/prs my-private-repo` fails with an error
about the repository not being visible — which is accurate but misleading,
because the repository is fine and the credential is the problem.

`gh auth status` in your terminal says you are logged in the whole time, which
is what makes this confusing. Your shell has an unlocked keyring; the service
did not.

PTAL handles this as well as it can: it retries after a minute, widening to
ten, and any command you type triggers an immediate attempt. But it cannot
read a locked keyring, so there is a window.

## Pick the setup that matches your machine

### A desktop or laptop you log into

**Use the CLI, and accept the window.**

```bash
ptal config use-gh-cli true
```

You log in shortly after booting anyway, and PTAL recovers within a minute of
the keyring unlocking. Nothing to store, nothing to rotate, and the token
never touches a file.

### A machine that stays on, or one you rarely log into graphically

**Use a classic personal access token.**

```bash
# https://github.com/settings/tokens/new?scopes=repo,notifications
ptal config use-gh-cli false
printf 'GH_PAT_TOKEN=ghp_...\n' >> ~/.config/ptal/.env
ptal restart
```

The file is written with mode `0600` and read at startup, with no keyring
involved — so the daemon is authenticated from its first cycle after a reboot.

If your repositories belong to an organization with SSO, click **Configure SSO
→ Authorize** on the token after creating it. Without that step the
organization's repositories vanish from results with no explicit error.

**Give the token an expiration.** Some enterprises refuse tokens that never
expire, or that live longer than a year:

```
The 'AcmeCorp' enterprise forbids access via a personal access tokens (classic)
if the token's lifetime is greater than 366 days.
```

This applies to classic and fine-grained tokens alike, and it is easy to walk
into: "No expiration" is right there in the dropdown, and the token appears to
work — `gh api user` succeeds, REST search succeeds. Only GraphQL is refused,
so the symptom is PTAL running in reduced mode with no obvious reason. Pick
90 days, or anything under a year, and edit the expiration on an existing
token rather than creating another one.

### A server, a container, or CI

**A token is the only option.** There is no keyring to unlock, and often no
interactive login to unlock it with. Pass it as an environment variable rather
than a file where you can — that is what makes the container and GitHub
Actions paths work:

```bash
docker run -e GH_PAT_TOKEN=ghp_... ...
```

`GITHUB_TOKEN` is accepted as an alias.

## Keeping both

Setting `GH_PAT_TOKEN` **and** `USE_GH_CLI=true` is a reasonable belt-and-braces
setup: the CLI token is preferred, and the file is there for the boot window
before the keyring unlocks.

One caveat, learned the hard way. If the token in the file is **fine-grained**
(`github_pat_…`) and any of your repositories sit under an organization that
restricts those, the fallback is worse than no fallback: the daemon starts,
uses the file token, gets rejected by the organization policy, and degrades to
public search. It recovers on the next retry, but a classic token avoids the
detour entirely.

## Which one is in use

```bash
ptal doctor
```

reports both the credential this shell can reach **and** what the running
daemon is actually using — they are not always the same, for exactly the
reason above.

```
• GitHub token: OAuth · source: GitHub CLI (gh auth token)
  (what this shell can reach - see the daemon's own state below)
✓ GitHub: authenticated as @octocat
✗ The running daemon has NO GitHub credential.
  It is searching anonymously, so private repositories are invisible.
  This happens when it starts before the system keyring unlocks.
  Fix now:  ptal restart
```

`ptal status` carries the short version:

```
daemon mode:  rich (authenticated)
```

## What each credential can see

| | Classic PAT | Fine-grained PAT | GitHub CLI | Anonymous |
|---|---|---|---|---|
| Public repositories | yes | yes | yes | yes |
| Private repositories | yes | only those granted | yes | **no** |
| CI state, approvals, conflicts | yes | often blocked by policy | yes | no |
| Notifications API | yes | **no** | yes | no |
| Survives a reboot without login | **yes** | **yes** | no | n/a |
| Rejected when it never expires | by some enterprises | by some enterprises | n/a | n/a |
| Search requests per minute | 30 | 30 | 30 | 10 |
