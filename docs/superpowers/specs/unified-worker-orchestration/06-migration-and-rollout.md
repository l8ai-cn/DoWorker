# Migration and Rollout

## Migration Rule

The rollout may use temporary schema columns and offline migration commands.
It must never run two execution semantics for the same domain object.

Migration code may read legacy fields to generate a candidate draft. Runtime
code cannot fall back to legacy fields when a snapshot is missing or invalid.

## Phase 0: Repair Current Contracts

Before structural migration:

- complete `ModelBinding.ProtocolAdapter` propagation across fixtures,
  migrations, API, and consumers;
- restore focused WorkerSpec, Expert, Workflow, Mesh, and GoalLoop tests;
- replace ignored JSON and repository errors in touched paths;
- add an inventory test listing every fresh Pod entry point.

Exit gate: current behavior is testable and failures are explicit.

## Phase 1: Immutable Artifact Foundation

Add:

- `worker_definition_revisions`;
- expanded `worker_spec_snapshots`;
- `worker_spec_plans`;
- `worker_run_manifests`;
- `worker_launch_intents`;
- immutable Skill and knowledge revisions plus environment binding revisions;
- definition import and retention service;
- Runtime Materializer.

The schema migration widens WorkerSpec version checks for V2 and adds
restricting foreign keys from snapshots to retained definition revisions.

Current WorkerDefinition sync imports immutable revisions and separately rebuilds
Agent query projections. Backend startup health fails if the current catalog
cannot be imported or its projection is inconsistent.

Exit gate: a V2 or exact-classified V1 snapshot runs after the current
WorkerDefinition changes, unless its exact revision is revoked.

## Phase 2: Plan-Based Worker Creation

Change the Worker wizard and Connect APIs:

- preflight produces a persisted plan;
- create consumes plan ID and hash;
- Pod creation persists snapshot and run manifest;
- snapshot execution accepts only typed invocation fields.

Within the Worker creation domain, direct fresh legacy requests are disabled
and only resume remains. Expert, Workflow, and other product domains switch in
their own atomic cutover phases; no endpoint accepts both legacy runtime fields
and a plan or snapshot.

Exit gate: all new Web Worker creation uses plan apply.

## Phase 3: Legacy Lineage Audit

Classify every V1 snapshot, resumable Pod, active WorkflowRun, and persistent
session. Generate durable migration proposals instead of ordinary short-lived
plans. Objects without immutable evidence become `migration_required` or
`fresh_session_required` before any domain cutover.

Exit gate: no cutover object depends on legacy resume or fabricated history.

## Phase 4: Expert Cutover

Add `expert_revisions` and `experts.active_revision_id`.

Migration classification:

- snapshot-backed Expert with matching legacy and Git projections: create
  revision 1 from its snapshot;
- snapshot-backed Expert with divergent projections: require reviewed diff;
- legacy Expert with resolvable fields: strict migration command creates a
  plan and reports the proposed diff;
- unresolved Expert: set `migration_required`, preserve metadata, disable run;
- marketplace Expert: reinstall from a versioned package plan.

After review and apply:

- Expert runtime endpoints require revision or active revision;
- metadata and runtime edit APIs are separated;
- Git backing becomes an outbox projection;
- legacy runtime fields are no longer read.

Exit gate: every runnable Expert has an active revision.

## Phase 5: Workflow Cutover

Add `workflow_revisions` and pin every WorkflowRun.

The migration command validates every active Workflow:

- maps Agent and model to a valid WorkerSpec draft;
- resolves repository, bundle, Skill, and target identifiers;
- rejects malformed config and unsupported permission modes;
- produces a per-Workflow report and candidate plan.

Unresolved Workflows are disabled with `migration_required`. The cutover
release switches execution atomically to revisions and removes the legacy
AgentFile builder from the runtime path.

Trigger migration adds stable occurrence keys, execution principals, atomic
Cron cursor advancement, and persistent-session concurrency validation.

Exit gate: Cron, API, event, and manual triggers all pin a Workflow revision and
WorkerSpec snapshot.

## Phase 6: GoalLoop, Mesh, and Remaining Entry Points

- GoalLoop dispatches through Runtime Materializer.
- GoalLoop receives an organization-scoped composite snapshot foreign key.
- Mesh Ticket requires a resolved snapshot binding.
- quick task, session create/import, coordinator launch, and marketplace
  installation use plan or snapshot APIs.
- direct fresh legacy Pod creation is rejected at the transport boundary.

Exit gate: inventory test finds no fresh execution path accepting Agent,
model, repository, or AgentFile runtime fields.

## Phase 7: Schema Cleanup

After one stable release:

- drop Expert legacy runtime columns;
- drop Workflow legacy runtime columns;
- remove legacy request fields from Proto, REST, Rust Core, and Web;
- delete Expert and Workflow AgentFile reconstruction helpers;
- delete Mesh hardcoded Worker defaults;
- remove current-definition equality checks from snapshot replay;
- keep only explicit migration reporting tools.

## Deployment Sequence

1. Deploy additive schema and revision import.
2. Run read-only migration audit and publish durable migration proposals.
3. Resolve or explicitly disable blocking legacy objects.
4. Deploy atomic domain cutover.
5. Verify health, scheduled runs, Expert run, GoalLoop, Ticket, and resume.
6. Observe one release.
7. Deploy destructive column and API cleanup.

## Rollback

Rollback is release-based, not a runtime fallback. The exact compatibility
matrix and historical lineage rules are defined in
[Legacy Lineage and Rollback](09-legacy-lineage-and-rollback.md).

- Before cutover, rollback application code while additive tables remain.
- After cutover, rollback only to a build that understands revision tables.
- Never repopulate legacy fields from snapshots to restore old runtime paths.
- Schema cleanup begins only after the supported rollback window closes.

## Migration Evidence

The migration command emits machine-readable records:

```text
object_type
object_id
status: ready | migration_required | applied
migration_proposal_id
blocking_issue_codes
legacy_field_hash
candidate_hash
final_plan_id
reviewed_by
applied_at
```

Reports contain no secret values and are retained with release evidence.
