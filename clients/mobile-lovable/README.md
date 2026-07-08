# Agent Console (mobile-lovable)

Mobile web UI for AgentsMesh — ACP chat + CLI terminal attach.

## Prerequisites

- AgentsMesh dev stack: `bazel run //deploy/dev:up` (API on `http://localhost:10000`)
- Node.js 20+

## Dev server

```bash
cd clients/mobile-lovable
npm install
npm run dev -- --port 10021 --host
```

Open http://localhost:10021/login

Vite proxies `/auth`, `/v1`, `/api` → `http://localhost:10000`.

## Test account

- Email: `dev@agentsmesh.local`
- Password: `AdminAb123456`

## Modes

| Mode | API | UI |
|------|-----|-----|
| ACP Chat | `POST /v1/sessions`, SSE `/stream` | Session detail |
| CLI Terminal | WS `.../terminals/:id/attach` | `/sessions/:id/terminal` |
| Org experts | `GET/POST /api/v1/orgs/:org/experts` | Experts + `/new` |

## Env overrides (optional)

```bash
VITE_AGENTSMESH_API_URL=http://localhost:10000
VITE_AGENTSMESH_JWT=...
VITE_AGENTSMESH_ORG_SLUG=dev-org
```

## Build

```bash
npm run build
```
