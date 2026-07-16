# Acceptance Checklist

Loop: Worker Onboarding Loop Template

Update policy: Mark an item only after all criteria pass with durable evidence.

## Operating Rules

- Only change `[ ]` to `[x]` after every acceptance criterion passes.
- Evidence refs must point to existing artifacts, command outputs, screenshots, logs, diffs, or review notes.
- If later verification invalidates an item, change it back to `[ ]` and record the reason in `PROGRESS.md`.
- Terminal completion may only be claimed when every checklist item is checked and the terminal verifier passes.

## Items

- [ ] `accept-research-upstream`: Research the selected Worker from official sources.
  - Owner agent: `worker`
  - Atomic task: `research-upstream`
  - Acceptance criteria:
    - required research categories have official or captured-command evidence
    - unknown conclusions are recorded as blockers
  - Verification refs:
    - `research-evidence`
  - Evidence refs:
    - evidence/research/research.json
    - evidence/research/source-index.json

- [ ] `accept-define-worker-contract`: Define the selected Worker contract.
  - Owner agent: `worker`
  - Atomic task: `define-worker-contract`
  - Acceptance criteria:
    - Definition schema and AgentFile alignment pass
    - adapter_id and interaction modes are explicit
  - Verification refs:
    - `worker-contract`
  - Evidence refs:
    - artifacts/definition.json
    - artifacts/AgentFile
    - artifacts/schemas/

- [ ] `accept-define-credentials-config`: Define credentials and configuration documents.
  - Owner agent: `worker`
  - Atomic task: `define-credentials-config`
  - Acceptance criteria:
    - all secrets are references with declared env or document targets
    - form required and conditional behavior is declared
  - Verification refs:
    - `worker-contract`
  - Evidence refs:
    - artifacts/credential-bindings.json
    - artifacts/config-documents.json

- [ ] `accept-implement-frontend-backend`: Implement frontend and backend contract surfaces.
  - Owner agent: `worker`
  - Atomic task: `implement-frontend-backend`
  - Acceptance criteria:
    - Definition version and hash survive backend, proto, Rust Core, and Web mapping
    - form tests cover success, loading, error, disabled, and incompatible states
  - Verification refs:
    - `frontend-backend`
  - Evidence refs:
    - evidence/contracts/frontend-backend.json
    - evidence/tests/frontend-backend.txt

- [ ] `accept-implement-runner-runtime`: Implement Runner adapter and runtime support.
  - Owner agent: `worker`
  - Atomic task: `implement-runner-runtime`
  - Acceptance criteria:
    - Runner resolves exact adapter_id without fallback
    - runtime image uses a real pinned artifact and passes version probing
  - Verification refs:
    - `runner-runtime`
  - Evidence refs:
    - evidence/images/runtime-image.json
    - evidence/tests/runner-runtime.txt

- [ ] `accept-verify-worker-flow`: Verify the selected Worker end-to-end.
  - Owner agent: `reviewer`
  - Atomic task: `verify-worker-flow`
  - Acceptance criteria:
    - all deterministic Worker checks pass
    - browser evidence covers creation success and critical failure paths
  - Verification refs:
    - `worker-terminal`
  - Evidence refs:
    - evidence/tests/terminal.json
    - evidence/browser/worker-flow.json

- [ ] `accept-independent-review`: Independently accept or reject Worker evidence.
  - Owner agent: `reviewer`
  - Atomic task: `independent-review`
  - Acceptance criteria:
    - review confirms all evidence is verifier-backed
    - review confirms no fallback, mock artifact, or plaintext secret is accepted
  - Verification refs:
    - `worker-terminal`
  - Evidence refs:
    - evidence/review/acceptance.json
