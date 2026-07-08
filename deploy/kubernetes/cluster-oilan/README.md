# AgentsMesh on doops-oilan

Deploys the full platform (backend, relay, web, web-admin + Postgres/Redis/MinIO)
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

## Deploy

```bash
docker login repo.aiedulab.cn:8443           # one-time
./push-images.sh all                          # build + push every image to Harbor
DOOPS_TARGET=gw-oilan ./deploy.sh        # secrets + manifests + jobs via DoOps
```

`deploy.sh` defaults to `gw-oilan`. `push-images.sh` subsets: `platform` |
`infra` | `runners`.

> The runner Go binary resolves its config dir via `config.UserConfigDir()`
> (`~/.do-worker`, legacy `~/.agentsmesh`); older runner images that hardcoded
> `~/.agentsmesh` fail with `Runner not registered` in a fresh container.

## Endpoints

- App: http://doworker.l8an.cn:10007 (`/api`, `/proto.`, `/relay`, `/health`)
- Admin console: http://admin.doworker.l8an.cn:10007 (separate host — no `/admin` basePath)
- Object storage (presigned URLs): http://minio.doworker.l8an.cn:10007
- Test accounts: `dev@agentsmesh.local / AdminAb123456`, `admin@agentsmesh.local / Ab123456`

DNS for `*.l8an.cn` must point at the oilan node; ingress-nginx must expose **NodePort
10007** (HTTP) so external `:10007` reaches the controller. One-time patch on the
cluster (adjust namespace/service name if your install differs):

```bash
kubectl -n ingress-nginx patch svc ingress-nginx-controller -p \
  '{"spec":{"ports":[{"name":"http","port":80,"targetPort":"http","nodePort":10007}]}}'
```

TLS cert secret `l8an-wildcard-tls` must exist in the `default` namespace before
`deploy.sh` runs (copied into `agentsmesh`).

## Layout

| File | Purpose |
|------|---------|
| `00-namespace` `02-configmap` | namespace + shared non-secret env |
| `10/11/12-*` `13-minio-setup-job` | Postgres / Redis / MinIO + bucket |
| `20-migrate-job` `21/22-seed*` | DB migrate (embedded) + org/user/runner seed |
| `30-backend*` | backend Deployment/Service + SA/RBAC (kubectl via init container) |
| `31/32/33-*` | relay / web / web-admin (web-admin listens on :3001) |
| `34/35-runner-*` | standing runner pods |
| `40-ingress` | ingress-nginx routes (app / admin host / relay rewrite / minio host) |
| `60-prepull-daemonset` | warm agent-runtime image cache |

Secrets (`agentsmesh-secrets`, `agentsmesh-pki-ca`, `agentsmesh-regcred`) and the
one-shot Jobs are applied by `deploy.sh`, not kustomize. Generated material lives in
`_gen/` (git-ignored).
