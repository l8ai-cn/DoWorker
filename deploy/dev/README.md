# Dev Environment

One-click local stack: Postgres, Redis, MinIO, Traefik, Gitea, Backend, Relay, agent-specific Runner containers, plus local Next.js frontends with hot reload.

Go services use `air`; frontends use plain `next dev`. Wasm is built via `pnpm run build:wasm` when the package is missing or stale.

## Quick start

```bash
./dev.sh                         # docker infra + host backend/relay + runners + frontends
./dev.sh --backend-only          # CI-style: skip frontends
./dev.sh --clean                 # stop and wipe volumes
./dev.sh --reset-runners         # restart host runner+relay
./dev.sh --rebuild-runner        # rebuild runner binary + restart runner containers
```

**Low-memory / web-only frontend**:

```bash
./dev-lite.sh                    # air backend/relay + coordinator runners + web only
./dev-lite.sh --backend-only     # skip frontend
pnpm proto:gen-go                # regenerate proto/gen/go (needs protoc)
pnpm proto:gen-go-all            # proto + amesh convert sync
```

The script auto-generates `.env` with worktree-hashed ports so multiple worktrees can coexist. Actual ports are printed on startup (or read from `deploy/dev/.env`).

Test accounts seeded by `init-seed.sh`:

- **User**: `dev@agentcloud.local` / `AdminAb123456`
- **Admin**: `admin@agentcloud.local` / `Ab123456`

## Contributors in mainland China

Docker image pulls through `docker.io` can be slow or unreliable from mainland China. **Configure registry mirrors once on your machine** — the Dockerfiles in this repo intentionally use canonical image names so this works transparently, with automatic fallback to Docker Hub when a mirror is unavailable.

Edit `~/.docker/daemon.json` (Docker Desktop) or `/etc/docker/daemon.json` (Linux), then restart Docker:

```json
{
  "registry-mirrors": [
    "https://docker.1ms.run",
    "https://docker.m.daocloud.io",
    "https://dockerproxy.com"
  ]
}
```

Do **not** hard-code mirror prefixes into the Dockerfiles — mirror metadata occasionally drifts out of sync with Docker Hub, which breaks builds for *everyone* and can't be fixed without a repo change. The daemon-level config is per-machine, auto-falls-back, and doesn't affect non-China contributors.

## Logs

```bash
tail -f deploy/dev/runtime/backend/backend.log   # air + backend stdout
tail -f deploy/dev/runtime/relay/relay.log
tail -f deploy/dev/web.log                       # next dev (web)
docker compose logs -f postgres                  # docker infra
docker compose logs -f runner-e2e-echo runner-claude-code runner-codex-cli
```

## Common issues

**Port conflicts between worktrees**: ports are derived from the worktree directory name. If you see a port clash, it usually means two worktrees hashed to the same port — rename one or set `PORT_SEED` in `.env`.

**`docker compose build` fails with `failed to resolve source metadata ... not found`**: Your Docker daemon is routing through a broken registry mirror. See the China section above — either fix the mirror list in `daemon.json` or remove it entirely to use Docker Hub directly.

**Runner can't connect to backend**: check `GRPC_PUBLIC_ENDPOINT` in the generated `.env`. For local (non-Docker) runners, this must be reachable from the host — usually `grpcs://localhost:<GRPC_PORT>`.
