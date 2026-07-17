# Worker Onboarding Loop Template

## Purpose

Run one Do Worker type through research, contracts, implementation, verification, and independent review.

User goal: Bring one selected Worker slug to a verifiable, Definition-driven integration state.

Done definition: All seven Worker tasks have acceptance evidence and the Worker terminal verifier exits 0 for the selected slug.

## Clarification Policy

- Default action: `ask_user`
- Secondary user query: What approved decision should resolve this selected Worker's license, protocol, credential, or public-contract ambiguity?
- Block if: real credential use is required, upstream license or redistribution status is unclear, public API or migration changes lack approval
- Assumption record: `PROGRESS.md`

## Recursive Loop Topology

- `worker-goal`: Advance one Worker through explicit research, implementation, verification, and review stages.

Decomposition strategy: worker_contract_then_parallelizable_integration_then_independent_review

Split until:

- one task has one Worker slug
- one task has deterministic verifier evidence
- shared files have one writer

## Atomic Tasks

- `research-upstream`: Research the selected Worker from official sources and record installation, version, license, protocol, auth, configuration, persistence, and platform evidence.
- `define-worker-contract`: Define the selected Worker identity, adapter, modes, capabilities, AgentFile alignment, and configuration documents.
- `define-credentials-config`: Define credential references, required conditions, injection targets, configuration document schemas, and frontend field metadata.
- `implement-frontend-backend`: Implement the selected Worker Definition through backend APIs, proto and Rust Core mappings, and frontend form behavior.
- `implement-runner-runtime`: Implement the selected Worker Runner adapter, runtime image path, binary probe, and immutable catalog integration.
- `verify-worker-flow`: Run Definition, API, Runner, image, PTY or ACP, and browser checks for the selected Worker.
- `independent-review`: Independently review the selected Worker evidence and reject unsupported, secret-bearing, or fallback-based acceptance.

## Acceptance Checklist

- Path: `ACCEPTANCE.md`
- Update policy: Mark an item only after all criteria pass with durable evidence.

- `accept-research-upstream`: Research the selected Worker from official sources.
- `accept-define-worker-contract`: Define the selected Worker contract.
- `accept-define-credentials-config`: Define credentials and configuration documents.
- `accept-implement-frontend-backend`: Implement frontend and backend contract surfaces.
- `accept-implement-runner-runtime`: Implement Runner adapter and runtime support.
- `accept-verify-worker-flow`: Verify the selected Worker end-to-end.
- `accept-independent-review`: Independently accept or reject Worker evidence.

## Blocked Execution And Decision Policy

- Decision file: `DECISIONS.md`
- Decision log: `journal.jsonl`
- Proxy decision agent: `decision-proxy` with `delegated_low_risk` authority
- Supervisor agent: `loop-supervisor` every 2 iteration(s)
- User confirmation: May the decision proxy make only the listed low-risk sequencing and public-metadata retry decisions?

## Agents

- `orchestrator` (orchestrator): maintain selected Worker state, enforce task dependencies, enforce exits
- `worker` (worker): perform one scoped research or implementation task
- `reviewer` (reviewer): verify evidence independently, review browser proof, accept or reopen tasks
- `decision-proxy` (decision_proxy): resolve delegated low-risk ordering decisions
- `loop-supervisor` (supervisor): detect goal drift, enforce human gates

## Collaboration

- Patterns: orchestrator_workers, evaluator_optimizer, independent_reviewer
- Subagent activation: research, frontend-backend, and Runner-runtime tasks have disjoint write paths, independent review is required, one selected Worker exceeds primary context budget
- Token policy: Subagents return only decisions, changed paths, commands, exit codes, and evidence references under 2000 tokens.

## Context Strategy

- Max context tokens: 50000
- Retrieval: just_in_time
- Tool output trimming: Keep commands, exit codes, failing assertions, changed paths, and evidence references; exclude raw logs and secrets.
- Compaction trigger: 0.8
- Durable memory: `state.json`, `journal.jsonl`, `PROGRESS.md`

## Termination Policy

- Success: all seven Worker tasks are accepted, Worker terminal verifier exits 0, independent review evidence exists
- Failure: a real artifact is unavailable, license or protocol evidence remains unknown, a protected verifier must be weakened
- Budget exits: {'max_iterations': 8, 'wall_clock_minutes': 90, 'max_tokens': 120000}
- No-progress fields: active_verifier_ids, definition_hash, changed_paths, evidence_count, blocker_code
- Human gates: before real credential use, before license acceptance, before public contract change, before image publication, before push merge or deployment, before report-only promotion

## Verification

- `research-evidence`: `bash scripts/verify-research.sh`
- `worker-contract`: `bash scripts/verify-contract.sh`
- `frontend-backend`: `bash scripts/verify-api-contract.sh`
- `runner-runtime`: `bash scripts/verify-runner-runtime.sh`
- `worker-terminal`: `bash scripts/verify.sh`

Protected paths:

- `loop.json`
- `scripts/verify.sh`
- `scripts/verify-research.sh`
- `scripts/verify-contract.sh`
- `scripts/verify-api-contract.sh`
- `scripts/verify-runner-runtime.sh`
- `worker.json`
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
- Owner: Do Worker maintainer
- Channel: Codex thread
- Message template: Worker loop stopped for {worker_slug}: {reason}. Evidence: {evidence_ref}.
