# Progress

Loop: Worker Onboarding Loop Template

## Current Status

- Status: planned
- Active loop node: not started
- Active atomic task: not assigned
- Last verifier result: not run
- Last no-progress fingerprint: not recorded
- Current token estimate: 0

## Assumptions

- No assumptions recorded yet.

## Decisions

- Initial recursive loop contract generated from `loop.json`.

## Acceptance Trace

- Checklist path: `ACCEPTANCE.md`
- Checked items: none
- Reopened items: none

## Blocked Decision Trace

- Decision file: `DECISIONS.md`
- Decision log: `journal.jsonl`
- Delegation confirmed: pending
- Last proxy decision: none
- Last supervisor review: none

## Next Cycle

1. Re-read `loop.json`, `state.json`, `tasks.json`, `agents.json`, `ACCEPTANCE.md`, `DECISIONS.md`, and this file.
2. Evaluate `clarification_policy` before acting.
3. If blocked, apply `decision_policy`: confirm delegation, use proxy decisions only within low-risk authority, and escalate otherwise.
4. Run supervisor drift checks on the configured cadence before continuing.
5. Execute one eligible loop node or atomic task.
6. Record evidence, then check exactly the matching acceptance item only if its criteria and verifier refs pass.
7. Stop on success, failure, budget, no-progress, or human-gate conditions.
