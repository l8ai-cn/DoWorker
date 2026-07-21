# Resource-Native Orchestration Goal

## Objective

Deliver a resource-native orchestration module for Agent Cloud that keeps
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
| 0 | complete | Machine-checked execution inventory has no `legacy` constructor: every entry is Plan, Snapshot, or lineage |
| 1 | complete | Strict JSON/YAML codecs, schema registry, ResourceRef and SecretReference tests pass |
| 2 | complete | WorkerTemplate and Worker use persisted plans, immutable revisions, WorkerSpec snapshots, durable launch records, stale-plan checks, and typed Apply |
| 3 | complete | Every executable WorkerSpec snapshot persists one resolved dependency artifact and runtime materializes its non-Secret inputs only from that artifact |
| 4 | complete | Session, host, fork, import, switch, Coordinator, Expert, Workflow, GoalLoop, Mesh, Quick Task, and MCP create paths bind Plan, Snapshot, or lineage and fail closed on zero source |
| 5 | in verification | Form/YAML one-Draft, semantic Diff, permission and Secret states pass 27 focused Web test files; a healthy browser API proxy is still required for current-run Playwright evidence |
| 6 | in verification | Real PostgreSQL migration rehearsal through 000231, scoped backend checks, Rust protocol fixture, docs checks, and codegen pass; browser/API pairing is the remaining environment gate |

## Delivery Audit

Audit date: 2026-07-19.

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
| Permission revalidation | proved | target and resolved references are reauthorized; Apply locks current membership/role and replay rechecks it |
| Secret reference protection | proved | reference-only schemas, non-echoing errors, digest/redacted Diff, reference-only UI test |
| Resolved dependency immutability | proved | Artifact V1 is persisted atomically with every executable snapshot and supplies all non-Secret runtime inputs; pins and mismatches fail closed |
| Migration integrity | proved | 000219 through 000231 are unique and the PostgreSQL up/down/up lineage rehearsal passes |
| Public documentation | proved | Product/API/migration/YAML/Expert/Workflow manuals describe typed Apply, explicit actions, snapshots, permission, Secret, and lineage behavior |

## Verification Snapshot

The durable command-level and review evidence is maintained in
[`2026-07-14-resource-native-orchestration-verification.md`](./2026-07-14-resource-native-orchestration-verification.md).
This goal file remains the control contract and current status SSOT.

## Current Constraints
- The current worktree contains extensive unrelated uncommitted work.
- Implementation must not revert or reformat those changes.
- Files already modified by another task require diff review before editing.
- Released migrations through `000231` are immutable; future migration numbers
  require release-owner ordering and a real PostgreSQL rehearsal.
- Browser evidence requires one healthy Web/API proxy pair; the stale `12407`
  dev proxy must not be treated as product behavior.
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
