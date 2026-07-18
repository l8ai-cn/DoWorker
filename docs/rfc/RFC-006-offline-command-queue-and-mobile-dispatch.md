# RFC-006: 离线命令队列与移动端任务下发

| 属性 | 值 |
|------|-----|
| **状态** | In Progress |
| **作者** | Do Worker Team |
| **创建日期** | 2026-07-05 |
| **目标** | 手机经云端向本地 Runner「发了就走」地下发任务：Runner 离线/满载时任务持久排队，上线后自动重放 |
| **参考** | doops.sh 网关隧道机制（`tunnel_hub.go` / `cmd/agent/main.go`） |

---

## 1. 背景

### 1.1 问题

Do Worker 的 Runner 隧道（gRPC+mTLS 出站双向流）已经完整实现了「本地客户端监听、云端下发任务」。但下发是**纯在线语义**：

- `PodCoordinator.CreatePod` → `commandSender.SendCreatePod`，Runner 不在连接表立即返回 `ErrRunnerNotConnected`，Pod 被标 `init_failed`
- Channel `@mention` 转 prompt 时 Runner 离线，只写一条「offline」系统消息，任务丢失
- Runner 满载（`max_concurrent_pods`）时直接不可选，没有排队

桌面场景用户能看到失败并重试；**移动场景用户不会等** —— 手机发任务时家里电脑可能睡眠、断网、正忙，用户期望「发出去，电脑醒了自己跑，跑完通知我」。

### 1.2 对 doops.sh 隧道机制的调研结论

doops-gateway 有三个值得借鉴的机制，与本设计的取舍：

| doops 机制 | 说明 | 本设计取舍 |
|---|---|---|
| `waitForAgent` 重连宽限 | 下发时 agent 恰好掉线，100ms 轮询等 10s 再失败 | **不采用轮询阻塞**。我们有持久队列 + 上线事件驱动重放，闪断场景下 Runner 重连即毫秒级 drain，效果覆盖 grace 且不阻塞 HTTP 请求 |
| `opSlot`/资源锁/`MaxQueuedPerTarget` 忙碌排队 | 目标忙时排队而非失败，队满快速拒绝 | **采用统一队列**。离线与满载走同一条 pending 队列，drain 时统一检查容量 |
| 双 token + action grants + 审计 | user×cluster/instance×action 三元组授权 | 本 RFC 不做（P2，另立 RFC）。mTLS + org 隔离已覆盖当前威胁模型 |

### 1.3 非目标

- 前端（web-user 移动页、PWA、Web Push）—— 本 RFC 只预留事件钩子，前端另行开发
- Runner 级 ACL / 细粒度授权
- `pod_input`（原始终端字节）与 `terminate_pod` 的离线排队 —— 重放过期按键危险；terminate 已有 orphan 恢复语义
- K8s 集群级注册包装

---

## 2. 总体架构

```
┌────────────┐   POST /quick-tasks     ┌─────────────────────────────────────┐
│ 手机客户端  │ ──────────────────────► │ Backend                             │
│ (未来 PWA) │ ◄── 202 {queued} ────── │                                     │
└────────────┘                         │  PodOrchestrator                    │
                                       │      │ queue_if_offline             │
                                       │      ▼                              │
                                       │  PodCoordinator ──在线──► gRPC 下发 │
                                       │      │ 离线/满载                    │
                                       │      ▼                              │
                                       │  PendingCommandQueue (Postgres)     │
                                       │      ▲            │                 │
                                       │  上线事件      drain (FIFO,单飞)    │
                                       │      │            ▼                 │
                                       │  RunnerConnectionManager ──► Runner │
                                       └─────────────────────────────────────┘
                                                            │
                                              eventbus: pod.queued /
                                              pod.queue_dispatched /
                                              pod.queue_expired  (通知钩子)
```

核心不变式：

1. **隧道零改动** —— 复用现有 `ServerMessage` oneof（`create_pod=2`、`send_prompt=6`），队列只是下发前的持久缓冲
2. **Pod 记录先于下发** —— 排队的任务在 DB 中就是一个 `status=queued` 的 Pod，所有既有查询/展示/终止路径自然可见
3. **至少一次投递 + Runner 侧幂等** —— drain 成功后才删队列行；Backend 崩溃可能重放，Runner 侧按 `pod_key` / `command_id` 去重
4. **默认行为不变** —— 只有显式 `queue_if_offline=true`（或 quick-tasks 端点）才进入排队语义，存量调用方零感知

---

## 3. 详细设计

### 3.1 Pod 状态机扩展

`backend/internal/domain/agentpod/pod.go` 新增：

```go
StatusQueued = "queued" // Runner 离线/满载，命令在 pending 队列等待
```

状态转移（新增部分加粗）：

```
              ┌──────────────────────────────────────────────┐
  create ──► **queued** ──dispatch──► initializing ──► running ──► ...
                 │
                 ├─ cancel ──► terminated
                 └─ expire ──► error (error_code=QUEUE_EXPIRED)
```

约束：

- `queued` 属于 `IsActive()`（占用配额展示但**不占** Runner `current_pods` 计数 —— `IncrementPods` 延迟到 dispatch 时刻）
- `queued → initializing` 由 drain 触发；`queued → terminated` 由用户取消；`queued → error` 由过期清扫器触发
- 已存在的 `UpdateByKeyAndActiveStatus` 终止路径对 `queued` Pod 生效（用户可从 Pod 列表直接终止 = 取消排队）

### 3.2 数据模型

迁移 `backend/migrations/000177_pending_runner_commands.up.sql`：

```sql
CREATE TABLE pending_runner_commands (
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    runner_id       BIGINT NOT NULL REFERENCES runners(id) ON DELETE CASCADE,
    pod_key         VARCHAR(100) NOT NULL,
    command_type    VARCHAR(20) NOT NULL CHECK (command_type IN ('create_pod', 'send_prompt')),
    command_id      VARCHAR(64) NOT NULL,      -- 幂等键；create_pod 时等于 pod_key
    payload         BYTEA NOT NULL,            -- proto.Marshal(runnerv1.ServerMessage)
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_pending_cmds_runner_fifo ON pending_runner_commands (runner_id, id);
CREATE INDEX idx_pending_cmds_expiry      ON pending_runner_commands (expires_at);
CREATE UNIQUE INDEX uq_pending_cmds_command ON pending_runner_commands (command_id);
```

设计决策：

- **Postgres 而非 Redis**：需要持久性（Backend 重启不丢任务）、与 Pod 记录同库事务（enqueue 与 `status=queued` 原子写入）、量级极小（每 Runner 上限 20 条）
- **payload 存序列化后的完整 `ServerMessage`**：drain 时直接写入连接的 Send channel，不需要重建命令 —— 这保证排队时刻的配置快照（AgentFile eval 结果）就是执行时的配置，避免 drain 时重新 eval 产生漂移
- **`command_id` 唯一索引**：同一 prompt 重复入队（客户端重试）天然去重
- FIFO 依据 `id`（BIGSERIAL 单调），保证同一 Pod 的 `create_pod` 先于其后续 `send_prompt`

`down.sql`：`DROP TABLE pending_runner_commands;` 并回滚不涉及 Pod 表（`queued` 是纯代码层常量）。

### 3.3 领域层与仓储

```
backend/internal/domain/agentpod/pending_command.go        (~60 行)
    type PendingCommand struct { ID, OrganizationID, RunnerID int64;
        PodKey, CommandType, CommandID string; Payload []byte;
        ExpiresAt, CreatedAt time.Time }
    const CommandTypeCreatePod / CommandTypeSendPrompt
    var ErrQueueFull, ErrDuplicateCommand

backend/internal/domain/agentpod/pending_command_repo.go   (~30 行)
    type PendingCommandRepository interface {
        Enqueue(ctx, cmd *PendingCommand) error              // 违反唯一索引 → ErrDuplicateCommand
        CountByRunner(ctx, runnerID int64) (int, error)
        ListByRunnerFIFO(ctx, runnerID int64, limit int) ([]*PendingCommand, error)
        Delete(ctx, id int64) error
        DeleteByPodKey(ctx, podKey string) (int64, error)    // 取消排队
        ListExpired(ctx, now time.Time, limit int) ([]*PendingCommand, error)
    }

backend/internal/infra/pending_command_repo.go              (~90 行, GORM 实现)
```

### 3.4 队列服务：PendingCommandQueue

新文件 `backend/internal/service/runner/pending_queue.go`（~150 行）：

```go
type PendingCommandQueue struct {
    repo      agentpod.PendingCommandRepository
    podStore  PodStore                       // 复用 PodCoordinator 的接口
    eventBus  EventPublisher                 // 复用 infra/eventbus
    maxPerRunner int                         // 默认 20（对齐 doops MaxQueuedPerTarget 思路）
    defaultTTL   time.Duration               // 默认 30min
    logger    *slog.Logger
}

func (q *PendingCommandQueue) EnqueueCreatePod(ctx, orgID, runnerID int64,
    cmd *runnerv1.CreatePodCommand, ttl time.Duration) error
func (q *PendingCommandQueue) EnqueueSendPrompt(ctx, orgID, runnerID int64,
    podKey, commandID, prompt string, ttl time.Duration) error
func (q *PendingCommandQueue) CancelByPodKey(ctx, podKey string) error
func (q *PendingCommandQueue) QueuePosition(ctx, runnerID int64, podKey string) (int, error)
```

行为规范：

- `Enqueue*` 先 `CountByRunner` 检查 `maxPerRunner`，超限返回 `ErrQueueFull`（API 层映射 429）
- payload 用 `proto.Marshal(&runnerv1.ServerMessage{Payload: &runnerv1.ServerMessage_CreatePod{...}})` 打包
- 成功后发布 eventbus 事件 `pod.queued`（携带 `pod_key`、`runner_id`、`queue_position`）
- TTL 上限钳制 24h，下限 1min

### 3.5 重放器：PendingCommandDrainer

新文件 `backend/internal/service/runner/pending_drain.go`（~180 行）：

```go
type PendingCommandDrainer struct {
    repo          agentpod.PendingCommandRepository
    podStore      PodStore
    runnerRepo    runnerDomain.RunnerRepository
    commandSender RunnerCommandSender          // 复用现有接口
    coordinator   *PodCoordinator              // IncrementPods / MarkInitFailed
    eventBus      EventPublisher
    inflight      sync.Map                     // runnerID → struct{}，drain 单飞
    logger        *slog.Logger
}

func (d *PendingCommandDrainer) DrainRunner(runnerID int64)      // 异步入口
func (d *PendingCommandDrainer) StartExpirySweeper(ctx context.Context) // 每 60s
```

**触发点**（三处，全部是已有回调/事件的追加订阅，不改现有逻辑）：

1. **Runner 上线**：`RunnerConnectionManager.SetInitializedCallback` 目前只标记 available agents；接线处（`backend/cmd/server/services_init.go`）追加调用 `drainer.DrainRunner(runnerID)`。选 initialized 而非 connected，保证 Runner 已完成 agent 能力上报再收任务
2. **容量释放**：`PodCoordinator.handlePodTerminated` 尾部追加 `drainer.DrainRunner(runnerID)`（满载排队的任务在前序 Pod 结束后立即补位）
3. **兜底轮询**：ExpirySweeper 每轮顺带对「有积压且在线」的 Runner 触发一次 drain（防事件丢失）

**DrainRunner 算法**：

```
1. inflight.LoadOrStore(runnerID) 失败 → 已有 drain 在跑，直接返回（单飞）
2. loop:
   a. batch = ListByRunnerFIFO(runnerID, 10)；空 → 结束
   b. 对每条 cmd（严格 FIFO）：
      - 已过期 → 走过期处理（3.6），Delete，continue
      - commandSender.IsConnected(runnerID) == false → 中止本轮（等下次上线事件）
      - command_type == create_pod：
          · runner.CurrentPods >= MaxConcurrentPods → 中止本轮（等容量释放事件）
          · coordinator.IncrementPods → SendCreatePod → 成功后:
              podStore 更新 status: queued → initializing
              Delete(cmd.ID)；发布 pod.queue_dispatched
          · Send 失败 → DecrementPods，中止本轮（命令留在队列）
      - command_type == send_prompt：
          · 目标 Pod 非 active → Delete（Pod 已终止，prompt 作废）
          · SendPrompt → 成功 Delete；失败中止本轮
3. inflight.Delete(runnerID)
4. 若步骤 2 中有新命令入队（enqueue 在 drain 期间发生），由入队方再触发一次 DrainRunner
   （enqueue 时若 IsConnected 则总是补一次 DrainRunner，闭合竞态窗口）
```

**关键并发论证**：

- 单飞锁保证同一 Runner 不会并发 drain → FIFO 顺序性成立
- 「先 Send 成功、后 Delete」+ Backend 崩溃 → 命令可能重放一次 → 由 Runner 侧幂等兜底（3.7）
- 「Delete 成功、Send 其实没到达」不可能：`SendCreatePod` 写入连接 Send channel 即返回 nil，真正的丢失场景是连接断开 —— 此时 Pod 卡在 `initializing`，由**既有**的 init 超时/orphan 检测路径接管（与今天在线下发后断线的行为完全一致，不引入新故障模式）

### 3.6 过期处理

ExpirySweeper 每 60s：

```
rows = ListExpired(now, 100)
for row:
    Delete(row.ID)
    if row.CommandType == create_pod:
        podService.MarkInitFailed(pod_key, "QUEUE_EXPIRED",
            "Task expired after waiting %s for runner to come online")
        发布 pod.queue_expired
    // send_prompt 过期静默丢弃 + debug 日志（prompt 无独立生命周期实体）
```

`QUEUE_EXPIRED` 作为新 error_code 常量，复用 Pod 表已有的 `error_code/error_message` 列（`MarkInitFailed` 现成路径），前端未来据此渲染「任务已过期」。

### 3.7 幂等（Runner 侧）

**proto 变更**（`proto/runner/v1/runner.proto`，向后兼容的字段追加）：

```protobuf
message SendPromptCommand {
  string pod_key = 1;
  string prompt = 2;
  string command_id = 3;   // 新增：幂等键，空值 = 旧客户端行为（不去重）
}
```

`CreatePodCommand` 不加字段 —— `pod_key` 本身就是幂等键。

**Runner 侧改动**（两处，各 ~15 行）：

1. `runner/internal/runner/message_handler.go` `OnCreatePod` 开头：`podStore.Get(pod_key)` 已存在且状态非终态 → 记日志、重发 `pod_created` ACK、直接返回 nil（吸收重放）。当前代码无条件 Put placeholder，重放会重建 Pod —— 这是本设计必须补的洞
2. `OnSendPrompt`：新增每 Pod 容量 32 的 `command_id` 环形去重集（`runner/internal/runner/prompt_dedup.go`，~50 行）；`command_id` 为空跳过去重（兼容在线直发路径）

### 3.8 下发入口改造

**`PodOrchestrator`**（`backend/internal/service/agentpod/`）：

`OrchestrateCreatePodRequest` 新增字段：

```go
QueueIfUnavailable bool          // false = 现状行为（离线即失败）
QueueTTL           time.Duration // 0 = 默认 30min
```

`pod_orchestrator_create.go` 的 dispatch 分支改造（伪码）：

```go
if o.podCoordinator != nil && !req.DeferRunnerDispatch {
    err := o.podCoordinator.CreatePodOrQueue(ctx, req.RunnerID, podCmd, CreatePodOpts{
        Queue: req.QueueIfUnavailable, TTL: req.QueueTTL,
        OrgID: req.OrganizationID,
    })
    switch {
    case err == nil:                    // 已下发，status=initializing（现状）
    case errors.Is(err, ErrPodQueued):  // 已入队，Pod 建成 status=queued
        result.Queued = true
    default:                            // 失败，MarkInitFailed（现状）
    }
}
```

**`PodCoordinator`** 新增 `CreatePodOrQueue`（`pod_coordinator_queue.go`，~70 行）：

```
在线且有容量  → 现有 CreatePod 路径（IncrementPods + Send）
离线 或 满载  → Queue=false: 返回现状错误
              → Queue=true : queue.EnqueueCreatePod → 返回 ErrPodQueued
```

注意：Pod DB 记录的初始 status 由 `podService.CreatePod` 写入 —— 该请求结构追加 `InitialStatus string`（空 = `initializing`），排队路径传 `queued`。

**Channel @mention 路径**（`hook_pod_prompt.go`）：`RoutePrompt` 失败且错误为 `ErrRunnerNotConnected` 时改为 `EnqueueSendPrompt`（TTL 10min），系统消息文案从「offline」改为「offline, message queued for delivery」。该路径 `command_id = fmt.Sprintf("chmsg-%d", message.ID)`（消息 ID 天然幂等）。

### 3.9 REST API 契约

挂载于既有 `backend/internal/api/rest/v1` 路由组（`/api/v1/organizations/:slug/...`，走既有 auth + tenant 中间件）。

#### A. POST `/api/v1/orgs/:slug/quick-tasks`

> 2026-07-16 修订：资源原生控制面取代了早期的运行时参数选择设计。

移动端极简下发入口。调用方先为 `kind: Worker` 执行 Validate 和 Plan，然后
Quick Task 消费该 Plan，内部复用 Worker Apply、不可变 snapshot、持久 launch
和 dispatch outbox。

请求：

```json
{
  "plan_id": "11111111-1111-4111-8111-111111111111"
}
```

PromptRef、inputs、alias、模型、工具、知识、权限、运行镜像、计算目标和 Runner
放置都来自 Plan 固定的 Worker/WorkerTemplate 资源图。接口不再接受 prompt、
agent、Runner、repository、AgentFile 或 queue TTL 覆盖。

响应：

```json
// 202 Accepted；status 是 Apply 后读取到的当前 Pod 状态
{ "pod_key": "pd-x7k2m9", "status": "running" }
{ "pod_key": "pd-x7k2m9", "status": "queued", "queue_position": 3,
  "expires_at": "2026-07-17T08:30:00Z" }
```

错误：

| HTTP | code | 场景 |
|---|---|---|
| 400 | `WORKER_PLAN_INVALID` | Plan ID 或 Worker Plan 无效 |
| 403 | `ACCESS_DENIED` | 当前 actor 无 Apply 权限 |
| 404 | `WORKER_PLAN_NOT_FOUND` | 当前组织不存在该 Plan |
| 409 | `WORKER_PLAN_STATE_CHANGED` | stale、expired、consumed 冲突或基线变化 |
| 422 | `NO_RUNNER_FOR_AGENT` | snapshot 没有可用运行目标 |
| 429 | `QUEUE_FULL` | 已解析 Runner 的队列已满 |
| 503 | `WORKER_APPLY_UNAVAILABLE` | Worker Apply 控制面未就绪 |

#### B. 既有 POST `/pods` 扩展

`CreatePodRequest`（`pod_create.go`）追加可选字段，默认值保持现状行为：

```json
{ "queue_if_offline": false, "queue_ttl_minutes": 30 }
```

响应 201 body 中 Pod 对象 `status` 可能为 `"queued"`（复用既有序列化，无结构变化）。

#### C. GET `/api/v1/organizations/:slug/pods/queued`

```json
// 200
{ "items": [ { "pod_key": "pd-x7k2m9", "runner_id": 42, "agent_slug": "codex-cli",
    "alias": "修登录 bug", "queue_position": 1,
    "created_at": "...", "expires_at": "..." } ] }
```

实现：join `pods(status=queued)` 与 `pending_runner_commands(command_type=create_pod)`。

#### D. DELETE `/api/v1/organizations/:slug/pods/:pod_key/queue`

取消排队：`queue.CancelByPodKey`（删除该 pod_key 全部 pending 行）+ Pod `queued → terminated`。Pod 非 `queued` 状态 → 409 `NOT_QUEUED`。权限：Pod 创建者或 org admin（对齐既有 terminate 权限检查）。

### 3.10 事件（通知钩子，前端后续消费）

eventbus 新增三个事件类型（`backend/internal/infra/eventbus` 既有发布通道）：

| 事件 | 载荷 | 触发点 |
|---|---|---|
| `pod.queued` | pod_key, runner_id, queue_position, expires_at | Enqueue 成功 |
| `pod.queue_dispatched` | pod_key, runner_id, waited_seconds | drain 下发成功 |
| `pod.queue_expired` | pod_key, runner_id | 过期清扫 |

既有 `pod.status_changed` 广播路径（WebSocket → 前端）对 `queued` 状态自动生效，无需改动。未来 Web Push 投递器只需订阅这三个事件 + `pod.status_changed(completed|error)`，本 RFC 不实现。

### 3.11 配置

`backend/internal/config` 追加（env 驱动，全部有默认值）：

| 环境变量 | 默认 | 说明 |
|---|---|---|
| `PENDING_QUEUE_ENABLED` | `true` | 离线积压总开关；false 时离线 `queue_if_offline` 请求直接失败。已连接 Runner 的 owner-before-dispatch 路径仍可短暂写入事务 outbox，提交后立即 drain |
| `PENDING_QUEUE_MAX_PER_RUNNER` | `20` | 单 Runner 积压上限 |
| `PENDING_QUEUE_DEFAULT_TTL` | `30m` | 默认过期时间 |
| `PENDING_QUEUE_SWEEP_INTERVAL` | `60s` | 过期清扫周期 |

---

## 4. 文件变更清单

遵守单文件 <200 行与描述性命名约束：

| 文件 | 动作 | 内容 |
|---|---|---|
| `backend/migrations/000177_pending_runner_commands.{up,down}.sql` | 新增 | 表 + 索引 |
| `backend/internal/domain/agentpod/pod.go` | 修改 | `StatusQueued` 常量、`IsActive` 覆盖 |
| `backend/internal/domain/agentpod/pending_command.go` | 新增 | 实体 + 错误 |
| `backend/internal/domain/agentpod/pending_command_repo.go` | 新增 | 仓储接口 |
| `backend/internal/infra/pending_command_repo.go` | 新增 | GORM 实现 |
| `backend/internal/service/runner/pending_queue.go` | 新增 | Enqueue/Cancel/Position |
| `backend/internal/service/runner/pending_drain.go` | 新增 | DrainRunner + ExpirySweeper |
| `backend/internal/service/runner/pod_coordinator_queue.go` | 新增 | `CreatePodOrQueue` |
| `backend/internal/service/runner/pod_coordinator.go` | 修改 | `handlePodTerminated` 尾部触发 drain |
| `backend/internal/service/agentpod/pod_orchestrator.go` | 修改 | `OrchestrateCreatePodRequest` 追加字段 + `ErrPodQueued` |
| `backend/internal/service/agentpod/pod_orchestrator_create.go` | 修改 | dispatch 分支 |
| `backend/internal/service/agentpod/pod_service.go` | 修改 | `InitialStatus` 支持 |
| `backend/internal/service/channel/hook_pod_prompt.go` | 修改 | 离线转入队 |
| `backend/internal/api/rest/v1/quick_task.go` | 新增 | quick-tasks handler |
| `backend/internal/api/rest/v1/quick_task_types.go` | 新增 | 请求/响应结构 |
| `backend/internal/api/rest/v1/pod_queued.go` | 新增 | 排队列表 + 取消 handler |
| `backend/internal/api/rest/v1/pod_create.go` | 修改 | `queue_if_offline` 字段 |
| `backend/internal/api/rest/v1/router.go` | 修改 | 路由挂载 |
| `backend/cmd/server/services_init.go` | 修改 | 队列/重放器装配 + 回调接线 |
| `backend/internal/config/*` | 修改 | 4 个配置项 |
| `proto/runner/v1/runner.proto` | 修改 | `SendPromptCommand.command_id = 3` |
| `runner/internal/runner/message_handler.go` | 修改 | `OnCreatePod` 幂等 |
| `runner/internal/runner/prompt_dedup.go` | 新增 | prompt command_id 去重环 |
| 各包 `BUILD.bazel` | 再生成 | `bazel run //:gazelle` |

**明确不改**：Relay、`RunnerConnectionManager` 连接管理逻辑、mTLS/注册流程、gRPC adapter 的消息路由、MCP dispatch 表（Agent 间委托的 `create_pod` MCP 方法可后续透传 queue 参数，本期不做）。

---

## 5. 接口测试计划（任务完成条件）

完成标准 = 下列测试全部编写并通过 `bazel test //backend/... //runner/...`。测试遵循仓库既有模式（testify + 既有 mock 基建：`mockCommandSender`、`mockRunnerSelector`、grpc adapter 的 mock stream）。

### 5.1 单元测试 — 队列服务

`backend/internal/service/runner/pending_queue_test.go`

| 用例 | 断言 |
|---|---|
| `TestEnqueueCreatePod_Success` | 行写入、payload 可 Unmarshal 回等价 `CreatePodCommand`、`pod.queued` 事件发布 |
| `TestEnqueue_QueueFull` | 第 21 条返回 `ErrQueueFull`，无副作用 |
| `TestEnqueue_DuplicateCommandID` | 返回 `ErrDuplicateCommand`（唯一索引路径） |
| `TestEnqueue_TTLClamped` | 25h 请求钳制到 24h；0 → 默认 30m |
| `TestCancelByPodKey` | pending 行删除、幂等（二次取消不报错） |
| `TestQueuePosition` | FIFO 中位置正确 |

### 5.2 单元测试 — 重放器

`backend/internal/service/runner/pending_drain_test.go`

| 用例 | 断言 |
|---|---|
| `TestDrain_FIFOOrder` | 3 条命令按入队序到达 mockCommandSender |
| `TestDrain_SingleFlight` | 并发 10 次 `DrainRunner` 只执行一轮（sender 调用计数） |
| `TestDrain_StopsWhenDisconnected` | 第 2 条前断连 → 第 2、3 条保留在队列 |
| `TestDrain_RespectsCapacity` | `current_pods == max` → create_pod 不下发、不删行 |
| `TestDrain_CreatePod_TransitionsStatus` | Pod `queued → initializing`、`IncrementPods` 恰好一次、行删除、`pod.queue_dispatched` 事件 |
| `TestDrain_SendFailure_RollsBackIncrement` | Send 错 → `DecrementPods` 补偿、行保留 |
| `TestDrain_SendPrompt_SkipsInactivePod` | 目标 Pod 已 terminated → 行删除、不下发 |
| `TestDrain_ExpiredRowHandledInline` | drain 遇过期行 → MarkInitFailed(QUEUE_EXPIRED) + 删除 |
| `TestExpirySweeper_MarksInitFailed` | 过期 create_pod → Pod error + `pod.queue_expired` 事件；过期 send_prompt → 静默删除 |

### 5.3 单元测试 — 协调器与编排器

`backend/internal/service/runner/pod_coordinator_queue_test.go`

| 用例 | 断言 |
|---|---|
| `TestCreatePodOrQueue_OnlineDispatchesDirectly` | 在线 → 走现有路径，队列零写入 |
| `TestCreatePodOrQueue_OfflineQueueFalse_FailsAsToday` | 现状回归保护：错误类型与改造前一致 |
| `TestCreatePodOrQueue_OfflineQueueTrue_ReturnsErrPodQueued` | 入队 + `ErrPodQueued` |
| `TestCreatePodOrQueue_BusyQueueTrue_Enqueues` | 满载也入队（离线/满载统一语义） |

`backend/internal/service/agentpod/pod_orchestrator_queue_test.go`

| 用例 | 断言 |
|---|---|
| `TestOrchestrate_QueueIfUnavailable_PodCreatedAsQueued` | DB Pod `status=queued`、result.Queued=true、**不**调用 MarkInitFailed |
| `TestOrchestrate_Default_BehaviorUnchanged` | `QueueIfUnavailable=false` 全链路与现状 byte-for-byte 一致（回归） |

`backend/internal/service/channel/hook_pod_prompt_queue_test.go`

| 用例 | 断言 |
|---|---|
| `TestMentionOfflinePod_EnqueuesPrompt` | `ErrRunnerNotConnected` → EnqueueSendPrompt(command_id=chmsg-{id})、系统消息含 "queued" |
| `TestMentionOfflinePod_DuplicateMessageIdempotent` | 同一消息重投不产生第二条队列行 |

### 5.4 接口测试 — REST API

`backend/internal/api/rest/v1/quick_task_test.go`（httptest + gin，mock Worker Apply）

| 用例 | 断言 |
|---|---|
| `TestQuickTask_InvalidPlanID_400` | 缺失、非规范 UUID 和 nil UUID 均阻断 |
| `TestQuickTask_AppliesWorkerPlanAndReturnsQueueState` | 精确组织/actor scope、plan ID、当前 queued 状态和队列元数据 |
| `TestQuickTask_PlanReplayReturnsCurrentPodStatus` | 幂等重放返回原 Pod 的当前状态 |
| `TestQuickTask_UnconfiguredApplyService_503` | 控制面未 wiring 时明确失败 |
| `TestMapQuickTaskError` | invalid/not-found/stale/consumed/unavailable 使用稳定错误码 |
| `TestQuickTask_Unauthenticated_401` / 跨 org 403 | 既有中间件生效 |

`backend/internal/api/rest/v1/pod_queued_test.go`

| 用例 | 断言 |
|---|---|
| `TestListQueuedPods` | 只含本 org `status=queued`、position 升序 |
| `TestCancelQueuedPod` | 200、Pod → terminated、pending 行删除 |
| `TestCancelNonQueuedPod_409` | running Pod → 409 NOT_QUEUED |
| `TestCancelQueuedPod_ForbiddenForOtherMember` | 非创建者非 admin → 403 |

`pod_create.go` 扩展回归：`TestCreatePod_QueueIfOffline_Returns201Queued` 追加到既有测试文件。

### 5.5 集成测试 — 上线重放全链路

`backend/internal/api/grpc/runner_adapter_drain_test.go`（复用 `runner_adapter_test.go` 的 mock stream 基建）

| 用例 | 断言 |
|---|---|
| `TestRunnerReconnect_DrainsQueuedCreatePod` | 预置 2 条队列行 → 模拟 Runner Connect+Initialize → mock stream 按 FIFO 收到 2 条 `ServerMessage_CreatePod` → 队列清空、Pod 状态推进 |
| `TestRunnerReconnect_MidDrainDisconnect_Requeues` | 第 1 条后断流 → 第 2 条保留，再次连接后送达 |
| `TestPodTerminated_TriggersCapacityDrain` | 满载 + 1 条排队 → 上报 pod_terminated → 排队命令下发 |

### 5.6 单元测试 — Runner 侧幂等

`runner/internal/runner/message_handler_dedup_test.go` / `prompt_dedup_test.go`

| 用例 | 断言 |
|---|---|
| `TestOnCreatePod_DuplicatePodKey_Absorbed` | 二次同 pod_key → 不重建 Pod、返回 nil、重发 ACK |
| `TestOnSendPrompt_DuplicateCommandID_Dropped` | 同 command_id 二次投递 → PTY 只写入一次 |
| `TestOnSendPrompt_EmptyCommandID_NoDedup` | 兼容路径：空 id 两次都执行 |
| `TestPromptDedupRing_EvictsOldest` | 第 33 个 id 挤出第 1 个 |

### 5.7 迁移与构建验证

- `sequence_uniqueness_test.go`（既有）通过 —— 000177 序号唯一
- `bazel run //:gazelle` 后 `bazel build //backend/... //runner/...` 通过
- `bazel run //backend:lint` / `//runner:lint` 零新增告警

---

## 6. 安全与兼容性

| 维度 | 分析 |
|---|---|
| 认证/授权 | 全部新端点挂在既有 org 路由组之后（JWT + tenant 中间件），无新增认证面；取消排队复用 terminate 权限模型 |
| AgentFile 注入 | quick-tasks 的 prompt 进入 `PROMPT "..."` 声明，必须复用 MCP 工具的转义函数（`\` `"` `\n` `\t`），测试 5.4 强制覆盖 |
| payload 信任边界 | pending 行的 payload 由 Backend 自己序列化写入（服务端内部数据），drain 时不重新校验内容，但 Unmarshal 失败时删行 + error 日志（防脏数据卡死队列头） |
| proto 兼容 | `command_id = 3` 为追加字段；旧 Runner 忽略该字段（不去重，行为等同现状在线直发）；旧 Backend 不发该字段，新 Runner 空值跳过去重 —— 双向兼容，无需版本协商 |
| 行为兼容 | 离线排队语义保持 opt-in；`PENDING_QUEUE_ENABLED=false` 禁止离线积压，但不关闭已连接 Runner 的事务 outbox 崩溃恢复 |
| 多副本 Backend | drain 单飞锁是进程内 `sync.Map`；多副本下两副本可能同时 drain 同一 Runner → FIFO 顺序仍由 DB `id` 保证、重复下发由 Runner 幂等兜底，可接受。若未来水平扩容成为常态，将单飞锁升级为 `pg_advisory_xact_lock(runner_id)`（设计预留，本期不做） |

## 7. 风险与缓解

| 风险 | 缓解 |
|---|---|
| 用户遗忘排队任务，电脑数日后唤醒批量执行 | TTL 默认 30 分钟 + 过期显式置 error + `pod.queue_expired` 事件（未来推送）；TTL 上限 24h 硬钳制 |
| 排队时刻与执行时刻配置漂移（如 env bundle 已改） | payload 快照语义：按入队时刻的 eval 结果执行（明确记入文档，是 feature 不是 bug） |
| 队列头部脏数据阻塞 | Unmarshal 失败即删行；过期行 drain 内联处理 |
| drain 与用户取消竞态 | CancelByPodKey 先删行后改状态；drain 的 Delete 是 by-id，行已删则 Send 前 re-check `podStore` 状态为 queued，否则跳过 |
| `send_prompt` 重放导致 Agent 收到重复指令 | command_id 去重环（5.6 覆盖）；环容量 32 × 每 Pod，内存可忽略 |

## 8. 实施顺序

| 阶段 | 内容 | 依赖 |
|---|---|---|
| M1 | migration + domain + infra repo + PendingCommandQueue（含 5.1 测试） | 无 |
| M2 | proto 字段 + Runner 幂等（含 5.6 测试） | 无（与 M1 并行） |
| M3 | Drainer + Coordinator/Orchestrator 改造 + 回调接线（含 5.2、5.3 测试） | M1、M2 |
| M4 | REST 端点 quick-tasks / queued / cancel + channel hook（含 5.4 测试） | M3 |
| M5 | 集成测试（5.5）+ lint/gazelle/迁移验证（5.7） | M4 |

前端（web-user 移动页、PWA、Web Push 投递器）在本 RFC 全部完成后另行立项，消费 3.10 的事件契约。
