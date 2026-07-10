# Worker Phase 0C Repository Create Enforcement Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reject unavailable or ambiguous repositories before Worker persistence and return a consistent client validation error across transports.

**Architecture:** PodOrchestrator resolves one accessible repository exactly once before quota and persistence, stores it in its internal resolved request, and reuses it for command construction. A service sentinel hides existence details while REST, Connect, and MCP map it to client input errors.

**Tech Stack:** Go, AgentFile orchestration, Gin REST, Connect RPC, MCP adapter, Bazel rules_go, testify.

---

### Task 1: Resolve the Worker Repository Before Persistence

**Files:**
- Modify: `backend/internal/service/agentpod/pod_orchestrator.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_create.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_command.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_types.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_setup_test.go`
- Create: `backend/internal/service/agentpod/pod_orchestrator_repository_test.go`
- Modify: `backend/internal/service/agentpod/BUILD.bazel`

- [x] **Step 1: Write failing direct-ID and AgentFile tests**

Add tests for inaccessible direct ID, inaccessible AgentFile `REPO`, ambiguous slug, and one successful scoped repository. Rejection must create zero Pod rows and send no command. Success must resolve once and place the repository clone/preparation data in the command.

```go
_, err := orch.CreatePod(ctx, createReqWithRepository(repoID))
assert.ErrorIs(t, err, ErrCreateResourceUnavailable)
assert.Zero(t, podCount(t, db))
assert.False(t, coordinator.createPodCalled)
```

- [x] **Step 2: Verify RED**

Run: `bazel test //backend/internal/service/agentpod:agentpod_test --test_filter='TestCreatePod_(Rejects|Uses).*Repository'`

Expected: the sentinel is undefined, direct ID persists, and AgentFile lookup ignores access errors.

- [x] **Step 3: Add the sentinel and scoped resolver contract**

Add `ErrCreateResourceUnavailable`. Replace `RepositoryServiceForOrchestrator` with:

```go
GetAccessibleByID(ctx context.Context, id, orgID, userID int64) (*gitprovider.Repository, error)
FindAccessibleByOrgSlug(ctx context.Context, orgID, userID int64, slug string) (*gitprovider.Repository, error)
```

Map inaccessible, missing, and ambiguous repository results to the sentinel without exposing the underlying cause to clients. Propagate cancellation and infrastructure errors as internal errors.

- [x] **Step 4: Resolve once before quota and persistence**

Add the resolved repository to the internal `agentfileResolved` state. Resolve AgentFile slug immediately when parsing `REPO`; resolve the final direct ID after merge and before quota. If an ID is present and the service is nil, fail closed. `buildPodCommand` uses the resolved object and performs no unscoped lookup.

- [x] **Step 5: Verify GREEN and regression**

Run: `bazel test //backend/internal/service/agentpod:agentpod_test --nocache_test_results`

Run: `bazel run //:buildifier_check`

Expected: PASS.

- [x] **Step 6: Commit**

```bash
git add backend/internal/service/agentpod
git commit -m "fix(worker): reject unavailable repositories"
```

### Task 2: Map Repository Validation Across Transports

**Files:**
- Modify: `backend/internal/api/rest/v1/pod_create.go`
- Modify: `backend/internal/api/rest/v1/pod_create_test.go`
- Modify: `backend/internal/api/connect/pod/mount.go`
- Modify: `backend/internal/api/connect/pod/server_test.go`
- Modify: `backend/internal/api/grpc/runner_adapter_mcp_pod.go`
- Modify: `backend/internal/api/grpc/runner_adapter_mcp_ticket_test.go`

- [x] **Step 1: Write failing transport mapping tests**

Add table cases for `ErrCreateResourceUnavailable`:

```go
{name: "repository unavailable", err: agentpod.ErrCreateResourceUnavailable, wantStatus: http.StatusBadRequest}
{name: "repository unavailable", err: agentpod.ErrCreateResourceUnavailable, wantCode: connect.CodeInvalidArgument}
```

The MCP assertion requires code 400 and a generic unavailable message.

- [x] **Step 2: Verify RED**

Run: `bazel test //backend/internal/api/rest/v1:rest_test --test_filter=TestMapOrchestratorErrorToHTTP`

Run: `bazel test //backend/internal/api/connect/pod:pod_test --test_filter=TestMapServiceError`

Run: `bazel test //backend/internal/api/grpc:grpc_test --test_filter=TestMapOrchestratorErrorToMCP`

Expected: all transports currently map the sentinel to internal error behavior.

- [x] **Step 3: Add explicit mappings**

REST returns HTTP 400 with `VALIDATION_FAILED` and `Selected repository is unavailable`. Connect returns `CodeInvalidArgument`. MCP returns code 400 with the same generic meaning. Do not include repository ID, slug, organization, or underlying permission error.

- [x] **Step 4: Verify GREEN and transport regression**

Run the three scoped commands, then:

```bash
bazel test //backend/internal/api/rest/v1:rest_test \
  //backend/internal/api/connect/pod:pod_test \
  //backend/internal/api/grpc:grpc_test --nocache_test_results
```

Expected: PASS.

- [x] **Step 5: Commit**

```bash
git add backend/internal/api/rest/v1 backend/internal/api/connect/pod backend/internal/api/grpc
git commit -m "fix(worker): map repository validation errors"
```
