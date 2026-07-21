# Loop 与 Workflow 领域拆分设计

## 目标

将当前名为 `Loop` 的定时任务产品拆成两个语义真实的领域：

- `Loop` 是一次目标导向任务，持续迭代直到进入经过验证的终态。
- `Workflow` 是可复用任务定义，可通过手动、API、事件或定时计划触发。

Worker 创建保持独立。Worker 字段描述执行能力，Loop 和 Workflow 字段描述要完成的工作。

## 产品定义

| 产品 | 回答的问题 | 生命周期 |
|---|---|---|
| Worker | 谁来工作，使用什么运行时和能力？ | 可保留的运行配置或实例 |
| Loop | 这个 Worker 必须达成什么单次目标？ | 一个目标生命周期 |
| Workflow | 什么任务需要复用，在什么时候执行？ | 一个版本化定义，多次运行 |
| Workflow Run | 这次执行发生了什么？ | 一条执行记录 |

产品术语 `Loop` 只表示目标导向控制循环。内部轮询、传输循环和定时重复属于实现细节，不得暴露为产品概念。

## 外部产品依据

Codex app-server 的线程目标包含目标内容、生命周期状态、可选 Token 预算、使用量、暂停、恢复、清除和完成行为。Claude Code 将持续推进的 `/goal` 与按间隔重复 Prompt 的 `/loop` 分开。Agent Cloud 采用目标模式的行为，将定时自动化命名为 `Workflow`。

参考：`https://developers.openai.com/codex/app-server`、
`https://code.claude.com/docs/en/goal`、
`https://code.claude.com/docs/en/loop`。

## Worker 边界

Worker 创建负责：

- Worker 类型和运行镜像
- 模型资源
- 交互方式和自动化能力
- 放置策略、计算目标、部署模式和资源规格
- 凭证和类型专属配置
- 仓库、分支、Skills、知识库挂载和环境变量包
- Worker 指令和运行时生命周期
- 展示身份和 Expert 来源

Worker 创建不负责：

- 目标和成功标准
- Cron、事件或 API 触发器
- 迭代、Token 或任务预算
- Workflow 并发和历史保留
- Workflow 回调

`WorkerSpec.workspace.initial_task` 是任务输入，不是 Worker 身份。它需要移出标准 WorkerSpec。“创建并开始”应先创建 Worker，再使用该 Worker 创建 Loop。

## Loop 领域

Loop 是一条持久化的目标执行记录。它没有调度计划，也不再生成嵌套的 Loop Run。

### Loop 字段

- 组织和创建者
- 稳定 slug 和展示名称
- 目标 Worker 或 WorkerSpec 快照
- 目标描述
- 机器可检查的成功标准
- 验证器定义和受保护的验证器引用
- 状态
- 最大迭代次数
- 可选 Token 预算
- 最大运行时长
- 无进展阈值
- 相同错误阈值
- 人工升级策略
- 持久化进度状态引用
- 开始、更新和结束时间；终止原因和证据

### Loop 状态

```text
draft -> active -> completed
                -> blocked
                -> failed
                -> budget_exhausted
                -> cancelled
active <-> paused
```

只有确定性验证器可以将 Loop 标记为 `completed`，Agent 自报完成不能作为证据。迭代、Token、时间和无进展限制由控制器强制执行。不可逆操作必须经过明确的人工检查点。

## Workflow 领域

Workflow V1 是对当前定时 Loop 能力的真实重命名。V1 仍然是单 Worker 任务定义，在多步骤图执行真正实现前，不宣称支持 DAG 工作流。

### Workflow 字段

- 组织、创建者、slug、名称和描述
- 目标 Worker 配置引用
- Prompt 模板和变量
- 仓库和分支执行上下文
- 手动、API、事件或 Cron 触发配置
- 执行模式
- 沙箱和会话持久化策略
- 并发策略和最大并发运行数
- 总超时和空闲超时
- 运行历史保留策略
- 完成回调
- 启用、禁用或归档状态
- 下次和上次运行时间
- 运行统计

每次触发创建一个 `WorkflowRun`。关联 Pod 后，Pod 状态继续作为执行状态的 SSOT。

## 创建体验

### 创建 Worker

配置运行时、能力、工作区、凭证、资源和生命周期。结果是 WorkerSpec 快照和 Worker 实例。页面不出现调度或目标控制字段。

### 创建 Loop

选择 Worker，定义目标和验证器，设置预算和人工升级策略，然后立即启动或保存为草稿。页面不出现 Cron、回调、并发或历史保留字段。

### 创建 Workflow

选择 Worker 配置，定义可复用任务模板和输入，然后配置触发器、并发、超时、持久化、历史保留和回调。页面不得重复 Worker 运行时字段。

## Clean-Cut 迁移

这是有意设计的破坏性迁移，不增加静默 fallback、双写、路由别名或兼容 facade。

现有定时 Loop 栈必须原子迁移：

- `loops` 表改为 `workflows`
- `loop_runs` 表改为 `workflow_runs`
- `backend/internal/domain/loop` 改为 `backend/internal/domain/workflow`
- `backend/internal/service/loop` 改为 `backend/internal/service/workflow`
- `proto.loop.v1.LoopService` 改为 `proto.workflow.v1.WorkflowService`
- 定时任务 REST `/loops` 改为 `/workflows`
- API 权限 `loops:*` 改为 `workflows:*`
- MCP 定时任务工具 `*_loop` 改为 `*_workflow`
- 事件名 `loop_run:*` 改为 `workflow_run:*`
- Web 路由、Store、ViewModel、组件、i18n 和文档全部改为 Workflow

迁移完成后，`/loops` 和 `proto.loop.v1` 只用于新的目标导向 Loop。历史 migration 文件不重写；新增 migration 负责重命名现有表、索引、约束和序列。

## 交付阶段

1. 将现有定时领域跨数据库、后端、Proto、API、Runner MCP、Rust Core、Web、事件、权限和文档整体改为 Workflow。
2. 新增目标导向 Loop 的领域模型、持久化、服务、API 和确定性生命周期测试。
3. 新增独立的 Loop 创建和监控界面。
4. 从新 WorkerSpec 草稿中移除 `initial_task`，实现明确的 Worker 后创建 Loop 的“创建并开始”编排。
5. 执行跨栈回归、浏览器用户路径验证和文档一致性检查。

每个阶段都必须保持自身可构建、可测试。第一阶段必须作为一个完整合并单元交付，因为只完成部分命名迁移会破坏公共契约。

## 验证

- Migration 测试证明现有定时任务和运行历史在表重命名后不丢失，状态和统计不漂移。
- 后端测试覆盖 Workflow CRUD、Cron 抢占、原子触发、并发、取消、Pod 派生状态、回调和权限。
- Proto 生成和 Connect procedure 测试证明不存在残留的定时 `LoopService`。
- MCP E2E 使用 Workflow 工具名，并保持现有执行行为。
- Loop 测试覆盖所有状态转换、仅验证器可完成、预算退出、无进展检测、人工升级和取消。
- WorkerSpec 测试拒绝将任务目标继续持久化为 Worker 身份。
- 浏览器测试覆盖创建 Worker、创建 Loop、创建 Workflow、Workflow 定时执行、Loop 暂停恢复、阻塞升级和终态证据。
- 用户文档和 API 文档必须用一个明确对照表解释 Worker、Loop、Workflow 和 Run。

## 非目标

- Workflow V1 不实现多步骤 DAG 引擎。
- 不保留定时 `/loops` 的兼容别名。
- Workflow 字段不 fallback 到 WorkerSpec 字段。
- Agent 不得修改或弱化验证器。
- 除移除任务归属外，不顺手重构无关 Worker 创建逻辑。
