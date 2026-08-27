# Mandatory patterns — concrete code

Language-specific implementations of the principles in `SKILL.md`. Pick the language you are working in. Patterns are equivalent across languages — the API names differ.

---

## Path APIs

**Python**
```python
# Wrong — breaks on Windows, hides intent
path = "data/" + filename
log_file = config_dir + "/logs/app.log"
parts = filepath.split("/")

# Right
from pathlib import Path
path = Path("data") / filename
log_file = config_dir / "logs" / "app.log"
parts = Path(filepath).parts
```

**TypeScript / Node**
```typescript
// Wrong
const p = "data/" + filename;
const logFile = configDir + "/logs/app.log";

// Right
import path from "node:path";
const p = path.join("data", filename);
const logFile = path.join(configDir, "logs", "app.log");
```

**Go**
```go
// Wrong
p := "data/" + filename

// Right
import "path/filepath"
p := filepath.Join("data", filename)
```

**Java**
```java
// Wrong
String p = "data/" + filename;

// Right
import java.nio.file.Path;
Path p = Path.of("data", filename);
```

---

## File I/O with explicit encoding

**Python**
```python
# Wrong — locale-dependent on Windows
with open("config.json") as f:
    data = f.read()
Path("notes.txt").read_text()

# Right
with open("config.json", encoding="utf-8") as f:
    data = f.read()
Path("notes.txt").read_text(encoding="utf-8")
```

**TypeScript / Node**
```typescript
// Wrong — Buffer default toString is not guaranteed UTF-8
const data = fs.readFileSync("config.json");

// Right
const data = fs.readFileSync("config.json", "utf-8");
```

**Go**
```go
// Right — Go reads bytes; decode explicitly when needed
data, err := os.ReadFile("config.json")  // bytes
text := string(data)                      // UTF-8 by convention
```

---

## Subprocess without shell

**Python**
```python
# Wrong — fails on Windows; shell injection risk
os.system("rm -rf temp/")
subprocess.run("ls -la /var/log", shell=True)
subprocess.run(f"grep {pattern} {file}", shell=True)

# Right — stdlib operations, no shell
import shutil, os
shutil.rmtree("temp", ignore_errors=True)
files = os.listdir("/var/log")

# When you must invoke an external program
import sys, subprocess, shutil
subprocess.run([sys.executable, "script.py", "--flag"], check=True)

# For tools on PATH, declare and check
if shutil.which("git") is None:
    raise RuntimeError("git is required but not found on PATH")
```

**TypeScript / Node**
```typescript
// Wrong
import { execSync } from "node:child_process";
execSync(`rm -rf ${dir}`);

// Right
import { rm } from "node:fs/promises";
import { execFileSync } from "node:child_process";
await rm(dir, { recursive: true, force: true });
execFileSync(process.execPath, ["script.js", "--flag"]);
```

**Go**
```go
// Right — exec.Command splits args; no shell
cmd := exec.Command("git", "status")
out, err := cmd.Output()
```

---

## Line endings

**Python**
```python
# Fragile — breaks on CRLF
lines = content.split("\n")

# Robust — handles \n, \r\n, \r
lines = content.splitlines()

# Writing: text mode lets the runtime translate per OS
with open("out.txt", "w", encoding="utf-8") as f:
    f.write("line one\nline two\n")

# When the line ending must be exact (protocols, binary formats), use binary mode
with open("crlf.txt", "wb") as f:
    f.write(b"line\r\n")
```

**TypeScript / Node**
```typescript
// Robust split
const lines = content.split(/\r\n|\r|\n/);

// Writing: use the OS-appropriate line ending
import os from "node:os";
fs.writeFileSync("out.txt", `line one${os.EOL}line two${os.EOL}`, "utf-8");
```

---

## OS-aware temp / cache / config directories

**Python**
```python
# Wrong
cache = "/tmp/myapp"                              # no /tmp on Windows
config = os.path.expanduser("~/.config/myapp")    # XDG-only

# Right
import tempfile
from platformdirs import user_cache_dir, user_config_dir, user_data_dir

with tempfile.TemporaryDirectory() as tmp:
    work_dir = Path(tmp)

cache = Path(user_cache_dir("myapp"))
config = Path(user_config_dir("myapp"))
```

**TypeScript / Node**
```typescript
import os from "node:os";
import path from "node:path";
import envPaths from "env-paths";

const paths = envPaths("myapp");
// paths.cache, paths.config, paths.data, paths.temp

// One-off temp file
import { mkdtempSync } from "node:fs";
const tmp = mkdtempSync(path.join(os.tmpdir(), "myapp-"));
```

---

## Filename validation

```python
WINDOWS_RESERVED = (
    {"CON", "PRN", "AUX", "NUL"}
    | {f"COM{i}" for i in range(1, 10)}
    | {f"LPT{i}" for i in range(1, 10)}
)
FORBIDDEN_CHARS = set('<>:"/\\|?*') | {chr(c) for c in range(32)}

def is_safe_filename(name: str) -> bool:
    if not name or name in {".", ".."}:
        return False
    if name.upper().split(".")[0] in WINDOWS_RESERVED:
        return False
    if any(c in FORBIDDEN_CHARS for c in name):
        return False
    if name.endswith((" ", ".")):  # Windows trims these
        return False
    return True
```

---

## Time: UTC at boundaries

**Python**
```python
# Wrong — local time at storage layer is unportable and ambiguous
created_at = datetime.now()

# Right
from datetime import datetime, timezone
created_at = datetime.now(timezone.utc)
```

**TypeScript / Node**
```typescript
// Wrong
const createdAt = new Date().toLocaleString();

// Right — store as ISO 8601 UTC
const createdAt = new Date().toISOString();  // "2026-05-27T13:42:00.000Z"
```

**Go**
```go
// Right
createdAt := time.Now().UTC()
```
