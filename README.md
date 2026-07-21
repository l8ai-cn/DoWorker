<p align="center">
  <img src="docs/images/logo.svg" alt="Agent Cloud" height="72" />
</p>

<h1 align="center">Agent Cloud</h1>

<h3 align="center">Where teams scale beyond headcount.</h3>

<p align="center">
  The AI Agent Workforce Platform.<br/>
  Run a hundred AI agents across your own machines — and command them all from one console.
</p>

<p align="center">
  <a href="https://agentcloud.ai">Website</a> ·
  <a href="https://agentcloud.ai/docs">Docs</a> ·
  <a href="#quick-start">Quick Start</a> ·
  <a href="https://github.com/l8ai-cn/AgentCloud">GitHub</a> ·
  <a href="https://discord.gg/3RcX7VBbH9">Discord</a> ·
  <a href="https://x.com/agentcloudai">X</a> · <a href="https://x.com/stone0dong">X (founder)</a> ·
  <a href="https://www.linkedin.com/company/agentcloud">LinkedIn</a>
</p>

<p align="center">
  <a href="https://github.com/l8ai-cn/AgentCloud/actions/workflows/ci.yml"><img src="https://github.com/l8ai-cn/AgentCloud/actions/workflows/ci.yml/badge.svg?branch=main" alt="CI" /></a>
  <a href="https://github.com/l8ai-cn/AgentCloud/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-BSL--1.1-blue" alt="License" /></a>
  <a href="https://hub.docker.com/u/agentcloud"><img src="https://img.shields.io/badge/docker-hub-blue?logo=docker" alt="Docker Hub" /></a>
</p>

<p align="center">
  <a href="https://youtu.be/FZrUO0tim0U">
    <img src="https://img.youtube.com/vi/FZrUO0tim0U/maxresdefault.jpg" alt="Agent Cloud Demo Video" width="720" />
  </a>
</p>

---

## The problem: one operator, a hundred agents

AI coding agents have made individual engineers wildly productive — but individual productivity has a ceiling. The next 10x isn't a smarter agent; it's **running many agents at once**, and directing them like a team.

That ambition breaks the moment you try it for real:

- A hundred agents won't fit on one laptop.
- Nobody can babysit a hundred terminals.
- Each agent needs its own clean, isolated workspace — or they corrupt each other's state.
- Long-running agents stall, get stuck, and silently die.
- Agents working in isolation never compound into a team.

What's missing isn't the agent. It's the **control layer** that turns one operator into the director of an agent workforce — the layer that schedules agents onto machines, isolates them, keeps them alive, lets them collaborate, and puts all of it on one screen.

**Agent Cloud is that layer.**

## From problem to platform

Every part of Agent Cloud exists to answer one question: *how does a single person reliably run, watch, and steer a hundred agents?* Each capability is the direct answer to a wall you hit when you scale.

| The wall you hit | What Agent Cloud gives you |
|---|---|
| 100 agents won't run on one machine | **Runner fleet** — install self-hosted runners across any number of machines. Each advertises its capacity (`max_concurrent_pods`), and agents are scheduled onto the runner you pick or an available one from the pool. Your code never leaves your infrastructure. |
| Every agent needs a clean, isolated environment | **Workspace isolation** — each agent runs in its own pod with a dedicated Git worktree sandbox (`sandboxes/{pod}/workspace/`), private credentials, and its own branch. Concurrent agents never step on each other. |
| You can't watch a hundred terminals | **One web console, every screen** — paginated pod sidebar, multi-pane workspace, and real-time terminal streaming let one person hold many agents in view. |
| Long-running agents stall and need babysitting | **Autopilot** — a control agent watches a pod and sends the next instruction the moment it goes idle, with iteration caps, decision history, and human takeover/handback. Self-healing, unattended runs. |
| Agents working alone don't compound | **Channels & Tickets** — organize Workers around shared work, communicate with `@mentions`, and track execution against a ticket. |

The rest is plumbing built so that chain holds up under load: a **control-plane / data-plane split** — orchestration over gRPC with mTLS, terminal bytes over a stateless Relay cluster — so the backend never bottlenecks on PTY traffic, no matter how many agents are streaming at once.

## Core concepts

- **Worker** — one isolated execution environment: a PTY or ACP session, a Git worktree sandbox, typed configuration, and a real-time output stream. `Pod` is the internal lifecycle object and legacy API name.
- **Runner** — a self-hosted daemon you install on your own machines. It connects to the backend over gRPC+mTLS and starts Workers. Register as many as you need; Workers schedule across the fleet.
- **Workspace** — the per-Worker sandbox: an isolated Git worktree plus scoped credentials, so concurrent runs do not collide.
- **Workflow** — an explicit task and acceptance loop that references an immutable Worker configuration.
- **Channel** — a shared collaboration space where Workers and people communicate with `@mentions`.
- **Ticket** — a unit of work on a Kanban board, bindable to a Worker with progress and MR/PR tracking.

## Architecture

Agent Cloud separates the **control plane** from the **data plane**: orchestration commands travel over gRPC with mTLS, while terminal I/O streams through a stateless Relay cluster. The backend never touches a single PTY byte — which is what lets the fleet scale.

<p align="center">
  <img src="docs/images/architecture.svg" alt="Agent Cloud Architecture" width="680" />
</p>

**Server-side (Go)**

| Component | Role |
|-----------|------|
| **Backend** | API server (Gin + GORM) — auth, org/team/user, pod lifecycle, tickets, billing, and the PKI that issues runner certs |
| **Relay** | WebSocket relay for the terminal data plane — low-latency pub/sub between runners and clients |
| **Runner** | Self-hosted daemon — connects to the backend (gRPC+mTLS), spawns isolated PTY pods that run the actual agents |

**Client-side**

| Component | Role |
|-----------|------|
| **Rust Core** | Business-logic SSOT — shared crates compiled to WASM for web. One cache, one set of services. |
| **Web** | Next.js console — terminal, Kanban, real-time mesh topology |
| **Web-Admin** | Internal admin console — user/org/runner management, audit logs |

## Getting Started

The fastest way to use Agent Cloud is the hosted service at **[agentcloud.ai](https://agentcloud.ai)** — sign up, connect your Git provider, and start running agents in minutes. Bring your own AI API keys (**BYOK**): no usage caps, full cost control.

### 1. Install a Runner

The Runner is a lightweight daemon that runs on your machine and executes AI agents locally. Your code stays on your infrastructure. Install one per machine you want in the fleet.

```bash
curl -fsSL https://agentcloud.ai/install.sh | sh
```

> See the [Runner README](runner/) for more installation options (deb, rpm, Windows, etc.)

### 2. Login

```bash
agent-cloud-runner register
```

This opens your browser to authenticate. For headless environments (SSH, remote server):

```bash
agent-cloud-runner register --headless
```

For self-hosted deployments, add `--server`:

```bash
agent-cloud-runner register --server https://your-server.com
```

### 3. Run

```bash
agent-cloud-runner run
```

Or install as a system service for always-on operation:

```bash
agent-cloud-runner service install
agent-cloud-runner service start
```

Once the runner is online, create a **Worker** from the web console. The wizard preflights the selected Worker type, model resource, credentials, workspace, and Runner before creation.

## Quick Start

Run the whole stack locally with one command.

```bash
git clone https://github.com/l8ai-cn/AgentCloud.git
cd AgentCloud
./deploy/dev/dev.sh
```

This starts the full stack: PostgreSQL, Redis, MinIO, Backend, Relay, Traefik, and the Next.js frontend with hot reload.

**Access (main worktree / offset 0):**

| Service | URL |
|---------|-----|
| Web Console | http://localhost:10007 |
| API | http://localhost:10000/api |
| Admin Console | http://localhost:10011 |

**Test Accounts:**

| Role | Email | Password |
|------|-------|----------|
| User | dev@agentcloud.local | AdminAb123456 |
| Admin | admin@agentcloud.local | Ab123456 |

> Ports are dynamically allocated per worktree. Check `deploy/dev/.env` for actual values.

<details>
<summary><strong>Manual Setup</strong></summary>

**Prerequisites:** Go, Node.js, pnpm, Docker, and the agent CLIs required by the Workers you plan to run.

```bash
# 1. Start infrastructure + host services
./deploy/dev/dev.sh

# 2. Tail logs
tail -f deploy/dev/runtime/backend/backend.log
tail -f deploy/dev/web.log

# Low-memory alternative:
# cd deploy/dev && ./dev-lite.sh
```

</details>

<details>
<summary><strong>Production Deployment</strong></summary>

Docker images are published to Docker Hub on every push to `main`:

```
agentcloud/backend:sha-xxxxxxx
agentcloud/web:sha-xxxxxxx
agentcloud/web-admin:sha-xxxxxxx
agentcloud/relay:sha-xxxxxxx
```

Tagged releases (`v*`) get semver tags:

```
agentcloud/backend:1.0.0
agentcloud/backend:1.0
```

See [deploy/selfhost/](deploy/selfhost/) for the self-hosted deployment guide.

</details>

## Worker Runtime Status

Agent Cloud does not treat an arbitrary terminal command as an integrated Worker.
Each formal type needs a Definition, explicit adapter, a published immutable
runtime image that can be pulled by digest, credential or model-resource
mapping, and product-path evidence before it is marked supported.

| Worker type | Current status | Creation requirement |
|-------------|----------------|----------------------|
| Codex CLI | Not formally deployable | Local create, ACP prompt, and cleanup passed; configured release digest is not pullable |
| Claude Code | Not formally deployable | Configured release digest is not pullable; compatible resource and full product path remain unverified |
| Gemini CLI | Not formally deployable | Configured release digest is not pullable; compatible Google resource is absent |
| Aider | Not formally deployable | No published immutable digest; upstream image build is blocked |
| Cursor CLI | Not formally deployable | No published immutable digest |
| Do Agent | Not formally deployable | No published immutable digest |
| Grok Build | Not formally deployable | No published immutable digest |
| Hermes | Not formally deployable | No published immutable digest; upstream image build is blocked |
| Loopal | Not formally deployable | No published immutable digest or accepted real artifact |
| MiniMax CLI | Not formally deployable | No published immutable digest |
| OpenClaw | Not formally deployable | No published immutable digest |
| OpenCode | Not formally deployable | No published immutable digest |

No Worker type is formally deployable until all release gates pass. The
machine-readable sources are `config/worker-types/`,
`backend/internal/domain/workerruntime/runtime_catalog.lock.json`,
`tools/loops/worker-onboarding/catalog-loop/evidence/runtime-lock-probes.json`,
and `clients/web/src/generated/worker-runtime-catalog.json`.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go (Gin + GORM) |
| Client Core | Rust → WASM (web) — shared business logic |
| Web | Next.js (App Router) + TypeScript + Tailwind |
| Database | PostgreSQL + Redis |
| Storage | MinIO (S3-compatible) |
| API | REST + gRPC (bidirectional streaming) |
| Security | mTLS for runner connections, JWT for web auth |
| Real-time | gRPC streaming (Runner ↔ Backend), WebSocket (Relay ↔ Client) |
| Reverse Proxy | Traefik |

## Project Structure

```
AgentCloud/
├── backend/          # Go API server
├── relay/            # Terminal relay server (Go)
├── runner/           # Self-hosted runner daemon (Go)
├── agentfile/        # AgentFile DSL
├── clients/
│   ├── core/         # Rust business-logic SSOT (WASM)
│   ├── web/          # Next.js console
│   ├── web-admin/    # Admin console (Next.js)
│   └── web-user/     # Hive / session UI
├── proto/            # Protocol Buffers definitions
├── packages/         # Shared TS packages (service-runtime, …)
├── deploy/
│   ├── dev/          # Docker Compose + host-side development services
│   └── selfhost/     # Self-hosted deployment guide
├── tests/            # E2E / hive smoke suites
└── docs/             # Architecture docs and RFCs
```

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Security Policy](SECURITY.md)

## License

[Business Source License 1.1](LICENSE) (BSL-1.1)

- **Change Date:** 2030-02-28
- **Change License:** GPL-2.0-or-later

The BSL allows you to use, copy, and modify the software for non-production purposes. Production use requires a commercial license until the change date, after which the software becomes available under GPL-2.0-or-later. See [LICENSE](LICENSE) for the full terms and additional use grant.
