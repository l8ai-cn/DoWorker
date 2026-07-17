# Legacy Lineage and Rollback

## Historical Evidence Rule

Existing V1 snapshots cannot be assumed to support exact replay. Current rows
contain semantic spec and summary, but may lack retained WorkerDefinition
content, compiled AgentFile, and immutable dependency revisions.

Migration classifies each snapshot:

| Class | Required evidence | Allowed behavior |
| --- | --- | --- |
| `exact` | Definition bundle and every non-secret dependency revision retained | Replay and resume |
| `configuration_only` | Semantic spec valid, mutable historical content missing | Create reviewed V2 proposal |
| `unrecoverable` | Definition or required identity unavailable | `migration_required` |

Current resource state cannot be presented as historical evidence. A generated
V2 snapshot always records that it is a new reviewed revision, not a recovered
original.

## Evidence Sources

Migration may use only immutable or auditable evidence:

- snapshot canonical JSON and summary;
- Git history containing a bundle with the recorded definition hash;
- persisted Pod AgentFile layer and resolved config;
- immutable Skill or artifact digest;
- Runner command or release audit record;
- repository commit already stored on the Pod or run;
- explicit user review and newly generated plan.

Mutable current defaults, renamed resources, current branch HEAD, and current
secret values cannot fill historical gaps.

## Migration Proposal

Batch audit does not create ordinary 15-minute plans. It writes a durable
`migration_proposal`:

- object type and ID;
- legacy source hash;
- candidate V2 draft and candidate hash;
- evidence classification;
- field-level conflicts and blocking issue codes;
- reviewer and decision;
- final plan and application IDs after approval.

After review, the backend creates a fresh plan. Apply requires the current
legacy source hash and candidate hash to match the proposal, preventing edits
between audit and cutover.

## Expert Divergence

Snapshot-backed Experts are not migrated automatically when their legacy
runtime fields or Git projection differ from the snapshot.

Migration normalizes and compares:

- bound WorkerSpec snapshot;
- mutable Expert runtime columns;
- generated Git AgentFile and expert metadata.

No difference creates revision 1 from the snapshot. Any semantic difference
creates `migration_required` with a field-level diff. The reviewer chooses the
existing snapshot or applies a new V2 revision. Migration never guesses which
copy was intended.

## Pod and Run Lineage

Before resume cutover, every non-terminal or resumable Pod is classified:

- snapshot and immutable execution evidence available: backfill exact manifest;
- snapshot available but effective dependencies uncertain: disable resume and
  require a new Worker launch;
- no snapshot: mark `resume_migration_required`;
- active persistent Workflow session: also classify its WorkflowRun and
  predecessor relationship.

Backfilled manifests record `source = legacy_evidence` and the evidence hashes.
They are created only when the command and effective configuration can be
proven. They are not synthesized from current Worker defaults.

## Persistent Workflow Cutover

For each Workflow with `LastPodKey` or an active run:

1. classify the source Pod lineage;
2. pin the Workflow revision and snapshot;
3. establish a valid predecessor run;
4. enforce persistent concurrency limit `1`;
5. either enable resume or mark the next run `fresh_session_required`.

Choosing a fresh session is an explicit migration decision and is shown before
cutover. Runtime never attempts legacy resume and silently starts fresh.

## Remaining Fresh Pod Entry Points

The execution inventory includes:

- Worker wizard and quick task;
- Expert, Workflow, GoalLoop, and Mesh Ticket;
- session create, import, fork, switch, MCP update, and session rebuild;
- host workspace binding;
- coordinator and Runner MCP creation;
- marketplace installation;
- resume and fork lineage.

Session fork or switch that changes Worker, model, tools, or workspace creates a
new V2 plan. Host execution uses an authorized `host_binding_id`, never a raw
per-run local path. MCP and tool bindings are explicit V2 capability
references. Runner MCP cannot submit raw AgentFile for a fresh Pod.

## GoalLoop Foreign Key

GoalLoop migration replaces its snapshot ID-only foreign key with
`(organization_id, worker_spec_snapshot_id)` and adds a PostgreSQL migration
test proving cross-organization references are rejected.

## Cutover Compatibility Matrix

Every production cutover publishes a release manifest containing:

- additive-schema build SHA;
- audit-tool SHA;
- cutover build SHA;
- revision-aware rollback build SHA;
- minimum and maximum compatible schema version;
- completed migration report hash;
- rollback rehearsal evidence.

The revision-aware rollback build is deployed and smoke-tested before cutover.
It understands WorkerSpec V2, revisions, launch intents, and run manifests.

## Rollback Policy

After data cutover, rollback to a legacy-only binary is forbidden. Rollback
means switching to the prevalidated revision-aware rollback build.

If no such build can preserve the current schema and runtime semantics, the
release is explicitly roll-forward-only and requires:

- a tested corrective deployment path;
- longer canary and observation gates;
- operator approval;
- no destructive schema cleanup.

Legacy runtime columns are dropped only after the rollback window closes and a
later release confirms that no supported rollback artifact reads them.
