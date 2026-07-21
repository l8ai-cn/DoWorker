# Worker-Per-Runner Pod Design

## Goal

Make the first oilan-ready implementation of dynamic workers match the product model that one user-facing worker maps to one dedicated Kubernetes pod.

The implementation should preserve the existing runner, PTY, ACP, relay, and agent launch code. Instead of rewriting each worker as a native Kubernetes workload immediately, the backend will provision a dedicated Runner pod per user worker and set `MAX_CONCURRENT_PODS=1`. The runner then starts exactly one logical doworker pod process inside that Kubernetes pod.

This gives the platform the operational behavior of one worker per Kubernetes pod while minimizing protocol and agent runtime changes.

## Non-Goals

- Do not rewrite Claude, Codex, or DoAgent execution to bypass the runner.
- Do not require dynamic Kubernetes Ingress per worker in the first implementation.
- Do not change the browser terminal or ACP interaction model.
- Do not remove standing runners; keep existing runner selection for environments that still use pooled runners.
- Do not make oilan depend on ports in the public host name. Public application traffic should use `dowork.l8ai.cn`.

## Current Model

The backend currently creates a logical pod row, resolves the agent launch command, selects a connected runner, and sends a `CreatePodCommand` over the runner gRPC stream.

The Kubernetes launcher currently provisions a runner pod keyed by organization and agent slug. That runner pod can host multiple logical doworker pods as local agent processes.

Current shape:

```text
Kubernetes Runner Pod
  -> runner process
    -> logical doworker pod A = local agent process
    -> logical doworker pod B = local agent process
```

Target first-phase shape:

```text
Kubernetes Worker Runner Pod for podKey A
  -> runner process with MAX_CONCURRENT_PODS=1
    -> logical doworker pod A = local agent process
```

## Proposed Architecture

Add a dedicated-worker provisioning path used when coordinator-managed workers are enabled.

The create flow becomes:

```text
CreatePod request
  -> resolve agent and AgentFile
  -> create logical pod row
  -> provision dedicated runner pod for podKey
  -> wait for runner registration/connection
  -> dispatch CreatePodCommand to that runner
  -> update pod status from runner events
```

The runner pod name, node id, labels, and annotations should be derived from the logical `podKey`, not only from `orgID + agentSlug`.

## Components

### WorkerProvisioner

Add a backend service that owns dedicated worker runner lifecycle.

Responsibilities:

- Build a `WorkerLaunchSpec` from organization, user, agent slug, pod key, runtime image, and coordinator config.
- Ensure a runner registration exists for the generated node id.
- Ask the configured launcher to create the runner pod.
- Wait until the runner is connected or fail with a bounded timeout.
- Return the selected runner id to `PodOrchestrator`.
- Delete or mark stale worker runner pods when logical pod termination is requested.

### RunnerLauncher

Extend the launcher model without breaking existing launchers.

Keep the existing `Launch(ctx, orgID, agentSlug)` path for pooled runner compatibility. Add a dedicated launch path, either as a new interface or an optional extension:

```go
type DedicatedRunnerLauncher interface {
    LaunchDedicated(ctx context.Context, spec WorkerLaunchSpec) error
}
```

The K8s implementation should support the dedicated path first. Docker and script launchers may return a clear unsupported error unless needed for local testing.

### WorkerLaunchSpec

Use an explicit spec so Kubernetes rendering does not reach back into orchestration request structs.

```go
type WorkerLaunchSpec struct {
    OrgID      int64
    UserID     int64
    PodKey     string
    AgentSlug  string
    Image      string
    Namespace  string
    RunnerName string
    RunnerNodeID string
    RunnerOrgSlug string

    BackendURL   string
    GRPCEndpoint string
    RelayBaseURL string

    MaxConcurrentPods int
    Labels      map[string]string
    Annotations map[string]string
}
```

For the first implementation, `MaxConcurrentPods` must be forced to `1` for dedicated workers, regardless of the global pooled-runner default.

### K8s Manifest

The dedicated K8s pod should include labels that make lifecycle, debugging, and future gateway routing deterministic:

```yaml
labels:
  app.kubernetes.io/name: agentcloud-runner
  app.kubernetes.io/component: dedicated-worker
  doworker.io/pod-key: <podKey>
  doworker.io/agent-slug: <agentSlug>
  doworker.io/org-id: "<orgID>"
```

The pod should set:

```yaml
env:
  - name: RUNNER_NODE_ID
    value: <dedicated node id>
  - name: MAX_CONCURRENT_PODS
    value: "1"
```

### Gateway and External Access

The first implementation should not create one Kubernetes Ingress per worker. It should leave public ingress stable and introduce a gateway-compatible metadata model.

Recommended external shape:

```text
https://dowork.l8ai.cn/preview/<podKey>/*
```

Initial backend implementation can store or compute a future preview endpoint without enabling proxying yet. If preview routing is implemented in this phase, prefer a fixed Gateway/Preview service that authenticates the request, looks up `podKey`, and proxies to the dedicated worker pod or runner-local preview proxy.

Dynamic Ingress remains a later route mode for special cases that require a dedicated subdomain or root path.

## Configuration

Add a feature flag for the dedicated worker path:

```text
COORDINATOR_DEDICATED_WORKER_PODS=true
```

Keep existing image mapping:

```text
COORDINATOR_RUNNER_IMAGES=codex-cli=...,do-agent=...,claude-code=...
```

Add optional naming controls:

```text
COORDINATOR_WORKER_NODE_ID_PREFIX=coord-worker-
COORDINATOR_WORKER_POD_NAME_PREFIX=amesh-worker-
COORDINATOR_WORKER_READY_TIMEOUT_SECONDS=120
```

When the flag is disabled, current runner selection and pooled runner behavior must continue unchanged.

## Error Handling

- If dedicated pod provisioning fails before runner dispatch, mark the logical pod as dispatch failed with a clear provisioning error.
- If the runner pod becomes ready but never connects, delete the Kubernetes pod when safe and mark the logical pod as runner unreachable.
- If `CreatePodCommand` dispatch fails after provisioning, keep existing dispatch failure handling and attempt best-effort cleanup of the dedicated runner pod.
- Termination should clean both the logical runner process and the dedicated Kubernetes runner pod.

## Testing

Add focused unit tests before implementation where possible:

- K8s manifest rendering for dedicated worker labels, node id, image, and `MAX_CONCURRENT_PODS=1`.
- Dedicated launcher uses `podKey`-derived resource names instead of `orgID + agentSlug`.
- Dedicated mode in orchestration provisions before dispatch and dispatches to the returned runner id.
- Feature flag disabled preserves existing runner selection path.
- Provisioning failure marks pod dispatch failed and does not send `CreatePodCommand`.

Integration coverage can use fake command runners and fake runner connection managers rather than a live Kubernetes cluster.

## Implementation Order

1. Introduce `WorkerLaunchSpec` and dedicated K8s manifest rendering tests.
2. Add `LaunchDedicated` support to `K8sLauncher`.
3. Add coordinator config for dedicated worker pods.
4. Add `WorkerProvisioner` with fakeable interfaces.
5. Wire `PodOrchestrator` so dedicated mode provisions after pod row creation and before dispatch.
6. Add cleanup hooks for dedicated worker termination.
7. Add oilan deployment config for dedicated worker mode.

## Open Decisions

The first implementation will use fixed Ingress plus future Gateway routing, not dynamic per-worker Ingress. A dynamic Ingress route mode can be added later if a worker requires subdomain isolation or root-path compatibility.

The runner remains the runtime boundary in this phase. A later architecture can make each worker a pure agent container without the runner, but that should be a separate migration because it changes terminal, ACP, relay, and recovery semantics.
