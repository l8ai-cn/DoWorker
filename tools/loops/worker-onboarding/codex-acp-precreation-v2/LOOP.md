# Codex ACP Precreation Verification Loop

## Goal

Verify Codex CLI ACP through Definition, immutable image, connected Runner,
required model binding, browser precreation, and database state. Stop before any
provider-backed Worker creation.

## Cycle

1. Reload durable state and evidence.
2. Run `bash scripts/verify.sh`.
3. Repair any stale local fact at its root cause.
4. Request named non-production approval for the selected organization model resource.
5. After approval only, run the disposable live-launch path and
   `bash scripts/verify-live-launch.sh`.

## Completion

Only a passing live verifier can mark Codex supported. A missing approval is
`blocked_human_gate`, not success.
