#!/usr/bin/env python3
import json
import re
import sys

CURL_RE = re.compile(r"\bcurl\s", re.IGNORECASE)
LOCAL_URL_RE = re.compile(r"https?://(localhost|127\.0\.0\.1|0\.0\.0\.0)[:/]", re.IGNORECASE)

CONTEXT = (
    "The user's request includes a cURL command or a local endpoint. Treat it as the ACCEPTANCE TEST: "
    "validate the work against the real running system, not by reading code. "
    "Run the request BEFORE the change (prove the bug / missing behavior) and AFTER (prove the fix/feature), "
    "paste the actual responses, check the persisted state in the database, and drive the UI end-to-end "
    "when the flow has a frontend. No assumptions, no 'should work' — real requests, real data, real evidence. "
    "Follow skills/validating-in-reality."
)


def main() -> None:
    try:
        data = json.load(sys.stdin)
    except Exception:
        sys.exit(0)

    prompt = data.get("prompt") or ""
    if not (CURL_RE.search(prompt) or LOCAL_URL_RE.search(prompt)):
        sys.exit(0)

    print(json.dumps({
        "hookSpecificOutput": {
            "hookEventName": "UserPromptSubmit",
            "additionalContext": CONTEXT,
        }
    }))


if __name__ == "__main__":
    main()
