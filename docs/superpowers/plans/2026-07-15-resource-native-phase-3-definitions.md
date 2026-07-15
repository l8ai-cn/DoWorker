# Resource-Native Phase 3: Referenced Definitions

## Goal

Add strict resource schemas and planners for `Prompt`, `Worker`, `Expert`,
`Workflow`, and `GoalLoop`, while preserving `WorkerSpecSnapshot` as the only
runtime configuration truth.

## Decisions

- `Prompt` is a versioned resource containing prompt text and declared
  variables. It never contains credentials.
- `Worker`, `Expert`, `Workflow`, and `GoalLoop` reference a
  `WorkerTemplate` revision. Plan resolves that revision to its immutable
  `worker_spec_snapshot_id`.
- `Expert` and `Workflow` may reference a `Prompt` revision.
- Invocation input stays outside Worker capability. A `Worker` may carry an
  alias and prompt inputs, but cannot override model, tools, environment,
  placement, or lifecycle.
- Workflow scheduling, concurrency, timeout, retention, and callback behavior
  belong to the Workflow resource, not the Worker snapshot.
- GoalLoop objective, verifier, budgets, no-progress limits, and escalation
  policy are immutable after typed Apply.

## Resource Specs

### Prompt

```yaml
spec:
  content: "Review {{change}}"
  variables:
    change:
      required: true
      default: ""
```

### Worker

```yaml
spec:
  workerTemplateRef: {kind: WorkerTemplate, name: codex-reviewer}
  promptRef: {kind: Prompt, name: review-task}
  inputs: {change: "PR-42"}
  alias: reviewer-42
```

### Expert

```yaml
spec:
  workerTemplateRef: {kind: WorkerTemplate, name: codex-reviewer}
  promptRef: {kind: Prompt, name: review-system}
  description: Reviews code changes
  category: engineering
  releaseNotes: Initial revision
```

### Workflow

```yaml
spec:
  workerTemplateRef: {kind: WorkerTemplate, name: codex-reviewer}
  promptRef: {kind: Prompt, name: nightly-review}
  inputs: {}
  executionMode: direct
  cronExpression: "0 2 * * *"
  sandboxStrategy: fresh
  sessionPersistence: false
  concurrencyPolicy: skip
  maxConcurrentRuns: 1
  maxRetainedRuns: 30
  timeoutMinutes: 60
  idleTimeoutSeconds: 30
```

### GoalLoop

```yaml
spec:
  workerTemplateRef: {kind: WorkerTemplate, name: codex-reviewer}
  objective: Fix checkout
  acceptanceCriteria: [Tests pass]
  verificationCommand: go test ./...
  maxIterations: 10
  timeoutMinutes: 60
  noProgressLimit: 3
  sameErrorLimit: 2
  escalationPolicy: pause
```

## Implementation

1. Register strict schemas with unknown-field, identifier, size, enum, and
   cross-field validation.
2. Add deterministic reference extraction for every new kind.
3. Extend the revision resolver to return a pinned WorkerTemplate snapshot.
4. Add planners that emit typed artifacts containing only safe domain data,
   pinned snapshot IDs, and pinned Prompt identities.
5. Add focused domain, planner, stale-reference, cross-tenant, and secret
   rejection tests.
6. Update product documentation only after the schema and planner tests pass.

## Verification

```bash
go test ./backend/internal/domain/orchestrationresource/... -count=1
go test ./backend/internal/service/orchestrationworker/... -count=1
go test ./backend/internal/service/orchestrationcontrol/... -count=1
go test -race ./backend/internal/service/orchestrationworker/... -count=1
```

## Exit Gate

- All five new Kinds validate and Plan.
- Every Worker-capable Plan contains a positive snapshot ID from the exact
  pinned WorkerTemplate revision.
- Prompt references are pinned and no prompt or error path accepts secret-like
  fields.
- No typed Apply or runtime behavior is claimed complete in this phase.
