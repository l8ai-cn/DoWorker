# 模型管理 + Worker 挂载模型 执行计划（Model Pool）

> 状态：已被 Unified AI Resource Management 替代。旧“模型配置”方案仅作为历史背景保留。
> 目标：创建 Worker（web-user 新建会话）时选择一个精确 AI model resource，可显示配额入口；do-agent 通过所选资源生成运行配置。

## 1. 目标（可验收）

1. **AI Resource Center**：org/user 级维护 provider connection 与 model resource；凭据只在 provider connection 中加密保存。
2. **创建会话时选择模型资源**：web-user 新建会话，从可见资源中选一个 chat 资源，并可设置该会话的**配额（成本上限 USD）**。
3. **运行链路**：所选资源 ID → backend 校验可见性/启用状态/能力兼容 → 生成 harness 运行配置 → runner → do-agent 读到 provider key + model。
4. **开箱即用**：dev seed 应通过 AI Resource Center 写入一条 MiniMax 默认资源。

## 2. 现状与复用

- 旧 `UserAIProvider`/模型池表不再作为 Worker 认证入口。
- 缺口已收束为：将历史模型池数据迁移到 provider connections/model resources，并删除旧注入路径。
- `handleCreateSession` 必须提交 `model_resource_id`，不能自动回退到任何隐式认证路径。
- do-agent agentfile 已声明 `CONFIG model` + `config_json`→settings.json；backend 只从所选资源生成运行配置。

## 3. 工作包

### M1 — 模型池数据模型（migration）
迁移旧 provider/model 数据到 `provider_connections` 与 `model_resources`：
- 加 `organization_id`（org 级可见，NULL=user 私有）
- 加 `provider_type` 允许值 `minimax`
- 加 `models JSONB`（可用 model id 列表）、`default_model`、`base_url`
- 加 `cost_budget_usd`（该配置默认配额，可空）
- `ProviderEnvVarMapping` 增 `minimax`（`MINIMAX_API_KEY` / `MINIMAX_BASE_URL`，按 do-agent 约定）

### M2 — 模型池 CRUD + 列表 API
- org/user 级 CRUD 走 AI Resource Connect API。
- web-user 选择用：`GET /v1/model-resources`（返回可见资源：id/name/provider/model/default，**不含明文 key**）

### M3 — 创建会话注入
- `createSessionBody` 加 `model_resource_id`、`model`（具体 model id）、`cost_budget_usd`
- `handleCreateSession`：按资源 id 解析 provider connection → 解密凭证 → 生成运行配置：
  - `CONFIG model = "<model>"`
  - credential env（走 EnvBundle 或直接 env 注入，取 do-agent 需要的 `MINIMAX_API_KEY` 等）
  - `cost_budget_usd` → 写 session label（复用已有 cost_control 机制）

### M4 — web-user 模型管理界面
- 设置区新增「模型管理」：列表 + 增删改（provider / key / models / default / 配额）
- 表单按 provider 类型显示对应字段（minimax/anthropic/openai）

### M5 — web-user 新建会话选择模型 + 配额
- `NewChatDialog` 增「模型」选择器（拉 `/v1/model-resources`）+ 具体 model 下拉 + 配额输入
- POST body 带 `model_resource_id` / `model` / `cost_budget_usd`
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
