# WorkerSpec Resolved Dependency Artifact

## Status
This design is migration-frozen at `000224_validate_migration_lineage`;
candidate `000225` is unreserved until the owner orders the local `000221`
collision and pending migrations. Runner, GoalLoop, migration, database,
service-start, and browser actions are forbidden.
The strict domain codec and Plan-owned builder are implemented and verified.
## Decision
Keep WorkerSpec V1 unchanged. Add one complete immutable artifact per new
snapshot; never add optional V1 fields or reconstruct facts from current rows.
Identity is `(organization_id, worker_spec_snapshot_id)`. After cutover it is
the only source of resolved non-Secret dependency facts.
## Candidate Schema
After the owner disposes of the local `000221` collision, the artifact migration
uses the next assigned post-`000224` number. It must not assume `000225`.
```sql
CREATE TABLE worker_spec_resolved_dependencies (
    organization_id BIGINT NOT NULL, worker_spec_snapshot_id BIGINT NOT NULL,
    version SMALLINT NOT NULL, document_json JSONB NOT NULL,
    digest VARCHAR(71) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (organization_id, worker_spec_snapshot_id),
    FOREIGN KEY (organization_id, worker_spec_snapshot_id)
        REFERENCES worker_spec_snapshots (organization_id, id) ON DELETE RESTRICT,
    CHECK (version = 1), CHECK (jsonb_typeof(document_json) = 'object'),
    CHECK (document_json->>'version' = version::TEXT),
    CHECK (digest ~ '^sha256:[0-9a-f]{64}$')
);
```

`ON DELETE RESTRICT` protects historical facts. A trigger rejects artifact
UPDATE/DELETE; the snapshot gains no mutable artifact ID. The stored SHA-256 is
over canonical JSON, encoded once and inserted with the snapshot transaction.

## Document V1
The implemented domain contract is
`backend/internal/domain/workerdependency.Document`. The root object contains:

```json
{
  "version": 1,
  "organization_id": 7,
  "namespace": "team-alpha",
  "worker": {
    "worker_type": "codex-cli",
    "adapter_id": "codex-app-server",
    "spec_version": 1, "spec_digest": "sha256:...",
    "definition_hash": "...",
    "model_managed_fields": [], "credential_bundle_fields": [],
    "agentfile_source": "...",
    "agentfile_source_digest": "sha256:..."
  },
  "models": {"primary": null, "tools": []},
  "repository": null,
  "skills": [], "knowledge_bases": [],
  "runtime_bundles": [], "secret_refs": [],
  "placement": {}
}
```

Domain-backed dependencies use `ResourcePin`: exact `ResourceRef` plus domain
row ID. ToolBinding has no domain row, so it stores only its exact ResourceRef
and the separately resolved ModelBinding pin.

`models` records ModelBinding pins, AI resource/connection revisions, provider,
adapter, model ID, BaseURL, modalities, capabilities, role, and environment
targets. `worker` records type, adapter, Definition field ownership, and the
complete merged AgentFile; bundles/literals cannot override managed targets.

`repository` records pin, clone endpoints, branch, commit, preparation, and a
`credential_ref`: explicit no-auth or fixed user credential ID/type/
owner, never token/private-key data. `runner_local` is rejected until an exact
Runner Secret/attestation resource exists. Default changes do not affect old
snapshots. Skills and knowledge bases record immutable package/clone facts.

`runtime_bundles` is ordered because later `USE_ENV_BUNDLE` overrides earlier
values. Each entry stores pin, kind, exact non-Secret values/digest, and config
metadata. Config bundles require the runtime `__json` key with a JSON object.
Values inside one bundle are sorted; the bundle list is not.

`secret_refs` records pin, target, source key, and owner only for
Definition-declared credential fields. It never stores values, hashes, or
provider error bodies.

`placement` records catalog revision, immutable image ID/reference/digest,
ComputeTarget and optional ResourceProfile pins, and exact WorkerSpec placement.

Unknown fields/versions fail. All sections exist; collections are arrays,
optionals are explicit `null`, UUIDs are canonical, and size is at most 1 MiB.
The stored dependency digest is over the same canonical bytes used by Plan.

## Writer Boundary
`workercreation.Prepared` cannot build this document: it lacks Plan references,
merged Definition AgentFile/adapter, model metadata, ToolBinding identity,
commits, package/bundle facts, fixed Git credential identity, ownership, and image
reference. The writer receives WorkerSpec plus typed resolver facts;
each domain ID/reference pair is created through `BindResourceProjection`.
An admission budget rejects oversized facts, WorkerSpec, Plan references,
non-tree JSON, and custom marshalers before allocation. Apply cannot rerun resolvers.

`workerdependencyartifact.Build` materializes Document itself and requires exact
closure with Plan direct references. It reparses and rehashes the raw
Worker Definition JSON plus base AgentFile, checks its typed projection, merges
the AgentFile layer, and derives adapter and field policy from that
verified snapshot. ToolBinding-to-ModelBinding is the only transitive edge and
must have explicit typed resolution evidence. Models, workspace, Secrets,
image, and placement are checked against one canonical WorkerSpec.

## Plan Artifact
Plan uses `WorkerTemplateBuild`, not a WorkerSpec-shaped compatibility payload.
Its versioned canonical JSON contains `workerSpec`, `resolvedDependencies`, and
`resolvedDependenciesDigest`. Plan hash binds the composite digest and direct
reference set. Its Plan DTO names tool environment targets explicitly
(`api_key_target`, etc.), so the control-plane Secret guard stays structural.
Strict decode restores V1 and revalidates both documents without current rows.

## Write Transaction
Apply must perform these writes in one transaction:

1. lock and reauthorize the Plan and target resource;
2. use `DecodeApplyPlan` on the persisted build, then verify live Secret access;
3. insert its WorkerSpec snapshot and exactly one dependency artifact;
4. insert the resource revision and domain projection;
5. consume the Plan and commit.

Failure rolls back snapshot and projections. Retry reuses the persisted Plan
result and cannot regenerate from newer rows.

## Runtime Read
After runtime cutover, launch loads the artifact only by
`(organization_id, worker_spec_snapshot_id)`, verifies version and digest, and
decodes strictly. Missing, malformed, or digest-mismatched artifacts are
explicit precondition failures.

All non-Secret dependency facts come only from the artifact; current/latest
rows, names, AgentFile, and legacy Agent data are forbidden.

Secret values stay live and resolve after tenant, permission, active, purpose,
and target checks. Rotation remains visible; unavailable refs fail without data.

## Historical Audit
Historical rows materialize only when every fact is proven by immutable
evidence. Current dependency rows cannot guess old values; no proof means no
artifact.

Audit reports snapshot, owner refs, proof status, and first missing fact.
Unprovable rows stay absent, never partial. Active definitions re-Plan/Apply;
completed history remains readable but unbound snapshots cannot launch.

## Staged Cutover
1. Have the owner order every pending migration from candidate `000225`.
2. Apply the schema-only table, checks, and immutability trigger.
3. Deploy all writers for transactional snapshot/artifact insertion.
4. Audit deterministic history and re-Plan/Apply active unbound definitions.
5. Enable one global fresh-launch fence across Worker Apply, Expert, Workflow,
   GoalLoop, Session, Quick Task, Runner MCP, Mesh, and Coordinator entry points.
6. Drain in-flight launches, stop every old-reader backend, start the complete
   artifact-only fleet behind the fence, and verify missing artifacts fail.
7. Remove the fence only after routing proves no old-reader instance remains.
8. Confirm zero new unbound snapshots and active-history coverage.
9. Use a later owner-confirmed migration to add a DEFERRABLE INITIALLY DEFERRED
   trigger requiring one same-organization artifact at snapshot insert commit.

The enforcement trigger follows writers. Readers never roll gradually; the
fence prevents mixed old/new execution.

## Rollback
Before reader cutover, roll back the app and leave immutable rows unused. After
cutover/enforcement, old readers are prohibited: forward-fix or stop writes and
restore matching DB/app releases. Down migration fails if artifacts exist.

## Acceptance
- New snapshot and artifact commit atomically; injected artifact failure leaves
  neither row nor domain projection.
- Artifact UPDATE and DELETE fail; snapshot DELETE is restricted.
- Canonical digest changes for any runtime fact and is stable for equivalent
  canonical input.
- `WorkerTemplateBuild` rejects unknown fields, version drift, non-canonical
  bytes, swapped WorkerSpec/dependencies, or a mismatched nested digest.
- Changing Worker type or EnvironmentBundle order changes the digest; changing
  UUID spelling or within-bundle value order does not.
- ToolBinding identity is a ResourceRef and never a fabricated domain row ID.
- AgentFile is complete, parseable, digest-bound, and contains no literal for a
  model-managed or credential-managed field; adapter ID is snapshot-bound.
- Repository credentials are identity-only live refs; rotation is visible,
  default selection changes are not, `runner_local` fails, and lost access fails.
- Runtime/Secret bundles cannot claim model-managed fields; Secret targets must
  be Definition-declared credential-bundle fields.
- Duplicate domain IDs, duplicate config document IDs, embedded URL userinfo,
  unnamed OCI digest references, and Secret-like materialized values fail.
- Updating model ID, BaseURL, Skill package, repository commit, knowledge
  content, bundle values, catalog, or placement does not change an old launch.
- Secret rotation is visible to old snapshots without persisting the value.
- Missing artifact, content, Secret access, or proof fails explicitly.
- Audit never materializes an unprovable historical row.
- Runtime tests prove no current-row, latest-by-name, AgentFile, or Agent fallback.
- Writer tests prove exact Plan closure, typed ToolBinding model provenance,
  WorkerSpec consistency, admission budgets, and immutable output ownership.
