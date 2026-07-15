# Loop 积木与代码双向联动技术可行性评审

- 日期：2026-07-14
- 状态：已评审，V1 MVP 已验证
- 评审对象：Loop Blockly MVP、AgentsMesh GoalLoop、Workflow、AgentFile
- 结论：可行，但代码侧必须是受限的 LoopScript DSL，不能承诺任意 TypeScript 与积木无损互转

## 1. 目标与判断标准

本评审验证四个目标：

1. 同一个 Loop 可以用积木或代码声明。
2. 任一视图修改后，另一视图得到等价结构。
3. 发布结果可被后端确定性编译，并交给 Agent 执行。
4. 运行步骤、错误和证据能定位回代码节点与积木。

判断标准是语义可逆、发布可复现、运行可停止、错误不被静默忽略。

## 2. 现状评审

| 能力 | 当前状态 | 结论 |
| --- | --- | --- |
| Blockly 视觉编辑 | 已完成 MVP | 可复用交互和积木定义方式 |
| 积木到规范 JSON | 已完成 | 应升级为积木到统一 AST |
| 自定义积木 | 仅任务文本宏 | 可保留，但正式版本必须版本化 |
| 代码编辑 | 未实现 | 需要专用 DSL、解析器和格式化器 |
| 代码到积木 | 未实现 | Blockly 不提供通用逆向解析能力 |
| 真实 Agent 执行 | GoalLoop 已有扁平执行链 | 可承接 V1 子集，不能承接任意控制流 |
| 步骤级事件 | 缺失 | 必须新增 node_id 级运行事件 |
| 服务端程序版本 | 缺失 | 必须新增不可变 ProgramVersion |

现有证据：

- `prototypes/loop-blockly-mvp/src/domain/compile-loop.ts`
- `prototypes/loop-blockly-mvp/src/blockly/workspace-to-draft.ts`
- `backend/internal/domain/goalloop/goal_loop.go`
- `backend/internal/service/goalloop/goal_loop_execution.go`
- `backend/internal/domain/workflow/workflow.go`
- `clients/core/crates/state/src/app_state.rs`

## 3. 官方能力边界

Blockly 官方提供工作区 JSON 序列化、事件和积木到文本的代码生成器；这些能力适合保存视觉状态和向外生成代码。官方接口没有提供“解析任意代码并恢复积木”的通用能力，因此代码回积木必须建立在 Loop 自己定义的语法和 AST 上。

CodeMirror 6 提供事务化文本状态；Lezer 和 Tree-sitter 支持增量语法树；LSP 的文档版本与诊断模型适合定义 revision、range 和 stale response 处理。这些能力足以支撑双向编辑，但不替代 Loop 的语义模型。

参考：

- [Blockly serialization](https://docs.blockly.com/guides/configure/serialization/)
- [Blockly code generation](https://docs.blockly.com/guides/create-custom-blocks/code-generation/overview/)
- [Blockly events](https://docs.blockly.com/guides/configure/events)
- [CodeMirror reference](https://codemirror.net/docs/ref/)
- [Lezer guide](https://lezer.codemirror.net/docs/guide/)
- [Tree-sitter parsers](https://tree-sitter.github.io/tree-sitter/using-parsers/)
- [LSP 3.17 specification](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/)

## 4. 方案比较

### 方案 A：Blockly 直接生成 TypeScript，再逆向解析 TypeScript

优点：用户熟悉 TypeScript，生态丰富。

问题：

- TypeScript 允许函数、闭包、动态值、反射和副作用，绝大多数语法没有积木对应物。
- AST 等价不代表运行语义等价，无法安全执行用户任意代码。
- 格式、注释和节点身份在双向改写时容易漂移。

结论：拒绝作为正式双向模型。可以在未来提供只读 TypeScript SDK 导出。

### 方案 B：TypeScript Builder DSL

示例：`loop("x", ({ agent, verify }) => { ... })`。

优点：代码感强，可复用 TypeScript 编辑器。

问题：

- 必须限制为非常窄的调用表达式子集。
- 用户一旦加入变量、条件或辅助函数，就失去可逆性。
- 编译器需要区分“合法 TypeScript”和“可投影 TypeScript”，心智成本高。

结论：技术上可行，但错误边界不清晰，不作为首选。

### 方案 C：专用 LoopScript DSL

优点：

- 语法只表达可执行 Loop，每个语法节点都有对应积木。
- 可强制节点标识、预算、验证器和副作用声明。
- 服务端可以确定性解析、检查、格式化和编译。
- 可以明确拒绝不可投影语法，不发生有损转换。

代价：需要维护语法、解析器、格式化器和编辑器支持。

结论：推荐。该投入是双向联动和安全执行的必要成本。

## 5. 架构决策

1. `LoopAST` 是代码和积木共享的语义模型。
2. LoopScript 是可编辑代码表示；Blockly workspace 是视觉表示。
3. Blockly JSON 只保存坐标、折叠和缩放等视图状态，不是执行语义。
4. 发布由后端重新解析、校验和编译，客户端结果不能直接执行。
5. 所有可执行节点必须有稳定 identifier，遵循 `slugkit` 规则。
6. 代码存在语法或类型错误时，发布和积木结构编辑均被阻止。
7. 不支持的语法是硬错误，不做静默删除、近似转换或自动降级。

## 6. 与现有系统的边界

### GoalLoop

现有 GoalLoop 可复用：

- WorkerSpec 不可变快照；
- Pod 创建与 Autopilot；
- 最大迭代、Token、超时、无进展和同错边界；
- Runner 外部命令验证；
- pause/fail 升级策略。

现有 GoalLoop 不能表达多步骤、分支、局部循环、多个 Worker 和步骤级证据。
当前代码能存储并下传 Token 预算，但本次证据未证明跨 Agent 的精确用量账本和 Controller 强制扣减；V2 必须补齐该闭环。
现有命令验证也没有通用的“测试与校验配置未被弱化”证明，正式接入需要 VerifierTrustPolicy。

### Workflow

Workflow 继续负责 cron、手动/API 触发、并发策略和运行历史。它可以引用已发布的 LoopProgramVersion，但不负责解释 LoopScript。

### AgentFile

AgentFile 继续作为 WorkerSpec 的运行配置产物，负责模型、环境、Skill、Knowledge、MCP 和启动方式。LoopScript 负责“何时调用哪个 Worker、如何验证和停止”。禁止把 Loop 控制流塞入 AgentFile。

## 7. 可行性评级

| 子能力 | 评级 | 前置条件 |
| --- | --- | --- |
| 积木生成 LoopScript | 高 | AST 格式化器 |
| LoopScript 恢复积木 | 高 | 受限语法和完整节点映射 |
| 双向实时联动 | 中高 | revision、事务来源、错误态锁定 |
| GoalLoop V1 接入 | 高 | 明确目标和 VerifierTrustPolicy |
| 多步骤控制器 | 中 | 新 LoopRun/StepRun 与解释器 |
| 自定义积木 | 中高 | 不可变定义版本和 AST 宏展开 |
| 多人同时编辑 | 中低 | 不纳入首期，后续再引入 OT/CRDT |

## 8. 首期范围

产品 P0 只实现单用户编辑、单 Loop、已发布 WorkerSpec 快照、顺序任务、命令验证、受限重复、人工确认、预算和失败升级。Phase 2 的 V1 真实运行只覆盖扁平子集；人工确认与结构化控制流在 Phase 3 的 V2 Controller 执行。

以下内容不进入首期：任意 TypeScript、任意 Shell 作为流程语言、动态网络插件、递归宏、并行分支、多人实时协作、运行中修改已发布版本。

## 9. 主要风险

1. 若以 Blockly block id 作为领域标识，重建工作区会破坏运行映射。
2. 若代码错误时仍允许积木继续改，两个视图会产生不可判定冲突。
3. 若客户端编译结果直接运行，用户可绕过权限和验证约束。
4. 若 Agent 能修改验证器或停止条件，会出现奖励投机。
5. 若发布版本可原地修改，历史运行无法复现。

这些风险在详细设计中分别由语义 node id、单写者事务、服务端编译、只读验证器和不可变版本解决。

## 10. V1 验证结果

V1 实现采用 Blockly 13、CodeMirror 6、Go LoopScript 编译器、ConnectRPC
和 Rust Core `LoopState`。验证结果如下：

| 验收项 | 结果 |
| --- | --- |
| 积木编辑后生成规范 LoopScript | 通过 |
| LoopScript 编辑后重建等价积木 | 通过 |
| `@id` 在双向投影中保持稳定 | 通过 |
| 非法代码保留最后有效 AST，并锁定积木运行 | 通过 |
| 后端重新编译后创建并启动真实 GoalLoop | 通过 |
| 过期 Worker snapshot 阻止启动 | 通过 |
| 桌面和 390px 移动端渲染 | 通过 |
| 任意 TypeScript 无损往返 | 不支持，按设计拒绝 |

真实执行已到达 Runner 目标进程和 Autopilot 启动。当前未闭合的最后一段
来自既有 Autopilot 固定 `claude` 控制命令与单运行时 runner 镜像的契约
冲突；这不会改变双向编辑和 V1 启动接入的可行性结论，但会阻止把
“任意 Worker runtime 都能完成自治控制循环”作为当前产品承诺。
