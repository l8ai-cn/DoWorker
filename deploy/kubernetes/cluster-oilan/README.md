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

When the build host cannot reach the node-local Harbor, dispatch the
`Oilan Image Publish` workflow on the verified `main` commit. It builds the
affected platform images into immutable Docker Hub staging tags and publishes
the `oilan-staging-release` artifact. Download its JSON manifest on an operator
machine that can reach both registries, then promote the exact manifests:

```bash
gh run download <run-id> -n oilan-staging-release -D /tmp/oilan-release
docker login docker.io
docker login repo.aiedulab.cn:8443
./promote-staged-images.sh /tmp/oilan-release/oilan-staging-release.json
git status --short
```

Harbor API calls use normal TLS verification. When Harbor uses a private CA,
set `HARBOR_CA_CERT=/path/to/harbor-ca.pem`; insecure TLS is not supported.
Docker credentials may use `auths`, `credHelpers`, or `credsStore`.

Promotion refuses a staging manifest from a different `main` commit, pulls
and verifies every source before changing Harbor, checks each `linux/amd64`
platform and source revision label, pushes all layers, and requires the Harbor
digest to equal the staged digest. It then updates the immutable release lock
and provenance metadata. Commit and push those generated files before running
`deploy.sh`.

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
commit and every platform and managed Runner image's exact source revision, so
incremental image releases can retain older immutable digests without weakening provenance.
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

`deploy-mobile-access.sh` refuses mutable image tags and requires the same DoSql
database evidence as the full reconcile. It applies the shared ConfigMap, rolls
out Backend, runs `worker-definition-sync`, then rolls out Relay, Mobile
Ingress, and Mobile. It does not run migrations or seed SQL.

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
3. **DoSql database gate** — the repository verifier requires a real append-only
   DoSql journal and immutable evidence artifact for
   `DOSQL_RELEASE_DB_TARGET`, `DOSQL_RELEASE_DB_MODE`,
   `DOSQL_RELEASE_DB_SESSION`, `DOSQL_RELEASE_CHANGE_ID`, and
   `DOSQL_RELEASE_OPERATION_ID`. It verifies the canonical
   `db_agentsmesh_prod_postgres` target identity, the change-specific
   hash-chained journal, the immutable evidence fingerprint, and the latest
   schema version. The deploy path never runs DDL/DML, creates a migration Job,
   or queries PostgreSQL directly.
4. **`kubectl apply -f /tmp/agentsmesh-release.yaml`** — after the audited
   database gate, re-apply every Deployment/ConfigMap/Ingress from git.
   Live hotfixes (`kubectl set env …`) are **overwritten** on the next deploy.
5. **Storage and catalog jobs** — ensure the MinIO bucket and its one-day
   `workspace-artifacts/` expiry rule, then sync Worker definitions.

Historical migration repairs are documented in [`MIGRATION_REPAIRS.md`](MIGRATION_REPAIRS.md).
The executable GitOps rollback procedure is in [`ROLLBACK.md`](ROLLBACK.md).
`21-seed-configmap.yaml` is idempotent SQL source material for an audited DoSql
change; normal deploy does not execute it.

> The runner Go binary resolves its config dir via `config.UserConfigDir()`
> (`~/.do-worker`, legacy `~/.agentsmesh`); older runner images that hardcoded
> `~/.agentsmesh` fail with `Runner not registered` in a fresh container.

## Endpoints

Public hostnames, DNS, wildcard preview, and ingress notes are in
[`ENDPOINTS.md`](ENDPOINTS.md).

TLS cert secret `l8ai-wildcard-tls` must exist in `agentsmesh`. On the first
deployment, `deploy.sh` copies it from `default`; later deployments keep and
validate the existing `agentsmesh` Secret.

## Layout

| File | Purpose |
|------|---------|
| `00-namespace` `02-configmap` | namespace + shared non-secret env |
| `10/11/12-*` `13-minio-setup-job` | Postgres / Redis / MinIO + bucket and temporary artifact TTL |
| `21-seed-configmap` | seed SQL source material for the audited DoSql change plan |
| `30-backend*` | backend Deployment/Service + SA/RBAC (kubectl via init container) |
| `31/32/33/42-*` | relay / web / web-admin / mobile |
| `34/35-runner-*` | standing runner pods |
| `38-marketplace` | Marketplace API Deployment/Service |
| `39-marketplace-web` | independent public Marketplace Storefront |
| `release/kustomization` | immutable platform image digests |
| `40-ingress` | ingress-nginx routes (app / admin host / relay rewrite / minio host) |
| `44-preview-ingress` | isolated HTTPS pod preview gateway |
| `60-prepull-daemonset` | warm agent-runtime image cache |

Secrets (`agentsmesh-secrets`, `agentsmesh-pki-ca`, `agentsmesh-regcred`) are
applied by `deploy.sh`, not kustomize. Existing values are read from the cluster
for each release; generated `_gen/` material is removed locally on every exit.
