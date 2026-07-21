# Marketplace Catalog and Listing Data Design

- **Date:** 2026-07-11
- **Scope:** 发布方、目录资源、不可变版本、市场上架项

## 1. `marketplace_publishers`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | bigint | PK |
| `slug` | varchar(100) | UNIQUE identifier |
| `publisher_type` | varchar(16) | `user/organization/platform` |
| `platform_user_id` | bigint | user 类型必填 |
| `platform_org_id` | bigint | organization 类型必填 |
| `display_name` | varchar(120) | 必填 |
| `summary` | varchar(240) | 可空 |
| `logo_asset_key` | varchar(500) | 可空 |
| `verification_status` | varchar(20) | `unverified/pending/verified/revoked` |
| `verified_at` | timestamptz | 可空 |
| `created_at/updated_at` | timestamptz | 审计时间 |

类型与外部引用必须匹配，禁止同时填写 user 和 organization。

## 2. `marketplace_catalog_items`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | bigint | PK |
| `publisher_id` | bigint | FK |
| `slug` | varchar(100) | 发布方内唯一 |
| `resource_type` | varchar(20) | `application/skill/mcp_connector/resource` |
| `name` | varchar(120) | 原始名称 |
| `summary` | varchar(240) | 原始摘要 |
| `platform_resource_type` | varchar(40) | `expert/skill/mcp_item/ai_resource/...` |
| `platform_resource_id` | bigint | Agent Cloud 外部引用 |
| `status` | varchar(20) | `draft/active/deprecated/blocked` |
| `latest_version_id` | bigint | 可空 |
| `created_by_platform_user_id` | bigint | 创建人 |
| `created_at/updated_at` | timestamptz | 审计时间 |

唯一索引为 `(publisher_id, slug)` 和
`(platform_resource_type, platform_resource_id)`。

## 3. `marketplace_catalog_item_versions`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | bigint | PK |
| `catalog_item_id` | bigint | FK |
| `version` | varchar(50) | SemVer |
| `source_revision` | varchar(100) | Git SHA、resource revision 或 digest |
| `content_digest` | char(64) | SHA-256 |
| `manifest` | jsonb | 类型化安装清单 |
| `permissions` | jsonb | 权限和 MCP tool scopes |
| `compatibility` | jsonb | Worker、模型、区域和资源要求 |
| `dependency_lock` | jsonb | 固定版本和 digest |
| `artifact_key` | varchar(500) | 可空 |
| `validation_status` | varchar(20) | `pending/passed/failed/deprecated` |
| `created_by_platform_user_id` | bigint | 创建人 |
| `created_at` | timestamptz | 创建后不可修改 |

唯一索引为 `(catalog_item_id, version)` 和
`(catalog_item_id, content_digest)`。只有 `passed` 版本可以提交上架。

Manifest 公共字段：

```text
schema_version
resource_type
source_ref
outcomes[]
dependencies[]
permissions[]
compatibility
configuration_schema
verification
```

Application manifest 保存可发布 WorkerSpec 模板和固定依赖；Skill 保存 artifact
digest；MCP Connector 保存 transport、tool scopes 和凭证 schema；Resource 保存
资源类型、绑定模式和计量 meter。

## 4. `marketplace_listings`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | bigint | PK |
| `marketplace_id` | bigint | FK |
| `catalog_item_id` | bigint | FK |
| `slug` | varchar(100) | 市场内唯一 URL identifier |
| `status` | varchar(24) | `draft/submitted/validating/needs_changes/approved/published/suspended/deprecated/removed` |
| `visibility` | varchar(16) | `public/members/hidden` |
| `access_mode` | varchar(16) | `direct/approval/grant_only` |
| `current_version_id` | bigint | 当前公开 ListingVersion |
| `submitted_by_platform_user_id` | bigint | 可空 |
| `published_at/suspended_at` | timestamptz | 可空 |
| `created_at/updated_at` | timestamptz | 审计时间 |

唯一索引为 `(marketplace_id, slug)` 和
`(marketplace_id, catalog_item_id)`；不同市场拥有独立 Listing。

## 5. `marketplace_listing_versions`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | bigint | PK |
| `listing_id` | bigint | FK |
| `catalog_item_version_id` | bigint | FK |
| `revision` | integer | Listing 内递增 |
| `display_name` | varchar(120) | 市场展示名 |
| `tagline` | varchar(160) | 一句话价值 |
| `description` | text | 完整介绍 |
| `outcomes/use_cases` | jsonb | 字符串数组，单项 2-120 字 |
| `target_audience` | jsonb | 字符串数组 |
| `requirements` | jsonb | 外部账号、资源和前置条件 |
| `tags` | text[] | 最多 12 项 |
| `icon_asset_key/hero_asset_key` | varchar(500) | 可空 |
| `gallery` | jsonb | 最多 6 个受控媒体资源 |
| `documentation_url/support_url` | varchar(1000) | HTTPS |
| `quota_plan_id` | bigint | 可空 |
| `release_notes` | text | 版本变化 |
| `review_status` | varchar(20) | `draft/submitted/approved/rejected` |
| `created_at` | timestamptz | 批准后不可修改 |

唯一索引为 `(listing_id, revision)`。`current_version_id` 必须指向同一 Listing。

## 6. `marketplace_listing_spaces`

字段为 `listing_id bigint`、`space_id bigint`、`is_primary boolean`、
`sort_order integer`，主键为 `(listing_id, space_id)`。每个 Listing 必须且只能
有一个 primary Space，由该 Space Reviewer 负责审批。

## 7. Collections

### `marketplace_collections`

字段为 `id bigint`、`marketplace_id bigint`、`slug varchar(100)`、
`name varchar(120)`、`summary varchar(240)`、`status draft/published/hidden`、
`sort_order integer`、`created_by_platform_user_id bigint` 和审计时间。

### `marketplace_collection_listings`

主键为 `(collection_id, listing_id)`，另有 `sort_order integer`。Collection 只能
引用同一市场的 Listing；公开 Collection 至少包含一个 published Listing。

## 8. Review and Publication Rules

- `submitted` 后锁定本次 ListingVersion，修改必须生成新 revision。
- 自动验证覆盖 schema、digest、依赖、权限、干净组织安装和验收测试。
- Reviewer 只能审批自己负责 Space 的 Listing，发布方不得审核自己的提交。
- 权限扩大必须创建新 CatalogItemVersion 和 ListingVersion。
- 暂停保留详情、版本历史和现有安装，不允许新获取。
- 存在 Entitlement 或 Installation 的版本不得物理删除。
- 所有外部资源引用写入前通过 Runtime Bridge 校验。
