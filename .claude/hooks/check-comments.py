#!/usr/bin/env python3
import json
import re
import sys
from pathlib import Path

CODE_EXTS = {".py", ".pyi", ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".go", ".java", ".kt", ".cs", ".rs", ".rb", ".swift"}
SKIP_PARTS = {"node_modules", ".venv", "venv", "__pycache__", "build", "dist", "target", ".next", "__generated__", ".pytest_cache", ".mypy_cache", ".ruff_cache"}
SKIP_NAME_PATTERNS = (re.compile(r"^test_"), re.compile(r"_test\.(py|go|ts|tsx|js|jsx)$"), re.compile(r"\.test\.(ts|tsx|js|jsx)$"), re.compile(r"\.spec\.(ts|tsx|js|jsx)$"))
SKIP_DIR_NAMES = {"tests", "test", "__tests__", "spec", "fixtures", "migrations", "alembic"}

ALLOW_RE = re.compile(
    r"^\s*(#!|"
    r"#\s*-\*-\s*coding|"
    r"#\s*noqa|#\s*type:\s*ignore|#\s*pylint:|#\s*pyright:|"
    r"//\s*eslint-disable|//\s*@ts-(ignore|expect-error|nocheck)|"
    r"///|//!|"
    r"#\s*Why:|//\s*Why:|"
    r"#\s*(Arrange|Act|Assert)\b|//\s*(Arrange|Act|Assert)\b|"
    r"#\s*(TODO|FIXME|HACK|XXX)\s*[\(\[](?:[A-Z]+-\d+|GH-?\d+|#\d+)|"
    r"//\s*(TODO|FIXME|HACK|XXX)\s*[\(\[](?:[A-Z]+-\d+|GH-?\d+|#\d+))",
    re.IGNORECASE,
)
DIVIDER_RE = re.compile(r"^\s*(#|//)\s*[=\-*]{3,}.*[=\-*]{3,}\s*$")
SECTION_BANNER_RE = re.compile(r"^\s*(#|//)\s*={2,}\s+\S.*\s+={2,}\s*$")
PY_COMMENT_RE = re.compile(r"^\s*#")
JS_COMMENT_RE = re.compile(r"^\s*//")
LICENSE_RE = re.compile(r"(copyright|licensed|spdx-license-identifier|apache license|mit license)", re.IGNORECASE)
PY_IMPORT_RE = re.compile(r"^\s*(import\s+|from\s+\S+\s+import\s+)")
JS_IMPORT_RE = re.compile(r"^\s*(import\s+|export\s+(\*|\{|default)\s+from|use\s+|using\s+|package\s+|#include\b)")

MAX_COMMENT_RATIO = 0.10
MAX_CONSECUTIVE_COMMENTS = 3


def should_skip(path: Path) -> bool:
    if path.suffix not in CODE_EXTS:
        return True
    parts = set(path.parts)
    if parts & SKIP_PARTS or parts & SKIP_DIR_NAMES:
        return True
    name = path.name
    for pat in SKIP_NAME_PATTERNS:
        if pat.search(name):
            return True
    return False


def is_comment_line(line: str, lang: str) -> bool:
    if lang == "py":
        return bool(PY_COMMENT_RE.match(line))
    return bool(JS_COMMENT_RE.match(line))


def is_import_line(line: str, lang: str) -> bool:
    if lang == "py":
        return bool(PY_IMPORT_RE.match(line))
    return bool(JS_IMPORT_RE.match(line))


def detect_lang(path: Path) -> str:
    if path.suffix in {".py", ".pyi"}:
        return "py"
    return "js"


def check(content: str, path: Path) -> list[str]:
    lang = detect_lang(path)
    lines = content.splitlines()
    issues: list[str] = []
    code_lines = 0
    counted_comments = 0
    consecutive = 0
    flagged_run = False

    in_license_zone = True
    for lineno, raw in enumerate(lines, start=1):
        stripped = raw.strip()
        if not stripped:
            consecutive = 0
            continue
        if in_license_zone and lineno <= 20 and LICENSE_RE.search(stripped):
            continue
        if lineno > 5:
            in_license_zone = False

        is_comment = is_comment_line(raw, lang)
        if is_comment:
            if SECTION_BANNER_RE.match(raw) or DIVIDER_RE.match(raw):
                issues.append(f"{path}:{lineno}: section/divider comment — split into its own file instead of a divider")
                consecutive = 0
                continue
            if ALLOW_RE.match(raw):
                consecutive = 0
                continue
            counted_comments += 1
            consecutive += 1
            if consecutive >= MAX_CONSECUTIVE_COMMENTS and not flagged_run:
                issues.append(f"{path}:{lineno}: {consecutive}+ consecutive comment lines — rewrite as prose-around-example or remove (rule: WHAT-comments forbidden)")
                flagged_run = True
            continue

        consecutive = 0
        if is_import_line(raw, lang):
            continue
        code_lines += 1

    if code_lines >= 10 and counted_comments / max(code_lines, 1) > MAX_COMMENT_RATIO:
        pct = round(counted_comments / code_lines * 100, 1)
        issues.append(f"{path}: comment density {pct}% over {code_lines} code lines (limit 10%) — code should be self-explanatory; reserve comments for non-obvious WHY only")

    return issues


def main() -> None:
    try:
        data = json.load(sys.stdin)
    except Exception:
        sys.exit(0)

    if data.get("tool_name") not in {"Write", "Edit", "MultiEdit"}:
        sys.exit(0)

    file_path = data.get("tool_input", {}).get("file_path")
    if not file_path:
        sys.exit(0)

    path = Path(file_path)
    if not path.exists() or should_skip(path):
        sys.exit(0)

    try:
        content = path.read_text(encoding="utf-8", errors="replace")
    except Exception:
        sys.exit(0)

    issues = check(content, path)
    if issues:
        print("Comment policy violation:", file=sys.stderr)
        for issue in issues:
            print(f"  - {issue}", file=sys.stderr)
        print("\nRule: default to no comments. Only WHY-comments where the reason is non-obvious. No section dividers. See developing-features/SKILL.md.", file=sys.stderr)
        sys.exit(2)


if __name__ == "__main__":
    main()
