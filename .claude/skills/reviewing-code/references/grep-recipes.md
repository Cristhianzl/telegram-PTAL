# Grep recipes for reviewers

Copy-paste-able commands to catch common violations during review. Adjust globs and patterns to the project's language(s).

---

## PII in logs

```bash
# Potential PII in logs — adjust patterns for your codebase
grep -rn "email" --include="*.py" --include="*.ts" --include="*.js" | grep -E "(log|print|console)"
grep -rn "first_name\|last_name\|full_name" --include="*.py" --include="*.ts" --include="*.js"
grep -rn "user\.phone\|user\.address\|user\.email" --include="*.py" --include="*.ts" --include="*.js" \
  | grep -E "(log|print|console)"
```

## Weak typing (TypeScript)

```bash
grep -rn ": any" --include="*.ts" --include="*.tsx"
grep -rn ": object" --include="*.ts" --include="*.tsx"
grep -rn "as any" --include="*.ts" --include="*.tsx"
```

## Empty catch blocks

```bash
grep -rn "catch.*{" -A1 --include="*.py" --include="*.ts" --include="*.js" \
  | grep -E "pass|//|{}"
```

## TODOs without a ticket

```bash
grep -rn "TODO" --include="*.py" --include="*.ts" --include="*.js" \
  | grep -v "TODO.*#\|TODO.*TICKET\|TODO.*JIRA"
```

## Stray debug statements

```bash
# JavaScript / TypeScript
grep -rn "console.log" --include="*.ts" --include="*.tsx" --include="*.js"

# Python
grep -rn "print(" --include="*.py"
```

## File size violations

```bash
# Find files with 500+ lines
find . -name "*.py" -o -name "*.ts" -o -name "*.js" | xargs wc -l | sort -n | tail -20
```

## Functions-per-file (Python)

```bash
for f in $(find . -name "*.py"); do
  count=$(grep -c "^def \|^    def " "$f" 2>/dev/null)
  if [ "$count" -gt 5 ]; then echo "$count functions: $f"; fi
done
```

## Cyclomatic complexity

```bash
# Python (radon)
radon cc -s -a -nc .
# Shows all functions ranked C or worse — anything ≥ CC 11 needs refactor

# TypeScript / JavaScript (ESLint rule)
# Add to .eslintrc:  "rules": { "complexity": ["error", 10] }
npx eslint . --rule '{"complexity": ["error", 10]}' --no-eslintrc

# Go
gocyclo -over 10 .
```

## Deep nesting (4+ levels of indentation)

```bash
grep -rn "                    " --include="*.py" --include="*.ts" --include="*.js" \
  | grep -v "test\|spec\|mock"
```

## Law of Demeter — long call chains (3+ dots)

```bash
grep -rn "\w\.\w\+\.\w\+\.\w\+" --include="*.py" --include="*.ts" --include="*.js" \
  | grep -v "import\|require\|from\|test\|spec\|mock\|node_modules"
```

## YAGNI — abstract classes with single implementation

```bash
grep -rn "class.*ABC\|abstract class\|interface " --include="*.py" --include="*.ts" -l | while read f; do
  class=$(grep -oP "class\s+\K\w+" "$f" | head -1)
  count=$(grep -rn "$class" --include="*.py" --include="*.ts" \
            | grep -v "import\|#\|//\|test\|spec" | wc -l)
  if [ "$count" -le 2 ]; then echo "Possible YAGNI: $class in $f (only $count references)"; fi
done
```

## AI / LLM red flags

```bash
# Direct DB access in AI / chat handlers
grep -rn "db\.\|repository\.\|query(" --include="*.ts" --include="*.py" \
  | grep -iE "(chat|llm|ai|bot|agent|prompt)"

# eval / exec on AI output
grep -rn "eval(\|exec(" --include="*.ts" --include="*.py" --include="*.js"

# Raw LLM output rendered without sanitization
grep -rn "innerHTML\|dangerouslySetInnerHTML\|render.*llm\|render.*ai" \
  --include="*.tsx" --include="*.jsx" --include="*.html"

# Secrets in system prompts
grep -rn "system_prompt\|SYSTEM_PROMPT" --include="*.ts" --include="*.py" \
  | grep -iE "(password|secret|key|token|conn|database)"

# Naive prompt building (system + user concatenation)
grep -rn "systemPrompt\s*+\s*userInput\|prompt.*\+.*req\.\|f\"{system_prompt}.*{user" \
  --include="*.ts" --include="*.py" --include="*.js"

# AI calls without explicit timeout
grep -rn "anthropic\|openai\|completion\|chat\.create\|messages\.create" --include="*.py" --include="*.ts" \
  | grep -v "timeout\|deadline"

# Logging full prompts / completions (PII / secret leak)
grep -rn "log.*prompt\|log.*completion\|logger.*messages" --include="*.py" --include="*.ts"

# Sync LLM call inside an HTTP handler / route
grep -rn -B2 "anthropic\|openai" --include="*.py" --include="*.ts" \
  | grep -E "router\.|@app\.|@router|def.*request|async def.*req"
```

## Third-party integration smells

```bash
# Signature compared with === (timing attack)
grep -rn "signature.*===\|=== .*signature\|hmac.*===\|=== .*hmac" --include="*.ts" --include="*.py" --include="*.js"

# Catching exceptions in verification functions (silent pass)
grep -rn "catch" --include="*.ts" --include="*.py" --include="*.js" -A2 \
  | grep -B1 "return true\|return { valid\|isValid.*true"

# Hardcoded webhook secrets
grep -rn "webhook.*secret\|WEBHOOK_SECRET\|signing.*secret" --include="*.ts" --include="*.py" --include="*.js" \
  | grep -v "process.env\|os.environ\|config\."
```

## Platform-agnostic red flags

```bash
# Hardcoded POSIX temp paths
grep -rnE '"/tmp/|"/var/|"~/\.config' --include="*.py" --include="*.ts" --include="*.js" --include="*.go" \
  | grep -v test/ | grep -v node_modules

# String concatenation forming paths
grep -rnE '["\x27][a-zA-Z_./-]+/["\x27]\s*\+' --include="*.py" --include="*.ts" --include="*.js"

# open() without encoding (Python)
grep -rnE 'open\([^)]*\)' --include="*.py" | grep -v "encoding=" | grep -v test_

# subprocess with shell=True
grep -rn "shell=True" --include="*.py"

# Hardcoded shell utilities in subprocess / system calls
grep -rnE '(os\.system|subprocess\.(run|call|Popen))\([^)]*["\x27](rm|cp|mv|ls|grep|cat|chmod) ' --include="*.py"

# Hardcoded interpreter
grep -rnE 'subprocess\.[a-z]+\(\s*\[?\s*["\x27]python3?["\x27]' --include="*.py"

# datetime.now() without UTC
grep -rn "datetime.now()" --include="*.py" | grep -v "timezone.utc" | grep -v test_

# split("\n") on text — likely will miss CRLF
grep -rnE '\.split\(["\x27]\\n["\x27]\)' --include="*.py" --include="*.ts" --include="*.js" | grep -v test_

# Forbidden POSIX-only constructs
grep -rn "os.fork\b\|os.posix_spawn\b" --include="*.py"
```

## Cohesion / "utils dump" red flag

```bash
# Catch-all helper file names
find . -name "utils.py" -o -name "helpers.py" -o -name "misc.py" -o -name "common.py" -o -name "shared.py" \
  | grep -v test/ | grep -v node_modules

# JS/TS equivalents
find . -name "utils.ts" -o -name "helpers.ts" -o -name "misc.ts" -o -name "common.ts" -o -name "shared.ts" \
  | grep -v test/ | grep -v node_modules
```

---

## Correctness: state machines & control-flow exceptions

Recipes for the third bug class (right on the happy path, silently wrong on an edge) live in
`correctness-checks.md`: auditing every enumeration site when an enum gains a member, and finding
broad `except`/`catch` blocks that swallow or rewrap control-flow signals (pause/cancel/retry).

---

## When to ignore grep results

- Test files often legitimately use patterns banned in production (e.g., `open()` without encoding for fixtures). Filter `test_` / `spec` / `__tests__` when checking production-only rules.
- Vendored / generated code is exempt. Filter `node_modules`, `vendor/`, `dist/`, `build/`, `__generated__/`, `proto/`.
- Some red flags are noise in specific languages (e.g., `: any` in TypeScript declaration files might be legitimate). Confirm in context before raising the finding.

---

## When to write your own grep

The recipes above catch common patterns. If your PR is in an area not covered (e.g., a new domain-specific anti-pattern the team has agreed to ban), add the grep to `learnings/` so it's available next time.
