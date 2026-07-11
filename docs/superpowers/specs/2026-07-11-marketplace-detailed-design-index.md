# Marketplace Detailed Design Index

- **Date:** 2026-07-11
- **Parent:** `2026-07-11-multi-tenant-industry-marketplace-platform-design.md`
- **Status:** Product and architecture design; implementation has not started

## Documents

| Document | Scope |
| --- | --- |
| `2026-07-11-marketplace-tenancy-space-data-design.md` | 市场租户、品牌、域名、成员、专区 |
| `2026-07-11-marketplace-catalog-listing-data-design.md` | 发布方、目录资源、版本、上架项 |
| `2026-07-11-marketplace-entitlement-installation-design.md` | 获取申请、使用权限、安装事务 |
| `2026-07-11-marketplace-quota-ledger-design.md` | 额度方案、计量、预占、账本、用量 |
| `2026-07-11-marketplace-storefront-interaction-design.md` | 市场前台页面、获取流程、状态文案 |
| `2026-07-11-marketplace-console-copy-design.md` | 管理台页面、字段、中文文案 |
| `2026-07-11-marketplace-service-api-architecture-design.md` | 独立服务、前后端代码结构、API、鉴权 |
| `2026-07-11-marketplace-public-api-contract-design.md` | Storefront、获取流程、Console 数据契约 |
| `2026-07-11-marketplace-internal-api-contract-design.md` | Runtime Bridge、预占和 Usage 数据契约 |
| `2026-07-11-marketplace-runtime-bridge-migration-design.md` | 现有 Expert/Skill/MCP 接入、事件和迁移 |

## Shared Decisions

- URL identifier 使用 `slugkit`，展示名称不得用于 lookup。
- Marketplace API 独占市场数据，不跨库读取 Backend 内部表。
- Storefront、Console 和 Marketplace API 分别部署。
- 市场版本、目录资源版本、上架版本和安装计划均不可变。
- 用户必须显式选择目标组织，禁止自动选择第一个组织。
- 安装使用 `Plan -> Apply -> Verify -> Settle`，部分成功不是成功。
- 市场额度、模型 Token 和 API 访问令牌是三个不同概念。
- 所有失败返回稳定错误码、具体原因和可执行恢复动作。
- 旧硬编码市场在新路径验证后删除，不做双写或静默 fallback。
