# web-user

End-user agent workbench — chat, terminal, files, session collaboration.

Vendored from [Omnigent](https://github.com/omnigent-ai/omnigent) `web/` (Apache-2.0, see `THIRD_PARTY.md`).
Will be wired to **AgentsMesh Backend**, not Omnigent Server.

## Three frontends

| Package | Role | Stack |
|---------|------|-------|
| `clients/web` | Org dashboard, IDE, pods, channels | Next.js + Rust WASM |
| `clients/web-admin` | System admin console | Next.js |
| **`clients/web-user`** | **Direct agent usage for end users** | Vite + React |

## Dev

Started with the rest of the dev stack (`bazel run //deploy/dev:up`). Port is
worktree-scoped via `WEB_USER_PORT` in `deploy/dev/.env` (default offset 0:
**10020**). Vite proxies `/v1`, `/auth`, `/api` → traefik (`HTTP_PORT`).

Manual start:

```bash
source deploy/dev/.env
cd clients/web-user
AGENTSMESH_API_URL="http://127.0.0.1:${HTTP_PORT}" npm run dev -- --port "${WEB_USER_PORT}" --host 127.0.0.1
```

## Adaptation target

Proxy `/v1` → AgentsMesh API (`PRIMARY_DOMAIN/api` or dedicated session-compat routes).
See `docs/rfc/web-user-omnigent-compat.md` for the full API gap list.
