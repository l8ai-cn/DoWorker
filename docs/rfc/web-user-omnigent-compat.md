# web-user × AgentsMesh 兼容层方案

> `clients/web-user`（Omnigent 前端）要「直接能用」，靠的不是把 Omnigent Python 搬进来，
> 而是在 AgentsMesh Backend 上实现 **Omnigent `/v1/*` 契约的兼容层**，把 Session 语义映射到 Pod/Runner/Relay。

## 1. 结论先行

| 问题 | 答案 |
|------|------|
| Hive 机制（Capability / Policy / Usage / Resume）集成完后，前端能直接用吗？ | **不能自动可用**。机制是 Runner/AgentFile 层；前端吃的是 **HTTP Session API + SSE + WS**。 |
| 还要做什么？ | 新建 **Workstream F：Session Compat API**（Backend BFF 或 `backend/internal/api/compat/omnigent/`）。 |
| 兼容层完成后，前端改多少？ | **很少**：改 Vite 代理目标 + JWT 鉴权 + 去 Omnigent 品牌/托管 host 等可选特性；**UI 组件基本不动**。 |
| 语言要转 Rust 吗？ | **否**。Compat 层用 **Go**；`web-user` 保持 **TypeScript**；Rust WASM 仍只服务 `clients/web` 管理面。 |

## 2. 三前端分工（已定）

```
clients/web-admin     → 系统管理（Runner/审计/配额）
clients/web          → 组织管理面（IDE、Pod、Channel、设置）
clients/web-user     → 终端用户直接用 Agent（聊天/终端/文件）
```

## 3. Omnigent 前端依赖的 API 清单

从 `clients/web-user/src` 静态扫描得出，按 **能否直接映射 AgentsMesh** 分级。

### P0 — 没有就无法打开应用

| Omnigent API | 用途 | AgentsMesh 现状 | 需要拓展 |
|--------------|------|-----------------|----------|
| `GET /v1/me` | 当前用户 | JWT 中间件有 user，无此路由 | 薄封装返回 `{id, email, is_admin}` |
| `GET /v1/info` | 服务能力探测 | 无 | 返回 `{auth_enabled, features: [...]}` |
| `POST /v1/sessions` | 创建会话 | **Pod 创建**（不同模型） | **Session→Pod 适配器**：创建 Pod + 生成 `conv_*` id |
| `GET /v1/sessions` | 侧边栏列表 | Pod 列表 API 有，字段不同 | 列表投影：Pod → Session 摘要 |
| `GET /v1/sessions/{id}` | 会话快照 | Pod GET + ACP snapshot 分散 | 聚合：metadata + items + pending_elicitations |
| `GET /v1/sessions/{id}/stream` | **SSE 事件流** | **无**（Relay 是终端二进制） | **新建**：ACP/Relay 事件 → Omnigent SSE envelope |
| `POST /v1/sessions/{id}/events` | 发消息/审批/中断 | gRPC send_prompt + relay permission | 统一入口，分发到 Runner |
| `GET /v1/sessions/{id}/items` | 历史分页 | 无 conversation_items 表 | 持久化 transcript 或 Pod 事件日志投影 |
| `GET /v1/runners` | 选 Runner | Runner 列表有 | 字段映射（online、available_agents） |
| `GET /v1/agents` | Agent 选择器 | `ListAgents` Connect API | 字段映射 + **capabilities**（Workstream A） |

### P1 — 核心体验完整

| Omnigent API | AgentsMesh 映射 / 新建 |
|--------------|------------------------|
| `PATCH /v1/sessions/{id}` | 绑定 runner、改 model → Pod 更新 + ACP set_model |
| `POST /v1/sessions/{id}/fork` | Workstream D fork API |
| `WS /v1/sessions/updates` | 侧边栏实时刷新（Pod 状态变更 pub/sub） |
| `POST .../elicitations/{id}/resolve` | 现有 ACP permission 响应桥接 |
| `GET /v1/sessions/{id}/resources/...` | Sandbox 文件 API（worktree 路径） |
| Terminal attach WS | **Relay WS** 协议转换层（Omnigent terminal 帧 ↔ Relay binary） |
| `GET /v1/harnesses` | **Capability 声明 API**（A3），替代 harness 注册表 |

### P2 — 协作 / 治理（可分期）

| Omnigent API | Hive 关联 |
|--------------|-----------|
| `GET/PUT /v1/sessions/{id}/permissions` | 新表 session_permissions 或复用 Pod ACL |
| `GET/POST /v1/sessions/{id}/policies` | Workstream B Policy |
| `GET/POST /v1/policies` | Org 级 policy CRUD |
| `GET /v1/policy-registry` | Builtin policy 元数据 |
| `GET /v1/hosts` + `POST /v1/hosts/{id}/runners` | 映射为 Runner fleet；**Managed Host 不做** |
| `GET /v1/hosts/{id}/directories` | Runner workspace 浏览（Host 协议子集） |
| `usage_by_model` / `total_cost_usd` on session | Workstream C 实时 usage |

### P3 — Omnigent 特有（可 stub 或砍掉）

| API | 建议 |
|-----|------|
| `POST /v1/sessions/{id}/switch-agent` | 阶段 3：终止 Pod + 新建 |
| `.../codex_goal` | Codex 专用；DoAgent 为主时延后 |
| Native harness SSE（terminal_pending、todos…） | ACP agent 不需要；UI 按 capability 隐藏 |
| `GET /health?session_ids=` | Runner 健康检查新路由 |
| Comments / projects 侧边栏 | 可二期 |

## 4. AgentsMesh 必须拓展的能力（相对 Omnigent）

这些是 **Backend/Runner 要新建或加厚** 的，与 Hive A–E 正交：

### F1. Session 资源模型（SSOT）

```
Session (conv_id) 1:1 Pod (pod_key)  — 或 session 为 Pod 的视图层
conversation_items[]  — transcript SSOT（Omnigent 有，AgentsMesh 无）
session_events SSE    — 控制面事件总线（非 Relay PTY 字节）
```

### F2. 事件归一化管道（学 Omnigent forwarder，用 Go 实现）

```
Runner ACP events ──gRPC──► Backend SessionEventWriter ──SSE──► web-user
Relay terminal    ──WS────► TerminalCompatBridge        ──WS──► web-user (xterm)
```

AgentsMesh 已有 ACP 事件到 Relay；缺的是 **语义层 conversation_item / session.status / session.usage** 的持久化 + SSE 推送。

### F3. 鉴权桥

- Omnigent：`/v1/me` + optional magic token
- AgentsMesh：现有 JWT → compat 层校验 org 成员 + session permission_level

### F4. 文件 / Workspace API

- Omnigent：`/v1/sessions/{id}/resources/environments/default/filesystem/...`
- AgentsMesh：Runner sandbox 读写在，需 HTTP 包装 + session 作用域

## 5. 集成后 web-user 能否「直接」用？

**分三档：**

| 档位 | 条件 | 前端改动 |
|------|------|----------|
| **MVP 可用** | P0 全部 + Terminal/Relay 桥 + JWT | 改 `vite.config.ts` 代理到 AgentsMesh；删 Omnigent auth token 逻辑 |
| **体验对齐** | + P1（fork、files、harnesses、实时 usage） | 按 capability 隐藏不支持的 Omnigent 控件 |
| **完全等价** | + P2/P3 + Managed Host + 全部 native harness | **不值得**；与 AgentsMesh ACP 路线冲突 |

**推荐目标：MVP → 体验对齐**，不追求 Omnigent 100% 特性 parity。

## 6. 推荐实施顺序

```
已完成:  web-user 拷贝 + Workstream A1/A2 (CAPABILITY parser)
下一步:  A3 capabilities API
并行:    F1 Session↔Pod 域模型 + migration
然后:    F2 SSE 事件管道（绑定已有 ACP）
然后:    P0 compat 路由（/v1/me, /v1/sessions, /v1/runners）
然后:    Terminal compat（Relay 桥）
最后:    web-user vite 代理切到 AgentsMesh + 冒烟 E2E
```

## 7. 与 Hive Workstream 的依赖

| Hive WS | 对 web-user 的价值 |
|---------|-------------------|
| A Capability | `GET /v1/harnesses`、UI 功能开关 |
| B Policy | elicitation 自动规则 + `/v1/policies` |
| C Usage | session 卡片上的 `total_cost_usd` |
| D Resume/Fork | 继续会话 / 分支按钮 |
| E 管理面 UX | **不影响 web-user**（web-user 已是 Omnigent 体验） |
| **F Session Compat** | **web-user 能跑起来的前提** |
