# Do Worker Mobile

Mobile Worker entry point for ACP visual conversations and PTY command lines.

## Prerequisites

- Node.js 20+ and the root pnpm workspace dependencies.
- Backend and Relay running through `./deploy/dev/dev.sh` or the equivalent host services.

## Dev server

```bash
pnpm run mobile:dev
```

The command reads `deploy/dev/.env` when present, uses its
`MOBILE_LOVABLE_PORT` (default `10021`), and proxies the API to
`BACKEND_HTTP_PORT` (default `10015`).

Open `http://127.0.0.1:10021/login`.

To inspect the UI from a device on the same LAN, bind Vite explicitly:

```bash
MOBILE_DEV_HOST=0.0.0.0 pnpm run mobile:dev
```

The local Relay URL must also be reachable by that device. For end-to-end
mobile testing, use the HTTPS deployment rather than a phone-local `localhost`
Relay URL.

## Test account

- Email: `dev@agentsmesh.local`
- Password: `AdminAb123456`

## Modes

| Mode | API | UI |
|------|-----|-----|
| ACP Chat | `POST /v1/sessions`, SSE `/stream` | Session detail |
| CLI Terminal | `GET /v1/sessions/:id/relay-connection`, Relay WebSocket + control lease | `/sessions/:id/terminal` |
| Org experts | `GET/POST /api/v1/orgs/:org/experts` | Experts + `/new` |

## Env overrides (optional)

```bash
DO_WORKER_API_URL=http://127.0.0.1:10015
VITE_AGENTSMESH_JWT=...
VITE_AGENTSMESH_ORG_SLUG=dev-org
```

## Build

```bash
pnpm run mobile:build
```
