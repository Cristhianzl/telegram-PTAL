#!/usr/bin/env python3
# Why: PostToolUse hooks scan the whole file, so editing one line in a legacy file would flag pre-existing debt; these helpers isolate just the operation's delta (new lines + reconstructed previous content).
from __future__ import annotations

import json
import sys
from pathlib import Path


def read_payload() -> dict:
    try:
        return json.load(sys.stdin)
    except Exception:
        sys.exit(0)


def get_path(data: dict) -> Path | None:
    fp = data.get("tool_input", {}).get("file_path")
    return Path(fp) if fp else None


def _line_span(content: str, offset: int, text: str) -> tuple[int, int]:
    start = content.count("\n", 0, offset) + 1
    return start, start + text.count("\n")


def _spans_for(content: str, new_string: str, replace_all: bool) -> list[tuple[int, int]]:
    if not new_string:
        return []
    spans: list[tuple[int, int]] = []
    pos = 0
    while True:
        off = content.find(new_string, pos)
        if off == -1:
            break
        spans.append(_line_span(content, off, new_string))
        if not replace_all:
            break
        pos = off + len(new_string)
    return spans


def changed_lines(data: dict, content: str) -> set[int] | None:
    # None = the operation owns the whole file (Write) -> check everything.
    tool = data.get("tool_name")
    ti = data.get("tool_input", {})
    if tool == "Write":
        return None
    spans: list[tuple[int, int]] = []
    if tool == "Edit":
        spans = _spans_for(content, ti.get("new_string", ""), bool(ti.get("replace_all")))
    elif tool == "MultiEdit":
        for edit in ti.get("edits", []):
            spans += _spans_for(content, edit.get("new_string", ""), bool(edit.get("replace_all")))
    else:
        return None
    lines: set[int] = set()
    for a, b in spans:
        lines.update(range(a, b + 1))
    return lines


def previous_content(data: dict, content: str) -> str | None:
    # Revert new_string -> old_string to reconstruct the prior file; None when it cannot be reconstructed (Write overwrites without old).
    tool = data.get("tool_name")
    ti = data.get("tool_input", {})
    if tool == "Edit":
        old, new = ti.get("old_string", ""), ti.get("new_string", "")
        return content.replace(new, old, -1 if ti.get("replace_all") else 1)
    if tool == "MultiEdit":
        prev = content
        for edit in reversed(ti.get("edits", [])):
            old, new = edit.get("old_string", ""), edit.get("new_string", "")
            prev = prev.replace(new, old, -1 if edit.get("replace_all") else 1)
        return prev
    return None
