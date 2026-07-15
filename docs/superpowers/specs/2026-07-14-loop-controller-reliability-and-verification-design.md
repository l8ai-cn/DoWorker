# Loop Controller 可靠性与验证详细设计

- 日期：2026-07-14
- 状态：待评审
- 依赖：Loop 发布与版本治理设计

## 1. 持久状态

`loop_step_runs` 至少保存：

```text
run_id
node_id
attempt
status
claim_owner
lease_expires_at
fence_token
dispatch_key
idempotency_key
progress_fingerprint
error_fingerprint
input_ref
output_ref
evidence_ref
started_at
finished_at
```

大输出、终端日志和文件产物保存为 artifact/evidence 引用，不直接塞入事件或主表。

## 2. Claim 与 Dispatch

Controller 不能从“ready”直接调用外部动作。固定协议：

1. 数据库事务用 `FOR UPDATE SKIP LOCKED` 选择 ready node。
2. 写入 `claimed`、owner、lease 和单调 fence token。
3. 同一事务写 durable outbox，包含 dispatch key、fence 和 idempotency key。
4. outbox sender 投递到 Runner 或 Action adapter。
5. 接收结果时仅接受当前 fence token。
6. 持久化 step result 后再发布事件并推进依赖节点。

lease 到期允许新 owner 领取并增加 fence。旧 owner 的迟到结果因 fence 不匹配被拒绝。

## 3. 幂等与不确定结果

- Agent、Verifier 和平台 Action 必须接收 `dispatch_key`。
- 可重试副作用必须有 provider 级 idempotency key。
- Runner/adapter 使用 durable inbox 与 `loop_dispatch_receipts` 保存 request hash、fence、状态和 result ref，重复 dispatch 返回原 receipt。
- receipt 至少保留到 run 数据保留期结束；活动 run 的 receipt 禁止清理。
- 外部系统不支持幂等且 dispatch 结果不确定时，Controller 必须 pause 等待人工确认，不能自动重试。
- cancel 只阻止新 dispatch；已发出的不可取消动作仍等待确定结果或人工裁决。

本设计保证 at-least-once 投递和幂等效果，不宣称对任意外部系统实现 exactly-once。provider side effect 与 receipt 无法原子提交且不可查询时，结果必须进入 ambiguous 状态。

## 4. 恢复算法

服务重启后：

1. 从数据库读取非终态 run。
2. 回收已过期 claim。
3. 重发未确认 outbox。
4. 对 `dispatching` 且结果未知的节点查询 Runner/adapter dispatch key。
5. 只有 durable receipt 或 provider reconciliation 确认未执行时才重新 dispatch。
6. 根据持久 fingerprint、预算账本和 step result 继续。

运行状态不依赖 Controller 内存或对话历史。

## 5. VerifierTrustPolicy

可信成功默认使用 `sealed-snapshot`：

1. Agent 完成后冻结 candidate workspace tree/hash。
2. Verifier 在独立 Pod/容器运行，使用不可变 verifier image、PATH 和 command definition。
3. verifier scripts、CI config 和 protected paths 从 ProgramVersion trust bundle 只读挂载。
4. candidate source 只读挂载，构建输出写入隔离的临时 overlay。
5. 输出记录 image、command、tree、trust bundle 和 evidence hash。

Verifier sandbox 默认无 Secret、无网络、非 root、`no-new-privileges`，不挂宿主 socket/目录，并启用 seccomp/AppArmor、只读根文件系统、CPU/内存/PID/磁盘/时间/输出上限。网络、凭证或额外系统调用必须由 trust policy 显式允许。

stdout/stderr、JUnit、coverage 和 artifact 在进入证据库前执行大小限制、控制字符清理、Secret 扫描和内容类型校验。

若任务允许修改测试，修改后的测试只能作为 candidate evidence；trusted success 还必须满足独立 CI、固定隐藏测试、覆盖基线或人工 Approval 中至少一种策略。

现有同 Pod Runner command 只能标记 `basic` assurance。要求 trusted success 的 ProgramVersion 不能用该模式终止成功。

## 6. 停止与 Fingerprint

`progress_fingerprint` 由规范化的相关 repo tree、artifact manifest、Verifier 结果和声明的状态字段计算。`error_fingerprint` 由 node kind、tool、稳定错误码和规范化错误类别计算，排除时间戳和随机 ID。

每个运行必须具备：

- 受信 Verifier、Approval 或类型化状态的成功出口；
- 不可恢复错误或 FailureNode 的失败出口；
- 迭代、Token、墙钟时间预算出口；
- 连续 N 次 progress fingerprint 不变的无进展出口；
- 连续 N 次 error fingerprint 相同的同错出口；
- pause 并写入可操作原因的人工升级。

Token 预算由 Controller usage ledger 根据 Runner usage 事件原子扣减。缺少可靠 usage 事件的 Worker 不能发布带 Token 硬上限的 V2 计划。

## 7. 权限与副作用

每个节点产生 `effects[]`：

```text
read_repository
write_repository
run_command
call_skill
call_mcp_tool
send_channel_message
create_ticket
publish_external
```

发布时检查创建者权限，dispatch 前按运行主体再次检查。Secret 只保存引用；不可逆动作默认要求 ApprovalNode。

## 8. 运行事件

```text
loop_step.claimed
loop_step.started
loop_step.progress
loop_step.evidence
loop_step.completed
loop_step.failed
loop_run.paused
loop_run.completed
loop_run.failed
```

事件携带 run id、program version id、node id、attempt、fence token、timestamp、status 和 evidence/error reference。事件是通知，不是状态 SSOT。

## 9. BDD 验收

### Dispatch 后崩溃

- Given outbox 已投递但 step result 未落库
- When Controller 崩溃并重启
- Then 系统按 dispatch key 读取 durable receipt；若外部结果不可确认则 pause，而不是盲目重试。

### 旧 Owner 迟到

- Given lease 已过期且新 owner 获得更高 fence
- When 旧 owner 返回成功
- Then 结果被拒绝，不覆盖新 attempt。

### 验证污染

- Given Agent 修改 package script 或 verifier config
- When sealed verifier 执行
- Then 使用 trust bundle 中的固定配置，污染结果不能成为 trusted success。

### 无进展恢复

- Given 连续两次 fingerprint 相同后 Controller 重启
- When 第三次仍相同并达到阈值
- Then 根据持久记录触发 no-progress exit。
