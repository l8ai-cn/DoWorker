# Marketplace Service and API Architecture Design

- **Date:** 2026-07-11
- **Scope:** 独立服务、前后端目录、鉴权、Storefront/Console API

## 1. Deployment Boundary

新增三个部署单元：

```text
marketplace-api
clients/marketplace-web
clients/marketplace-console
```

Marketplace API 与 Backend 位于同一 Go module，但拥有独立 binary、数据库
schema、DB role、迁移、配置和发布流程。它禁止 import `backend/internal/**`，
只通过生成契约调用 Runtime Bridge。

初期共用 PostgreSQL cluster、Redis、对象存储和 Traefik。市场表位于
`marketplace` schema，数据库账号无权访问 runtime `public` schema。

## 2. Backend Structure

```text
marketplace/
  cmd/server/
  internal/api/public/
  internal/api/console/
  internal/api/internal/
  internal/domain/market/
  internal/domain/catalog/
  internal/domain/publishing/
  internal/domain/entitlement/
  internal/domain/installation/
  internal/domain/quota/
  internal/domain/audit/
  internal/service/
  internal/infra/postgres/
  internal/infra/redis/
  internal/infra/objectstorage/
  internal/integration/runtime/
  internal/integration/identity/
  migrations/
```

Domain 只定义实体、状态、错误、Repository interface 和纯规则。Service 编排
事务；infra 实现 Repository；API 只做 wire conversion、鉴权和错误映射。禁止
通用 `helpers.go`、`utils.go` 和大一统 service。

## 3. Frontend Structure

```text
clients/marketplace-web/src/
  app/(storefront)/
  app/(account)/
  features/market/
  features/catalog/
  features/acquisition/
  features/library/
  features/usage/
  lib/api/
  messages/{locale}/

clients/marketplace-console/src/
  app/(console)/
  features/market-settings/
  features/spaces/
  features/catalog/
  features/listings/
  features/publishing/
  features/members/
  features/quota/
  features/audit/
  lib/api/
  messages/{locale}/
```

两个应用使用 Next.js App Router、React 19、现有语义 token 和 next-intl，但不
import `clients/web/**` 或 `clients/web-admin/**`。Route 只组合 regions；feature
拆分 header、filters、table/list、detail/editor、dialogs 和状态组件。

Storefront 不加载 WASM。运行资源选择和执行通过 API 或显式跳转完成。

## 4. Authentication

当前 HS256 JWT 要求服务共享签名密钥，使 Marketplace API 具备伪造核心 Token
的能力。新服务上线前必须：

1. Core Auth 支持 OIDC Authorization Code + PKCE。
2. Core Auth 使用非对称签名并公开 JWKS。
3. Storefront 和 Console 注册为独立 OIDC client。
4. Marketplace API 只持有 issuer、audience 和 JWKS。

自定义域名先按 Host 解析 Marketplace，再发起登录。Callback state 绑定 market
ID、return path 和 PKCE verifier。禁止 query token。

Marketplace role 与 Runtime organization role 分别校验，任一不足都明确失败。

市场自定义 OIDC/SAML 由 Core Auth 的 Identity Integration API 创建和测试。
Marketplace API 只保存 provider ID 与 allowed domain，不接收 client secret、
证书或 SAML metadata。

## 5. Storefront API

版本前缀 `/api/marketplace/v1`：

```text
GET  /storefront/resolve
GET  /markets/{marketSlug}
GET  /markets/{marketSlug}/spaces
GET  /markets/{marketSlug}/listings
GET  /markets/{marketSlug}/listings/{listingSlug}
POST /markets/{marketSlug}/acquisition-requests
POST /markets/{marketSlug}/listings/{listingSlug}/plans
POST /installation-operations/{operationId}/apply
GET  /installation-operations/{operationId}
GET  /me/installations
GET  /me/quota
GET  /me/usage
```

公开接口以 Host 和 market slug 双重校验，防止跨自定义域读取。详情响应不返回
内部 platform resource ID、密钥 schema 默认值或未公开版本。

## 6. Console API

前缀 `/api/marketplace/v1/console/markets/{marketSlug}`，资源包括：

```text
market, brand, domains, members, spaces
publishers, catalog-items, catalog-item-versions
listings, listing-versions, submissions
quota-plans, quota-accounts, usage, audit-events
```

PATCH 携带 `If-Match` revision；冲突返回 `409 REVISION_CONFLICT`，禁止
last-write-wins。发布、暂停、审批、额度调整等命令使用显式 action endpoint。

## 7. API Conventions

- 创建和命令接口必须接受 `Idempotency-Key`。
- 大表使用 cursor pagination，响应为 `items`、`next_cursor`。
- 筛选、排序和 cursor 必须服务端执行。
- 时间统一 RFC3339 UTC；额度和 quantity 使用字符串 decimal。
- 错误包含 `code`、`message`、`detail`、`field_issues`、`request_id`。
- 写响应返回新 revision；读取支持 ETag。
- 上传只接受预签名对象存储流程和受控 MIME/大小。

## 8. Service Interfaces

主要 application services：

```text
MarketLifecycleService
DomainBindingService
MembershipService
SpaceService
CatalogRegistrationService
ListingPublishingService
AcquisitionService
InstallationOrchestrationService
QuotaService
UsageSettlementService
AuditQueryService
```

每个 service 依赖窄 Repository 和 integration port。跨域命令由 orchestration
service 调用，不允许 domain repository 互相访问。

Identity integration port 至少提供 CreateProvider、TestProvider、
EnableProvider 和 ResolveAuthenticatedUser；Runtime integration port 仅处理资源、
安装和运行权限，两者不得合并。

## 9. Configuration

Marketplace API 独立配置至少包含 HTTP 端口、DATABASE_URL、Redis、object
storage、OIDC issuer/audience、Runtime Bridge 地址、mTLS identity、public base
domain、domain verification 和 usage outbox 告警阈值。

所有配置进入 GitOps 模板；生产环境禁止默认 JWT secret、默认管理员和自动建表。

## 10. Verification

- Domain/repository 测试覆盖 identifier、状态机、不可变版本和租户隔离。
- PostgreSQL contract tests 覆盖 CHECK、UNIQUE、revision 和 cursor。
- API tests 覆盖 Host、role、ETag、idempotency 和错误码。
- 前端测试覆盖 URL 筛选、dirty、permission、conflict 和恢复。
- 浏览器 E2E 覆盖自定义域名、登录回跳、市场发布和移动端。
- 部署验证独立 DB role、OIDC/JWKS、mTLS 和健康检查。
