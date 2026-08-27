#!/usr/bin/env python3
import json
import re
import sys
from pathlib import Path

CODE_EXTS = {".py", ".pyi", ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".go", ".java", ".kt", ".cs", ".rs", ".rb", ".swift"}
SKIP_PARTS = {"node_modules", ".venv", "venv", "__pycache__", "build", "dist", "target", ".next", "__generated__", ".pytest_cache", ".mypy_cache", ".ruff_cache", "migrations", "alembic"}
SKIP_DIR_NAMES = {"tests", "test", "__tests__", "spec", "fixtures"}
SKIP_NAME_PATTERNS = (re.compile(r"^test_"), re.compile(r"_test\.(py|go|ts|tsx|js|jsx)$"), re.compile(r"\.test\.(ts|tsx|js|jsx)$"), re.compile(r"\.spec\.(ts|tsx|js|jsx)$"), re.compile(r"^conftest\.py$"))

PY_IMPORT_RE = re.compile(r"^\s*(import\s+|from\s+\S+\s+import\s+)")
JS_IMPORT_RE = re.compile(r"^\s*(import\s+|export\s+(\*|\{|default)\s+from|use\s+|using\s+|package\s+|#include\b)")
PY_COMMENT_RE = re.compile(r"^\s*#")
JS_COMMENT_RE = re.compile(r"^\s*//")

HARD_LIMIT = 500
FLEX_LIMIT = 700


def should_skip(path: Path) -> bool:
    if path.suffix not in CODE_EXTS:
        return True
    parts = set(path.parts)
    if parts & SKIP_PARTS or parts & SKIP_DIR_NAMES:
        return True
    for pat in SKIP_NAME_PATTERNS:
        if pat.search(path.name):
            return True
    return False


def count_code_lines(content: str, path: Path) -> int:
    is_py = path.suffix in {".py", ".pyi"}
    count = 0
    in_block_string = False
    for raw in content.splitlines():
        stripped = raw.strip()
        if not stripped:
            continue
        if is_py:
            triple_double = stripped.count('"""')
            triple_single = stripped.count("'''")
            if triple_double % 2 == 1 or triple_single % 2 == 1:
                in_block_string = not in_block_string
                continue
            if in_block_string:
                continue
            if PY_COMMENT_RE.match(raw):
                continue
            if PY_IMPORT_RE.match(raw):
                continue
        else:
            if JS_COMMENT_RE.match(raw):
                continue
            if JS_IMPORT_RE.match(raw):
                continue
        count += 1
    return count


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

    loc = count_code_lines(content, path)
    if loc > FLEX_LIMIT:
        print(f"File size violation: {path} has {loc} code lines (hard cap {FLEX_LIMIT}).", file=sys.stderr)
        print("Split into files by responsibility (validation, formatting, repository, etc.). See developing-features/references/file-structure.md.", file=sys.stderr)
        sys.exit(2)
    if loc > HARD_LIMIT:
        print(f"File size warning: {path} has {loc} code lines (soft limit {HARD_LIMIT}).", file=sys.stderr)
        print(f"Acceptable only if SRP, separation by responsibility, and no mixed prefixes hold. Otherwise split. Beyond {FLEX_LIMIT} is blocked.", file=sys.stderr)


if __name__ == "__main__":
    main()
