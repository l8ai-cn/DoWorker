# Loop 与 Workflow 产品边界

## 定义

Agent Cloud 将可复用配置、一次性运行、目标闭环和可重复自动化拆为不同概念：

| 概念 | 用途 | 完成或触发方式 |
| --- | --- | --- |
| WorkerTemplate | 可复用的执行能力与运行环境配置 | 被 Worker、Expert、Workflow 或 GoalLoop 引用 |
| Worker | 一次性启动声明和执行实例 | Apply 后创建 launch 与 Pod |
| Loop | 为一个明确目标进行一次自主执行 | 仅验证命令退出码为 `0` 才完成 |
| Workflow | 可重复运行的自动化任务 | Cron、API 或事件触发每一次运行 |

Loop 的设计采用目标、验收标准、执行和独立验证的闭环。它不把 Agent 的文字声明当作完成证据。该边界与 Codex 的目标及完成条件、Claude Code 的计划和执行分离一致：先明确可验证的交付，再允许自主执行。

Workflow 是重复触发器，不是长时间追求同一个目标的 Loop。每天代码审查、每周依赖扫描、CI 失败回调等场景应使用 Workflow。

## 表单字段

### 创建 WorkerTemplate

WorkerTemplate 只定义执行能力，不定义业务目标、循环上限或调度规则。

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| 资源名称、显示名称 | 是/否 | identifier 与展示名称分离 |
| Worker 类型、运行镜像 | 是 | 例如 Codex CLI、Claude Code、Gemini CLI |
| ComputeTarget | 是 | Runner 池或集群绑定资源 |
| ModelBinding | 按 Worker 类型 | 模型 API 资源引用，凭据不进入 YAML |
| 仓库与分支 | 否 | 为工作区提供代码上下文 |
| EnvironmentBundle、Skill、知识库 | 否 | 通过版本化资源引用补充运行环境 |
| 生命周期与资源限制 | 是 | 终止策略、超时、profile 或 custom resources |

### 创建 Worker

Worker 引用一个 WorkerTemplate，可附加 Prompt、输入和别名。它是 create-only
资源；需要再次运行时创建新的 Worker identity，不能更新已有 Worker 声明。

### 创建目标 Loop

Loop 必须有完成定义。GoalLoop 资源引用 WorkerTemplate，Plan 固定实际 revision
并编译 WorkerSpec 快照；Apply 只创建 `draft`，不会立即启动。Runner、Agent、
凭据和工作区来自固定快照，不能在 GoalLoop 中复制或覆盖。

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| 名称 | 是 | 目标任务名称 |
| WorkerTemplate | 是 | Plan 时固定 revision 并编译 WorkerSpec 快照 |
| 目标 | 是 | 这一次必须达成的结果 |
| 验收标准 | 是 | 每条都应可检查 |
| 验证命令 | 是 | Runner 在工作区执行；退出码 `0` 才会完成 |
| 最大迭代次数 | 否 | 默认 `10` |
| Token 预算 | 否 | 单次 Loop 的消耗上限 |
| 总运行时长 | 否 | 默认 `60` 分钟 |
| 无进展阈值、同错阈值 | 否 | 防止重复消耗 |
| 升级策略 | 否 | 暂停等待人工处理，或标记失败 |

Loop 不包含 Cron、并发策略、回调地址、跨运行会话或历史保留。
创建完成后，用户必须在 Loop 列表显式启动；Apply 成功不等于已进入 active。

### 创建 Workflow

Workflow 负责反复执行同一自动化定义。它不需要单次目标验收闭环。
定义必须在资源编辑器中通过 `Validate -> Plan -> ApplyWorkflowPlan` 创建或更新；
旧 Connect `CreateWorkflow` 和 Runner MCP `create_workflow` 不再接受 Agent、
Runner、仓库和 Prompt 字段直接创建。

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| 资源名称、显示名称 | 是/否 | Workflow identity 与展示文本 |
| WorkerTemplate | 是 | 固定每次运行的 WorkerSpec 快照 |
| Prompt | 是 | 固定每次运行的任务模板 |
| 输入参数 | 否 | 替换 Prompt 变量 |
| 触发器 | 是 | API 触发始终可用；可额外启用 Cron |
| Cron 表达式 | 启用 Cron 时必填 | 周期性触发规则 |
| 执行模式、沙箱策略、会话保持 | 否 | 定义每次运行方式 |
| 并发策略、最大并发数 | 否 | 控制重叠触发 |
| 超时时间、历史保留数量 | 否 | 控制每次运行和历史记录 |
| 回调地址 | 否 | 运行完成后通知外部系统 |

## 状态与停止条件

Loop 的状态为 `draft`、`active`、`verifying`、`paused`、`completed`、`failed` 或 `cancelled`。

1. `draft` 或 `paused` 的 Loop 可以启动。
2. 目标 Agent 达到自主迭代停止条件后，系统进入 `verifying`。
3. Runner 在该 Loop 的 Pod 工作区执行验证命令。
4. 退出码为 `0` 时，Loop 标记为 `completed`。
5. 非零退出码、超出预算、无进展或重复错误时，按升级策略暂停或失败，并终止该 Pod。

Workflow 的每次运行是独立记录。其成功或失败不改变 Workflow 的重复触发定义。
Trigger 会固定 resource revision、WorkerSpec snapshot、输入参数和解析后的
Prompt，并保存完整 execution manifest。manifest 包含执行模式、Autopilot、
沙箱、会话来源、回调、Ticket、历史保留、timeout 和 idle timeout。启动、完成
回调和超时扫描只读取 run 的固定 manifest；即使 Workflow 已 Apply 新 revision，
旧 run 也不会读取新定义或写入新 revision 的持久运行状态。Enable、Disable、
Trigger 和 Cancel run 是领域动作；Prompt、Cron 或运行策略的定义变更必须
Apply 新 revision。

首次运行从 run 固定的 WorkerSpec snapshot 物化 Pod。持久化 Workflow 有可恢复
的前序 Pod 时，只使用该 `source_pod_key` 作为执行来源，不会同时提交 snapshot
来源；PodOrchestrator 会从 lineage 重新验证组织、WorkerSpec 和 Worker Definition。
本次 run 解析后的 Prompt 作为 invocation metadata 追加到继承的配置层，既不会
修改 Workflow resource revision，也不会生成新的 Worker 身份。

## 数据与兼容性

历史定时 Loop 已原表改名为 Workflow，保留原有 ID、运行记录、Pod 关联和调度
配置。旧 `/loops` 定时任务路由不保留兼容分支；新的 `/loops` 仅表示目标 Loop，
`/workflows` 仅表示可重复 Workflow。

资源迁移不自动改写历史 Expert 或 Workflow。历史 Workflow 仍可读取，但未绑定
resource revision 与 WorkerSpec 快照时不能触发新 run；需要创建 resource-native
替代定义并切换调度。Apply 失败不会改走旧创建路径。

## 参考

- [OpenAI Codex Goals](https://developers.openai.com/codex/goals/)
- [OpenAI Codex Scheduled Tasks](https://developers.openai.com/codex/scheduled-tasks/)
- [Claude Code Goals](https://code.claude.com/docs/en/goal)
