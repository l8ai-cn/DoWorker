# Marketplace Runtime Bridge and Migration Design

- **Date:** 2026-07-11
- **Scope:** Expert、Skill、MCP、Resource 接入，用量事件，旧市场迁移

## 1. Runtime Bridge Contract

新增 `proto/marketplace_runtime/v1/marketplace_runtime.proto`，由现有 Backend
实现 ConnectRPC，仅允许 Marketplace API 的 mTLS service identity 调用。

方法：

```text
DescribePublishableResource
ValidateCatalogItemVersion
PreflightMarketplaceInstallation
ApplyMarketplaceInstallation
GetMarketplaceInstallationOperation
SuspendMarketplaceInstallation
UninstallMarketplaceInstallation
```

`DescribePublishableResource` 输入 resource type/id 和 actor，返回展示元数据、
immutable revision、可发布性和权限要求。

`PreflightMarketplaceInstallation` 输入 installation UUID、目标 org、manifest、
bindings 和 actor，复用 `workercreation.Service.Preflight`、Skill/MCP
compatibility 和组织权限检查，返回：

```text
plan_id, plan_digest, expires_at
resolved_resources[]
required_permissions[]
mutations[]
estimated_usage[]
blocking_issues[]
warnings[]
```

`Apply` 必须携带 plan ID、digest 和 idempotency key。计划过期、版本变化或资源
revision 变化返回 `FAILED_PRECONDITION`，不得静默重新解析。

## 2. Runtime Bridge Domain

新增：

```text
backend/internal/domain/marketplacebridge/
backend/internal/service/marketplacebridge/
backend/internal/api/connect/marketplacebridge/
```

### `marketplace_runtime_installations`

字段为 `id bigint`、`marketplace_installation_id uuid UNIQUE`、
`organization_id bigint`、`listing_version_ref varchar(100)`、
`status varchar(20)`、`plan_digest char(64)`、`created_by_id bigint`、
`created_at/updated_at`。

### `marketplace_runtime_installation_resources`

字段为 `id bigint`、`runtime_installation_id bigint`、
`resource_type varchar(30)`、`resource_id bigint`、`role varchar(30)`、
`provisioning_mode created/reused`、`created_by_operation_id uuid`、`created_at`。
created 资源必须有 operation ID，reused 资源必须为空。唯一索引为
`(runtime_installation_id, resource_type, resource_id, role)`。

Bridge 表记录映射，不把 marketplace 字段散落到 Expert、Skill 和 MCP 表。

WorkerSpec metadata 新增 `source_marketplace_installation_id` UUID，使使用事件
稳定归因，不能通过可变 Expert slug 反查。

## 3. Resource Adapters

| Type | Existing source | Adapter responsibility |
| --- | --- | --- |
| Application | `domain/expert`, WorkerSpec | 从 manifest 创建 org-owned Expert |
| Skill | `domain/skill`, `InstalledSkill` | 校验 digest，创建 pinned install |
| MCP Connector | `McpMarketItem`, `InstalledMcpServer` | 创建模板实例，密钥留在 Runtime |
| Resource | AI resource/model/compute/knowledge | 建立 entitlement 或 binding |

Application adapter 不调用当前 `InstallMarketApplication`。它根据不可变 manifest
创建 Expert，并通过明确 adapter 安装依赖。任一步失败进入补偿。

Marketplace Skill 必须写入 `PinnedVersion`、`ContentSha` 和 `StorageKey`，不得
使用现有 unpinned live-follow 行为。

MCP Connector 只复制模板和 schema；HTTP header、env 和 OAuth token 由 Runtime
安全存储，Marketplace API 永不接收明文。

## 4. Required Runtime Model Corrections

- Marketplace Application 必须引用 Expert 的 immutable WorkerSpec snapshot；
  新增 `experts.worker_spec_snapshot_id` FK，Marketplace Expert 必填。
- Expert Run 检测 snapshot 后必须通过 WorkerSpec prepare/run 路径执行并校验
  immutable resource revision，不再从 `SkillSlugs` 或 Agentfile 字段重建配置。
- `InstalledSkill.RepositoryID` 和 `InstalledMcpServer.RepositoryID` 当前强制非空，
  迁移为 `target_type=organization/repository` 与可空 `repository_id`，并用 DB
  CHECK 约束字段组合。
- 现有 `scope=org/user` 保留为所有权范围，不再承担安装目标含义。
- 既有安装回填 `target_type=repository`，禁止创建虚拟仓库伪装组织级安装。
- Marketplace Skill 写入 WorkerSpec `skill_ids` 和 bridge resource rows，不通过
  slug 在运行时重新选择版本。
- Application 的 primary Expert 必须按 Installation 创建，不能标记 reused。
- Resource mapping 不用于推断计费来源。每次 Marketplace 发起的运行或 MCP
  调用必须携带 Bridge 签发的 usage context，绑定 installation、org、实际 user、
  expiry 和 nonce；Usage emitter 只按该 context 归因。
- 同一 reused Skill/MCP 可服务多个 Installation，但无 usage context 的普通调用
  不得记入 Marketplace 账本。

## 5. Installation Execution

Runtime Apply 阶段：

```text
validate service identity
validate actor and organization role
load and lock plan
create bridge installation
install pinned dependencies
create primary resource
run post-install verification
mark bridge active
return opaque runtime refs
```

补偿只反向删除 `provisioning_mode=created` 且 operation ID 匹配的资源。reused
资源不得删除；崩溃恢复使用 resource rows 重建补偿范围。

## 6. Usage Event Flow

Runtime 产生模型 Token、Worker 时间、MCP call、存储或 application run 时先写
本地 outbox，再调用：

```text
POST /api/marketplace/v1/internal/quota-reservations
POST /api/marketplace/v1/internal/usage-events
POST /api/marketplace/v1/internal/quota-reservations/{id}/release
```

事件字段：event UUID、marketplace installation UUID、platform org/user、meter
key、quantity、reservation UUID、occurred_at、source 和非敏感 metadata。

Marketplace API 以 event UUID 幂等写 UsageEvent 和 LedgerEntry。投递失败留在
outbox 重试并告警，不能丢弃，也不能让运行请求同步等待结算。

## 7. Existing Code Touchpoints

| Existing path | Integration |
| --- | --- |
| `backend/internal/service/workercreation` | 复用 WorkerSpec Prepare/Preflight |
| `backend/internal/domain/expert` | 创建组织所有 Expert |
| `backend/internal/domain/skill` | 校验 catalog version 和 digest |
| `backend/internal/domain/extension` | 创建 pinned Skill/MCP install |
| `backend/internal/domain/tokenusage` | 产生模型 token usage evidence |
| `backend/internal/domain/billing/usage.go` | 仅作通用 usage 参考，不充当市场账本 |
| `backend/internal/domain/admin/audit_log.go` | 不复用，市场拥有独立 audit |
| `backend/pkg/slugkit` | 新 identifier 统一校验 |

## 8. Legacy Market Migration

1. 将 `expert/marketplace.go` 三个硬编码应用导入默认市场的 Publisher、
   CatalogItem、CatalogItemVersion、Listing 和 ListingVersion。
2. 使用新 Preflight/Apply 在测试组织验证安装和补偿。
3. 将 `/marketplace` 域名路由切换到 Marketplace Web。
4. 将旧页面改为一次性明确跳转，不代理读写。
5. 删除 `PublicMarketHandler`、旧 install route、`public-market-api.ts` 和卡片直装。
6. 保留组织 Skill 管理页，它是 Runtime 管理界面。

迁移不双写。失败通过 GitOps 回滚服务版本和路由，不恢复硬编码目录写路径。

## 9. Error and Observability

Runtime Bridge 稳定错误至少包括：

```text
RESOURCE_NOT_PUBLISHABLE
ORGANIZATION_ACCESS_DENIED
PLAN_EXPIRED
PLAN_DIGEST_MISMATCH
RESOURCE_REVISION_CHANGED
DEPENDENCY_INCOMPATIBLE
RUNTIME_CAPACITY_UNAVAILABLE
INSTALLATION_CONFLICT
COMPENSATION_PENDING
```

Trace attributes：marketplace installation、operation、listing version、platform
organization、runtime resource 和 usage event。指标覆盖 preflight 阻塞率、安装
成功率、补偿时长、usage 延迟、重复事件和投递 backlog。

## 10. Verification

- Connect contract tests 覆盖 service identity、plan digest、过期和 revision drift。
- Adapter tests 覆盖 Expert、pinned Skill、MCP secret boundary 和 Resource binding。
- PostgreSQL tests 覆盖 bridge 唯一性、operation 幂等和补偿 ownership。
- E2E 覆盖直接启用、审批启用、依赖失败、验证失败、升级权限扩大和卸载。
- Deployment checks 覆盖 mTLS、outbox worker、路由切换和回滚。
