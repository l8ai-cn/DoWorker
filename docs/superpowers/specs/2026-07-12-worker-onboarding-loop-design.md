# Worker Onboarding Recursive Loop Design

## Goal
Generate a reusable engineering loop that brings every formal Agent Cloud type to
a verifiable state. It covers upstream research, Worker definition, credentials,
configuration documents, frontend forms, backend contracts, Runner adapters,
runtime images, and deterministic verification.

The output contains two `RecursiveLoopOrchestration` v2.0 projects:

```text
tools/loops/worker-onboarding/
  catalog-loop/
  worker-loop-template/
```

`catalog-loop` owns inventory, shared prerequisites, pilots, the Worker queue,
and terminal verification. `worker-loop-template` is instantiated once per
Worker slug and owns one full integration cycle.

These are engineering orchestration artifacts, not the product `Loop` domain.
They do not publish images, push commits, merge pull requests, or deploy.

## Inventory And SSOT
Formal Workers:

```text
claude-code
codex-cli
gemini-cli
aider
opencode
loopal
cursor-cli
do-agent
minimax-cli
grok-build
openclaw
hermes
```

`e2e-echo` is test-only and excluded from catalog completion.

Bootstrap discovery compares Agent migrations, Runner registrations, runtime
build scripts, runtime catalogs, and frontend credential metadata. The loop
reports drift but cannot choose a legacy source as authoritative. After the
shared-contract task is accepted, the Git source of truth is:

```text
config/worker-types/catalog.json
config/worker-types/<slug>/definition.json
config/worker-types/<slug>/AgentFile
config/worker-types/<slug>/schemas/
```

## Worker Contract
Each versioned `definition.json` declares:

```text
identity
runtime: executable, adapter_id, interaction_modes, version_probe
configuration: fields, config_documents
credentials: bindings
capabilities
image
verification
```

Backend dispatch carries `worker_type_slug`, `adapter_id`, and
`definition_hash`. Runner resolves that exact adapter or fails. It must not infer
a transport from the executable or fall back to ACP.

A configuration document declares ID, path, format, schema, merge policy, file
mode, and template data. `.json` path heuristics are not a valid contract.

## Credentials, Forms, And API
Credential bindings contain an ID, source kind, required condition, target, and
display metadata. Source kinds are `model_resource` and `credential_bundle`;
targets are `env` and `config_document`.

WorkerSpec stores references only. Backend validates authorization and
compatibility, resolves references immediately before dispatch, and sends
materialized runtime values through the authenticated Runner control channel.
Plaintext values must not enter WorkerSpec, Definition responses, evidence, or
durable Loop state.

The frontend renders configuration from the Definition API. It does not use
per-slug fallback lists or manually duplicated fields. Each field includes
required state, visibility conditions, options, secret-reference semantics,
validation errors, and help text.

| Operation | Responsibility |
| --- | --- |
| `ListWorkerTypes` | Selectable summaries and availability reasons. |
| `GetWorkerTypeDefinition` | Versioned form and runtime contract. |
| `GetWorkerCreationOptions` | Compatible models, images, targets, and profiles. |
| `PreflightWorkerDraft` | Validate values, refs, documents, authorization, and compatibility. |
| `CreateWorker` | Persist only a preflight-valid WorkerSpec and immutable snapshot. |

Proto, Connect handlers, Rust Core, and Web view models carry Definition version
and hash. The browser never assembles an executable command.

## Catalog Loop
The Catalog Loop uses a goal primitive and persists each observation:

```text
discovering -> foundation -> piloting -> processing-workers
            -> catalog-verifying -> ready-for-review
```

Its atomic tasks are:

1. Discover inventory and legacy drift.
2. Establish shared Definition, credential, API, and verifier contracts.
3. Run the Codex special-adapter pilot.
4. Run the Gemini standard-ACP pilot.
5. Instantiate and monitor one Worker Loop per remaining catalog entry.
6. Verify catalog-wide consistency.
7. Stop for independent human review.

Catalog success requires one accepted Worker Loop instance per formal slug, no
unresolved drift, and zero exit codes from catalog checks.

## Worker Loop

Each Worker Loop uses a goal primitive with seven atomic tasks:

1. Research installation, version, license, protocol, auth, configuration,
   persistence, and platform support.
2. Define the Worker contract and AgentFile alignment.
3. Define credential bindings and configuration documents.
4. Implement frontend and backend contract surfaces.
5. Implement Runner adapter and runtime image support.
6. Run contract, image, smoke, and browser verification.
7. Run independent review and record acceptance evidence.

Per-Worker state is isolated:

```text
tools/loops/worker-onboarding/runs/<slug>/
  state.json
  PROGRESS.md
  ACCEPTANCE.md
  DECISIONS.md
  journal.jsonl
  evidence/
```

Research records source URL, source kind, check date, upstream version, evidence
hash, and conclusion. Protocol, installation, license, and configuration facts
need official documentation, an official repository or package registry, or a
captured command result. `unknown` blocks implementation.

## Verification

The Worker terminal verifier runs:

```text
research schema
definition and AgentFile schema
credential reference safety
API and proto contract tests
Runner adapter contract tests
runtime image build, binary probe, and digest check
PTY or ACP smoke tests
browser creation-form E2E
independent review evidence
```

The catalog verifier compares Definition, database, Runner, runtime catalog,
image build list, and frontend metadata for missing or extra formal Workers.

## Budgets, Gates, And Completion

Each Worker Loop has at most eight iterations, 120000 tokens, and 90 minutes.
The Catalog Loop executes one queue decision per cycle and cannot invent Workers
outside `catalog.json`.

The no-progress fingerprint contains active verifier IDs, Definition hash,
changed paths, evidence count, and blocker code. Three identical consecutive
fingerprints stop that Loop as `blocked`. Missing real artifacts, unsupported
licenses, unavailable protocols, or attempts to weaken protected verification
are failures. Budget exhaustion and blocked states are never success.

Human approval is required before using real credentials, accepting unclear
licensing, marking a Worker host-provided, changing public Proto/API or
migrations, publishing images, pushing, merging, deploying, or promoting from
report-only to assisted execution.

Generated projects protect manifests, verifier scripts, acceptance criteria, CI,
and test-count checks. A protected verifier change requires a separate human
approval.

Loop-project delivery is complete when both manifests and generated projects
validate, all required files exist, verifier scripts are executable, no generated
file exceeds the repository line limit, and the implementation plan exists.

Catalog execution is complete only when every formal Worker is accepted, the
catalog terminal verifier passes, and a human accepts the final review gate.
