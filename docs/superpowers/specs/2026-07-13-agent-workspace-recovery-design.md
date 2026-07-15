# Agent Workspace Component Design

**Status:** implemented and browser-verified
**Scope:** shared Agent conversation workspace, PTY terminal surface, Web
integration, React integration, and capability-scoped iframe embedding.

## Decision

The reusable product unit is `AgentWorkspace`, backed by injected runtime
contracts. It is not a copied chat page and does not depend on either
application's authentication, REST client, WASM bridge, Relay pool, Zustand
store, or router.

Conversation and terminal are separate atoms:

- `AgentSessionRuntime` owns durable and live Agent conversation behavior.
- `TerminalRuntime` owns PTY bytes, resize, connection, and control leases.
- `AgentWorkspace` composes the atoms only when the runtime advertises the
  corresponding capability and resource.

This preserves rich Agent behavior without forcing Web and Web User onto one
transport.

## Source Allocation

| Concern        | Source retained from Web                                                | Source retained from Web User                                                       |
| -------------- | ----------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| Agent control  | Rust Relay pool, ACP commands, interrupt, model and permission commands | Durable session event API                                                           |
| Live state     | ACP snapshot/event projection, plan, tools, reasoning, permissions      | SSE session status and item stream                                                  |
| History        | ACP snapshot hydration                                                  | REST session/items snapshot and pagination                                          |
| Rendering      | Existing Web status and workspace shell                                 | Block reducer for messages, tools, reasoning, routing, compaction, and elicitations |
| Terminal       | Relay control-lease semantics and xterm behavior                        | Session terminal resource API and iframe Relay connection                           |
| Embedding      | Web consumes the shared component inside `AgentPanel`                   | React export, iframe bootstrap, origin handshake, scoped session client             |
| Authentication | Existing Rust `AuthManager` remains the Web authority                   | Embed token redemption is isolated from legacy Web User login/session state         |

Neither application's store is shared. Each application adapts its native
state into `AgentSessionSnapshot`.

## Package Boundary

`packages/agent-ui` exports:

```ts
interface AgentSessionRuntime {
  open(sessionId: string): Promise<void>;
  close(sessionId: string): void;
  getSnapshot(sessionId: string): AgentSessionSnapshot;
  subscribe(sessionId: string, listener: () => void): () => void;
  sendMessage(
    sessionId: string,
    commandId: string,
    input: { text: string },
  ): Promise<void>;
  interrupt(sessionId: string, commandId: string): Promise<void>;
  resolvePermission(
    sessionId: string,
    commandId: string,
    permissionId: string,
    result: PermissionResult,
  ): Promise<void>;
  updateConfiguration(
    sessionId: string,
    commandId: string,
    patch: Record<string, unknown>,
  ): Promise<void>;
  loadOlder(sessionId: string, beforeItemId?: string): Promise<void>;
}

interface TerminalRuntime {
  connect(resource: TerminalResource): Promise<void>;
  disconnect(resourceId: string): void;
  subscribeOutput(resourceId: string, listener: OutputListener): () => void;
  subscribeStatus(resourceId: string, listener: StatusListener): () => void;
  write(resourceId: string, bytes: Uint8Array): Promise<void>;
  resize(resourceId: string, columns: number, rows: number): Promise<void>;
  acquireControl(
    resourceId: string,
    clientLabel: string,
  ): Promise<TerminalControlLease>;
  renewControl(resourceId: string, leaseId: string): Promise<void>;
  releaseControl(resourceId: string, leaseId: string): Promise<void>;
}
```

`AgentWorkspace` contains:

```text
AgentWorkspace
├── WorkspaceHeader
├── Conversation
│   ├── PlanStrip
│   ├── ActivityTimeline
│   ├── ApprovalDock
│   └── ConversationComposer
└── TerminalSurface (only for a real PTY resource)
```

The component accepts `runtime`, `sessionId`, and optional
`terminalRuntime`. The host owns layout and lifecycle outside that boundary.

## Capability Rules

Controls are rendered from runtime truth, not from the reference design:

- Send requires `sendMessage`, an idle session, and a non-empty draft.
- Running or permission-waiting sessions keep the Stop action primary even
  when a draft exists; ACP rejects a second prompt until idle.
- Approval actions require `resolvePermission`.
- Terminal labels and tabs require all three: terminal capability, a terminal
  resource, and a `TerminalRuntime`.
- Terminal input requires a granted control lease.
- Model, attachment, skill, application, and slash-command controls are not
  shown until their command contracts exist.

The visual reference informs hierarchy: centered empty-state intent, a large
composer, compact capability labels, and one primary circular action. It does
not authorize fake controls.

## ACP And PTY Boundary

ACP is not a terminal.

`ACPPodIO` has no shell process, PTY, VT aggregator, terminal detach, or PTY
snapshot. Its Relay data plane carries `AcpCommand`, `AcpEvent`, and
`AcpSnapshot`. Therefore:

- Web ACP snapshots advertise no terminal.
- Session terminal listing returns no terminal for ACP pods.
- Relay terminal connection rejects ACP pods.
- `TerminalSurface` is used only for PTY sessions.

Adding an ACP terminal later requires a new explicit runner protocol or a real
PTY resource. A connected ACP Relay socket must never be presented as an
interactive terminal.

## Runtime Adapters

### Web

`WebAcpSessionRuntime`:

1. Opens one Rust Relay subscription for the pod.
2. Projects the existing ACP store into the shared timeline, plan, approval,
   status, and configuration contract.
3. Sends ACP prompt, interrupt, permission, model, and permission-mode
   commands through the Rust Relay manager.
4. Leaves Rust Core and the existing control-lease overlay authoritative.

`AgentPanel` renders `AgentWorkspace` in place of the previous ACP-only view.
PTY terminal panes elsewhere in Web continue to use their native terminal
path; they are not mislabeled as ACP resources.

### Web User And Iframe

`EmbeddedAgentSessionRuntime`:

1. Hydrates the granted session, durable items, and PTY terminal resources.
2. Reduces Web User item shapes into the shared timeline model.
3. Consumes the session SSE stream for live state.
4. Re-hydrates durable session/items/terminal state after reconnect so events
   missed in the gap are recovered.
5. Stops reconnecting on non-recoverable HTTP responses such as expired or
   forbidden embed tokens.

Reconnect is snapshot reconciliation, not event-log replay. The client does
not send `Last-Event-ID`; persisted items and session state are authoritative
after a reconnect. An ephemeral delta that was never persisted can be absent
until the backend exposes it through a later durable snapshot.

`EmbeddedTerminalRuntime` obtains a scoped Relay connection and applies the
same lease, output, input, resize, renew, and release contract as the shared
terminal surface. The iframe workspace is fixed to `100dvh`, and the terminal
mount clips overflow so xterm's fit observer cannot feed its measured height
back into the document and grow the iframe indefinitely.

## Iframe Protocol

1. A session manager calls `POST /v1/sessions/:id/embed-context` with exact
   parent origins and explicit capabilities.
2. The backend returns a five-minute `agent_embed_context` JWT plus an
   independent random `redemption_proof`. Redis stores only the proof hash,
   keyed by the context `jti`. The parent keeps the proof out of the iframe URL.
3. `/iframe.html` calls `POST /v1/embed-contexts/inspect` with the context.
   This validates that the context is signed, unexpired, and still unredeemed,
   then returns only the exact parent origins and expiry.
4. The frame removes the context from its URL, installs its message listener,
   and sends `agentsmesh.embed.ready` only to those origins. The parent replies
   with `agentsmesh.embed.open` carrying `redemptionProof`. Both
   `event.source` and exact origin must match.
5. The frame calls `POST /v1/embed-contexts/redeem` with the context and proof.
   A Redis script compares the stored hash and deletes it in one operation.
   A wrong proof cannot consume a valid context; a successful proof cannot be
   replayed across backend instances.
6. The frame receives a fifteen-minute
   `agent_embed_session` token bound to one organization, user, session,
   capability set, and parent-origin set.
7. Embedded REST, SSE, approval, terminal, and Relay routes enforce the exact
   session and capability. Relay browser and runner tokens cannot outlive the
   embed session token.

The context is accepted only by the inspect and redemption endpoints, never by
normal auth middleware. The session token is never accepted outside
`/v1/embed/sessions/:id`.

## Integration Modes

| Mode                   | Entry                                                   | Authority                                 |
| ---------------------- | ------------------------------------------------------- | ----------------------------------------- |
| Shared React component | `<AgentWorkspace runtime={...} sessionId={...} />`      | Host-injected runtime                     |
| Web workspace          | `AgentPanel`                                            | Rust AuthManager, Connect-RPC, Rust Relay |
| Web User same-root     | exported `EmbeddedAgentWorkspace` or full `DoWorkerApp` | Host or Web User auth                     |
| Iframe                 | `/iframe.html?embed_context=...`                        | Context plus parent-held redemption proof |
| Standalone Web User    | `/worker.html`                                          | Web User application session              |

## Verification

- Shared package: component, composer lifecycle, approval, terminal lease, and
  TypeScript checks.
- Web: ACP projection tests, full Vitest suite, TypeScript check, and a real
  browser prompt returning `WEB_SHARED_WORKSPACE_OK`.
- Web User: embed client, hydration, reconnect recovery, fatal stream error,
  workspace projection, terminal runtime, StrictMode-safe one-time context
  redemption, iframe, and production embed build.
- Backend: embed token, one-time redemption, origin/capability authorization,
  session SSE writer, terminal mode boundary, Relay token lifetime, and session
  route tests.
- Browser: desktop Web, desktop iframe, PTY terminal connection, live SSE
  response, control lease acquire/release, and 390px iframe layout.

## Acceptance

- Web and iframe render the same shared conversation component.
- A real Agent response arrives without refresh after a submitted prompt.
- Reconnect reconciles durable events missed between streams.
- Running ACP sessions cannot submit a second prompt.
- ACP never advertises a terminal.
- PTY terminal input is impossible without a control lease.
- Embed contexts are one-time, origin-bound, session-bound, capability-bound,
  and fail closed when Redis is unavailable.
- A stolen iframe context cannot be redeemed without the parent-held proof.
- React StrictMode initialization sends one inspection and one redemption
  request.
