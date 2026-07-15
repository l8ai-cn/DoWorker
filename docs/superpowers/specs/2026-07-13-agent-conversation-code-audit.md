# Agent Conversation Code Audit

## Scope and acceptance

This audit covers the current conversation path in `clients/web-user`, the ACP
path in `clients/web`, the experimental shared package in
`packages/agent-ui`, and the REST session API. A usable Agent Worker must:

1. discover a runnable agent and execution target;
2. create or resume a session and bind its transport;
3. render the full structured stream without flattening it;
4. send, interrupt, and resolve structured permissions with the correct
   authorization;
5. work as a page, an iframe, and a React component.

The current `iframe.html` meets none of the launch requirements. It is an
existing-session viewer with a generic composer, not a Worker.

## Current paths

### Web User REST path

`clients/web-user/src/pages/ChatPage.tsx` routes a missing session id to
`NewChatLandingScreen`. `NewChatLandingScreen` loads agents and hosts, selects
a compatible execution target, creates `POST /v1/sessions`, records the first
prompt, then navigates to the session. `chatStore.switchTo` hydrates and opens
SSE; `chatStore.send` posts typed events and optimistically renders the user
message.

The source has the working launch logic already:

| File | Current responsibility | Audit result |
| --- | --- | --- |
| `clients/web-user/src/shell/NewChatDialog.tsx:1661` | Agent, host, workspace, model, policy, file, and first-prompt launch flow | Reuse its domain flow; do not replace it with a blind text box. |
| `clients/web-user/src/store/chatStore.ts:777` | Optimistic typed message send, upload, SSE reconciliation | Authoritative Web User client state. |
| `clients/web-user/src/store/chatStore.ts:1171` | Session switch, history hydrate, stream lifecycle | Existing-session runtime behavior. |
| `clients/web-user/src/store/chatStore.ts:1539` | Create empty session, bind runner, then stream before first prompt | Correct ordering for first-turn events. |
| `clients/web-user/src/lib/chatStoreConversationRuntime.ts:54` | Experimental adapter for an existing session | Ignores `target.agentId`; cannot create or select a Worker. |
| `clients/web-user/src/lib/chatStoreConversationItems.ts:5` | Converts rich blocks to four generic item kinds | Loses Markdown, attachments, structured elicitation, tool detail, terminal state, and ordering fidelity. |
| `clients/web-user/src/iframe.tsx:12` | Boots generic iframe and waits for `open_session` | No launcher, no standalone mode, no session bootstrap. |
| `clients/web-user/src/lib/bootstrapEmbeddedConversation.ts:8` | Initializes store, identity, and server info | Does not load agents, hosts, or create/bind a session. |

### Web ACP path

`AgentPanel` is a pod-scoped ACP workspace. Pod creation, relay subscription,
session hydration, terminal state, plans, and permissions are already owned by
the Web/Rust path. It is not REST session creation.

| File | Current responsibility | Audit result |
| --- | --- | --- |
| `clients/web/src/components/workspace/AgentPanel.tsx:62` | Creates the experimental runtime and replaces the rich stream | Regression: `AcpActivityStream` was replaced by a generic flat surface. |
| `clients/web/src/components/workspace/acp/createAcpConversationRuntime.ts:13` | Maps ACP store and sends relay commands | `getSnapshot()` allocates on every call, violating `useSyncExternalStore` snapshot identity. |
| `clients/web/src/components/workspace/acp/createAcpConversationRuntime.ts:100` | Opens an already selected pod only | `podKey` is treated as a session id; no Worker creation or selection exists. |
| `clients/web/src/components/workspace/acp/AcpActivityStream.tsx` | Rich Markdown, tools, thinking, logs, scrolling | Must remain the ACP renderer until a typed equivalent preserves every state. |
| `clients/web/src/components/workspace/acp/AcpPromptInput.tsx` | ACP-aware prompt, controls, and error state | Preserved as a slot, but the generic runtime send path is unused. |

### Shared package and iframe protocol

| File | Current responsibility | Audit result |
| --- | --- | --- |
| `packages/agent-ui/src/conversation-types.ts:45` | Four flattened display item kinds and an existing-session runtime | Missing catalog, launch, binding, capability, attachment, and typed-permission contracts. |
| `packages/agent-ui/src/ConversationSurface.tsx:22` | Generic cards, textarea, approve/decline UI | This is a demo renderer, not an Agentic workbench. |
| `packages/agent-ui/src/conversation-context.tsx:34` | `useSyncExternalStore` subscription | Correct only when adapters return a stable snapshot between mutations. |
| `packages/agent-ui/src/iframe-bridge.ts:19` | Parent/frame messages | Only `open_session`; cannot launch, navigate, or report capability state. Pending opens never time out. |
| `packages/agent-ui/src/conversation-iframe-protocol.ts:4` | Message schema | `agentId` is optional metadata and is ignored by the Web User adapter. |
| `packages/agent-ui/src/conversation-context.test.tsx:7` | Runtime subscription test | Stale snapshot shape and missing runtime methods; Vitest transpiles it without type validation. |

## Confirmed defects

| Severity | Defect | Evidence | Required correction |
| --- | --- | --- | --- |
| P0 | The direct iframe URL has no session and cannot start one. | `iframe.tsx:32-67` only accepts `open_session`; screenshot reproduces the empty surface. | Separate a Worker page from the embedded surface; the page must launch/resume. |
| P0 | Shared contract cannot describe a Worker launch. | `ConversationRuntime` only opens an existing `sessionId`. | Add a launch controller contract before implementing UI. |
| P0 | Read-only users can mutate sessions. | `session_events.go:17` and `session_elicitations.go:31` call `authorizeSession`, which grants `levelRead`; neither calls `requireSessionLevel(levelEdit)`. | Require edit permission and add handler tests. |
| P1 | ACP adapter can cause repeated external-store renders. | `createAcpConversationRuntime.ts:96` creates a snapshot for every read. | Cache by ACP store version/state identity and render-test it. |
| P1 | ACP feature regression. | `AgentPanel.tsx:129` replaces `AcpActivityStream`; generic model flattens tools/reasoning/logs. | Restore native ACP stream; integrate only stable component boundaries. |
| P1 | ACP relay rejection is swallowed and clears the prompt. | `relayConnection.ts:128` discarded the async result; `AcpPromptInput.tsx:42-43` cleared after dispatch. | Return the relay promise, clear only after it resolves, and retain input on failure. |
| P1 | Web User feature regression. | `chatStoreConversationItems.ts:5-115` flattens blocks and attachment content. | Preserve Web User's typed block renderer instead of converting it. |
| P2 | ACP `info` logs are discarded. | `acpEventDispatcher.ts:67-72` persisted only warnings and errors. | Store every runner log level so `AcpActivityStream` can display it. |
| P1 | Iframe trust is query-controlled. | `iframe.tsx:19-30` accepts caller-provided `parent_origin`; bridge trusts it. | Use server-issued embed context with allowed origins; do not trust URL origin input. |
| P2 | Iframe host API has no timeout or lifecycle protocol. | `iframe-bridge.ts:112-128` leaves promises pending indefinitely. | Add request timeout, `close`, and explicit error/ready states. |

## Target design

Use two layers, not one fake universal runtime:

1. **Worker controller** owns catalog, target selection, create/resume, and
   transport binding. It exposes `listAgents`, `listExecutionTargets`,
   `createSession`, `resumeSession`, and `openSession`. Web User implements it
   with REST/SSE. Web implements it with Rust/Connect pod creation and ACP.
2. **Conversation renderer** receives the native stream model through typed
   slots. It owns page layout, loading/error/empty states, navigation, and
   iframe lifecycle. Web User keeps its block renderer and composer; Web keeps
   ACP activity, prompt, and permission components. No lossy common union is
   allowed on the critical rendering path.

Entry points are distinct:

| Entry | Behavior |
| --- | --- |
| React component | Caller provides a controller and native renderer slots. It launches or resumes a Worker. |
| Standalone page | Owns launcher state and renders the component. This is the directly usable URL. |
| Iframe | Receives a server-issued, short-lived embed context that fixes tenant, allowed parent origins, and allowed actions. It renders the same component; it does not accept a parent origin as authorization. |

The first implementation slice is Web User because it already has the full
REST launch chain. Web ACP will restore its native stream immediately and
adopt only the controller boundary after the Connect/Rust launch adapter is
implemented and verified.

## Implemented correction boundary

`clients/web-user/src/embed.tsx` already exports `DoWorkerApp`, the actual
React component integration point. It carries the complete Web User provider
tree and reuses the existing Agent launch and session renderer. The replacement
`worker.html` and `iframe.html` both mount it through `HashRouter`; direct
navigation now lands on `NewChatLandingScreen`, and an iframe renders the same
real Worker UI.

This correction does not expose parent-driven `postMessage` control. The prior
bridge trusted a caller-provided origin and only opened arbitrary existing
sessions. Cross-origin parent control must wait for a signed server-issued
embed context; no query-authorized bridge remains.

## Verification matrix

1. Unit: controller creates/resumes and binds before first send.
2. Unit: read permission cannot send, interrupt, or resolve an elicitation.
3. Browser: standalone and iframe documents both mount the complete Worker
   launcher rather than an existing-session-only composer.
4. Browser: standalone page selects an agent and target, creates a session,
   sends a prompt, receives SSE output, handles loading/error/disabled states.
5. Security: parent-driven cross-origin control stays unavailable until signed
   embed contexts and their origin checks are implemented.
6. Browser: Web ACP still renders Markdown, tools, thinking, logs, composer,
   and permission interactions.
