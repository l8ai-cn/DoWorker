# Progress

Loop: Codex ACP Precreation Verification Loop

## Current Status

- Status: `blocked_human_gate`
- Active task: `verify-live-launch`
- Last verifier: `codex-precreation` passed
- Last fingerprint: `codex-cli|verify-live-launch|approval-missing`

## Verified

- Codex Definition requires an `openai-compatible` ModelBinding and supports ACP.
- The image digest, Docker version probe, and `dev-runner-codex` agree on `0.144.5`.
- Browser validation, plan, and apply created
  `codex-acp-precreation-template@r1` with WorkerSpec snapshot `5`.
- The snapshot pins `qa-primary-model` to `openai/gpt-5` through
  `openai-compatible`; no Worker launch or Pod exists.
- Personal runtime environment variables are plaintext preferences and explicitly
  cannot be used as credential evidence. The organization AI Resource page owns
  the model connection and credential-rotation surface.

## Blocker

No named non-production authorization exists for use of the organization model
resource in a disposable Codex ACP Worker. Do not create a Worker, use ACP, or
send a provider request until that approval is recorded.
