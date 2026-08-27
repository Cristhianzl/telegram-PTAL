---
name: ensuring-cross-platform
description: Apply platform-agnostic engineering rules so code, tests, and Docker builds behave identically on Linux, macOS, and native Windows. Use when writing or reviewing code that touches the filesystem, shell, subprocess, encoding, time, paths, native dependencies, or CI matrix — or when the user mentions Windows, macOS, Linux, "works on my machine", cross-platform, portability, or Docker.
license: MIT
---

# Ensuring Cross-Platform Behavior

A feature is not "done" because it works on the developer's machine. It is done when it produces identical, correct behavior on every supported platform — verified by tests that run on every supported platform.

This skill is the **single source of truth** for platform agnosticism. Other skills (`developing-features`, `developing-features-tdd`, `fixing-bugs`, `writing-tests`, `reviewing-code`, `documenting-features`) reference this one and apply its principles to their respective phases.

## Read first (always)

List `learnings/` and read every file relevant to the current task. Project-specific platform constraints, supported-OS lists, or known traps live there and override the defaults below. If a learning conflicts with this file, **the learning wins** — mention it to the user.

## Default supported platforms

Until the project says otherwise, assume the defensive case:

| Platform     | Architectures                  | Notes                                               |
|--------------|--------------------------------|-----------------------------------------------------|
| Linux        | x86_64, arm64                  | Glibc distros (Ubuntu 22.04+, Debian 12+, RHEL 9+)  |
| macOS        | x86_64, arm64                  | macOS 13+                                           |
| Windows      | x86_64                         | Windows 10 (22H2) / 11 — **native, not WSL**        |
| Docker       | linux/amd64, linux/arm64       | Linux containers; host OS may be any of the above   |

If the project's real target is narrower (e.g., Linux only), the user must say so explicitly. **WSL is not a substitute for native Windows support** unless the user documents that exception.

## Pre-implementation platform check

Before writing any code that touches the filesystem, the shell, the environment, or external processes, answer:

1. **Which platforms must this run on?** (See the table above or the project's override.)
2. **Am I making any POSIX assumption that is not guaranteed on Windows?** — path separator, line endings, encoding, fork, signals, file permissions, atomic rename.
3. **Am I calling any external tool I am not declaring as a dependency?** — `git`, `make`, `curl`, `tar`, `unzip`, anything assumed on `PATH`.
4. **What is the failure mode on the platform I am NOT testing on right now?** — "It probably works" is not an answer.
5. **Is this code in a hot path that varies per OS?** — filesystem watching, process spawning, GUI, GPU. If yes, design the abstraction first (see `references/platform-specific-code.md`).

If any answer is "I don't know" → stop, find out, then continue.

## Mandatory patterns

These are language-agnostic principles. For concrete code in Python, TypeScript, Java, Go, C#, etc., see `references/patterns.md`.

- **Path APIs, not string concatenation.** Use the language's path-join API (`pathlib.Path` / `path.join` / `filepath.Join`). Never concatenate paths with `/` or `\`.
- **Always specify encoding for text I/O.** Default to UTF-8. Never rely on locale-dependent defaults — they break on Windows hosts.
- **No hardcoded shell commands.** Use language-native APIs (`shutil.rmtree`, `fs.rm`, etc.). When you must invoke a subprocess, pass arguments as a list, never with `shell=True`. For tools assumed on PATH, check existence first and declare the dependency.
- **Line endings: read tolerantly, write per-OS.** Read with `splitlines()`-equivalent (handles `\n`, `\r\n`, `\r`). Write in text mode and let the runtime translate, **except** for files where the line ending must be exact (protocols, binary formats) — write in binary mode then.
- **OS-aware temp / cache / config directories.** Use `tempfile`-equivalents and a `platformdirs`-equivalent. Never hardcode `/tmp`, `~/.config`, or `C:\...`.
- **Validate filenames before writing them** when the name comes from user input or an external source. Reject Windows reserved names (`CON`, `PRN`, `AUX`, `NUL`, `COM1-9`, `LPT1-9`), forbidden chars (`< > : " / \ | ? *` + control chars), trailing space/dot, and `.` / `..`.
- **Time: UTC at boundaries, local at presentation.** Never store local time. Never compare timezones with naive datetimes.
- **Native dependencies must have prebuilt artifacts** for every target architecture (manylinux + macosx + win_amd64 wheels, prebuilt Node binaries, etc.). Anything calling `LoadLibrary("foo.dll")` / `dlopen("libfoo.so")` is platform-specific by definition — wrap and isolate.

The full anti-pattern reference table (POSIX-only mistakes and their portable fixes) is in `references/anti-patterns.md`.

## CI matrix is mandatory

> Tests that only run on Linux prove nothing about Windows or macOS. A green build on one OS is necessary but never sufficient.

The CI configuration MUST run the test suite on **every** supported OS in the matrix, with `fail-fast: false` so that a Windows failure does not hide a parallel Linux pass that confirms the bug is OS-specific.

A PR is NOT mergeable when:
- Any matrix cell is failing.
- A matrix cell was removed to "make CI green".
- The matrix was bypassed via merge override "just this once".

If the project ships Docker, the image MUST build successfully from at least the Linux runner, and a `docker run` smoke test that exercises one critical path MUST be part of the matrix. See `references/docker.md` for image hygiene and `.dockerignore` baseline.

## Platform-specific code, when justified

Some functionality is genuinely OS-specific (filesystem watchers, GPU, system notifications, hardware sensors). When that happens:

1. **Isolate behind a single abstraction.** A `Notifier` protocol; one factory; lazy imports inside each branch so importing a Linux-only module on Windows does not crash startup.
2. **Detect at runtime, not at import time.**
3. **Test each implementation on its target platform** in CI.
4. **Document the branch map** — every place the implementation differs per OS, why, and the test that covers each branch (see `documenting-features` skill).
5. **Loud, early failure for unsupported platforms.** Raise `NotImplementedError` (or equivalent) with a clear message naming the platform and the feature. Silent fallback to a no-op is forbidden.

Concrete shape and example code in `references/platform-specific-code.md`.

## Validation checklist (before delivery)

Apply this in addition to the checklists in `developing-features`, `writing-tests`, etc. Every box must be ticked.

**Code**
- [ ] All paths constructed with a path API — no string concatenation, no hardcoded `/` or `\`.
- [ ] All file I/O specifies encoding explicitly (UTF-8 unless documented otherwise).
- [ ] No hardcoded shell commands; subprocess calls use list args, no shell expansion.
- [ ] No hardcoded `/tmp`, `~/.config`, `C:\…` — uses temp / platformdirs equivalents.
- [ ] No reliance on POSIX utilities (`ls`, `grep`, `cat`, `rm`) called from code.
- [ ] Time stored in UTC at boundaries; local time only at presentation.
- [ ] Filename validation rejects Windows-reserved names and forbidden chars when input is external.
- [ ] Native dependencies have prebuilt artifacts for every target architecture (verified on the registry).

**Tests**
- [ ] Tests pass on every supported platform (CI matrix), not just the developer's machine.
- [ ] Tests do not assume case-sensitive filesystems.
- [ ] Tests do not assume specific line endings.
- [ ] Tests do not depend on `/tmp`, `/etc`, or other POSIX-specific locations.
- [ ] Platform-specific code has tests that run on each target platform.

**Docker (when applicable)**
- [ ] Image builds reproducibly from Linux, macOS, AND Windows hosts.
- [ ] `.dockerignore` excludes host-OS artifacts (`.DS_Store`, `Thumbs.db`, `node_modules`, etc.).
- [ ] Smoke test runs `docker run` against one critical path in CI.
- [ ] No bind mounts in production compose; named volumes only.

**Documentation**
- [ ] Supported OS versions are listed in README.
- [ ] Known platform-specific limitations are documented.
- [ ] Platform-specific code branches are documented (which OS, why, where).

If any box fails → fix before delivering. There is no "I'll add Windows support next sprint" — by then the design has calcified around POSIX assumptions and the cost is 10x.

## Final step: capture a learning

Before closing the task, ask: *did I encounter a platform constraint, trap, or convention not in this SKILL.md or `references/`?* If yes, append a `learnings/YYYY-MM-DD-slug.md` per `learnings/README.md`. If no, skip — don't write filler.

## See also

- `references/patterns.md` — Python/TypeScript/Go/Java/C# code for the mandatory patterns.
- `references/anti-patterns.md` — POSIX-only mistakes with the portable fix.
- `references/docker.md` — image hygiene, `.dockerignore` baseline, volume mounts.
- `references/platform-specific-code.md` — protocol + factory pattern for OS-conditional code.
- `learnings/` — project-specific platform constraints accumulated over time.
