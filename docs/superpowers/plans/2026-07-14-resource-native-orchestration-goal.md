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
| 0 | complete | Full backend verifier passes; inventory records 17 constructors: 13 legacy and 4 snapshot |
| 1 | complete | Strict JSON/YAML codecs, schema registry, ResourceRef and SecretReference tests pass |
| 2 | complete | WorkerTemplate and Worker use persisted plans, immutable revisions, WorkerSpec snapshots, durable launch records, stale-plan checks, and typed Apply |
| 3 | complete | Prompt, Expert, and Workflow Apply persist immutable resource revisions and pin WorkerSpec snapshots and resolved references |
| 4 | in progress | GoalLoop has schema, reference resolution, and Plan artifact; typed Apply and resource-backed domain creation are not connected |
| 5 | in progress | One Draft, form/YAML round trip, diff, bindings, and typed Apply UI are verified for supported kinds; GoalLoop form and Apply UI are absent |
| 6 | in progress | Full backend, full web, Rust/WASM, build, docs, migration, and supported-kind browser checks pass; GoalLoop delivery remains open |

## Delivery Audit

Audit date: 2026-07-15.

| Requirement | Status | Authoritative evidence |
| --- | --- | --- |
| Unified Resource envelope and ResourceRef | proved | `orchestrationresource.Manifest`, draft/resolved `Reference` validation, protocol Resource identity |
| Validate and persisted Plan | proved | control service validation/planning tests and Connect handlers |
| Binding typed Apply | proved | eight registered binding kinds, binding Apply service, protocol and web dispatch |
| Prompt typed Apply | proved | Prompt Apply service, immutable revision tests, protocol and web dispatch |
| WorkerTemplate typed Apply | proved | snapshot compiler, Apply service, revision/snapshot PostgreSQL tests |
| Worker typed Apply | proved | durable launch claim/completion, one-shot constraint, Pod result contract |
| Expert typed Apply | proved | Expert projection stores resource revision and WorkerSpec snapshot |
| Workflow typed Apply | proved | Workflow projection and runs pin resource revision and WorkerSpec snapshot |
| GoalLoop typed Apply | missing | schema and Plan exist; no Apply RPC, service, projection, or frontend kind |
| One form/YAML Draft | proved | reducer invalidates stale plans; YAML parse returns canonical draft |
| Semantic Diff | proved | server digest-only changes and dedicated redaction rendering tests |
| Permission revalidation | proved | target and every resolved reference are reauthorized immediately before Apply |
| Secret reference protection | proved | reference-only schemas, non-echoing errors, digest/redacted Diff, reference-only UI test |
| Migration integrity | proved | 000211/000212 constraints and PostgreSQL mutation/concurrency tests |
| Public documentation | proved | product/API manuals, public docs route, multilingual page, desktop/mobile anonymous browser evidence |

The active blocker is not a validation or documentation ambiguity. Completing
the goal requires a typed GoalLoop Apply path that creates a GoalLoop domain
record from the immutable plan artifact and pins its resource revision and
WorkerSpec snapshot. Current coordination instructions prohibit editing
GoalLoop or Runner files, so this path must remain explicit rather than falling
back to the legacy GoalLoop create API.

## Current Constraints

- The current worktree contains extensive unrelated uncommitted work.
- Implementation must not revert or reformat those changes.
- Files already modified by another task require diff review before editing.
- Runtime code must never fall back to legacy Agent, model, repository,
  AgentFile, or mutable Expert/Workflow fields.
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
