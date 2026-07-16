# Progress

Loop: Worker Integration Evidence Rebuild

## Current Status

- Status: budget exhausted
- Active atomic task: none
- Current plan: Worker Integration Evidence Rebuild Plan
  (`docs/superpowers/plans/2026-07-12-worker-integration-evidence-rebuild.md`)
- Last verifier: `loop-control-plane-consistency` passed
- Budget exit: the durable token estimate reached the configured `720000` cap;
  no new Worker task may start until the maintainer explicitly re-budgets it.
- Current support claims: no Worker is formally deployable; `codex-cli` is
  verified in local development only

## Current Cycle Evidence

- The official runtime catalog is release-gated: all three historical
  published digests remain declared for audit but are disabled because actual
  registry pulls return `not_found`. The default Worker wizard exposes no
  formal deployable Worker.
- Local development generates an explicit ignored catalog from actual Docker
  image IDs only. It currently contains Codex CLI and Gemini CLI, and starts
  only `runner-codex-cli` and `runner-gemini-cli`; it never promotes a mutable
  local tag to the official catalog.
- `./deploy/dev/dev.sh --backend-only` initially failed because it built every
  runner, including the rejected Loopal mock sidecar. The dev lifecycle now
  builds only services represented in the explicit local catalog. Codex and
  Gemini both registered over mTLS and were detected as Codex `0.144.1` and
  Gemini `0.50.0`.
- The current Chromium suite uses the project auth setup rather than the
  hydration-sensitive login form. It created a Codex Worker, acquired control,
  received ACP `READY`, and cleaned up. It also showed Gemini's missing-model
  resource error, disabled Next, and visibly blocked all ten unavailable
  Worker types.
- The release gate, Go package tests, frontend typecheck, frontend lint,
  runtime lock probes, Kustomize render, generated docs checks, and public
  docs Chromium checks passed. Frontend lint has 181 pre-existing warnings and
  no errors.
- The local Gemini CLI `0.50.0` bundle distinguishes Gemini API
  `GEMINI_API_KEY` from Vertex AI Express `GOOGLE_API_KEY`. The platform's
  `gemini` model resource targets `generativelanguage.googleapis.com`, so
  Gemini CLI Definition version `2`, AgentFile, legacy provider mapping, and
  create-time injection now consistently use `GEMINI_API_KEY`. Vertex remains
  unsupported because its project, location, and credential contract is not
  declared.
- OpenClaw `2026.6.11` passes only its version probe. The formal bare
  `openclaw` AgentFile command exits in non-interactive execution because it
  enters onboarding, so its formal launch contract is blocked. The platform
  does not infer a provider, model, `agent --local` invocation, or credential
  fallback from upstream help text.
- OpenClaw and Hermes legacy model paths, model-pool preferences, and Web
  filtering now accept only the OpenAI-compatible protocol declared in their
  formal Definitions. The removed Anthropic/Gemini injection branch had no
  matching formal AgentFile configuration.
- A Definition/AgentFile update requires the documented
  `worker-definition-sync` development transaction to refresh the `agents`
  projection. The current sync repaired the live Gemini mismatch, after which
  the formal Chromium wizard test again selected Gemini and verified its
  missing-model-resource block.
- The active execution plan, state machine, and generated loop controls had
  diverged after the evidence-rebuild redesign. `loop.json`, task graph, and
  monitoring now use the rebuild identity; the stale duplicate manifest was
  removed. This was recorded as a budget exit, not a completed catalog.

## Reopened Acceptance

The former inventory and shared-contract acceptance were invalidated on
2026-07-12. The reason and affected product paths are recorded in
`evidence/revocations/2026-07-12-invalid-shared-contract.md`.

## Observed Baseline

- An explicit `worker-definition-sync` transaction now projects all twelve
  formal Definitions into `agents`. It repaired Claude Code and MiniMax
  AgentFile/adapter drift and created Hermes/OpenClaw without copying legacy
  migration text.
- Development starts this projection sync after migration and seed. The
  Kubernetes deployment path runs the same command as a Job before workload
  rollout, using the backend Deployment's immutable image reference.
- The runtime catalog is an embedded, versioned digest lock. The three entries
  previously marked enabled (Codex CLI, Claude Code, Gemini CLI) all fail an
  actual registry pull with `not found`; none is a deployable published
  runtime. The catalog's enabled state, coordinator image mapping, and
  Kubernetes manifests currently use different references and must be
  reconciled before any production claim.
- The build script now names all twelve declared targets. Cursor, Claude,
  Gemini, OpenCode, Grok, and MiniMax have current `linux/amd64` image and
  binary smoke evidence. This is not product-path evidence.
- The old catalog hash omitted AgentFile bytes. Definition hashes now cover the
  JSON document and AgentFile bundle; worker creation consumes this bundle and
  checks the database projection.
- Authenticated API E2E is available and EnvBundle API checks pass. The formal
  Worker browser test is still unverified because the old Pod-create test uses
  a removed `#worker-image-select` control instead of the Worker wizard.
- `codex-cli` now has a real `linux/amd64` image build and CLI smoke result:
  Codex CLI `0.144.1` ran as the non-root `runner` user, its configuration
  directory existed, and the image contained no substituted mock sidecars.
  This is runtime evidence only; the Runner control path and product flow
  remain unverified.
- `cursor-cli` initially copied only its launcher, which failed because the
  launcher requires its sibling Cursor runtime tree. The image now copies that
  tree to `/opt/cursor-agent`; `agent --version` succeeds.
- Aider image construction is blocked by repeated Debian apt repository 502
  responses while installing `python3-pip`. The Dockerfile retries the actual
  package installation eight times and exits on failure; no substitute image
  or mock binary was introduced.
- Credential-bundle forms now only expose fields declared as
  `credential_bundle` in the formal Worker Definition. Model-resource fields
  are removed at the Backend schema boundary, so the UI cannot save a bundle
  that the WorkerSpec resolver later rejects.
- `loopal` was found to stage an ignored local binary containing
  `runner/internal/agents/mockagent`. Its image build now requires an explicit
  non-mock `LOOPAL_BINARY`; without a real artifact its runtime is blocked.
- `do-agent` has a real `linux/amd64` image smoke result with version `0.2.7`
  and a passing Runner adapter unit test. Its model-resource and product-path
  gates remain unverified.
- OpenClaw has a real `linux/amd64` image smoke result with version
  `2026.6.11`; its database projection now exists, but it has no published
  immutable runtime image lock entry.
- Hermes exhausts the same bounded Debian apt retry path as Aider and remains
  blocked by repeated HTTP 502 responses.
- The live `ListWorkerCreateOptions` response returns all twelve definitions.
  With the official catalog it marks every formal Worker unavailable. In local
  development, only Codex CLI and Gemini CLI become selectable through the
  explicit generated catalog.
- Browser validation now proves the formal Worker wizard consumes the current
  Rust Core contract: Codex reaches server preflight without creating a Worker;
  Claude and Gemini show an explicit missing-model-resource error and disable
  the next step; every runtime-image-blocked type is visibly unavailable.
- The Rust `prost` generated `WorkerTypeOption` had discarded
  `requires_model_resource` because the generated Core code lagged the proto
  source. A raw protobuf round-trip test now protects field 8. The browser had
  previously hidden the model resource control and allowed progress.
- `scripts/build-wasm.sh` now builds into a temporary directory and publishes
  static assets before the JS entrypoint. Reloading the running Worker wizard
  during and after `pnpm run build:wasm` no longer produced a
  `do-worker-wasm` module-not-found failure.
- A disposable local Codex Worker was created from the real four-step wizard
  and received the ACP prompt `Reply with exactly READY. Do not modify files.`
  It moved to Executing, returned `READY`, and was terminated through
  `PodService/TerminatePod`; both API and browser then showed it was no longer
  active. Its product path is verified locally, but the catalog remains
  unpromoted until the formal Playwright flow is current and repeatable.
- The independent formal Worker wizard Playwright test now logs in without the
  stale global storage snapshot, creates Codex, acquires control, receives
  `READY`, and terminates its own Worker. `codex-cli` is therefore
  `verified_local_dev`; it is not a production or remote-deployment claim.
- A second formal Worker wizard Playwright test proves Gemini cannot advance
  without a compatible model resource and lists the ten missing-image Worker
  types. The Gemini run is explicitly blocked by the absent real resource; no
  fake provider endpoint or credential bundle was created.
- Public Next.js documentation now uses a generated Worker runtime catalog
  compiled from the formal Definitions, immutable runtime lock, and actual
  pull evidence. It presents 12 formal definitions, 0 formally deployable
  Workers, 1 local-development proof blocked from release, and 12 release
  blockers rather than a static supported-tool list.
- Docs navigation now exposes Worker Types & Runtime and removes the obsolete
  AgentPod and Mesh entries. Legacy paths redirect to the current Worker
  pages. The current four-step creation flow, model-resource references,
  credential bundles, typed configuration, and preflight are documented and
  covered by unauthenticated desktop and 390px Chromium checks.
- Runner tunnel events now have a complete state path: the gRPC adapter
  persists connected/disconnected events, the service synchronizes the active
  Runner cache, Connect exposes the fields, the WASM/cache projections retain
  them, and the Runner detail page renders their status and latest
  confirmation time.
- Live Gemini Runner `19` reported `connected` with
  `tunnel_last_seen_at=2026-07-12T22:24:18Z`. The PostgreSQL row, authenticated
  binary Connect `GetRunner` response, and browser detail page agreed. The
  first live check exposed an active-cache stale-read defect; a red service
  test reproduced it before the cache synchronization fix.
- The public documentation header no longer constrains itself to a 1536px
  centered container. On a 2048px viewport its logo now aligns with the
  viewport/sidebar gutter; Chromium checks cover both this wide layout and
  the 390px layout.
- The former external REST Worker guide incorrectly described legacy
  `agent_slug` and raw AgentFile fields as the formal Worker API. It now
  distinguishes the four-step Connect `WorkerSpecDraft` workflow from the
  legacy external REST endpoint, which has no formal API-key equivalent.
- `WorkerSpec.runtime.image` is persisted as a formal selection but never
  reaches the Runner create command. Platform-managed runner startup instead
  consumes a separate `COORDINATOR_RUNNER_IMAGES` mapping. This is a model
  boundary defect: self-hosted runner capability and managed-container release
  selection are conflated and have no common verifier.

## Next Cycle

1. Publish a real immutable image digest for a Worker only after its image,
   Runner, credential, product-path, and browser evidence all pass. Enable it
   in the official catalog and managed deployment mapping in the same change.
2. Run a real Runner create/preflight/ACP/cleanup path for Gemini CLI only
   after an authorized compatible Gemini API model resource is created in the
   product. Do not create a disposable endpoint, fake credential, or
   support-status promotion.
3. Obtain a versioned, real Loopal artifact and resolve Aider/Hermes package
   repository failures through their real upstream paths.
4. Redesign the OpenClaw formal AgentFile only after proving an exact
   non-interactive command, state directory, provider/model mapping, and
   prompt/cleanup behavior against a real credential.
