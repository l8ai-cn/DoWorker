# 知识库（Knowledge Base）模块开发文档

组织级知识库：Git 为底座，遵循 Karpathy llm-wiki 三层模型 + llms.txt 索引规范，
可只读/读写挂载到 Agent Pod，并通过 connector 单向同步飞书/钉钉/Google 外部知识源。

## 设计理念

- **Wiki 是编译产物，raw 是不可变事实源**：每个 KB 是内部 Gitea 里一个标准布局的
  Git 仓库；`raw/` 只进不改，`wiki/` 由 LLM 维护，随时可从 raw 重编译。
- **Agent 就是 wiki 维护者**：仓库根部的 `AGENTS.md` 定义页面结构、ingest 流程、
  交叉引用与矛盾标记规范。"ingest / 重编译"就是一次带 rw 挂载的 Pod 运行。
- **llms.txt 是索引层**：agent 挂载后用它做低成本导航。v1 不做向量检索，
  靠 grep/文件树（个人/团队规模下结构化 wiki 优于 RAG）。
- **Git 是历史/审计/回滚层**：agent 写回 = commit + push，人可 review diff。
- **连接器统一物化为 Git**：外部源同步为 markdown 落入 `raw/{source_type}/`，
  挂载管道只认 Git，一条链路。

### KB 仓库标准布局（scaffold 模板：`backend/internal/service/knowledgebase/scaffold/`）

```
llms.txt        # 索引：H1 名称 + blockquote 摘要 + H2 分节链接列表
AGENTS.md       # schema：wiki 维护规范
raw/            # 不可变原始资料（connector 同步目录 raw/feishu/...）
wiki/
  index.md      # 总览
  log.md        # 变更日志（## [YYYY-MM-DD] operation | Title）
```

## 数据模型

迁移 `backend/migrations/000166_knowledge_bases.up.sql`：

- `knowledge_bases`：org 作用域（`UNIQUE(organization_id, slug)`，slug 走 slugkit
  CHECK），`git_repo_path` / `http_clone_url` / `default_branch` 指向 Gitea 仓库，
  `source_type IN ('git','feishu','dingtalk','google')`，`source_config` JSONB
  存 connector 配置，`sync_status` / `last_synced_at` 存同步状态。
- `knowledge_base_agent_mounts`：agent 默认挂载（`agent_slug` + `mode ro|rw`）。
  Pod 级挂载走创建请求，不落此表。

`source_config` 中的连接器凭证和仓库级 SSH deploy private key 在写入时加密
（`enc:v1:` 前缀，平台 `crypto.Encryptor`）。连接器凭证只在 sync worker 调用
外部源前解密；deploy private key 只在组装 Pod 挂载命令时解密，且不会返回 API。
见 `service/knowledgebase/source_secrets.go`。

## 代码地图

| 层 | 位置 | 职责 |
|---|---|---|
| 域 | `backend/internal/domain/knowledgebase/` | 实体 + Repository 接口 + slugkit hooks |
| Gitea client | `backend/internal/infra/gitea/` | 建/删仓、提交、tree/contents API、deploy key 生命周期 |
| 服务 | `backend/internal/service/knowledgebase/` | CRUD、provisioner（scaffold 初始化）、mounts 解析、search、sync |
| 连接器 | `.../knowledgebase/connector/` | `Connector` 接口 + feishu/dingtalk/google 实现 |
| Connect API | `backend/internal/api/connect/knowledgebase/` | 9 个 RPC（CRUD、挂载、文件/目录读） |
| MCP 分发 | `backend/internal/api/grpc/runner_adapter_mcp_kb.go` | `kb_list/kb_search/kb_read/kb_write` 的 backend 侧 |
| Runner 挂载 | `runner/internal/runner/pod_builder_kb*.go` | 用仓库级 SSH key clone 到 `{sandbox}/kb/{slug}`，校验路径并按 ro/rw 收敛凭据 |
| Runner MCP | `runner/internal/mcp/http_tools_kb.go` + `grpc_client_kb.go` | Pod 内 kb_* 工具 HTTP 面 |
| Agentfile | `agentfile/`（parser/extract/merge/serialize） | `KNOWLEDGE team-docs, product-wiki [rw]` 声明 |
| Rust core | `clients/core/crates/api-client/src/modules/knowledgebase.rs` + `wasm/src/service_kb*.rs` | Connect 客户端 + WASM 绑定 |
| Web | `clients/web/src/components/knowledgebase/` + `app/(dashboard)/[org]/knowledge-base/` | 列表/详情/文件树/Ingest 入口 |
| 内置 skill | `backend/internal/service/agent/builtin_skills/am-knowledge.md` | 教 agent 按 llm-wiki 工作流使用挂载与工具 |

## 关键链路

### 挂载（Pod 创建 → 沙箱可用）

1. 前端 `KnowledgeBaseMountSelect` 产出 `{slug, mode}` 列表，
   `agentfile-layer.ts` 生成 `KNOWLEDGE` 声明进 Agentfile layer。
2. orchestrator（`pod_orchestrator_*`）合并 Agentfile 声明 + agent 默认挂载表，
   经 `service/knowledgebase/mounts.go` 解析为 `KnowledgeMount`
   （`proto/runner/v1`，含 SSH clone URL、仓库级 private key 和固定 host key）。
3. Runner 在主 worktree 建好后、prep script 前校验挂载目标的真实路径，拒绝越出
   sandbox 的 symlink，再用 `StrictHostKeyChecking=yes` 逐个 clone。
4. ro 挂载清除写凭据；rw 挂载安装对应 deploy key。Pod 恢复时会重新收敛
   remote URL、key 和 known_hosts，取消挂载时删除 Runner 管理的凭据文件。

### MCP kb_*（跨库检索 + 未挂载访问）

Pod 内 MCP HTTP server → Runner gRPC → backend `dispatchMcpMethod` →
knowledgebase service。org 隔离用 `authenticatePod` 的 `TenantContext`。
`kb_write` 由 backend 代提交，commit author 标注 pod。

### 外部源同步

`sync_worker.go`（仿 MarketplaceWorker，周期 `KB_SYNC_INTERVAL`，默认 1h）：
`ListExternal` 找出非 git 源的 KB → 取对应 `Connector` → `SyncFromConnector`。
增量判断用 Git blob SHA（`gitBlobSHA` vs Gitea tree API 的 entry SHA），
内容未变则跳过 commit；完成后更新 `sync_status` / `last_synced_at`。

## 配置与部署依赖

- `KB_GITEA_URL` / `KB_GITEA_TOKEN` / `KB_GITEA_ORG`（默认 `am-kb`）/
  `KB_GITEA_CLONE_URL` / `KB_GITEA_SSH_CLONE_URL` / `KB_GITEA_KNOWN_HOSTS`：
  任一控制面或 SSH 挂载配置缺失时 KB 服务不启用。Gitea 是 KB 功能的必选生产依赖。
  KB namespace org 在首次建仓时由 provisioner 的 `EnsureNamespace` 自动创建。
- `KB_SYNC_INTERVAL`：外部源同步周期，默认 1h。
- dev 环境：`deploy/dev/gitea/init-gitea.sh` 签发 backend token 到
  `runtime/gitea/backend-token`，`lib/host_services.sh` 据此导出 `KB_GITEA_*`。

## 扩展点

- **新增外部源**：实现 `connector.Connector`（`SourceType` / `ListDocs` /
  `FetchDoc → markdown`），注册进 `connector.NewRegistry`，
  在迁移的 `source_type` CHECK 里加枚举值。
- **检索升级**：当 KB 规模超上下文预算时再评估向量检索；当前 `search.go`
  为服务端 grep。
- **同步后自动 ingest**：策略可存 `source_config`，同步完成后触发带 rw
  挂载 + ingest prompt 的 Pod（预留，未实现）。

## 关键取舍

- 连接器单向物化（外部 → raw/），不做双向同步，避免冲突解决复杂度。
- Pod 级挂载不落库，只存在于创建请求与 Agentfile；持久默认挂载才进
  `knowledge_base_agent_mounts`。
- KB 删除先撤销 deploy keys 并删除 Gitea 仓库，再删除 DB 行。若远端已成功而 DB
  删除失败，服务会清空仓库定位字段并标记同步失败，使挂载立即拒绝；后续删除调用
  只重试 DB 删除，不会重新访问已经删除的仓库。
