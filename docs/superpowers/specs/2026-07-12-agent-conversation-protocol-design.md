# Agent Conversation Protocol Design

**Status:** proposed
**Date:** 2026-07-12
**Depends on:** `2026-07-12-agent-conversation-component-design.md`

## 1. Protocol Split

| Plane | Protocol | Owner | Browser use |
| --- | --- | --- | --- |
| Session command | Connect-RPC unary | Backend session service | create, send, interrupt, approve, configure |
| Session state | Connect-RPC unary + server stream | Backend session service | snapshot, history, resumable events |
| Terminal data | Relay binary WebSocket | Rust Relay runtime | PTY output/input/resize/resync/control lease |
| Runner control | Existing Backend to Runner command path | Backend | never called by UI |

The conversation UI does not consume Relay ACP events. Runner ACP events enter
Backend, are normalized into canonical session events, update durable items or
the active-turn buffer, and are then published through `WatchSession`.

## 2. Service

Add `proto/agent_session/v1/agent_session.proto`:

```text
service AgentSessionService {
  rpc CreateSession(CreateSessionRequest) returns (CreateSessionResponse);
  rpc GetSessionSnapshot(GetSessionSnapshotRequest) returns (SessionSnapshot);
  rpc ListSessionItems(ListSessionItemsRequest) returns (ListSessionItemsResponse);
  rpc WatchSession(WatchSessionRequest) returns (stream SessionEvent);
  rpc SendMessage(SendMessageRequest) returns (CommandReceipt);
  rpc SendSlashCommand(SendSlashCommandRequest) returns (CommandReceipt);
  rpc InterruptSession(InterruptSessionRequest) returns (CommandReceipt);
  rpc ResolvePermission(ResolvePermissionRequest) returns (CommandReceipt);
  rpc UpdateSessionConfiguration(UpdateSessionConfigurationRequest)
      returns (CommandReceipt);
}
```

Existing session listing, sharing, policies, comments, filesystem, and admin
operations may migrate separately. They are not prerequisites for the atomic
conversation runtime.

## 3. Identifiers and Ordering

- `session_id`, `pod_key`, `item_id`, `response_id`, `elicitation_id`, and
  `command_id` are explicit fields, never inferred from display names.
- `command_id` is a caller-generated UUID used as the idempotency key.
- `event_sequence` is a monotonically increasing `uint64` per session.
- `state_revision` increments for every canonical session-state mutation.
- `event_id` is an opaque stable identifier used for tracing and diagnostics.
- Events include `causation_command_id` when caused by a browser command.

The Backend stores command receipts by `(session_id, command_id)`. Repeating
the same command returns the original receipt and cannot dispatch the Runner
again. Reusing a command ID with a different payload returns
`ALREADY_EXISTS`.

## 4. Snapshot

`SessionSnapshot` includes:

```text
session
state_revision
latest_event_sequence
items[]
history_page_info
active_turn
pending_permissions[]
pending_commands[]
configuration
capabilities
usage
resources[]
presence[]
child_sessions[]
```

`active_turn` contains `response_id`, status, accumulated assistant text,
reasoning, plan, tool calls, started time, and last update time. It lets a
reconnecting browser resume a live turn without replaying every text delta.

`pending_commands` contains accepted commands that have not reached their
terminal result. It includes `command_id`, kind, created time, optional
`item_id`, and delivery state.

## 5. History

`ListSessionItems` uses an exclusive `before_item_id` cursor and a bounded
limit. The response returns ordered durable items and the next cursor.

Conversation items are a protobuf `oneof`:

```text
message
reasoning
tool_call
tool_result
native_tool
error
compaction
slash_command
routing_decision
terminal_command
file
system
```

Unknown future item variants remain unknown protobuf fields and do not crash
the reducer. The UI may render a generic unsupported-item row in development,
but production does not invent a substitute semantic type.

## 6. Events

`SessionEvent` carries the common envelope plus a `oneof payload`:

| Event | Meaning |
| --- | --- |
| `session_status_changed` | running, waiting, idle, failed, stopped |
| `command_state_changed` | accepted, delivered, completed, denied, failed |
| `item_appended` | durable item added |
| `item_updated` | durable partial/final item changed |
| `assistant_text_delta` | transient active response text |
| `reasoning_delta` | transient active reasoning |
| `tool_call_changed` | tool lifecycle and arguments |
| `plan_changed` | current Agent plan |
| `permission_requested` | durable pending elicitation |
| `permission_resolved` | elicitation terminal result |
| `configuration_changed` | model, effort, permission mode, collaboration mode |
| `usage_changed` | token and cost aggregate |
| `resource_changed` | terminal, file, or environment resource |
| `presence_changed` | current viewers |
| `child_session_changed` | sub-agent session lifecycle |
| `sandbox_changed` | managed launch state |
| `session_superseded` | canonical replacement session |

Every state-changing event carries the resulting `state_revision`. Deltas for
an active response carry `response_id`. Item events carry `item_id`.

## 7. Stream Resume

`WatchSession(session_id, after_event_sequence)` emits only events with a
larger sequence. Events for one session are strictly ordered.

If the cursor is no longer retained, the server returns
`FAILED_PRECONDITION` with reason `SESSION_CURSOR_EXPIRED` and current
revision/sequence metadata. The runtime then explicitly fetches a new snapshot
and reopens the stream from that sequence. This is a protocol-defined resync,
not a silent transport fallback.

On network loss, the Rust runtime reconnects with the last fully reduced
sequence. Duplicate sequences are ignored; a forward gap triggers snapshot
resync. Reducer application and cursor advancement occur atomically.

Command execution, permission resolution, terminal framing, stable errors, and
delivery verification are specified in
`2026-07-12-agent-conversation-delivery-design.md`.
