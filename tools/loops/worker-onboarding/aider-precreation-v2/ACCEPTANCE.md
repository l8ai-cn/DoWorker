# Acceptance Checklist

Loop: Aider Precreation Verification Loop

- [x] `accept-contract`: Aider uses `aider-pty` with PTY; model binding is optional.
  - Evidence: `evidence/aider-precreation-2026-07-16.json`
  - Verifier: `bash scripts/verify.sh`
- [x] `accept-runtime`: Catalog, Docker image, and Runner report the current Aider runtime.
  - Evidence: `evidence/aider-precreation-2026-07-16.json`
  - Verifier: `bash scripts/verify.sh`
- [x] `accept-precreation`: Browser validation, plan, and template apply passed; no launch or Pod exists.
  - Evidence: `evidence/aider-precreation-2026-07-16.json`
  - Verifier: `bash scripts/verify.sh`
- [x] `accept-review`: Evidence is precreation-only and contains no plaintext credential or provider result.
  - Evidence: `evidence/aider-precreation-review-2026-07-16.json`
  - Verifier: `bash scripts/verify.sh`
- [ ] `accept-live-launch`: Named approval, reference injection, disposable Pod, PTY, provider smoke, and cleanup.
  - Evidence: `evidence/aider-live-launch-approval-and-result.json`
  - Verifier: `bash scripts/verify-live-launch.sh`

The unchecked live-launch item blocks terminal success and any supported claim.
