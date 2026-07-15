# Agent Conversation Integration Design

> **Status: superseded.** The implemented runtime and security contract is
> documented in `2026-07-13-agent-workspace-recovery-design.md`.

## Decision

Do not merge Web and Web User by converting their streams into a lowest-common-
denominator `message | activity | notice | permission` model. The two clients
share the Worker lifecycle, but their transports and rich render data differ:

- Web User owns REST session creation, runner placement, SSE history, uploads,
  native-terminal state, and its block renderer.
- Web owns Rust Core state, Connect-RPC pod creation, ACP relay, terminal
  controls, plans, and ACP-native tool/thinking/permission rendering.

The reusable unit is therefore a **Worker workspace composition contract**:
launch or resume a Worker, then render the owning client's native conversation
surface. It is not a universal chat renderer.

## Source capabilities

| Capability         | Take from Web User                                           | Take from Web                                       |
| ------------------ | ------------------------------------------------------------ | --------------------------------------------------- |
| Agent catalog      | `useAvailableAgents` and agent capability metadata           | Rust/Connect worker creation options                |
| Execution target   | `useHosts`, workspace picker, managed sandbox choice         | Pod placement and runner selection in Rust Core     |
| New Worker         | `POST /v1/sessions` with host/workspace/model/policy         | Connect pod creation flow                           |
| History and stream | `chatStore.switchTo`, SSE, block reducer                     | ACP session store, relay subscription, hydration    |
| Composer           | Attachments, mentions, slash commands, model/policy controls | ACP prompt, interrupt, permission mode              |
| Timeline           | Web User block components, native terminal cards             | `AcpActivityStream`, tool cards, plans, permissions |
| React integration  | `DoWorkerApp` is already a real same-root component          | Existing `AgentPanel` remains pod-scoped            |

## Component boundary

The shared package, when introduced, exports controller and host contracts only:

```ts
export type WorkerRef = {
  workerId: string;
  sessionId?: string;
  podKey?: string;
};

export type WorkerLaunchRequest = {
  agentId: string;
  executionTargetId?: string;
  workspace?: string;
  initialPrompt?: string;
};

export interface WorkerWorkspaceController {
  listAgents(): Promise<readonly WorkerAgent[]>;
  listExecutionTargets(agentId: string): Promise<readonly ExecutionTarget[]>;
  launch(request: WorkerLaunchRequest): Promise<WorkerRef>;
  resume(ref: WorkerRef): Promise<void>;
}

export interface WorkerWorkspaceSlots {
  renderLauncher(controller: WorkerWorkspaceController): React.ReactNode;
  renderConversation(ref: WorkerRef): React.ReactNode;
}
```

The shell decides whether to show the launcher or a resumed Worker. The slots
keep protocol-specific render data within their source client. A controller
must fail on unsupported launch options; it must not silently route to another
transport.

## Implemented entry points

`clients/web-user/src/embed.tsx` already exports `DoWorkerApp`, which provides
the direct React-component path. It owns the full Web User provider stack but
expects the host's router:

```tsx
<BrowserRouter>
  <DoWorkerApp basename="/agent-worker" />
</BrowserRouter>
```

For a server-authorized existing session, the package also exports:

```tsx
const client = createEmbedSessionClient(access);
<EmbeddedSessionTimeline client={client} />;
```

`EmbedSessionAccess` only comes from the server's proof-backed
context-redemption response.
The component uses Web User's native session DTOs, SSE parser, block reducer,
and block renderer. It does not use the full Worker launcher or an invented
generic chat model.

The document entries are deliberately different:

| Entry           | URL                              | Router       | Use                              |
| --------------- | -------------------------------- | ------------ | -------------------------------- |
| Standalone      | `/worker.html`                   | `HashRouter` | Directly usable Agent Worker     |
| Iframe document | `/iframe.html?embed_context=...` | None         | Restricted existing-session view |
| Same-root React | `DoWorkerApp`                    | Host router  | Embed in an existing React page  |

`worker.html` retains the real full Worker experience: agent selection, target
selection, workspace choice, session creation, and the native conversation
surface. `iframe.html` never exposes that launcher. It redeems an authorized
context, hydrates one specified session, streams its timeline, and conditionally
shows the text composer only when the redeemed token has `write`.

## Cross-origin iframe protocol

The old `parent_origin` plus nonce design is rejected because it put
authorization data under parent control. The implemented protocol is:

1. A manager calls `POST /v1/sessions/:id/embed-context` with exact allowed
   origins and explicit capabilities. The backend issues an
   `agent_embed_context` bearer token and an independent `redemption_proof`.
2. The parent places only the context in
   `/iframe.html?embed_context=<opaque>` and retains the proof in memory.
3. The frame calls `POST /v1/embed-contexts/inspect`, learns the signed exact
   origin set, removes the context from its URL, and installs its message
   listener.
4. The frame sends `{ type: "agentsmesh.embed.ready", version: 1 }` to each
   allowed origin. The parent replies with
   `{ type: "agentsmesh.embed.open", version: 1, redemptionProof }`.
5. The frame accepts `open` only when both
   `event.source === window.parent` and `event.origin` exactly match the signed
   origin set, then redeems the context and proof at
   `POST /v1/embed-contexts/redeem`.
6. Redis atomically compares the stored proof hash and deletes it. A wrong
   proof cannot consume the context, and a successful redemption cannot be
   replayed. The response has a distinct fifteen-minute
   `agent_embed_session` token, exact session id, capabilities, and allowed
   parent origins.

The `ready` stage prevents the parent from losing an `open` message while the
iframe is bootstrapping. The development fixture
`clients/web-user/e2e/embed-host.html` demonstrates this sequence and clears
the context and proof from the parent URL after assigning the iframe source.

The origin check determines which parent can deliver the server-issued proof;
the backend still performs the decisive proof comparison and remains the
source of truth for session access and write permissions.

## Backend authorization

The issuer must hold `levelManage` on the session. It creates a context with a
non-empty `read` and/or `write` capability set and exact `http` or `https`
origins. Wildcards, URL paths, query strings, fragments, and duplicate origins
are rejected.

The token uses are purpose-separated:

- `agent_embed_context` is accepted only by `/v1/embed-contexts/inspect` and
  `/v1/embed-contexts/redeem`.
- `agent_embed_session` is accepted only under `/v1/embed/sessions/:id`.
- Normal REST, Connect-RPC, and service-token validation reject either embed
  token use.

The embedded API surface is intentionally narrow:

- `GET /v1/embed/sessions/:id`
- `GET /v1/embed/sessions/:id/items`
- `GET /v1/embed/sessions/:id/stream`
- `POST /v1/embed/sessions/:id/events` only with `write`
- `POST /v1/embed/sessions/:id/elicitations/:id/resolve` only with `approve`
- `GET /v1/embed/sessions/:id/resources/terminals` only with `terminal`
- `GET /v1/embed/sessions/:id/relay-connection` only with `terminal` and
  `control`

The token is matched to its exact session id. It cannot list sessions, launch
an agent, or access hosts. Terminal and elicitation access require explicit
capabilities. Read-only embeds render a timeline with a disabled composer and
send control.

## Acceptance criteria

- A direct Worker URL presents agent, runner, workspace, and first-prompt
  controls, then launches a real session.
- An iframe requires a server-issued context, verifies a signed parent origin
  at handshake time, and renders only the authorized session.
- A React host renders `DoWorkerApp` without another React root or router.
- A React host can render `EmbeddedSessionTimeline` using a redeemed
  `EmbedSessionAccess` without importing the full Worker launcher.
- Web ACP retains rich Markdown, tool, thinking, log, prompt, and permission
  behavior.
- A read-only session share cannot mutate Agent state.
