# Marketplace Quota and Ledger Design

- **Date:** 2026-07-11
- **Scope:** 额度方案、meter、账户、预占、账本、usage、审计

## 1. Concepts

- 访问令牌用于 API 认证。
- 模型 Token 是模型服务商计量单位。
- 市场额度是用户可分配和消费的统一 Credit。
- 现有 `TokenQuota` 是报告型 token ceiling，不能作为市场账本。

## 2. `marketplace_quota_plans`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | bigint | PK |
| `marketplace_id` | bigint | FK |
| `slug` | varchar(100) | 市场内唯一 |
| `name` | varchar(100) | 展示名 |
| `description` | varchar(500) | 可空 |
| `period` | varchar(16) | `monthly/total` |
| `grant_credits` | numeric(20,6) | 每周期额度 |
| `charge_scope` | varchar(16) | `marketplace/organization/group/user` |
| `renewal_day` | smallint | monthly 时 1-28 |
| `status` | varchar(16) | `draft/active/retired` |
| `created_at/updated_at` | timestamptz | 审计时间 |

## 3. Meter and Rate

### `marketplace_meter_definitions`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | bigint | PK |
| `marketplace_id` | bigint | FK |
| `key` | varchar(100) | 市场内唯一 identifier |
| `display_name` | varchar(100) | 如“模型 Token” |
| `unit` | varchar(40) | `token/second/call/gb_day/run` |
| `aggregation` | varchar(16) | `sum/count/max` |
| `status` | varchar(16) | `active/disabled` |

### `marketplace_quota_rates`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | bigint | PK |
| `quota_plan_id` | bigint | FK |
| `listing_id` | bigint | 可空，空表示默认 |
| `meter_definition_id` | bigint | FK |
| `credits_per_unit` | numeric(20,9) | 大于等于 0 |
| `minimum_credits` | numeric(20,6) | 单次最低扣减 |
| `rounding_scale` | smallint | 0-9 |

Listing 费率优先于方案默认。找不到费率时阻止计量型运行，不能按 0 执行。

## 4. `marketplace_quota_accounts`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | uuid | PK |
| `marketplace_id` | bigint | FK |
| `subject_type` | varchar(16) | `marketplace/organization/group/user` |
| `subject_ref` | varchar(100) | 外部 ID 或内部 group ID |
| `quota_plan_id` | bigint | FK |
| `status` | varchar(16) | `active/suspended/closed` |
| `period_start/period_end` | timestamptz | 当前周期 |
| `created_at/updated_at` | timestamptz | 审计时间 |

唯一索引为 `(marketplace_id, subject_type, subject_ref, quota_plan_id)`。

## 5. Reservation and Ledger

### `marketplace_quota_reservations`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | uuid | PK |
| `quota_account_id` | uuid | FK |
| `reservation_type` | varchar(20) | `installation/runtime_execution` |
| `subject_ref` | varchar(100) | operation ID 或 runtime execution ID |
| `idempotency_key` | uuid | UNIQUE |
| `reserved_credits` | numeric(20,6) | 大于 0 |
| `status` | varchar(16) | `held/settled/released/expired` |
| `expires_at` | timestamptz | 必填 |
| `created_at/updated_at` | timestamptz | 审计时间 |

### `marketplace_quota_ledger_entries`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | uuid | PK |
| `quota_account_id` | uuid | FK |
| `entry_type` | varchar(20) | `grant/reserve/debit/release/adjust/grant_expire` |
| `available_delta` | numeric(20,6) | 可正可负 |
| `reserved_delta` | numeric(20,6) | 可正可负 |
| `consumed_delta` | numeric(20,6) | 仅 debit 为正 |
| `shortfall_delta` | numeric(20,6) | 未被额度覆盖的实际消费 |
| `reservation_id/usage_event_id/operation_id` | uuid | 可空 |
| `reason` | varchar(240) | 必填 |
| `created_by_platform_user_id` | bigint | 系统事件可空 |
| `created_at` | timestamptz | 创建后不可修改 |

余额由 delta 聚合。任何事务后 available 和 reserved 不得小于 0。

## 6. `marketplace_usage_events`

| Field | Type | Constraint / Meaning |
| --- | --- | --- |
| `id` | uuid | PK，Runtime 幂等键 |
| `marketplace_id` | bigint | FK |
| `installation_id` | uuid | FK |
| `listing_id` | bigint | 固定审计引用 |
| `reservation_id` | uuid | 可空，FK |
| `platform_org_id/platform_user_id` | bigint | 使用主体 |
| `meter_key` | varchar(100) | MeterDefinition key |
| `quantity` | numeric(24,9) | 大于等于 0 |
| `occurred_at` | timestamptz | Runtime 事件时间 |
| `source` | varchar(32) | `worker/model/mcp/storage/application` |
| `metadata` | jsonb | 非敏感证据 |
| `status` | varchar(16) | `accepted/rejected/settled` |
| `rejection_code` | varchar(80) | 可空 |
| `received_at/settled_at` | timestamptz | 可空 |

Usage 和 ledger 在同一事务写入；重复 event ID 返回首次结果。

## 7. Ledger Operations

```text
grant:   available +N
reserve: available -N, reserved +N
debit:   reserved -min(actual, reserved), then available, consumed +actual
release: reserved -N, available +N
grant_expire: available -remaining
adjust:  explicit signed delta with reason and actor
```

UsageEvent 到达时执行 debit；completion 只 release 剩余 reserved，不再次扣款。
actual 大于剩余 reservation 时原子扣减 available。仍不足时 available 保持 0，
`consumed_delta` 记录全部实际用量，未覆盖部分写入 `shortfall_delta`，账户进入
suspended 并阻止新运行，直到 grant 或管理员 adjustment 清偿 shortfall。
不得丢弃、缩小或篡改实际用量。

Reservation 到期使用 release 规则并记录 reason=`reservation_expired`；
`grant_expire` 只用于周期额度到期。

## 8. Account Resolution

QuotaPlan 的 `charge_scope` 决定层级：organization、marketplace 和 group 在
Plan 时解析并固化 account ID；group 必须显式选择并验证成员关系。user scope
只固化层级，每次 Runtime execution 按实际 `platform_user_id` 解析用户账户，
不能永久绑定安装者。

不存在匹配 active account 时返回 `QUOTA_ACCOUNT_NOT_FOUND`，禁止向上级账户
隐式回退或拆分扣费。

## 9. Renewal and Audit

月度发放使用 `(quota_account_id, period_start, entry_type=grant)` 唯一键，重复
任务不得重复发放。关闭账户不会删除历史账本。

`marketplace_audit_events` 记录额度方案、费率、人工调整、续期失败、短缺处理和
安全冻结。字段包括 actor、action、target、old/new data、IP、UA 和时间。

## 10. Enforcement

- Preflight 返回预估额度，不扣款。
- Apply 原子预占；不足时返回 required 和 available。
- 每次应用运行在 Runtime 启动前独立预占，不复用安装 reservation。
- Runtime usage 通过 outbox 异步投递，不能丢弃。
- 每个 UsageEvent 原子扣减 reservation；完成事件只释放剩余额度。
- 运行不能静默切换低价模型或 meter。
- 用户界面分别展示可用、已预占、已消费和周期结束时间。
