# Aider Precreation Verification Loop

## Goal

Maintain a current, verifier-backed Aider integration record. The loop may verify
local contracts, images, Runner state, browser precreation, and database state
without approval. It must stop before credential reference injection, Pod
creation, provider requests, or cleanup.

## Cycle

1. Reload `worker.json`, `state.json`, `PROGRESS.md`, `DECISIONS.md`, and evidence.
2. Run `bash scripts/verify.sh`.
3. If current evidence is stale, repair the root cause and rerun focused checks.
4. If precreation is valid, require named non-production approval for the live gate.
5. Only after approval, run `bash scripts/verify-live-launch.sh`.

## Completion

Terminal success requires all acceptance items, including a disposable live launch
with PTY, provider result, and cleanup. `blocked_human_gate` is an honest
terminal pause, not success.

## Limits

- Maximum iterations: 6
- Maximum wall time: 120 minutes
- Maximum tokens: 60,000
- No progress: two identical task/verifier/blocker fingerprints

## Human Gate

The approval must name the non-production credential bundle and resource scope,
allow one disposable Pod, and allow cleanup. The loop must not read credential
values or send an external request before that record exists.
