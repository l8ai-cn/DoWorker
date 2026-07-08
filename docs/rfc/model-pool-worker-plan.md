# 模型管理 + Worker 挂载模型 执行计划（Model Pool）

> 状态：M1–M3 + web-user UI 已落地 | 待：migration 176 跑完 + MINIMAX_API_KEY seed + 端到端验证
> 目标：多 Worker 管理端有「模型管理」，创建 Worker（web-user 新建会话）时可选择挂载模型 + 设置配额；do-agent 默认走 MiniMax，开箱即用。

## 1. 目标（可验收）

1. **模型池（Model Pool）**：org 级维护一组「模型配置」，每条 = provider 类型 + 凭证（加密）+ 一个或多个可用 model id + 默认 model + endpoint。
2. **创建会话时挂载模型**：web-user 新建会话，从模型池选一个模型配置，并可设置该会话的**配额（成本上限 USD）**。
3. **注入链路**：所选模型配置 → `handleCreateSession` 拼进 AgentfileLayer（`CONFIG model = ...` + `USE_ENV_BUNDLE`/credential env） → runner → do-agent 读到 provider key + model，不再 exit 1。
4. **开箱即用**：dev 启动 seed 一条 MiniMax 模型配置（key 从 `deploy/dev/.env` 读），新建会话默认选它即可直接对话。

## 2. 现状与复用

- **`UserAIProvider`**（`backend/internal/domain/agentpod/settings.go`）：已有加密凭证、`provider_type`、`is_default`、`ProviderEnvVarMapping`。**复用并扩展**，不新造。
- 缺：org 级可见（当前仅 user scope）、MiniMax provider type、可选 model 列表、默认 model、配额概念、web-user 侧 CRUD/选择 UI、`handleCreateSession` 注入。
- `handleCreateSession`（session API）当前忽略 `model_override`；注入点在此。
- do-agent agentfile 已声明 `CONFIG model` + `ENV *_API_KEY SECRET OPTIONAL` + `config_json`→settings.json，**注入机制已完备**，只差把模型配置喂进去。

## 3. 工作包

### M1 — 模型池数据模型（migration）
扩展 `user_ai_providers` → 语义升级为「模型配置」（保留表名或新增 org 级表 `ai_model_configs`）：
- 加 `organization_id`（org 级可见，NULL=user 私有）
- 加 `provider_type` 允许值 `minimax`
- 加 `models JSONB`（可用 model id 列表）、`default_model`、`base_url`
- 加 `cost_budget_usd`（该配置默认配额，可空）
- `ProviderEnvVarMapping` 增 `minimax`（`MINIMAX_API_KEY` / `MINIMAX_BASE_URL`，按 do-agent 约定）

### M2 — 模型池 CRUD + 列表 API
- org 级：`GET/POST/PUT/DELETE /v1/organizations/:slug/model-configs`（复用 AIProviderService，扩 org 查询）
- web-user 选择用：`GET /v1/model-configs`（返回可见模型配置：id/name/provider/models/default_model，**不含明文 key**）

### M3 — 创建会话注入
- `createSessionBody` 加 `model_config_id`、`model`（具体 model id）、`cost_budget_usd`
- `handleCreateSession`：按 id 取配置 → 解密凭证 → AgentfileLayer 追加：
  - `CONFIG model = "<model>"`
  - credential env（走 EnvBundle 或直接 env 注入，取 do-agent 需要的 `MINIMAX_API_KEY` 等）
  - `cost_budget_usd` → 写 session label（复用已有 cost_control 机制）

### M4 — web-user 模型管理界面
- 设置区新增「模型管理」：列表 + 增删改（provider / key / models / default / 配额）
- 表单按 provider 类型显示对应字段（minimax/anthropic/openai）

### M5 — web-user 新建会话选择模型 + 配额
- `NewChatDialog` 增「模型」选择器（拉 `/v1/model-configs`）+ 具体 model 下拉 + 配额输入
- POST body 带 `model_config_id` / `model` / `cost_budget_usd`
- 默认选中 org 的 default 模型配置（MiniMax）

### M6 — dev seed 开箱即用
- 启动时若无模型配置，从 `deploy/dev/.env`（`MINIMAX_API_KEY`）seed 一条 org 级 MiniMax 默认配置
- runner-entrypoint 的 do-agent 默认 settings.json 逻辑改为「无 key 时给清晰错误」而非静默 minimax

### M7 — 验证
- 新建 do-agent 会话选 MiniMax → 发消息 → 收到真实回复（非 exit 1）
- 无 key 时报清晰错误（不再 100ms 静默失败）

## 4. 顺序与依赖
```
M1 → M2 → M3 (backend 链路先通)
M4 / M5 (前端，依赖 M2 API)
M6 (dev seed，依赖 M1/M2)
M7 收尾验证
```

## 5. 范围外
- 真实计量扣费/配额强制阻断已有 `cost_budget` 机制（P4 已落地），本计划只做「创建时设置配额」的入口，复用既有 gate。
- provider 的真实 model 列表拉取（走 provider API discovery）暂不做，手工维护 model 列表。
