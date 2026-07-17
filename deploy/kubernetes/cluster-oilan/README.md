# AgentsMesh on doops-oilan

Deploys the full platform (backend, Marketplace API, Marketplace Storefront,
relay, web, web-admin, mobile + Postgres/Redis/MinIO)
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
DOOPS_SESSION=<release-session> ./configure-harbor-upload-token.sh
./push-images.sh all                          # build + push every image to Harbor
git status --short                            # review generated locks/evidence
git add deploy/kubernetes/cluster-oilan/release \
  deploy/kubernetes/cluster-oilan/30-backend.yaml \
  deploy/kubernetes/cluster-oilan/60-prepull-daemonset.yaml \
  docker/agent-runtime/do-agent-release.json \
  backend/internal/domain/workerruntime/runtime_catalog.lock.json \
  config/worker-types tools/loops/worker-onboarding/catalog-loop
git commit
git push
DOOPS_TARGET=gw-oilan-node ./deploy.sh         # secrets + manifests + jobs via DoOps
```

Runner runtime builds resolve their Node base only through the locked
`runner-node-base@sha256:...` Harbor reference and fail before building if its
digest differs or either `linux/amd64` or `linux/arm64` is absent.
`configure-harbor-upload-token.sh` applies the 120-minute Harbor system setting
with the cluster-held administrator Secret. Image publishing uses only the
Docker push credential to verify the issued token lifetime before any build
starts. Large runtime layers can exceed Harbor's 30-minute default; an expired
token makes Harbor reject the final blob commit after receiving the entire
layer.

Build and deploy scripts refuse a dirty tree, detached HEAD, a commit that is
not the current remote branch HEAD, or missing release-specific CI checks.
The status of explicitly named deployment and migration jobs for the
independent US West/CN environments does not block an Oilan release; every
other check must still finish successfully, and the three Loop/Seedance release
checks are always required.
`release/source.json` is mandatory release provenance. It records the release
commit and each platform image's exact source revision, so incremental image
releases can retain older immutable digests without weakening provenance.
Commit it with the generated digest locks and runtime evidence.

`deploy.sh` defaults to `gw-oilan-node`. `push-images.sh` subsets: `platform` |
`marketplace-core` | `video-expert` | `video-runtime` | `web` | `infra` | `runners`.
`video-expert` rebuilds Backend, Marketplace API, Marketplace Web, Core Web,
and Web Admin while retaining the current Relay digest. `video-runtime` builds
and pushes only the Video Studio runner image. `marketplace-core` rebuilds
Backend, Marketplace API, Marketplace Web, and Core Web while pinning the current
Relay/Web-Admin registry digests. `web` rebuilds only Core Web and retains all
other digests.

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
3. **Pre-migration backup** — write a verified custom-format PostgreSQL dump to
   `/root/backups/agentsmesh/pre-migrate-<UTC>.dump`; any backup failure stops
   the release before changing the PostgreSQL image or running migrations.
4. **Pinned migration Job** — use the exact Backend digest from the rendered
   release to run `/app/server migrate up`, and block until all pending
   migrations complete. A dirty migration state is rejected; it is never
   force-cleared by the normal deployment path.
5. **`kubectl apply -f /tmp/agentsmesh-release.yaml`** — only after migration
   success, re-apply every Deployment/ConfigMap/Ingress from git.
   Live hotfixes (`kubectl set env …`) are **overwritten** on the next deploy.
6. **Seed and storage jobs** — run the idempotent seed, ensure the MinIO bucket
   and its one-day `workspace-artifacts/` expiry rule, then sync Worker definitions.

Production migration state `222 dirty=true` caused by the historical
`video-studio` insert must be repaired only after the corrected Backend image is
published and committed:

```bash
MIGRATION_REPAIR_ACK=repair-dirty-222-video-studio \
DOOPS_SESSION=<release-session> \
bash deploy/kubernetes/cluster-oilan/repair-migration-222.sh
```

The repair requires a clean, pushed `main` commit with successful GitHub
checks, verifies the exact dirty state and schema preconditions, stops
application writes, creates a checksummed backup, and reruns migration `000222`
from version `221`. Any failed precondition leaves the database untouched.

The executable database and GitOps rollback procedure is in
[`ROLLBACK.md`](ROLLBACK.md).

The seed SQL is **idempotent** (`WHERE NOT EXISTS`): it ensures `dev-org`, the
admin user, subscription, and pre-registered runner `node_id`s exist. It does **not**
truncate pods/users/orgs; it also does **not** seed `dev@agentsmesh.local` (admin-only
on this cluster). Re-running seed alone does not change relay/web env — only step 2 does.

> The runner Go binary resolves its config dir via `config.UserConfigDir()`
> (`~/.do-worker`, legacy `~/.agentsmesh`); older runner images that hardcoded
> `~/.agentsmesh` fail with `Runner not registered` in a fresh container.

## Endpoints

- App: https://dowork.l8ai.cn (`/api`, `/proto.`, `/relay`, `/health`)
- Isolated Pod preview: `https://<pod-key>.preview.dowork.l8ai.cn` (`/preview` only)
- Mobile Worker entry: https://mobile.l8ai.cn
- Marketplace Storefront: https://market.l8ai.cn
- Marketplace API: https://market.l8ai.cn/api/marketplace/v1
- Organization marketplace: https://dowork.l8ai.cn/dev-org/marketplace
- Admin console: https://admin.l8ai.cn (separate host — no `/admin` basePath)
- Pod preview gateway: `https://<pod-key>.preview.dowork.l8ai.cn/preview`
- Object storage (presigned URLs): https://minio.dowork.l8ai.cn
- Test account: `admin@agentsmesh.local / Ab123456`

DNS for `dowork.l8ai.cn` / `market.l8ai.cn` / `mobile.l8ai.cn` /
`admin.l8ai.cn` / `*.preview.dowork.l8ai.cn` / `minio.dowork.l8ai.cn` must point at the oilan node.
The `agentsmesh` namespace must contain `dowork-preview-wildcard-tls` covering
`*.preview.dowork.l8ai.cn`. All
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
| `10/11/12-*` `13-minio-setup-job` | Postgres / Redis / MinIO + bucket and temporary artifact TTL |
| `20-migrate-job` | Digest-pinned embedded Backend migrations before workload rollout |
| `21/22-seed*` | idempotent org/user/runner seed |
| `30-backend*` | backend Deployment/Service + SA/RBAC (kubectl via init container) |
| `31/32/33/42-*` | relay / web / web-admin / mobile |
| `34/35-runner-*` | standing runner pods |
| `38-marketplace` | Marketplace API Deployment/Service + init migration |
| `39-marketplace-web` | independent public Marketplace Storefront |
| `release/kustomization` | immutable platform image digests |
| `40-ingress` | ingress-nginx routes (app / admin host / relay rewrite / minio host) |
| `44-preview-ingress` | isolated HTTPS pod preview gateway |
| `60-prepull-daemonset` | warm agent-runtime image cache |

Secrets (`agentsmesh-secrets`, `agentsmesh-pki-ca`, `agentsmesh-regcred`) and
one-shot seed jobs are applied by `deploy.sh`, not kustomize. Existing values
are read from the cluster for each release; generated `_gen/` material is
removed locally on every exit.
