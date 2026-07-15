# Agent Workbench Protocol V2 Design

**Status:** normative appendix
**Date:** 2026-07-15
**Parent:** `2026-07-15-agent-workbench-v2-design.md`

## 1. Source Of Truth

Protocol V2 is defined in `proto/agent_workbench/v2`. Go, Rust, and TypeScript types are generated from that IDL. Handwritten transport DTOs may wrap generated types but may not redefine timeline, command, resource, grant, or artifact payloads.

Every extensible union contains an `unsupported` variant with the original namespace, type, schema version, and raw structured payload. Adapters preserve unknown data and render it visibly; they never drop it or reinterpret display text.

## 2. Session Identity And Ordering

```ts
type SessionRef = {
  sessionId: string;
  streamEpoch: string;
};

type EventEnvelope = SessionRef & {
  revision: bigint;
  sequence: bigint;
  turnId?: string;
  itemId: string;
  parentId?: string;
  causationCommandId?: string;
  createdAt: string;
};
```

`streamEpoch` changes whenever the server cannot preserve sequence continuity. `sequence` is contiguous within one epoch. `revision` identifies the durable session projection and increases after each atomic mutation.

## 3. Snapshot And Delta Contract

```ts
type SessionSnapshot = SessionRef & {
  revision: bigint;
  latestSequence: bigint;
  activeTurnId?: string;
  history: TimelineItem[];
  commandReceipts: CommandReceipt[];
  permissionRequests: PermissionRequest[];
  grants: AuthorizationGrant[];
  resources: SessionResource[];
  artifacts: ArtifactDescriptor[];
};

type SessionDeltaBatch = SessionRef & {
  baseRevision: bigint;
  revision: bigint;
  firstSequence: bigint;
  lastSequence: bigint;
  events: AgentEvent[];
};
```

The reducer applies a batch only when epoch, base revision, sequence range, and event order all match the current cursor. An already-applied batch is ignored only when its epoch, revision, range, and digest are identical. A duplicate with different content, a gap, an epoch change, or a base-revision mismatch enters `resync_required`; the UI remains readable but disables commands until a new snapshot is committed.

Snapshot application is atomic: validate the complete payload, replace durable state and cursor, then publish one revision notification. Batch application follows the same validate, reduce, commit, notify sequence.

## 4. Host Runtime Boundaries

The shared runtime is an interface over a session projection, not a second Web business-state store:

```ts
interface AgentSessionRuntime {
  getSnapshot(): AgentSessionView;
  subscribe(listener: () => void): () => void;
  execute(command: CommandEnvelope): Promise<CommandReceipt>;
  requestResync(reason: ResyncReason): Promise<void>;
}
```

Web implements snapshot and batch reduction in Rust Core through `applySnapshot` and `applyDeltaBatch`; TypeScript only adapts generated values and subscribes to the single Rust revision signal.

Web User starts the live tail before loading history, buffers events, fetches a revision-watermarked REST snapshot, commits it, and replays only buffered events after the watermark. Snapshot data and `latestSequence` come from one atomic server projection.

SSE event IDs are `<streamEpoch>:<sequence>` and resume through `Last-Event-ID`. JSON transports encode every 64-bit integer as a decimal string; protobuf transports use native `int64`, and generated TypeScript exposes `bigint`. The pre-snapshot buffer is capped at 5,000 events or 4 MiB, whichever comes first; overflow enters `resync_required` and discards the partial buffer.

## 5. Timeline And Content

Timeline variants are `message`, `reasoning`, `tool_execution`, `plan`, `artifact_reference`, `approval`, `status`, `error`, `system`, and `unsupported`. Session events also include typed `command_receipt_changed`, permission, resource, artifact, and terminal-lease changes.

Message content variants are `text`, `markdown`, `code`, `json`, `table`, `diff`, `command`, `log`, `progress`, `error`, `image`, `video`, `audio`, `html`, `live_preview`, `pdf`, `presentation`, `spreadsheet`, `file`, `link`, `citation`, `artifact_ref`, `restricted_iframe`, and `unsupported`.

Adapters map source records directly to these generated variants. They do not route through the current flattened `AgentTimelineItem`, `BlockStream`, or display-only tool representation.

## 6. Tool Identity

```ts
type ToolIdentity = {
  namespace: string;
  semanticKey: string;
  schemaVersion: string;
  sourceToolName?: string;
};
```

Each adapter owns a reviewed mapping catalog from source protocol tool identity to `ToolIdentity`. Missing mappings become `unsupported`; substring matching and wildcard semantic classification are forbidden.

A tool execution carries phase, structured input, progress, typed result blocks, artifact references, actions, timing, executor context, retry lineage, and approval reference. Phase values are `queued`, `running`, `waiting_approval`, `completed`, `failed`, and `cancelled`.

## 7. Commands And Idempotency

Core commands are a generated protobuf `oneof`, represented in TypeScript as:

```ts
type CoreCommand =
  | { case: "send_prompt"; value: SendPromptCommand }
  | { case: "interrupt"; value: InterruptCommand }
  | { case: "change_configuration"; value: ChangeConfigurationCommand }
  | { case: "resolve_permission"; value: ResolvePermissionCommand }
  | { case: "artifact_action"; value: ArtifactActionCommand }
  | { case: "terminal_operation"; value: TerminalOperationCommand }
  | { case: "extension"; value: ExtensionCommand };

type CommandEnvelope = SessionRef & {
  commandId: string;
  command: CoreCommand;
  payloadDigest: string;
  expectedRevision?: bigint;
  issuedAt: string;
};

type CommandReceipt = {
  commandId: string;
  state:
    | "received"
    | "accepted"
    | "running"
    | "succeeded"
    | "failed"
    | "rejected"
    | "cancelled";
  payloadDigest: string;
  resultingRevision?: bigint;
  error?: AgentError;
};
```

`ExtensionCommand` contains namespace, semantic type, schema version, and raw structured payload. Unknown extensions become `unsupported`; core commands never use a string command type with `unknown` payload.

Receipts are durable by `(sessionId, commandId)`. Repeating the same ID and digest returns the stored receipt. Reusing an ID with a different digest fails with `command_id_conflict`. Every event caused by a command includes `causationCommandId`.

Receipt transitions are monotonic: `received` may become `accepted` or `rejected`; `accepted` may become `running`, `succeeded`, `failed`, or `cancelled`; `running` may become `succeeded`, `failed`, or `cancelled`; terminal states are immutable. Each transition is a `command_receipt_changed` event and is replayable through SSE.

`POST /sessions/{sessionId}/commands` executes the generated envelope. `GET /sessions/{sessionId}/commands/{commandId}` returns the durable receipt. Snapshots include all nonterminal receipts and retained terminal receipts required by visible history; older terminal receipts remain queryable according to the server's declared retention policy.

Send, interrupt, model and mode changes, approvals, artifact actions, terminal operations, and extensions use this lifecycle. Optimistic UI may show local pending state, but durable state changes only after a receipt or event is committed.

## 8. Support And Authorization

`SupportCapabilities` declares what a runtime can represent or execute. `AuthorizationGrant` declares what this user, embed, resource, and revision may do. A visible control requires both support and a valid grant; support never implies authorization.

Grants include grant ID, issuer, subject, session, resource scope, actions, issued and expiry times, and optional revision constraints. Permission requests reference the exact command or artifact action they authorize.

## 9. Verification

Generated Go, Rust, and TypeScript fixtures must round-trip without field loss. Reducer tests cover duplicate batches, gaps, epoch changes, stale snapshots, buffered-tail replay, command replay, digest conflicts, unsupported variants, and authorization changes.

The same fixture corpus must produce equivalent visible timelines and artifact catalogs in Web, Web User, plain mount, and iframe hosts.
