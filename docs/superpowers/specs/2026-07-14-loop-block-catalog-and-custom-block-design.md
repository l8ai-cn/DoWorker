# Loop 积木目录与自定义积木详细设计

- 日期：2026-07-14
- 状态：待评审
- 依赖：`2026-07-14-loop-language-and-ast-design.md`

## 1. P0 积木

| 分类 | 积木 | AST | 必要参数 |
| --- | --- | --- | --- |
| 程序 | Loop | `LoopProgram` | loop id |
| Worker | Worker Snapshot | `WorkerRef` | alias、snapshot id |
| 控制 | 顺序执行 | `SequenceNode` | node id |
| 控制 | 重复直到 | `RepeatNode` | node id、max、condition |
| 任务 | Agent 任务 | `AgentNode` | node id、worker、prompt |
| 验证 | 命令验证 | `CommandVerifierNode` | node id、command、accept |
| 人工 | 人工确认（V2） | `ApprovalNode` | node id、说明、超时 |
| 证据 | 记录证据 | `EvidenceNode` | node id、key、source |
| 结束 | 成功结束 | `SuccessNode` | node id、condition |
| 结束 | 失败结束 | `FailureNode` | node id、reason |
| 边界 | 执行边界 | `LoopLimits` | 五项硬边界 |
| 升级 | 失败处理 | `FailurePolicy` | pause/fail |

P0 的 `Agent 任务` 只向指定 Worker 发送一个明确 prompt。它不能决定自身是否成功；后续 Verifier、Approval 或平台状态负责判定。

## 2. P1 积木

| 分类 | 积木 | 约束 |
| --- | --- | --- |
| 控制 | 条件分支 | 条件只能读取类型化节点结果 |
| 能力 | 调用 Skill | 固定 Skill 版本、结构化输入 |
| 能力 | 调用 MCP Tool | 固定 server/tool、JSON Schema 输入 |
| 协作 | 调用 Worker | 固定 WorkerSpec 快照 |
| 数据 | 设置变量 | 仅 JSON 标量和结构化结果 |
| 数据 | 读取节点结果 | 必须声明来源 node id |

## 3. P2 积木

- 并行执行与 join 策略；
- Ticket、Channel 和外部系统 Action；
- 子 Loop 调用；
- 事件触发与 Workflow 参数；
- 补偿动作。

P2 的外部副作用节点必须声明权限、幂等键、重试策略和人工审批边界。

## 4. 参数编辑

积木内显示短参数：identifier、Worker、状态、次数。长 prompt、命令、JSON 输入和验收说明在右侧检查器编辑。

每次编辑产生一个语义命令：

```text
SetField(node_id, field_path, value)
InsertNode(parent_id, index, node)
MoveNode(node_id, parent_id, index)
DeleteNode(node_id)
ReplaceNode(node_id, node)
```

命令携带 `base_revision`。后端拒绝过期 revision，不自动覆盖新版本。

## 5. 代码映射

| LoopScript | Blockly |
| --- | --- |
| `loop checkout-fix {}` | Loop 根积木 |
| `worker coder = snapshot(42)` | Worker Snapshot 积木 |
| `agent fix-tax(using: coder)` | Agent 任务积木 |
| `verify tests {}` | 命令验证积木 |
| `repeat fix-cycle(...) {}` | C 型重复积木 |
| `on_failure pause` | 失败处理积木 |

Blockly 自身的 block id 只用于当前画布。每个积木的 `data` 保存 AST `node_id`；重建画布时允许生成新的 block id，但语义身份不变。

## 6. 连接规则

- 一个 workspace 只显示当前 AST 的一个根 Loop。
- 所有 executable/control 积木必须连接到根结构。
- 输入插槽按 AST 类型检查，不允许任意积木连接。
- 删除仍被条件或数据引用的节点时，先显示引用列表并阻止直接删除。
- C 型循环体内节点按 AST 顺序排列；拖动只改变显式顺序。
- 未知积木、无 node id 积木和孤立积木都产生阻塞诊断。

## 7. 双击创建积木

画布双击打开上下文插入菜单：

1. 根据当前插槽类型筛选可用积木。
2. 选择内置积木时生成唯一 identifier 和 AST node。
3. 选择“创建自定义积木”时进入定义向导。
4. 新节点通过 `InsertNode` 提交；成功后才进入正式 workspace 状态。

向导要求填写名称、slug、参数、展开模板、分类和预期副作用。创建和发布自定义积木是两个动作；未发布版本只能用于当前草稿预览，不能发布 LoopProgram。

## 8. 自定义积木

自定义积木是不可变、版本化的 AST 宏，不是 JavaScript 插件。定义源使用独立 LoopBlock DSL；LoopProgram 只允许 `use block` 和类型化调用。

```loop
block fix-module v1(worker: worker-ref, module: text, test: command) {
  @id(n-template-work)
  agent work(using: worker) {
    prompt "修复 {{module}} 并补充测试。"
  }
  @id(n-template-check)
  verify check {
    command test
    accept "{{module}} 测试通过"
  }
}
```

实例：

```loop
use block fix-module@1
@id(n-billing-fix)
fix-module billing-fix(worker: coder, module: "billing", test: "pnpm test")
```

展开节点的显示 path 为 `billing-fix/work`、`billing-fix/check`；真实 node id 由 program version、实例 node id 和 template node id 确定性生成，并满足 slugkit。

## 9. 自定义积木版本

每个版本包含：

- slug、version、owner organization；
- 参数 JSON Schema；
- AST template；
- 展开后允许的 effects；
- UI 分类、图标、颜色 token；
- content hash、发布时间和废弃状态。

禁止递归宏、运行时下载代码、Secret 字面量和未声明副作用。发布后的定义不可原地修改；升级由 Loop 草稿显式选择新版本。

ProgramVersion 保存 expansion source map：expanded node id -> instance node id、template node id、调用处 source range 和 Blockly instance block id。运行高亮宏调用处，并可下钻模板节点。

## 10. 积木投影器

投影器输入 typed AST，输出 Blockly block tree 和 node index：

```text
node_id -> block_id
block_id -> node_id
node_id -> source_range
```

更新时按 node id 比较：

- 相同 kind：更新字段、连接和顺序；
- kind 改变：替换该节点积木；
- 新增或删除：局部创建或移除；
- 根结构变化：重建 workspace，但恢复匹配 node id 的视图状态。

投影器不解释语义，也不编译执行计划。

## 11. UI 状态

必须呈现：

- 空白、解析中、有效、代码错误、冲突、发布中；
- 当前选中节点、运行中节点和历史证据节点；
- 自定义积木草稿、已发布、已废弃、缺失版本；
- 只读代码模式、只读积木模式和运行版本查看模式。

错误点击后同时定位积木和代码 range。运行历史查看使用发布版本，不切换到当前 draft 的节点。

## 12. 验收

1. 任一内置积木都能生成规范 LoopScript，并从该代码恢复同构 AST。
2. 积木移动只改变 AST 顺序，不改变未移动节点的 node id。
3. 自定义积木展开结果稳定，参数缺失时阻止发布。
4. 未知、孤立和类型错误积木不会被忽略。
5. 重建 workspace 后 block id 可变化，但 node id、代码和运行定位保持一致。
