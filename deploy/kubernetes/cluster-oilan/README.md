# AgentsMesh on doops-oilan

Deploys the full platform (backend, Marketplace API, Marketplace Storefront,
relay, web, web-admin, mobile + Postgres/Redis/MinIO)
to the single-node `doops-oilan` k3s cluster (`gpu-ampere01`, amd64) via DoOps,
with all images served from the in-cluster Harbor
(`repo.aiedulab.cn:8443/agentsmesh/*`, node-local so pulls are effectively free).

## Worker model

The backend runs on K8s. AI Workers are PTY processes hosted inside **Runner
pods**:

- **Standing runners** (`34-runner-e2e-echo.yaml`, `35-runner-video-studio.yaml`) are
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
git add release/kustomization.yaml && git commit
DOOPS_TARGET=gw-oilan-node ./deploy.sh         # secrets + manifests + jobs via DoOps
```

`deploy.sh` defaults to `gw-oilan-node`. `push-images.sh` subsets: `platform` |
`marketplace-core` | `video-expert` | `video-runtime` | `web` | `infra` | `runners`.
`video-expert` rebuilds Backend, Marketplace API, Marketplace Web, Core Web,
and Web Admin while retaining the current Relay digest. `video-runtime` builds
and pushes only the Video Studio runner image. `marketplace-core` rebuilds
Backend, Marketplace API, Marketplace Web, and Core Web while pinning the current
Relay/Web-Admin registry digests. `web` rebuilds only Core Web and retains all
other digests.

The July 2026 OILAN database repair for the historical dirty migration 208 is a
one-time, fail-closed operation. Run it only when `schema_migrations` contains
exactly `208 / dirty=true` and migration 207's `agents.adapter_id` is absent:

```bash
DOOPS_TARGET=gw-oilan-node ./repair-migration-208.sh
```

The repair asserts migrations 205-206, applies the repository-exact 207-208 SQL,
uses the pinned Backend image to mark 208 clean and migrate through 222, then
removes its Job and ConfigMap. It is intentionally excluded from kustomize.

The migration 222 Backend hotfix is reproducibly built from the prior immutable
runtime image when Docker Hub is unavailable:

```bash
./build-backend-migration-hotfix.sh
```

The script locks the Go version, linux/amd64 target, Server checksum, base image,
and resulting image digest before updating the Harbor `latest` tag.

For the mobile Worker access path, do not run the full reconcile when unrelated
workloads are newer in the cluster. Build the three affected images, pin their
immutable digests in `30-backend.yaml`, `31-relay.yaml`, and `42-mobile.yaml`,
then deploy only those resources:

```bash
./push-images.sh mobile-access
DOOPS_TARGET=gw-oilan-node ./deploy-mobile-access.sh
```

`deploy-mobile-access.sh` refuses mutable image tags. It applies the shared
ConfigMap, runs the migration Job with the Backend digest pinned in
`30-backend.yaml`, rolls out that Backend, then runs its
`worker-definition-sync` command before rolling out Relay, Mobile Ingress, and
Mobile. Do not bypass the migration or sync: the migration protects the
Pod/session schema and the sync publishes Codex ACP/PTY metadata.

After rollout, run the release smoke from a trusted operator machine. It fails
closed when the Backend does not expose Codex ACP/PTY mode metadata or exactly
one default model resource:

```bash
MOBILE_SMOKE_USERNAME=admin@agentsmesh.local \
MOBILE_SMOKE_PASSWORD='...' \
./verify-mobile-worker-access.sh
```

Set `MOBILE_SMOKE_RUN_INTERACTIONS=true` only in a test organization with a
configured Codex model resource. That mode creates disposable ACP and PTY
Workers, verifies an ACP response and PTY control lease over the direct Relay
data plane, and deletes both sessions.

### What `./deploy.sh` does (often called "reseed")

Each run is a **full reconcile**, not a DB-only reset:

1. **Secrets** — restore existing cluster secrets or generate first-deploy values.
2. **Immutable release** — reject any platform image not pinned by registry digest.
3. **`kubectl apply -k .`** — re-apply every Deployment/ConfigMap/Ingress from git.
   Live hotfixes (`kubectl set env …`) are **overwritten** on the next deploy.
4. **Init migrations** — Backend and Marketplace run embedded migrations before
   their application containers start.
5. **Seed job** — delete old `seed` job, run `21-seed-configmap.yaml` → `seed.sql`.
6. **Operator catalog** — run the idempotent video Skill/Expert submission,
   review, and publication bootstrap with model resource `1` and runtime image
   `4`.

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
- Marketplace Storefront: https://market.l8ai.cn
- Marketplace API: https://market.l8ai.cn/api/marketplace/v1
- Organization marketplace: https://dowork.l8ai.cn/dev-org/marketplace
- Admin console: https://admin.l8ai.cn (separate host — no `/admin` basePath)
- Pod preview gateway: https://preview.l8ai.cn/preview
- Object storage (presigned URLs): https://minio.dowork.l8ai.cn
- Test account: `admin@agentsmesh.local / Ab123456`

DNS for `dowork.l8ai.cn` / `market.l8ai.cn` / `mobile.l8ai.cn` /
`admin.l8ai.cn` / `preview.l8ai.cn` / `minio.dowork.l8ai.cn` must point at the oilan node. All
public URLs share one domain family so relay/WebSocket URLs from
`GetPodConnection` match the page origin (mixed `l8an.cn` / `l8ai.cn` hosts caused
403 on terminal attach). Ingress-nginx may expose **NodePort 10007** (HTTP) so
external `:10007` reaches the controller. One-time patch on the cluster (adjust
namespace/service name if your install differs):

```bash
kubectl -n ingress-nginx patch svc ingress-nginx-controller -p \
  '{"spec":{"ports":[{"name":"http","port":80,"targetPort":"http","nodePort":10007}]}}'
```

TLS cert secret `l8ai-wildcard-tls` must exist in `agentsmesh`. On the first
deployment, `deploy.sh` copies it from `default`; later deployments keep and
validate the existing `agentsmesh` Secret.

## Layout

| File | Purpose |
|------|---------|
| `00-namespace` `02-configmap` | namespace + shared non-secret env |
| `10/11/12-*` `13-minio-setup-job` | Postgres / Redis / MinIO + bucket |
| `21/22-seed*` | idempotent org/user/runner seed |
| `30-backend*` | backend Deployment/Service + SA/RBAC (kubectl via init container) |
| `31/32/33/42-*` | relay / web / web-admin / mobile |
| `34/35-runner-*` | standing e2e and video-studio runner pods |
| `38-marketplace` | Marketplace API Deployment/Service + init migration |
| `39-marketplace-web` | independent public Marketplace Storefront |
| `release/kustomization` | immutable platform image digests |
| `40-ingress` | ingress-nginx routes (app / admin host / relay rewrite / minio host) |
| `44-preview-ingress` | isolated HTTPS pod preview gateway |
| `60-prepull-daemonset` | warm agent-runtime image cache |

Secrets (`agentsmesh-secrets`, `agentsmesh-pki-ca`, `agentsmesh-regcred`) and
one-shot seed jobs are applied by `deploy.sh`, not kustomize. Generated secret
material lives in `_gen/` (git-ignored).
