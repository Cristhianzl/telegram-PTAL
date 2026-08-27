# File structure — hard limits

> These are **hard limits**, not guidelines. They apply to ALL programming languages and frameworks.

---

## 1. Hard limits per file

| Metric                                                     | Max         | Flexibility                                                                                              |
|------------------------------------------------------------|-------------|----------------------------------------------------------------------------------------------------------|
| Lines of code (excluding imports, types, docs)             | **500**     | Up to 600–700 acceptable ONLY if all other rules (SRP, separation by responsibility, no mixed prefixes) pass |
| Functions with **different** responsibilities              | **5**       | No flexibility                                                                                            |
| Functions with **same** responsibility (same prefix)       | **10**      | No flexibility                                                                                            |
| Main classes per file                                      | **1**       | Small related classes (exceptions, DTOs, enums) up to 5 are OK                                            |

If you are about to exceed any limit → **stop and split into multiple files first.**

> The 600–700 line flexibility is not a free pass. A 650-line file where all functions share the same responsibility prefix is acceptable. A 550-line file with mixed responsibilities is not — it fails regardless of line count.

### Understanding the function limit

**Allowed — 10 functions with same prefix (same responsibility):**
```
# helpers/validation.py
validate_email()
validate_phone()
validate_cpf()
validate_cnpj()
validate_date()
validate_address()
validate_name()
validate_password()
validate_username()
validate_age()
```

**Violation — 6 functions with different prefixes (mixed responsibilities):**
```
# mixed.py
validate_email()       # validation
validate_phone()       # validation
format_currency()      # formatting
format_date()          # formatting
fetch_user()           # data access
save_user()            # data access
```

### Understanding the class limit

**Allowed — multiple small related classes:**
```
# feature_exceptions.py — exceptions only
class FeatureNotFoundError(Exception): ...
class FeatureValidationError(Exception): ...
class FeaturePermissionError(Exception): ...

# feature_types.py — DTOs only
class CreateUserRequest: ...
class UpdateUserRequest: ...
class UserResponse: ...
```

**Violation — multiple unrelated main classes:**
```
# bad.py
class UserService: ...
class OrderService: ...
```

---

## 2. The single responsibility rule

Every file MUST have **one reason to exist** and **one reason to change**.

- ✓ "Validates user input"
- ✓ "Formats data for display"
- ✓ "Handles database persistence"
- ✗ "Validates input and formats output" → two files
- ✗ "Manages users or handles authentication" → two files

---

## 3. Mandatory separation by responsibility

Code must be separated into files based on these responsibility categories:

| Responsibility            | Function-name patterns                              | Must be in separate file       |
|---------------------------|-----------------------------------------------------|---------------------------------|
| Validation                | `validate*`, `check*`, `is_valid*`, `assert*`       | `validation` file               |
| Formatting / Serialization| `format*`, `build*`, `serialize*`, `to_*`           | `formatting` / `serialization`  |
| Parsing                   | `parse*`, `extract*`, `from_*`                      | `parsing` file                  |
| External communication    | `fetch*`, `send*`, `call*`, `request*`              | `client` / `api` file           |
| Data persistence          | `save*`, `load*`, `find*`, `delete*`, `query*`      | `repository` file               |

**Rule:** Functions with **different** prefixes from this table must not coexist in the same file.

---

## 4. STOP conditions — refactor immediately

Stop writing code and split into multiple files when **any** is true:

| Condition                                                  | Required action                                |
|------------------------------------------------------------|------------------------------------------------|
| File exceeds 500 lines                                     | Split by responsibility category               |
| File has 6+ functions with **different** prefixes          | Move related functions to a separate file      |
| File has 11+ functions even with **same** prefix           | Split into logical subgroups                   |
| File has 2+ main classes                                   | Each main class gets its own file              |
| About to add a function with a **new** prefix              | Create the new file first, then add it there   |
| You write a section comment like `# Validation helpers`    | That section becomes its own file              |

---

## 5. Cohesion — avoid over-engineering

Separation is good. **Excessive fragmentation is bad.** Don't create unnecessary files.

Do **not** create a separate file if:

- The file would have only 1–2 functions with <20 lines total.
- The function is used in one place and is not reusable.
- The "helper" is a one-liner that doesn't warrant extraction.

| Scenario                                                       | Extract to a separate file? |
|----------------------------------------------------------------|------------------------------|
| Function used in 2+ files                                      | Yes                          |
| Function with 20+ lines of logic                               | Yes                          |
| Function with its own tests                                    | Yes                          |
| One-liner used once                                            | No                           |
| Private helper (`_func`) used only locally                     | No                           |
| Planned file would have <3 functions and <30 lines             | Probably not                 |

---

## 6. The required pattern

This structure is **mandatory** regardless of language:

```
feature/
├── types/models file       → ModelA, ModelB (data structures only)
├── constants file          → CONSTANT_A, CONSTANT_B
├── helpers/
│   ├── validation file     → validate_a, validate_b
│   └── formatting file     → format_a, format_b
├── services/
│   └── external client     → fetch_from_api
├── repositories/
│   └── data access file    → save_to_database
└── handlers/
    └── request handler     → handle_request
```

---

## 7. Responsibility layer rules

| Layer                       | CAN                                                    | CANNOT                                                       |
|-----------------------------|--------------------------------------------------------|--------------------------------------------------------------|
| **Handler / Controller**    | Receive input, delegate to service, return output      | Contain business logic, call DB directly, call external APIs directly |
| **Service / Orchestrator**  | Coordinate operations, apply business rules            | Know about HTTP/transport, execute SQL/queries directly      |
| **Repository / Data access**| Execute queries, map data                              | Make business decisions, call external APIs                  |
| **Helper**                  | Transform data, validate, format                       | Have side effects, do I/O, maintain state                    |
| **External client**         | Communicate with external services                     | Contain business logic, access the database                  |

---

## 8. Pre-code planning (mandatory)

Before writing any code:

1. **List** all functions/classes you will need.
2. **Categorize** each by responsibility (table in section 3).
3. **Group** functions of the same category.
4. **Create** the file structure first (empty files).
5. **Then** write code into the correct files.

### Pre-code checklist

| Question                                                       | If yes →                                       |
|----------------------------------------------------------------|------------------------------------------------|
| Will I have data structures/models?                            | Create types/models file                       |
| Will I have constants or config values?                        | Create constants file                          |
| Will I have validation functions?                              | Create validation helper file                  |
| Will I have formatting/serialization functions?                | Create formatting helper file                  |
| Will I have external API calls?                                | Create external client file                    |
| Will I have database operations?                               | Create repository/data access file             |
| Will I have 6+ functions total?                                | Plan multiple files by category                |

---

## 9. File naming convention

Use the language's idiomatic naming convention, but follow this semantic pattern:

| Responsibility   | Naming pattern                                                              |
|------------------|------------------------------------------------------------------------------|
| Types / Models   | `{feature}_types`, `{feature}_models`                                        |
| Constants        | `{feature}_constants`, `{feature}_config`                                    |
| Validation       | `validation`, `validators`, `{feature}_validation`                           |
| Formatting       | `formatting`, `serialization`, `{feature}_formatter`                         |
| External calls   | `{service}_client`, `api_client`, `external_{service}`                       |
| Data access      | `{feature}_repository`, `{feature}_store`, `{feature}_dao`                   |
| Orchestration    | `{feature}_service`, `{feature}_manager`                                     |
| Handlers         | `{feature}_handler`, `{feature}_controller`, `{feature}_router`              |

**Never use generic names as standalone files:** `utils`, `helpers`, `misc`, `common`, `shared`.

---

## 10. Final validation checklist

Before delivering any code:

- [ ] No file exceeds 500 lines (600–700 acceptable only if all other rules pass).
- [ ] No file has more than 5 functions with **different** responsibilities.
- [ ] No file has more than 10 functions even with **same** responsibility.
- [ ] No file has more than 1 main class.
- [ ] All functions in a file share the **same** responsibility category.
- [ ] No file mixes prefixes (`validate*` and `format*` in the same file = violation).
- [ ] Types/models are in dedicated type files.
- [ ] Constants are in dedicated constant files.
- [ ] Each file can be described in **one sentence without "and" or "or"**.
- [ ] File names follow semantic naming (no `utils`, `helpers`, `misc`).
- [ ] No over-engineering: no files with only 1-2 trivial functions.

If any item fails → refactor before delivering.
