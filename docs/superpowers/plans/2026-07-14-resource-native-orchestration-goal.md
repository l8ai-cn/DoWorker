# Resource-Native Orchestration Goal

## Objective

Deliver a resource-native orchestration module for AgentsMesh that keeps
`WorkerSpecSnapshot` as the runtime SSOT while adding:

- a typed resource envelope and immutable `ResourceRef` resolution;
- persisted validate, plan, and target-specific apply operations;
- versioned Worker, Expert, Workflow, and GoalLoop integration;
- referenceable model, prompt, Skill, knowledge, tool, environment, repository,
  compute target, and resource profile resources;
- one frontend draft shared by domain forms and an advanced YAML view;
- semantic diff, revision history, status, permissions, and secret-reference
  protection.

The delivery target is merge-ready code with deterministic backend, protocol,
Rust Core, frontend, migration, and browser evidence. Production deployment and
destructive schema cleanup require separate approval.

## Architecture Boundary

The approved architecture is:

```text
domain form / YAML / API
          |
          v
typed Resource Draft
          |
          v
validate -> plan -> target apply
          |
          v
WorkerSpecSnapshot / domain revision
          |
          v
WorkerRunManifest -> Pod
```

The resource envelope is an authoring and control-plane contract. It does not
replace `WorkerSpecSnapshot`, introduce local state, or create a second runtime
path. Commands such as run, trigger, pause, resume, publish, archive, and
terminate remain typed domain actions.

## Delivery Phases

| Phase | Deliverable | Machine-checkable exit |
| --- | --- | --- |
| 0 | Repair current WorkerSpec contracts and execution inventory | Focused WorkerSpec, Expert, Workflow, GoalLoop, Mesh, and Pod tests pass |
| 1 | Resource envelope, references, schema registry, YAML codec | Domain and codec tests reject unknown fields, invalid identifiers, mutable status, and secret values |
| 2 | Persisted validate/plan/apply vertical slice for Worker | Concurrent apply, stale plan, hash, tenancy, idempotency, snapshot, and launch tests pass |
| 3 | Immutable definitions, dependencies, and run manifests | Historical snapshot and exact revision tests pass after current catalogs change |
| 4 | Expert, Workflow, GoalLoop, Mesh, and remaining entry points | Every fresh execution pins a snapshot and no legacy runtime builder is reachable |
| 5 | Domain forms plus YAML, references, diff, status, and history | Vitest and Playwright cover form/YAML round trip and all blocking states |
| 6 | Migration, rollout, documentation, and independent review | Migration rehearsal, cross-stack tests, browser evidence, docs checks, and review pass |

Detailed implementation plans live in separate phase documents. A phase plan
must exist before code for that phase is edited.

## Loop Control

### Trigger

The active Codex goal drives one bounded implementation task at a time. Each
cycle starts by reading this file, the current phase plan, Git status, and the
latest verification evidence.

### Observation

Each cycle records:

- current phase and checked plan item;
- relevant Git diff;
- exact verifier command and exit status;
- blocker or next failing assertion;
- whether the verification fingerprint changed.

### Action

One cycle performs one narrow change set with one deterministic verifier.
Independent read-only audits or disjoint write sets may run in parallel.

### Success Exit

The goal succeeds only when all phase exit gates pass, the fresh execution
inventory contains no legacy runtime inputs, browser evidence covers the primary
and blocking paths, and an independent final review has no unresolved P0/P1
findings.

### Failure And Budget Exits

- Maximum 32 implementation cycles before scope review.
- Maximum 45 minutes of active work per cycle.
- Maximum two concurrent writing agents with disjoint file ownership.
- Stop after two consecutive cycles with the same failing verifier fingerprint
  and no relevant diff.
- Stop after three failed attempts to repair the same root cause.
- Never weaken tests, validators, migration checks, CI, permission checks, or
  secret redaction to make a gate pass.

### Human Escalation

Stop and request a decision before:

- deleting legacy columns or public API fields;
- changing tenant, authorization, credential, or secret ownership semantics;
- selecting a migration result when immutable historical evidence is missing;
- deploying, pushing, merging, or applying production migrations;
- editing a file that has overlapping uncommitted changes whose intent cannot be
  established from code and tests.

## No-Progress Fingerprint

The fingerprint is:

```text
phase + plan item + verifier command + first stable error code/assertion
```

A changed error caused only by nondeterministic text, timestamps, ports, or IDs
does not count as progress.

## Durable Progress

| Phase | Status | Evidence |
| --- | --- | --- |
| 0 | complete | Execution inventory remains machine-checked and currently records 16 constructors: 6 legacy, 2 plan, 5 snapshot, and 3 lineage |
| 1 | complete | Strict JSON/YAML codecs, schema registry, ResourceRef and SecretReference tests pass |
| 2 | complete | WorkerTemplate and Worker use persisted plans, immutable revisions, WorkerSpec snapshots, durable launch records, stale-plan checks, and typed Apply |
| 3 | in progress | Prompt, Expert, Workflow, Worker, and GoalLoop pin resource revisions and WorkerSpec snapshots; resolved non-Secret model, repository, Skill, knowledge, environment, and placement facts still require a versioned dependency snapshot |
| 4 | in progress | Expert/Workflow definition mutation guards, Workflow/GoalLoop create gates, full WorkflowRun manifests, snapshot-backed Mesh launch, Plan-only Quick Task/Runner MCP, durable Session create/fork/import ownership, and Coordinator atomic task claims are complete; Session switch replacement ownership and 6 legacy snapshot-binding constructors remain |
| 5 | in progress | Draft concurrency, YAML gates, typed/redacted errors, Plan retirement, partial reference loading, mobile actions, exact numeric handling, Workflow create/new-revision editors, and locked identity have focused tests; Playwright execution remains frozen |
| 6 | in progress | Focused backend, migration-static, documentation, and independent subtask reviews pass; PostgreSQL rehearsal, browser evidence, and final whole-goal review remain |

## Delivery Audit

Audit date: 2026-07-17.

| Requirement | Status | Authoritative evidence |
| --- | --- | --- |
| Unified Resource envelope and ResourceRef | proved | `orchestrationresource.Manifest`, draft/resolved `Reference` validation, protocol Resource identity |
| Validate and persisted Plan | proved | control service validation/planning tests and Connect handlers |
| Binding typed Apply | proved | eight registered binding kinds, binding Apply service, protocol and web dispatch |
| Prompt typed Apply | proved | Prompt Apply service, immutable revision tests, protocol and web dispatch |
| WorkerTemplate typed Apply | proved | snapshot compiler, Apply service, revision/snapshot PostgreSQL tests |
| Worker typed Apply | proved | durable launch claim/completion, one-shot constraint, Pod result contract |
| Expert typed Apply | proved | Expert projection stores resource revision and WorkerSpec snapshot; old Update/Delete reject resource-managed definitions before store or Git mutation |
| Workflow typed Apply | proved | Workflow projection and runs pin resource revision, WorkerSpec snapshot, trigger params, resolved prompt, and full execution manifest; runtime completion and timeout/idle scans do not read the latest Workflow; old AgentFile runtime builder is removed |
| GoalLoop typed Apply | proved | `CreateGoalLoopFromPlan` creates a pinned draft without a Pod; public legacy Create and RunLoopProgram are gated while Start/Verify/Cancel remain explicit actions |
| One form/YAML Draft | proved | YAML/form transition is atomic; parsed YAML is locally gated; stale responses cannot replace current state; expired or terminal Apply Plans cannot be replayed |
| Semantic Diff | proved | server digest-only changes and dedicated redaction rendering tests |
| Permission revalidation | in verification | target and resolved references are reauthorized; Apply transaction locks current membership/role and replay rechecks it. Unit tests pass; PostgreSQL concurrency tests are authored but frozen |
| Secret reference protection | proved | reference-only schemas, non-echoing errors, digest/redacted Diff, reference-only UI test |
| Resolved dependency immutability | Plan builder proved; persistence frozen | Strict Artifact V1 and `WorkerTemplateBuild` codecs bind canonical WorkerSpec, exact Plan refs, typed ToolBinding model provenance, model/repository/content/image facts, ordered bundles, field ownership, and reference-only Secrets. Table and artifact-only runtime remain frozen |
| Migration integrity | in verification | Formal release mainline is through 000224; local `000221_worker_spec_optional_model_binding` conflicts with formal `000221_add_expert_revision`. Sequence may restart at candidate 000225 only after owner ordering; DB execution is frozen |
| Public documentation | in verification | Product/API/migration/YAML/Expert/Workflow manuals describe Apply-only definitions, explicit actions, snapshot consistency, Worker Definition credential/config-document bindings, REST lineage-only resume, Plan-only Quick Task and Runner MCP, and legacy entry-point errors; focused locale and docs checks pass |

## Verification Snapshot

The durable command-level and review evidence is maintained in
[`2026-07-14-resource-native-orchestration-verification.md`](./2026-07-14-resource-native-orchestration-verification.md).
This goal file remains the control contract and current status SSOT.

## Current Constraints
- The current worktree contains extensive unrelated uncommitted work.
- Implementation must not revert or reformat those changes.
- Files already modified by another task require diff review before editing.
- Mainline is at `000224`; the dirty local view through `000221` confers no
  ownership of later numbers. Have the owner order work from candidate `000225`
  before file creation. Do not add, renumber, run, or deploy while frozen.
- Do not modify Runner or GoalLoop during the Loop/Seedance production release.
- Runtime code must never fall back to legacy Agent, model, repository,
  AgentFile, or mutable Expert/Workflow fields.
- Runtime code must not fall back from a missing resolved dependency artifact
  to current model, repository, Skill, knowledge, environment, compute, or
  resource-profile rows. Secret values remain live references and are never
  materialized into snapshot JSON.
- Plan `optionsRevision` is an opaque bounded string so target planners can
  bind native catalog or policy revision tokens without lossy numeric mapping.

## Source Designs

- `docs/superpowers/specs/2026-07-13-unified-worker-orchestration-design.md`
- `docs/superpowers/specs/unified-worker-orchestration/`
- `docs/superpowers/plans/2026-07-14-resource-native-phase-0.md`
- `docs/superpowers/plans/2026-07-14-resource-native-phase-1a-contract.md`
- `docs/superpowers/plans/2026-07-14-resource-native-phase-1b-codec-registry.md`
- `docs/superpowers/plans/2026-07-14-resource-native-phase-2a-control-plane.md`
- `docs/product/resource-native-orchestration.md`
- `docs/product/resource-yaml-manual.md`
