# Omnigent 语义全量迁移执行计划（Full Parity）

> 状态：W1+W2+W3 已落地 | 标准：web-user 每一个 `/v1/*` 调用都有真实实现（非 stub）
> 前置：`hive-execution-plan.md` S0–S7 已完成（compat MVP + 机制层）
> 本文是后续代码开发的 SSOT：按工作包顺序执行，每包有明确验收。

## 1. 目标标准

**全量可用** = 打开 web-user 任意页面、任意交互路径，控制台无 404/501；
所有语义（会话、文件、协作、策略、agent 控制）走 AgentsMesh 真实数据。
Omnigent Python 控制面仍不迁移；语义全部落 compat 层 + Pod/Runner。

## 2. 缺口矩阵（web-user 实际调用 vs compat 现状）

| Endpoint | 现状 | 工作包 |
|---|---|---|
| `DELETE /v1/sessions/{id}` (`?delete_branch=`) | **无路由** | W1 ✅ |
| `GET /v1/sessions/projects` + PATCH project 字段 | **无路由** | W1 ✅ |
| `GET /v1/sessions/{id}/agent` | **无路由** | W1 ✅ |
| `GET /v1/sessions/{id}/read-state` | 只有 PUT | W1 ✅ |
| conversation_items 全类型（tool_call/reasoning） | 部分持久化 | W1 ✅ |
| `GET/PUT .../environments/default/filesystem[/*path]` | GET 501，PUT 无 | W2 ✅ |
| `GET .../environments/default/changes` | **无路由** | W2 ✅ |
| `GET .../environments/default/diff/{path}` | **无路由** | W2 ✅ |
| `GET .../environments/default/search` | **无路由** | W2 ✅ |
| `GET /v1/hosts/{id}/filesystem` | 假数据（静态两层） | W2 ✅ |
| `POST /v1/sessions/{id}/resources/files`（上传） | **无路由** | W3 ✅ |
| `GET .../files/{id}/content`（SessionImage） | **无路由** | W3 ✅ |
| `GET/POST/DELETE .../comments`、`/comments/send` | **无路由** | W4 |
| `PUT/DELETE /v1/sessions/{id}/permissions` | 501 | W4 |
| `GET /v1/sessions/{id}/child_sessions` | **无路由** | W4 |
| `POST /v1/sessions/{id}/switch-agent` | 501 | W5 |
| `GET/POST/DELETE .../agent/mcp-servers[/{name}]` | **无路由** | W5 |
| `GET/POST/DELETE .../codex_goal`、`/status` | **无路由** | W5 |
| `POST /v1/sessions/{id}/policies`（session 级） | 501 | W6 |
| `PATCH /v1/policies/{id}`（更新非禁用） | 501 | W6 |
| `POST /v1/hosts/{id}/directories` | 501 | W7 |
| `GET /v1/agents?after=`（游标分页） | 忽略参数 | W8 |

## 3. 工作包

### W2 — Sandbox 文件系统 gRPC 通道（先行，最大块）

其余文件类功能全依赖它。proto 无任何文件命令 → 新增：

1. `proto/runner/v1`：`SandboxFsCommand{op: list|read|write|changes|diff|search, pod_key, path, payload}` + `SandboxFsResultEvent{request_id, entries|content|error}`（request/response over bidi stream，参照 `QuerySandboxesCommand` 模式）
2. Runner：`runner/internal/runner/sandbox_fs_handler.go` — 限制在 `pod.SandboxPath` 内（path traversal 防护）；changes/diff 用 `git status`/`git diff`（worktree sandbox）；search 用 walk + 上限
3. Backend：`backend/internal/api/grpc/runner_adapter_fs.go` — 带超时的同步 request/response 关联
4. Compat：filesystem GET/PUT、changes、diff、search 五个 handler + host filesystem 真实化（runner 报家目录树，替换 `compatFilesystemEntries` 假数据）

验收：FileViewer 读/写文件、ChangedFiles 面板、diff 视图、workspace 搜索全通；`hive-s5-smoke.mjs` 断言 write→read 回读一致 + git diff 出现修改。

### W1 — Session 生命周期补全

1. `DELETE /v1/sessions/{id}`：TerminatePod + agent_sessions 软删；`delete_branch=true` 透传 runner 删 worktree 分支
2. `agent_sessions.project` 列（migration 171）+ `GET /v1/sessions/projects`（DISTINCT 聚合）+ PATCH 支持 project
3. `GET /v1/sessions/{id}/agent`：pod → agent 投影（id/name/harness/model/capabilities）
4. `GET .../read-state`：复用 `session_read_states`
5. conversation_items 补 `tool_call` / `reasoning` 类型：EventBridge 已有 tap，扩 `event_bridge_assistant.go` 的 item 种类

验收：侧边栏删除会话、项目分组、Inbox 已读、历史含工具调用条目。

### W3 — 文件上传资源

1. `session_files` 表（migration 172：id/session_id/filename/bytes/minio_key）
2. `POST .../resources/files`（multipart → MinIO，复用 `backend/internal/domain/file`）+ `GET .../files/{id}/content`
3. `POST /events` 的 message 块引用 file id → prompt 注入文件路径/内容

验收：聊天框粘贴图片上传成功，SessionImage 渲染。

### W4 — 协作（comments / permissions / lineage）

1. `session_comments` 表（migration 173）+ CRUD + `/comments/send`（send = 评论转 prompt 注入）
2. `session_permissions` 表（migration 173）+ PUT/DELETE 真实现；GET 从表读（owner 恒 level 4）；被授权者 List/Get 放行（`authorizeSession` 扩权）
3. `child_sessions`：fork 时已存 source → 反查 `agent_sessions.forked_from`

验收：两账号共享会话（A 授权 B，B 能打开）；文件评论闭环；fork 树展示。

### W5 — Agent 控制

1. `switch-agent`：终止旧 Pod → 以同 session_id + resume_external_session 建新 Pod（能 resume 的 agent 续上下文，否则冷启动带 items 摘要）
2. `agent/mcp-servers` CRUD：`agent_sessions.mcp_servers` JSONB；变更后走 switch-agent 同路径重建 Pod 生效
3. `codex_goal` 三路由：桥接 DoAgent goal RPC（`doagent.rpc` relay 通道已有 goal/list、goal/pause）；非 DoAgent harness 返回空

验收：会话中途换 agent 继续对话；MCP server 添加后新 turn 可用；DoAgent 会话 goal 面板可读可停。

### W6 — Policy 补全

1. session 级 policy：`permission_policies.session_id` 列（migration 174）；create/delete 后即时 `SendUpdatePodPolicyRules` 推送该 Pod
2. `PATCH /v1/policies/{id}` 支持改 verdict/pattern（不只 disable）

验收：单会话 deny 规则只影响该会话；org 规则可编辑。

### W7 — Host 目录管理

`POST /v1/hosts/{id}/directories`：W2 fs 通道加 `mkdir` op（限 runner 配置的 workspace root 内）；`ResumeWithDirectoryDialog` 全流程可用。

### W8 — 收尾

1. `GET /v1/agents?after=` 游标分页
2. `hive_smoke.sh` 增 S5（W1–W7 各一断言）；CI `hive-e2e` 覆盖
3. web-user 全页面手工过一遍：console 零 404/501
4. 更新本文档状态头 + `hive-execution-plan.md`

## 4. 执行顺序与依赖

```
W2 (fs 通道, proto 变更) ──► W7 (mkdir)
W1 (session 生命周期)    ──► W4 (child_sessions 用 fork 列)
W3 / W5 / W6 互相独立，可并行
W8 最后收尾
建议顺序：W2 → W1 → W3 → W4 → W5 → W6 → W7 → W8
```

proto 变更（W2）需 backend + runner 同步构建；migration 序号从 **171** 起。
每个工作包完成即本地 commit（不开 PR），跑 `bash deploy/dev/hive_smoke.sh` 回归。

## 5. 范围外（唯二例外，需单独决策）

- **Managed Host / credential proxy**：Omnigent 云托管特有，与 self-hosted Runner 模型冲突，前端 `/v1/info` 已报 `managed_sandboxes_enabled=false` 走本地分支
- **native harness 专有 SSE**（terminal_pending/todos 等）：ACP agent 无此语义，UI 按 capability 自动隐藏

除此之外 web-user 引用的全部端点都在 W1–W8 内，无其它砍项。
