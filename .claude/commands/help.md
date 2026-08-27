---
description: List the available commands
---

# /help

## Session

|Command|What it does|
|-------|---------|
|`/init`|Session briefing (branch, phase, next tasks)|
|`/check`|Health check (lint + tests on the changed areas)|
|`/test`|Runs tests for the modified files|
|`/review`|Review the diff against the rules|

## Deliver

|Command|What it does|
|-------|---------|
|`/commit [msg]`|Lint + tests + commit|
|`/done [msg]`|Review + check + commit|
|`/push`|Push the current branch|
|`/pr [title]`|Create a PR on GitHub|

## Roadmap

|Command|What it does|
|-------|---------|
|`/roadmap`|Status of all phases|
|`/next`|Pick the next pending task|
|`/task [name]`|Mark a task as done|

## Project health

|Command|What it does|
|-------|---------|
|`/sync`|Check docs and configs vs code|
|`/security`|Security audit|

## Typical flow

```text
/init                  # start
/next                  # pick a task
... code ...
/done "feat: ..."      # finalize
/push                  # send
/pr                    # open PR
```
