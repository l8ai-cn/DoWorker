# Acceptance Checklist

Loop: Worker Integration Evidence Rebuild

Only mark an item after its verifier and every required runtime artifact pass.
An invalidated acceptance remains visible through the revocation record.

- [x] `accept-runtime-evidence-baseline`
  - Criteria: exactly 14 target slugs have durable baseline rows; every
    observed database/image/fixture mismatch is recorded without inferred support.
  - Verifier: `bash scripts/verify-rebuild-state.sh`
  - Evidence: `catalog/worker-evidence-matrix.json`,
    `evidence/revocations/2026-07-12-invalid-shared-contract.md`

- [x] `accept-canonical-definition-chain`
  - Criteria: Worker creation consumes the embedded definition; the database
    projection hash is checked; credential and configuration metadata survive.
  - Verifier: `bash scripts/verify-definition-chain.sh`
  - Evidence: focused Backend tests and per-slug definition evidence.

- [x] `accept-audit-worker-contracts`
  - Criteria: real command evidence corrects selected credential, launch, and
    adapter mappings without promoting an unsupported Worker.
  - Verifier: focused Backend, Web, Definition, and projection checks.
  - Evidence: `evidence/openclaw-noninteractive-launch.json` and journal entry
    `worker_definition_projection_and_contract_reconciled`.

- [ ] `accept-no-hidden-product-fallbacks`
  - Criteria: unknown adapters fail, mock binaries are absent, and missing form
    definitions produce a visible error rather than a fallback form.
  - Verifier: adapter, image, and Web contract tests.

- [ ] `accept-isolated-e2e-fixture`
  - Criteria: a guarded fixture creates only disposable test identities and
    allows authenticated Browser/API validation.
  - Verifier: fixture bootstrap and teardown checks.

- [ ] `accept-pilot-codex`
  - Criteria: Codex image, Runner, API, Core, and browser gates pass.
  - Verifier: `bash scripts/verify-worker-run.sh codex-cli`

- [ ] `accept-pilot-gemini`
  - Criteria: Gemini image, explicit generic ACP adapter, API, Core, and
    browser gates pass.
  - Verifier: `bash scripts/verify-worker-run.sh gemini-cli`

- [ ] `accept-worker-queue`
  - Criteria: every remaining slug is either fully verified or explicitly
    blocked with reproducible evidence; no status is inferred.
  - Verifier: `bash scripts/verify-worker-runs.sh`

- [ ] `accept-catalog-terminal`
  - Criteria: all product and browser gates pass, the terminal verifier exits
    zero, and independent review records remaining human gates.
  - Verifier: `bash scripts/verify.sh`
