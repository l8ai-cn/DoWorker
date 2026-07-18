# Acceptance Checklist

Loop: Definition Driven Worker Create Contract

Update policy: Mark an item only after its verifier and durable evidence pass; a rejected or blocked gate remains unchecked.

## Operating Rules

- Only change `[ ]` to `[x]` after every acceptance criterion passes.
- Evidence refs must point to existing artifacts, command outputs, screenshots, logs, diffs, or review notes.
- If later verification invalidates an item, change it back to `[ ]` and record the reason in `PROGRESS.md`.
- Terminal completion may only be claimed when every checklist item is checked and the terminal verifier passes.

## Items

- [x] `accept-current-contract-baseline`: Record the authoritative current Worker create boundary and evidence state.
  - Owner agent: `reviewer`
  - Atomic task: `current-contract-baseline`
  - Acceptance criteria:
    - baseline verifier exits 0
    - no Worker support claim changes
  - Verification refs:
    - `contract-baseline`
  - Evidence refs:
    - evidence/current-contract-baseline.json

- [x] `accept-named-binding-approval`: Obtain the public-contract decision for named configuration documents.
  - Owner agent: `orchestrator`
  - Atomic task: `named-binding-approval`
  - Acceptance criteria:
    - human decision is recorded
    - approval verifier exits 0
  - Verification refs:
    - `approval-gate`
  - Evidence refs:
    - evidence/named-binding-approval.json

- [x] `accept-named-binding-implementation`: Replace positional configuration bundle references with named Definition bindings.
  - Owner agent: `worker`
  - Atomic task: `named-binding-implementation`
  - Acceptance criteria:
    - focused contract tests exit 0
    - unknown, duplicate, wrong-kind, and malformed bindings fail
  - Verification refs:
    - `focused-contract-tests`
  - Evidence refs:
    - evidence/named-binding-implementation.json

- [x] `accept-resource-projection`: Return only Definition-compatible model and credential references to the form.
  - Owner agent: `worker`
  - Atomic task: `resource-projection`
  - Acceptance criteria:
    - focused contract tests exit 0
    - projection contains no secret values
  - Verification refs:
    - `focused-contract-tests`
  - Evidence refs:
    - evidence/resource-projection.json

- [x] `accept-definition-driven-web`: Render named documents and server-filtered resources without static Worker mappings.
  - Owner agent: `worker`
  - Atomic task: `definition-driven-web`
  - Acceptance criteria:
    - focused contract tests exit 0
    - stale draft references reset on Worker type change
  - Verification refs:
    - `focused-contract-tests`
  - Evidence refs:
    - evidence/definition-driven-web.json

- [x] `accept-runner-materialization`: Migrate snapshots and materialize Definition-owned document targets at Runner.
  - Owner agent: `worker`
  - Atomic task: `runner-materialization`
  - Acceptance criteria:
    - focused contract tests exit 0
    - Runner writes only resolved Definition targets
  - Verification refs:
    - `focused-contract-tests`
  - Evidence refs:
    - evidence/runner-materialization.json

- [ ] `accept-real-worker-release`: Record authorized lifecycle, cleanup, and browser evidence for each eligible Worker.
  - Owner agent: `reviewer`
  - Atomic task: `real-worker-release`
  - Acceptance criteria:
    - real lifecycle evidence is recorded
    - every promoted Worker has cleanup evidence
  - Evidence refs:
    - evidence/real-worker-release.json
  - Current evidence: OpenCode completed browser-created template, Pod, exact
    prompt response, and cleanup without a platform model binding. DoAgent
    completed cleanup but its model request failed at the unreachable provider
    endpoint. Cursor CLI completed ACP initialization but its unauthenticated
    `session/new` failed and cleaned up automatically. Grok Build was correctly
    blocked before creation because `XAI_API_KEY` is required. This item remains
    unchecked until every promoted Worker has equivalent cleanup evidence.
