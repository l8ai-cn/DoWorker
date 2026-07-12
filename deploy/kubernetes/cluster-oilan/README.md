# AgentsMesh on doops-oilan

Deploys the full platform (backend, relay, web, web-admin, mobile + Postgres/Redis/MinIO)
to the single-node `doops-oilan` k3s cluster (`gpu-ampere01`, amd64) via DoOps,
with all images served from the in-cluster Harbor
(`repo.aiedulab.cn:8443/agentsmesh/*`, node-local so pulls are effectively free).

## Worker model

The backend runs on K8s. AI Workers are PTY processes hosted inside **Runner
pods**:

- **Standing runners** (`34-runner-e2e-echo.yaml`, `35-runner-claude.yaml`) are
  pre-registered in the DB seed and connect on boot — the reliable default. Creating
  a Worker from the UI spawns a PTY inside a standing runner.
- **On-demand runners**: the backend `K8sLauncher` (`COORDINATOR_RUNNER_LAUNCHER=k8s`)
  `kubectl apply`s a runner pod per org/agent when the coordinator needs an agent
  with no online runner. The pod's node_id is
  `coord-runner-amesh-runner-<orgID>-<agent>`; those node_ids are **pre-registered by
  the seed** (`21-seed-configmap.yaml`) so the pod passes mTLS `validateRunner` on
  first connect (the backend does not trust-on-first-use unknown node_ids).

Runners authenticate to the backend over gRPC/mTLS using certs derived from a shared
CA (`agentsmesh-pki-ca` secret, mounted at `/app/ssl`). The runner's TLS verification
is chain-only, so in-cluster runners dial `backend:9090` directly.

Because Harbor is node-local, runner + backend-launcher image pull policy is `Always`
(cheap, and avoids stale `:latest` cache after image rebuilds).

## Deploy / reseed

```bash
docker login repo.aiedulab.cn:8443           # one-time
./push-images.sh all                          # build + push every image to Harbor
DOOPS_TARGET=gw-oilan ./deploy.sh        # secrets + manifests + jobs via DoOps
```

`deploy.sh` defaults to `gw-oilan`. `push-images.sh` subsets: `platform` |
`infra` | `runners`.

For the mobile Worker access path, do not run the full reconcile when unrelated
workloads are newer in the cluster. Build the three affected images, pin their
immutable digests in `30-backend.yaml`, `31-relay.yaml`, and `42-mobile.yaml`,
then deploy only those resources:

```bash
./push-images.sh mobile-access
DOOPS_TARGET=gw-oilan-node ./deploy-mobile-access.sh
```

`deploy-mobile-access.sh` applies the shared ConfigMap, backend, relay, mobile
Ingress, and mobile Deployment only. It refuses mutable image tags.

### What `./deploy.sh` does (often called "reseed")

Each run is a **full reconcile**, not a DB-only reset:

1. **Secrets** — generate or reuse `_gen/` CA + app secrets + Harbor pull creds.
2. **`kubectl apply -k .`** — re-apply every Deployment/ConfigMap/Ingress from git.
   Live hotfixes (`kubectl set env …`) are **overwritten** on the next deploy.
3. **Migrate job** — delete old `migrate` job, run embedded SQL migrations (idempotent).
4. **Seed job** — delete old `seed` job, run `21-seed-configmap.yaml` → `seed.sql`.

The seed SQL is **idempotent** (`WHERE NOT EXISTS`): it ensures `dev-org`, the
admin user, subscription, and pre-registered runner `node_id`s exist. It does **not**
truncate pods/users/orgs; it also does **not** seed `dev@agentsmesh.local` (admin-only
on this cluster). Re-running seed alone does not change relay/web env — only step 2 does.

> The runner Go binary resolves its config dir via `config.UserConfigDir()`
> (`~/.do-worker`, legacy `~/.agentsmesh`); older runner images that hardcoded
> `~/.agentsmesh` fail with `Runner not registered` in a fresh container.

## Endpoints

- App: https://dowork.l8ai.cn (`/api`, `/proto.`, `/relay`, `/health`)
- Mobile Worker entry: https://mobile.l8ai.cn
- Admin console: https://admin.dowork.l8ai.cn (separate host — no `/admin` basePath)
- Object storage (presigned URLs): https://minio.dowork.l8ai.cn
- Test account: `admin@agentsmesh.local / Ab123456`

DNS for `dowork.l8ai.cn` / `mobile.l8ai.cn` / `admin.dowork.l8ai.cn` / `minio.dowork.l8ai.cn` must point at the
oilan node. All public URLs share one domain family so relay/WebSocket URLs from
`GetPodConnection` match the page origin (mixed `l8an.cn` / `l8ai.cn` hosts caused
403 on terminal attach). Ingress-nginx may expose **NodePort 10007** (HTTP) so
external `:10007` reaches the controller. One-time patch on the cluster (adjust
namespace/service name if your install differs):

```bash
kubectl -n ingress-nginx patch svc ingress-nginx-controller -p \
  '{"spec":{"ports":[{"name":"http","port":80,"targetPort":"http","nodePort":10007}]}}'
```

TLS cert secret `l8ai-wildcard-tls` must exist in the `default` namespace before
`deploy.sh` runs (copied into `agentsmesh` alongside `l8an-wildcard-tls` if needed).

## Layout

| File | Purpose |
|------|---------|
| `00-namespace` `02-configmap` | namespace + shared non-secret env |
| `10/11/12-*` `13-minio-setup-job` | Postgres / Redis / MinIO + bucket |
| `20-migrate-job` `21/22-seed*` | DB migrate (embedded) + org/user/runner seed |
| `30-backend*` | backend Deployment/Service + SA/RBAC (kubectl via init container) |
| `31/32/33/42-*` | relay / web / web-admin / mobile |
| `34/35-runner-*` | standing runner pods |
| `40-ingress` | ingress-nginx routes (app / admin host / relay rewrite / minio host) |
| `60-prepull-daemonset` | warm agent-runtime image cache |

Secrets (`agentsmesh-secrets`, `agentsmesh-pki-ca`, `agentsmesh-regcred`) and the
one-shot Jobs are applied by `deploy.sh`, not kustomize. Generated material lives in
`_gen/` (git-ignored).
