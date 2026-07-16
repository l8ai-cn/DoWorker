# Agent Workbench V2 Progress

## Objective

Deliver one generated-protocol Agent Workbench that is shared by Web and Web User,
supports React mounting, plain-page mounting, and iframe embedding, and renders
real Runner results for code, HTML, image editing, presentations, and video.

## Completion Evidence

- Runner events reach the frontend as `agent_workbench.v2` snapshots, delta
  batches, receipts, tools, and artifacts without handwritten ACP projections.
- Tool identities select renderers through the exact-key tool registry.
- Artifact media types and manifests select HTML, image, image comparison,
  presentation, video, and generic-file viewers through the content registry.
- Web and Web User use the same shared workspace and generated V2 state.
- A real connected Runner completes programming, HTML, image, presentation, and
  video scenarios; produced files open in the results pane.
- Desktop, mobile, plain mount, and iframe paths pass browser interaction,
  console, network, and screenshot checks.

## Current State

- V2 protobuf contracts, generated Go/TypeScript/Rust bindings, fixtures, source
  tool catalog, and lossless reducers are implemented.
- Shared workspace, split results pane, iframe/plain/React mounting, preview
  security, image comparison/edit, presentation, video, terminal, and HTML
  viewers exist.
- The shared TypeScript runtime uses the official Connect client, refreshes the
  bearer token for every RPC, hashes deterministic protobuf command envelopes
  identically to Go, consumes server streams, resnapshots on cursor gaps, and
  stops after bounded no-progress retries.
- `AgentSessionRuntimeV2` exposes atomic open/subscribe/snapshot/send/interrupt/
  permission/configuration/artifact-action/resource-loading methods. It does
  not append optimistic transcript items or infer media from filenames.
- Artifact actions carry an explicit action schema version, artifact revision,
  representation, and client action ID. Built-in image-edit and presentation
  interactions use this contract.
- Rust Core has generated snapshot/stream/execute bindings, canonical state,
  WASM service/state facades, explicit access scopes, strict stream completion
  reporting, and configuration replacement semantics.
- Runner emits generated V2 timeline, tool, status, permission, unsupported, and
  explicitly declared artifact batches. Artifact revisions retain a stable
  producer and tool execution identity instead of being inferred from transcript
  text or filenames.
- Backend persists deterministic source events, projections, deltas, and
  receipts; mounts snapshot/stream/command Connect handlers; validates normal
  JWT and embed tokens against exact org/session/capability scope; and decorates
  viewer-specific session and artifact grants from the exact advertised agent
  operations.
- Runner and Backend project current model and permission mode only when ACP
  reports explicit values. The UI never selects the first supported option as a
  fabricated current value.
- Runner and Backend stream raw artifact bytes in validated 4 MiB ranges. Web
  and Web User consume the same endpoint as `Blob` data, so large video and PPT
  files are no longer base64 encoded or truncated by the legacy 1 MiB sandbox
  response.
- Web uses Rust Workbench state and the shared Agent UI. Web User uses the same
  V2 runtime and renderer registry for React mounting, plain-page mounting, and
  iframe embedding; the legacy duplicate ACP session projections were removed.
- Built-in renderers cover safe static HTML, image, image comparison/edit,
  presentation, video, code/text, and generic files. Image and presentation
  controls are enabled only when the agent explicitly advertises the matching
  artifact action and the viewer grant permits it.
- Backend Connect, Runner ACP/workbench, shared renderer, Web loader, Web User
  embed tests, Web/Web User type checks, and the Web User embed production build
  pass.
- Same-cursor snapshots now treat digests and viewer grants as refreshable
  metadata while preserving the applied-batch window. Duplicate-range digest
  conflicts disable commands and require resynchronization in both Rust and
  TypeScript.
- Snapshot position validation now matches the Backend invariant:
  `revision <= latest_sequence`, and revision/sequence are either both zero or
  both nonzero.
- Embedded artifact downloads reject cross-origin URLs before reading the
  embedded session token. Published plain mounts add the `.do-worker-app` scope
  class and remove it on unmount.
- Narrow workbenches do not mount result renderers until the user opens Results.
  Video viewers load only the active version plus its poster and rebuild the
  player when the selected version changes.
- Mobile conversation/results and conversation/terminal tabs have linked
  tabpanels, roving focus, arrow-key navigation, and at least 44 px touch
  targets.

## Final Verification

- Isolated stack: API `29950`, Web `29957`, Relay `29967`, Web User `29970`.
  Runner `dev-runner-codex` reconnected after deploying the lifecycle build.
- Real lifecycle session `conv_ee6103e62a1f0e33` ran a foreground 90-second
  task. The UI changed to Executing, exposed Stop, and returned to Idle after
  interrupt. The Runner stopped the loop at line 79 instead of completing 90.
- Real video session `conv_0444cbc013c4460f` rendered a typed video artifact.
  The MP4 was verified as H.264, 1280x720, 30 fps, 8.000 seconds, with a poster
  and three thumbnails. Browser media reached `readyState=4`.
- Desktop Web rendered the real result. React mount, plain-page mount, and
  iframe mount all opened the same V2 session runtime.
- A 390x700 iframe had no horizontal overflow. Before opening Results it had no
  video element and an empty results panel. After opening Results, the real
  video loaded at 298 px wide with `readyState=4`.
- Switching `playable` to `original` changed the browser's `currentSrc` and the
  replacement video also reached `readyState=4`.
- Published `dist-embed` JS/CSS was loaded in a standalone QA host. The mount
  root received `.do-worker-app`; a scoped `h-8` probe computed to 32 px.
- The current `29970` iframe session produced no console errors or warnings.
- Shared Agent UI: 56 files, 276 tests passed; TypeScript check passed.
- Rust Agent Workbench state: 22 tests passed. Final WASM release build passed.
- Web User focused embed tests: 10 passed; type check, application build, and
  embed distribution build passed.
- Web type check and production Next build passed.
- Backend Connect/service/infra and Runner client/Codex/workbench/runner focused
  suites passed during the same verification cycle.
- `git diff --check` passed; edited production files remain below the 200-line
  project limit.

## Residual Baseline

- The full legacy Web User suite is not green: 70 of 3,608 tests fail in old
  Sidebar, AgentInfo, filesystem hook, settings navigation, and related areas.
  The failures are outside the shared Agent Workbench paths verified above.
- Full Web User lint also reports existing errors outside this component scope.
  Lint on the edited embed files passes.
- `clients/web/scripts/check-no-wasm-in-marketing.sh` passes its marketing
  negative check against the production `.next` output, but its positive check
  assumes route-local `WasmProvider` text. The current Next build extracts that
  symbol into a shared chunk, so this validation script needs a separate
  manifest-aware repair.

## Loop Guardrails

- Stop on verified completion or on an external dependency that prevents three
  consecutive attempts from making progress.
- Reject changes that weaken protocol validation, tests, browser checks, auth,
  preview isolation, or artifact grants.
- Re-read this file, the implementation plans, git diff, and test/browser
  evidence after context transitions.
- The user explicitly requested no execution budget cap; no-progress detection
  and human escalation remain mandatory.
