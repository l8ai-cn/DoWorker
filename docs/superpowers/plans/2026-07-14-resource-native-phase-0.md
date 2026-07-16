# Resource-Native Orchestration Phase 0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> `subagent-driven-development` or `executing-plans` to implement this plan
> task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore a trustworthy WorkerSpec baseline and add a deterministic
inventory proving which fresh Pod entry points still bypass snapshots.

**Architecture:** Repair the model binding contract at every current fixture
and persistence boundary, then encode fresh execution entry points as an
explicit test-owned inventory. Do not add compatibility behavior.

**Tech Stack:** Go 1.25, testify, GORM, embedded SQL migrations, Connect-RPC.

---

### Task 1: Complete Protocol Adapter Test Fixtures

**Files:**
- Modify: `backend/internal/service/expert/publish_test.go`
- Modify: `backend/internal/infra/worker_spec_snapshot_repo_contract_test.go`

- [x] **Step 1: Preserve the failing Expert evidence**

Run:

```bash
go test ./backend/internal/service/expert \
  -run TestPublishFromPodBindsWorkerSpecSnapshot -count=1
```

- [x] **Step 2: Add the immutable adapter identity to the Expert fixture**

Add this field beside `ProviderKey`:

```go
ProtocolAdapter: slugkit.MustNewForTest("openai-compatible"),
```

- [x] **Step 3: Add the same identity to the snapshot repository fixture**

Add the same `ProtocolAdapter` field to `workerSpecForRepoContract`.

- [x] **Step 4: Verify focused packages**

Run:

```bash
go test -count=1 \
  ./backend/internal/domain/workerspec \
  ./backend/internal/service/workerspec \
  ./backend/internal/service/workercreation \
  ./backend/internal/service/expert \
  ./backend/internal/infra
```

Expected: PASS.

### Task 2: Verify Migration And Serialization Boundaries

**Files:**
- Modify only when a failing assertion proves a missing adapter:
  `backend/migrations/worker_spec_protocol_adapter_migration_test.go`
- Modify only when a failing assertion proves a missing adapter:
  `backend/migrations/worker_spec_model_binding_postgres_test.go`

- [x] **Step 1: Run adapter migration tests**

```bash
go test ./backend/migrations \
  -run 'TestMigration.*WorkerSpec.*(ProtocolAdapter|ModelBinding)' -count=1
```

- [x] **Step 2: Run WorkerSpec JSON contract tests**

```bash
go test ./backend/internal/domain/workerspec \
  -run 'TestWorkerSpec.*(ModelBinding|Summary|Normalize|Decode)' -count=1
```

### Task 3: Add A Fresh Execution Inventory

**Files:**
- Create: `backend/internal/service/agentpod/fresh_execution_inventory_test.go`

- [x] **Step 1: Write the failing inventory test**

Parse non-test backend Go files and discover every
`OrchestrateCreatePodRequest` composite literal. Key each constructor by source
file and enclosing function.

The expected inventory classifies each constructor as `legacy`, `plan`,
`snapshot`, or `lineage` and records its currently accepted runtime fields.
The test fails on an unknown constructor, a stale expected entry, or a field
set that differs from the reviewed inventory.

```go
[]string{
	"agent_slug",
	"model",
	"repository_id",
	"agentfile_layer",
}
```

Resume and exact fork may use `lineage`. Phase 0 records explicit `legacy`
entries; Phase 6 changes the terminal assertion to require zero legacy entries.

- [x] **Step 2: Run the inventory test**

```bash
go test ./backend/internal/service/agentpod \
  -run TestFreshExecutionInventory -count=1
```

Expected: FAIL listing all unreviewed constructors.

- [x] **Step 3: Implement the minimal explicit registry**

Keep the inventory in the test. Each entry records:

```go
type FreshExecutionEntry struct {
	Source          string
	Mode            string
	AcceptedRuntime []string
}
```

Valid modes are `legacy`, `plan`, `snapshot`, and `lineage`.

- [x] **Step 4: Re-run the inventory test**

Expected: PASS only after every known entry point is represented.

### Task 4: Restore Focused Domain Baseline

**Files:**
- Modify only files with a proved root-cause failure in:
  `backend/internal/service/workflow/`
  `backend/internal/service/goalloop/`
  `backend/internal/service/mesh/`
  `backend/internal/service/agentpod/`

- [x] **Step 1: Run the focused baseline**

```bash
go test -count=1 \
  ./backend/internal/service/expert \
  ./backend/internal/service/workflow \
  ./backend/internal/service/goalloop \
  ./backend/internal/service/mesh \
  ./backend/internal/service/agentpod
```

- [x] **Step 2: Repair only explicit contract failures**

Do not introduce legacy fallback reads. Malformed JSON, missing resources,
stale projections, and cross-organization references remain errors.

- [x] **Step 3: Run the Phase 0 terminal verifier**

```bash
go test -count=1 \
  ./backend/internal/domain/workerspec \
  ./backend/internal/service/workerspec \
  ./backend/internal/service/workercreation \
  ./backend/internal/service/expert \
  ./backend/internal/service/workflow \
  ./backend/internal/service/goalloop \
  ./backend/internal/service/mesh \
  ./backend/internal/service/agentpod \
  ./backend/internal/infra \
  ./backend/migrations
```

Expected: PASS.

### Task 5: Record Phase Evidence

**Files:**
- Modify: `docs/superpowers/plans/2026-07-14-resource-native-orchestration-goal.md`

- [x] **Step 1: Check completed Phase 0 items**

Record exact verifier commands and results. Skipped failures do not pass.

- [x] **Step 2: Create the Phase 1 implementation plan**

The Phase 1 plan must specify exact files and tests for:

- resource envelope and metadata;
- immutable `ResourceRef`;
- kind/schema registry;
- strict JSON and YAML codecs;
- status and server-managed field protection;
- Secret value rejection and reference-only encoding.
