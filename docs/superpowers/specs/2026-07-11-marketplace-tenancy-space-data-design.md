# Marketplace Tenancy and Space Data Design

- **Date:** 2026-07-11
- **Scope:** 市场、品牌、域名、成员、专区

## 1. Shared Rules

- 主键使用 `BIGSERIAL`；跨服务命令和事件使用 UUID。
- slug 满足 `slugkit` 规则，长度 2-100，创建后不可修改。
- `platform_user_id`、`platform_organization_id` 是外部引用，不建跨库外键。
- 所有市场内唯一索引包含 `marketplace_id`。
- Marketplace API 独占写入，Agent Cloud Backend 不直接查询这些表。

## 2. `marketplaces`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | bigint | PK |
| `slug` | varchar(100) | UNIQUE identifier |
| `name` | varchar(120) | 必填展示名 |
| `summary` | varchar(240) | 必填，一句话定位 |
| `description` | text | 完整介绍 |
| `status` | varchar(24) | `draft/configuring/review/published/suspended/archived` |
| `visibility` | varchar(16) | `public/private` |
| `template_key` | varchar(50) | `blank/cross-border-commerce/higher-education/enterprise` |
| `default_locale` | varchar(16) | 默认 `zh-CN` |
| `registration_mode` | varchar(16) | `public/invite/sso` |
| `owner_platform_org_id` | bigint | 市场所有者组织 |
| `default_quota_plan_id` | bigint | 可空，发布前必填 |
| `created_by_platform_user_id` | bigint | 创建人 |
| `published_at/suspended_at` | timestamptz | 可空 |
| `created_at/updated_at` | timestamptz | 审计时间 |

状态只能沿产品状态机迁移。发布前必须存在主域名、已发布 Space、已发布
Listing 和有效额度方案。

## 3. `marketplace_brand_configs`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `marketplace_id` | bigint | PK + FK |
| `brand_name` | varchar(120) | 页头品牌名 |
| `logo_asset_key` | varchar(500) | 对象存储键 |
| `favicon_asset_key` | varchar(500) | 对象存储键 |
| `hero_asset_key` | varchar(500) | 可空 |
| `primary_color` | varchar(7) | `#RRGGBB`，发布时校验 AA |
| `layout_preset` | varchar(24) | `catalog/editorial/institution` |
| `homepage_config` | jsonb | 版本化区块配置 |
| `updated_by_platform_user_id` | bigint | 最后修改人 |
| `updated_at` | timestamptz | 最后修改时间 |

禁止任意 CSS、脚本或 HTML。`homepage_config` 只允许 `hero`、`space_grid`、
`featured_collection` 和 `latest_listings` 等服务端 schema。

## 4. `marketplace_domains`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | bigint | PK |
| `marketplace_id` | bigint | FK |
| `host` | varchar(253) | UNIQUE，小写 punycode |
| `kind` | varchar(16) | `platform/custom` |
| `status` | varchar(20) | `pending/verifying/active/failed/disabled` |
| `verification_token` | varchar(100) | DNS 验证随机值 |
| `is_primary` | boolean | 一个市场最多一个 true |
| `verified_at` | timestamptz | 可空 |
| `last_error_code` | varchar(80) | 可空 |
| `created_at/updated_at` | timestamptz | 审计时间 |

自定义域名必须完成所有权验证和 TLS 健康检查后才能设为主域名。

## 5. Identity and Invitations

### `marketplace_identity_providers`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | bigint | PK |
| `marketplace_id` | bigint | FK |
| `platform_identity_provider_id` | bigint | Core Auth 外部引用 |
| `provider_type` | varchar(16) | `oidc/saml` |
| `name` | varchar(100) | 展示名 |
| `allowed_domains` | text[] | 可空 |
| `status` | varchar(16) | Core Auth 状态镜像 |
| `last_synced_at` | timestamptz | 可空 |
| `created_at/updated_at` | timestamptz | 审计时间 |

Issuer、client secret、证书和 SAML metadata 由 Core Auth 管理；Marketplace
只保存 provider 引用和市场准入策略。

### `marketplace_member_invitations`

字段为 `id uuid`、`marketplace_id bigint`、`email varchar(320)`、
`role varchar(20)`、`token_digest char(64)`、`status varchar(16)`、
`invited_by_platform_user_id bigint`、`expires_at/accepted_at/created_at`。
同一 market/email 最多一个 active invitation，数据库只保存 token digest。

## 6. `marketplace_members`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | bigint | PK |
| `marketplace_id` | bigint | FK |
| `platform_user_id` | bigint | 外部用户引用 |
| `role` | varchar(20) | `owner/admin/quota_admin/publisher/member` |
| `status` | varchar(16) | `invited/active/suspended/removed` |
| `invited_by_platform_user_id` | bigint | 可空 |
| `joined_at` | timestamptz | 可空 |
| `created_at/updated_at` | timestamptz | 审计时间 |

唯一索引为 `(marketplace_id, platform_user_id)`。任何操作后必须至少保留一个
active owner。

## 7. Groups

### `marketplace_groups`

字段为 `id bigint`、`marketplace_id bigint`、`slug varchar(100)`、
`name varchar(120)`、`group_type varchar(20)`、`status varchar(16)`、
`created_at/updated_at`。`group_type` 为 `department/class/cohort/team/custom`。

### `marketplace_group_members`

主键为 `(group_id, platform_user_id)`，另有 `role member/manager` 和
`created_at`。额度账户可以绑定 group，但成员身份仍来自 Marketplace Member。

## 8. `marketplace_spaces`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | bigint | PK |
| `marketplace_id` | bigint | FK |
| `slug` | varchar(100) | 市场内唯一 |
| `name` | varchar(80) | 用户界面称“专区” |
| `summary` | varchar(240) | 卡片摘要 |
| `description` | text | 专区介绍 |
| `icon_asset_key` | varchar(500) | 可空 |
| `status` | varchar(16) | `draft/published/hidden/archived` |
| `sort_order` | integer | 默认 0 |
| `created_by_platform_user_id` | bigint | 创建人 |
| `published_at` | timestamptz | 可空 |
| `created_at/updated_at` | timestamptz | 审计时间 |

同一市场唯一索引为 `(marketplace_id, slug)`。Space 可以先发布再上架 Listing，
避免首次发布形成循环依赖；隐藏 Space 不出现在导航。

## 9. `marketplace_space_members`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `space_id` | bigint | FK |
| `platform_user_id` | bigint | 外部用户引用 |
| `role` | varchar(16) | `maintainer/reviewer` |
| `created_at` | timestamptz | 审计时间 |

主键为 `(space_id, platform_user_id, role)`。Reviewer 可以审核但不能修改 Space；
Maintainer 可以维护 Space 和审核其 Listing。

## 10. Tenant Invariants

- 每个查询、缓存 key、对象存储路径和审计事件必须携带 `marketplace_id`。
- Host 只能解析到一个 active Marketplace，无法解析时返回 `MARKET_NOT_FOUND`。
- 市场暂停后保留成员和已启用实例读取，阻止新获取、发布和安装。
- 删除采用状态迁移；存在 Listing、Entitlement 或 Installation 时禁止物理删除。
