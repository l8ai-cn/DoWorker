# Worker Creation and Publishing Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the split Worker creation paths with one four-step workflow backed by an immutable WorkerSpec snapshot, then publish Experts and selected sandbox Skills from that server-owned state.

**Architecture:** `CreatePodRequest` carries a structured WorkerSpec draft through TypeScript, protobuf, Rust Core, and Connect. The backend resolves organization-scoped references, immutable runtime metadata, and the exact model resource before creating a snapshot and compiling the existing AgentFile runtime layer. Expert publication references that snapshot; Skill publication discovers and packages candidates through the Runner sandbox filesystem channel without exposing host paths.

**Tech Stack:** Go/GORM/PostgreSQL, Connect-RPC/Protobuf, Rust Core/WASM, Next.js/TypeScript, Vitest, Playwright.

---

## File Map

- `backend/internal/domain/workerruntime/catalog.go`: immutable runtime image, compute target, deployment, and resource profile catalog.
- `backend/internal/service/workercreation/`: draft conversion, scoped resolution, preflight, AgentFile compilation, and AI draft filling.
- `backend/internal/service/agentpod/`: orchestration integration and persisted snapshot linkage only.
- `backend/internal/service/expert/`: snapshot-backed publish and run.
- `backend/internal/service/skill/worker_publish.go`: selected Worker sandbox Skill publication.
- `runner/internal/runner/worker_skill_*.go`: secure candidate discovery and package export below allowlisted roots.
- `proto/pod/v1/pod.proto`: WorkerSpec draft, create options, preflight, fill, Pod snapshot ID, and Skill publish RPC messages.
- `clients/core/crates/{api-client,services,wasm}/`: binary Worker RPC bridge.
- `clients/web/src/components/pod/CreatePodForm/`: reducer-owned four-step workflow.
- `clients/web/src/components/workers/`: Fill with AI integration and active Worker publish surfaces.

### Task 1: Worker Runtime Catalog and Wire Contract

**Files:**
- Create: `backend/internal/domain/workerruntime/catalog.go`
- Create: `backend/internal/domain/workerruntime/catalog_test.go`
- Modify: `backend/internal/domain/workerspec/placement.go`
- Modify: `backend/internal/domain/workerspec/placement_validation.go`
- Modify: `proto/pod/v1/pod.proto`
- Regenerate ignored Go/Rust mirrors and commit the TypeScript proto mirror
- Test: `backend/internal/domain/agentpod/pod_proto_contract_test.go`

- [x] **Step 1: Write failing catalog and wire tests**

Assert that:

```go
catalog := workerruntime.DefaultCatalog()
require.NotEmpty(t, catalog.ImagesFor("codex-cli"))
assert.Regexp(t, `^sha256:[a-f0-9]{64}$`, catalog.ImagesFor("codex-cli")[0].Digest)
assert.True(t, catalog.Target(1).SupportsPooled)
assert.False(t, catalog.Target(1).SupportsDedicated)
```

The protobuf contract must expose `WorkerSpecDraft`, `ListWorkerCreateOptions`,
`PreflightWorker`, `FillWorkerDraft`, `worker_spec_snapshot_id`, and selected
Skill publish messages without legacy default-auth fields.

- [x] **Step 2: Run tests and verify RED**

Run:

```bash
go test ./backend/internal/domain/workerruntime ./proto/gen/go/pod/v1 -count=1
```

Expected: missing catalog types and protobuf fields.

- [x] **Step 3: Implement the immutable catalog and protobuf messages**

Use the verified OCI index digests for the currently published Codex, Claude,
and Gemini runtime images. Keep unavailable Worker types absent instead of
inventing an image. Define the pooled organization-runner target and a managed
Kubernetes target whose disabled reason is returned when dedicated
provisioning is not configured. Define server-owned resource profiles with
positive CPU and memory request/limit pairs.

- [x] **Step 4: Regenerate mirrors and verify GREEN**

Run:

```bash
pnpm proto:gen-ts
pnpm proto:gen-go-all
go test ./backend/internal/domain/workerruntime ./proto/gen/go/pod/v1 -count=1
```

- [x] **Step 5: Commit**

```bash
git add backend/internal/domain/workerruntime backend/internal/domain/workerspec \
  backend/internal/domain/agentpod/pod_proto_contract_test.go \
  proto/pod/v1 proto/gen/ts/pod/v1 clients/core/crates/proto-gen/src/domains.rs
git commit -m "feat(worker): define creation wire contract"
git push origin main
```

### Task 2: Scoped Resolution, Preflight, and Snapshot Persistence

**Files:**
- Create: `backend/internal/service/workercreation/contracts.go`
- Create: `backend/internal/service/workercreation/runtime_catalog.go`
- Create: `backend/internal/service/workercreation/worker_type.go`
- Create: `backend/internal/service/workercreation/workspace.go`
- Create: `backend/internal/service/workercreation/compiler.go`
- Create: `backend/internal/service/workercreation/preflight.go`
- Create: `backend/internal/service/workercreation/service_test.go`
- Create: `backend/migrations/000199_worker_spec_links.up.sql`
- Create: `backend/migrations/000199_worker_spec_links.down.sql`
- Modify: `backend/internal/domain/agentpod/pod.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_types.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_create.go`
- Modify: `backend/internal/service/agentpod/pod_service.go`
- Modify: `backend/internal/infra/worker_spec_snapshot_repo.go`
- Modify: `backend/cmd/server/main_startup.go`

- [x] **Step 1: Write failing resolution and persistence tests**

Cover exact model binding, Worker-type hash, image digest, target kind, profile
resources, repository branch, Skill IDs, knowledge IDs, environment bundle IDs,
automation, lifecycle, and alias. Reject cross-organization references,
unsupported deployment, unknown automation, stale options, and missing
dependencies.

```go
result, err := service.Preflight(ctx, scope, draft)
require.NoError(t, err)
require.Empty(t, result.BlockingErrors)
assert.Equal(t, int64(42), result.Resolved.Spec.Runtime.ModelBinding.ResourceID)
```

Create must persist one immutable snapshot and set
`pods.worker_spec_snapshot_id` before dispatch.

- [x] **Step 2: Verify RED**

```bash
go test ./backend/internal/service/workercreation \
  ./backend/internal/service/agentpod ./backend/migrations -run 'WorkerSpec|WorkerCreation' -count=1
```

- [x] **Step 3: Implement scoped resolvers and compiler**

Resolve every external ID through its organization/user-scoped service. Convert
the existing Agent config schema into WorkerSpec type schema, compute a
canonical SHA-256 definition hash, and compile the resolved spec into the
existing AgentFile DSL using `agentfile/serialize`. Secret values must never be
written into the draft or snapshot. Model-resource-managed config and
credential fields must not appear in the Worker type schema or runtime
EnvBundles.

- [x] **Step 4: Link snapshots to Pods**

Migration:

```sql
ALTER TABLE pods
  ADD COLUMN worker_spec_snapshot_id BIGINT
  REFERENCES worker_spec_snapshots(id);
CREATE INDEX idx_pods_worker_spec_snapshot_id
  ON pods(worker_spec_snapshot_id)
  WHERE worker_spec_snapshot_id IS NOT NULL;
```

Fresh structured creates require the WorkerSpec service. Resume inherits the
source snapshot ID and does not re-resolve mutable references.

- [x] **Step 5: Verify GREEN and commit**

```bash
go test ./backend/internal/service/workercreation ./backend/internal/service/agentpod \
  ./backend/internal/infra ./backend/migrations -count=1
git diff --check
git commit -m "feat(worker): persist resolved workerspec on create"
git push origin main
```

### Task 3: Connect, Rust Core, and Web API Boundary

**Files:**
- Modify: `backend/internal/api/connect/pod/server.go`
- Modify: `backend/internal/api/connect/pod/mount.go`
- Modify: `backend/internal/api/connect/pod/mutations.go`
- Create: `backend/internal/api/connect/pod/worker_creation.go`
- Test: `backend/internal/api/connect/pod/worker_creation_test.go`
- Modify: `clients/core/crates/api-client/src/modules/pod.rs`
- Modify: `clients/core/crates/services/src/pod.rs`
- Modify: `clients/core/crates/wasm/src/service_pod.rs`
- Modify: `clients/web/src/lib/api/connect/podConnect.ts`
- Modify: `clients/web/src/lib/api/facade/podApi.ts`
- Test: `clients/web/src/lib/api/__tests__/podWorkerCreation.test.ts`

- [ ] **Step 1: Write failing transport tests**

Test protobuf binary round-trip for all draft fields, options response disabled
reasons, preflight blocking errors versus warnings, and Fill with AI patch
response. Missing services return Unavailable; invalid references return
InvalidArgument or PermissionDenied, never an empty success response.

- [ ] **Step 2: Verify RED**

```bash
go test ./backend/internal/api/connect/pod -run WorkerCreation -count=1
pnpm exec vitest run --config clients/web/vitest.config.ts \
  clients/web/src/lib/api/__tests__/podWorkerCreation.test.ts
```

- [ ] **Step 3: Implement Connect and client bridge**

Add the new unary handlers to the existing Pod service. Decode and encode only
protobuf bytes across Rust/WASM. TypeScript exposes snake_case business shapes
and throws on transport or decoding failure.

- [ ] **Step 4: Verify GREEN and commit**

```bash
go test ./backend/internal/api/connect/pod ./backend/cmd/server -count=1
pnpm exec vitest run --config clients/web/vitest.config.ts \
  clients/web/src/lib/api/__tests__/podWorkerCreation.test.ts
git commit -m "feat(worker): expose creation preflight APIs"
git push origin main
```

### Task 4: One Four-Step Worker Creation State Machine

**Files:**
- Create: `clients/web/src/components/pod/hooks/workerCreateDraft.ts`
- Create: `clients/web/src/components/pod/hooks/useWorkerCreateDraft.ts`
- Create: `clients/web/src/components/pod/hooks/useWorkerCreateOptions.ts`
- Create: `clients/web/src/components/pod/CreatePodForm/WorkerRuntimeStep.tsx`
- Create: `clients/web/src/components/pod/CreatePodForm/WorkerTypeConfigStep.tsx`
- Create: `clients/web/src/components/pod/CreatePodForm/WorkerWorkspaceStep.tsx`
- Create: `clients/web/src/components/pod/CreatePodForm/WorkerPreflightStep.tsx`
- Modify: `clients/web/src/components/pod/CreatePodForm/WorkerCreateStepper.tsx`
- Modify: `clients/web/src/components/pod/CreatePodForm/CreatePodFormFields.tsx`
- Modify: `clients/web/src/components/pod/CreatePodForm/index.tsx`
- Modify: `clients/web/src/components/workers/CreateWorkerPageContent.tsx`
- Modify: `clients/web/src/components/workers/NlWorkerCreate.tsx`
- Modify locale files under `clients/web/src/messages/*`
- Test: `clients/web/src/components/pod/CreatePodForm/__tests__/WorkerCreateFlow.test.tsx`

- [ ] **Step 1: Write failing reducer and UI tests**

Cover:

- four named steps and gated navigation;
- model, Worker type, image, target, deployment, and profile order;
- disabled incompatible options with concrete reasons;
- load failures rendered as errors;
- type switch confirmation and incompatible-value clearing;
- lifecycle request parity;
- preflight blocking errors and warnings;
- one `Create Worker` action;
- `Fill with AI` patches the same reducer without submitting or remounting.

- [ ] **Step 2: Verify RED**

```bash
pnpm exec vitest run --config clients/web/vitest.config.ts \
  clients/web/src/components/pod/CreatePodForm/__tests__/WorkerCreateFlow.test.tsx
```

- [ ] **Step 3: Implement reducer-owned workflow**

Replace local field `useState` ownership with one reducer. Required async
sections use `idle | loading | ready | error` states. Remove the separate quick
task submit path and source/form mode toggle; raw AgentFile remains an advanced
panel inside step 2 and must pass server preflight.

- [ ] **Step 4: Verify GREEN, affected lint/typecheck, and commit**

```bash
pnpm exec vitest run --config clients/web/vitest.config.ts \
  clients/web/src/components/pod/CreatePodForm/__tests__
pnpm exec eslint --config clients/web/eslint.config.mjs \
  clients/web/src/components/pod/CreatePodForm clients/web/src/components/workers
pnpm exec tsc --noEmit -p clients/web/tsconfig.json
git commit -m "feat(worker): unify four-step creation workflow"
git push origin main
```

### Task 5: Snapshot-Backed Expert Publishing

**Files:**
- Modify: `backend/internal/domain/expert/expert.go`
- Modify: `backend/internal/service/expert/service.go`
- Modify: `backend/internal/service/expert/publish.go`
- Modify: `backend/internal/service/expert/run.go`
- Modify: `backend/internal/api/rest/v1/expert_handler.go`
- Modify: `backend/internal/api/rest/v1/expert_handler_types.go`
- Modify: `backend/migrations/000199_worker_spec_links.up.sql`
- Modify: `backend/migrations/000199_worker_spec_links.down.sql`
- Modify: `clients/web/src/components/experts/PublishExpertDialog.tsx`
- Test: `backend/internal/service/expert/publish_test.go`

- [ ] **Step 1: Write failing snapshot publication tests**

Publishing must copy the Pod snapshot ID, reject missing/cross-org snapshots,
ignore client-supplied runtime configuration, and reproduce the source spec on
run except for alias/prompt overrides.

- [ ] **Step 2: Verify RED**

```bash
go test ./backend/internal/service/expert -run 'Publish|WorkerSpec' -count=1
```

- [ ] **Step 3: Implement fail-closed snapshot publication**

Add `experts.worker_spec_snapshot_id`. New publication accepts identity fields
only. Legacy Experts with no snapshot remain readable but return a typed
republish-required error when run; do not reconstruct a WorkerSpec from
incomplete legacy columns.

- [ ] **Step 4: Verify GREEN and commit**

```bash
go test ./backend/internal/service/expert ./backend/internal/api/rest/v1 -run Expert -count=1
git commit -m "feat(expert): publish from workerspec snapshots"
git push origin main
```

### Task 6: Selected Sandbox Skill Publishing

**Files:**
- Create: `runner/internal/runner/worker_skill_scan.go`
- Create: `runner/internal/runner/worker_skill_scan_test.go`
- Modify: `runner/internal/runner/sandbox_fs_handler.go`
- Create: `backend/internal/service/skill/worker_publish.go`
- Create: `backend/internal/service/skill/worker_publish_test.go`
- Create: `backend/internal/api/rest/v1/worker_skill_handler.go`
- Modify: `backend/internal/api/rest/v1/routes_skill.go`
- Create: `clients/web/src/components/workers/PublishSkillDialog.tsx`
- Modify active Worker header component selected by code search

- [ ] **Step 1: Write failing Runner security tests**

Reject absolute paths, `..`, roots outside `skills`, `.agents/skills`, and
`.codex/skills`, escaping symlinks, invalid/missing frontmatter, invalid
identifiers, packages over limits, and hash changes between discovery and
publish.

- [ ] **Step 2: Verify RED**

```bash
go test ./runner/internal/runner -run WorkerSkill -count=1
go test ./backend/internal/service/skill -run WorkerPublish -count=1
```

- [ ] **Step 3: Implement discovery and selected publication**

Use `SandboxFsCommand` operations `skill_discover` and `skill_package`.
Discovery returns relative path, slug, name, summary, hash, validation state,
and new/modified/published status. Publication revalidates the selected path and
hash on Runner, then the Skill service provisions or updates the Git-backed
catalog row from the packaged files.

- [ ] **Step 4: Verify GREEN and commit**

```bash
go test ./runner/internal/runner ./backend/internal/service/skill \
  ./backend/internal/api/rest/v1 -run 'WorkerSkill|PublishSkill' -count=1
git commit -m "feat(skill): publish selected worker skills"
git push origin main
```

### Task 7: Browser Acceptance and Final Integration

**Files:**
- Create or modify focused Playwright specs under `clients/web/e2e-playwright/tests/pod/`
- Modify: `docs/superpowers/plans/2026-07-10-worker-platform-redesign-progress.md`
- Modify user-visible Worker/Expert/Skill documentation affected by the flow

- [ ] **Step 1: Run full affected regression**

```bash
go test ./backend/internal/domain/workerspec ./backend/internal/domain/workerruntime \
  ./backend/internal/service/workerspec ./backend/internal/service/workercreation \
  ./backend/internal/service/agentpod ./backend/internal/service/expert \
  ./backend/internal/service/skill ./backend/internal/api/connect/pod \
  ./backend/internal/api/rest/v1 ./backend/cmd/server ./runner/internal/runner -count=1
pnpm proto:gen-ts
pnpm proto:gen-go-all
pnpm run web:test
pnpm run web:lint
pnpm run web:typecheck
git diff --check
```

- [ ] **Step 2: Start the real development services**

Use the repository's current non-Bazel dev launcher. Confirm backend health,
frontend readiness, Runner connectivity, migrations, and test account login.

- [ ] **Step 3: Execute desktop and mobile browser scenarios**

Verify:

1. successful pooled Worker create;
2. incompatible and failed-loading states;
3. Fill with AI updates but does not submit;
4. preflight blocking and warning rendering;
5. Expert publish and subsequent run preserve WorkerSpec;
6. candidate discovery, preview, selected Skill publish, and invalid-path rejection;
7. console and network contain no relevant errors.

- [ ] **Step 4: Final review, update durable progress, and push**

Review staged files against the approved spec and the shared dirty worktree.
Confirm no unrelated files, no fallback paths, no weakened tests, and no new
production file over 200 lines.

```bash
git status --short
git diff --cached --name-only
git push origin main
git fetch origin main
git branch -r --contains HEAD
git show --no-patch --format=fuller HEAD
```
