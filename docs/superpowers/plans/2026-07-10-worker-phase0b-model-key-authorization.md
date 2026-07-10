# Worker Phase 0B Model and Virtual-Key Authorization Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prevent Worker and session creation from resolving model credentials or virtual keys outside the caller's organization and user scope.

**Architecture:** AI model lookup exposes a visible-row boundary before decryption. Virtual-key creation and resolution require exact key ownership and a visible underlying model; caller scope is propagated through every orchestration interface.

**Tech Stack:** Go, GORM, Gin session API, Bazel rules_go, testify.

---

### Task 1: Resolve Only Visible AI Models

**Files:**
- Modify: `backend/internal/domain/aimodel/repository.go`
- Modify: `backend/internal/infra/ai_model_repo.go`
- Create: `backend/internal/infra/ai_model_repo_test.go`
- Modify: `backend/internal/infra/BUILD.bazel`
- Modify: `backend/internal/service/aimodel/service.go`
- Create: `backend/internal/service/aimodel/model_resolution.go`
- Create: `backend/internal/service/aimodel/service_test.go`
- Modify: `backend/internal/service/aimodel/BUILD.bazel`

- [ ] **Step 1: Write failing service tests**

Create a complete fake `aimodel.Repository` and `TestResolveVisible`. Cover same-org shared, current-user private, other-org, other-user, disabled, and missing models. Track decrypt calls through an Encryptor fixture or malformed encrypted JSON so an inaccessible row proves it was rejected before decryption.

```go
resolved, err := service.ResolveVisible(ctx, model.ID, userID, orgID)
require.NoError(t, err)
assert.Equal(t, model.ID, resolved.Model.ID)

_, err = service.ResolveVisible(ctx, foreign.ID, userID, orgID)
assert.ErrorIs(t, err, ErrNotFound)
```

- [ ] **Step 2: Verify RED**

Run: `bazel test //backend/internal/service/aimodel:aimodel_test --test_filter=TestResolveVisible`

Expected: target or method does not exist.

- [ ] **Step 3: Add repository and service boundaries**

Extend the repository with:

```go
GetVisibleByID(ctx context.Context, id, userID, orgID int64) (*AIModel, error)
```

The GORM query is `id = ? AND is_enabled = ? AND (organization_id = ? OR user_id = ?)`. Add `Service.GetVisible` for non-secret validation and `Service.ResolveVisible` that calls it before `resolveRow`. Return `ErrNotFound` for missing or invisible rows.

Move model-resolution types and methods from the already-near-limit `service.go` into `model_resolution.go`; keep existing unscoped method behavior unchanged until its callers are migrated in Task 2. Both production files must remain below 200 lines.

- [ ] **Step 4: Verify GREEN and infra build**

Run: `bazel test //backend/internal/service/aimodel:aimodel_test`

Run: `bazel test //backend/internal/infra:infra_test --test_filter=AIModel`

The infra test must exercise the real GORM query for same-org shared, current-user private, other-org, other-user, disabled, and missing rows. Expected: service and infra tests PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain/aimodel backend/internal/infra/ai_model_repo.go backend/internal/service/aimodel
git commit -m "fix(model): scope credential resolution"
```

### Task 2: Propagate Model Scope Through Worker and Session Creation

**Files:**
- Modify: `backend/internal/service/agentpod/pod_orchestrator_worker_model.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_worker_model_test.go`
- Modify: `backend/internal/api/rest/v1/session/session_worker_model.go`

- [ ] **Step 1: Write the failing orchestrator test**

Change the fake pool to record ID, user ID, and organization ID. Add `TestResolvePoolModel_ExplicitModelUsesCallerScope`.

```go
_, _, err := orchestrator.resolvePoolModel(ctx, req, agentDef)
require.NoError(t, err)
assert.Equal(t, []int64{modelID, userID, orgID}, pool.resolveVisibleArgs)
```

- [ ] **Step 2: Verify RED**

Run: `bazel test //backend/internal/service/agentpod:agentpod_test --test_filter=TestResolvePoolModel_ExplicitModelUsesCallerScope`

Expected: fake interface only receives model ID.

- [ ] **Step 3: Replace unscoped resolution**

Change `AIModelPoolForOrchestrator.Resolve` to:

```go
ResolveVisible(ctx context.Context, id, userID, orgID int64) (*aimodelsvc.ResolvedModel, error)
```

Pass `req.UserID` and `req.OrganizationID`. In `sessionapi.resolveWorkerModel`, replace `AIModels.Resolve` with `ResolveVisible(ctx, id, userID, orgID)` using the authenticated session scope already passed to the function.

- [ ] **Step 4: Verify GREEN and session compilation**

Run: `bazel test //backend/internal/service/agentpod:agentpod_test --test_filter='Test(ResolvePoolModel|ApplyWorkerModel)'`

Run: `bazel test //backend/internal/api/rest/v1/session:sessionapi_test`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/agentpod backend/internal/api/rest/v1/session/session_worker_model.go
git commit -m "fix(worker): propagate model visibility scope"
```

### Task 3: Scope Virtual-Key Creation and Resolution

**Files:**
- Modify: `backend/internal/domain/virtualkey/repository.go`
- Modify: `backend/internal/infra/virtual_api_key_repo.go`
- Modify: `backend/internal/service/virtualkey/service.go`
- Create: `backend/internal/service/virtualkey/service_test.go`
- Modify: `backend/internal/service/virtualkey/BUILD.bazel`

- [ ] **Step 1: Write failing service tests**

Create complete fake repositories and add `TestCreateRejectsInvisibleModel` plus `TestResolveModelForScope`. Cover wrong org, wrong user, revoked key, invisible underlying model, and success. Assert `TouchLastUsed` runs only after full success.

```go
_, _, err := service.ResolveModelForScope(ctx, key.ID, orgID, userID)
require.NoError(t, err)
assert.Equal(t, 1, repo.touchCalls)

_, _, err = service.ResolveModelForScope(ctx, key.ID, otherOrgID, userID)
assert.ErrorIs(t, err, ErrNotFound)
assert.Equal(t, 1, repo.touchCalls)
```

- [ ] **Step 2: Verify RED**

Run: `bazel test //backend/internal/service/virtualkey:virtualkey_test --test_filter='Test(CreateRejectsInvisibleModel|ResolveModelForScope)'`

Expected: scoped repository and service methods do not exist.

- [ ] **Step 3: Implement exact key scope and underlying-model validation**

Add repository method:

```go
GetByIDForScope(ctx context.Context, id, orgID, userID int64) (*VirtualAPIKey, error)
```

Query exact `id`, `organization_id`, and `user_id`. `Create` calls `models.GetVisible` before minting or persistence. `ResolveModelForScope` loads the scoped active key, calls `models.ResolveVisible`, and touches usage only after success. Return `ErrNotFound` for invisible keys.

- [ ] **Step 4: Verify GREEN**

Run: `bazel test //backend/internal/service/virtualkey:virtualkey_test`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain/virtualkey backend/internal/infra/virtual_api_key_repo.go backend/internal/service/virtualkey
git commit -m "fix(model): scope virtual key resolution"
```

### Task 4: Propagate Virtual-Key Scope Into Worker Creation

**Files:**
- Modify: `backend/internal/service/agentpod/pod_orchestrator_worker_model.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_worker_model_test.go`

- [ ] **Step 1: Write the failing propagation test**

Add `TestResolvePoolModel_VirtualKeyUsesCallerScope`; the fake records key ID, organization ID, and user ID.

- [ ] **Step 2: Verify RED**

Run: `bazel test //backend/internal/service/agentpod:agentpod_test --test_filter=TestResolvePoolModel_VirtualKeyUsesCallerScope`

Expected: current interface receives only key ID.

- [ ] **Step 3: Replace the interface and call**

Change `VirtualKeyPoolForOrchestrator.ResolveModel` to `ResolveModelForScope(ctx, keyID, orgID, userID)` and pass request scope without reordering it.

- [ ] **Step 4: Verify GREEN and package regression**

Run: `bazel test //backend/internal/service/agentpod:agentpod_test`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/agentpod/pod_orchestrator_worker_model.go backend/internal/service/agentpod/pod_orchestrator_worker_model_test.go
git commit -m "fix(worker): scope virtual key binding"
```
