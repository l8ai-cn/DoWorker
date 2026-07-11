# Marketplace Entitlement and Installation Design

- **Date:** 2026-07-11
- **Scope:** 获取申请、使用权限、安装计划、安装操作

## 1. `marketplace_acquisition_requests`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | uuid | PK，公开请求标识 |
| `marketplace_id` | bigint | FK |
| `listing_id` | bigint | FK |
| `requester_platform_user_id` | bigint | 申请人 |
| `target_platform_org_id` | bigint | 必填，显式目标组织 |
| `status` | varchar(16) | `pending/approved/rejected/cancelled/expired` |
| `reason` | varchar(500) | 可空 |
| `decision_note` | varchar(500) | 可空 |
| `decided_by_platform_user_id` | bigint | 可空 |
| `decided_at/expires_at` | timestamptz | 可空 |
| `created_at/updated_at` | timestamptz | 审计时间 |

同一用户、Listing、目标组织最多存在一个 pending 请求。`direct` 模式在权限
事务中直接创建 Entitlement，不生成伪审批记录。

## 2. `marketplace_entitlements`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | uuid | PK |
| `marketplace_id` | bigint | FK |
| `listing_id` | bigint | FK |
| `subject_type` | varchar(16) | `user/organization` |
| `subject_platform_id` | bigint | 外部用户或组织 ID |
| `target_platform_org_id` | bigint | 权限唯一适用的目标组织 |
| `status` | varchar(16) | `active/suspended/revoked/expired` |
| `source` | varchar(16) | `direct/approval/grant` |
| `source_request_id` | uuid | 可空 |
| `starts_at/expires_at` | timestamptz | 有效期 |
| `granted_by_platform_user_id` | bigint | 可空 |
| `created_at/updated_at` | timestamptz | 审计时间 |

有效唯一约束包含 Listing、subject、target org 和有效时间窗口。user entitlement
不能用于其他组织；撤销权限不直接删除运行资源。

## 3. `marketplace_installations`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | uuid | PK，跨服务 installation ID |
| `marketplace_id` | bigint | FK |
| `listing_id/listing_version_id` | bigint | 固定安装来源 |
| `entitlement_id` | uuid | FK |
| `target_platform_org_id` | bigint | 目标组织 |
| `quota_charge_scope` | varchar(16) | Plan 固化的扣费层级 |
| `quota_account_id` | uuid | 非 user scope 的固定账户，可空 |
| `installed_by_platform_user_id` | bigint | 操作人 |
| `status` | varchar(20) | `planning/installing/verifying/active/failed/suspended/uninstalled` |
| `runtime_ref` | varchar(200) | Runtime 返回的 opaque ref |
| `config_snapshot` | jsonb | 非密钥配置快照 |
| `plan_digest` | char(64) | 已确认计划摘要 |
| `current_operation_id` | uuid | 可空 |
| `last_verified_at` | timestamptz | 可空 |
| `created_at/updated_at` | timestamptz | 审计时间 |

同一 Listing 和目标组织可以有多个实例，但必须由 manifest 明确
`single_instance` 或 `multiple_instances`。

## 4. `marketplace_installation_operations`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | uuid | PK |
| `installation_id` | uuid | FK |
| `operation_type` | varchar(16) | `install/upgrade/suspend/resume/uninstall` |
| `idempotency_key` | uuid | UNIQUE |
| `status` | varchar(20) | `planned/running/succeeded/failed/compensating/compensated` |
| `stage` | varchar(40) | 确定性阶段 |
| `plan` | jsonb | Runtime preflight 不可变计划 |
| `result` | jsonb | 成功结果，可空 |
| `error_code` | varchar(80) | 可空 |
| `error_message` | varchar(500) | 可空，用户可展示 |
| `started_at/completed_at` | timestamptz | 可空 |
| `created_at` | timestamptz | 审计时间 |

阶段固定为 `entitlement/quota/runtime/dependencies/create/verify/settle`。重复
idempotency key 返回原操作。

## 5. Plan Contract

Plan 保存并返回：

```text
plan_id
plan_digest
expires_at
listing_version_id
target_platform_org_id
resolved_resources[]
required_permissions[]
required_configuration_schema
mutations[]
estimated_usage[]
blocking_issues[]
warnings[]
```

任何 ListingVersion、依赖 digest、组织资源 revision 或权限变化都会使 Plan
失效。Apply 不允许重新解析后静默执行。

## 6. State and Transaction Rules

- 用户必须显式选择目标组织；禁止自动选择第一个组织。
- 获取权限、创建 Installation、预占额度分别使用短事务。
- Apply 启动后通过 operation ID 查询，不依赖浏览器连接持续存在。
- Runtime 失败进入 `compensating`，补偿完成后才能释放额度。
- 部分依赖成功、主资源失败时必须清理本次创建的全部资源。
- 升级保留原版本、配置快照、Plan 和操作历史。
- 卸载不删除账本、审计和历史版本引用。
- 失败响应必须提供稳定 error code、stage、detail 和 recovery action。

## 7. Required Scenarios

- direct Listing 直接创建 Entitlement 并进入 Plan。
- approval Listing 无权限用户只能发起申请。
- grant_only Listing 只能由管理员发放权限。
- 额度不足在 Apply 前阻止，不创建 runtime 资源。
- Plan 过期或资源 revision 变化要求重新检查。
- 重复 Apply 返回原 operation，不重复安装。
- 补偿失败保持显式状态并告警，不伪装成已完成。
