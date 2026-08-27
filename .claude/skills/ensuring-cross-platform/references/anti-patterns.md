# Anti-pattern reference table

Every row is a POSIX-only assumption that breaks on Windows or fails silently on a non-default locale. The fix column is the portable equivalent.

| Anti-pattern                                      | Why it fails                                          | Portable fix                                            |
|---------------------------------------------------|-------------------------------------------------------|----------------------------------------------------------|
| `"a/" + "b"` for paths                            | Wrong separator on Windows                            | Use a path-join API (`Path / "b"`, `path.join`)          |
| `open(f)` without encoding                        | Locale-dependent on Windows                           | `open(f, encoding="utf-8")`                              |
| `subprocess.run("cmd …", shell=True)`             | Shell varies; injection risk                          | List args, no shell                                       |
| `os.system("rm -rf x")`                           | POSIX-only, no error handling                         | `shutil.rmtree("x")` / `fs.rm`                            |
| `"\n".join(lines)` written in binary mode         | Wrong line endings on Windows                         | Text mode + `f.write` per line                            |
| `text.split("\n")`                                | Misses CRLF                                           | `text.splitlines()` / regex `/\r\n|\r|\n/`                |
| `tempfile = "/tmp/x"`                             | No `/tmp` on Windows                                  | `tempfile.TemporaryDirectory()` / `os.tmpdir()`           |
| `~/.config/app` hardcoded                         | Wrong location on Windows/macOS                       | `platformdirs.user_config_dir` / `env-paths`              |
| Calling `ls`, `grep`, `cat` from code             | Not available on Windows                              | Use stdlib (`os.listdir`, regex)                          |
| `python3` as interpreter literal                  | Windows uses `python` or `py`                         | `sys.executable`                                          |
| Case-sensitive filename equality                  | Windows ignores case                                  | Normalize with `Path.resolve()`                           |
| Using symlinks unconditionally                    | Windows requires admin or Developer Mode              | Check support, fall back to copy                          |
| `os.fork()`                                       | Not available on Windows                              | Use `multiprocessing` (spawn) / process pools             |
| Unix domain sockets                               | Limited on Windows                                    | TCP localhost or named pipes                              |
| `\0` or `:` in filenames                          | Forbidden on Windows                                  | Validate before writing                                   |
| `os.rename` across drives on Windows              | Fails                                                 | `shutil.move`                                             |
| `localhost` for tests                             | May resolve to `::1` on some stacks, breaks bindings  | Pin to `127.0.0.1`                                        |
| Binding low ports in CI                           | Windows reserves dynamic ranges                       | Pick a high, unreserved port; check before binding         |
| `strftime("%-d")` (no-pad)                        | Glibc-only                                            | `strftime("%d").lstrip("0")` or platform check            |
| Locale-dependent number parsing                   | Some locales use `,` for decimal                      | Use `decimal.Decimal` / explicit parser, never `float()`   |
| Locale-dependent sort                             | `sort()` varies                                       | Binary / codepoint sort for stable results                |
| `LoadLibrary("foo.dll")` / `dlopen("libfoo.so")`  | Platform-specific by definition                       | Wrap and isolate behind a factory; lazy import            |
