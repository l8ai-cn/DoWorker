# web-user

End-user agent workbench — chat, terminal, files, session collaboration.

UI vendored from [Omnigent](https://github.com/omnigent-ai/omnigent) `web/` (Apache-2.0, see `THIRD_PARTY.md`).
Wired to **Do Worker Backend** via `/v1/*` session API + JWT auth.

## Client layer

All Do Worker integration lives under `src/lib/do-worker/`:

| Module | Role |
|--------|------|
| `auth-session.ts` | JWT storage (`do-worker-auth/*`, shared key shape with `clients/web`) |
| `host-config.ts` | Embed host transport + WebSocket URL resolution |
| `api-client.ts` | Authenticated fetch with org slug headers |
| `server-info.ts` | `GET /v1/info` capabilities probe |
| `session-labels.ts` | Session label key SSOT (`do-worker.ui`, `do-worker.wrapper`, …) |
| `cli-commands.ts` | Runner reconnect command snippets |
| `storage-keys.ts` | localStorage key prefix + legacy `omnigent:` migration |

## Three frontends

| Package | Role | Stack |
|---------|------|-------|
| `clients/web` | Org dashboard, IDE, pods, channels | Next.js + Rust WASM |
| `clients/web-admin` | System admin console | Next.js |
| **`clients/web-user`** | **Direct agent usage for end users** | Vite + React + Do Worker `/v1` API |

## Dev

Started **in parallel** with web / web-admin when you run `bazel run //deploy/dev:up`.
Port is worktree-scoped via `WEB_USER_PORT` in `deploy/dev/.env` (default **10020**).
Vite proxies `/v1`, `/auth`, `/api` → traefik (`HTTP_PORT` via `localhost`, not `127.0.0.1`).

```bash
source deploy/dev/.env
cd clients/web-user
DO_WORKER_API_URL="http://localhost:${HTTP_PORT}" npm run dev -- --port "${WEB_USER_PORT}" --host 127.0.0.1
```

See `docs/rfc/web-user-omnigent-compat.md` for backend compat mapping.
