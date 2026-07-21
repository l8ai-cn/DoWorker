# Definition Driven Worker Create Contract

## Purpose

Implement and verify the Definition-owned Worker creation contract without fallback paths.

User goal: Create many Worker types through Definition-specific adapters, credentials, model resources, and configuration documents, then verify the real product path.

Done definition: Named configuration bindings, compatible resource projection, Runner materialization, browser evidence, and authorized lifecycle evidence pass without inferred support claims.

## Clarification Policy

- Default action: `ask_user`
- Secondary user query: Approve named Definition configuration-document bindings as the replacement for positional config bundle IDs?
- Block if: a public API contract or database migration has no approval, a provider request would require a real credential
- Assumption record: `PROGRESS.md`

## Recursive Loop Topology

- `root`: Run the goal-based contract loop and stop at irreversible or unverified boundaries.

Decomposition strategy: definition_contract_then_per_worker_release_evidence

Split until:

- each task has one owner
- each task has deterministic evidence
- each Worker lifecycle has its own evidence record

## Atomic Tasks

- `current-contract-baseline`: Record the current typed and untyped Worker creation boundaries from source and real local evidence.
- `named-binding-approval`: Obtain an explicit decision to replace positional configuration bundle IDs with named Definition bindings.
- `named-binding-implementation`: Implement named configuration-document bindings throughout the canonical Worker create contract.
- `resource-projection`: Project compatible model resources and typed credential references from the selected Definition.
- `definition-driven-web`: Render Definition-declared documents and backend-filtered resources without static Worker compatibility maps.
- `runner-materialization`: Migrate Worker snapshots and preserve resolved Definition document targets into Runner materialization.
- `real-worker-release`: Verify authorized Worker preflight, lifecycle, cleanup, and browser paths with real non-production resources.

## Acceptance Checklist

- Path: `ACCEPTANCE.md`
- Update policy: Mark an item only after its verifier and durable evidence pass; a rejected or blocked gate remains unchecked.

- `accept-current-contract-baseline`: Record the authoritative current Worker create boundary and evidence state.
- `accept-named-binding-approval`: Obtain the public-contract decision for named configuration documents.
- `accept-named-binding-implementation`: Replace positional configuration bundle references with named Definition bindings.
- `accept-resource-projection`: Return only Definition-compatible model and credential references to the form.
- `accept-definition-driven-web`: Render named documents and server-filtered resources without static Worker mappings.
- `accept-runner-materialization`: Migrate snapshots and materialize Definition-owned document targets at Runner.
- `accept-real-worker-release`: Record authorized lifecycle, cleanup, and browser evidence for each eligible Worker.

## Blocked Execution And Decision Policy

- Decision file: `DECISIONS.md`
- Decision log: `journal.jsonl`
- Proxy decision agent: `decision-proxy` with `delegated_low_risk` authority
- Supervisor agent: `loop-supervisor` every 2 iteration(s)
- User confirmation: May the decision-proxy choose only low-risk evidence order and one read-only verifier retry?

## Agents

- `orchestrator` (orchestrator): maintain state, enforce gates, integrate verified changes
- `worker` (worker): implement one isolated contract slice, run focused checks
- `reviewer` (reviewer): reject incomplete evidence, review migration and API boundaries
- `decision-proxy` (decision_proxy): choose low-risk evidence sequencing, record blocked states
- `loop-supervisor` (supervisor): monitor drift, enforce no-progress exits, escalate gates

## Collaboration

- Patterns: orchestrator_workers, evaluator_optimizer, independent_reviewer
- Subagent activation: read-only audits are independent, review needs fresh context, write paths do not overlap
- Token policy: Subagents return commands, exit codes, changed paths, evidence paths, and unresolved questions under 2000 tokens.

## Context Strategy

- Max context tokens: 60000
- Retrieval: just_in_time
- Tool output trimming: Keep commands, exit codes, assertions, changed paths, and evidence references; exclude secret values and raw provider output.
- Compaction trigger: 0.8
- Durable memory: `state.json`, `journal.jsonl`, `PROGRESS.md`

## Termination Policy

- Success: all acceptance items are checked and independent review evidence exists
- Failure: a protected verifier would need weakening, the named binding contract is rejected, a required upstream artifact cannot be obtained
- Budget exits: {'max_iterations': 20, 'wall_clock_minutes': 360, 'max_tokens': 360000}
- No-progress fields: active_task_id, changed_paths, verifier_id, verifier_exit_code, blocker_code
- Human gates: before protobuf or database migration changes, before provider credential use, before browser mutation, push, merge, publish, or deploy

## Verification

- `contract-baseline`: `bash scripts/verify-contract-baseline.sh`
- `approval-gate`: `bash scripts/verify-contract-approval.sh`
- `focused-contract-tests`: `bash scripts/verify-focused-contract-tests.sh`

Protected paths:

- `loop.json`
- `scripts/verify.sh`
- `scripts/verify-contract-baseline.sh`
- `scripts/verify-contract-approval.sh`
- `scripts/verify-focused-contract-tests.sh`
- `tests`
- `.github/workflows`

## Human Gates

- approve public contract change
- run database migration
- use provider credentials
- create non-test Worker
- push, merge, publish, or deploy

## Escalation

- Condition: approval, credential, browser, budget, or no-progress gate is reached
- Owner: Agent Cloud maintainer
- Channel: Codex thread
- Message template: Definition-driven Worker loop stopped: {reason}. Evidence: {evidence_ref}.
