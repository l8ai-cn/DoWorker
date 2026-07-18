# Resource-Native Phase 4B: Release And Verification

This document continues the
[execution entry cutover design](2026-07-16-resource-native-phase-4b-execution-cutover.md)
with the immutable dependency, migration, rollout, and acceptance contracts.

## Resolved Dependency Snapshot Contract

Pinning a Resource revision and WorkerSpec ID is necessary but not sufficient.
The runtime must not reconstruct historical execution from mutable AI resource,
repository, Skill, KnowledgeBase, EnvironmentBundle, compute target, or resource
profile rows.

Add one complete immutable resolved-dependency artifact per WorkerSpec snapshot.
The exact schema, document contract, transaction boundary, deterministic
backfill rules, cutover stages, and rollback limits are defined in
`2026-07-16-worker-spec-resolved-dependency-artifact.md`.

Operational AI resource changes do not alter configuration revisions:
validation state, display name, enable state, and credential rotation are
metadata. Base URL, model ID, modalities, and capabilities are runtime
configuration and advance the relevant revision. Validation persistence must
also compare the credential ciphertext identity so a probe started before
rotation cannot approve the new credentials.

Required behavior scenarios:

- Given an applied WorkerSpec, when its Skill is updated, then the old snapshot
  launches with the old package digest and a new revision uses the new digest.
- Given an applied WorkerSpec, when model display state or credentials change,
  then the old snapshot remains launchable without a configuration revision
  change.
- Given an applied WorkerSpec, when model ID or base URL changes, then the old
  snapshot uses its materialized values and a new revision uses the new values.
- Given a referenced Secret is disabled, deleted, or unauthorized, when launch
  starts, then it fails explicitly without exposing Secret data.
- Given a historical row cannot be deterministically materialized, when audit or
  launch reaches it, then it stays unbound and fresh launch requires re-plan.

## Database Changes

The formal release mainline head is `000225_agent_workbench_stream`.
`000226` through `000228` are assigned in dependency order:

1. `000226_enforce_orchestration_domain_snapshot_consistency` makes each bound
   Expert, Workflow, Workflow run, GoalLoop, and Worker launch reference the
   exact Resource revision and WorkerSpec snapshot tuple.
2. `000227_workflow_run_execution_manifest` persists and validates the immutable
   execution manifest used by active resource-native Workflow runs.
3. `000228_worker_spec_optional_model_binding` permits model-free Worker types
   to persist the canonical empty model binding.

Later execution-entry migrations must start after `000228` and be confirmed
again with the release mainline owner. Split the remaining database work by
invariant:

1. add the resolved-dependency artifact schema and immutability guard;
2. enable the deferred requirement that every new WorkerSpec snapshot has
   exactly one artifact by transaction commit;
3. add Session snapshot and Worker launch ownership;
4. add a composite Pod binding key so Session organization, Pod key, and
   snapshot identity must agree;
5. add coordinator project resource revision and snapshot ownership;
6. add the same immutable binding to coordinator executions;
7. reject partially populated coordinator project or execution bindings;
8. verify each bound snapshot belongs to the pinned resource revision;
9. leave historical rows nullable and fail fresh execution from unbound rows;
10. avoid guessing snapshots from legacy `agent_slug`, model, or AgentFile data.

Static filename uniqueness is not authorization to create, execute, or renumber
a migration.

## Cutover Order

1. Add `ExecutionSource` validation and permit ticket/session association for
   snapshot launches.
2. Add a deferred-dispatch Worker launch boundary for owning-domain binding.
3. Add migration and repositories for Session and Coordinator bindings.
4. Convert session create/fork/import/switch/host flows and their Web User
   clients.
5. Convert Quick Task and its Web client.
6. Convert Runner MCP schema and backend adapter.
7. Convert Coordinator project API, execution pinning, and dispatch recovery.
8. Remove legacy request builders and require zero legacy inventory entries.
9. Update product, API, migration, and operator documentation.

No step keeps a legacy branch. A route is switched only when its caller and
functional test are switched in the same change.

Current checkpoint: PodOrchestrator resolves plan, snapshot, and source lineage
before preparation and rejects multiple immutable sources. Fresh Workflow runs
use their pinned snapshot; persistent Workflow runs use source lineage and carry
the next prompt only as invocation metadata. Six migration-gated legacy
constructors remain and are not represented as an `ExecutionSource`.

The current lineage loader returns without a WorkerSpec when a historical
source Pod has no `worker_spec_snapshot_id`. That path cannot be accepted at the
exit gate. It may only be removed after the six legacy constructors are
converted, because Session message recovery can currently target Pods created
by those paths. The cutover must then reject missing-snapshot lineage and replace
the legacy resume fixtures with snapshot-backed sources in the same change.

## Deterministic Verification

- Inventory reports zero `legacy` constructors.
- Each converted entry rejects zero, multiple, cross-tenant, missing, and stale
  execution sources.
- Session tests prove failed persistence terminates the unbound Pod.
- Switch tests prove the old Pod remains usable until the new binding commits.
- Mesh and Coordinator tests prove ticket metadata does not change snapshot
  identity.
- Coordinator tests prove a linked task without an active execution retries,
  and failed execution attachment terminates the new Pod.
- MCP tests prove the Runner schema exposes only `plan_id` and the backend
  rejects runtime fields and snapshot IDs.
- PostgreSQL tests prove revision/snapshot ownership and transaction rollback.
- Browser tests cover new session, fork, import, switch, quick task, and ticket
  launch success plus missing-plan blocking states.

## Exit Gate

- Zero legacy fresh-execution constructors.
- Source lineage without `worker_spec_snapshot_id` fails closed.
- No automatic Worker, model, repository, Runner, or AgentFile selection outside
  Validate and Plan.
- Every fresh Pod and owning domain row agree on WorkerSpec snapshot identity.
- No Runner command is released before its owning Session or Coordinator
  execution binding commits.
- Migration rehearsal, service startup, and browser paths pass.
- Documentation exposes ResourceRef and plan concepts, never Secret values or
  mutable runtime overrides.
