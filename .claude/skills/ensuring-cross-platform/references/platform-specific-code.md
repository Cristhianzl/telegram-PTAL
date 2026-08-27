# Platform-specific code, when justified

Some functionality is genuinely OS-specific: filesystem watchers, GPU APIs, system notifications, hardware sensors, deep OS integrations. When that is the case, the goal is to **isolate the platform branch behind a single, narrow abstraction** so the rest of the codebase stays portable.

## 1. Isolate behind one protocol

Define a protocol / interface that captures the behavior, then have one factory pick the implementation at runtime.

**Python**
```python
# core/notifications.py — platform-agnostic interface
from typing import Protocol

class Notifier(Protocol):
    def notify(self, title: str, body: str) -> None: ...

# adapters/notifier_factory.py — runtime selection, single entry point
import sys

def get_notifier() -> Notifier:
    if sys.platform == "darwin":
        from .macos_notifier import MacOSNotifier
        return MacOSNotifier()
    if sys.platform.startswith("linux"):
        from .linux_notifier import LinuxNotifier
        return LinuxNotifier()
    if sys.platform == "win32":
        from .windows_notifier import WindowsNotifier
        return WindowsNotifier()
    raise RuntimeError(f"No notifier implementation for platform {sys.platform!r}")
```

**TypeScript**
```typescript
// core/notifier.ts
export interface Notifier {
  notify(title: string, body: string): Promise<void>;
}

// adapters/notifierFactory.ts
export async function getNotifier(): Promise<Notifier> {
  switch (process.platform) {
    case "darwin":  return (await import("./macosNotifier.js")).MacOSNotifier;
    case "linux":   return (await import("./linuxNotifier.js")).LinuxNotifier;
    case "win32":   return (await import("./windowsNotifier.js")).WindowsNotifier;
    default:
      throw new Error(`No notifier for platform ${process.platform}`);
  }
}
```

## 2. Detect at runtime, not at import time

Importing a Linux-only module unconditionally on Windows crashes the import — the entire app dies before `main` runs.

- Python: lazy `import` inside the factory branch.
- Node: dynamic `import()` inside the factory branch.
- Go: build tags (`//go:build linux`) — the compiler enforces it.

## 3. Test each implementation on its target platform

A `MacOSNotifier` that nobody runs in CI on a macOS runner is unverified code. The CI matrix MUST exercise each branch on the OS it targets.

If a branch can only be exercised manually (no headless variant of the OS API exists), document that in the test file and in the feature's documentation, and add a smoke-test instruction to the release checklist.

## 4. Document the branch map

In the feature documentation (see `documenting-features` skill), list every place the implementation differs per OS, why it differs, and the test that covers each branch. Example:

| Capability        | Linux                 | macOS                 | Windows               | Test                              |
|-------------------|-----------------------|-----------------------|-----------------------|------------------------------------|
| Notifications     | `notify-send` CLI     | `osascript` AppleScript | `WinRT.UI.Notifications` | `test_notifier_<platform>.py`     |
| Filesystem watch  | `inotify` (pyinotify) | `FSEvents`            | `ReadDirectoryChangesW` | `test_watcher_<platform>.py`     |

## 5. Failure mode for unsupported platforms

When a feature genuinely cannot work on a platform, the failure must be:

- **Loud** — raise `NotImplementedError` (or equivalent) with a clear message naming the platform and the feature.
- **Early** — detected at startup, not three actions deep into a user flow.
- **Documented** — listed in the README and the feature docs.

**Silent fallback to a no-op is forbidden.** A notification system that silently fails to notify on Windows is worse than one that refuses to start — the user thinks it worked.

Acceptable fallback: a degraded behavior the user is informed about. `Notifier` on Windows might log to stderr "system notifications unavailable; falling back to terminal output" and proceed. That is loud, explicit, and the user can detect it.

## When **not** to add a platform branch

Most code does not need this pattern. Reach for the platform-agnostic API first. Only add OS-conditional code when the language/runtime forces you to — typically:

- Direct system API calls (notifications, sensors, GPU, FS events).
- Process model differences (fork vs spawn, signal handling).
- Performance-critical paths where the portable abstraction is too slow.

If the language has an abstraction that works (e.g., `pathlib`, `tempfile`, `platformdirs`, `os.path`), **use it**. Adding a manual `if sys.platform == "win32"` branch where the stdlib already handles the difference is a code smell.
