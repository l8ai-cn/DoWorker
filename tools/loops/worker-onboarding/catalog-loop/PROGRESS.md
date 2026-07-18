# Progress

Loop: Worker Integration Evidence Rebuild

## Current Status

- Status: iteration budget exhausted; new per-Worker work is tracked separately
- Active task: `none` in this catalog loop
- Current plan: Worker Integration Evidence Rebuild Plan
- Formal catalog: 13 Worker types
- Support claim: no Worker is formally supported or launch-verified in this
  rebuild.
- Human gate: do not send external model requests or read provider credentials
  without explicit authorization for a named non-production resource.

## Current Evidence

The deterministic definition chain is current:

- `config/worker-types/catalog.json` declares 13 slugs.
- `worker-definition-sync` projected all 13 Definitions into the development
  database.
- `sync-worker-run-contracts`, generated documentation catalog validation,
  definition-chain verification, and runtime catalog validation passed.

The current workspace runtime baseline is
`evidence/runtime-instance-baseline-2026-07-16.json`:

- The testable API origin is `http://[::1]:12400`.
- The local runtime catalog contains verified `codex-cli`, `gemini-cli`,
  `do-agent`, `aider`, `claude-code`, and `cursor-cli` images.
- `dev-runner-codex` (`codex-cli`), `dev-runner-gemini` (`gemini-cli`),
  `dev-runner-do-agent` (`do-agent`, `seedance-expert`),
  `dev-runner-aider` (`aider`), `dev-runner-claude` (`claude-code`), and
  `dev-runner-cursor` (`cursor-cli`) are online with connected tunnels;
  `dev-runner` (`e2e-echo`) is also online for fixture use. The other formal
  Worker Runners remain offline or unstarted.
- Every formal Worker is returned by `ListWorkerCreateOptions`. The eight
  types without a current local runtime image are explicitly disabled; Codex,
  Gemini, Aider, Claude Code, and Cursor are selectable only because each has
  a stable online compatible Runner.
- Earlier API evidence against `http://[::1]:10000` belongs to another
  worktree stack and is invalid for this workspace. It must not be used for
  preflight, Runner, form, credential, or lifecycle acceptance.

The local runtime catalog generator and service mapping now cover all 13
formal Worker types. It still includes an entry only after Docker reports an
immutable local image ID, so the current API intentionally remains blocked
until each image is rebuilt and the Backend reloads the generated catalog.

`evidence/aider-template-apply-and-launch-plan-2026-07-16.json` records the
current Aider image, Runner, encrypted credential form, applied WorkerTemplate,
and un-applied launch plan. It excludes injection, Pod lifecycle, terminal, and
provider execution; Aider remains not formally supported.

`evidence/cursor-cli-local-preflight-2026-07-16.json` records the equivalent
Cursor path, including its exact image version and the browser's PTY/ACP mode
contract. `evidence/amesh-codegen-transient-2026-07-16.json` records a
non-reproducible signal-9 interruption during the first lifecycle restart; no
source change was made and the unmodified retry passed.

`evidence/gemini-cli-local-preflight-2026-07-16.json` records Gemini CLI
0.50.0, its current local image digest, Runner tunnel, continuous heartbeat,
API availability, and browser PTY/ACP selection. It also records a failed
WorkerTemplate compatibility contract: the only visible `ModelBinding` points
to an OpenAI-compatible `gpt-5` resource, while Gemini requires the `gemini`
protocol. The backend compiler rejects that mismatch later, but the current
generic reference catalog exposes it to the user. Gemini is not supported.

`evidence/claude-code-local-preflight-2026-07-16.json` records the equivalent
Claude Code 2.1.211 image, local catalog, connected Runner, heartbeat, API,
and browser runtime-selection facts. Claude Code requires `anthropic`; the
same generic catalog still exposes the incompatible OpenAI binding. Claude
Code is not supported.

`evidence/do-agent-local-preflight-2026-07-16.json` records DoAgent 0.2.7,
its immutable local image, connected Runner, model-binding metadata, applied
WorkerTemplate, and browser Worker launch plan. The plan pins
`do-agent-smoke-template@r1`, whose WorkerSpec snapshot is `2`. Applying that
launch plan would create a Pod and may use a configured provider credential,
so it was not applied without named non-production authorization. DoAgent is
not supported.

`evidence/seedance-expert-local-preflight-2026-07-16.json` records the shared
DoAgent image and Runner, Seedance's required `seedance-video` tool contract,
the encrypted Doubao credential form, and the missing-resource browser block.
The plan path regression for hyphenated tool roles was repaired and browser
rechecked; no Doubao credential, Worker, or Pod was created. Seedance Expert
is not supported.

The Worker create options project Definition primary-model protocol adapters
through Go, Connect, TypeScript, and Rust protobuf mirrors. The
Resource-native create form uses those returned adapters for model-resource
filtering rather than a static slug map. This has focused Go and Web test
evidence only; it does not replace browser, Runner, credential, or lifecycle
evidence.

`evidence/web-connect-and-worker-template-2026-07-16.json` records the real
browser path: removing an overlapping Next.js Connect rewrite restored local
login, and `/dev-org/workers/new` rendered all 13 types. The 12 image-blocked
types remained visible and disabled with their reason; the Resource-native
Worker Template exposed runtime, placement, resource, model, configuration,
Secret-reference, and workspace-reference fields without Secret plaintext.
The most recent clean reload returned `200` from `ListWorkerCreateOptions`,
had no unassociated HTML labels in the runtime form or console errors/issues,
and exercised the custom CPU, memory, storage, and GPU resource path.

Config-kind bundles now require `__json` to decode to a JSON object on Create
and Update. Existing malformed rows also fail Pod command construction instead
of being omitted by `ParseConfigDocuments`; a live Connect request received
`400 invalid_argument` before persistence. This closes silent config loss but
does not bind `config_bundle_ids` to Definition `config_documents`.

Web type checking passed. The most recent `pnpm run web:test` execution ran
285 files and failed only three untracked Workflow tests outside this Worker
change; the blocker details are retained below. `web:lint` reported 183
existing warnings and no errors.

## Corrected Root Causes

- The local runtime catalog had a five-runtime allowlist. It now has explicit
  metadata, image arguments, and Runner service mappings for all 13 formal
  Worker types, with `do-agent` and `seedance-expert` sharing one image.
- The runtime catalog preserves its immutable-image-ID gate; this change does
  not infer that a missing image is ready.
- Preflight metadata paths use `ResolveMetadata` rather than decrypting a
  provider credential. Actual Pod launch remains the only decrypting path.
- `runner-aider` now explicitly builds from `python-runtime-base`, matching its
  `pip3 install aider-chat` Dockerfile branch. The current rebuilt image is
  cataloged, version-probed, and reported by a connected Aider Runner; lifecycle
  evidence still requires named non-production authorization.
- `runner-entrypoint.sh` rejected `AGENT_RUNTIME=cursor-cli` despite a matching
  Compose service and Dockerfile branch. The red/green contract repair is
  recorded in `evidence/cursor-runner-entrypoint-contract-2026-07-16.json`.
- The development Backend started a full MCP Registry import by default. Its
  large failing database batch starved Runner heartbeat and readiness updates.
  Development config now writes and exports `MCP_REGISTRY_ENABLED=false`;
  production behavior is unchanged. See
  `evidence/dev-runtime-stability-2026-07-16.json`.

## Current Blockers

- The unrelated Docker build queue cleared. The development infrastructure,
  Codex Runner, and authenticated Worker-options request have now been
  revalidated. This is infrastructure stability evidence only, not a Worker
  lifecycle or support result.
- Codex, Gemini, DoAgent, Aider, Claude Code, and Cursor currently have
  verified local images in this workspace. The next runtime step is to build
  one remaining Worker image at a time and record its Docker, Runner, API,
  and browser outcome.
- `:10000` is not this workspace's API origin. All earlier live evidence that
  used it is invalidated; use only `:12400` for this rebuild.
- `pnpm run web:test` currently fails in three untracked Workflow tests outside
  this Worker change: one imports a nonexistent `WorkflowRevisionDialog`, and
  two assert a Workflow API/Header contract that current source does not
  implement. These tests were not deleted, skipped, or changed.
- A full lifecycle test needs an authorized non-production provider credential
  and may incur external usage.
- Local port `10000` is shadowed on IPv4 by `netdisk_s`; startup now rejects
  this condition rather than misrouting API traffic.
- Runner startup logs an OpenTelemetry schema-URL conflict. The tunnel still
  connected, but observability initialization must be separately repaired.
- Five orphaned Pods created on July 14-15 remain active in the development
  database. They predate this loop and were not modified; they must be
  classified or cleaned up by an authorized owner before lifecycle testing.
- Full Go package linking and WASM rebuild are blocked by local disk exhaustion
  (less than 1 GiB free). The shared 23 GiB Cargo target was not deleted.
- The development Backend logs a separate missing
  `workflow_runs.execution_manifest` database migration. It was observed but
  not modified during this Worker-contract change.
- Definition `config_documents` are still not bound to a typed create request:
  the form and WorkerSpec use a generic `config_bundle_ids` list. The proposed
  document-ID binding migration remains a user-confirmation gate.
- The Resource-native form uses Definition protocol adapters, but the legacy
  form still contains a static model-resource mapping. It cannot be removed
  safely until that path either consumes create options or stops accepting
  formal Workers.
- The WorkerTemplate form lists every `ModelBinding` without resolving it
  against the selected Definition, and ToolBinding roles remain manually keyed.
  Seedance now reports its missing `seedance-video` path precisely, but
  definition-derived compatible-binding selection remains a confirmation gate.

## Next Cycle

1. The current loop has reached its fixed 24-iteration limit; record any new
   Worker work in a newly authorized loop rather than extending this one.
2. Before changing WorkerTemplate model selection, confirm the proposed
   server-owned compatible `ModelBinding` query and its resource-document
   boundary.
3. After explicit credential authorization, create and clean up one
   disposable Worker per launch-ready type.
4. Before changing the configuration-document boundary, confirm
   [the proposed runtime architecture](../../../../docs/superpowers/specs/2026-07-16-worker-integration-runtime-architecture.md).
   Physical Runner selection remains a scheduler result, not a normal form
   field. The required contract changes are an explicit placement-policy
   control and per-Definition configuration-document bindings instead of the
   current untyped config-bundle ID list.
