# Unified Worker Orchestration SSOT Design

## Status

- Date: 2026-07-13
- Status: Proposed for review
- Scope: WorkerDefinition, WorkerSpec, Expert, Workflow, GoalLoop, Mesh Ticket, Pod

## Decision

Agent Cloud will use `WorkerSpecSnapshot` as the only reusable runtime
configuration contract. Expert, Workflow, GoalLoop, Mesh Ticket, marketplace
installation, and direct Worker creation must all resolve to one immutable
snapshot before a Pod can be created.

The first implementation does not add HCL, local state files, or a second
Kastor-style runtime. Terraform-style behavior is implemented in the existing
backend control plane:

1. `validate` checks structure and schema without persisting runtime state.
2. `plan` resolves references, policy, compatibility, and an immutable proposed
   snapshot.
3. Target-specific `apply` consumes the exact plan to create a Worker, publish
   an Expert revision, or update a Workflow revision.
4. Target-specific lifecycle commands terminate, archive, or delete objects.

Database state remains authoritative. A future HCL or YAML adapter may compile
into `WorkerSpecDraft`, but it cannot become another runtime source of truth.

## Why Two Immutable Records Are Required

`WorkerSpecSnapshot` freezes configuration intent. It includes the Worker
definition revision, image digest, model binding revisions, typed
configuration, workspace references, placement intent, and lifecycle.

`WorkerRunManifest` freezes what one Pod actually received. It additionally
records the selected Runner or cluster, repository commit, resolved dependency
revisions, policy overlay, compiled AgentFile hashes, and secret reference
revisions without storing secret values.

This distinction prevents two incorrect promises:

- A snapshot cannot guarantee that a revoked credential or deleted compute
  target remains executable.
- A branch-based workspace cannot guarantee identical repository contents
  across runs without recording the resolved commit.

## Design Invariants

1. Every fresh Pod has one non-null `worker_spec_snapshot_id`.
2. Every dispatched Pod has one immutable `worker_run_manifest_id`.
3. Snapshot-backed execution rejects legacy runtime fields instead of merging
   them.
4. Expert and Workflow runtime edits always create new immutable revisions.
5. Metadata edits do not create runtime revisions.
6. Current WorkerDefinition changes do not invalidate retained historical
   definition revisions.
7. Revocation and authorization are live policy checks and can block replay.
8. Secret values never enter plans, snapshots, revisions, diffs, logs, or
   manifests.
9. AgentFile is a compiled artifact, not a user-maintained business SSOT.
10. Missing resources, malformed JSON, stale plans, and projection drift fail
    explicitly.

## Detailed Documents

| Document | Question answered |
| --- | --- |
| [Current State and Decisions](unified-worker-orchestration/00-current-state-and-decisions.md) | What conflicts exist and which approach was selected? |
| [Target Architecture](unified-worker-orchestration/01-target-architecture.md) | Which components own configuration and execution state? |
| [Domain and Storage Model](unified-worker-orchestration/02-domain-and-storage-model.md) | Which immutable and mutable records are required? |
| [Control Plane Protocol](unified-worker-orchestration/03-control-plane-protocol.md) | How do validate, plan, and apply work? |
| [Domain Integration](unified-worker-orchestration/04-domain-integration.md) | How do Expert, Workflow, Loop, Mesh, and Market consume snapshots? |
| [Runtime Materialization](unified-worker-orchestration/05-runtime-materialization.md) | How is one snapshot converted into one Pod safely? |
| [Migration and Rollout](unified-worker-orchestration/06-migration-and-rollout.md) | How is the legacy runtime removed without fallback? |
| [Verification and Acceptance](unified-worker-orchestration/07-verification-and-acceptance.md) | Which tests and release gates prove completion? |
| [Reliability and Security](unified-worker-orchestration/08-reliability-and-security.md) | How are launch, trigger, identity, and callback failures recovered safely? |
| [Legacy Lineage and Rollback](unified-worker-orchestration/09-legacy-lineage-and-rollback.md) | Which historical objects can migrate, resume, or roll back? |
| [Resource-Native Contract](unified-worker-orchestration/10-resource-native-contract.md) | How do typed resources, references, schemas, and domain actions fit together? |
| [Resource Editor Frontend](unified-worker-orchestration/11-resource-editor-frontend.md) | How do forms, YAML, plans, diffs, revisions, and status share one draft? |

## Implementation Order

1. Repair the current WorkerSpec contract and restore focused tests.
2. Persist WorkerDefinition revisions and compiled snapshot artifacts.
3. Add plans, launch intents, durable commands, and WorkerRunManifest.
4. Audit V1 snapshots, resumable Pods, and persistent Workflow lineage.
5. Cut Expert creation, editing, publishing, marketplace installation, and
   execution to revisions.
6. Cut Workflow execution to immutable Workflow revisions.
7. Require snapshots for Mesh Ticket and every remaining fresh Pod entry point.
8. Delete legacy runtime builders, fields, request shapes, and UI controls.

## Related Existing Designs

- [Worker Creation and Publishing Redesign](2026-07-10-worker-creation-publishing-design.md)
- [Loop and Workflow Domain Split](2026-07-11-loop-workflow-domain-split-design.md)
- [WorkerSpec Runtime Foundation](../plans/2026-07-10-worker-phase1-workerspec-runtime-foundation.md)
- [Expert Platform Capability Map](../../product/expert-platform-capability-map.md)
