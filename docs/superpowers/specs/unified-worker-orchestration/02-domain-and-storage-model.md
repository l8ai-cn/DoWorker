# Domain and Storage Model

## WorkerDefinitionRevision

`worker_definition_revisions` retains every executable definition bundle used
by an applied snapshot.

| Field | Meaning |
| --- | --- |
| `definition_hash` | Lowercase SHA-256 primary identifier |
| `worker_type_slug` | Stable Worker type identifier |
| `definition_version` | Human-readable source version |
| `definition_json` | Canonical strict definition document |
| `base_agentfile` | Exact base AgentFile source |
| `config_schema_json` | Derived typed configuration schema |
| `executable` | Runtime executable |
| `adapter_id` | Runner protocol adapter |
| `created_at` | Import time |

Content columns are immutable. Availability and revocation live in a separate
policy record so historical content remains inspectable.

## WorkerSpecPlan

`worker_spec_plans` is an expiring server-owned proposal.

| Field | Meaning |
| --- | --- |
| `id` | Opaque plan identifier |
| `organization_id`, `created_by_id` | Scope and ownership |
| `intent_kind`, `target_id`, `expected_revision` | Authorized apply purpose |
| `draft_json` | Canonical normalized input |
| `proposed_spec_json` | Fully resolved WorkerSpec |
| `summary_json` | Safe display summary |
| `dependency_manifest_json` | Pinned dependency revisions |
| `compiled_agentfile_layer` | Deterministic WorkerSpec layer |
| `plan_hash` | Hash of all immutable proposal fields |
| `options_revision`, `policy_revision` | Staleness guards |
| `status` | `ready`, `applied`, `expired`, or `cancelled` |
| `applied_snapshot_id`, `apply_result_json` | Durable application result |
| `expires_at`, `applied_at` | Lifecycle timestamps |

Plans never contain secret values. A plan is scoped to one organization and
creator unless an explicit organization policy grants shared apply rights.

## WorkerSpecSnapshot

The existing `worker_spec_snapshots` table remains the core immutable record.
It gains:

- `spec_sha256` for canonical semantic spec identity;
- `snapshot_sha256` for organization-local content addressing and deduplication;
- `worker_definition_hash` for indexed exact revision lookup;
- `dependency_manifest_json`;
- `compiled_agentfile_layer`;
- `compiled_agentfile_sha256`;
- `created_by_id`.

The immutable trigger covers every content field. `spec_json`, summary,
dependency manifest, and compiled layer must be cross-validated on reads.

`snapshot_sha256` covers canonical spec bytes, dependency manifest bytes,
WorkerDefinition hash, and compiled layer hash. Two snapshots with the same
spec JSON but different pinned dependency revisions are not the same snapshot.

## WorkerSpec V2 Boundary

New plans emit WorkerSpec V2:

- `workspace.initial_task` is removed and becomes invocation input;
- repository selection becomes `{id, ref_policy, ref}`;
- workspace source is `none`, `repository`, or authorized `host_binding`;
- Skill, knowledge, environment, MCP, and tool bindings are explicit
  capability references;
- `metadata.alias` becomes invocation or product metadata;
- `metadata.source_expert_id` becomes snapshot provenance;
- runtime semantics remain limited to runtime, placement, type configuration,
  workspace capability, and lifecycle.

Existing V1 snapshots remain decodable. Only snapshots classified `exact` by
immutable evidence may replay; other V1 rows require a reviewed V2 proposal or
remain `migration_required`. No new V1 snapshot is created after cutover.

## Dependency Manifest

The manifest records stable identifiers needed to interpret the compiled
snapshot:

```json
{
  "worker_definition_hash": "0123...",
  "runtime_image": {"id": 12, "digest": "sha256:..."},
  "model_binding": {
    "resource_id": 31,
    "resource_revision": 7,
    "connection_id": 44,
    "connection_revision": 3
  },
  "skills": [{"id": 9, "revision": 4, "digest": "sha256:..."}],
  "knowledge": [{"id": 18, "revision": 6, "mode": "ro"}],
  "env_bundles": [{"id": 22, "revision": 5}],
  "repository": {"id": 27, "ref_policy": "branch", "ref": "main"}
}
```

Secret bundle revisions identify metadata and authorization state, never
decrypted values.

## WorkerRunManifest

`worker_run_manifests` records one materialization:

| Field | Meaning |
| --- | --- |
| `id`, `organization_id`, `pod_key` | Identity |
| `worker_spec_snapshot_id` | Desired configuration |
| `definition_hash`, `image_digest` | Exact runtime artifacts |
| `compiled_agentfile_sha256` | Snapshot layer identity |
| `effective_agentfile_sha256` | Final merged program identity |
| `command_sha256` | Exact Runner create command identity |
| `invocation_summary_json` | Correlation IDs and protected payload hashes |
| `policy_revision`, `policy_overlay_json` | Applied restrictions |
| `placement_json` | Selected target, Runner, cluster, and resources |
| `workspace_resolution_json` | Repository commit and dependency revisions |
| `secret_reference_manifest_json` | Referenced secret revisions only |
| `created_at` | Materialization time |

The Pod gets non-null foreign keys to both snapshot and run manifest.
Snapshot definition hashes use `ON DELETE RESTRICT`; referenced artifact
revisions are archived instead of physically deleted.

## Expert Revisions

`experts` becomes mutable identity and product metadata:

- organization, slug, name, description, avatar, category;
- status: `draft`, `published`, `migration_required`, `archived`;
- `active_revision_id`;
- run counters and projection status.

`expert_revisions` is immutable:

- `expert_id`, monotonically increasing `version`;
- `worker_spec_snapshot_id`;
- source Pod, marketplace package, or previous revision metadata;
- release notes, creator, and creation time.

Runtime configuration columns are removed from `experts` after migration.

## Workflow Revisions

`workflows` owns identity, enabled state, active revision, scheduling cursor,
and aggregate statistics.

`workflow_revisions` owns immutable execution definition:

- `worker_spec_snapshot_id`;
- prompt template and variable schema/defaults;
- execution principal reference;
- execution mode and persistence policy;
- concurrency policy and limit;
- total and idle timeout;
- retained run policy;
- callback configuration.

Every `workflow_run` pins `workflow_revision_id` and
`worker_spec_snapshot_id`. Editing a Workflow creates a revision and atomically
changes the active pointer.

## GoalLoop and Mesh

GoalLoop keeps its required `worker_spec_snapshot_id`. A started Loop cannot
change that reference.

Mesh Ticket stores a resolved `worker_spec_snapshot_id` and optional
`source_expert_revision_id` for audit. Selecting an Expert copies its active
revision binding at assignment time; later Expert updates do not mutate the
Ticket.

## Identifier and Foreign-Key Rules

- Public slugs follow `slugkit` and database checks.
- Every cross-domain snapshot foreign key includes organization scope.
- Immutable revision rows use `ON DELETE RESTRICT`.
- Product objects are archived instead of deleting referenced revisions.
- Plan and apply APIs never accept organization IDs inferred from untrusted
  payloads; tenant scope comes from authenticated context.
