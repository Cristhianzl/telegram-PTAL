#!/usr/bin/env python3
import re
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from _hooklib import changed_lines, get_path, read_payload  # noqa: E402

SKIP_PARTS = {"node_modules", ".venv", "venv", "__pycache__", "build", "dist", "target", ".next", "__generated__", ".pytest_cache", ".mypy_cache", ".ruff_cache"}
TEST_DIR_NAMES = {"tests", "test", "__tests__", "spec", "fixtures"}
TEST_NAME_PATTERNS = (re.compile(r"^test_"), re.compile(r"_test\.(py|go|ts|tsx|js|jsx)$"), re.compile(r"\.test\.(ts|tsx|js|jsx)$"), re.compile(r"\.spec\.(ts|tsx|js|jsx)$"), re.compile(r"^conftest\.py$"))
SCRIPT_DIR_NAMES = {"scripts", "bin", "tools", "hooks"}
DTS_RE = re.compile(r"\.d\.ts$")

RULES = [
    {
        "id": "shell-true",
        "exts": {".py"},
        "skip_tests": True,
        "regex": re.compile(r"\bshell\s*=\s*True\b"),
        "message": "subprocess shell=True is forbidden. Use a list of args, no shell. See ensuring-cross-platform/references/patterns.md.",
    },
    {
        "id": "ts-any",
        "exts": {".ts", ".tsx"},
        "skip_tests": True,
        "skip_dts": True,
        "regex": re.compile(r":\s*any\b(?!\s*\[\])"),
        "message": ": any is forbidden. Strong-type or use unknown + a type guard.",
    },
    {
        "id": "py-open-no-encoding",
        "exts": {".py"},
        "skip_tests": True,
        "regex": re.compile(r"\bopen\s*\([^)]*\)"),
        "extra_check": "open_missing_encoding",
        "message": "open() must specify encoding (default UTF-8). See ensuring-cross-platform/references/patterns.md.",
    },
    {
        "id": "py-datetime-now-no-utc",
        "exts": {".py"},
        "skip_tests": True,
        "regex": re.compile(r"datetime\.now\s*\(\s*\)"),
        "message": "datetime.now() without timezone is forbidden. Use datetime.now(timezone.utc).",
    },
    {
        "id": "hardcoded-paths",
        "exts": {".py", ".ts", ".tsx", ".js", ".jsx", ".go"},
        "skip_tests": True,
        "regex": re.compile(r"""["'](/tmp/|/var/|~/\.config/|C:\\\\|C:/Users)"""),
        "message": "Hardcoded platform-specific path. Use tempfile / platformdirs / equivalent. See ensuring-cross-platform.",
    },
    {
        "id": "console-log",
        "exts": {".ts", ".tsx", ".js", ".jsx"},
        "skip_tests": True,
        "skip_dts": True,
        "regex": re.compile(r"\bconsole\.log\s*\("),
        "message": "console.log in production code is forbidden. Use the project logger.",
    },
    {
        "id": "py-print",
        "exts": {".py"},
        "skip_tests": True,
        "skip_scripts": True,
        "regex": re.compile(r"^\s*print\s*\("),
        "multiline": True,
        "message": "print() in production code is forbidden. Use structlog/loguru/the project logger.",
    },
    {
        "id": "py-os-fork",
        "exts": {".py"},
        "regex": re.compile(r"\bos\.fork\s*\("),
        "message": "os.fork is POSIX-only and breaks on Windows. Use multiprocessing with spawn.",
    },
    {
        "id": "eval-exec",
        "exts": {".py", ".ts", ".tsx", ".js", ".jsx"},
        "skip_tests": True,
        "regex": re.compile(r"^\s*\b(eval|exec)\s*\("),
        "multiline": True,
        "message": "eval/exec on dynamic input is forbidden. Sandbox or refactor.",
    },
    {
        "id": "dangerous-html",
        "exts": {".tsx", ".jsx"},
        "regex": re.compile(r"\bdangerouslySetInnerHTML\b"),
        "message": "dangerouslySetInnerHTML enables stored XSS. Sanitize and use safe rendering.",
    },
    {
        "id": "py-bare-except",
        "exts": {".py"},
        "regex": re.compile(r"^\s*except\s*:\s*$"),
        "multiline": True,
        "message": "Bare 'except:' swallows everything including KeyboardInterrupt. Catch specific exceptions.",
    },
    {
        "id": "ts-as-any",
        "exts": {".ts", ".tsx"},
        "skip_tests": True,
        "skip_dts": True,
        "regex": re.compile(r"\bas\s+any\b"),
        "message": "'as any' bypasses the type system. Type properly or use 'unknown' + guards.",
    },
    {
        "id": "localStorage-token",
        "exts": {".ts", ".tsx", ".js", ".jsx"},
        "regex": re.compile(r"localStorage\.setItem\s*\([^)]*[Tt]oken"),
        "message": "Never store tokens in localStorage. Use httpOnly cookies with Secure + SameSite.",
    },
]


def is_test_path(path: Path) -> bool:
    parts = set(path.parts)
    if parts & TEST_DIR_NAMES:
        return True
    for pat in TEST_NAME_PATTERNS:
        if pat.search(path.name):
            return True
    return False


def is_script_path(path: Path) -> bool:
    return bool(set(path.parts) & SCRIPT_DIR_NAMES)


def open_missing_encoding(line: str) -> bool:
    return "open(" in line and "encoding=" not in line


def datetime_now_utc_present(line: str) -> bool:
    return "timezone" in line or "tzinfo" in line or "UTC" in line


def run_rule(rule: dict, content: str, path: Path) -> list[tuple[int, str]]:
    if path.suffix not in rule["exts"]:
        return []
    if set(path.parts) & SKIP_PARTS:
        return []
    if rule.get("skip_tests") and is_test_path(path):
        return []
    if rule.get("skip_scripts") and is_script_path(path):
        return []
    if rule.get("skip_dts") and DTS_RE.search(path.name):
        return []

    hits: list[tuple[int, str]] = []
    for lineno, raw in enumerate(content.splitlines(), start=1):
        if not rule["regex"].search(raw):
            continue
        if rule["id"] == "py-open-no-encoding":
            if not open_missing_encoding(raw):
                continue
            if "encoding=" in raw:
                continue
            if "open_text" in raw or "io.open" in raw and "encoding=" in raw:
                continue
        if rule["id"] == "py-datetime-now-no-utc":
            if datetime_now_utc_present(raw):
                continue
        if rule["id"] == "hardcoded-paths" and any(s in raw for s in ("os.environ", "Path(__file__)", "tempfile.")):
            continue
        hits.append((lineno, raw.strip()[:120]))
    return hits


def main() -> None:
    data = read_payload()

    if data.get("tool_name") not in {"Write", "Edit", "MultiEdit"}:
        sys.exit(0)

    path = get_path(data)
    if not path or not path.exists():
        sys.exit(0)

    try:
        content = path.read_text(encoding="utf-8", errors="replace")
    except Exception:
        sys.exit(0)

    # Only report patterns on lines the operation wrote; pre-existing debt elsewhere does not block the edit.
    changed = changed_lines(data, content)

    all_hits: list[str] = []
    for rule in RULES:
        for lineno, line in run_rule(rule, content, path):
            if changed is not None and lineno not in changed:
                continue
            all_hits.append(f"  - {path}:{lineno} [{rule['id']}] {line}\n    → {rule['message']}")

    if all_hits:
        print("Banned-pattern violation:", file=sys.stderr)
        for h in all_hits:
            print(h, file=sys.stderr)
        sys.exit(2)


if __name__ == "__main__":
    main()
