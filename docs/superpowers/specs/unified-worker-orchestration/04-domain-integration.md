# Domain Integration

## Expert

### Create

The Expert form edits a WorkerSpec draft plus Expert metadata. Submission plans
the draft and applies `CreateExpertFromPlan`.

A created Expert always has an active immutable revision. There is no runnable
Expert without a snapshot.

### Edit

The UI separates:

- metadata: name, description, avatar, category;
- runtime revision: Worker, model, image, capabilities, workspace, and policy.

Metadata updates modify `experts`. Runtime updates import the active snapshot
into a draft, produce a plan and diff, then create a new Expert revision.

### Run

Run resolves either the active revision or an explicitly requested version.
The request can supply alias and task prompt input only. It cannot alter Worker
capability.

### Publish From Pod

Publishing creates an Expert revision from the source Pod snapshot. The source
run manifest is retained as provenance, but the Expert revision references the
WorkerSpec snapshot rather than copying effective system policy or secret
values.

## Expert Git Backing

Git is a generated projection:

- `expert.json` contains identity, revision number, snapshot hash, dependency
  references, and safe metadata;
- AgentFile content is rendered from the immutable snapshot artifact;
- avatar and documentation remain ordinary files;
- projection status is `pending`, `ready`, or `failed`.

Git failure does not reconstruct runtime state from cached Expert columns.
Export and marketplace publication may require projection status `ready`.

## Marketplace

A marketplace package contains:

- Expert metadata;
- a WorkerSpec template without organization-local IDs;
- required Worker type and protocol adapter constraints;
- Skill, knowledge, tool, and secret requirement descriptors;
- optional Workflow revisions;
- package version and content digest.

Installation resolves organization-local resources and creates a plan. It
either produces a runnable Expert revision or fails with explicit unresolved
requirements. The current direct `CreateExpertRequest` installation path is
removed.

## Workflow

### Definition

Workflow runtime identity is `worker_spec_snapshot_id`. Agent, model, Runner,
repository, environment bundle, permission, and raw config fields are removed
from the Workflow aggregate.

Prompt template, variables, trigger behavior, concurrency, timeout,
persistence, retention, and callback belong to Workflow revision.

### Trigger

Every manual, API, event, or Cron trigger:

1. derives a stable trigger key and locks the scheduling cursor;
2. reads the active revision and validates the execution principal;
3. validates input variables;
4. creates a WorkflowRun pinned to revision and snapshot;
5. advances the cursor and writes a launch outbox in the same transaction;
6. renders the prompt, records its hash, and launches idempotently.

Changing the Workflow later cannot change an already created WorkflowRun.

### Persistent Sessions

Persistent Workflow execution resumes only from the previous successful run
whose snapshot and session are compatible with the current Workflow revision.
If the active revision changes WorkerSpec, the next run starts a fresh session.

Workflow V1 requires `max_concurrent_runs = 1` when session persistence is
enabled and records an explicit predecessor run.

No compatibility fallback silently resumes an incompatible Pod.

## GoalLoop

GoalLoop already pins `worker_spec_snapshot_id`. The design adds:

- a run manifest link for every Pod used by the Loop;
- explicit invocation input for objective and current progress;
- immutable verification command and acceptance criteria after activation;
- current live authorization checks on every resume.

The Loop controller cannot change WorkerSpec, weaken verification, or select a
different Expert revision after activation.

## Mesh Ticket

Ticket assignment accepts either a WorkerSpec snapshot selection or an Expert
revision selection in the UI. The backend always stores the resolved snapshot
ID and optional source Expert revision ID.

`CreatePodForTicket` sends:

- snapshot ID;
- ticket correlation ID;
- task prompt derived from the Ticket;
- terminal dimensions and idempotency key.

Hardcoded Worker type, model, permission, and AgentFile defaults are deleted.

## Direct Worker and Session Entry Points

Every fresh Pod path must converge on plan or snapshot execution:

- Worker creation wizard;
- quick task;
- session create, import, fork, switch, MCP update, and rebuild;
- authorized host workspace binding;
- Ticket creation;
- coordinator launch;
- Runner MCP creation;
- Workflow and GoalLoop;
- Expert run;
- marketplace application.

Resume and fork use the source Pod snapshot. Fork may create a new invocation
but cannot mutate runtime configuration unless the user explicitly creates and
applies a new WorkerSpec plan.

## Ownership Matrix

| Concern | Owner |
| --- | --- |
| Harness executable and adapter | WorkerDefinitionRevision |
| Model and credential references | WorkerSpecSnapshot |
| Worker capabilities and workspace | WorkerSpecSnapshot |
| Expert identity and active version | Expert |
| Reusable task and trigger policy | WorkflowRevision |
| One goal and verifier | GoalLoop |
| Ticket task context | Mesh Ticket |
| Per-run rendered task input | Invocation / WorkflowRun |
| Effective runtime evidence | WorkerRunManifest |
| Pod lifecycle | Pod |
