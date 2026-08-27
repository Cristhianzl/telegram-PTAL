# Docker as a common denominator

When the application targets Docker for deployment, the image becomes the contract — but only if the build is reproducible across host operating systems.

## Image hygiene

- Use small Linux base images: `alpine`, `slim`, `distroless`. Pin a **digest**, not just a tag.
- Multi-stage builds: build dependencies in one stage, copy only artifacts to the runtime stage.
- Set a non-root `USER` in the runtime stage. Never run production containers as root.
- `HEALTHCHECK` for long-running services so orchestrators know when the container is unhealthy.

## Reproducibility

The same `Dockerfile` MUST produce the same image regardless of host OS. The leaks that break this in practice:

- `.DS_Store` (macOS) and `Thumbs.db` (Windows) — add to `.dockerignore`.
- Line-ending differences in text files — enforce LF in `.gitattributes`.
- File permissions — `RUN chmod` explicitly; do not rely on host mode bits.
- Build cache poisoning from `node_modules`, `.venv`, `target/`, `dist/`, `build/` — exclude in `.dockerignore`.

## `.dockerignore` baseline

```gitignore
.git
.gitignore
.DS_Store
Thumbs.db
*.log
.env
.env.*
.venv/
venv/
__pycache__/
node_modules/
target/
dist/
build/
.pytest_cache/
.mypy_cache/
.idea/
.vscode/
```

## Volume mounts

- Bind mounts behave differently per host OS (case sensitivity, permissions, performance).
- **Dev environments:** document the assumed host OS in the `docker-compose.yml`.
- **Tests:** prefer named volumes or in-image test data over bind mounts. A test that depends on host case-insensitivity is a flake waiting to happen.
- **Production:** named volumes only. No bind mounts in production compose files.

## CI signal for Docker

If the production target includes Docker:

- The image MUST build successfully from a Linux runner.
- The image SHOULD also build from a macOS runner if the team develops on macOS (catches `.DS_Store` and case-insensitivity surprises early).
- A `docker run` smoke test that exercises one critical path MUST be part of the matrix. "It builds" is not "it runs".

## Multi-arch

If the deployment target is `linux/amd64` AND `linux/arm64` (common for AWS Graviton, Apple Silicon dev):

- Use `docker buildx build --platform linux/amd64,linux/arm64 --push`.
- Verify the runtime stage installs only architecture-portable artifacts (no `node-pre-gyp` binary pinned to one arch).
- Smoke-test each arch in CI, not just the build.
