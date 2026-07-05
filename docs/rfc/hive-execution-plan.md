# Hive 全量改造执行计划

> 状态：S0–S7 验收通过 | S8 Omnigent parity 代码完成、验收待跑 | 更新：2026-07-05

## 目标定义

在 **不引入 Omnigent Python 控制面** 的前提下，完成 Hive 五工作流 + web-user 对接，使：

1. `clients/web-user` 通过 AgentsMesh `/v1/*` compat 层可启动、建会话、发消息
2. `clients/web` 管理面由 AgentFile `CAPABILITY` 驱动，消除 slug 硬编码
3. Backend/Runner 具备 Policy、Usage、Resume 的 **可扩展基础**（表 + proto + 评估桩）

## 阶段与验收

| 阶段 | 工作流 | 交付物 | 验收 |
|------|--------|--------|------|
| **P0** | A 收尾 + F MVP | capabilities 全 builtin；session CRUD；events；vite 代理 | web-user 能登录、选 agent、建 session |
| **P1** | F 体验 + B 最小 | SSE stream；PATCH runner；permission_policies + runner eval 桩 | 侧边栏有 session；发消息到 Pod |
| **P2** | C + D 基础 | model_prices；external_session_id；PodUsageEvent proto | session 显示 status；表结构就绪 |
| **P3** | B/C/D 完整 + E | policy 热更新；实时 usage USD；resume/fork；CreatePod 简易模式 | 阶段 2–3 hive-platform-plan 验收 |
| **P4** | F P1/P2 | fork、files、harnesses、WS updates | web-user 体验对齐 |

## P0 任务清单（本迭代执行）

### A — Capability
- [x] A3 proto + Connect enrichment
- [x] A4 do-agent migration
- [x] A4b migration 000163：claude/codex/gemini/opencode/cursor/loopal
- [x] A6 `agent-capability-axes.ts` + `agentSupports()`

### F — Session Compat
- [x] F1 agent_sessions migration（修正 conv_* id）
- [x] F1b agentsession domain + service
- [x] F2 POST/GET/LIST/PATCH sessions → Pod
- [x] F4 POST events（message → SendPrompt）
- [x] F3 SSE stream（status + keepalive）
- [x] web-user vite 默认代理 AgentsMesh

### B/C/D — 基础
- [x] 000163 permission_policies 表
- [x] 000163 model_prices 表
- [x] 000163 pods.external_session_id
- [x] runner policy eval 桩（fail-safe ASK）

## P1 任务清单（进行中）

### F — Session 体验
- [x] F5 conversation_items 持久化 + SSE 文本流（content_delta / message_done）
- [x] F5b permission_request → `response.elicitation_request` SSE
- [x] F5c `POST /v1/sessions/{id}/elicitations/{eid}/resolve` → gRPC `acp_relay`
- [x] F5d list agents 快速 capabilities 扫描（`capability.ScanDeclarations`）
- [x] F5e Pod 状态变更 → `session.status` SSE；ACP 无 turn 时自动 `StartTurn`
- [x] F5f `GET /health?session_ids=` session liveness
- [x] F6 Terminal/Relay WS 桥接

### B — Policy 最小
- [x] runner policy eval 桩（fail-safe ASK）
- [x] B2 policy 快照下发 + Runner ACP handler 集成

## P2+ 待办
- [x] C2 PodUsageEvent 实时上报 + USD 聚合
- [x] D2 resume_external_session + fork API
- [x] E CreatePodForm simple/advanced 分层

## P3 冲刺任务清单（下一步执行）

> 排序原则：先修 blocker（S0），再验证浏览器闭环（S1），再补 compat 缺口（S2），
> 最后补 Hive 机制（S3）。S0 不通过，S1–S3 全部无法验收。

### S0 — 消息往返 E2E（blocker，最高优先）

当前故障：`POST /v1/sessions/{id}/events` 202 后，assistant 回复未落
`conversation_items`（smoke 第 4 步失败）。链路：SendPrompt → Runner →
e2e-echo ACP → `AcpSessionEvent` → `ForwardAcpSession` → EventBridge。

- [x] S0.1 诊断 docker runner gRPC `handshake EOF`：根因是 traefik passthrough 在 Mac Docker 下不稳定；mTLS 直连 host 正常
- [x] S0.2 `docker-compose.runners.yml` 改为 `GRPC_ENDPOINT: host.docker.internal:10016`；替换 legacy `agentsmesh-main-runner-1` 为 `runner-e2e-echo`
- [x] S0.3–S0.4 compat 层注入 `MODE acp`（`compat_agentfile_layer.go`）；runner `acpSendPromptWhenReady` 修复 create/send_prompt 竞态
- [x] S0.5 `node output/api-integration-smoke.mjs` **五步全绿**（2026-07-05 验证）

验收：assistant item 落库、SSE 收到 `response.output_text.delta` +
`response.completed`、`GET /items` 返回两条消息。

### S1 — web-user 浏览器闭环

- [x] S1.1 浏览器 E2E：登录 → 选 agent → 建 session → 发消息 → 气泡渲染
      （`node output/hive-browser-integration.mjs` 全绿，2026-07-05）
      修复项：CORS 放行 `:5173`；`GET /v1/hosts/{id}/filesystem` stub；
      `POST /events` 接受 `input_text` 块类型
- [x] S1.2 `session.status` + 侧边栏：`GET /v1/sessions` 返回 Omnigent
      `conversation` wire（`updated_at`/`runner_online`）；`WS /v1/sessions/updates`
- [x] S1.3 elicitation 闭环：`scenario=permission_request_edit` →
      `POST .../elicitations/{id}/resolve` → assistant 继续（`hive-s1-smoke.mjs`）
- [x] S1.4 Terminal attach WS：`pty_only` session → relay 字节（`hive-s1-smoke.mjs`）

验收：无 console error；一次完整对话 + 一次审批 + 一次终端 attach。

### S2 — compat API 缺口（按 web-user 404 影响排序）

- [x] S2.1 `GET /v1/harnesses`：builtin agents → `{id, label}`（`hive-s2-smoke.mjs`）
- [x] S2.2 `/v1/policies` CRUD + `/v1/policy-registry` → `permission_policies`
      表（`acp_tool_rule` handler 投影）
- [x] S2.3 `PUT .../read-state` + `GET .../permissions` + `GET .../owner`：
      Postgres `session_read_states`；permissions 空数组（单用户）；`permission_level=4`
- [x] S2.4 `WS /v1/sessions/updates`：S1.2 已实现
- [x] S2.5 `switch-agent` / `POST .../directories` / sandbox filesystem：501 stub

验收：web-user 控制台无 404 噪音；策略页可增删 org 规则。

### S3 — Hive 机制补完（对应 platform-plan 阶段 2–3）

- [x] S3.1 A5 ACP runtime 校准：`CalibrateDeclaredCapabilities` +
      `LogDeclaredRuntimeMismatches` WARN
- [x] S3.2 B 热更新：policy CRUD → `SendUpdatePodPolicyRules` 推送在跑 Pod
- [x] S3.3 C 展示：turn 完成 → `PodUsageEvent` → `total_cost_usd`（session
      头部经 SSE `session.usage` 实时更新；web-user `AgentInfo` 已接）
- [x] S3.4 D fork：`resume_external_session` 注入 CreatePod；fork 复制 items +
      续对话

验收：`bazel test //deploy/dev:hive_smoke --test_tag_filters=hive` 或
`bash deploy/dev/hive_smoke.sh` 全绿（2026-07-05）。

### S4 — 收尾

- [x] S4.1 统一 smoke：`deploy/dev/hive_smoke.sh` 串联 S0–S4；
      Bazel `//deploy/dev:hive_smoke`（tag `hive` + `manual`）
- [x] S4.2 CI：`bazel.yml` smoke job 跑 `runner_runtime_contract_test`
- [x] S4.3 `reset_runners` 改为 docker cp 热更新，跳过 apt image rebuild
- [x] S4.4 更新本文件与 `hive-platform-plan.md` 状态头

### S5 — 审查修复（2026-07-05 复查后补强）

- [x] T1 `schema_migrations` 修复：删除手工插入的 164 双行；migrate oneshot
      up/idempotence 验证通过（version=166, dirty=f）；`hive_smoke.sh`
      preflight 加 `count|dirty = 1|f` 一致性检查
- [x] T2 policy 热推送行为断言：`hive-s3-smoke.mjs` 对在跑 Pod 推
      `tool_pattern=Edit, verdict=deny` → 触发 permission 场景 → 断言
      elicitation 不出现且 turn 正常完成（runner 自动拒绝）
- [x] T3 真实 usage 透传：`acp.TurnUsage` + `OnUsage` 回调；标准 ACP
      transport 解析 `session/prompt` 响应的 `usage` 块；codex transport
      解析 `turn/completed` 的 `turn.usage`；bridge 累计后经
      `PodUsageEvent` 上报；无 agent 报告时保留 len/4 估算 fallback
- [x] T4 smoke 校验模型名：mock 上报 `gpt-4o`（≠ 估算 fallback 默认
      `gpt-4o-mini`），断言 `usage_by_model["gpt-4o"]` 证明真实透传路径
- [x] T5 CI dev-stack job：`bazel.yml` 新增 `hive-e2e` job（backend_only
      → pnpm + chromium → `bash deploy/dev/hive_smoke.sh` S0–S4 全量）

### S6 — P4 平台延伸（2026-07-05 纳入验收）

- [x] S6.1 read-state 落库：`session_read_states` 表 + GORM upsert；
      `PUT/GET` wire 经 Postgres 持久化（`hive-s4-smoke.mjs`）
- [x] S6.2 `session_cost_budget` policy handler + backend turn gate
      （超预算 `402 cost_budget_exceeded`，Pod 保持存活）
- [x] S6.3 `GET /v1/org/usage/summary`：聚合 org 下全部 `pod_session_usage`
- [x] S6.4 Claude live usage：`assistant` message `usage` → `OnUsage` 回调
- [x] S6.5 DoAgent `session/resume` vendor E2E：orchestrator 注入
      `AGENTSMESH_RESUME_EXTERNAL_SESSION` → runner `session/resume`；
      fork smoke 断言 `RESUMED_OK`（`hive-s4-smoke.mjs`）
- [x] S6.6 web Usage 接 compat：`UsageLiveSessionCost` 拉
      `/v1/org/usage/summary`；Connect GetDashboard 合并 `pod_session_usage` token

验收：`bash deploy/dev/hive_smoke.sh` S0–S4 全绿；migration `170`。

- [x] S7.1 gemini/opencode/loopal `CAPABILITY resume acp`（migration 170；标准 ACP `session/resume`）
- [x] S7.2 Codex `thread/resume` transport + `resume cli` fork 续聊
- [x] S7.3 Claude `external_session_id`：`system/init` → `OnSessionID` → backend 捕获
- [x] S7.4 Workstream E：CreatePodForm advanced 折叠；DoAgent 任务入口首位 + 默认选中

### S8 — Omnigent Full Parity（W1–W8，2026-07-05）

> SSOT 细节见 `omnigent-full-parity-plan.md`。代码已落地；本节跟踪验收尾项。

- [x] W1–W7 全部 endpoint 真实现（compat 层无 501 stub）
- [x] W8.1 agents `?after=` 游标分页
- [x] W8.2 `hive-s5-smoke.mjs` + `hive_smoke.sh` parity suite
- [x] W8.3 Bazel `hive_smoke_scripts` 纳入 s5 脚本
- [ ] W8.4 `hive_smoke.sh` S0–S5 全绿（migration 175 + dev stack）
- [ ] W8.5 web-user 全页面手工验收

验收：`bash deploy/dev/hive_smoke.sh` S0–S5 全绿；`bazel.yml` `hive-e2e` job 随 shell 脚本自动覆盖 S5。

## 文件地图（新增/修改）

```
backend/internal/domain/agentsession/
backend/internal/service/agentsession/
backend/internal/api/compat/omnigent/session_*.go
backend/migrations/000163_hive_foundations.up.sql
clients/web/src/lib/agent-capability-axes.ts
clients/web-user/vite.config.ts
docs/rfc/hive-execution-plan.md
```

## 不在本迭代范围

- Omnigent Managed Host / tmux 刮屏 / credential proxy
- conversation_items 全量持久化（需 Relay tap 或 gRPC ACP 事件流）
- web-user 100% 特性 parity
