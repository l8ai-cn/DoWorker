# Agent Workbench Backend Session Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide durable V2 session snapshots, replayable ordered deltas, idempotent commands, and authorization-aware artifact actions.

**Architecture:** Each session has one persisted projection row and append-only event log. Event append, command receipt transition, projection bytes, revision, and sequence commit in one database transaction. SSE replays by epoch/sequence before joining the live Hub.

**Tech Stack:** Go, Gin, GORM/PostgreSQL, protobuf, SSE, testify.

---

### Task 1: Add Durable Session Stream Tables

**Files:**
- Create: `backend/migrations/000217_agent_workbench_stream.up.sql`
- Create: `backend/migrations/000217_agent_workbench_stream.down.sql`
- Create: `backend/migrations/agent_workbench_stream_test.go`
- Create: `backend/internal/domain/agentworkbench/session_state.go`
- Create: `backend/internal/domain/agentworkbench/repository.go`
- Create: `backend/internal/infra/agent_workbench_repo.go`
- Create: `backend/internal/infra/agent_workbench_repo_test.go`

- [ ] **Step 1: Write migration and atomic append tests**

```go
require.NoError(t, repo.Append(ctx, sessionID, expectedRevision, events, projection))
state, _ := repo.GetSnapshot(ctx, sessionID)
require.Equal(t, uint64(1), state.Revision)
require.Equal(t, uint64(2), state.LatestSequence)
require.Len(t, repo.ListAfter(ctx, sessionID, epoch, 0, 10), 2)
```

- [ ] **Step 2: Verify failure**

Run: `go test ./backend/migrations ./backend/internal/infra -run AgentWorkbench`
Expected: FAIL because tables and repository do not exist.

- [ ] **Step 3: Implement schema**

Create `agent_workbench_session_states` with session ID PK, epoch, revision, latest sequence, projection bytes, digest, and timestamps. Create `agent_workbench_events` with unique `(session_id, stream_epoch, sequence)`, revision, payload bytes, digest, and causation command ID. Create `agent_workbench_command_receipts` with PK `(session_id, command_id)`, payload digest, state, receipt bytes, and timestamps.

- [ ] **Step 4: Run and commit**

Run: `go test ./backend/migrations ./backend/internal/infra -run AgentWorkbench`
Expected: PASS including revision conflict rollback and duplicate sequence rejection.

```bash
git add backend/migrations/000217_agent_workbench_stream.* backend/migrations/agent_workbench_stream_test.go backend/internal/domain/agentworkbench backend/internal/infra/agent_workbench_repo.go backend/internal/infra/agent_workbench_repo_test.go
git commit -m "feat(backend): persist workbench session streams"
```

### Task 2: Build The Atomic Projection Service

**Files:**
- Create: `backend/internal/service/agentworkbench/projector.go`
- Create: `backend/internal/service/agentworkbench/projector_test.go`
- Create: `backend/internal/service/agentworkbench/runner_event_mapper.go`
- Create: `backend/internal/service/agentworkbench/runner_event_mapper_test.go`
- Modify: `backend/internal/api/rest/v1/session/pod_event_sink.go`
- Modify: `backend/internal/api/rest/v1/session/session_stream_publisher.go`
- Modify: `backend/internal/api/rest/v1/session/session_stream_acp.go`
- Modify: `backend/internal/api/rest/v1/session/session_stream_assistant.go`

- [ ] **Step 1: Add direct mapping tests**

```go
event := mapper.ToolCall(acpToolCallWithImageAndUnknownBlock())
require.Equal(t, "agentcloud.acp", event.GetToolExecution().Identity.Namespace)
require.NotNil(t, event.GetToolExecution().ResultBlocks[1].GetUnsupported())
```

- [ ] **Step 2: Implement one transaction per logical batch**

Map Runner/ACP events directly to generated V2 events. Projector loads the current projection, validates expected revision, applies the complete batch, stores projection and event rows, then publishes the committed batch. Missing source-tool mappings produce `unsupported`.

- [ ] **Step 3: Verify and commit**

Run: `go test ./backend/internal/service/agentworkbench ./backend/internal/api/rest/v1/session -run Workbench`
Expected: PASS for tool blocks, permissions, receipts, artifacts, unsupported payloads, and atomic publish-after-commit.

```bash
git add backend/internal/service/agentworkbench backend/internal/api/rest/v1/session
git commit -m "feat(backend): project runner events into workbench v2"
```

### Task 3: Add Snapshot And Replayable SSE Endpoints

**Files:**
- Create: `backend/internal/api/rest/v1/session/session_workbench_snapshot.go`
- Create: `backend/internal/api/rest/v1/session/session_workbench_snapshot_test.go`
- Create: `backend/internal/api/rest/v1/session/session_workbench_stream.go`
- Create: `backend/internal/api/rest/v1/session/session_workbench_stream_test.go`
- Modify: `backend/internal/api/rest/v1/session/routes.go`
- Modify: `backend/internal/api/rest/v1/session/session_hub.go`

- [ ] **Step 1: Write watermark and replay tests**

```go
require.Equal(t, `"41"`, snapshot.LatestSequence)
require.Equal(t, "epoch-a:42", response.Header().Get("Last-Event-ID"))
require.Equal(t, []uint64{42, 43}, decodedSequences(response.Body))
```

- [ ] **Step 2: Implement endpoints**

Add `GET /sessions/:id/workbench/snapshot` and `GET /sessions/:id/workbench/stream` to authenticated and embed route groups. Snapshot returns protobuf or JSON with decimal uint64 strings from one repository read. SSE parses `Last-Event-ID`, replays persisted events after the cursor, then subscribes to the live Hub without an event gap.

- [ ] **Step 3: Run and commit**

Run: `go test ./backend/internal/api/rest/v1/session -run 'WorkbenchSnapshot|WorkbenchStream'`
Expected: PASS for exact watermark, reconnect, epoch mismatch resync, authorization, and embed read-only access.

```bash
git add backend/internal/api/rest/v1/session/session_workbench_snapshot* backend/internal/api/rest/v1/session/session_workbench_stream* backend/internal/api/rest/v1/session/routes.go backend/internal/api/rest/v1/session/session_hub.go
git commit -m "feat(api): expose workbench snapshot and replay stream"
```

### Task 4: Add Idempotent Command Endpoints

**Files:**
- Create: `backend/internal/service/agentworkbench/command_service.go`
- Create: `backend/internal/service/agentworkbench/command_service_test.go`
- Create: `backend/internal/api/rest/v1/session/session_workbench_commands.go`
- Create: `backend/internal/api/rest/v1/session/session_workbench_commands_test.go`
- Modify: `backend/internal/api/rest/v1/session/routes.go`
- Modify: `backend/internal/api/rest/v1/session/session_events_message.go`
- Modify: `backend/internal/api/rest/v1/session/session_events_interrupt.go`
- Modify: `backend/internal/api/rest/v1/session/session_elicitations.go`

- [ ] **Step 1: Write command replay tests**

```go
first := service.Execute(ctx, command("cmd-1", "sha256:a"))
second := service.Execute(ctx, command("cmd-1", "sha256:a"))
require.Equal(t, first, second)
require.ErrorIs(t, service.Execute(ctx, command("cmd-1", "sha256:b")), ErrCommandIDConflict)
```

- [ ] **Step 2: Implement generated command routing**

Add `POST /sessions/:id/commands` and `GET /sessions/:id/commands/:command_id`. Validate support plus authorization grant, insert `received`, route the generated command case to existing message/interrupt/configuration/permission/artifact/terminal services, and publish each monotonic receipt transition.

- [ ] **Step 3: Verify and commit**

Run: `go test ./backend/internal/service/agentworkbench ./backend/internal/api/rest/v1/session -run WorkbenchCommand`
Expected: PASS for same-digest replay, digest conflict, terminal immutability, stale revision, read-only embed, and causation ID.

```bash
git add backend/internal/service/agentworkbench backend/internal/api/rest/v1/session
git commit -m "feat(api): execute idempotent workbench commands"
```

### Task 5: Retire The V1 Session Event Route

**Files:**
- Delete: `backend/internal/api/rest/v1/session/session_events.go`
- Delete: `backend/internal/api/rest/v1/session/session_events_delivery_test.go`
- Delete: `backend/internal/api/rest/v1/session/embed_event_capability_test.go`
- Modify: `backend/internal/api/rest/v1/session/routes.go`
- Modify: `backend/internal/api/rest/v1/session/embed_context_validation.go`

- [ ] **Step 1: Prove clients migrated**

Run: `rg 'sessions/.*/events|handlePostEvent' clients packages backend --glob '!backend/internal/api/rest/v1/session/session_events.go'`
Expected: no client or route consumer remains.

- [ ] **Step 2: Delete V1 and run session suites**

Run: `go test ./backend/internal/api/rest/v1/session ./backend/internal/service/agentworkbench`
Expected: PASS with all commands using the generated envelope.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/api/rest/v1/session/session_events.go backend/internal/api/rest/v1/session/session_events_delivery_test.go backend/internal/api/rest/v1/session/embed_event_capability_test.go backend/internal/api/rest/v1/session/routes.go backend/internal/api/rest/v1/session/embed_context_validation.go
git commit -m "refactor(api): remove the v1 session event route"
```
