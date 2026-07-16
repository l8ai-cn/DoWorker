# Agent Workbench Protocol And Runtime Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace flattened conversation DTOs with generated Protocol V2 types, deterministic session reduction, durable command receipts, and exact renderer registries.

**Architecture:** Protobuf is the only wire schema. Generated Go/Rust/TypeScript types feed direct adapters. A framework-neutral runtime validates and applies snapshots or ordered delta batches; React reads it with `useSyncExternalStore`.

**Tech Stack:** Protobuf, Buf, Go, Rust/prost, TypeScript, Vitest, Cargo test.

---

### Task 1: Add The Versioned IDL And Generated Mirrors

**Files:**
- Create: `proto/agent_workbench/v2/content.proto`
- Create: `proto/agent_workbench/v2/tool.proto`
- Create: `proto/agent_workbench/v2/artifact.proto`
- Create: `proto/agent_workbench/v2/command.proto`
- Create: `proto/agent_workbench/v2/session.proto`
- Modify: `clients/core/Cargo.toml`
- Modify: `clients/core/crates/proto-gen/src/domains.rs`
- Create: `clients/core/crates/proto/agent_workbench/Cargo.toml`
- Generate: `clients/core/crates/proto/agent_workbench/src/lib.rs`
- Modify: `clients/core/crates/types/Cargo.toml`
- Modify: `clients/core/crates/types/src/lib.rs`

- [ ] **Step 1: Write schema contract tests**

Create `packages/agent-ui/src/protocol/generatedContract.test.ts`:

```ts
const command = create(CommandEnvelopeSchema, {
  commandId: "cmd-1",
  command: { case: "sendPrompt", value: { text: "build it" } },
  payloadDigest: "sha256:abc",
});
expect(command.command.case).toBe("sendPrompt");
```

Add Rust compilation use:

```rust
use agentsmesh_types::proto_agent_workbench_v2::CommandEnvelope;
let command = CommandEnvelope::default();
assert!(command.command.is_none());
```

- [ ] **Step 2: Verify failure**

Run: `pnpm proto:gen-ts && pnpm proto:gen-go && (cd clients/core && cargo test -p agentsmesh_types)`
Expected: FAIL because the V2 schema and Rust crate do not exist.

- [ ] **Step 3: Define exact unions**

`content.proto` defines all standard blocks plus `UnsupportedValue`. `tool.proto` defines exact identity, phase, progress, results, actions, and approval reference. `artifact.proto` defines descriptor, representation, provenance, and grants. `command.proto` defines the core command `oneof`, extension command, receipt, and error. `session.proto` defines envelope, snapshot, delta batch, event `oneof`, capabilities, grants, resources, and cursor values as `uint64`.

- [ ] **Step 4: Generate and commit**

Run: `pnpm proto:gen-ts && pnpm proto:gen-go && pnpm proto:gen-amesh && (cd clients/core && cargo run -p do_worker_proto_gen)`
Expected: generated Go, TypeScript, and Rust mirrors contain `proto.agent_workbench.v2`.

```bash
git add proto/agent_workbench proto/gen clients/core/Cargo.toml clients/core/crates/proto-gen clients/core/crates/proto/agent_workbench clients/core/crates/types
git commit -m "feat(protocol): generate agent workbench v2 contracts"
```

### Task 2: Add Cross-Language Fixtures And Direct Mapping

**Files:**
- Create: `packages/agent-ui/src/protocol/fixtures/sessionFixture.ts`
- Create: `packages/agent-ui/src/protocol/fixtures/sessionFixture.test.ts`
- Create: `backend/internal/api/rest/v1/session/agent_workbench_fixture_test.go`
- Create: `clients/core/crates/state/src/agent_workbench_fixture_tests.rs`
- Create: `packages/agent-ui/src/protocol/sourceToolCatalog.ts`
- Create: `packages/agent-ui/src/protocol/sourceToolCatalog.test.ts`
- Delete: `packages/agent-ui/src/toolPresentation.ts`

- [ ] **Step 1: Add a lossless fixture**

The fixture contains Markdown, image, video, presentation, tool input/results, unknown content, permission request, artifact revision, terminal lease, running and terminal receipts, and one `causationCommandId`.

```ts
expect(decoded.history.at(-1)?.payload.case).toBe("unsupported");
expect(encodeToBinary(SessionSnapshotSchema, decoded)).toEqual(bytes);
```

- [ ] **Step 2: Add exact source mapping tests**

```ts
expect(resolveSourceTool("acp", "shell")).toEqual({
  namespace: "agentsmesh.acp",
  semanticKey: "shell",
  schemaVersion: "1",
});
expect(resolveSourceTool("acp", "shell_exec")).toBeUndefined();
```

- [ ] **Step 3: Run fixtures**

Run: `pnpm --dir packages/agent-ui exec vitest run src/protocol && go test ./backend/internal/api/rest/v1/session && (cd clients/core && cargo test -p agentsmesh_state agent_workbench_fixture)`
Expected: PASS with byte-stable unsupported payloads and no substring mapping. Delete the old tool-name presentation guesser after every consumer uses exact source identities.

- [ ] **Step 4: Commit**

```bash
git add packages/agent-ui/src/protocol packages/agent-ui/src/toolPresentation.ts backend/internal/api/rest/v1/session/agent_workbench_fixture_test.go clients/core/crates/state/src/agent_workbench_fixture_tests.rs
git commit -m "test(protocol): add lossless workbench fixtures"
```

### Task 3: Implement The Framework-Neutral Runtime Reducer

**Files:**
- Create: `packages/agent-ui/src/runtime/agentSessionReducer.ts`
- Create: `packages/agent-ui/src/runtime/agentSessionReducer.test.ts`
- Create: `packages/agent-ui/src/runtime/AgentSessionRuntime.ts`
- Create: `packages/agent-ui/src/runtime/AgentSessionStore.ts`
- Create: `packages/agent-ui/src/runtime/AgentSessionStore.test.ts`
- Modify: `packages/agent-ui/src/useAgentSessionSnapshot.ts`

- [ ] **Step 1: Write gap, duplicate, epoch, and receipt tests**

```ts
expect(applyDeltaBatch(state, duplicateSameDigest)).toBe(state);
expect(() => applyDeltaBatch(state, duplicateDifferentDigest)).toThrow("delta_digest_conflict");
expect(applyDeltaBatch(state, sequenceGap).status).toBe("resync_required");
expect(() => transitionReceipt("succeeded", "running")).toThrow("receipt_terminal");
```

- [ ] **Step 2: Implement validate-reduce-commit**

```ts
export function applyDeltaBatch(state: SessionState, batch: SessionDeltaBatch): SessionState {
  const validation = validateBatch(state.cursor, batch);
  if (validation === "duplicate") return state;
  if (validation !== "apply") return { ...state, status: "resync_required" };
  return reduceEventsAtomically(state, batch);
}
```

The store publishes once after snapshot or batch commit. `execute` returns a receipt; `getCommandReceipt` and `requestResync` are required runtime methods.

- [ ] **Step 3: Run and commit**

Run: `pnpm --dir packages/agent-ui exec vitest run src/runtime`
Expected: PASS for duplicate, gap, epoch, stale revision, receipt transition, and single-notification tests.

```bash
git add packages/agent-ui/src/runtime packages/agent-ui/src/useAgentSessionSnapshot.ts
git commit -m "feat(agent-ui): add deterministic session runtime"
```

### Task 4: Implement Exact Tool And Content Registries

**Files:**
- Create: `packages/agent-ui/src/registry/ToolRendererRegistry.ts`
- Create: `packages/agent-ui/src/registry/ToolRendererRegistry.test.ts`
- Create: `packages/agent-ui/src/registry/ContentRendererRegistry.ts`
- Create: `packages/agent-ui/src/registry/ContentRendererRegistry.test.ts`
- Create: `packages/agent-ui/src/registry/rendererKeys.ts`
- Modify: `packages/agent-ui/package.json`

- [ ] **Step 1: Write exact lookup and conflict tests**

```ts
registry.register(key, renderer, "builtin");
expect(() => registry.register(key, renderer, "host")).toThrow("renderer_key_conflict");
expect(registry.lookup({ ...key, schemaVersion: "2" })).toBeUndefined();
registry.replace(key, renderer2, { expectedSourceId: "builtin", sourceId: "host" });
```

- [ ] **Step 2: Implement registries and subpath exports**

Export `./protocol`, `./runtime`, `./registry`, `./react`, and `./embed`. Iframe configuration accepts renderer IDs only; executable registrations remain React/plain-mount inputs.

- [ ] **Step 3: Verify package and budgets**

Run: `pnpm --dir packages/agent-ui exec vitest run src/registry && pnpm exec tsc --noEmit -p packages/agent-ui/tsconfig.json`
Expected: PASS with no wildcard or last-write-wins lookup.

- [ ] **Step 4: Commit**

```bash
git add packages/agent-ui/src/registry packages/agent-ui/package.json
git commit -m "feat(agent-ui): add exact renderer registries"
```
