# Agent Workbench Web Rust Core Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Rust Core the atomic Agent Workbench session SSOT and replace Web's TypeScript ACP reconstruction with a thin generated-protocol adapter.

**Architecture:** Relay frames decode to generated V2 snapshots or delta batches. Rust validates cursor and receipt transitions, commits one session revision, and exposes byte snapshots/selectors through WASM. Zustand stores only the revision tick and UI state.

**Tech Stack:** Rust, prost, wasm-bindgen, TypeScript, Zustand, Cargo test, Vitest.

---

### Task 1: Add The Rust Session Projection

**Files:**
- Create: `clients/core/crates/state/src/agent_workbench_state.rs`
- Create: `clients/core/crates/state/src/agent_workbench_reducer.rs`
- Create: `clients/core/crates/state/src/agent_workbench_state_tests.rs`
- Modify: `clients/core/crates/state/src/app_state.rs`
- Modify: `clients/core/crates/state/src/lib.rs`

- [ ] **Step 1: Write atomic reducer tests**

```rust
let before = manager.revision("session-1");
manager.apply_delta_batch(batch_with_two_events()).unwrap();
assert_eq!(manager.revision("session-1"), before + 1);
assert_eq!(notifications.take(), 1);
assert_eq!(manager.snapshot("session-1").history.len(), 2);
```

Add duplicate same-digest, duplicate conflicting digest, sequence gap, epoch change, stale base revision, unsupported payload, and receipt terminal-state tests.

- [ ] **Step 2: Verify failure**

Run: `(cd clients/core && cargo test -p agentcloud_state agent_workbench)`
Expected: FAIL because the Workbench state modules do not exist.

- [ ] **Step 3: Implement generated-type state**

`AgentWorkbenchState` stores generated `SessionSnapshot`, cursor, batch digests, command receipts, and status. `apply_snapshot` and `apply_delta_batch` validate the complete input before mutation. Errors are typed as `ResyncRequired`, `DigestConflict`, `ReceiptTransition`, or `InvalidPayload`.

- [ ] **Step 4: Run and commit**

Run: `(cd clients/core && cargo test -p agentcloud_state agent_workbench)`
Expected: PASS.

```bash
git add clients/core/crates/state/src/agent_workbench_state.rs clients/core/crates/state/src/agent_workbench_reducer.rs clients/core/crates/state/src/agent_workbench_state_tests.rs clients/core/crates/state/src/app_state.rs clients/core/crates/state/src/lib.rs
git commit -m "feat(core): add atomic agent workbench state"
```

### Task 2: Decode V2 Relay Messages Into Atomic Actions

**Files:**
- Modify: `clients/core/crates/protocol/src/msg_type.rs`
- Modify: `clients/core/crates/protocol/src/codec.rs`
- Modify: `clients/core/crates/protocol/src/tests.rs`
- Modify: `clients/core/crates/relay/src/types.rs`
- Modify: `clients/core/crates/relay/src/dispatch.rs`
- Modify: `clients/core/crates/relay/src/driver/session.rs`
- Create: `clients/core/crates/relay/src/agent_workbench_dispatch_tests.rs`

- [ ] **Step 1: Add snapshot and batch frame tests**

```rust
let action = decode(frame(MsgType::AgentWorkbenchDeltaBatch, bytes)).unwrap();
assert!(matches!(action, RelayAction::ApplyAgentWorkbenchDeltaBatch(_)));
assert_eq!(dispatch_count_after_two_event_batch(), 1);
```

- [ ] **Step 2: Implement message boundaries**

Add distinct message types for V2 snapshot, delta batch, command receipt, and command envelope. Codec rejects a V2 type carrying a V1 JSON body. Relay dispatch forwards one atomic action per frame and never decomposes a batch into old ACP mutations.

- [ ] **Step 3: Run and commit**

Run: `(cd clients/core && cargo test -p agentcloud_protocol -p agentcloud_relay agent_workbench)`
Expected: PASS for binary decode, wrong schema, batch dispatch, and reconnect cursor.

```bash
git add clients/core/crates/protocol/src clients/core/crates/relay/src
git commit -m "feat(core): dispatch workbench v2 relay frames"
```

### Task 3: Expose WASM Runtime Methods

**Files:**
- Create: `clients/core/crates/wasm/src/state_agent_workbench.rs`
- Create: `clients/core/crates/wasm/src/state_agent_workbench_tests.rs`
- Modify: `clients/core/crates/wasm/src/api.rs`
- Modify: `clients/core/crates/wasm/src/lib.rs`
- Modify: `clients/core/crates/wasm/src/protocol.rs`
- Modify: `clients/core/crates/wasm/src/relay_manager.rs`

- [ ] **Step 1: Write binding tests**

```rust
manager.apply_snapshot(&snapshot.encode_to_vec()).unwrap();
manager.apply_delta_batch(&batch.encode_to_vec()).unwrap();
assert_eq!(manager.revision("session-1"), 2);
assert!(!manager.snapshot_bytes("session-1").is_empty());
```

- [ ] **Step 2: Implement byte-oriented bindings**

Expose `apply_snapshot`, `apply_delta_batch`, `snapshot_bytes`, `revision`, `send_command`, and `request_resync`. Inputs and outputs are protobuf bytes; no JSON projection crosses WASM. One Rust commit increments one revision observable by React.

- [ ] **Step 3: Run and commit**

Run: `(cd clients/core && cargo test -p agentcloud_wasm state_agent_workbench) && pnpm run build:wasm`
Expected: PASS and regenerated WASM TypeScript declarations expose all six methods.

```bash
git add clients/core/crates/wasm packages/agent-cloud-wasm
git commit -m "feat(wasm): expose agent workbench runtime"
```

### Task 4: Replace Web TypeScript Reconstruction

**Files:**
- Modify: `clients/web/src/stores/relayProtocol.ts`
- Modify: `clients/web/src/stores/acpEventDispatcher.ts`
- Modify: `clients/web/src/stores/acpSession.ts`
- Modify: `clients/web/src/stores/acpSessionTypes.ts`
- Modify: `clients/web/src/stores/relayConnection.ts`
- Modify: `clients/web/src/components/workspace/AgentPanel.tsx`
- Create: `clients/web/src/components/workspace/__tests__/AgentPanel.test.tsx`
- Modify: `clients/web/src/components/workspace/agent-ui/WebAcpSessionRuntime.ts`
- Modify: `clients/web/src/components/workspace/agent-ui/WebAcpSessionRuntime.test.ts`
- Delete: `clients/web/src/components/workspace/agent-ui/webAcpSnapshot.ts`
- Delete: `clients/web/src/components/workspace/agent-ui/webAcpSnapshot.test.ts`
- Delete: `clients/web/src/components/workspace/agent-ui/webAcpPermissionProjection.ts`
- Delete: `clients/web/src/components/workspace/agent-ui/webAcpArtifactProjection.ts`

- [ ] **Step 1: Add single-SSOT tests**

```ts
expect(wasm.apply_delta_batch).toHaveBeenCalledOnce();
expect(store.getState()).toEqual(expect.objectContaining({ revisionTick: 12 }));
expect(store.getState()).not.toHaveProperty("toolCalls");
expect(runtime.getSnapshot().history[0].payload.case).toBe("toolExecution");
```

- [ ] **Step 2: Implement the thin adapter**

Relay code forwards V2 bytes to WASM. `WebAcpSessionRuntime.getSnapshot` decodes `snapshot_bytes`; subscription listens only to the Rust revision tick. `AgentPanel` mounts the shared conversation/results workbench. Commands encode generated envelopes and call `send_command`.

- [ ] **Step 3: Delete projection state**

Remove the four Web projection files and every import. `rg "projectWebAcp|webAcpSnapshot|toolPresentation\\(" clients/web/src` must return no Workbench path.

- [ ] **Step 4: Run and commit**

Run: `pnpm run web:test -- WebAcpSessionRuntime acpEventDispatcher acpSession relayConnection && pnpm run web:typecheck`
Expected: PASS with no business projection stored in Zustand.

```bash
git add clients/web/src/stores clients/web/src/components/workspace/AgentPanel.tsx clients/web/src/components/workspace/__tests__/AgentPanel.test.tsx clients/web/src/components/workspace/agent-ui
git commit -m "refactor(web): read agent workbench state from rust"
```

### Task 5: Remove The V1 ACP Workbench Contract
**Files:**
- Delete: `proto/acp_state/v1/acp_state.proto`
- Delete: `proto/gen/go/acp_state/v1/acp_state.pb.go`
- Delete: `proto/gen/ts/acp_state/v1/acp_state_pb.ts`
- Delete: `clients/core/crates/proto/acp_state`
- Delete: `clients/core/crates/state/src/acp_session.rs`
- Delete: `clients/core/crates/state/src/acp_session_tests.rs`
- Delete: `clients/core/crates/state/src/acp_types.rs`
- Delete: `clients/core/crates/wasm/src/state_acp.rs`
- Modify: `clients/core/Cargo.toml`
- Modify: `clients/core/Cargo.lock`
- Modify: `clients/core/crates/proto-gen/src/domains.rs`
- Modify: `clients/core/crates/state/src/lib.rs`
- Modify: `clients/core/crates/types/Cargo.toml`
- Modify: `clients/core/crates/types/src/lib.rs`
- Modify: `clients/core/crates/wasm/src/api.rs`
- Modify: `clients/core/crates/wasm/src/lib.rs`
- Modify: `clients/web/src/lib/__tests__/wasm-contract.test.ts`
- Modify: `clients/web/src/stores/__tests__/protoRoundtripOpaque.test.ts`
- Modify: `clients/web/src/test/wasm-mock-acp.ts`
- [ ] **Step 1: Prove all consumers migrated**
Run: `rg "acp_state_proto|proto_acp_state_v1|WasmAcpSessionManager|AcpSessionManager" clients proto`
Expected: only files listed for deletion appear.
- [ ] **Step 2: Delete and regenerate**
Run: `pnpm proto:gen-ts && pnpm proto:gen-go && pnpm proto:gen-amesh && (cd clients/core && cargo run -p agent_cloud_proto_gen)`
Expected: no generated `acp_state/v1` output remains.

- [ ] **Step 3: Run full Core/Web checks and commit**

Run: `(cd clients/core && cargo test --workspace) && pnpm run build:wasm && pnpm run web:check`
Expected: PASS.

```bash
git add proto/acp_state/v1/acp_state.proto proto/gen/go/acp_state/v1/acp_state.pb.go proto/gen/ts/acp_state/v1/acp_state_pb.ts clients/core/Cargo.toml clients/core/Cargo.lock clients/core/crates/proto-gen/src/domains.rs clients/core/crates/proto/acp_state clients/core/crates/state/src/acp_session.rs clients/core/crates/state/src/acp_session_tests.rs clients/core/crates/state/src/acp_types.rs clients/core/crates/state/src/lib.rs clients/core/crates/types/Cargo.toml clients/core/crates/types/src/lib.rs clients/core/crates/wasm/src/api.rs clients/core/crates/wasm/src/lib.rs clients/core/crates/wasm/src/state_acp.rs clients/web/src/lib/__tests__/wasm-contract.test.ts clients/web/src/stores/__tests__/protoRoundtripOpaque.test.ts clients/web/src/test/wasm-mock-acp.ts
git commit -m "refactor(core): remove the v1 acp workbench state"
```
