# Structural checks during review

The non-security half of the review. File-structure limits, SOLID, pragmatic principles, architecture, code quality, error handling, observability, comments, testing.

For full rationale, see the `developing-features` skill. This file lists the **review-time** checks — what to look for in a diff.

---

## File structure — hard limits (Blockers when violated)

| Metric                                                                    | Max         |
|---------------------------------------------------------------------------|-------------|
| Lines of code (excluding imports, types, docs)                            | **500**     |
| Acceptable range with all other rules passing                             | 600–700     |
| Functions with **different** responsibilities                             | **5**       |
| Functions with **same** responsibility (same prefix)                      | **10**      |
| Main classes per file                                                     | **1**       |
| Cyclomatic complexity per function (new code)                             | **10**      |
| Cyclomatic complexity per function (legacy code being touched)            | **15** (only if your change doesn't increase complexity) |
| Maximum nesting depth per function                                        | **4**       |

A 650-line file passing all SRP and separation checks: acceptable. A 400-line file with mixed responsibilities: NOT acceptable. A 30-line function with cyclomatic complexity 14: NOT acceptable. Linear 80-line code: fine. Tangled 30-line code: not.

### Mandatory separation by responsibility

Functions with **different** prefixes from the table below MUST NOT coexist in the same file:

| Responsibility            | Function prefixes                                   | Must be in                       |
|---------------------------|------------------------------------------------------|----------------------------------|
| Validation                | `validate*`, `check*`, `is_valid*`, `assert*`        | `validation` file                |
| Formatting / Serialization| `format*`, `build*`, `serialize*`, `to_*`           | `formatting` / `serialization`   |
| Parsing                   | `parse*`, `extract*`, `from_*`                       | `parsing` file                   |
| External communication    | `fetch*`, `send*`, `call*`, `request*`               | `client` / `api` file            |
| Data persistence          | `save*`, `load*`, `find*`, `delete*`, `query*`       | `repository` file                |

### Cohesion — avoid over-engineering

- No files with only 1–2 trivial functions (< 20 lines total).
- Private helpers (`_func`) stay in the file that uses them.
- One-liner utilities are not extracted to separate files.

### File-level structural checks

- Types / models in dedicated `*_types` or `*_models` file.
- Constants in dedicated `*_constants` file.
- Each file can be described in **one sentence without "and" or "or"**.
- No generic file names: `utils`, `helpers`, `misc`, `common`, `shared` (standalone).
- No `index` files as main logic containers.

---

## SOLID — how to spot violations

### SRP

- Each file has one clear responsibility.
- Each function does one thing.
- Each class has one reason to change.
- No function names with "and", "or", "then", or vague "process".
- No monolithic files that "do everything".

**Good names:** `validateEmail()`, `formatCurrency()`, `fetchUserById()`, `parseResponse()`.
**Bad names:** `validateAndFormatUser()`, `processOrder()`, `handleEverything()`, `doStuff()`.

### OCP

- No functions with growing `if/elif/else` or `switch` chains for type/category handling.
- New feature requests should not require modifying existing tested functions.
- Polymorphism or strategy pattern used where 3+ variants exist and more are expected.

Red flag:

```python
if type == "credit": ...
elif type == "debit": ...
elif type == "pix": ...        # added this sprint
elif type == "boleto": ...     # added last sprint
```

### LSP

- No subclass throws `NotImplementedError` on an inherited method.
- No subclass silently ignores a method (empty body where parent does work).
- No subclass narrows accepted inputs compared to parent.
- No subclass adds side effects the parent doesn't have.
- Replacing parent with child in any caller would not break behavior.

### ISP

- No class has methods that are empty or return dummy values just to satisfy an interface.
- No interface has 7+ methods where most implementors use only 2–3.
- A change in one interface method does not force changes in classes that don't use it.

### DIP

- High-level modules do not directly instantiate low-level modules.
- Constructors receive abstractions, not concrete classes.
- Swapping an implementation (MySQL → Postgres) doesn't require changing business logic.

---

## Pragmatic principles — how to spot violations

### YAGNI

The reviewer's question: *"Is this code solving a problem that exists right now?"*

- No code solving future / imaginary requirements.
- No abstract base classes with only one implementation (and no near-term plan for a second).
- No configuration parameters nobody requested.
- No caching before a performance problem is measured.
- No generic plugin / event bus when only one plugin / event exists.

### Law of Demeter

- No call chains deeper than one dot from a direct dependency (`a.b.c.method()`).
- Callers ask for what they need, not dig through structures.

Red flag:

```python
order.customer.address.zip_code
```

Right:

```python
order.get_customer_zip()
```

### KISS

- No factory / strategy / builder pattern where a simple function suffices.
- No generics where a concrete type is clear and stable.
- No unnecessary abstraction layers (class wrapping a single function with no added value).

### DRY (with nuance)

- No duplicate types, classes, or 5+ line logic blocks across 2+ places.
- **But:** prefer duplication over wrong abstraction. If two blocks look similar by coincidence but change for different reasons, leave them duplicated.

---

## Architecture & layer separation

| Layer                       | CAN                                                    | CANNOT                                                       |
|-----------------------------|--------------------------------------------------------|--------------------------------------------------------------|
| Handler / Controller        | Receive request, delegate, return response             | Contain business logic, call DB directly                     |
| Service / Domain            | Business rules, orchestration                          | Know about HTTP, execute queries directly                    |
| Repository                  | Execute queries, map data                              | Make business decisions, call external APIs                  |
| Helper                      | Transform data, validate, format                       | Have side effects, do I/O, maintain state                    |
| External client             | Communicate with external services                     | Contain business logic, access the database                  |

- Domain logic independent from frameworks / infrastructure.
- Clear boundaries between modules and layers.
- No business logic in controllers / handlers / routes.
- No HTTP / DB concerns in service / domain layer.
- No framework imports in pure business logic.

---

## Code quality (clean code)

- **Immutability:** `const`, `readonly`, `final`, `val` where possible.
- **Strong typing:** no `any`, `object`, `dynamic`, `Object` loose types.
- **Early returns:** guard clauses to reduce nesting.
- **No magic values:** extract numbers / strings to named constants.
- **Dependency injection** where appropriate for testability.
- **No global state:** no singletons-as-state, no module-level mutable dicts.

### Naming

- Names reveal intent.
- Names consistent with codebase conventions.
- Boolean variables / functions use `is`, `has`, `can`, `should` prefix.
- Functions use verbs: `get`, `set`, `create`, `update`, `delete`, `validate`.
- No abbreviations unless universally understood (`id`, `url`, `api`).

---

## Error handling

- Expected errors handled explicitly.
- No silent failures (empty catch blocks).
- No generic exceptions (`Exception`, `Error`, raw `object`).
- Errors include meaningful context.
- Errors are part of the API contract.
- Inputs validated at system boundaries (fail fast).
- Clear distinction between recoverable errors and fatal exceptions.

Wrong vs right:

```python
# Wrong
try:
    result = do_something()
except:
    pass

# Right
try:
    result = do_something()
except ValidationError as e:
    logger.warning("Validation failed", extra={"error": str(e), "auth_id": user.auth_id})
    raise DomainError(f"Invalid input: {e.field}") from e
```

```typescript
// Wrong
try {
  const result = await doSomething();
} catch (e) {
  // silent
}

// Right
try {
  const result = await doSomething();
} catch (error) {
  if (error instanceof ValidationError) {
    logger.warn('Validation failed', { error: error.message, authId: user.authId });
    throw new DomainError(`Invalid input: ${error.field}`);
  }
  throw error;
}
```

---

## Observability (Recommended)

- Structured logging at key decision points.
- Appropriate log levels (`debug`, `info`, `warn`, `error`).
- Logs include: operation name, relevant IDs, outcome, duration (if relevant).
- No sensitive data in logs (passwords, tokens, PII).
- No redundant logs duplicating obvious information.
- No logs in no-op / stub implementations.

---

## Comments (Important)

- **Default: no comments.** Code should be self-explanatory through good naming.
- Comments only explain **WHY** the non-obvious decision was made (hidden constraint, subtle invariant, workaround for a specific bug). Never explain **WHAT** — the identifiers already do that.
- One short line max. No multi-paragraph docstrings on functions whose name and signature already tell the story.
- No section-divider comments (`# === Section ===`). If a section feels worth dividing, it's its own file.
- No comments referencing the current task or AI session ("added for the new flow", "used by X"). The PR description is for that.
- No commented-out code — use version control.
- No TODO without a ticket reference.

A diff that adds 12 lines of code and 8 lines of comments to "help future readers" is a finding — request the comments be removed unless each one's WHY would not be obvious without it.

---

## Testing (apply `writing-tests` rules)

### Quality

- Clear test names: `should_[expected]_when_[condition]`.
- Independent tests (no shared state between tests).
- Deterministic tests (no flakiness).
- External dependencies mocked / faked (DB, APIs, filesystem, time).
- Tests validate behavior, not implementation details.
- Clear structure: Arrange-Act-Assert.

### Tests must challenge the code, not only confirm it

- Both happy path AND adversarial tests exist.
- Edge cases and boundary values — null, empty, zero, max, one past the limit.
- Invalid and unexpected inputs — wrong types, malformed data, missing fields.
- Error states and failure modes.
- What should NOT happen (e.g., "should not render delete button for read-only users").
- Error messages and error types verified — not just that "it fails".

Reviewer questions:

1. Are there both happy path and adversarial tests?
2. Would these tests catch a regression if someone broke the logic?
3. Are there edge cases or failure modes not tested?
4. If I remove a line of business logic, would at least one test fail?

### Coverage validation

- Coverage was run AND output shown (backend + frontend if both).
- Coverage ≥ 75% (target 80%) on all tested source code.
- Coverage report was actually executed, not just claimed.
- All created tests pass — zero failures.

---

## Platform agnosticism — reviewer checklist

For every PR (see `ensuring-cross-platform` skill for the full rules):

- **Paths**: every path constructed via `Path` / `path.join`? Any string concatenation with `/` or `\`?
- **Encoding**: every `open()` / `read_text()` / `readFileSync()` specifies encoding?
- **Shell**: any `shell=True`? Any string-form subprocess command? Any literal `rm`, `cp`, `mv`, `ls`, `grep`, `cat`?
- **Interpreter**: any hardcoded `python3`, `node`, `bash`? Should be `sys.executable` / `process.execPath`.
- **Temp / config dirs**: any literal `/tmp`, `/var/`, `~/.config`, `C:\…`?
- **Line endings**: any `text.split("\n")` or hardcoded `"\n".join` written to text files?
- **External tools**: reliance on `git`, `make`, `curl`, etc. without `shutil.which` check or documented dependency?
- **Time**: any `datetime.now()` without `timezone.utc`?
- **CI**: does the CI configuration cover every supported platform? Was a matrix entry quietly removed?

### Red flags that often hide portability bugs

- A new file with `import os` but no `import pathlib` — suspect string-based path manipulation.
- `@pytest.mark.skipif(sys.platform == "win32")` without an explanation of the alternative coverage.
- Helpers named `run_command`, `shell_exec`, `exec_bash` — almost certainly hardcoded shell semantics.
- `try/except OSError: pass` around a filesystem operation — OS-specific failure being silently swallowed.
- A diff that touches CI configuration alongside production code — verify the matrix wasn't trimmed.
- A `Dockerfile` change without a corresponding `.dockerignore` review.

---

## Legacy code awareness

- Do NOT prolong bad patterns — even if surrounding code is bad, write good code.
- Do NOT copy-paste from legacy code without reviewing quality.
- Flag legacy patterns you encounter for future cleanup (don't fix them in unrelated PRs).
- Isolate new code from legacy where possible.
- Document any necessary interaction with legacy systems.
