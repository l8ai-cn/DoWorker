# Current State and Decisions

## Existing Foundation

The repository already contains most required primitives:

- `WorkerDefinition` bundles load from `config/worker-types`.
- `WorkerSpec` has normalized runtime, placement, type configuration,
  workspace, lifecycle, and metadata sections.
- `worker_spec_snapshots` are protected from updates by a database trigger.
- structured Pod creation rejects simultaneous legacy Worker fields.
- GoalLoop already requires a WorkerSpec snapshot.
- Expert publishing from a Pod already verifies the Pod snapshot.
- AgentFile has a typed parser, serializer, evaluator, and WorkerSpec compiler.

The design extends the existing WorkerSpec foundation. It does not introduce
another configuration language or orchestration runtime.

## Confirmed Conflicts

### Expert Split-Brain

Expert REST and Web creation still accept mutable Agent, repository, prompt,
Skill, knowledge, environment, and config fields. The service can persist an
Expert without `worker_spec_snapshot_id`, while Expert execution rejects that
record.

Snapshot-backed Expert execution ignores later changes to those legacy fields.
The UI can therefore report a successful runtime edit that does not affect the
next run.

Evidence:

- `backend/internal/domain/expert/expert.go`
- `backend/internal/service/expert/crud.go`
- `backend/internal/service/expert/run.go`
- `backend/internal/api/rest/v1/expert_handler_types.go`
- `clients/web/src/components/experts`

### Workflow Is a Parallel Runtime Specification

Workflow stores Agent, model, Runner, repository, branch, environment bundle,
permission, and raw config fields. It reconstructs AgentFile at execution time
and sends legacy Pod fields.

The builder ignores malformed config JSON, unavailable repositories, and
missing environment bundles in some paths. This violates explicit failure
semantics and makes Workflow a second runtime SSOT.

Evidence:

- `backend/internal/domain/workflow/workflow.go`
- `backend/internal/service/workflow/workflow_orchestrator_agentfile.go`
- `backend/internal/service/workflow/workflow_pod_request.go`

### Historical Snapshot Replay Depends on Current Definitions

Snapshot preparation recompiles against current dependencies and validates the
snapshot Worker type against the current WorkerDefinition. Updating the
definition therefore makes a historical snapshot unrunnable even though the
snapshot is labeled immutable.

Evidence:

- `backend/internal/service/workercreation/snapshot.go`
- `backend/internal/service/workercreation/worker_type_snapshot.go`
- `backend/internal/service/agentpod/pod_orchestrator_worker_spec.go`

### WorkerSpec Contract Migration Is Incomplete

`ModelBinding.ProtocolAdapter` is now required by WorkerSpec validation, but at
least the Expert publish fixture still constructs an incomplete binding. The
focused Expert test currently fails with an empty protocol adapter slug.

Evidence:

- `backend/internal/domain/workerspec/runtime.go`
- `backend/internal/domain/workerspec/validation.go`
- `backend/internal/service/expert/publish_test.go`

### Mesh Ticket Bypasses WorkerSpec

Mesh Ticket Pod creation hardcodes a Worker type, model default, permission
mode, and AgentFile layer.

Evidence:

- `backend/internal/domain/mesh/legacy_ticket_pod.go`
- `backend/internal/service/mesh/ticket_pod_orchestration.go`

### WorkerDefinition Projection Has Runtime Authority

File-backed WorkerDefinition is canonical, but runtime resolution also requires
an exact matching Agent row. A missing or stale projection is discovered on a
user request instead of at service health validation.

Evidence:

- `backend/internal/service/workerdefinition/projection_sync.go`
- `backend/internal/service/workercreation/worker_type.go`
- `backend/internal/service/workercreation/worker_type_projection.go`

## Considered Approaches

### Approach A: Add HCL and Keep Existing Runtime Paths

HCL would generate current Expert, Workflow, and Pod fields while legacy APIs
remain writable.

Rejected because it adds a third representation and leaves all current
split-brain behavior intact.

### Approach B: Keep Mutable Domain Fields and Recompile on Every Run

Expert and Workflow remain editable aggregates. Every run resolves current
resources and rebuilds AgentFile.

Rejected because historical behavior is not reproducible, edits race with
scheduled runs, and deleted or renamed resources cause implicit drift.

### Approach C: Immutable WorkerSpec and Domain Revisions

All reusable products pin a WorkerSpec snapshot. Configuration changes create
new revisions. Every Pod receives a materialized run manifest.

Selected because it matches the existing WorkerSpec work, provides auditable
runtime identity, and removes duplicate configuration logic.

## Explicit Product Decisions

- WorkerSpec describes execution capability, not a reusable task schedule.
- Expert is a versioned business capability whose revision pins WorkerSpec.
- Workflow is a versioned reusable task whose revision pins WorkerSpec.
- GoalLoop is a single goal execution and pins WorkerSpec directly.
- Mesh Ticket pins a resolved WorkerSpec at assignment time.
- Prompt templates belong to Workflow revisions.
- Durable Worker instructions belong to WorkerSpec.
- Per-run task input belongs to invocation context, not WorkerSpec identity.
- Alias and display metadata may be overridden without changing runtime
  semantics.
- Raw AgentFile editing is not a supported primary configuration path.

## Compatibility Boundary

Migration tooling may read legacy fields to produce an explicit report and a
candidate WorkerSpec draft. Runtime code must never fall back to those fields.

Rows that cannot be migrated safely become `migration_required` and cannot run
until a user reviews a generated draft and applies a valid plan.
