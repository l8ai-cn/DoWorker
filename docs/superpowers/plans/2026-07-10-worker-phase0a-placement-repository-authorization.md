# Worker Phase 0A Placement and Repository Authorization Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reject ineligible explicit Runners and inaccessible repositories before a Worker row or command is created.

**Architecture:** Runner selection and exact-ID selection share one service eligibility policy, with offline/capacity relaxation only for explicit queueing. Repository resolution gains org-and-user scoped methods used by both AgentFile and direct-ID paths.

**Tech Stack:** Go, Gin service layer, GORM repositories, Bazel rules_go, testify.

---

### Task 1: Resolve an Explicit Runner Through Eligibility Policy

**Files:**
- Modify: `backend/internal/service/runner/query_eligible.go`
- Modify: `backend/internal/service/runner/query_visibility_test.go`

- [x] **Step 1: Write the failing service test**

Add `TestResolveRunnerForCreate` with table cases for eligible, other organization, invisible private, disabled, unsupported Agent, at capacity, offline, and queued offline. Construct each Runner in `activeRunners` and the test DB. Assert only `allowUnavailable=true` relaxes online, heartbeat, and capacity.

```go
got, err := service.ResolveRunnerForCreate(ctx, r.ID, 1, userID, "codex-cli", false)
require.NoError(t, err)
assert.Equal(t, r.ID, got.ID)

_, err = service.ResolveRunnerForCreate(ctx, otherOrg.ID, 1, userID, "codex-cli", false)
assert.ErrorIs(t, err, ErrNoRunnerForAgent)
```

- [x] **Step 2: Verify RED**

Run: `bazel test //backend/internal/service/runner:runner_test --test_filter=TestResolveRunnerForCreate`

Expected: compile failure because `ResolveRunnerForCreate` does not exist.

- [x] **Step 3: Implement the exact-ID resolver**

Add this public method and keep detailed conditions private:

```go
func (s *Service) ResolveRunnerForCreate(
    ctx context.Context,
    runnerID, orgID, userID int64,
    agentSlug string,
    allowUnavailable bool,
) (*runnerDomain.Runner, error)
```

Use `s.repo.ListByOrg(ctx, orgID, userID)` to establish org/private/grant visibility. Always require `IsEnabled` and `SupportsAgent(agentSlug)`. With `allowUnavailable=false`, require the exact ID to appear in `collectEligibleRunners`; with `true`, allow offline/stale/full but no other relaxation. Return `ErrNoRunnerForAgent` for all rejected IDs.

- [x] **Step 4: Verify GREEN**

Run: `bazel test //backend/internal/service/runner:runner_test --test_filter=TestResolveRunnerForCreate`

Expected: PASS.

- [x] **Step 5: Commit**

```bash
git add backend/internal/service/runner/query_eligible.go backend/internal/service/runner/query_visibility_test.go
git commit -m "fix(runner): scope explicit worker placement"
```

### Task 2: Enforce Runner Eligibility in PodOrchestrator

**Files:**
- Modify: `backend/internal/service/agentpod/pod_orchestrator.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_create.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_setup_test.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_create_test.go`

- [x] **Step 1: Replace the bypass test with failing behavior tests**

Replace `TestCreatePod_ExplicitRunnerID_SkipsAutoSelect` with tests asserting the exact-ID resolver receives runner/org/user/agent and `QueueIfUnavailable`; rejection returns `ErrNoAvailableRunner`, creates zero Pods, and never dispatches.

```go
_, err := orch.CreatePod(ctx, &OrchestrateCreatePodRequest{
    OrganizationID: 1, UserID: 7, RunnerID: 9,
    AgentSlug: "codex-cli", QueueIfUnavailable: true,
})
assert.ErrorIs(t, err, ErrNoAvailableRunner)
assert.False(t, coord.createPodCalled)
var podCount int64
require.NoError(t, db.Model(&podDomain.Pod{}).Count(&podCount).Error)
assert.Zero(t, podCount)
```

- [x] **Step 2: Verify RED**

Run: `bazel test //backend/internal/service/agentpod:agentpod_test --test_filter=TestCreatePod_ExplicitRunner`

Expected: explicit Runner still bypasses validation and creates a Pod.

- [x] **Step 3: Extend the orchestrator dependency and validate before agent/model resolution**

Add `ResolveRunnerForCreate` to `RunnerSelectorForOrchestrator`. In the non-resume path, call it for nonzero `RunnerID`; pass `req.QueueIfUnavailable` as `allowUnavailable`. Missing resolver or any resolver error returns `ErrNoAvailableRunner`. Keep automatic affinity selection unchanged.

- [x] **Step 4: Verify GREEN and regression**

Run: `bazel test //backend/internal/service/agentpod:agentpod_test --test_filter='TestCreatePod_(ExplicitRunner|AutoSelectRunner)'`

Expected: PASS.

- [x] **Step 5: Commit**

```bash
git add backend/internal/service/agentpod
git commit -m "fix(worker): validate explicit runner placement"
```

### Task 3: Add Scoped Repository Resolvers

**Files:**
- Modify: `backend/internal/service/repository/interfaces.go`
- Modify: `backend/internal/service/repository/service.go`
- Modify: `backend/internal/service/repository/service_crud_test.go`

- [ ] **Step 1: Write failing access tests**

Add `TestGetAccessibleByID` and `TestFindAccessibleByOrgSlug`. Cover organization-visible in the same org, private imported by caller, private imported by another user, and organization-visible in another org.

```go
got, err := service.GetAccessibleByID(ctx, repo.ID, orgID, importerID)
require.NoError(t, err)
assert.Equal(t, repo.ID, got.ID)

_, err = service.GetAccessibleByID(ctx, repo.ID, otherOrgID, importerID)
assert.ErrorIs(t, err, ErrNoPermission)
```

- [ ] **Step 2: Verify RED**

Run: `bazel test //backend/internal/service/repository:repository_test --test_filter='Test(Get|Find)Accessible'`

Expected: compile failure because scoped resolvers do not exist.

- [ ] **Step 3: Implement shared access validation**

Add both methods to `RepositoryServiceInterface`:

```go
GetAccessibleByID(ctx context.Context, id, orgID, userID int64) (*gitprovider.Repository, error)
FindAccessibleByOrgSlug(ctx context.Context, orgID, userID int64, slug string) (*gitprovider.Repository, error)
```

Both require `repo.OrganizationID == orgID`. A private repository additionally requires `ImportedByUserID != nil && *ImportedByUserID == userID`. Return `ErrNoPermission` for nonexistent or inaccessible resources at this boundary.

- [ ] **Step 4: Verify GREEN**

Run: `bazel test //backend/internal/service/repository:repository_test --test_filter='Test(Get|Find)Accessible'`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/repository
git commit -m "fix(repository): scope worker repository access"
```

### Task 4: Reject Inaccessible Worker Repositories Before Persistence

**Files:**
- Modify: `backend/internal/service/agentpod/pod_orchestrator.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_create.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_command.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_setup_test.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_create_test.go`

- [ ] **Step 1: Write failing direct-ID and AgentFile tests**

Add `TestCreatePod_RejectsInaccessibleRepositoryID` and `TestCreatePod_RejectsInaccessibleAgentfileRepository`. Each fake resolver returns `repository.ErrNoPermission`; assert zero Pod rows and no dispatch.

- [ ] **Step 2: Verify RED**

Run: `bazel test //backend/internal/service/agentpod:agentpod_test --test_filter='TestCreatePod_RejectsInaccessible.*Repository'`

Expected: direct ID creates a Pod and AgentFile lookup ignores visibility.

- [ ] **Step 3: Use only scoped repository methods**

Replace `RepositoryServiceForOrchestrator` methods with the two scoped methods. Resolve AgentFile `REPO` through `FindAccessibleByOrgSlug`. Validate the effective direct repository ID after AgentFile merge and before quota or `PodService.CreatePod`. In `buildPodCommand`, use `GetAccessibleByID` and propagate errors instead of silently omitting clone data.

- [ ] **Step 4: Verify GREEN and package regression**

Run: `bazel test //backend/internal/service/agentpod:agentpod_test`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/agentpod
git commit -m "fix(worker): reject inaccessible repositories"
```
