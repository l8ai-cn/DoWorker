# Pod Workspace Artifacts Implementation Plan

**Goal:** Display deliverables generated in an ACP Worker's sandbox, including
Seedance MP4 output, directly in the existing Agent workspace.

**Architecture:** Add read-only Pod workspace filesystem routes backed by the
Runner `SandboxFsService`. Discover supported deliverables when a Web ACP
session opens and after processing returns to idle, then render them through
the existing Agent UI artifact components.

**Tech stack:** Go, Gin, Runner gRPC sandbox filesystem operations, TypeScript,
React, Vitest, Playwright.

## Task 1: Pod Workspace API

Files:

- `backend/internal/api/rest/v1/pod_workspace_artifacts.go`
- `backend/internal/api/rest/v1/pod_workspace_artifacts_test.go`
- `backend/internal/api/rest/v1/pods.go`
- `backend/internal/api/rest/v1/routes_pod_queue.go`

Steps:

- [x] Add failing authorization, list, read, Runner error, and content tests.
- [x] Add a small sandbox executor interface to `PodHandler`.
- [x] Authorize through `PodPolicy.AllowRead`.
- [x] Execute `changes` and `read` against the Pod's Runner and Pod key.
- [x] Register the read-only routes under `/api/v1/orgs/:slug/pods/:key`.
- [x] Verify focused handler tests.

Routes:

```text
GET /api/v1/orgs/:slug/pods/:key/resources/workspace/changes
GET /api/v1/orgs/:slug/pods/:key/resources/workspace/filesystem/*filepath
```

Verification:

```bash
cd backend
go test ./internal/api/rest/v1 \
  -run 'TestPodWorkspace|TestGetPodPreview|TestSendPodPrompt' -count=1
```

## Task 2: Frontend API Client

Files:

- `clients/web/src/lib/api/podWorkspaceArtifactApi.ts`
- `clients/web/src/lib/api/__tests__/podWorkspaceArtifactApi.test.ts`

Steps:

- [x] Add failing list, binary read, and truncation tests.
- [x] Use Pod routes directly without resolving an embedded session.
- [x] Normalize the platform REST root to `/api/v1` when API base is empty.
- [x] Decode base64 content into typed Blobs.
- [x] Project supported workspace files into `AgentArtifactItem` records.

Verification:

```bash
cd clients/web
node ../../node_modules/vitest/vitest.mjs run \
  src/lib/api/__tests__/podWorkspaceArtifactApi.test.ts \
  --config vitest.config.ts
```

## Task 3: Web ACP Discovery

Files:

- `clients/web/src/components/workspace/agent-ui/WebAcpSessionRuntime.ts`
- `clients/web/src/components/workspace/agent-ui/webAcpArtifactDiscovery.ts`
- `clients/web/src/components/workspace/agent-ui/webAcpArtifactProjection.ts`
- `clients/web/src/components/workspace/agent-ui/webAcpRuntimeDefaults.ts`
- `clients/web/src/components/workspace/agent-ui/webAcpRuntimeTypes.ts`
- `clients/web/src/components/workspace/agent-ui/webAcpSnapshot.ts`
- focused tests beside the runtime and snapshot

Steps:

- [x] Refresh artifacts when the runtime opens.
- [x] Refresh after a non-idle to idle transition.
- [x] Deduplicate concurrent refreshes.
- [x] Merge discovered and tool-emitted artifacts by `artifactId`.
- [x] Surface discovery errors without fabricating artifacts.
- [x] Exclude hidden runtime paths and deleted files in shared projection.

Verification:

```bash
cd clients/web
node ../../node_modules/vitest/vitest.mjs run \
  src/lib/api/__tests__/podWorkspaceArtifactApi.test.ts \
  src/components/workspace/agent-ui/WebAcpSessionRuntime.test.ts \
  src/components/workspace/agent-ui/webAcpSnapshot.test.ts \
  --config vitest.config.ts
```

## Task 4: Worker Lifecycle

Files:

- `backend/internal/service/agentpod/pod_orchestrator_worker_spec_snapshot.go`
- `backend/internal/service/agentpod/pod_worker_spec_test.go`

Steps:

- [x] Reproduce a manual-lifecycle Expert launching a non-perpetual Pod.
- [x] Project WorkerSpec `termination_policy=manual` to `Perpetual=true`.
- [x] Verify idle and completed policies remain non-perpetual.
- [x] Confirm the Seedance Expert snapshot launches a persistent Worker.

Verification:

```bash
cd backend
go test ./internal/service/agentpod \
  -run 'TestProjectWorkerSpecAppliesLifecyclePolicy|TestProjectWorkerSpecOmitsModelResourceForCredentialManagedWorker' \
  -count=1
```

## Task 5: Browser MP4 Verification

- [x] Generate a deterministic local H.264 MP4.
- [x] Place it at `output/seedance-platform-artifact-smoke.mp4` in the real
  Seedance Worker sandbox.
- [x] Confirm the artifact card, video preview, open action, and download action.
- [x] Confirm `readyState=4`, duration `2`, dimensions `640x360`, and no media
  error.
- [x] Confirm no artifact errors and no browser console errors.
- [x] Save evidence to
  `output/playwright/seedance-platform-video-artifact-final.png`.

The deterministic clip is local platform evidence, not an Ark-generated result.

## Task 6: Final Regression

- [x] Run all focused backend and frontend tests together.
- [x] Run `git diff --check` on the task files.
- [x] Scan the task diff for committed API keys or credentials.
- [x] Record the final Worker key, status, lifecycle, and screenshot.
- [ ] Clear the unrelated full Web typecheck failure in
  `WorkerTypeConfigStep.tsx`; the task-specific tests pass.

Final evidence:

- Worker: `7-standalone-30dcd07d`
- State: `running/idle`
- Lifecycle: `perpetual=true`
- WorkerSpec snapshot: `19`
- Browser video: H.264, `640x360`, 2 seconds, `readyState=4`
- Screenshot:
  `output/playwright/seedance-platform-video-artifact-final.png`
