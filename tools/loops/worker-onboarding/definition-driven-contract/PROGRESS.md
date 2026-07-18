# Progress

Loop: Definition Driven Worker Create Contract

## Current Status

- Status: running
- Active loop node: `root`
- Active atomic task: `real-worker-release`
- Public contract: approved and implemented as named Definition document bindings,
  with no positional compatibility read or dual write.
- Cross-layer verifier: `focused-contract-tests` passed on July 16, 2026.
  It covers Backend, Connect, WorkerSpec, migration, Runner materialization,
  Rust API and service wire preservation, a fresh WASM build, freshness
  detection, and the targeted Web form, filter, and responsive-draft tests.
- Browser evidence: the formal WorkerTemplate flow selected `DoAgent`, exposed
  `settings` as a required JSON document targeting `DO_AGENT_SETTINGS`, and
  wrote the named binding into YAML. The selected document remained visible after
  changing from desktop to mobile with no horizontal overflow.
- WASM root cause: the loaded package predated
  `ListResourcesResponse.applied_environment_bundle_filter`, so the browser
  transport dropped the control-plane acknowledgement even though Backend had
  applied the filter. Rebuilding WASM, clearing `.next`, and restarting Next
  restored the field; a direct WASM service byte round trip returned
  `{ purpose: 2, workerType: "do-agent" }`. Dev startup now rebuilds when
  `.proto`, Rust Core source, or build scripts are newer.
- Responsive-state root cause: `ResponsiveShell` swaps `IDEShell` for
  `MobileShell`, which unmounted the local ResourceEditor reducer and erased
  the selected Worker type. The WorkerTemplate editor now uses a Dashboard
  session keyed by organization, preserving the draft through that relayout.
- Migration safety: the development database was repaired only by deleting
  unreferenced smoke records and ignoring expired WorkerTemplate plans. The
  migration was then run with `--apply`, followed by a zero-update check:
  four snapshots and four WorkerTemplate revisions migrated cleanly.
- DoAgent lifecycle: a real browser-created configuration bundle, WorkerTemplate,
  WorkerSpec snapshot, Pod, ACP session, prompt, and termination ran on
  `dev-runner-do-agent`. Materialization and transport passed. The exact prompt
  failed only when the configured OpenAI endpoint timed out from both host and
  Runner. Termination removed the active pod, relay, MCP registration, and Runner
  slot. The sandbox was retained because the request set `delete_branch=false`.
- OpenCode lifecycle: its Definition declared no required platform model
  resource. A browser-created WorkerTemplate without `modelRef` produced
  WorkerSpec snapshot 8, created Pod `1-standalone-d1d72cdc` on
  `dev-runner-opencode`, accepted the exact prompt `Reply with exactly: READY`,
  and returned `READY`. Browser-authorized termination then removed the active
  pod, relay, ACP session, MCP registration, and Runner slot.
- Cursor CLI lifecycle: its Definition allowed omission of both `modelRef` and
  the optional `CURSOR_API_KEY` bundle. The browser-created template produced
  WorkerSpec snapshot 9 and Pod `1-standalone-0580fff5` on
  `dev-runner-cursor`. ACP initialization succeeded, then Cursor rejected
  `session/new` with `Authentication required`. Runner automatically removed
  the Pod and sandbox; the browser rendered the exact failure and the Runner
  slot returned to zero. This is a credential blocker, not a platform-model or
  frontend failure.
- Grok Build preflight: the browser correctly made `modelRef` optional but
  required an `XAI_API_KEY` EnvironmentBundle, because its AgentFile declares
  `ENV XAI_API_KEY SECRET` without `OPTIONAL`. No bundle exists locally, so no
  template, Worker, Pod, or provider request was created. Model optionality
  does not mean credential optionality.
- Cursor process cleanup: `ACPClient.Stop` already waits for the managed direct
  child to exit through `processmgr`. The one reaper warning after the rejected
  Cursor session was an untracked descendant created by the external Cursor CLI,
  not an unclosed Runner direct child. No unsupported wait-layer patch was made.
- Stable browser retest after Backend hot reload showed `Pod completed`, zero
  Runner slots, and no console error or warning.
- Image supply: Alibaba Cloud or Tencent Cloud sources are authorized. The
  build contract accepts exact `NODE_BASE_IMAGE` and `PYTHON_BASE_IMAGE`
  references in Compose, local K8s, and cluster build flows. This environment
  has no authenticated ACR/TCR namespace or verified full image reference, so
  a clean rebuild is not claimed.
- Runner egress: Docker Compose and generated local K8s manifests now expose
  optional `RUNNER_HTTP_PROXY`, `RUNNER_HTTPS_PROXY`, and `RUNNER_NO_PROXY`
  without proxying Backend, gRPC, Relay, OTel, or cluster-internal addresses.
- Release-wide formal Worker support remains incomplete. OpenCode has one
  successful local end-to-end proof; DoAgent has one provider-network failure;
  Cursor CLI has one credential failure; the remaining 10 formal Worker types
  have no lifecycle proof.

## Acceptance Trace

- Checklist path: `ACCEPTANCE.md`
- Checked items: baseline, approval, named binding, resource projection,
  Definition-driven Web, and Runner materialization.
- Remaining item: real Worker release evidence. `real-worker-release.json`
  records the successful OpenCode run plus failed DoAgent and Cursor runs; the
  acceptance item remains unchecked until every promoted Worker has cleanup
  evidence.

## Blocked Decision Trace

- Decision file: `DECISIONS.md`
- Decision log: `journal.jsonl`
- Last lifecycle result: Cursor CLI reached ACP initialization but requires a
  `CURSOR_API_KEY` EnvironmentBundle to create a session.
- Last preflight result: Grok Build requires an `XAI_API_KEY` EnvironmentBundle
  even though it does not require a platform model resource.
- Current external gates: individual lifecycle proof for the remaining 10
  Worker types, reachable credentials where required, and a verified ACR/TCR
  image reference.

## Next Cycle

1. Bind disposable `CURSOR_API_KEY` and `XAI_API_KEY` EnvironmentBundles before
   retrying Cursor CLI and starting Grok Build.
2. Use each remaining Worker type's declared adapter, credentials, and config
   documents to create, prompt, terminate, and inspect cleanup one type at a
   time.
3. Configure one disposable credential for an endpoint reachable from Runner
   before retrying DoAgent and other provider-backed Worker types.
4. Obtain one exact authenticated Alibaba Cloud ACR or Tencent Cloud TCR
   Node/Python base-image pair and execute an audit-tag rebuild.
