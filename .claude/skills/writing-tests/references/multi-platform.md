# Multi-platform test strategy

A test suite that only runs on one OS lies about coverage. The greenness of CI on Linux says nothing about whether Windows is healthy.

This is the testing companion to the `ensuring-cross-platform` skill. Same principles, applied specifically to test code.

---

## The CI matrix is not optional

Every test in the suite — unit, integration, e2e — must be runnable on every supported platform without modification.

```yaml
# Required CI matrix (adapt to your stack)
strategy:
  fail-fast: false       # critical: a Windows failure must NOT abort the Linux run
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
runs-on: ${{ matrix.os }}
```

`fail-fast: false` is mandatory. When one OS fails, you NEED the others' state to confirm the scope: OS-specific bug vs universal bug. Aborting on first failure hides that information.

---

## Implicitly Linux-only tests — and how to fix

| Implicit Linux assumption                                       | Portable equivalent                                                 |
|------------------------------------------------------------------|---------------------------------------------------------------------|
| `Path("/tmp/fixture")`                                           | `tmp_path` (pytest fixture), `tempfile.TemporaryDirectory()`, `os.tmpdir()` |
| `assert content == "line1\nline2\n"`                             | `assert content.splitlines() == ["line1", "line2"]`                 |
| `subprocess.run("ls", capture_output=True)`                      | Use `os.listdir(...)` or `Path.iterdir()` directly                  |
| `assert os.access(f, os.X_OK)` for permissions                   | Skip on Windows or use `pytest.mark.skipif(sys.platform == "win32")` |
| `open("data.txt").read()`                                        | `open("data.txt", encoding="utf-8").read()`                         |
| Asserting filename equality after `os.listdir`                   | Normalize case before compare on Windows                            |
| `time.sleep(0.001)` for ordering                                 | Use proper synchronization; Windows timer resolution is ~15ms       |
| Tests that `fork()` worker processes                             | Use `multiprocessing` with the `spawn` start method                 |

---

## Platform-conditional tests — last resort, documented

Some tests genuinely cannot run on every OS (e.g., a Linux-only filesystem watcher).

```python
import sys
import pytest

@pytest.mark.skipif(
    sys.platform == "win32",
    reason="inotify is Linux-only; Windows uses a separate adapter tested in test_windows_watcher.py",
)
def test_inotify_emits_events(): ...
```

**Rules:**

- The skip reason **must** name the alternative coverage (the test that covers the same behavior on the other OS).
- A test skipped on every OS but one is a **red flag** — either the feature is genuinely OS-specific (document it in the feature docs) or the test is hiding a portability problem in the production code.

---

## Test data and line endings

Test fixtures committed to the repo will have their line endings rewritten on Windows checkouts unless `.gitattributes` enforces a policy:

```gitattributes
# Suggested baseline
* text=auto eol=lf
*.bat text eol=crlf
*.cmd text eol=crlf
*.ps1 text eol=crlf
# Binary fixtures must NEVER be touched
tests/fixtures/**/*.bin binary
tests/fixtures/**/*.png binary
tests/fixtures/**/*.zip binary
```

Without this, a fixture asserting `len(content) == 142` may pass on Linux and fail on Windows because Git rewrote `\n` to `\r\n` on checkout.

---

## Coverage counts per platform

The 80%-coverage gate applies **per OS in the matrix**. A Linux-only coverage report is misleading: branches that only run on Windows are reported as 0% covered when only the Linux runner contributes data.

When merging coverage across the matrix, use a tool that combines reports:

- Python: `coverage combine`
- Node: `nyc merge`
- Go: `go test -coverprofile=...` then combine with a script
- .NET: `coverlet` + `ReportGenerator`

---

## Docker tests

If the deliverable includes Docker:

- A smoke test that runs `docker build` + `docker run` against one critical path **must** be in the matrix.
- The Dockerfile itself is test-worthy: a `hadolint` lint step is recommended.
- Bind-mount tests are flaky across OSes — prefer **named volumes** or **baked-in test data**.

---

## Common cross-platform fix mistakes (watch in PR review)

| Mistake                                                                 | How to catch it                                              |
|-------------------------------------------------------------------------|--------------------------------------------------------------|
| Adding `if sys.platform == "win32"` without testing the other branches  | All branches must have test coverage                          |
| Replacing `/` with `os.sep` in one place but not another                | Grep the diff for residual `/` literals in path contexts      |
| "Fixing" line endings by force-writing `\n` in binary mode              | Verify file opens correctly on Windows readers                |
| Catching a Windows-specific exception silently                          | Re-raise with context; never swallow OS-specific errors       |

---

## When a platform-specific test breaks production

If a test passed on Linux but a Windows-specific bug shipped to production, the **suite did not run on Windows**. That is a separate finding to address:

1. Extend the CI matrix to include the missing OS (this PR, not a follow-up).
2. Add the platform-aware regression test (from `fixing-bugs`).
3. Document the gap in the post-mortem.

Without extending the matrix, the next instance of the same blind spot ships through the same gate.
