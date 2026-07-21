# Loop 与 Agent Cloud 运行领域接入详细设计

> 2026-07-15 更新：运行环境快照不再进入 LoopScript、AST 或发布版本。
> 启动请求显式携带 `worker_spec_snapshot_id`，后端在创建 GoalLoop 前完成
> 归属、可用性和新鲜度校验。当前权威边界见
> `2026-07-15-loop-runtime-separation-design.md`。

- 日期：2026-07-14
- 状态：已评审，V1 垂直切片已联调
- 依赖：Loop Controller 可靠性与验证设计

## 1. 单一运行所有权

`LoopRun` 是用户可见的统一 run identity、ProgramVersion 引用和审计入口，但每个 run 只能有一个 execution backend：

```text
goal-loop-v1
loop-plan-v2
```

- V1：GoalLoop 是生命周期状态 SSOT；`LoopRun.status` 只做派生 projection，不独立迁移。
- V2：LoopRun/StepRun 是生命周期状态 SSOT；不创建 GoalLoop。
- 同一个 run 不允许同时绑定两个 backend。

## 2. WorkflowRun 边界

Program-backed Workflow 必须保存准确 `program_version_id` 和 `execution_principal_id`。触发时：

1. 创建 WorkflowRun。
2. 创建一个关联 LoopRun。
3. WorkflowRun 的有效状态从 LoopRun 派生。
4. WorkflowRun 不再单独创建 Pod、Autopilot 或 GoalLoop。

Legacy Workflow 保持现有 Pod 派生逻辑。两种模式由显式 execution kind 区分，不读取 draft 或 `latest_published_version_id` 作为运行时隐式默认。

## 3. LoopRun 字段

```text
program_version_id
execution_backend
backend_run_key?
trigger_type
trigger_source
requested_by_id
execution_principal_type
execution_principal_id
credential_binding_policy
effective_budget
status_projection
terminal_reason
```

手动/API 运行把当前调用者解析为 execution principal 并持久化。cron 必须配置组织 service principal；禁止运行时默认继承 Workflow 创建者。

## 4. GoalLoop V1 启动

1. `RunLoopProgramVersion` 创建 backend=`goal-loop-v1` 的 LoopRun。
2. 服务读取 ProgramVersion 中的 `GoalLoopLaunchSpec`。
3. 创建 GoalLoop，并写入 program version id、plan hash 和 loop run id。
4. LoopRun 保存 GoalLoop key 作为 backend run key。
5. 后续状态、cancel、pause、verification 从 GoalLoop 派生到 LoopRun projection。

发布阶段不创建 GoalLoop。V1 使用现有 WorkerSpec、Pod、Runner command
queue 和 Runner verification 链，但 assurance 受 VerifierTrustPolicy
限制。

## 5. LoopPlan V2 启动

1. 创建 backend=`loop-plan-v2` 的 LoopRun。
2. 固定 effective budget、execution principal 和 dependency lock。
3. 创建根 StepRun 或 ready set。
4. Controller 按 claim/outbox/fence 协议执行。
5. 运行完成后写 terminal reason 和最终 evidence manifest。

## 5.1 V1 联调结论

2026-07-14 的 V1 联调已经证明以下真实链路：

1. 浏览器提交 LoopScript。
2. 后端重新解析、校验并编译 GoalLoop launch spec。
3. Worker snapshot 在启动前经过当前 WorkerDefinition 一致性校验。
4. GoalLoop 被创建并进入启动流程。
5. Pod 被真实创建，Runner 启动目标 Worker。

初始垂直切片证据对应 GoalLoop `checkout-fix-2`、Pod
`7-standalone-62c1f8c9` 和 Worker snapshot `23`。控制器修复后的真实
浏览器联调对应 GoalLoop `checkout-fix-7` 和 Pod
`7-standalone-616af48d`：Worker 进入等待后触发 verifier，GoalLoop 未创建
Autopilot controller，取消后 Pod 和循环均进入终态。

GoalLoop V1 不再创建第二个 Autopilot 控制进程。控制循环由后端状态机
确定性驱动：

1. 目标 Worker 进入 `waiting` 后，后端在同一 Pod workspace 运行 verifier。
2. verifier 成功时完成 GoalLoop 并终止 Pod。
3. verifier 非零退出时，后端持久化迭代、验证输出、进展指纹和错误指纹，
   再把验证证据作为下一轮 prompt 发回同一个 Worker。
4. 达到最大迭代、同错、无进展或墙钟时间上限时，按 escalation policy
   进入 paused 或 failed。
5. Runner/验证器自身错误直接升级，不伪装成可由 Worker 修复的任务失败。

验证结果以 `verification_request_id` 做精确 CAS 消费。失败结果和下一轮
`retry_prompt_command_id` 在同一数据库更新中持久化，随后进入现有 Runner
命令队列。命令 ID 在一次迭代内稳定，Runner 按命令 ID 去重，因此重复事件
和恢复扫描不会重复执行成功写入的 prompt；若目标 Worker 尚未就绪或写入
失败，Runner 释放该 command ID，允许恢复扫描用同一 ID 重试。成功或结果
不确定的 prompt ID 持久化在 Runner 的受控 receipt 目录，Runner 重启后
仍按 at-most-once 语义吸收重复命令。只有时间戳
不早于该命令创建时间的
`executing` 事件才能激活下一轮，旧 `waiting`/`executing` 事件不能推进
状态机；事件携带的 agent status 还必须与 Pod 当前持久化状态一致。超时
监控先终止本轮已超预算的 Loop，再恢复其余活跃 Loop 尚未入队的 retry
prompt 和 verifier 请求。Verifier 使用原 request ID 恢复发送；Runner
先原子持久化结果再回传，同 request ID 在进程或 Runner 重启后直接重放
结果。Pod 终止入口同步删除该 Pod 的待发送命令和 receipt。

该实现消除了固定 `claude` 控制命令与单运行时 runner 的契约冲突，也不
需要向 Worker 镜像静默安装第二套 CLI。Worker runtime 自身仍必须通过
catalog、镜像和非交互启动契约校验；控制器修复不构成对无效 Worker
launch contract 的兼容或降级。

## 6. WorkerSpec 与 AgentFile

Worker 节点引用 ProgramVersion dependency lock 中的 WorkerSpec snapshot 和内容 hash。运行时复用现有 Pod/Runner 物化链。

AgentFile 继续是 WorkerSpec 的运行配置编译产物。Loop 不把控制流写入 AgentFile，也不在运行时重新解析最新 Skill/MCP 内容。

启动前必须验证锁定的 WorkerDefinition、Skill、MCP 和 Verifier revision 仍可取得且 hash 相同；不一致直接阻止启动。

## 7. Approval 契约

`loop_approvals` 保存：

```text
run_id
node_id
attempt
fence_token
plan_hash
requested_at
expires_at
decided_by_id
decision
decided_at
consumed_at
```

Approval API 必须携带 run、node、attempt、fence 和 plan hash。决定仅可消费一次；过期、旧 attempt、旧 fence 或不同 plan 的审批全部拒绝。

审批权限由 node policy 决定，不能仅凭能查看 run 即可批准。

## 8. Run API

```text
RunLoopProgramVersion
GetLoopRun
ListLoopRuns
CancelLoopRun
ResumeLoopRun
ApproveLoopStep
RejectLoopStep
```

启动前重新校验 program version、dependency lock、execution principal、权限和预算。失效依赖直接阻止启动，不替换成最新版本。

## 9. UI 与事件

Rust Core 的 `LoopState` 接收 LoopRun/StepRun 事件并保存统一 projection。界面通过 ProgramVersion source map 将 opaque node id 映射到：

- 程序 source range；
- Blockly instance block；
- 自定义积木 template node。

例如事件携带 `node_id=n-01j2-fix-tax` 和 `node_path=[fix-cycle,fix-tax]`；lookup 只使用 node id。

## 10. 错误模型

- compile-target：当前 backend 不支持节点；
- dependency-lock：锁定资源丢失或 hash 不同；
- principal/policy：执行主体或审批权限无效；
- runtime：Worker、Tool、Verifier、预算或基础设施失败；
- ambiguous-effect：外部动作结果不确定，需要人工裁决。

错误必须包含稳定 code、node id、可操作信息和是否可重试。未知节点、未知版本和无法解析的结果全部 fail closed。

## 11. BDD 验收

### Workflow 精确版本

- Given Workflow 绑定 ProgramVersion 3
- When Version 4 发布后 cron 触发
- Then 新 WorkflowRun 仍创建 Version 3 的 LoopRun。

### V1 单一 SSOT

- Given V1 LoopRun 已关联 GoalLoop
- When GoalLoop 进入 verifying
- Then LoopRun 显示派生状态，不能独立写入另一状态。

### V2 不创建 GoalLoop

- Given backend 为 loop-plan-v2
- When 运行启动
- Then 只创建 LoopRun/StepRun，不创建 GoalLoop。

### 审批防重放

- Given attempt 2 正在等待审批
- When 客户端提交 attempt 1 的 approval
- Then 请求被拒绝，不推进节点。

## 12. 分阶段交付

1. Phase 0：冻结语言、AST、compile target 和 conformance corpus。
2. Phase 1：双向编辑和 Rust Core LoopState。
3. Phase 2：ProgramVersion、dependency lock、V1 adapter 和统一 LoopRun projection。
4. Phase 3：V2 Controller、sealed verifier、StepRun 事件和 Approval。
5. Phase 4：Skill/MCP/Worker 调用和 program-backed Workflow。
