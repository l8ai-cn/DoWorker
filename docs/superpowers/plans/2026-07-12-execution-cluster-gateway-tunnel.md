# Execution Cluster and Gateway Tunnel Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将线上与本地作为同一组织下两个可见、可调度的逻辑执行集群；Runner 归属集群并通过现有 Relay 的反向隧道主动接入，使 Worker 创建、运行状态和管理界面具有一致且可验证的行为。

**Architecture:** `ExecutionCluster` 是组织内的调度边界，`Runner` 是 Cluster 内的执行节点；创建 Worker 时提交 `execution_cluster_id`，后端只在该 Cluster 内选择兼容且在线的 Runner，并将 Cluster 写入 Pod 审计快照。Runner 继续使用现有 gRPC 控制面和 Relay WebSocket 数据面；Relay 负责持有实时隧道注册表，Backend 通过 Runner 的持续状态上报和启动重调度得到可展示的隧道状态。Preview 仍只能访问显式声明的 Pod loopback 服务，绝不扩展为任意主机/端口代理。

**Tech Stack:** Go 1.24、Gin/Connect RPC、GORM、PostgreSQL、gRPC、WebSocket、Next.js、TypeScript、Rust/WASM Core、Vitest、Go test、Playwright。

---

## 范围与不变量

- 一个组织的默认 Cluster 固定为 `online` 与 `local`；这是两个逻辑集群，不是两个 Runner，也不是“每个 Agent 一个集群”。
- 一个 Cluster 可包含任意数量 Runner。当前线上 Codex、Claude、MiniMax、e2e Runner 都归入 `online`。
- 本地 Runner 使用 `runner register --server https://dowork.l8ai.cn --token ...` 主动建立 mTLS gRPC 和 Gateway WebSocket；不要求本机暴露公网端口。
- V1 由 Backend 集中调度 Runner；不部署第二个“集群调度服务”。未来 Kubernetes Cluster 才增加 Cluster-side adapter，不在本次实现。
- Cluster 是组织级边界；没有现有项目领域模型时，不新增虚假的“项目绑定”表。
- 不能把 doops 的鉴权、Token、协议或进程引入 AgentCloud；仅复用“Agent 主动出站连接、Gateway 负责授权和可观测性”的拓扑原则。
- 不使用兼容/降级路径掩盖 Cluster 缺失：新建 Worker 指定 Cluster 后，Cluster 无可用 Runner 必须返回明确不可调度错误。
- 本计划不删除任何远端 Runner、Pod、证书或历史数据。数据修正通过可回滚迁移和部署后核验完成。

## 文件结构

| 路径 | 职责 |
| --- | --- |
| `backend/internal/domain/executioncluster/cluster.go` | Cluster 常量、实体、状态与 identifier 校验边界 |
| `backend/internal/domain/executioncluster/repository.go` | Cluster 读取、默认集群保证、Runner 归属查询的领域接口 |
| `backend/internal/service/executioncluster/service.go` | 组织作用域、默认 Cluster、注册 Token 绑定和可展示状态 |
| `backend/internal/infra/execution_cluster_repository.go` | PostgreSQL/GORM 查询；所有查询均强制 `organization_id` |
| `backend/internal/domain/runner/runner.go` | Runner 增加 `ClusterID` 与 tunnel 状态快照字段 |
| `backend/internal/domain/runner/certificate.go` | PendingAuth、注册 Token 增加不可变 `ClusterID` |
| `backend/internal/service/runner/registration_token.go` | Token 创建与 Runner 注册均使用服务端绑定的 Cluster |
| `backend/internal/service/runner/registration_interactive.go` | 交互注册授权时保留预先选定 Cluster |
| `backend/internal/infra/runner_repo*.go` | Cluster 约束的 Runner 查询和原子注册 |
| `backend/internal/service/agentpod/pod_orchestrator_*.go` | Cluster 内 Runner 选择、Pod Cluster 审计快照、恢复 Worker 的 Cluster 一致性 |
| `backend/internal/api/connect/executioncluster/*.go` | Cluster 列表、注册命令和 Cluster 管理 Connect API |
| `proto/execution_cluster/v1/execution_cluster.proto` | 浏览器/Rust Core 使用的 Cluster 管理协议 |
| `proto/pod/v1/pod.proto` | `CreatePodRequest.execution_cluster_id` |
| `proto/pod/v1/worker_creation.proto` | `PreflightWorkerRequest.execution_cluster_id`，不污染可发布的 WorkerSpec |
| `runner/internal/tunnel/client.go` | 隧道连接、断线/重连状态事件 |
| `runner/internal/client/*tunnel*.go` | 向 Backend 上报当前隧道状态 |
| `backend/cmd/server/eventbus_tunnel.go` | 初始化与周期性隧道重调度 |
| `backend/internal/api/grpc/runner_adapter_message.go` | 消费实时 Tunnel 状态事件并更新 Runner |
| `relay/internal/tunnel/registry.go` | 查询单个 Runner 是否存在当前隧道、记录连接代次 |
| `relay/internal/server/handler_tunnel.go` | 注册/注销事件的可观测性与安全验证 |
| `clients/core/crates/api-client/src/modules/execution_cluster.rs` | Cluster Connect 客户端 |
| `clients/core/crates/services/src/execution_cluster.rs` | Rust 服务层的 SSOT |
| `clients/core/crates/wasm/src/service_execution_cluster.rs` | Web 调用 Rust 服务的 WASM 导出 |
| `clients/web/src/lib/api/connect/executionClusterConnect.ts` | Cluster WASM/Connect 适配器 |
| `clients/web/src/lib/api/facade/executionCluster.ts` | Web Cluster 读取与注册命令 facade |
| `clients/web/src/components/pod/CreatePodForm/ExecutionClusterSelect.tsx` | 创建向导 Cluster 选择器 |
| `clients/web/src/components/pod/CreatePodForm/WorkerRuntimeStep.tsx` | 第一步的模型、镜像、Cluster、部署和资源规格顺序 |
| `clients/web/src/components/pod/hooks/useExecutionClusters.ts` | 创建向导的 Cluster 加载状态 |
| `clients/web/src/components/infra/ExecutionClusterDetail.tsx` | 组织基础设施的 Cluster 状态与注册命令 |
| `clients/web/src/app/(dashboard)/[org]/infra/*` | Cluster/Runner 两层视图 |
| `clients/web/src/components/ide/sidebar/WorkspaceSidebarContent.tsx` | 工作区只保留 Worker 列表；移除 Runner/导入会话混入 |
| `backend/migrations/000206_execution_clusters.*.sql` | Cluster、Runner/Pod/注册凭据的原子数据迁移 |
| `backend/migrations/execution_clusters*_test.go` | SQL 形状、真实 PostgreSQL up/down、约束测试 |

## 验收场景

1. **线上 Cluster 调度**
   - Given `online` 下有多个在线 Runner、`local` 下没有在线 Runner
   - When 用户在创建向导选择 `online`、Codex 和合法模型
   - Then 后端只从 `online` 的 Codex Runner 中选择容量最低者，Pod 的 `cluster_id` 等于 `online`。

2. **本地 Cluster 未连接**
   - Given `local` 已创建但尚未注册本地 Runner
   - When 用户选择 `local` 创建 Worker
   - Then 预检显示“本地集群暂无可用 Runner”，创建 API 返回可辨识的不可调度错误，不会静默改派到线上。

3. **本地 Cluster 受控注册**
   - Given 组织管理员为 `local` 生成一次性注册命令
   - When 本地运行 `runner register --server https://dowork.l8ai.cn --token <token>`
   - Then Token 的 `cluster_id` 被原子写入 Runner，客户端输入不能覆盖它，Runner 建立 gRPC 与隧道后 `local` 显示在线。

4. **Gateway 断线恢复**
   - Given Runner 已在线且隧道已连接
   - When Relay 或 Backend 重启
   - Then Runner 自动重连，Backend 重调度连接命令，管理页面在一个心跳周期内显示 `connected`；若不能连接，显示明确失败原因和最近时间。

5. **安全边界**
   - Given 任意组织成员持有另一个组织的 Cluster ID、Runner ID 或注册 Token ID
   - When 其请求 Cluster、生成注册命令或创建 Worker
   - Then API 返回 404/权限错误，不能读写或调度跨组织资源。
   - Given Preview URL
   - Then 仍只能代理绑定 Pod 的 loopback `preview_target` 和规范化路径，不能访问任意 Cluster IP、域名或端口。

6. **工作区管理**
   - Given 用户打开最新部署的工作区
   - When Worker 已 `completed` 或 `terminated`
   - Then Worker 卡片和右键/更多操作菜单均提供“唤醒”；列表中不再混入 Runner 或导入会话。

### Task 1: 锁定迁移编号与 Cluster 领域契约

**Files:**
- Create: `backend/internal/domain/executioncluster/cluster.go`
- Create: `backend/internal/domain/executioncluster/repository.go`
- Create: `backend/internal/domain/executioncluster/cluster_test.go`
- Create: `backend/migrations/000206_execution_clusters.up.sql`
- Create: `backend/migrations/000206_execution_clusters.down.sql`
- Create: `backend/migrations/execution_clusters_test.go`
- Create: `backend/migrations/execution_clusters_postgres_test.go`

- [x] **Step 1: 确认 `000206` 在写入前仍然为空**

Run: `rg --files backend/migrations | rg '/000206_'`

Expected: no output. If another concurrent change already owns `000206`, stop this task before creating SQL and choose the next unused pair only after recording the collision in the plan progress log; never create duplicate migration sequences.

- [x] **Step 2: 写出失败的 Cluster identifier 与状态测试**

```go
func TestClusterValidateRejectsNonIdentifierSlug(t *testing.T) {
    cluster := Cluster{OrganizationID: 7, Slug: "Local_Cluster", Kind: KindLocal}
    require.Error(t, cluster.Validate())
}

func TestClusterValidateAcceptsLocalCluster(t *testing.T) {
    cluster := Cluster{OrganizationID: 7, Slug: "local", Name: "本地集群", Kind: KindLocal, Status: StatusPending}
    require.NoError(t, cluster.Validate())
}
```

- [x] **Step 3: 运行测试确认失败**

Run: `cd backend && go test ./internal/domain/executioncluster -run TestClusterValidate -count=1`

Expected: FAIL because package and `Cluster` do not yet exist.

- [x] **Step 4: 实现最小 Cluster 领域模型和 repository 接口**

```go
const (
    KindOnline = "online"
    KindLocal = "local"
    StatusReady = "ready"
    StatusPending = "pending"
    StatusOffline = "offline"
)

type Cluster struct {
    ID int64 `gorm:"primaryKey" json:"id"`
    OrganizationID int64 `gorm:"not null;index" json:"organization_id"`
    Slug slugkit.Slug `gorm:"size:100;not null" json:"slug"`
    Name string `gorm:"size:255;not null" json:"name"`
    Kind string `gorm:"size:32;not null" json:"kind"`
    Status string `gorm:"size:32;not null" json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type Repository interface {
    ListByOrganization(ctx context.Context, organizationID int64) ([]*Cluster, error)
    GetByIDAndOrganization(ctx context.Context, id, organizationID int64) (*Cluster, error)
    EnsureDefaults(ctx context.Context, organizationID int64) ([]Cluster, error)
}
```

`Validate` 必须调用 `slugkit.ValidateIdentifier("execution_clusters.slug", string(c.Slug))`，并拒绝未知 `kind`、未知 `status`、非正 `OrganizationID` 与空 `Name`。

- [x] **Step 5: 写出迁移及其真实数据库 RED 测试**

迁移必须一次性完成以下约束：

```sql
CREATE TABLE execution_clusters (
  id BIGSERIAL PRIMARY KEY,
  organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  slug VARCHAR(100) NOT NULL,
  name VARCHAR(255) NOT NULL,
  kind VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT execution_clusters_org_slug_unique UNIQUE (organization_id, slug),
  CONSTRAINT execution_clusters_slug_check
    CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) BETWEEN 2 AND 100),
  CONSTRAINT execution_clusters_kind_check CHECK (kind IN ('online', 'local')),
  CONSTRAINT execution_clusters_status_check CHECK (status IN ('ready', 'pending', 'offline'))
);
```

迁移为每个既有组织创建 exactly `online` 和 `local` 两行；将所有现有 Runner 归属到同组织 `online`；为 `runners.cluster_id`、`pods.cluster_id`、`runner_grpc_registration_tokens.cluster_id`、`runner_pending_auths.cluster_id` 回填后加同组织复合外键。`execution_clusters` 必须拥有 `UNIQUE (id, organization_id)`，而宿主表必须使用 `FOREIGN KEY (cluster_id, organization_id) REFERENCES execution_clusters(id, organization_id)`，从数据库层拒绝跨组织 Cluster。`runner_pending_auths` 在未授权、未选择组织时允许两列同时为空；其余状态由 CHECK 要求 `organization_id` 和 `cluster_id` 同时存在。

迁移还必须在 `runners` 建立可展示的隧道状态字段：

```sql
ALTER TABLE runners
  ADD COLUMN cluster_id BIGINT,
  ADD COLUMN tunnel_state VARCHAR(32) NOT NULL DEFAULT 'disconnected',
  ADD COLUMN tunnel_last_seen_at TIMESTAMPTZ,
  ADD COLUMN tunnel_last_error VARCHAR(255);
```

在完成所有 Cluster ID 回填后，`runners.cluster_id`、`pods.cluster_id` 和 `runner_grpc_registration_tokens.cluster_id` 必须为 `NOT NULL`。`runner_pending_auths.cluster_id` 仅能在未授权且尚未选择组织时为空；所有非空归属均必须通过 `(cluster_id, organization_id)` 复合外键。

真实 PostgreSQL 测试至少插入两个组织、跨组织 Cluster ID 和既有 Runner/Pod/Token/PendingAuth；断言 up 后：

```go
require.Equal(t, int64(2), clusterCount(t, ctx, conn, 10))
require.Equal(t, "online", runnerClusterSlug(t, ctx, conn, 100))
require.Equal(t, "online", podClusterSlug(t, ctx, conn, 200))
require.Error(t, execSQL(ctx, conn, `
    UPDATE runners SET cluster_id = $1 WHERE id = 100
`, otherOrgClusterID))
```

- [x] **Step 6: 运行领域与迁移测试确认通过**

Run: `cd backend && go test ./internal/domain/executioncluster ./migrations -run 'TestClusterValidate|TestMigration000206' -count=1`

Expected: PASS; PostgreSQL 合约测试在设置 `MIGRATIONS_POSTGRES_TEST_DSN` 时执行 up/down 和跨组织拒绝断言。

- [x] **Step 7: 提交领域和迁移边界**

Run:

```bash
git add backend/internal/domain/executioncluster \
  backend/migrations/000206_execution_clusters.up.sql \
  backend/migrations/000206_execution_clusters.down.sql \
  backend/migrations/execution_clusters_test.go \
  backend/migrations/execution_clusters_postgres_test.go
git commit -m "feat(cluster): add execution cluster storage"
```

Expected: 一个只包含 Cluster schema 与领域契约的原子提交。

### Task 2: 将 Runner 注册凭据绑定到 Cluster

**Files:**
- Modify: `backend/internal/domain/runner/runner.go`
- Modify: `backend/internal/domain/runner/certificate.go`
- Modify: `backend/internal/domain/runner/repository.go`
- Modify: `backend/internal/infra/runner_repo.go`
- Modify: `backend/internal/infra/runner_repo_registration.go`
- Create: `backend/internal/infra/runner_repo_registration_test.go`
- Create: `backend/internal/infra/execution_cluster_repository.go`
- Modify: `backend/internal/service/runner/grpc_registration.go`
- Modify: `backend/internal/service/runner/registration_token.go`
- Modify: `backend/internal/service/runner/registration_interactive.go`
- Create: `backend/internal/service/runner/registration_authorize.go`
- Create: `backend/internal/service/runner/registration_cluster.go`
- Modify: `backend/internal/service/runner/grpc_registration_token_test.go`
- Modify: `backend/internal/service/runner/grpc_registration_auth_test.go`
- Create: `backend/internal/service/runner/registration_cluster_test.go`
- Modify: `backend/internal/api/connect/runner/handlers_auth.go`
- Modify: `backend/internal/api/connect/runner/handlers_tokens.go`
- Modify: `backend/internal/api/connect/runner/server.go`
- Modify: `backend/internal/api/rest/v1/runners_grpc_token.go`
- Modify: `backend/internal/api/rest/v1/runners_grpc_types.go`
- Modify: `proto/runner_api/v1/runner.proto`
- Modify: `proto/gen/ts/runner_api/v1/runner_pb.ts`
- Modify: `clients/core/crates/api-client/src/modules/runner.rs`
- Modify: `clients/core/crates/types/src/runner.rs`
- Modify: `clients/web/src/lib/api/facade/runner.ts`

- [x] **Step 1: 写失败测试，证明客户端标签不能决定 Runner Cluster**

```go
func TestRegisterWithTokenUsesPersistedClusterAndPreservesLabels(t *testing.T) {
    token := registrationToken(orgID, localClusterID)
    result, err := service.RegisterWithToken(ctx, &RegisterWithTokenRequest{
        Token: token.Plaintext,
        NodeID: "local-mac",
    }, issuer)
    require.NoError(t, err)
    runner := mustRunner(t, repo, result.RunnerID)
    require.Equal(t, localClusterID, runner.ClusterID)
    require.Contains(t, runner.Tags, "cluster=online")
}
```

- [x] **Step 2: 运行 RED 测试**

Run: `cd backend && go test ./internal/service/runner -run TestRegisterWithTokenUsesPersistedClusterAndPreservesLabels -count=1`

Expected: FAIL because registration tokens and runners do not expose `ClusterID`.

- [x] **Step 3: 最小实现 Cluster 绑定**

在所有三个实体添加明确字段：

```go
ClusterID int64 `gorm:"not null;index" json:"cluster_id"`
```

扩展 `GenerateGRPCRegistrationTokenRequest`：

```go
type GenerateGRPCRegistrationTokenRequest struct {
    Name string
    ClusterID int64
    Labels map[string]string
    SingleUse bool
    MaxUses int
    ExpiresIn int
}
```

服务在创建 Token 前执行：

```go
cluster, err := s.clusterRepo.GetByIDAndOrganization(ctx, req.ClusterID, orgID)
if err != nil || cluster == nil {
    return nil, ErrExecutionClusterNotFound
}
```

`RegisterWithToken` 从已持久化 Token 复制 `ClusterID`；交互授权在已确定组织的 `AuthorizeRunner` 动作中验证并原子写入 `PendingAuth.ClusterID` 与新 Runner。禁止从 CLI 参数、标签或主机名推断 Cluster。保留 Labels，但将其复制进 `Runner.Tags`，以修复现有“Token labels 丢失”的真实数据问题。

- [x] **Step 4: 扩展交互注册**

`RequestAuthURLRequest` 保持无 Cluster：Runner 发起请求时尚未有组织授权上下文。`AuthorizeRunner` 必须携带 `ClusterID`，先以当前组织验证该 Cluster，再用 CAS 将 `organization_id` 与 `cluster_id` 一起写入未授权 PendingAuth；创建 Runner 时只使用该服务端已验证的 Cluster。

- [x] **Step 5: 运行注册回归**

Run:

```bash
cd backend && go test ./internal/service/runner -run 'Test(RegisterWithToken|AuthorizeRunner|GenerateGRPCRegistrationToken)' -count=1
go test ./internal/infra -run 'TestRunner.*Registration' -count=1
go test ./internal/api/connect/runner -run 'Test(MapServiceError|AuthorizeRunner)' -count=1
go build ./cmd/server
cd ../clients/core && cargo test -p agentcloud_types -p agentcloud_api_client -p agentcloud_services
cd ../.. && pnpm run web:typecheck
```

Expected: PASS，包含 token 已过期、使用次数耗尽、跨组织 Cluster、interactive 授权以及 Labels 保留。

- [x] **Step 6: 完成 Cluster 选择入口后提交注册绑定**

质量审查发现，现有“添加 Runner”和交互授权页没有合法 Cluster ID 来源；在 Connect handler 强制 `cluster_id` 后会直接失败。不能用默认 `online` 或前端标签推断代替用户选择。已先实现 Task 4 的 Cluster 列表 API，并将两个入口改为必选 Cluster 选择器和精确 ID 传参；Dashboard 入口经 Rust/WASM service，授权页使用认证前轻量 Connect 读取。

Run:

```bash
git add backend/cmd/server/services_workspace_init.go \
  backend/internal/domain/runner backend/internal/infra/execution_cluster_repository.go \
  backend/internal/infra/runner_repo_registration.go backend/internal/infra/runner_repo_registration_test.go \
  backend/internal/testkit/schema_runner.go backend/internal/api/connect/runner \
  backend/internal/api/rest/v1/runners_grpc_token.go backend/internal/api/rest/v1/runners_grpc_types.go \
  proto/runner_api/v1/runner.proto proto/gen/ts/runner_api/v1/runner_pb.ts \
  clients/core/crates/api-client/src/modules/runner.rs clients/core/crates/types/src/runner.rs \
  clients/web/src/lib/api/facade/runner.ts docs/superpowers/plans/2026-07-12-execution-cluster-gateway-tunnel.md
git commit -m "feat(cluster): bind runner registration to cluster"
```

Expected: 一个包含必选 Cluster 入口的原子提交；不修改 Pod Cluster 选址或 Relay 隧道协议。

### Task 3: Cluster 作用域的 Runner 选择与 Pod 审计快照

**Files:**
- Modify: `backend/internal/domain/agentpod/pod.go`
- Modify: `backend/internal/domain/runner/repository.go`
- Modify: `backend/internal/infra/runner_repo.go`
- Modify: `backend/internal/service/runner/query.go`
- Modify: `backend/internal/service/runner/query_affinity.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_types.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_runner_placement.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_create.go`
- Modify: `backend/internal/service/agentpod/pod_orchestrator_resume.go`
- Modify: `backend/internal/api/connect/pod/create_pod.go`
- Modify: `backend/internal/api/connect/pod/worker_creation.go`
- Modify: `proto/pod/v1/pod.proto`
- Modify: `proto/pod/v1/worker_creation.proto`
- Modify: `clients/core/crates/api-client/src/modules/pod.rs`
- Modify: `clients/core/crates/services/src/pod.rs`
- Modify: `clients/core/crates/wasm/src/service_pod.rs`
- Modify: `clients/web/src/lib/api/connect/podConnect.ts`
- Modify: `clients/web/src/lib/api/connect/podWorkerCreationConnect.ts`
- Modify: `clients/web/src/lib/api/connect/podWorkerCreationTypes.ts`
- Test: `backend/internal/service/agentpod/pod_orchestrator_runner_placement_test.go`
- Test: `backend/internal/api/connect/pod/create_pod_test.go`
- Test: `clients/web/src/lib/api/connect/podConnect.test.ts`

- [ ] **Step 1: 写 Cluster 内调度的失败测试**

```go
func TestResolveRunnerForFreshCreateOnlyUsesRequestedCluster(t *testing.T) {
    req := &OrchestrateCreatePodRequest{
        OrganizationID: 1, UserID: 2, AgentSlug: "codex-cli", ExecutionClusterID: 11,
    }
    selector := newSelector(
        runnerInCluster(21, 11, "codex-cli", 2),
        runnerInCluster(22, 12, "codex-cli", 0),
    )
    err := orchestrator(selector).resolveRunnerForFreshCreate(ctx, req)
    require.NoError(t, err)
    require.Equal(t, int64(21), req.RunnerID)
}

func TestResolveRunnerForFreshCreateDoesNotCrossClusterWhenUnavailable(t *testing.T) {
    req := &OrchestrateCreatePodRequest{
        OrganizationID: 1, UserID: 2, AgentSlug: "codex-cli", ExecutionClusterID: 11,
    }
    err := orchestrator(selectorWithOnlyCluster12Runner()).resolveRunnerForFreshCreate(ctx, req)
    require.ErrorIs(t, err, ErrNoAvailableRunner)
}
```

- [ ] **Step 2: 运行 RED 测试**

Run: `cd backend && go test ./internal/service/agentpod -run TestResolveRunnerForFreshCreateOnlyUsesRequestedCluster -count=1`

Expected: FAIL because `ExecutionClusterID` does not exist and selector methods have no cluster argument.

- [ ] **Step 3: 在协议与 Rust/WASM 边界中新增 Cluster 选择**

在 `CreatePodRequest` 追加新字段，不重用 reserved field：

```proto
optional int64 execution_cluster_id = 21;
```

在 `PreflightWorkerRequest` 追加独立字段：

```proto
optional int64 execution_cluster_id = 3;
```

在浏览器输入、Rust `CreatePodRequest` 映射和 Connect handler 中保持同一字段名；`0` 仅允许旧的 resume 请求，正常创建没有 Cluster ID 必须返回 `invalid_argument`。`WorkerSpecDraft` 保持可移植，不能出现 Cluster ID。

- [ ] **Step 4: 在领域与查询层实现硬边界**

```go
type OrchestrateCreatePodRequest struct {
    OrganizationID int64
    UserID int64
    ExecutionClusterID int64
    RunnerID int64
    // existing fields remain unchanged
}

func (s *Service) SelectAvailableRunnerForAgentInCluster(
    ctx context.Context, orgID, userID, clusterID int64, agentSlug string,
) (*runner.Runner, error)
```

数据库查询必须包含：

```sql
WHERE organization_id = ?
  AND cluster_id = ?
  AND status = 'online'
  AND is_enabled = true
  AND current_pods < max_concurrent_pods
```

显式 Runner 请求必须验证 `runner.cluster_id == execution_cluster_id`。Pod 创建持久化 `Pod.ClusterID = req.ExecutionClusterID`。恢复同一 Pod 时继承源 Pod Cluster，调用方提交不一致 Cluster 必须被拒绝。

`PreflightWorker` 在解析 WorkerSpec 后调用 Cluster 可调度检查器：

```go
if req.Msg.ExecutionClusterId == nil {
    return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("execution_cluster_id is required"))
}
if err := s.clusterAvailability.Check(ctx, tenant.OrganizationID, tenant.UserID,
    req.Msg.GetExecutionClusterId(), draft.WorkerSpec.WorkerTypeSlug); err != nil {
    return connect.NewResponse(preflightIssueResponse(
        "execution-cluster-unavailable", "execution_cluster_id", err.Error(),
    )), nil
}
```

该检查器不得选择或保留 Runner；仅证明目标 Cluster 在当前权限与 Agent 能力下存在至少一个可用 Runner。真正创建时仍重新执行 Cluster 内选址以避免 TOCTOU。

- [ ] **Step 5: 运行服务器端回归**

Run:

```bash
cd backend && go test ./internal/service/agentpod -run 'Test(ResolveRunnerForFreshCreate|CreatePod_.*Cluster|CreatePod_Resume)' -count=1
go test ./internal/service/runner -run 'TestSelect.*Cluster' -count=1
go test ./internal/api/connect/pod -run TestCreatePod -count=1
```

Expected: PASS，包含 Cluster 无可用 Runner、跨 Cluster 显式 Runner、恢复 Worker 以及 Pod 的 `cluster_id` 审计快照。

- [ ] **Step 6: 再生协议代码并运行 Rust/WASM 断言**

Run:

```bash
pnpm proto:gen-go-all
cd clients/core && cargo test -p api-client -p services -p agentcloud-wasm
pnpm run build:wasm
```

Expected: PASS；生成物只包含 `execution_cluster_id` 的兼容字段新增，不改写无关 proto。

- [ ] **Step 7: 提交调度边界**

Run:

```bash
git add proto/pod/v1/pod.proto proto/gen \
  backend/internal/domain/agentpod backend/internal/domain/runner/repository.go \
  backend/internal/infra/runner_repo.go backend/internal/service/runner \
  backend/internal/service/agentpod backend/internal/api/connect/pod \
  clients/core/crates/api-client clients/core/crates/services clients/core/crates/wasm \
  clients/web/src/lib/api/connect/podConnect.ts clients/web/src/lib/api/connect/podWorkerCreationTypes.ts
git commit -m "feat(cluster): schedule workers within execution clusters"
```

Expected: 一个可独立创建线上 Cluster Worker 的原子提交。

### Task 4: 提供 Cluster 管理与本地注册命令 API

> 2026-07-12 进度：Task 2 的浏览器入口不能在没有 Cluster 列表 API 的情况下保持严格绑定；因此先实现本 Task 的服务端 List/CreateRegistrationCommand、Connect 挂载和前端直连读取，再回填 Rust/WASM SSOT 适配与 Cluster 基础设施页面。当前本机浏览器的两个文档开发账户都被现有数据库拒绝，已完成协议/单元/类型验证，但需要有效本地测试会话补做已登录 UI 端到端验证。

**Files:**
- Create: `proto/execution_cluster/v1/execution_cluster.proto`
- Create: `backend/internal/api/connect/executioncluster/server.go`
- Create: `backend/internal/api/connect/executioncluster/handlers.go`
- Create: `backend/internal/api/connect/executioncluster/handlers_test.go`
- Create: `backend/internal/service/executioncluster/service.go`
- Create: `backend/internal/service/executioncluster/service_test.go`
- Create: `backend/internal/infra/execution_cluster_repository.go`
- Create: `clients/core/crates/api-client/src/modules/execution_cluster.rs`
- Create: `clients/core/crates/services/src/execution_cluster.rs`
- Create: `clients/core/crates/wasm/src/service_execution_cluster.rs`
- Create: `clients/web/src/lib/api/connect/executionClusterConnect.ts`
- Create: `clients/web/src/lib/api/facade/executionCluster.ts`
- Modify: `backend/cmd/server/main.go`
- Modify: `clients/core/crates/api-client/src/modules/mod.rs`
- Modify: `clients/core/crates/services/src/lib.rs`
- Modify: `clients/core/crates/wasm/src/lib.rs`

- [x] **Step 1: 写跨组织 API RED 测试**

```go
func TestCreateRegistrationCommandRejectsOtherOrganizationCluster(t *testing.T) {
    resp, err := server.CreateRegistrationCommand(
        tenantRequest(orgOneUser, &executionclusterv1.CreateRegistrationCommandRequest{
            OrgSlug: "org-one", ClusterId: otherOrgClusterID,
        }),
    )
    require.Nil(t, resp)
    require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}
```

- [x] **Step 2: 运行 RED 测试**

Run: `cd backend && go test ./internal/api/connect/executioncluster -run TestCreateRegistrationCommandRejectsOtherOrganizationCluster -count=1`

Expected: FAIL because the service does not exist.

- [x] **Step 3: 定义最小 Connect 协议**

```proto
service ExecutionClusterService {
  rpc ListExecutionClusters(ListExecutionClustersRequest) returns (ListExecutionClustersResponse);
  rpc CreateRegistrationCommand(CreateRegistrationCommandRequest) returns (CreateRegistrationCommandResponse);
}

message ExecutionCluster {
  int64 id = 1;
  string slug = 2;
  string name = 3;
  string kind = 4;
  string status = 5;
  int32 runner_count = 6;
  int32 online_runner_count = 7;
  int32 available_runner_count = 8;
  string tunnel_status = 9;
  optional string tunnel_last_seen_at = 10;
  optional string tunnel_last_error = 11;
}

message CreateRegistrationCommandRequest {
  string org_slug = 1;
  int64 cluster_id = 2;
  string node_name = 3;
}

message CreateRegistrationCommandResponse {
  string command = 1;
  string expires_at = 2;
}
```

`CreateRegistrationCommand` 只允许组织管理员；服务生成的命令必须只包含 `runner register --server <canonical-server-url> --token <one-time-secret>`，不包含 Cluster ID、证书私钥或其他权限信息。

- [x] **Step 4: 实现 Cluster 服务**

`ListExecutionClusters` 先调用 `EnsureDefaults`，再聚合 Cluster 下 Runner 数、可用数与最近 Tunnel 状态。`CreateRegistrationCommand` 将 Cluster ID 放进 `GenerateGRPCRegistrationTokenRequest`，并使用 `MaxUses=1`、`SingleUse=true`、`ExpiresIn=900`。

- [x] **Step 5: 运行 API、服务和代码生成回归**

Run:

```bash
pnpm proto:gen-go-all
cd backend && go test ./internal/domain/executioncluster ./internal/service/executioncluster ./internal/api/connect/executioncluster -count=1
cd clients/core && cargo test -p api-client -p services -p agentcloud-wasm
```

Expected: PASS；跨组织 ID 不能泄露，非管理员不能生成本地注册命令，组织列表总能返回 `online`/`local`。

- [x] **Step 6: 提交 Cluster API 与注册绑定**

Run:

```bash
git add proto/execution_cluster proto/gen backend/internal/domain/executioncluster \
  backend/internal/service/executioncluster backend/internal/infra/execution_cluster_repository.go \
  backend/internal/api/connect/executioncluster backend/cmd/server/main.go \
  clients/core/crates/api-client clients/core/crates/services clients/core/crates/wasm
git commit -m "feat(cluster): expose cluster management and registration"
```

Expected: 与 Task 2 必选 Cluster 入口一起形成一个不修改 Gateway 连接机制的原子提交。注册链接只能使用服务端签发的一次性命令；轻量授权页保留精确十进制 Cluster ID，拒绝不安全数值。已验证 Go 领域/服务/Connect/REST、Rust types/api-client/services/wasm、WASM 产物和 Web typecheck；已登录浏览器路径仍需有效本地测试会话，在 Task 8 补做。

### Task 5: 建立隧道持续状态与重调度闭环

**Files:**
- Modify: `proto/runner/v1/runner.proto`
- Modify: `backend/internal/domain/runner/runner.go`
- Modify: `backend/internal/api/grpc/runner_adapter_message.go`
- Modify: `backend/cmd/server/eventbus_tunnel.go`
- Create: `backend/cmd/server/eventbus_tunnel_reconcile.go`
- Create: `backend/cmd/server/eventbus_tunnel_reconcile_test.go`
- Modify: `runner/internal/tunnel/client.go`
- Create: `runner/internal/tunnel/status.go`
- Modify: `runner/internal/runner/message_handler_tunnel.go`
- Modify: `runner/internal/client/grpc_handler_ready.go`
- Modify: `runner/internal/client/protocol.go`
- Modify: `relay/internal/tunnel/registry.go`
- Modify: `relay/internal/server/handler_tunnel.go`
- Create: `relay/internal/tunnel/registry_status_test.go`
- Create: `relay/internal/server/handler_tunnel_status_test.go`

- [ ] **Step 1: 写 Runner 断线后状态上报的失败测试**

```go
func TestTunnelClientReportsDisconnectedAfterRelayClose(t *testing.T) {
    reporter := &recordingTunnelReporter{}
    client := newTunnelClient(t, relayURL, reporter)
    require.NoError(t, client.Connect())
    closeRelayConnection(t)
    require.Eventually(t, func() bool {
        return reporter.contains(TunnelStatusDisconnected)
    }, time.Second, 10*time.Millisecond)
}
```

- [ ] **Step 2: 运行 RED 测试**

Run: `cd runner && go test ./internal/tunnel -run TestTunnelClientReportsDisconnectedAfterRelayClose -count=1`

Expected: FAIL because `TunnelStatusEvent` and reporter do not exist.

- [ ] **Step 3: 扩展 Runner gRPC 协议并上报状态**

追加一个新 Runner 消息，不改写既有 `TunnelConnectionResultEvent`：

```proto
message TunnelStatusEvent {
  string state = 1; // connecting | connected | disconnected
  int64 observed_at_unix_ms = 2;
  string error_code = 3;
  string error_message = 4;
}
```

`runner/internal/tunnel/client.go` 必须在首次连接、连接断开、每次重连成功时调用 reporter。错误文本必须截断到 256 字符且不得包含 tunnel token、URL query 或证书内容。

- [ ] **Step 4: 后端保存可展示状态，Relay 保持实时权威**

Runner 增加以下持久字段并由迁移 Task 1 创建：

```go
TunnelState string `gorm:"size:32;not null;default:'disconnected'"`
TunnelLastSeenAt *time.Time
TunnelLastError *string `gorm:"size:255"`
```

Backend 只接受当前 gRPC 认证 Runner 的状态消息；禁止携带或信任 client-supplied Runner ID。Relay `Registry` 增加：

```go
func (r *Registry) HasTunnel(runnerID int64) bool
func (r *Registry) Generation(runnerID int64) uint64
```

这两个方法只用于 Relay 的内部实时判断；Backend 的 UI 状态来自受认证 Runner 上报，避免开放未经授权的 Relay 查询接口。

- [ ] **Step 5: 补齐 Backend restart/reconnect 调度**

`setupTunnelConnectCallback` 保留初始化触发，并增加有界周期 reconciler：

```go
func reconcileRunnerTunnels(ctx context.Context, interval time.Duration, query OnlineRunnerQuery, sender runner.RunnerCommandSender) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    for {
        if err := connectOnlineRunners(ctx, query, sender); err != nil {
            slog.Warn("tunnel reconciliation failed", "error", err)
        }
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
        }
    }
}
```

`connectOnlineRunners` 必须限制并发为 8，跳过 `TunnelState=connected` 且 `TunnelLastSeenAt` 未过期的 Runner，对过期/失败 Runner 重新发短期 token。它不能把无法连接的 Runner 标记为可用，也不能无限创建 goroutine。

- [ ] **Step 6: 运行隧道三组件测试**

Run:

```bash
cd runner && go test ./internal/tunnel ./internal/runner ./internal/client -run 'Test.*Tunnel' -count=1
cd relay && go test ./internal/tunnel ./internal/server -run 'Test.*Tunnel|TestPreview' -count=1
cd backend && go test ./cmd/server ./internal/api/grpc -run 'Test.*Tunnel' -count=1
```

Expected: PASS；包含 token 不进入错误日志、Relay reconnect takeover、Backend 重启后的 reconciling、Preview 仍然拒绝非 Pod-bound token。

- [ ] **Step 7: 提交隧道闭环**

Run:

```bash
git add proto/runner/v1/runner.proto proto/gen \
  backend/cmd/server backend/internal/domain/runner backend/internal/api/grpc \
  runner/internal/tunnel runner/internal/runner runner/internal/client \
  relay/internal/tunnel relay/internal/server
git commit -m "feat(gateway): reconcile runner tunnel connectivity"
```

Expected: 一个可独立部署的隧道状态和重连闭环提交。

### Task 6: 在统一四步创建向导中选择 Cluster

**Files:**
- Create: `clients/web/src/components/pod/CreatePodForm/ExecutionClusterSelect.tsx`
- Create: `clients/web/src/components/pod/CreatePodForm/__tests__/ExecutionClusterSelect.test.tsx`
- Modify: `clients/web/src/components/pod/hooks/workerCreateDraft.ts`
- Modify: `clients/web/src/components/pod/hooks/workerCreateValidity.ts`
- Modify: `clients/web/src/components/pod/hooks/useWorkerCreateDraft.ts`
- Create: `clients/web/src/components/pod/hooks/useExecutionClusters.ts`
- Modify: `clients/web/src/components/pod/CreatePodForm/WorkerRuntimeStep.tsx`
- Modify: `clients/web/src/components/pod/CreatePodForm/CreatePodFormFields.tsx`
- Modify: `clients/web/src/lib/api/connect/podConnect.ts`
- Modify: `clients/web/src/lib/api/connect/podWorkerCreationConnect.ts`
- Modify: `clients/web/src/messages/zh/workforce.json`
- Modify: `clients/web/src/messages/en/workforce.json`

- [ ] **Step 1: 写向导字段顺序和离线禁用的失败测试**

```tsx
it("renders model, image, cluster, deployment and resource profile in order", () => {
  render(<WorkerRuntimeStep {...readyProps} />);
  expect(screen.getAllByTestId("runtime-field").map((node) => node.dataset.runtimeField))
    .toEqual(["model", "runtime-image", "execution-cluster", "deployment-mode", "resource-profile"]);
});

it("does not allow a pending local cluster to advance", async () => {
  render(<ExecutionClusterSelect value={2} clusters={[pendingLocal]} onChange={vi.fn()} />);
  expect(screen.getByRole("option", { name: /本地集群/ })).toBeDisabled();
});
```

- [ ] **Step 2: 运行 RED 测试**

Run: `pnpm exec vitest run clients/web/src/components/pod/CreatePodForm/__tests__/ExecutionClusterSelect.test.tsx`

Expected: FAIL because the Cluster selector does not exist.

- [ ] **Step 3: 实现 Step 1 的明确顺序**

Worker 创建第一步渲染顺序固定为：

```tsx
<ModelField ... />
<RuntimeImageField ... />
<ExecutionClusterSelect
  value={executionClusterID}
  clusters={executionClusters}
  onChange={onExecutionClusterChange} />
<DeploymentModeField ... />
<ResourceProfileField ... />
<WorkerTypeField ... />
```

`WorkerType` 保留在同一步但放在运行时字段之后，避免用户在尚未选择运行环境时先填写 Agent 私有表单。Cluster 名称显示 Runner 数、可用数和 tunnel 状态；状态不是 `ready` 或可用 Runner 数为零时禁用并展示具体原因。

- [ ] **Step 4: 校验创建 payload**

`WorkerCreateDraftState` 新增 placement-only 字段和 action：

```ts
interface WorkerCreateDraftState {
  executionClusterID: number;
  draft: WorkerSpecDraft;
}

case "set_execution_cluster":
  return invalidatePreflight(state, {
    ...state,
    executionClusterID: action.clusterID,
  });
```

`WorkerSpecDraft` 不新增 `execution_cluster_id`。`useWorkerCreateDraft` 通过 `useExecutionClusters` 加载 Cluster，`workerCreateValidity` 只有在已选择 `ready` 且有可用 Runner 的 Cluster 时才允许进入下一步。`createPod` 把 `state.executionClusterID` 写入 `CreatePodRequest.executionClusterId`；`preflightWorker` 把同一值写入 `PreflightWorkerRequest.executionClusterId`。

- [ ] **Step 5: 运行前端回归**

Run:

```bash
pnpm exec vitest run clients/web/src/components/pod/CreatePodForm/__tests__ \
  clients/web/src/lib/api/connect/podConnect.test.ts
pnpm run web:typecheck
pnpm run web:lint
```

Expected: PASS；没有 Cluster 时前进按钮不可用，`local` 未连接时不允许提交，`online` 可用时 payload 精确包含 Cluster ID。

- [ ] **Step 6: 提交创建向导**

Run:

```bash
git add clients/web/src/components/pod clients/web/src/lib/api/connect/podConnect.ts \
  clients/web/src/messages/zh/workforce.json clients/web/src/messages/en/workforce.json
git commit -m "feat(worker): select execution cluster during creation"
```

Expected: 一个不改动 WorkerType 私有配置的原子前端提交。

### Task 7: 整理基础设施与工作区管理界面

**Files:**
- Create: `clients/web/src/components/infra/ExecutionClusterDetail.tsx`
- Create: `clients/web/src/components/infra/ExecutionClusterList.tsx`
- Create: `clients/web/src/components/infra/__tests__/ExecutionClusterList.test.tsx`
- Modify: `clients/web/src/app/(dashboard)/[org]/infra/page.tsx`
- Modify: `clients/web/src/app/(dashboard)/[org]/infra/_components/RunnerSection.tsx`
- Modify: `clients/web/src/components/ide/sidebar/WorkspaceSidebarContent.tsx`
- Modify: `clients/web/src/components/ide/sidebar/PodListItem.tsx`
- Modify: `clients/web/src/components/ide/sidebar/SidebarPodActionsMenu.tsx`
- Modify: `clients/web/src/components/ide/sidebar/SidebarPodContextMenu.tsx`
- Modify: `clients/web/src/components/workspace/TerminalPane.tsx`
- Test: `clients/web/src/components/ide/sidebar/__tests__/SidebarPodActionsMenu.test.tsx`
- Test: `clients/web/src/components/ide/sidebar/__tests__/SidebarPodContextMenu.test.tsx`

- [ ] **Step 1: 写列表层级与 Worker 唤醒的失败测试**

```tsx
it("shows two clusters and nests runners under the selected cluster", () => {
  render(<ExecutionClusterList clusters={[online, local]} />);
  expect(screen.getAllByRole("button", { name: /集群/ })).toHaveLength(2);
  expect(screen.queryByText("dev-runner-codex")).not.toBeInTheDocument();
});

it("offers wake from both right-click and visible action menu for completed workers", async () => {
  render(<PodListItem pod={completedPod} {...handlers} />);
  await userEvent.click(screen.getByLabelText("Worker actions"));
  expect(screen.getByText("唤醒 Worker")).toBeVisible();
});
```

- [ ] **Step 2: 运行 RED 测试**

Run: `pnpm exec vitest run clients/web/src/components/infra/__tests__/ExecutionClusterList.test.tsx`

Expected: FAIL because Cluster UI does not exist.

- [ ] **Step 3: 实现基础设施双层视图**

基础设施页默认 tab 改为 `clusters`，`ExecutionClusterList` 只渲染 Cluster 行。点击 Cluster 后显示 `ExecutionClusterDetail`：Cluster 状态、隧道最近状态、Runner 数和“添加此 Cluster Runner”按钮。Runner 明细只能通过选定 Cluster 进入，不能将一个 Runner 伪装为一个 Cluster。

- [ ] **Step 4: 清理工作区列表责任**

`WorkspaceSidebarContent` 删除 `ImportedSessionsSection` 与 `RunnerSection` 的渲染，仅保留 Worker 搜索、筛选、列表和 Worker 操作。导入会话入口迁移到 Worker 创建页的“导入已有会话”操作，不删除底层导入数据。

现有右键菜单与 `MoreHorizontal` 菜单必须保持同一组操作：重命名、分享、移动访问、发布专家、启用/关闭常驻、唤醒、终止、删除。已完成/已终止 Worker 显示“唤醒”；运行中 Worker 显示“停止”和常驻切换。

- [ ] **Step 5: 运行 UI 回归**

Run:

```bash
pnpm exec vitest run clients/web/src/components/infra/__tests__ \
  clients/web/src/components/ide/sidebar/__tests__ \
  clients/web/src/components/workspace/__tests__
pnpm run web:typecheck
```

Expected: PASS；工作区没有 Runner/导入会话列表，所有 Worker 管理入口可被键盘和右键访问。

- [ ] **Step 6: 提交管理 UI**

Run:

```bash
git add clients/web/src/app/'(dashboard)'/'[org]'/infra \
  clients/web/src/components/infra clients/web/src/components/ide/sidebar \
  clients/web/src/components/workspace
git commit -m "feat(workspace): manage workers through cluster-aware views"
```

Expected: 一个集中在页面责任和 Worker 管理动作的原子提交。

### Task 8: 端到端验证、GitOps 发布与数据核验

**Files:**
- Create: `docs/operations/execution-cluster-runbook.md`
- Create: `docs/operations/execution-cluster-acceptance.md`
- Modify: `deploy/dev/runner_runtime_contract_test.sh`
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: 写真实浏览器 E2E 场景**

```text
Given 管理员登录 dev-org
When 打开 基础设施 > 集群
Then 只看见 线上集群 与 本地集群
When 选择 本地集群 并生成注册命令
Then 命令使用 canonical server URL，页面不显示 token 的明文历史
When 使用线上集群创建常驻 Codex Worker
Then Worker 列表显示 running，右键菜单显示停止/常驻操作
When 终止该 Worker
Then TerminalPane 和右键菜单显示唤醒；点击唤醒后打开新的运行 Worker
```

- [ ] **Step 2: 运行本地构建与数据库合同**

Run:

```bash
cd backend && go test ./internal/domain/executioncluster ./internal/service/executioncluster \
  ./internal/service/runner ./internal/service/agentpod ./internal/api/connect/... ./internal/api/grpc
MIGRATIONS_POSTGRES_TEST_DSN='postgres://postgres:postgres@localhost:10002/agentcloud?sslmode=disable' \
  go test ./migrations -run 'TestMigration000206|TestNoDuplicateMigrationSequence' -count=1
go build ./cmd/server
cd ../runner && go test ./internal/tunnel ./internal/runner ./internal/client
cd ../relay && go test ./internal/tunnel ./internal/server
cd .. && pnpm run web:lint && pnpm run web:typecheck && pnpm run web:build
```

Expected: PASS。任何 migration 数据库连接失败都必须报告为环境阻塞，不能跳过真实 SQL 合同测试后宣称完成。

- [ ] **Step 3: 做真实浏览器验证**

用运行中的本地浏览器完成 Task 8 Step 1，并保存以下证据：

```text
docs/operations/evidence/execution-cluster-list.png
docs/operations/evidence/execution-cluster-worker-create.png
docs/operations/evidence/execution-cluster-wake-menu.png
```

检查浏览器 console 无未处理异常，Network 中 Cluster List、Create Registration Command、Create Pod 均为成功响应；验证 `local` 离线时不能创建，验证 `online` 创建不发生跨 Cluster 调度。

- [ ] **Step 4: 按 GitOps 部署，不进行手工热修**

按仓库现有 CI/CD 与 GitOps 配置发布已推送 commit。部署前记录 migration 版本；部署后执行：

```bash
doops -session execution-cluster-release exec --target gw-oilan-node -- \
  kubectl get deploy,pod -A
doops -session execution-cluster-release exec --target gw-oilan-node -- \
  kubectl logs deploy/relay --since=15m
```

验证 Backend、Relay、Runner 健康后，使用远程 `https://dowork.l8ai.cn/dev-org/workspace` 重跑浏览器场景。不得通过 SQL 直接删除/伪造 Runner 或证书来“修复”结果。

- [ ] **Step 5: 提交 runbook 与验收说明**

Run:

```bash
git add docs/operations deploy/dev/runner_runtime_contract_test.sh .github/workflows/ci.yml
git commit -m "docs(cluster): add deployment and acceptance runbook"
git push origin main
git fetch origin main
git branch -r --contains HEAD
```

Expected: `origin/main` 包含完整 SHA；最终交付报告列出 migration 版本、线上/本地 Cluster 状态、测试命令、浏览器证据和回滚方式。

## 审查门禁

- 每个提交前运行 `git diff --check`，并仅暂存本任务负责的明确路径。
- 领域/API/迁移提交需要安全审查：组织作用域、IDOR、Token 泄露、跨 Cluster 调度和 SQL 约束。
- Relay/Runner 提交需要协议审查：重连竞争、Token 刷新、goroutine 泄漏、日志敏感信息和 Preview SSRF 边界。
- 前端提交需要按 `frontend-design-skill` 执行加载、空、错误、禁用、权限与键盘操作的浏览器验证。
- 所有文件保持生产代码不超过 200 行、测试文件不超过 400 行；超过时按责任拆分。

## 回滚

- 应用回滚只允许回退到上一个已部署 Git SHA；不手工修改数据库掩盖发布问题。
- 迁移 down 仅在确认没有新 Cluster 绑定的 Runner、Pod、Token、PendingAuth 后执行；否则先停止新创建并按 runbook 导出审计数据。
- 任何一个 Runner 的证书/注册 Token 不在回滚中重造或复制；需要重新接入时生成新的 Cluster-bound 一次性 Token。
