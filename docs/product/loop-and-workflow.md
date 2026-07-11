# Loop 与 Workflow 产品边界

## 定义

Do Worker 将执行能力、一次性目标任务和可重复自动化任务拆为三个独立概念：

| 概念 | 用途 | 完成或触发方式 |
| --- | --- | --- |
| Worker | 可复用的执行能力与运行环境配置 | 被 Pod、Loop 或 Workflow 引用 |
| Loop | 为一个明确目标进行一次自主执行 | 仅验证命令退出码为 `0` 才完成 |
| Workflow | 可重复运行的自动化任务 | Cron、API 或事件触发每一次运行 |

Loop 的设计采用目标、验收标准、执行和独立验证的闭环。它不把 Agent 的文字声明当作完成证据。该边界与 Codex 的目标及完成条件、Claude Code 的计划和执行分离一致：先明确可验证的交付，再允许自主执行。

Workflow 是重复触发器，不是长时间追求同一个目标的 Loop。每天代码审查、每周依赖扫描、CI 失败回调等场景应使用 Workflow。

## 表单字段

### 创建 Worker

Worker 只定义执行能力，不定义业务目标、循环上限或调度规则。

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| Worker 名称 | 是 | 展示名称 |
| Runner / 集群 | 是 | 提供执行资源 |
| Agent / 镜像 | 是 | 例如 Codex、Claude Code、Gemini |
| 模型资源与凭据 | 按 Agent 要求 | 运行时模型绑定 |
| 仓库与分支 | 否 | 为工作区提供代码上下文 |
| 环境变量 Bundle、Skill、知识库 | 否 | 补充运行环境 |
| 初始提示 | 否 | 单次人工启动 Worker 时使用 |

### 创建目标 Loop

Loop 必须有完成定义，创建后可以保存为草稿或立即启动。它选择已有 Worker 的不可变快照，因此不能在 Loop 表单中重新填写 Runner、Agent、凭据或工作区细节。

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| 名称 | 是 | 目标任务名称 |
| 执行 Worker | 是 | 选择已有 Worker 的不可变配置快照 |
| 目标 | 是 | 这一次必须达成的结果 |
| 验收标准 | 是 | 每条都应可检查 |
| 验证命令 | 是 | Runner 在工作区执行；退出码 `0` 才会完成 |
| 最大迭代次数 | 否 | 默认 `10` |
| Token 预算 | 否 | 单次 Loop 的消耗上限 |
| 总运行时长 | 否 | 默认 `60` 分钟 |
| 无进展阈值、同错阈值 | 否 | 防止重复消耗 |
| 升级策略 | 否 | 暂停等待人工处理，或标记失败 |

Loop 不包含 Cron、并发策略、回调地址、跨运行会话或历史保留。

### 创建 Workflow

Workflow 负责反复执行同一自动化定义。它不需要单次目标验收闭环。

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| 名称、描述、Prompt 模板 | 是 | 每次运行使用的任务定义 |
| Agent、Runner、仓库、分支、模型资源 | 按任务需要 | 每次运行的执行配置 |
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

## 数据与兼容性

历史定时 Loop 已原表改名为 Workflow，保留原有 ID、运行记录、Pod 关联和调度配置。旧 `/loops` 定时任务路由不保留兼容分支；新的 `/loops` 仅表示目标 Loop，`/workflows` 仅表示可重复 Workflow。

## 参考

- [OpenAI Codex Goals](https://developers.openai.com/codex/goals/)
- [OpenAI Codex Scheduled Tasks](https://developers.openai.com/codex/scheduled-tasks/)
- [Claude Code Goals](https://code.claude.com/docs/en/goal)
