# Hive 方案：Do Worker 底座 + Omnigent 机制吸收

> 状态：Hive S0–S4 验收通过（2026-07-05）| 平台阶段 2+ 延伸见 execution-plan
> 前置调研：Omnigent clone（/private/tmp/omnigent_explain，HEAD 3124850）
> 结论前提：Do Worker 是唯一控制面 SSOT；Omnigent 只吸收契约与机制，不引入其 Python 控制面。

## 0. 双前端架构

| 前端 | 路径 | 角色 | 技术栈 |
|------|------|------|--------|
| 管理面 | `clients/web` + `clients/web-admin` | 组织、Runner、配额、审计、高级配置 | Next.js + Rust WASM |
| 用户工作台 | `clients/web-user` | 终端用户直接跑 agent、聊天、协作 | Vite + React（源自 Omnigent `web/`） |

`clients/web-user` 已从 Omnigent 摘出（Apache-2.0，见 `clients/web-user/THIRD_PARTY.md`），后续通过 **Session API**（`backend/internal/api/rest/v1/session`）对接 Do Worker Backend，不再依赖 Omnigent Server。

## 0.1 语言分工（不是整体转 Rust）

| 层 | 语言 | 职责 |
|----|------|------|
| 业务 SSOT | **Rust → WASM** | 现有 `clients/core`：auth、cache、DTO（管理面 web 继续用） |
| 控制面 / 执行面 | **Go** | Backend、Runner、AgentFile、Policy、Usage（Hive 机制落这里） |
| 管理面 UI | **TypeScript / Next.js** | `clients/web`、`clients/web-admin` |
| 用户工作台 UI | **TypeScript / Vite** | `clients/web-user`（Omnigent 前端改造） |

Omnigent 的 Python 代码**不迁移**；只迁移前端 UI 与机制契约（capability、policy、usage 事件模型）。

## 1. 定位

Hive = 在现有 Do Worker（Backend/Runner/Relay/Web）之上做四个机制升级
（Capability 声明、Policy 引擎、实时 Usage/Cost、Session 恢复统一）+ 一次
前端体验分层改造。所有升级都挂在已存在的扩展点上，不新建平行链路。

## 2. Workstream A — Capability 声明（P0，先行）✅ A1–A3 已落地

**已完成（2026-07-04）**：
- `agentfile/capability`：轴校验（resume/permission/usage/control/…）
- `CAPABILITY` declaration：parser → extract → `AgentSpec.capabilities`
- 单测：`parser_capability_test.go`、`capability/axis_test.go`
- **A3**：`proto/agent/v1/agent.proto` `capabilities` map + Connect handler enrichment
- **A4**：migration `000161_doagent_capabilities`（do-agent 首批声明）

**待做**：A5 ACP runtime 校准；其余 P1+ 见 `hive-execution-plan.md`。

## 2.1 Workstream F — Session Compat API（P0 MVP ✅）

**已完成**：
- migration `000162_agent_sessions`（conv_* id）
- `agentsession` domain + service
- `POST/GET/LIST/PATCH /v1/sessions` → PodOrchestrator
- `POST /v1/sessions/{id}/events`（message → SendPrompt）
- `GET /v1/sessions/{id}/stream`（SSE status + keepalive）
- web-user vite 默认代理 `http://localhost:10000`

**待做**：conversation_items、完整 SSE 事件管道、Terminal 桥接。

现状：capability 信息散在三处且硬编码 —— `agents.supported_modes` 单维度
（backend/internal/domain/agent/agent.go:33）、前端 slug 匹配
（clients/web/src/lib/runner-agent-capabilities.ts）、ACP initialize 只协商
permission modes。RFC frontend-runner-agent-capability-loop 只解决了
"runner 装没装"，没解决"agent 支持什么"。

设计（对齐 Omnigent harness_capabilities 的声明式模型，轴按需裁剪）：

1. AgentFile 新 declaration（SSOT 在 agent 定义里，不加 DB 平行列）：
   `CAPABILITY <axis> <value>`，首批轴：
   - `resume`: `none | cli`（CLI 支持 --resume 类参数）
   - `permission`: `none | acp | notification`（do-agent 是 notification）
   - `usage`: `none | exit | live`
   - `control`: 逗号分隔（`set_model,set_permission_mode`）
   - `interrupt` / `streaming` / `subagents`: `true|false`
   - `model_family`: `claude | gpt | gemini | multi`
2. 链路：agentfile/parser/ast_decl.go 加 `CapabilityDecl` → extract 进
   `AgentSpec`（agentfile/agentspec.go）→ agents API 返回 capabilities map。
3. 运行时校准：ACP initialize handshake 可覆盖声明（declared vs runtime
   两层，学 Omnigent 的 capabilities-bench seam，声明为假是 bug）。
4. 前端：runner-agent-capabilities.ts 扩展 `agentSupports(agent, axis)`，
   Create Pod / Loop / Coordinator 的功能开关全部由声明驱动，删除
   slug 字符串分支。

## 3. Workstream B — Permission Policy 引擎（P0）

现状：ACP tool permission 纯人工审批 + 60s 超时 deny
（AcpToolPermissionCard），无 org 级默认规则；backend/pkg/policy 是
RBAC，不是 tool policy。

设计（三档 verdict，ASK 是一等公民）：

1. 表 `permission_policies`：org_id、scope（org|project|pod）、
   agent_slug（可空）、tool_pattern、path_pattern、verdict
   （allow|deny|ask）、priority。两层即可，不照搬 Omnigent 三层。
2. 评估位置在 Runner：规则 SSOT 在 Backend，随 CreatePodCommand 快照
   下发（变更经 gRPC 推送）。Runner 在 `HandlePermissionRequest`
   （runner/internal/acp/handler.go:155）先评估：
   - ALLOW → 直接 RespondToPermission，事件仍上报 Backend 审计；
   - DENY → 拒绝 + 审计事件；
   - ASK / 无匹配 → 走现有 relay permissionRequest 卡片（行为不变）。
3. 现有 session 级 alwaysAllow（addRules）并入同一模型，成为 pod scope
   的临时规则。
4. 守住 Omnigent 的边界原则：平台 ALLOW 不压制 vendor 原生确认；
   policy 数据不可达时 fail-safe 到 ASK（不是 deny，先保可用性）。
5. 阶段 2 增加 `cost_budget` policy 类型：超预算不杀 pod，改为拒绝
   expensive model 的后续 turn（downgrade gate），依赖 Workstream C。

## 4. Workstream C — 实时 Usage 与成本（P0）

现状：token 只在 Pod 退出后由 tokenusage parser 批量采集；token_usages
无美元成本；无 session 级 budget。

设计：

1. proto 加 `PodUsageEvent`（cumulative input/output/cache tokens +
   model，SET 语义，学 Omnigent external_session_usage）。ACP transport
   在 turn 完成时上报（claude/codex/do-agent 先行）；PTY-only agent 保留
   退出后 parser 作为兜底。
2. Backend：token_usages 支持实时 upsert；新增 model_prices 表，聚合层
   算 USD（现有 AggregationFilter 复用）。
3. 前端：org Usage 图表接实时；Pod 详情显示当前 session 花费。
4. 产出物同时是 Workstream B cost gate 的数据输入。

## 5. Workstream D — Session 恢复统一 + Fork（P0 后半）

现状：Pod sandbox resume 完整（LocalPathStrategy），但 agent session
resume 只有 claude/codex 在各自 AgentFile 手写；vendor session id 无统一
存储；无 fork。

设计：

1. vendor session id 统一捕获：各 ACP transport 拿到 vendor session/thread
   id 后经新 gRPC 事件上报，存 pods 表 `external_session_id`（等价
   Omnigent 的 PATCH external_session_id 模式）。
2. systemConfigKeySet（pod_orchestrator_system_config.go）增加
   `resume_external_session`；AgentFile 用 CAPABILITY resume=cli +
   build logic 模板消费该 key，替代各家手写。
3. DoAgent 先行补 resume RPC（自家 runtime，transport_control.go 已有
   session/* 通道），作为抽象验证样板，再推广 gemini/opencode/loopal。
4. Fork（阶段 3）：create pod API 加 `fork_from_pod_key` —— 复制 sandbox
   （复用 LocalPathStrategy）+ 能 vendor-resume 的 agent 走 resume，
   否则冷启动。数据模型现在就定（source + 可选截断点）。

## 6. Workstream E — 前端体验分层（与 A-D 并行推进）

原则：Omnigent 赢在入口体验，输在平台模型。Hive 保留平台对象，但按
用户角色分层暴露：

1. 普通用户首页 = 任务工作台：入口是"开始一个任务"（选 agent + 描述 +
   可选 repo），Runner 自动选择（hasRunnerForAgent 已有），Pod/Runner/
   AgentFile 概念收进高级设置。CreatePodForm 已拆分组件，在其上做
   simple/advanced 两档。
2. 高级用户：Agent 配置、Repo、Knowledge、Permission 表单化
   （由 ResolveConfigSchema + Workstream A capability 驱动，不暴露 DSL）。
3. 管理员：Runner/Relay/Quota/Audit 保持在 admin console 与 org settings。
4. DoAgent 作为默认推荐 agent 出现在任务入口第一位。

## 7. 明确不做

- 不引入 Omnigent Server/Host/Runner 控制面（双控制面冲突）。
- 不做 tmux/TUI 刮屏、approval-mirror（ACP 原生协议优于刮屏）。
- 不用 Omnigent YAML spec 替换 AgentFile（只吸收字段思路）。
- 不做 WS-tunneled HTTP 控制通道（gRPC+mTLS 已覆盖）。
- Credential proxy（swap-on-access）与 Managed Host 列为远期安全/云化
  roadmap，不进本方案迭代。

## 8. 阶段计划

| 阶段 | 内容 | 验收 |
|---|---|---|
| 1 | A capability 声明全链路 + C 实时 usage 上报 | 前端零 slug 分支；pod 运行中可见 token 增长 |
| 2 | B policy 最小版（org 规则 + runner 评估）+ C 成本换算 + D vendor session id 捕获 | org 规则可 auto-allow/deny；usage 显示 USD |
| 3 | D DoAgent resume + fork API + B cost gate + E 任务工作台 | fork 出的 pod 能续上下文；超预算 turn 被 gate |
| 4（远期） | capability bench DRIFT 检测；credential proxy 评估 | 声明与行为不一致 CI 失败 |

## 9. 风险

- proto 变更需 Backend/Runner 同步发布（沿用现有 runner 版本协商）。
- policy 快照下发的规则更新延迟：先接受 pod 创建时快照，热更新走
  gRPC 推送在阶段 2 再做。
- vendor session id 的稳定性依赖各 CLI 行为，capability 声明须如实
  标注 `resume=none` 而不是乐观声明。
