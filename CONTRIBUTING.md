# Contributing to Do Worker

Thanks for your interest in contributing.

## Before You Start

- Read the [README](./README.md) for project overview and local setup.
- By contributing, you agree that your contributions are licensed under this repository's [BSL 1.1](./LICENSE).

## Development Setup

1. Start local dependencies and seed data (air + plain next + wasm as needed):

```bash
./deploy/dev/dev.sh
```

Optional web-only frontend (lower memory):

```bash
cd deploy/dev && ./dev-lite.sh
```

2. First-time proto codegen (if stubs are missing):

```bash
pnpm proto:gen-go-all
```

## Build and Test

Run relevant checks before opening a PR. CI is `.github/workflows/ci.yml`.

### Backend / Runner / Relay

```bash
go test ./backend/... ./runner/... ./relay/...
(cd backend && golangci-lint run)
(cd runner && golangci-lint run)
(cd relay && golangci-lint run)
```

### Web

```bash
pnpm install --frozen-lockfile   # one-shot at repo root
pnpm run build:wasm
pnpm run web:lint
pnpm run web:typecheck
pnpm run web:test
pnpm run web:build
```

### Rust Core

```bash
cd clients/core && cargo test --workspace
```

### Runner release / images

```bash
bash scripts/build-runner-release.sh
docker build -f backend/Dockerfile .
docker build -f runner/Dockerfile .
```

## Pull Request Guidelines

- Keep PRs focused and small where possible.
- Include context: what changed, why, and any tradeoffs.
- Add or update tests for behavior changes.
- Update docs when user-facing behavior or configuration changes.
- Ensure CI is green before requesting review.

## Commit Messages

Use clear, descriptive commit messages that explain intent.

## Reporting Bugs and Requesting Features

- Use GitHub Issues with the provided templates.
- For security-sensitive reports, see [SECURITY.md](./SECURITY.md).
