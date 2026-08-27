#!/usr/bin/env bash
# Why: permission rules match a command prefix, so a mutation hidden after && or ; in a compound
# line slips past them; this hook inspects the whole command string before it runs.

set -uo pipefail

input="$(cat)"
cmd="$(printf '%s' "$input" | jq -r '.tool_input.command // empty' 2>/dev/null)"
[ -z "$cmd" ] && exit 0

normalized="$(printf '%s' "$cmd" | tr '\n' ' ' | tr -s ' ')"

deny() {
  printf 'BLOCKED (read-only reviewer): %s\n' "$1" >&2
  printf 'This config reviews pull requests; it never mutates a repository.\n' >&2
  printf 'Put the change in a finding and let the author apply it. The human posts the review.\n' >&2
  exit 2
}

case "$normalized" in
  *"git commit"*)   deny "git commit — only the human commits" ;;
  *"git add"*)      deny "git add — the reviewer does not stage" ;;
  *"git push"*)     deny "git push" ;;
  *"git merge"*)    deny "git merge" ;;
  *"git rebase"*)   deny "git rebase" ;;
  *"git reset"*)    deny "git reset" ;;
  *"git revert"*)   deny "git revert" ;;
  *"git restore"*)  deny "git restore" ;;
  *"git clean"*)    deny "git clean" ;;
  *"git cherry-pick"*) deny "git cherry-pick" ;;
  *"git tag"*)      deny "git tag" ;;
  *"git branch -d"*|*"git branch -D"*) deny "git branch deletion" ;;
  *"gh pr review"*) deny "gh pr review — the human posts the review" ;;
  *"gh pr comment"*) deny "gh pr comment — the human posts the review" ;;
  *"gh pr merge"*)  deny "gh pr merge" ;;
  *"gh pr close"*)  deny "gh pr close" ;;
  *"gh pr edit"*)   deny "gh pr edit" ;;
  *"gh pr create"*) deny "gh pr create" ;;
  *"gh pr ready"*)  deny "gh pr ready" ;;
  *"gh issue"*)     deny "gh issue — the reviewer does not touch the tracker" ;;
  *"gh release"*)   deny "gh release" ;;
  *"gh api"*)
    case "$normalized" in
      *" -X POST"*|*" -X PATCH"*|*" -X PUT"*|*" -X DELETE"*|*"--method POST"*|*"--method PATCH"*|*"--method PUT"*|*"--method DELETE"*)
        deny "gh api with a mutating method" ;;
    esac
    ;;
esac

exit 0
