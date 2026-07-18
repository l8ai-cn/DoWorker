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

## Release And Verification

The resolved dependency contract, migration allocation, cutover order,
deterministic verification, and exit gate continue in
[Resource-Native Phase 4B: Release And Verification](2026-07-16-resource-native-phase-4b-release-verification.md).
