# Marketplace Storefront and Console API Contract

- **Date:** 2026-07-11
- **Scope:** 公共类型、Storefront、获取流程、Console write

## 1. Common Types

ID 按来源命名，不使用含糊的跨服务 `id`：

```json
{
  "marketplace_id": "42",
  "listing_id": "108",
  "installation_id": "7d296477-3b27-4bee-a381-bd61f86f71f7",
  "platform_organization_id": "9"
}
```

bigint 使用十进制字符串，时间使用 UTC RFC3339，decimal 使用字符串，枚举使用
小写 snake_case。

## 2. Error Envelope

```json
{
  "error": {
    "code": "QUOTA_INSUFFICIENT",
    "message": "市场额度不足",
    "detail": "需要 120.000000，可用 80.000000。",
    "field_issues": [
      {"field": "quota_account_id", "code": "insufficient"}
    ],
    "recovery_action": "view_quota",
    "request_id": "req_01J..."
  }
}
```

客户端只按 `code` 和 `recovery_action` 分支；`message/detail` 仅用于展示。

## 3. Listing Summary and Detail

`GET /markets/{marketSlug}/listings` item：

```json
{
  "listing_id": "108",
  "slug": "product-listing-optimizer",
  "resource_type": "application",
  "display_name": "商品 Listing 优化应用",
  "tagline": "生成符合目标平台规范的多语言商品详情",
  "publisher": {
    "slug": "commerce-lab",
    "display_name": "Commerce Lab",
    "verified": true
  },
  "spaces": [{"slug": "product-operations", "name": "商品运营"}],
  "quota": {"mode": "metered", "summary": "预计每次 5-20 额度"},
  "maintenance_status": "maintained",
  "published_at": "2026-07-11T08:00:00Z"
}
```

详情增加 `description`、`outcomes`、`use_cases`、`target_audience`、
`requirements`、`permissions`、`examples`、`verification`、`version`、
`release_notes`、`media`、`documentation_url` 和 `support_url`。

公开响应不包含 platform resource ID、secret 默认值或未发布版本。

## 4. Create Installation Plan

`POST /markets/{marketSlug}/listings/{listingSlug}/plans`

```json
{
  "listing_version_id": "301",
  "target_platform_organization_id": "9",
  "requested_configuration": {
    "repository_id": "77",
    "model_resource_id": "18"
  },
  "marketplace_group_id": null,
  "budget_limit_credits": "100.000000"
}
```

Response：

```json
{
  "installation_id": "7d296477-3b27-4bee-a381-bd61f86f71f7",
  "operation_id": "bbebd338-079d-43a1-8393-c9210adbd148",
  "plan": {
    "plan_id": "d7160a22-d578-4491-9a31-6c94af28fdd4",
    "plan_digest": "sha256:...",
    "expires_at": "2026-07-11T08:15:00Z",
    "resolved_resources": [],
    "required_permissions": [],
    "configuration_schema": {},
    "mutations": [],
    "estimated_usage": [],
    "blocking_issues": [],
    "warnings": []
  }
}
```

Requested configuration 只接受 Runtime secret reference，不接收明文 secret。

## 5. Apply and Operation

`POST /installation-operations/{operationId}/apply`

```json
{
  "plan_id": "d7160a22-d578-4491-9a31-6c94af28fdd4",
  "plan_digest": "sha256:...",
  "accepted_permission_keys": [
    "repository.write",
    "pull_request.create"
  ]
}
```

Header 必须包含 `Idempotency-Key`。Plan digest 已锁定 Listing 和 Runtime
revision；`If-Match` 仅用于 Console PATCH。成功返回 `202 Accepted`：

```json
{
  "operation_id": "bbebd338-079d-43a1-8393-c9210adbd148",
  "status": "running",
  "stage": "quota",
  "poll_after_ms": 1000
}
```

Operation 查询返回 status、stage、progress、result、error 和 compensation：

```json
{
  "status": "failed",
  "stage": "verify",
  "error": {
    "code": "POST_INSTALL_VERIFICATION_FAILED",
    "detail": "..."
  },
  "compensation": {
    "status": "succeeded",
    "released_credits": "20.000000"
  }
}
```

## 6. Acquisition Request

```json
{
  "listing_id": "108",
  "target_platform_organization_id": "9",
  "reason": "用于跨境电商运营团队商品发布"
}
```

响应包含 request ID、status、approver scope 和 expires_at。拒绝后不创建
Entitlement，审批命令必须携带 decision note。

## 7. Console Write Contract

创建 Marketplace：

```json
{
  "template_key": "cross-border-commerce",
  "name": "海贸通 AI 市场",
  "slug": "haitong-ai-market",
  "summary": "面向跨境电商团队的 AI 应用与系统连接市场",
  "visibility": "private",
  "registration_mode": "sso"
}
```

Listing draft 写入 catalog item version、展示内容、Space IDs、access mode、quota
plan、权限确认和支持信息。PATCH 使用 `If-Match`，响应返回 revision 和 ETag。

## 8. Compatibility

- API 不接受旧 MarketApplication slug 作为安装身份。
- API 不接受 `skill_slugs` 作为版本锁。
- Runtime resource ID 只出现在认证后的配置请求。
- 新增字段采用 additive versioning；语义变化发布新 API version。
