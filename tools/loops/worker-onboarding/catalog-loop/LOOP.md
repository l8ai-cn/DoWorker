# Worker Integration Evidence Rebuild

## Purpose

Coordinate every formal Agent Cloud type through a runtime-evidence queue.

User goal: Create varied Workers with their own connections, adapters, credentials,
and configuration documents, then prove every product path with real evidence.

Done definition: Every target has Definition, image, Runner, product, and browser
evidence. A target can be `supported` only when all applicable gates pass.

## Clarification Policy

- Default action: `ask_user`
- Secondary user query: Which license, protocol, credential, or public-contract decision should govern this Worker?
- Block if: a real secret value is needed, an irreversible action is required, a product security boundary is unclear
- Assumption record: `PROGRESS.md`

## Recursive Loop Topology

- `rebuild-goal`: Select one bounded evidence task, preserve failures, and drive
  the queue to independent terminal review.

Decomposition strategy: catalog_queue_with_per_worker_goal_loops

Split until:

- one Worker slug has one isolated Worker Loop run
- each task has one owner
- each task has deterministic evidence

## Atomic Tasks

- `reconcile-runtime-evidence`: Record actual source and runtime gaps without support claims.
- `canonicalize-definition-chain`: Make the embedded Definition product-authoritative.
- `remove-hidden-product-fallbacks`: Remove adapter, binary, and form fallbacks.
- `provision-isolated-e2e-fixture`: Create the guarded authentication fixture.
- `pilot-codex`: Run the special-adapter pilot through the Codex CLI Worker path.
- `pilot-gemini`: Run the standard ACP pilot through the Gemini CLI Worker path.
- `process-worker-queue`: Instantiate and process one isolated Worker Loop run for every remaining formal slug.
- `verify-catalog`: Verify catalog-wide Definition, backend, Runner, image, and frontend consistency.
- `human-review`: Present the catalog evidence for independent human approval without publishing or deploying.

## Acceptance Checklist

- Path: `ACCEPTANCE.md`
- Update policy: Mark an item only after every criterion and verifier reference has durable evidence.

- `accept-runtime-evidence-baseline`: Record initial evidence without claims.
- `accept-canonical-definition-chain`: Establish the product definition chain.
- `accept-no-hidden-product-fallbacks`: Reject unsafe implicit behavior.
- `accept-isolated-e2e-fixture`: Establish authenticated test isolation.
- `accept-pilot-codex`: Complete the Codex special-adapter pilot.
- `accept-pilot-gemini`: Complete the Gemini standard ACP pilot.
- `accept-process-worker-queue`: Process every remaining formal Worker.
- `accept-verify-catalog`: Verify catalog-wide integration consistency.
- `accept-human-review`: Record the independent human review gate.

## Blocked Execution And Decision Policy

- Decision file: `DECISIONS.md`
- Decision log: `journal.jsonl`
- Proxy decision agent: `decision-proxy` with `delegated_low_risk` authority
- Supervisor agent: `loop-supervisor` every 2 iteration(s)
- User confirmation: May the decision proxy make only the listed low-risk queue and retry decisions?

## Agents

- `orchestrator` (orchestrator): maintain queue state, dispatch independent Worker runs, enforce exits
- `worker` (worker): execute one bounded implementation or research task
- `reviewer` (reviewer): independently verify evidence, reject incomplete acceptance
- `decision-proxy` (decision_proxy): resolve delegated low-risk scheduling decisions
- `loop-supervisor` (supervisor): monitor goal drift, stop invalid queue progress

## Collaboration

- Patterns: orchestrator_workers, evaluator_optimizer, independent_reviewer
- Subagent activation: Worker runs have no overlapping write paths, research and review need independent context, one Worker task exceeds the primary context budget
- Token policy: Subagents return only decisions, changed paths, commands, exit codes, and evidence references under 2000 tokens.

## Context Strategy

- Max context tokens: 60000
- Retrieval: just_in_time
- Tool output trimming: Keep commands, exit codes, failing assertions, changed paths, and evidence references; exclude raw logs and secrets.
- Compaction trigger: 0.8
- Durable memory: `state.json`, `journal.jsonl`, `PROGRESS.md`

## Termination Policy

- Success: all catalog tasks are accepted, the terminal verifier exits 0, and independent review is recorded
- Failure: required upstream artifact cannot be obtained, a protected verifier would need weakening, or a required public contract lacks approval
- Budget exits: `max_iterations=24`, `wall_clock_minutes=720`, `max_tokens=720000`
- No-progress fields: verifier ID, definition hash, changed paths, evidence count, blocker code
- Human gates: before real credential use, before public contract change, before image publication, before push merge or deployment, before report-only promotion

## Verification

- `rebuild-state`: `bash scripts/verify-rebuild-state.sh`
- `definition-chain`: `bash scripts/verify-definition-chain.sh`
- `catalog-terminal`: `bash scripts/verify.sh`

Protected paths:

- `loop.json`
- `scripts/verify.sh`
- `scripts/verify-rebuild-state.sh`
- `scripts/verify-definition-chain.sh`
- `catalog/formal-worker-slugs.txt`
- `.github/workflows`
- `tests`

## Human Gates

- use real credentials
- accept licensing terms
- change public contracts
- publish images
- push commits
- merge changes
- deploy environments

## Escalation

- Condition: A human gate, blocker, budget exit, or no-progress exit is reached.
- Owner: Agent Cloud maintainer
- Channel: Codex thread
- Message template: Worker catalog loop stopped: {reason}. Evidence: {evidence_ref}.
