# Pre-commit validation — 8 steps

Test code is production code. Same quality bar. Below is the mandatory sequence before declaring any test task complete. Per-language commands included so you can copy/paste.

---

## Step 1 — Run each new test in isolation

Every new test MUST pass when executed alone. Verifies no dependence on side effects from earlier tests.

```bash
# Python (pytest)
pytest tests/users/test_user_service.py::test_should_return_error_when_email_is_invalid -v

# JavaScript / TypeScript (Jest)
npx jest tests/users/user.service.test.ts -t "should return error when email is invalid"

# JavaScript / TypeScript (Vitest)
npx vitest run tests/users/user.service.test.ts

# Go
go test ./users/ -run TestCreateUser_InvalidEmail -v

# Java (Maven)
mvn test -Dtest=UserServiceTest#shouldReturnErrorWhenEmailIsInvalid

# C# (.NET)
dotnet test --filter "FullyQualifiedName~UserServiceTests.ShouldReturnErrorWhenEmailIsInvalid"
```

---

## Step 2 — Run in random order

Tests MUST pass regardless of execution order. Catches hidden state coupling.

```bash
# Python (pytest)
pip install pytest-randomly
pytest --randomly-seed=random -v

# Jest
npx jest --randomize

# Vitest
npx vitest run --sequence.shuffle

# Go (1.17+)
go test ./... -shuffle=on -v

# JUnit 5 — configure in junit-platform.properties:
#   junit.jupiter.testmethod.order.default=org.junit.jupiter.api.MethodOrderer$Random

# xUnit (.NET) randomizes within a class by default
dotnet test
```

If any test fails in random order → it has a hidden dependency. Fix it immediately.

---

## Step 3 — Run all new tests together

Catches concurrency issues, shared-state leaks, race conditions between your new tests.

```bash
# pytest
pytest tests/users/test_user_service_creation.py tests/users/test_user_service_permissions.py -v
pytest tests/users/ -v

# Jest
npx jest tests/users/user.service.creation.test.ts tests/users/user.service.permissions.test.ts

# Vitest
npx vitest run tests/users/

# Go
go test ./users/... -v

# Maven
mvn test -Dtest="UserServiceCreationTest,UserServicePermissionsTest"

# .NET
dotnet test --filter "FullyQualifiedName~UserService"
```

If any test fails together but passes alone → shared state or concurrency. Fix before continuing.

---

## Step 4 — Verify coverage on changed code (intermediate)

Run coverage on the modules you changed. Focus on **uncovered branches**, not just percentages.

```bash
# pytest-cov
pytest tests/users/ --cov=src/users --cov-report=term-missing --cov-branch

# Jest
npx jest tests/users/ --coverage --collectCoverageFrom="src/users/**/*.ts"

# Vitest
npx vitest run tests/users/ --coverage

# Go
go test ./users/... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out

# Maven + JaCoCo
mvn test -Dtest="UserServiceCreationTest,UserServicePermissionsTest" jacoco:report

# .NET
dotnet test --filter "FullyQualifiedName~UserService" --collect:"XPlat Code Coverage"
```

This is an intermediate check during development. The final, mandatory coverage report is **Step 7**.

---

## Step 5 — Lint, format, type-check test files

Test code MUST meet the same quality bar as production. Run all configured tools on test files.

Identify the project's tools first:

| Tool                          | Config files                                                                 |
|-------------------------------|------------------------------------------------------------------------------|
| Python linter/formatter       | `ruff.toml`, `pyproject.toml [tool.ruff]`, `.flake8`, `.pylintrc`, `mypy.ini` |
| JS/TS linter/formatter        | `.eslintrc.*`, `eslint.config.*`, `.prettierrc.*`, `biome.json`, `tsconfig.json` |
| Go                            | `.golangci.yml`, `go vet`                                                    |
| Java/Kotlin                   | `checkstyle.xml`, `.editorconfig`, `spotless`, `ktlint`                      |
| C#                            | `.editorconfig`, `StyleCop`, `dotnet format` config                          |
| Rust                          | `clippy`, `rustfmt.toml`                                                     |

Then run them on the test files:

```bash
# Python — Ruff
ruff check tests/users/
ruff check tests/users/ --fix
ruff format tests/users/ --check
ruff format tests/users/

# Python — mypy
mypy tests/users/

# Python — flake8
flake8 tests/users/

# ESLint + Prettier
npx eslint tests/users/ --fix
npx prettier tests/users/ --write --check

# Biome
npx biome check tests/users/ --write

# Go
gofmt -w ./users/*_test.go
go vet ./users/...
golangci-lint run ./users/...

# Maven
mvn checkstyle:check
mvn spotless:apply

# .NET
dotnet format --verify-no-changes
dotnet format

# Rust
cargo fmt -- --check
cargo clippy --tests
```

**Rules:**

- Fix **all** reported errors and warnings.
- Run BOTH linter AND formatter — they check different things.
- Run the type checker if configured.
- Do NOT disable rules inline (`# noqa`, `# type: ignore`, `eslint-disable`, `@ts-ignore`, `@SuppressWarnings`) without a `Why:` comment.
- Do NOT skip because "it's just test code".

---

## Step 6 — All created/modified tests pass

**This is the most important step. Zero failures, zero exceptions.**

After running individually (Step 1), random order (Step 2), and together (Step 3), do a final combined run and confirm **100% green**.

```bash
pytest tests/users/test_user_service_creation.py tests/users/test_user_service_permissions.py -v
npx jest tests/users/user.service.creation.test.ts tests/users/user.service.permissions.test.ts
npx vitest run tests/users/user.service.creation.test.ts tests/users/user.service.permissions.test.ts
go test ./users/... -v
mvn test -Dtest="UserServiceCreationTest,UserServicePermissionsTest"
dotnet test --filter "FullyQualifiedName~UserService"
```

- All tests pass — zero failures, zero errors, zero skips.
- Never `@skip` / `@ignore` / `xit` / `@Disabled` to hide a failure.
- Never delete a test you wrote because it's failing.
- Never leave "to fix later" — fix now.

---

## Step 7 — Coverage report for ALL created tests, ≥ 75% (target 80%)

**Mandatory. Always run. Always show the output to the user.**

For every test file you created — backend AND frontend.

### 7.1 — Backend

```bash
# Single module
pytest tests/unit/agentic/services/test_assistant_service.py \
  --cov=src/backend/base/langflow/agentic/services/assistant_service \
  --cov-report=term-missing --cov-branch -v

# Multiple files, same module
pytest tests/unit/agentic/services/test_assistant_service_creation.py \
       tests/unit/agentic/services/test_assistant_service_streaming.py \
  --cov=src/backend/base/langflow/agentic/services \
  --cov-report=term-missing --cov-branch -v

# Entire directory
pytest tests/unit/agentic/ \
  --cov=src/backend/base/langflow/agentic \
  --cov-report=term-missing --cov-branch -v
```

### 7.2 — Frontend

```bash
# Jest — single file
npx jest src/components/assistantPanel/__tests__/assistant-panel.test.tsx \
  --coverage \
  --collectCoverageFrom="src/components/assistantPanel/**/*.{ts,tsx}"

# Jest — multiple files
npx jest src/components/assistantPanel/__tests__/assistant-header.test.tsx \
         src/components/assistantPanel/__tests__/assistant-input.test.tsx \
  --coverage \
  --collectCoverageFrom="src/components/assistantPanel/components/**/*.{ts,tsx}"

# Vitest
npx vitest run src/components/assistantPanel/__tests__/ --coverage
```

### 7.3 — What to do with the output

1. **Run** the commands for ALL your test files (backend + frontend).
2. **Read** the coverage output — per-file percentage and the `Missing` column.
3. **Show** the full output to the user (paste the terminal output).
4. **Check** every file is ≥ 80% (or 75% minimum after reasonable effort).
5. **If below 80%** → identify uncovered lines → write tests → re-run → repeat.

### Rules

- Target: 80%. Floor: 75%. Below 75% the task is **not** complete.
- Run for BOTH backend AND frontend.
- Always show the full output. Never say "coverage looks good" without showing it.
- Actually execute the commands — do not just list them.
- Focus on **branch** coverage (both sides of `if/else`, all `catch` blocks, all error paths).
- Never inflate coverage with meaningless assertions (see "Liar" anti-pattern).

---

## Step 8 — Final checklist

| Check                                                                                  | Status |
|----------------------------------------------------------------------------------------|--------|
| Each new test passes when run **individually**                                         | [ ]    |
| All new tests pass in **random order**                                                 | [ ]    |
| All new tests pass when run **together**                                                | [ ]    |
| Project linter ran on test files — **zero errors**                                     | [ ]    |
| Project formatter ran on test files — **zero diffs**                                   | [ ]    |
| Type checker ran on test files — **zero errors** (if applicable)                       | [ ]    |
| No `@skip` / `xit` / `@Disabled` without a ticket reference                            | [ ]    |
| All created/modified tests pass — **zero failures**                                    | [ ]    |
| Backend coverage report ran — **output shown to the user**                             | [ ]    |
| Frontend coverage report ran — **output shown to the user** (if applicable)            | [ ]    |
| Coverage ≥ 75% on ALL tested source code (backend AND frontend)                        | [ ]    |
| Attempted to reach 80% before accepting 75%                                            | [ ]    |
| Tests pass on **every supported OS** in the CI matrix (see ensuring-cross-platform)    | [ ]    |

If any box fails → fix before declaring done.
