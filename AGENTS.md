# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

## Project Overview

Do Worker is **The AI Agent Workforce Platform** — where teams scale beyond headcount. It supports Codex, Codex CLI, Gemini CLI, Aider, and more.

### Components

Server-side (Go):
- **Backend** (`backend/`): API server (Gin + GORM). REST for clients, gRPC + mTLS to Runner. Owns auth, org/team/user, pod lifecycle, ticket/channel, billing, PKI for runner certs. PostgreSQL + Redis.
- **Relay** (`relay/`): WebSocket relay for the terminal data plane. Browser ↔ Relay ↔ Runner (binary protocol). Backend never touches PTY bytes.
- **Runner** (`runner/`): Self-hosted daemon. Connects to Backend via gRPC bidi stream. Spawns isolated PTY pods that run the actual AI agents (Codex / Codex / Aider / …).

Client-side:
- **Rust Core** (`clients/core/`): **Business-logic SSOT.** Rust crates compiled to WASM for Web. Owns the authoritative cache, DTOs, and services — auth, blockstore, channels, tickets, mesh, autopilot. Front-ends are thin views over it. **Modify state in Rust, not in Zustand.**
- **Web** (`clients/web/`): Next.js (App Router + TS + Tailwind). Loads `agentsmesh-wasm` at boot; UI state mirrors Rust selectors via `_tick` triggers.
- **Web-Admin** (`clients/web-admin/`): Next.js admin console mounted at `/admin`. Internal-only — gated on `is_system_admin`.

## Development Environment

Go services (backend / runner / relay) run on the host via `air` hot-reload.
Next.js apps (web / web-admin) run via plain `next dev`. Docker only hosts
stateful infrastructure (PostgreSQL, Redis, MinIO, Traefik, Jaeger, Gitea,
OTel collector, Adminer). Wasm is built with `pnpm run build:wasm` when needed.

### Quick Start

```bash
./deploy/dev/dev.sh                  # docker infra + host backend/relay/runner + host web/web-admin
./deploy/dev/dev.sh --clean          # stop everything, drop docker volumes, clear runtime/
./deploy/dev/dev.sh --reset-runners  # only restart host runner+relay (backend stays up)
./deploy/dev/dev.sh --rebuild-runner # rebuild runner binary + restart containers
./deploy/dev/dev.sh --backend-only   # CI-style: skip frontends
```

**Low-memory / web-only frontend** (optional):

```bash
cd deploy/dev && ./dev-lite.sh       # air backend/relay + coordinator runners + web only
pnpm proto:gen-go-all                # first-time: proto + amesh codegen
```

Prerequisites (one-time):

```bash
# Go, Docker, pnpm required; air auto-installs on first start if missing
# protoc needed when regenerating proto stubs: brew install protobuf
npm i -g @anthropic-ai/Codex @openai/codex @google/gemini-cli  # for runner pods
```

The dev pipeline automatically:
1. Generates `.env` with worktree-isolated ports (3 host service ports added: BACKEND_HTTP_PORT / BACKEND_GRPC_PORT / RELAY_HTTP_PORT)
2. Generates traefik dynamic configs that route `host.docker.internal:<host-port>`
3. Starts the docker infra stack
4. Runs migrations via the `migrate/migrate` oneshot service (no backend container needed)
5. Launches `air` for backend / relay / runner in the background, with isolated `$HOME` for the runner so its `~/.Codex/*` writes don't touch your real configs
6. Runs `pnpm run build:wasm` if needed, then starts plain `next dev` for web and web-admin

### Services & Ports (offset 0 / main worktree)

| Service | URL | Notes |
|---------|-----|-------|
| **Frontend** | http://localhost:10007 | `next dev` (host) |
| **Admin Console** | http://localhost:10011 | `next dev` (host) |
| **web-user** | http://localhost:10020 | Vite dev server (host) |
| **API** | http://localhost:10000/api | traefik → host backend :10015 |
| **Relay** | ws://localhost:10000/relay | traefik → host relay :10017 |
| **gRPC mTLS** | grpcs://localhost:10001 | traefik passthrough → host backend :10016 |
| Postgres | localhost:10002 | docker |
| Redis | localhost:10003 | docker |
| MinIO API/Console | localhost:10004 / 10005 | docker |
| Adminer | localhost:10006 | docker |
| Traefik Dashboard | localhost:10008 | docker |
| Gitea HTTP/SSH | localhost:10009 / 10010 | docker |
| OTel gRPC/HTTP | localhost:10012 / 10013 | docker |
| Jaeger UI | localhost:10014 | docker |

Each worktree adds offset×50 to every slot.

Test accounts:
- **User**: dev@agentsmesh.local / AdminAb123456
- **Admin**: admin@agentsmesh.local / Ab123456

### Logs

```bash
tail -f deploy/dev/runtime/backend/backend.log   # air + backend stdout
tail -f deploy/dev/runtime/relay/relay.log
tail -f deploy/dev/runtime/runner/runner.log
tail -f deploy/dev/web.log                       # next dev (web)
tail -f deploy/dev/web-user.log                  # Vite (web-user)
docker compose logs -f postgres                  # docker infra
```

### Hot Reload

- **Frontend (web / web-admin)**: Next.js dev server fast refresh
- **Go services (backend / runner / relay)**: `air` watches `.go` changes and rebuilds incrementally

## Build Commands (for CI/testing outside Docker)

CI is defined in `.github/workflows/ci.yml`.

### Backend / Runner / Relay (Go)

```bash
go test ./backend/... ./runner/... ./relay/...
go test ./backend/internal/service/... -run TestAuth   # specific test

# Lint — run golangci-lint in each module directory
(cd backend && golangci-lint run)
(cd runner && golangci-lint run)
(cd relay && golangci-lint run)
```

### Images (Docker)

Dockerfiles live next to each service; build from the repo root:

```bash
docker build -f backend/Dockerfile .
docker build -f relay/Dockerfile .
docker build -f runner/Dockerfile .
docker build -f clients/web/Dockerfile .
docker build -f clients/web-admin/Dockerfile .
```

### Web (Next.js)

所有前端的依赖（web / web-admin）统一放在根 `package.json`：

```bash
pnpm install                 # Install at repo root (one-shot)
pnpm run build:wasm          # Build Rust → WASM package
pnpm run web:lint            # ESLint
pnpm run web:typecheck       # tsc --noEmit
pnpm run web:test            # Vitest
pnpm run web:build           # Production Next.js build
pnpm run web-admin:lint
pnpm run web-admin:typecheck
pnpm run web-admin:build

# Dev server (also started by ./deploy/dev/dev.sh)
(cd clients/web && node ../../node_modules/next/dist/bin/next dev --turbopack)
```

#### Wasm 加载边界（路由分层）

为避免 21MB wasm 在静态/营销页 block 渲染（手机基本跑不动），WasmProvider
**仅挂在三组 layout 中**，营销页保持 0 wasm：

| Layout | wasm | 路由 |
|---|---|---|
| `app/layout.tsx` (root) | ❌ | 全站基底，无 wasm |
| `app/(dashboard)/layout.tsx` | ✅ | `(dashboard)/[org]/**`、`/settings`、`/support` |
| `app/(auth)/layout.tsx` | ✅ | `/login`、`/register`、OAuth callback、verify-email、invite、onboarding、runners |
| `app/popout/layout.tsx` | ✅ | `/popout/terminal/[podKey]` |
| 其它营销/文档 (`/`、`/docs`、`/about`、`/blog`、`/changelog`、`/demo`、`/enterprise`、`/privacy`、`/terms`、`/mock-checkout` 等) | ❌ | 通过 `lib/light-session.ts` 直读 localStorage 判 auth；公开 API 走 `lib/public-api.ts` 的 fetch |

**约束**（违反会让营销页重新加载 wasm）：
- 营销页组件**不要 import** `@/lib/wasm-core` / `@agentsmesh/service-runtime` / `agentsmesh-wasm` / `@/stores/auth`（任意一个会通过依赖图把 21MB 拉进 chunk）
- 需要"已登录吗 + 当前 org slug"用 `useLightSession`（来自 `@/hooks/useLightSession`）
- 需要 CTA 用 `LightAuthButtons`（不是 `AuthButtons`）
- 需要公开 API（pricing 等）用 `fetch` 或 `lib/public-api.ts` 包装，不走 wasm

**校验**：CI / 本地构建后跑 `bash clients/web/scripts/check-no-wasm-in-marketing.sh` 验证营销 chunk 不含 wasm 符号。

### Runner release

```bash
bash scripts/build-runner-release.sh   # 6-platform tar.gz/zip + checksums
```

Cross-compiles linux/darwin/windows × amd64/arm64, packages tar.gz/zip plus
`checksums.txt`. The release.yml workflow stamps version into the staged
filenames and runs `rcodesign` over darwin binaries before `gh release create`.

### Proto

```bash
pnpm proto:gen-ts        # regenerate committed TypeScript protobuf files
pnpm proto:gen-go        # regenerate proto/gen/go (requires protoc)
pnpm proto:gen-go-all    # proto + amesh convert sync
cd clients/core && cargo run -p do_worker_proto_gen --bin gen-proto
```

Rust protobuf sources under `clients/core/crates/proto/*/src/` are gitignored and
are not regenerated by `cargo test`; run `gen-proto` after every `.proto`
change. Requires `protoc` plus `protoc-gen-go` / `protoc-gen-go-grpc`
(auto-installed by the script if missing). No Bazel fallback.

### Rust Core (Cargo workspace)

Rust 业务代码在 `clients/core/`（Cargo workspace）。依赖在各 crate 的
`Cargo.toml` 中声明；WASM 产物通过根目录 `pnpm run build:wasm` 产出。

```bash
cd clients/core && cargo test --workspace
pnpm run build:wasm      # from repo root — builds wasm package for web
```

**加新依赖**：编辑对应 crate 的 `Cargo.toml`，然后 `cargo test` / `pnpm run build:wasm` 验证。

### Database Migrations

Migrations are located in `backend/migrations/` using golang-migrate format.

**Development** (via Docker):
```bash
./deploy/dev/dev.sh      # automatically runs all migrations
```

**Production** (via backend container):
```bash
# Inside the backend container, golang-migrate is pre-installed
migrate -path /app/migrations -database "postgres://user:pass@host:5432/db?sslmode=disable" up
migrate -path /app/migrations -database "postgres://user:pass@host:5432/db?sslmode=disable" down 1
migrate -path /app/migrations -database "postgres://user:pass@host:5432/db?sslmode=disable" version
```

**Create new migration**:
```bash
# Install golang-migrate locally
brew install golang-migrate

# Create migration files
migrate create -ext sql -dir backend/migrations -seq add_new_feature
# This creates: 000024_add_new_feature.up.sql and 000024_add_new_feature.down.sql
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Web (Next.js)                            │
│                 localhost:3000                              │
└─────────────────────────────────────────────────────────────┘
                              │
                        REST / WebSocket
                         (terminal/events)
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Backend (Go + Gin)                        │
│            REST: localhost:8080 | gRPC: localhost:9443      │
│  - Auth (JWT + OAuth)                                       │
│  - Organization/Team/User management                        │
│  - Pod lifecycle management                                 │
│  - Ticket/Channel management                                │
│  - PostgreSQL + Redis                                       │
│  - PKI: Runner certificate issuance & revocation            │
└─────────────────────────────────────────────────────────────┘
                              │
                      gRPC + mTLS (port 9443)
                   (bidirectional streaming)
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Runner (Go daemon)                        │
│              Self-hosted by users                           │
│  - Connects via gRPC with mTLS client certificate           │
│  - Creates isolated PTY terminals (Pods)                    │
│  - Executes AI agents (Codex, Aider, etc.)            │
│  - Streams terminal output back to server                   │
│  - Auto certificate renewal before expiry                   │
└─────────────────────────────────────────────────────────────┘
```

## Backend Structure

```
backend/
├── cmd/server/           # Entry point
├── internal/
│   ├── api/rest/         # REST API handlers
│   │   └── v1/admin/     # Admin API handlers
│   ├── domain/           # Domain models (DDD style)
│   │   ├── user/         # User entity (includes is_system_admin)
│   │   ├── organization/ # Organization entity
│   │   ├── agentpod/     # AgentPod entity
│   │   ├── agent/        # Agent configuration entity
│   │   ├── ticket/       # Ticket entity
│   │   ├── channel/      # Channel entity
│   │   ├── runner/       # Runner entity
│   │   ├── billing/      # Billing/subscription entity
│   │   ├── invitation/   # Organization invitation
│   │   ├── promocode/    # Promo code entity
│   │   ├── gitprovider/  # Git provider (OAuth) entity
│   │   ├── repository/   # Repository entity
│   │   ├── mesh/         # Mesh topology entity
│   │   ├── file/         # File storage entity
│   │   └── admin/        # Admin audit log entity
│   ├── service/          # Business logic layer
│   │   └── admin/        # Admin service (dashboard, user/org management)
│   ├── infra/            # Infrastructure (DB, cache)
│   ├── config/           # Configuration loading (includes AdminConfig)
│   └── middleware/       # Auth, tenant isolation, AdminMiddleware
├── pkg/                  # Shared packages
│   ├── auth/             # JWT and OAuth utilities
│   ├── crypto/           # Encryption utilities
│   ├── i18n/             # Internationalization
│   └── audit/            # Audit logging
└── migrations/           # SQL migrations
```

## Web Structure

```
clients/web/src/
├── app/                  # Next.js App Router
│   ├── (auth)/           # Auth pages (login, register)
│   ├── (dashboard)/      # Dashboard pages
│   └── api/              # API routes
├── components/           # React components
├── lib/                  # Utilities, API clients
├── stores/               # Zustand state stores
├── hooks/                # Custom React hooks
├── messages/             # i18n translations (next-intl)
└── providers/            # Context providers
```

## Web-Admin Structure (Admin Console)

```
clients/web-admin/src/
├── app/                  # Next.js App Router (basePath: /admin)
│   ├── login/            # GitLab SSO login page
│   ├── auth/callback/    # OAuth callback handler
│   └── (dashboard)/      # Dashboard pages (protected)
│       ├── users/        # User management
│       ├── organizations/ # Organization management
│       ├── runners/      # Runner management
│       └── audit-logs/   # Audit log viewer
├── components/
│   ├── ui/               # Shadcn-style UI components
│   └── layout/           # Sidebar, Header
├── lib/
│   ├── api/              # Admin API client
│   └── utils.ts          # Utility functions
└── stores/
    └── auth.ts           # Zustand auth store (persist to localStorage)
```

## Runner Structure

```
runner/
├── cmd/runner/           # Entry point (register/run/service)
├── internal/
│   ├── runner/           # Core runner logic
│   │   ├── runner.go         # Main Runner struct
│   │   ├── pod_builder.go    # Builder pattern for Pods
│   │   ├── pod_store.go      # Pod storage
│   │   ├── message_handler.go # gRPC message routing
│   │   └── pty_forwarder.go  # Terminal output forwarding
│   ├── client/           # gRPC client (mTLS)
│   │   ├── grpc_connection.go   # gRPC bidirectional stream
│   │   ├── grpc_registration.go # Certificate registration
│   │   └── protocol.go          # Message types
│   ├── terminal/         # PTY management (creack/pty)
│   ├── process/          # Process management
│   ├── sandbox/          # Sandbox environment
│   │   └── plugins/      # worktree, tempdir plugins
│   ├── mcp/              # Model Context Protocol integration
│   ├── workspace/        # Git worktree management
│   └── console/          # Console UI
```

## Key Concepts

**Pod**: An isolated execution environment with PTY terminal, sandbox config, and output forwarder.

**Runner**: Self-hosted daemon that connects to backend via gRPC+mTLS, receives tasks, and manages Pod lifecycle.

**Sandbox**: Configurable environment created by plugins (worktree for Git isolation, tempdir for temporary workspace).

**Channel**: Multi-agent collaboration space where agents can communicate.

**Ticket**: Task management unit with kanban board integration.

## Message Flow (Runner ↔ Backend)

1. Runner registers via gRPC, receives mTLS certificate from PKI
2. Runner connects via gRPC bidirectional stream with mTLS
3. Backend sends `create_pod` → Runner creates Sandbox → Starts PTY/ACP process
4. Backend sends `subscribe_pod` → Runner connects to Relay WebSocket
5. Terminal I/O (data plane): Browser ↔ Relay ↔ Runner (WebSocket binary protocol)
6. Control commands (control plane): Backend → Runner via gRPC (`terminate_pod`, `send_prompt`, etc.)
7. Runner events → Backend via gRPC (`pod_created`, `pod_terminated`, `agent_status`, etc.)
8. Certificate auto-renewal before expiry (checked every hour)

## Configuration

**Development** (Docker): Run `./deploy/dev/dev.sh` — auto-generates all configs

**Runner**: `~/.agentsmesh/config.yaml` (created after `runner register`)

## Testing Patterns

- Backend: Standard Go testing with `testify`
- Web: Vitest + Testing Library
- Runner: Go testing, files ending with `_integration_test.go` for integration tests

## Admin Console

The Admin Console (`web-admin`) is an internal management interface for system administrators.

### Access Control

- **Authentication**: Email + Password login (same as main app)
- **Authorization**: `is_system_admin` flag on user record must be `true`
- **Audit Logging**: All admin actions are logged to `system_admin_audit_logs` table

### Features

- **Dashboard**: System statistics (users, organizations, runners, pods)
- **User Management**: View, disable/enable users, grant/revoke admin privileges
- **Organization Management**: View, update, delete organizations
- **Runner Management**: View, disable/enable, delete runners
- **Audit Logs**: View all admin actions with filtering

### Configuration

Admin Console is enabled by default. All components use unified domain configuration:

```bash
# All components use the same two variables (Backend, Relay, Web, Web-Admin)
PRIMARY_DOMAIN=localhost:10000                  # Primary domain for all URLs
USE_HTTPS=false                                 # Use HTTPS/WSS protocols

# Backend-specific
ADMIN_ENABLED=true                              # Enable admin console (default: true)
```

### Creating Admin Users

To grant admin privileges to a user, set `is_system_admin = true` in the database:

```sql
UPDATE users SET is_system_admin = true WHERE email = 'admin@example.com';
```

Or use an existing admin to grant privileges via the Admin Console UI.


## Principles
* Architecture must conform to SOLID, GRASP, and YAGNI.
* **代码即 SSOT — 不要解释 what，只解释 why。** 注释能删则删。可以写注释的场景：业务约束、跨模块契约、解决方案的非显然取舍 (workaround 原因)。绝不能写：复述函数名/类型名的注释、`// 创建 X` 之上紧跟 `CreateX()`、section banner、文档化签名的 JSDoc。代码不够自解释就改代码，不要靠注释补救。
* **Hard limit: every file must stay under 200 lines** (excluding test files, which should stay under 400 lines). When a file approaches this limit, proactively split it by SRP — extract types, helpers, hooks, or sub-components into separate files. A 210-line file is acceptable if splitting would break cohesion; a 300+ line file is never acceptable and must be split before committing.
* **Code is the single source of truth — comments that can be eliminated, must be eliminated.** Only comment to explain **why** something non-obvious exists (business constraints, cross-module contracts, workarounds). Never comment **what** code does — if the code isn't self-explanatory, rewrite the code. No JSDoc that restates the function signature, no `// Create X` above `CreateX()`, no section banners.
* **File names must be specific and descriptive.** Never use generic names like `helpers`, `utils`, `common`, `misc`, `shared`. Name files after what they contain — e.g., `mesh-status-info.ts` not `mesh-helpers.ts`, `runner-display-info.ts` not `runner-utils.ts`. 

## Identifier 字段契约

任何 UNIQUE string 列、URL path 段、@mention key、lookup 主键 **都是 identifier**，必须满足 `backend/pkg/slugkit` 的规则：`^[a-z0-9]+(-[a-z0-9]+)*$`，长度 2-100。

### 字段身份分层
- **认证身份** (`users.email`): 用户输入凭证，合法 email 格式即可
- **公开身份 / identifier** (`users.username`, `organizations.slug`, `channels.slug`, `pods.pod_key`...): 系统派生 + 严格 sanitize，全小写 + 数字 + 连字符
- **呈现层** (`users.name`, `channels.name`...): 任意 Unicode，仅用于 UI 显示

**`name` 字段不得当 identifier 用**。如果需要 lookup，加 `slug` 列。

### 写入路径强制约束
- 外部 raw 字符串（OAuth login / SAML attr / email local-part / AI 输出）**禁止**直接赋值到 identifier 字段
- 必须通过对应 service helper：`userService.EnsureUniqueUsername`、`orgService.CreatePersonal`、`channelService.EnsureUniqueSlug` 等
- helper 内部走 `slugkit.GenerateUnique(seed, dbExistsCheck)`，带 collision retry

### 新增 identifier 字段 checklist
1. migration 加 column + `CHECK (col ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(col) BETWEEN 2 AND 100)`
2. domain model 字段类型用 `slugkit.Slug`（新代码） 或 `string`（兼容老代码）
3. 加 `BeforeSave` hook 调 `slugkit.ValidateIdentifier("<table>.<col>", value)`
4. service 包加 `*Registry` helper 封装 `slugkit.GenerateUnique`
5. 单测覆盖含 `.`/`_`/uppercase/unicode 的输入，断言落库值通过 `slugkit.Validate`

参见 `backend/pkg/slugkit/doc.go` 完整说明，`.Codex/plans/sharded-imagining-bird.md` 重构 plan。
