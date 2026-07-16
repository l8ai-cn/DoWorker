# Control Plane Protocol

## Validate

`ValidateWorkerSpec` is deterministic and side-effect free.

Input:

- organization slug for scoped schema selection;
- `WorkerSpecDraft`;
- options revision.

Checks:

- strict decoding and unknown-field rejection;
- identifier, enum, range, and required-field rules;
- Worker type configuration schema;
- mutually exclusive value and secret reference fields;
- internal consistency such as branch requiring repository;
- allowed interaction and automation combinations.

Validation does not select a Runner, check capacity, decrypt secrets, create a
snapshot, or claim that external resources remain available.

## Plan

`PlanWorkerSpec` performs full resolution.

Input:

- validated draft;
- optional comparison snapshot ID;
- required intent such as `create_worker`, `revise_expert`, or
  `revise_workflow`;
- target ID and expected revision for updates.

Output:

- `plan_id`, `plan_hash`, and expiration;
- canonical proposed WorkerSpec and safe summary;
- path-addressed diff against the comparison snapshot;
- blocking issues and advisory warnings;
- options and policy revisions.

The default plan lifetime is 15 minutes. Its state machine is
`ready -> applied | cancelled | expired`. Apply locks the plan row, so only one
concurrent caller can leave `ready`.

Blocking checks include:

- organization ownership and authorization;
- exact WorkerDefinition and runtime image compatibility;
- model protocol adapter compatibility;
- repository, Skill, knowledge, and environment references;
- secret reference authorization without value exposure;
- compute target capability and resource profile compatibility;
- enforceable quota and policy constraints.

Warnings are non-blocking observations only. Missing or malformed resources are
never warnings.

## Diff Contract

Diff entries contain:

```text
path
change_type: add | remove | replace
before_summary
after_summary
sensitivity: public | reference
impact: metadata | restart | authorization
```

Secret values are represented as reference identity changes. Diffs never reveal
decrypted values, credential prefixes, or protected environment values.

## Apply

There is no generic operation that mutates every target. Target services
consume the same plan contract:

- `CreateWorkerFromPlan`
- `CreateExpertFromPlan`
- `CreateExpertRevisionFromPlan`
- `CreateWorkflowFromPlan`
- `CreateWorkflowRevisionFromPlan`
- `CreateGoalLoopFromPlan`
- `BindTicketWorkerFromPlan`

Each request includes:

- `plan_id` and `plan_hash`;
- idempotency key;
- expected current target revision when updating;
- target metadata or task definition that is outside WorkerSpec.

Idempotency keys are unique within organization, caller, operation, and target.
Reusing a key with different request hashes is rejected.

Apply rejects:

- expired, cancelled, already consumed, or wrong-organization plans;
- plan hash mismatch;
- a changed security policy revision requiring a new plan;
- target revision compare-and-swap failure;
- any legacy runtime configuration field.

A catalog options revision changing after plan creation does not by itself
invalidate exact pinned revisions. Apply still rechecks their availability,
revocation, and authorization.

## Transaction Boundary

Within one database transaction, apply:

1. locks the plan;
2. verifies scope, status, hash, expiration, and expected target revision;
3. creates or reuses the content-addressed WorkerSpec snapshot;
4. creates the target revision or durable Worker launch intent;
5. updates the active revision pointer when required;
6. marks the plan applied with snapshot and operation result references;
7. writes an outbox event.

External Git, Runner, and notification calls occur after commit through durable
outbox or pending-command workers.

## Worker Creation Apply

Worker creation apply persists the snapshot, preallocated Pod key, and
`worker_launch_intent`. A leased worker materializes and dispatches it through
the durable command path described in
[Reliability and Security](08-reliability-and-security.md).

The same idempotency key returns the same launch intent and eventual Pod or
terminal failure. It cannot create a second launch.

## Destroy Semantics

Terraform vocabulary is not copied where product semantics differ:

- Worker: `TerminatePod`, followed by retention cleanup policy.
- Expert: `ArchiveExpert`; revisions remain referenced and auditable.
- Workflow: `DisableWorkflow` or `ArchiveWorkflow`; active runs are handled by
  explicit cancellation policy.
- GoalLoop: `CancelGoalLoop`.
- Snapshot: no user destroy while referenced; retention removes only
  unreferenced artifacts.

Generic `destroy` would hide these distinct lifecycle and audit requirements
and is therefore out of scope.

## Error Codes

Stable machine-readable codes include:

- `invalid-draft`
- `stale-options`
- `plan-expired`
- `plan-hash-mismatch`
- `plan-already-applied`
- `target-revision-conflict`
- `definition-revoked`
- `definition-revision-unavailable`
- `resource-revision-unavailable`
- `authorization-denied`
- `policy-changed`
- `capacity-unavailable`
- `quota-exceeded`
- `migration-required`
- `conflicting-runtime-input`

Messages may be localized, but clients branch only on codes and field paths.

## API Evolution

The existing `PreflightWorker` can become `PlanWorkerSpec` after clients migrate.
During one release, both procedure names may map to the same new planner only
for transport migration. They must not implement different runtime semantics.

`CreatePodRequest.worker_spec` is replaced by `worker_plan_id` and
`worker_plan_hash`. Snapshot-backed internal calls continue accepting only
`worker_spec_snapshot_id` plus allowed invocation fields.
