# Resource-Native Phase 4: Typed Domain Apply

## Goal

Consume Phase 3 plans through explicit typed operations and make fresh Worker,
Expert, Workflow, and GoalLoop execution use pinned immutable revisions.

## Storage

- Resource revisions remain the immutable declaration history.
- Expert and Workflow mutable rows retain identity, state, counters, and an
  active resource revision pointer.
- Every Expert and Workflow resource revision stores
  `worker_spec_snapshot_id`.
- Every WorkflowRun stores the exact Workflow resource revision and snapshot.
- GoalLoop already stores `worker_spec_snapshot_id`; it gains its originating
  resource revision.
- Worker Apply stores a one-shot resource revision and creates a Pod from the
  pinned snapshot using an idempotency key derived from the plan.

## Typed Operations

```text
ApplyPromptPlan
CreateWorkerFromPlan
ApplyExpertPlan
ApplyWorkflowPlan
CreateGoalLoopFromPlan
```

No generic mutation selects behavior from an untrusted Kind.

## Transaction Rules

- Plan locking, staleness checks, resource revision insertion, active pointer
  update, and domain row creation happen in one database transaction.
- Worker launch uses a durable outbox/idempotency record because Runner
  dispatch cannot participate in the database transaction.
- Replaying an applied plan returns the same typed result.
- Apply verifies artifact kind, canonical hash, actor, tenant, base
  `resourceVersion`, expiry, and exact snapshot identity.

## Runtime Cutover

- Expert Run reads its active resource revision and snapshot.
- Workflow trigger pins the active resource revision and snapshot before
  enqueueing launch.
- GoalLoop Start uses the snapshot stored at creation.
- Worker creation from a resource Plan passes only snapshot ID and invocation
  input to the existing snapshot materializer.
- Mutable legacy Agent, model, repository, environment, and AgentFile fields
  are not consulted on these new paths.

## Verification

```bash
go test ./backend/internal/service/expert/... ./backend/internal/service/workflow/... ./backend/internal/service/goalloop/... -count=1
go test ./backend/internal/infra/... -run 'Orchestration|ExpertRevision|WorkflowRevision|GoalLoopResource' -count=1
go test -race ./backend/internal/service/orchestrationworker/... -count=1
go test ./backend/internal/service/agentpod/... -run 'FreshExecutionInventory|WorkerSpec' -count=1
```

Browser verification must cover:

1. create and apply a Prompt;
2. create an Expert from a pinned WorkerTemplate and run it;
3. create a Workflow, update its active revision, and prove an older run stays
   pinned;
4. create and start a GoalLoop from a Plan;
5. edit a draft after Plan and prove Apply is blocked;
6. attempt unauthorized and expired Apply and prove explicit failure.

## Exit Gate

- All typed Apply operations are transactional and idempotent.
- Every fresh domain execution stores the exact resource revision and
  WorkerSpec snapshot.
- Existing runtime builders are unreachable from the new resource paths.
- Migration rehearsal, API/Rust/Web bindings, user docs, and browser evidence
  pass before the feature is called complete.

The remaining non-domain entry points are specified in
`2026-07-16-resource-native-phase-4b-execution-cutover.md`.
