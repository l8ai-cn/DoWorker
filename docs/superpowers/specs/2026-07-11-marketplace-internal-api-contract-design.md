# Marketplace Runtime and Usage Internal API Contract

- **Date:** 2026-07-11
- **Scope:** Runtime Bridge、额度预占、UsageEvent

## 1. Runtime Bridge Preflight

Marketplace API 调用 Runtime Bridge：

```text
marketplace_installation_id
listing_version_ref
catalog_item_type
manifest_json
target_organization_id
actor_user_id
requested_bindings_json
idempotency_key
```

Runtime 返回：

```text
plan_id
plan_digest
runtime_revision
expires_at
resolved_resources[]
required_permissions[]
configuration_schema
mutations[]
estimated_meters[]
blocking_issues[]
warnings[]
```

Apply 重复传递 plan ID/digest 和 idempotency key，并返回 opaque runtime
installation ref、created resources 和 verification evidence。Marketplace API
不解析内部 resource ref。

## 2. Runtime Execution Reservation

`POST /api/marketplace/v1/internal/quota-reservations`

```json
{
  "event_id": "fdb63db4-32f9-4b99-9ed3-727bd13c81fb",
  "installation_id": "7d296477-3b27-4bee-a381-bd61f86f71f7",
  "platform_user_id": "14",
  "runtime_execution_ref": "pod:delivery-42",
  "estimated_meters": [
    {"meter_key": "model_tokens", "quantity": "10000"}
  ],
  "budget_limit_credits": "30.000000"
}
```

Response：

```json
{
  "reservation_id": "71debd22-edbe-4494-a775-8470cebb0215",
  "reserved_credits": "30.000000",
  "expires_at": "2026-07-11T10:00:00Z"
}
```

event ID 全局幂等。额度不足返回 `QUOTA_INSUFFICIENT`，Runtime 不启动执行。

## 3. Usage Event

`POST /api/marketplace/v1/internal/usage-events`

```json
{
  "event_id": "557164b3-c7bb-44dc-9976-5c98a663687a",
  "reservation_id": "71debd22-edbe-4494-a775-8470cebb0215",
  "installation_id": "7d296477-3b27-4bee-a381-bd61f86f71f7",
  "platform_organization_id": "9",
  "platform_user_id": "14",
  "meter_key": "model_tokens",
  "quantity": "3840",
  "source": "model",
  "occurred_at": "2026-07-11T08:04:12Z",
  "metadata": {"model": "provider/model"}
}
```

Usage event ID 全局幂等。accepted 事务同时写 UsageEvent 和本次 debit ledger；
metadata 禁止 prompt、模型响应、credential、header 和环境变量。响应包含本次
debited credits、reservation remaining、shortfall 和 account status。

## 4. Completion and Release

`POST /api/marketplace/v1/internal/quota-reservations/{id}/complete`

```json
{
  "completion_event_id": "8ef19427-c434-4476-9d7b-6b677e7106d9",
  "completed_at": "2026-07-11T08:10:00Z"
}
```

完成操作不重复 debit，只释放剩余 reserved，并返回 consumed、released 和 status。

`POST /api/marketplace/v1/internal/quota-reservations/{id}/release`

```json
{
  "release_event_id": "f11d529a-a0ae-4b40-9515-374651441bd1",
  "reason": "runtime_start_failed"
}
```

未产生使用的失败执行调用 release；release event ID 全局幂等。

## 5. Security and Delivery

- Internal API 只接受 Runtime Bridge mTLS identity。
- 请求同时校验 installation 与 platform organization 的归属。
- Runtime 先写本地 outbox，再投递 UsageEvent 和 completion。
- 投递失败保持 outbox pending 并告警，不能丢弃。
- Marketplace 返回稳定 accepted/rejected code；重试返回首次处理结果。
- Internal API 不接受浏览器 JWT，不暴露到公网 ingress。
