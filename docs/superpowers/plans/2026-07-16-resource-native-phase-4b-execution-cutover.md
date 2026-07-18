# Resource-Native Phase 4B: Execution Entry Cutover

## Goal

Remove the remaining legacy fresh-execution constructors without disabling
session, ticket, quick-task, MCP, or coordinator workflows. Every fresh Pod must
start from exactly one authorized immutable source:

```text
consumed Worker plan | WorkerSpec snapshot | source Pod lineage
```

Runtime fields such as Worker type, model, repository, AgentFile, and placement
must never be reconstructed from mutable domain rows or accepted beside an
immutable source.

## Current Inventory

Six constructors still create fresh Pods from legacy fields:

| Owner | Entry |
| --- | --- |
| Session | create, fork, import, switch, host bind |
| Coordinator | `claimAndDispatch` |

Session message recovery and REST recovery already use source lineage. Connect
Worker creation already uses a Worker plan or structured WorkerSpec draft. Mesh
ticket creation now launches an exact WorkerSpec snapshot with ticket and prompt
invocation metadata. Quick Task and Runner MCP now consume only a Worker plan.

## Invocation Contract

Introduce one service-layer value object:

```text
ExecutionSource
  plan_id?
  worker_spec_snapshot_id?
  source_pod_key?
```

Exactly one field is required. Public adapters may expose only the source kinds
that fit their user action:

- fresh user actions accept a consumed Worker plan result or an exact snapshot;
- replay and rebuild actions accept an exact snapshot;
- resume accepts only a source Pod key;
- no adapter accepts legacy runtime fields beside an immutable source.

`WorkerSpecPromptOverride`, ticket association, session identity, terminal size,
queue policy, and MCP server attachment are invocation metadata. They do not
change Worker identity. Runner, repository, model, Worker type, local path,
AgentFile, automation level, and knowledge mounts belong to the snapshot.

## Domain Ownership

### Session

Add `worker_spec_snapshot_id` and `orchestration_worker_launch_id` to
`agent_sessions`. The Session row, its initial conversation items, and its
immutable execution binding commit in one transaction. Pod materialization is
deferred until the immutable source is authorized; Runner dispatch is released
only after that transaction commits. If the transaction fails, the unbound Pod
and Worker launch are terminated instead of being exposed as a successful
Session.

- New session: the Session API accepts `plan_id`, consumes the Worker plan
  server-side, and returns only after the Session binding commits.
- Fork without a Worker change: launch the source session snapshot.
- Fork with a Worker change: accept `plan_id`; `agent_id` and model overrides
  are removed.
- Import: accept `plan_id`; `agent_id`, host, and runtime overrides are removed.
- Switch: launch the new snapshot, atomically replace the session binding, then
  terminate the old Pod. A Worker change never forwards the old provider
  `external_session_id`; Session items remain the conversation source.
- Host rebind: accept a new `plan_id` whose target/placement ResourceRef
  expresses the requested host. The endpoint cannot pass `RunnerID` or
  `workspace` beside the plan. A path host identifier is only an assertion
  against the materialized Runner, never a scheduling override.
- Message recovery remains source-lineage resume.

The lifecycle order is deterministic:

1. authorize the immutable source;
2. materialize a Pod without releasing Runner dispatch;
3. transactionally persist Session binding and conversation items;
4. release Runner dispatch;
5. for replacement flows, terminate the old Pod.

Session persistence failure terminates the new Pod. Replacement binding failure
leaves the old Pod and old Session binding unchanged.

### Quick Task

Quick Task becomes a short path over an already planned Worker:

```text
plan_id
```

The endpoint consumes the plan through the orchestration Worker Apply service.
PromptRef, inputs, alias, model, tools, knowledge, permissions, runtime and
placement remain part of the planned Worker/WorkerTemplate graph. The endpoint
does not accept invocation overrides, auto-select a Worker type, or select a
Runner outside snapshot materialization. The UI may preserve the one-click
experience by remembering the user's selected Worker draft, but absence of a
valid plan is a blocking state.

### Mesh Ticket

`CreatePodForTicket` accepts `worker_spec_snapshot_id` and optional prompt
override. Ticket association is invocation metadata. The fixed `do-agent`,
model string, permission mode, Runner selection, and generated AgentFile layer
are removed.

Snapshot materialization must allow `TicketID` and `TicketSlug` because they do
not alter the immutable WorkerSpec. It must continue rejecting Runner,
repository, model, local path, AgentFile, and automation overrides.

### Runner MCP

Replace the `create_pod` payload with:

```text
plan_id
```

The backend reauthorizes and consumes the plan for the caller's organization.
The Runner cannot submit snapshot IDs, Resource revisions, runtime fields, or
credentials. Prompt, ticket, model, repository, permission, runtime, and
placement settings remain inside the immutable plan. Generated Runner tool
schema and user documentation change in the same patch.

### Coordinator

Add the following immutable execution binding to `coordinator_projects`:

```text
orchestration_resource_id
orchestration_resource_revision
worker_spec_snapshot_id
```

Project create and explicit rebind accept a pinned WorkerTemplate ResourceRef,
resolve its applied revision, and persist the resulting revision and snapshot.
Dynamic task text is an invocation prompt override; the source repository used
to discover issues remains coordinator metadata and is never copied into the
Pod request.

Workflow ResourceRefs are not accepted by this direct Pod dispatcher. Supporting
Workflow dispatch requires the Workflow trigger service and its execution
manifest; extracting the Workflow's Worker snapshot here would bypass Workflow
semantics.

Each `coordinator_executions` row copies the project's resource ID, revision,
and snapshot before dispatch. Changing the project execution resource is an
explicit compare-and-swap rebind and never changes existing executions.

`RunnerEnsurer` cannot select or provision by mutable `agent_slug`. Runner
selection comes from the prepared WorkerSpec snapshot and runtime catalog. The
legacy pre-dispatch ensure step is removed unless it can consume that exact
prepared snapshot.

Dispatch also needs compensation and retry ownership:

1. claim the external task;
2. create or reuse the linked ticket idempotently;
3. persist an execution in `claimed` state with the immutable binding;
4. launch only from `worker_spec_snapshot_id`, task prompt, and ticket metadata;
5. attach the Pod and move the execution to `running`;
6. terminate the Pod if execution attachment fails.

An external link alone is not proof of successful dispatch. A linked task with
no active execution remains retryable instead of being skipped forever.

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

The formal release mainline head is `000224_validate_migration_lineage`.
`000222` through `000224` are owned by the Loop/Seedance release and must not be
occupied. Resource-native migrations may start at candidate `000225`, but no
specific migration is reserved or authorized until the owner orders the local
`000221` collision and the remaining schema work.

After confirmation, split the database work by invariant:

1. first assign a formal post-`000224` number to the colliding optional-model
   migration or prove that a formal migration already supersedes it;
2. a separately assigned migration adds the resolved-dependency artifact schema
   and immutability guard without assuming it owns `000225`;
3. a later confirmed migration enables the deferred requirement that every new
   WorkerSpec snapshot has exactly one artifact by transaction commit;
4. later confirmed migrations add Session snapshot and Worker launch ownership;
5. add a composite Pod binding key so Session organization, Pod key, and
   snapshot identity must agree;
6. add coordinator project resource revision and snapshot ownership;
7. add the same immutable binding to coordinator executions;
8. reject partially populated coordinator project or execution bindings;
9. verify each bound snapshot belongs to the pinned resource revision;
10. leave historical rows nullable and fail fresh execution from unbound rows;
11. avoid guessing snapshots from legacy `agent_slug`, model, or AgentFile data.

Static filename uniqueness or a locally missing `000222` through `000224` is not
authorization to create, execute, or renumber a migration.

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
