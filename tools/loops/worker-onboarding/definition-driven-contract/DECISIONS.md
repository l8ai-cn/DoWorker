# Blocked Execution Decisions

Loop: Definition Driven Worker Create Contract

## User Confirmation

- Required before delegation: true
- Prompt: May the decision-proxy choose only low-risk evidence order and one read-only verifier retry?
- Confirmation record: `PROGRESS.md`

## Proxy Decision Agent

- Agent: `decision-proxy`
- Authority: `delegated_low_risk`
- Default when uncertain: `ask_user`

Allowed decisions:

- choose evidence order
- split a read-only audit
- retry one deterministic verifier

Forbidden decisions:

- irreversible action
- public contract approval
- credential use
- production deploy

Decision records must include:

- `blocked_reason`
- `options`
- `selected_option`
- `rationale`
- `evidence_ref`

## Blocked Handling

Blocked signals:

- approval missing
- credential authorization missing
- same verifier fails twice
- browser capability unavailable

- Max blocked cycles: 1

Allowed resolution actions:

- record durable evidence
- ask the secondary user query
- process an independent read-only audit

Escalate when:

- public API decision
- database migration decision
- provider action
- protected verifier conflict

## Supervisor

- Agent: `loop-supervisor`
- Review cadence iterations: 2
- Report path: `monitoring-plan.json`

Drift checks:

- active task maps to the objective
- public contract changes have approval
- verifiers remain protected

Intervention actions:

- pause invalid progress
- record no-progress exit
- escalate to user

## Recorded Contract Decision

- Approved on July 16, 2026: replace positional references with
  `config_document_bindings[{document_id, config_bundle_id}]`.
- Implementation status: verified across Backend, Rust/WASM wire, Web,
  migration, and Runner materialization tests.
- Compatibility policy: no dual-read, dual-write, or anonymous fallback path.

## Remaining External Gates

- Historical-data gate: resolved on July 17, 2026. Only unreferenced smoke
  records were deleted, expired WorkerTemplate plans no longer block migration,
  and the applied migration was verified with a zero-update check.
- Image-source gate: Alibaba Cloud or Tencent Cloud is authorized, but an
  authenticated registry namespace and exact Node/Python image references have
  not passed manifest verification.
- Lifecycle gate: OpenCode completed real browser template creation, Pod launch,
  ACP prompt response, and cleanup without a platform model binding. DoAgent
  was also created and terminated with real browser and Runner evidence, but its
  provider prompt failed because the configured OpenAI endpoint is unreachable
  from both the host and Runner. Cursor CLI reached ACP initialization but
  rejected `session/new` without its optional `CURSOR_API_KEY` bundle. The
  remaining 10 Worker types still need individual lifecycle evidence and
  credentials where their Definitions require them.
- Runner egress gate: optional Runner-only proxy variables are now wired into
  Compose and generated local K8s manifests. A real proxy endpoint is not
  configured in this environment, so this is a configuration-path proof only.
