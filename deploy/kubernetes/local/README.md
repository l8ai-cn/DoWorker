# Local Kubernetes Runners

Dev runners can run on a local K8s cluster (Docker Desktop → Kubernetes) instead of `docker-compose.runners.yml`.

## Quick start

1. Enable Kubernetes in Docker Desktop.
2. Start the dev stack with cluster runners:

```bash
./deploy/dev/dev.sh --runners-k8s
```

This will:

- Start docker infra only (postgres, traefik, …) — **no** compose runner containers
- Build runner images locally via `docker compose build` (same images as compose mode)
- Generate `deploy/dev/runtime/runners-k8s/manifest.yaml` from `.env` ports + dev SSL/SSH
- `kubectl apply` the configured runner Deployments into namespace `agentsmesh`

## Files

| File | Role |
|------|------|
| `runners-workloads.yaml` | PVCs, Services, Deployments (placeholders for image tag + MCP hostPort) |
| `deploy/dev/lib/generate_runners_k8s_manifest.sh` | ConfigMaps/Secrets + patched manifest |
| `deploy/dev/lib/runners_k8s.sh` | Build, apply, hot-swap, teardown |

`runners-cluster.yaml` is legacy (monolithic); prefer the generator + `runners-workloads.yaml`.

## Operations

```bash
kubectl get pods -n agentsmesh
kubectl logs -n agentsmesh deployment/runner-e2e-echo -f
./deploy/dev/dev.sh --reset-runners   # hot-swap rebuilt binary into K8s pods
./deploy/dev/dev.sh --clean           # deletes namespace agentsmesh
```

Runners reach host backend/traefik via `host.docker.internal:<HTTP_PORT>` (worktree-aware).
