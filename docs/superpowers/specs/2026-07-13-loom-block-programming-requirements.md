# Loom 积木化 Loop 编程需求规格说明书

- 日期：2026-07-13
- 状态：MVP 已批准
- 产品名：Loom 工作台
- 执行对象：AgentsMesh Goal Loop

## 1. 产品目标

Loom 帮助不编写代码的用户，通过拖拽积木定义一个有明确目标、验收条件、
执行预算和人工升级路径的 AI Loop，并在运行前看到平台将执行的规范化定义。

MVP 证明以下闭环可行：

1. 用户从积木箱拖入一个 Goal Loop。
2. 用户配置 Worker、任务、验收条件、验证命令和执行边界。
3. 用户可双击画布快速插入积木。
4. 用户可创建带参数的声明式自定义宏块。
5. 工作台校验积木结构并生成 canonical Loop JSON。
6. 工作台模拟逐积木执行、高亮和证据输出。

MVP 不连接真实 AI、Runner、数据库或现有 GoalLoop API。

## 2. 用户与核心场景

主要用户是需要重复委派 AI 工作的团队成员。用户不应先理解 Pod、Autopilot、
WorkerSpec 或协议字段，只需回答：

- 由哪个 Worker 执行？
- 要完成什么？
- 什么证据代表完成？
- 最多允许执行多久、多少轮、多少 Token？
- 失败后暂停人工处理还是直接失败？

## 3. 验收场景

### 场景 A：创建有效 Loop

Given 用户已进入空白 Loom 工作台  
When 用户放置完整 Goal Loop 并点击“生成 Loop”  
Then 系统显示无错误的 canonical JSON，包含 Worker、目标、验收标准、验证器、
预算和升级策略。

### 场景 B：阻止不可执行定义

Given Loop 缺少 Worker、任务、验收标准、验证器或执行边界  
When 用户点击“验证”或“运行模拟”  
Then 系统定位对应积木或插槽，并阻止生成与运行。

### 场景 C：创建自定义积木

Given 用户从快速插入菜单选择“创建自定义积木”  
When 用户输入名称和含 `{{parameter}}` 占位符的任务模板  
Then 新积木出现在“我的积木”，每个参数可在积木内和右侧检查器编辑。

### 场景 D：模拟运行

Given 当前 Loop 已通过校验  
When 用户点击“运行模拟”  
Then 工作台按执行顺序高亮积木，记录开始、任务、验证和完成证据。

## 4. MVP 积木目录

| 分类 | 积木 | 参数 | 编译结果 |
| --- | --- | --- | --- |
| 控制 | Goal Loop | 名称 | 根程序 |
| Worker | 使用 Worker | 快照 ID、显示名 | `worker.snapshot_id` |
| 任务 | 执行任务 | 指令文本 | objective 指令 |
| 任务 | 自定义宏块 | 模板参数 | 展开后的 objective 指令 |
| 验收 | 验收条件 | 条件文本 | `acceptance_criteria[]` |
| 验证 | 运行验证命令 | command | `verification.command` |
| 边界 | 执行边界 | 轮数、Token、分钟、无进展、同错 | `limits` |
| 升级 | 失败处理 | pause / fail | `escalation_policy` |

连接规则：

- 一个工作区只能有一个顶层 Goal Loop。
- Goal Loop 必须连接一个 Worker、至少一个任务、至少一个验收条件、一个验证器、
  一个执行边界和一个失败处理。
- 任务与验收条件分别形成有序积木链。
- 所有数值边界必须为正数；最大迭代不得超过 100。
- 未连接的散落积木属于错误，不静默忽略。

## 5. 后续积木规划

### Phase 2：确定性控制流

- `顺序执行`
- `如果验证通过 / 未通过`
- `重复直到验证通过`
- `人工确认`
- `记录证据`
- `成功结束`
- `失败结束`

循环必须有最大次数、时间或 Token 上限。条件只能读取结构化状态、验证结果或
人工决定，不能依赖 AI 自报的自然语言结论。

### Phase 3：平台能力

- `读取仓库`
- `读取 Knowledge`
- `调用 Skill`
- `调用 MCP 工具`
- `创建 Ticket`
- `发送 Channel 消息`
- `调用另一个 Worker`
- `发布或写回外部系统`

有外部副作用的积木必须声明权限、幂等键和人工审批策略。Secret 只能使用引用。

## 6. 自定义积木契约

MVP 自定义积木是声明式任务宏，不是可执行插件：

- 名称用于呈现；内部 ID 由系统生成。
- 模板使用 `{{parameter}}` 声明文本参数。
- 参数名只允许小写字母、数字和连字符，长度 2-100。
- 实例保存参数值，编译时确定性展开。
- 不允许 JavaScript、Shell、网络请求或 Secret 字面量。
- 定义与工作区保存在浏览器 localStorage，仅用于 MVP。

正式产品中，自定义积木定义与发布版本必须组织隔离、不可变并显式升级。

## 7. Canonical Loop JSON

```json
{
  "kind": "goal-loop-program",
  "schema_version": 1,
  "name": "结算页修复",
  "worker": {"snapshot_id": 42, "label": "Codex"},
  "objective": "修复税额计算。\n补充边界测试。",
  "acceptance_criteria": ["完整测试集通过"],
  "verification": {"kind": "command", "command": "pnpm test"},
  "limits": {
    "max_iterations": 10,
    "token_budget": 80000,
    "timeout_minutes": 60,
    "no_progress_limit": 3,
    "same_error_limit": 2
  },
  "escalation_policy": "pause"
}
```

工作区 JSON 只负责恢复 Blockly 视觉状态；canonical JSON 才是预览执行语义。
正式产品必须由后端根据结构化源定义重新编译，不能信任客户端执行计划。

## 8. 工作台信息架构

- 顶部：Loop 名称、保存状态、验证、生成 Loop、运行模拟。
- 左侧：分类积木箱和“我的积木”。
- 中间：Blockly 画布，支持拖拽、缩放、撤销、双击快速插入。
- 右侧：选中积木的参数检查器和结构错误。
- 底部：诊断、生成 JSON、模拟运行证据。

短参数在积木内编辑；长文本和精确数值可在右侧检查器编辑。运行中画布只读，
当前积木高亮，停止后恢复编辑。

## 9. 安全与停止条件

- AI 完成声明不构成成功证据。
- 验证器退出码或受信结果才可确认成功。
- Controller 强制执行轮数、Token、时间、无进展和同错上限。
- 未知积木、无版本积木、类型错误和无退出条件循环必须阻止发布。
- 已发布程序和自定义积木不可原地修改。
- 不提供 fallback、静默跳过或“尽力执行”模式。

## 10. MVP 技术边界

- React 19、TypeScript、Vite、Blockly。
- 独立目录 `prototypes/loom-blockly-mvp/`。
- 独立依赖与锁文件，不修改根工作区依赖。
- 纯函数编译器负责结构校验和 canonical JSON。
- Blockly 只负责编辑与序列化。
- Vitest 验证编译器和自定义宏参数解析。
- 浏览器验证桌面、移动、空白、错误、有效、运行中和自定义积木路径。

## 11. 正式接入边界

正式版本应新增版本化 `LoopProgram` 领域，而不是把 Blockly JSON直接塞入
`GoalLoop`。发布版本编译后再实例化现有 GoalLoop，并复用当前 WorkerSpec
快照、Autopilot、Pod 与外部验证链路。

现有基础：

- `backend/internal/domain/goalloop/goal_loop.go`
- `backend/internal/service/goalloop/goal_loop_execution.go`
- `backend/internal/domain/workflow/workflow.go`
- `proto/goalloop/v1/goalloop.proto`
- `clients/core/crates/services/src/goal_loop_service.rs`

