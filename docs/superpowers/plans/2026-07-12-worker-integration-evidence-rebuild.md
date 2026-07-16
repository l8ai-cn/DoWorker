# Worker Integration Evidence Rebuild Plan

Status: active
Supersedes: `2026-07-12-worker-onboarding-foundation.md` for execution decisions

## Goal

Prove, per Worker type, the complete path from a versioned definition through
Backend, Runner, image, credentials, Rust Core, and Web. A Worker is never
called supported until its recorded runtime and browser evidence passes.

## Scope

The target matrix has 12 slugs:

`aider`, `claude-code`, `codex-cli`, `cursor-cli`, `do-agent`, `gemini-cli`,
`grok-build`, `hermes`, `loopal`, `minimax-cli`, `openclaw`, `opencode`.

`e2e-echo` is test-only and excluded. No production image publication,
deployment, push, or real credential use is in scope.

## Completion Model

Each Worker has five independent evidence gates:

1. Definition: schema, credential references, configuration document contract,
   interaction modes, image target, and explicit adapter identifier.
2. Runtime: non-mock image builds, starts, and passes the declared version
   probe.
3. Runner: reports the exact executable and starts the declared PTY or ACP
   adapter without command-name inference.
4. Product path: Backend options/preflight/create-pod responses carry the
   exact immutable definition hash; Rust Core and Web consume that response.
5. Browser: an isolated fixture executes the create flow, including loading,
   disabled, invalid-credential, and backend-error states.

Only an explicit `verified` result for every applicable gate permits
`supported`. Missing infrastructure, credentials, or an unsupported upstream
tool produces `blocked` with a reason; it is not a fallback path.

## Execution Order

1. Reconcile the observed database, build scripts, runtime catalog, Runner
   imports, adapters, and Web forms into the durable evidence matrix.
2. Make the embedded Worker definition the single product input. Existing
   database AgentFiles may remain only as a checked projection during migration;
   a hash mismatch blocks creation.
3. Carry `slug`, definition version/hash, modes, and explicit `adapter_id`
   through Backend, Proto, Rust Core, Web, and Runner.
4. Remove product-path fallbacks: generic command-derived ACP selection,
   mock-binary substitution, and frontend credential-form defaults.
5. Add a guarded, deterministic dev E2E fixture with a dedicated user,
   organization, Runner, and disposable credential references. It must not
   create or overwrite normal development accounts.
6. Run Codex CLI and Gemini CLI as pilots. Their results determine the exact
   per-Worker run template, not a claim for other types.
7. Run each remaining Worker through its own evidence run, then independently
   review the matrix and terminal verifier.

## Deterministic Verifiers

- `verify-rebuild-state.sh`: no invalid prior acceptance remains and every
  target starts non-supported.
- `verify-definition-chain.sh`: catalog parsing, consumer wiring, and checked
  database projection.
- `verify-runtime-image.sh <slug>`: digest, non-mock executable, and declared
  version probe.
- `verify-runner-path.sh <slug>`: Runner probe and explicit adapter behavior.
- `verify-product-path.sh <slug>`: API snapshot and stale-hash rejection.
- `verify-browser-path.sh <slug>`: browser screenshot, console, network, and
  state assertions.
- `verify.sh`: terminal gate; it requires evidence for every target and cannot
  pass while any target is unverified or blocked.

## Stop And Escalate

Stop the current Worker run when a required command is unavailable, an image
cannot build, a vendor credential is required, a public Proto contract changes,
or the same verifier fails twice without a code or environment change. Record
the command, exit code, evidence path, and blocker. Escalate before using real
credentials, publishing images, changing a public API, or writing shared dev
data outside the dedicated E2E fixture.

## Test Standard

Static tests establish contract correctness only. A runtime gate requires an
actual container command and Runner event. A product gate requires an
authenticated API call. A UI gate requires a real browser run with console and
network inspection. Each test artifact must include the definition hash and
the exact command or request that produced it.
