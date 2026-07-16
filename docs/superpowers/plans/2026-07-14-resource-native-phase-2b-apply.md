# Resource-Native Phase 2B Apply Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> `subagent-driven-development` task-by-task.

**Goal:** Consume actor-bound Plans to atomically persist binding resources,
WorkerTemplate revisions, and immutable WorkerSpec snapshots.

**Architecture:** Apply remains target-specific. `orchestrationcontrol`
continues to own plan and revision invariants; `orchestrationworker` builds
binding or WorkerTemplate mutations; the PostgreSQL repository persists the
target artifact, resource head, revision, and plan consumption in one
transaction. Runtime code continues to load only `WorkerSpecSnapshot`.

**Tech Stack:** Go, GORM, PostgreSQL, Connect RPC, protobuf, Rust Core/WASM.

---

## Invariants

- The authenticated organization and actor must equal the stored Plan scope.
- Apply accepts only a pending, unexpired Plan ID; it never accepts a manifest,
  mutable draft, artifact, or legacy Worker fields.
- Binding Apply accepts only the eight registered binding Kinds.
- WorkerTemplate Apply accepts only `WorkerTemplate -> WorkerSpec`.
- The stored revision must match the exact planned manifest, Spec, resolved
  references, artifact digest, and target identity.
- A WorkerTemplate revision and its WorkerSpec snapshot are committed or rolled
  back together.
- A replay, stale base, expired Plan, or concurrent create fails before a second
  revision or snapshot becomes visible.
- Metadata-only updates increment revision and resource version but not
  generation; Spec changes increment all three.
- No target Apply path invokes Worker creation, Pod launch, or legacy builders.

## BDD Scenarios

```text
Given a valid binding Plan
When ApplyBindingResourcePlan consumes it
Then one head and one immutable revision are stored and the Plan is applied

Given a valid WorkerTemplate Plan with a canonical WorkerSpec artifact
When ApplyWorkerTemplatePlan consumes it
Then the WorkerSpec snapshot, resource revision, and Plan transition commit atomically

Given snapshot persistence fails
When WorkerTemplate Apply runs
Then no resource, revision, snapshot, or Plan consumption remains

Given an update Plan whose Spec is unchanged
When it is applied
Then revision and resourceVersion increment while generation remains unchanged

Given two callers consume the same Plan
When both Apply concurrently
Then exactly one succeeds and the other returns consumed
```

### Task 1: Reconstruct A Trusted Snapshot From A Planned Artifact

**Files:**
- Modify: `backend/internal/service/workerspec/resolved_snapshot.go`
- Test: `backend/internal/service/workerspec/resolved_snapshot_test.go`

- [ ] Add `NewResolvedSnapshot(organizationID int64, spec workerspec.Spec)` as
  the only exported constructor for an already-resolved canonical Spec.
- [ ] Normalize and validate the Spec, derive its Summary, and emit canonical
  Spec and Summary JSON through the existing private `resolveSnapshot` path.
- [ ] Reject zero organization IDs, unsupported versions, invalid bindings,
  malformed placement, invalid lifecycle, and non-canonical input values.
- [ ] Prove the constructor returns detached byte slices and repository
  round-trips through `SnapshotRepository.Create`.
- [ ] Run:

```bash
go test ./backend/internal/service/workerspec ./backend/internal/infra \
  -run 'ResolvedSnapshot|WorkerSpecSnapshotRepository' -count=1
```

### Task 2: Add A Transaction-Aware Target Apply Boundary

**Files:**
- Create: `backend/internal/service/orchestrationworker/apply_contracts.go`
- Modify: `backend/internal/infra/orchestration_resource_apply.go`
- Create: `backend/internal/infra/orchestration_worker_template_apply.go`
- Create: `backend/internal/infra/orchestration_binding_apply.go`
- Test: matching focused PostgreSQL tests

- [ ] Define `BindingApplyRepository` and `WorkerTemplateApplyRepository`
  interfaces in `orchestrationworker`; neither exposes `*gorm.DB`.
- [ ] Define builders that receive the locked `LockedApplyState`; the
  WorkerTemplate builder also receives the snapshot ID created inside the
  transaction.
- [ ] Refactor the existing repository transaction into one private helper that
  passes its transaction to target-specific persistence while preserving the
  public generic `RunApplyTransaction` behavior and tests.
- [ ] Binding Apply must reject WorkerTemplate and unknown Kinds before calling
  the service builder.
- [ ] WorkerTemplate Apply must strictly decode `ArtifactJSON` as WorkerSpec,
  require byte-for-byte canonical JSON, create the snapshot with the transaction,
  then call the service builder with the persisted snapshot ID.
- [ ] Verify rollback after snapshot insert, replay rejection, concurrent
  consumption, expired Plan, stale base, and exact actor/tenant scope.
- [ ] Run:

```bash
AGENTSMESH_TEST_POSTGRES_DSN="$AGENTSMESH_TEST_POSTGRES_DSN" \
go test ./backend/internal/infra \
  -run 'Orchestration.*Apply|WorkerTemplateApply|BindingApply' -count=1
```

### Task 3: Build Binding And WorkerTemplate Revisions

**Files:**
- Create: `backend/internal/service/orchestrationworker/apply_manifest.go`
- Create: `backend/internal/service/orchestrationworker/binding_apply.go`
- Create: `backend/internal/service/orchestrationworker/worker_template_apply.go`
- Test: matching `_test.go` files

- [ ] Strictly decode the planned authoring manifest through the registered
  Registry and reject stored unknown fields as corruption.
- [ ] Build server-owned metadata from the locked state: UID, generation,
  resource version, timestamps, and canonical empty status.
- [ ] Compute revision and generation from the current revision's canonical
  Spec; never infer a Spec change from display metadata or labels.
- [ ] Copy resolved references from the Plan and preserve their exact revisions
  and digests.
- [ ] Binding revisions require `WorkerSpecSnapshotID == 0`.
- [ ] WorkerTemplate revisions require a positive snapshot ID and
  `ArtifactKind == "WorkerSpec"`.
- [ ] Return the applied head plus snapshot ID without returning the artifact
  JSON or raw canonical manifest.
- [ ] Run:

```bash
go test ./backend/internal/service/orchestrationworker -run Apply -count=1
go test -race ./backend/internal/service/orchestrationworker -run Apply -count=1
```

### Task 4: Expose Typed Apply RPCs

**Files:**
- Modify: `proto/orchestration_resource/v1/orchestration_resource.proto`
- Modify: generated Go, TypeScript, and Rust protobuf artifacts
- Modify: `backend/internal/api/connect/orchestration_resource/`
- Modify: Rust API client, service, WASM wrapper, and web adapter files created
  by Phase 2A
- Test: Connect and Rust adapter tests

- [ ] Add `ApplyBindingResourcePlan` and `ApplyWorkerTemplatePlan`; both requests
  contain only `org_slug` and `plan_id`.
- [ ] Return the applied `Resource`; WorkerTemplate also returns
  `worker_spec_snapshot_id`.
- [ ] Resolve tenant scope through the existing interceptor and never trust the
  request slug as an organization ID.
- [ ] Map expired, consumed, stale, and conflict errors to `Aborted`; keep
  corruption internal.
- [ ] Ensure Apply responses and errors cannot expose artifact JSON, canonical
  manifest, raw YAML, Secret values, SQL, or credentials.
- [ ] Mount the typed handlers beside the six Phase 2A methods and route the web
  adapter through Rust Core/WASM.
- [ ] Run:

```bash
pnpm proto:gen-go-all
pnpm proto:gen-ts
go test ./backend/internal/api/connect/orchestration_resource -count=1
cd clients/core && cargo test --workspace
cd ../../ && pnpm run build:wasm
```

### Task 5: Startup Wiring And End-To-End Transaction Gate

**Files:**
- Modify: `backend/cmd/server/orchestration_control_init.go`
- Modify: `backend/cmd/server/services_container.go`
- Modify: `backend/cmd/server/connect_init.go`
- Test: `backend/cmd/server/orchestration_control_init_test.go`
- Test: Connect handler integration tests

- [ ] Construct binding and WorkerTemplate Apply services from the same
  repository, Registry, and organization authorizer used for Plan.
- [ ] Fail startup when any planner, schema, Apply repository, snapshot
  constructor, or handler dependency is missing.
- [ ] Verify an authenticated create flow:
  `Validate -> Plan -> Apply -> Get -> Export`.
- [ ] Verify a WorkerTemplate Apply stores a positive
  `worker_spec_snapshot_id` on the exact resource revision.
- [ ] Verify a failed Apply leaves the Plan pending and no snapshot visible.
- [ ] Run:

```bash
go test ./backend/cmd/server \
  ./backend/internal/api/connect/orchestration_resource \
  ./backend/internal/service/orchestrationworker \
  ./backend/internal/infra -count=1
go vet ./backend/cmd/server \
  ./backend/internal/api/connect/orchestration_resource \
  ./backend/internal/service/orchestrationworker
```

### Task 6: Phase Gate

- [ ] Run focused PostgreSQL, race, vet, proto, Rust, WASM, TypeScript, and diff
  checks.
- [ ] Update `docs/product/resource-native-orchestration.md` and
  `docs/product/resource-yaml-manual.md` with the exact Plan expiry, typed Apply,
  revision, conflict recovery, and snapshot behavior.
- [ ] Record exact commands and results in
  `docs/superpowers/plans/2026-07-14-resource-native-orchestration-goal.md`.
- [ ] Obtain independent spec and code-quality reviews with no unresolved
  P0/P1 findings.
