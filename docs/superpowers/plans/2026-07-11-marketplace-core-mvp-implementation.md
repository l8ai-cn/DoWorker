# Marketplace Core MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> `subagent-driven-development` to implement this plan task-by-task.

**Goal:** 先交付独立 Marketplace API 的公开目录闭环，再扩展到可管理、可获取的
MVP。

**Architecture:** 新增根目录 `marketplace/` 独立 Go 服务、迁移和数据库访问层，
禁止导入 `backend/internal`。Core 只管理市场、专区、目录版本和 Listing；
安装计划与额度在 Core 稳定后通过明确接口加入。

**Tech Stack:** Go 1.25、Gin、GORM/PostgreSQL、golang-migrate、Testify、Next.js。

---

## Acceptance Scenarios

### Core

- Given Marketplace 数据库为空，When 执行迁移，Then 建立独立 schema、约束和索引。
- Given 市场存在已发布 Listing，When 匿名用户按市场 slug 查询，Then 只返回公开版本。
- Given 请求的 Host 不属于市场，When 查询详情，Then 返回 `MARKET_NOT_FOUND`。
- Given Listing 或 Space 属于其他市场，When 关联或查询，Then 不泄露跨租户数据。
- Given slug 非法，When 创建市场或 Listing，Then 返回字段错误且数据不落库。

### MVP

- Given 管理员创建市场和专区，When 发布满足条件的 Listing，Then Storefront 可见。
- Given 用户选择目标组织，When 创建安装计划，Then 返回不可变 plan digest。
- Given 额度不足，When Apply，Then 返回 `QUOTA_INSUFFICIENT` 且不调用 Runtime。
- Given安装成功，When 用户查看“我的应用”，Then 可看到 active installation。

## Task 1: Core Database Contract

**Files:**
- Create: `marketplace/migrations/000001_market_foundation.up.sql`
- Create: `marketplace/migrations/000001_market_foundation.down.sql`
- Create: `marketplace/migrations/000002_core_catalog.up.sql`
- Create: `marketplace/migrations/000002_core_catalog.down.sql`
- Create: `marketplace/migrations/embed.go`
- Test: `marketplace/migrations/core_catalog_test.go`

- [ ] 先写迁移静态契约测试，断言 schema、核心表、slug CHECK、租户唯一索引、
  published 状态约束和 down 顺序存在。
- [ ] 运行
  `go test ./marketplace/migrations -run TestCoreCatalogMigrationContract -count=1`，
  确认因迁移不存在而失败。
- [ ] 编写最小迁移，建立 marketplaces、domains、spaces、publishers、
  catalog_items、catalog_item_versions、listings、listing_versions 和
  listing_spaces。
- [ ] 再运行迁移测试并确认通过。

## Task 2: Domain Rules

**Files:**
- Create: `marketplace/internal/domain/market/market.go`
- Create: `marketplace/internal/domain/market/market_test.go`
- Create: `marketplace/internal/domain/catalog/catalog.go`
- Create: `marketplace/internal/domain/catalog/catalog_test.go`
- Create: `marketplace/internal/domain/listing/listing.go`
- Create: `marketplace/internal/domain/listing/listing_test.go`

- [ ] 先测试合法状态迁移、非法 slug、不可变版本和 published Listing 可见性。
- [ ] 运行 `go test ./marketplace/internal/domain/... -count=1` 并确认 RED。
- [ ] 使用 `backend/pkg/slugkit` 实现 identifier 校验和最小状态规则。
- [ ] 再运行 domain tests，确认全部 GREEN。

## Task 3: PostgreSQL Repositories

**Files:**
- Create: `marketplace/internal/infra/postgres/database.go`
- Create: `marketplace/internal/infra/postgres/market_repository.go`
- Create: `marketplace/internal/infra/postgres/catalog_repository.go`
- Create: `marketplace/internal/infra/postgres/listing_repository.go`
- Test: `marketplace/internal/infra/postgres/repository_test.go`

- [ ] 先用 SQLite transaction 测试 repository 的市场隔离、published 过滤和分页。
- [ ] 运行 repository test，确认接口未实现导致失败。
- [ ] 实现窄 Repository，所有查询显式包含 marketplace ID。
- [ ] 运行 repository tests；配置 PostgreSQL DSN 时再运行真实约束测试。

## Task 4: Independent API Process

**Files:**
- Create: `marketplace/internal/config/config.go`
- Create: `marketplace/internal/api/router.go`
- Create: `marketplace/internal/api/error_response.go`
- Create: `marketplace/internal/api/health_handler.go`
- Create: `marketplace/cmd/server/main.go`
- Test: `marketplace/internal/api/health_handler_test.go`

- [ ] 先测试 `/health/live`、`/health/ready` 和统一错误 envelope。
- [ ] 运行 API test，确认路由不存在。
- [ ] 实现独立配置、数据库连接、Gin router 和优雅关闭。
- [ ] 运行 `go test ./marketplace/internal/api ./marketplace/cmd/server -count=1`。

## Task 5: Public Storefront API

**Files:**
- Create: `marketplace/internal/service/storefront_service.go`
- Create: `marketplace/internal/api/public/market_handler.go`
- Create: `marketplace/internal/api/public/listing_handler.go`
- Test: `marketplace/internal/api/public/storefront_handler_test.go`

- [ ] 先测试市场信息、Listing 列表、详情、Host 双重校验和未发布数据隐藏。
- [ ] 运行 public API tests，确认 RED。
- [ ] 实现 `GET /markets/:marketSlug`、`/spaces`、`/listings` 和详情。
- [ ] 运行 API 与 domain 全量测试，确认 GREEN。

## Task 6: Minimal Console Write API

**Files:**
- Create: `marketplace/internal/service/market_lifecycle_service.go`
- Create: `marketplace/internal/service/listing_publishing_service.go`
- Create: `marketplace/internal/api/console/market_handler.go`
- Create: `marketplace/internal/api/console/listing_handler.go`
- Test: `marketplace/internal/api/console/console_handler_test.go`

- [ ] 先测试创建市场、创建专区、注册目录版本、创建 Listing 和发布命令。
- [ ] 测试 revision 冲突返回 `REVISION_CONFLICT`。
- [ ] 实现最小 Console API；认证先使用可替换 Actor middleware，不接共享 HS256。
- [ ] 运行 Console、Storefront 和 repository tests。

## Task 7: Storefront MVP UI

**Files:**
- Create: `clients/marketplace-web/package.json`
- Create: `clients/marketplace-web/src/app/layout.tsx`
- Create: `clients/marketplace-web/src/app/page.tsx`
- Create: `clients/marketplace-web/src/features/catalog/market-home.tsx`
- Create: `clients/marketplace-web/src/features/catalog/listing-card.tsx`
- Create: `clients/marketplace-web/src/lib/api/storefront.ts`
- Test: `clients/marketplace-web/src/features/catalog/market-home.test.tsx`

- [ ] 激活 `frontend-design-skill`，先写加载、空、错误和列表状态测试。
- [ ] 实现中文市场首页、专区筛选、搜索和 Listing 卡片，不加载 WASM。
- [ ] 执行 lint、typecheck、Vitest 和桌面/移动浏览器验证。

## Task 8: Installation and Quota Slice

**Files:**
- Create: `marketplace/migrations/000002_entitlement_installation_quota.up.sql`
- Create: `marketplace/internal/domain/installation/installation.go`
- Create: `marketplace/internal/domain/quota/account.go`
- Create: `marketplace/internal/service/installation_orchestration_service.go`
- Test: corresponding domain and service tests

- [ ] 先测试 plan digest、幂等 Apply、额度预占、Usage debit 和 completion release。
- [ ] 实现 entitlement、installation、operation、quota account 和 ledger 最小表。
- [ ] Runtime Bridge 使用 interface fake 验证，真实 ConnectRPC 契约单独提交。
- [ ] 运行 Marketplace 全量测试，确认无重复扣费和跨用户归因。

## Delivery Gates

- [ ] 所有新增非测试文件不超过 200 行。
- [ ] `go test ./marketplace/...` 通过。
- [ ] PostgreSQL migration up/down contract 通过。
- [ ] Storefront 真实浏览器覆盖成功、空、错、移动端状态。
- [ ] Docker/GitOps、健康检查和回滚命令进入仓库。
- [ ] 独立代码审查确认无跨库导入、静默 fallback 或旧市场双写。
