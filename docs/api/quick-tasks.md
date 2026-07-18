# Quick Task API

Quick Task 是已规划 Worker 的简化执行入口。它消费一个
`kind: Worker` 的有效 Plan，并复用 Worker Apply 的不可变 revision、
`worker_spec_snapshot_id`、持久 launch 和调度 outbox。

## 前置流程

1. 使用 `ValidateResource` 校验 Worker 资源。
2. 使用 `PlanResource` 创建 Worker Plan。
3. 将返回的 `plan_id` 提交给 Quick Task。

Worker 资源负责声明 `workerTemplateRef`、`promptRef`、`inputs` 和 `alias`。
模型、工具、Skill、KnowledgeBase、SecretReference、运行镜像、计算目标和权限
通过 WorkerTemplate 及其引用资源解析，不在 Quick Task 请求中重复声明。

## 创建

```http
POST /api/v1/orgs/{slug}/quick-tasks
Authorization: Bearer <jwt>
Content-Type: application/json

{
  "plan_id": "11111111-1111-4111-8111-111111111111"
}
```

`plan_id` 必须是当前 actor 在同一组织中创建、尚未过期并可应用的规范 UUID。
接口不接受 `agent_slug`、`runner_id`、`repository_id`、prompt、alias、
AgentFile 或 queue TTL 覆盖。

Apply 前会重新校验当前 actor 对 Worker 目标和每个固定 ResourceRef 的权限，
包括引用 revision 的 UID、revision 和 digest。权限撤销后不能依靠旧 Plan 启动。

## 响应

```json
{
  "pod_key": "7-standalone-12345678",
  "status": "queued",
  "queue_position": 3,
  "expires_at": "2026-07-17T08:30:00Z"
}
```

HTTP 状态为 `202 Accepted`。`status` 来自 Apply 后的当前 Pod；只有 Pod
确实处于 `queued` 时，响应才可能包含 `queue_position` 和 `expires_at`。
已消费 Plan 的幂等重放返回原 Worker launch 对应的当前 Pod，不创建第二个 Pod。

## 错误

| HTTP | code | 场景 |
| --- | --- | --- |
| 400 | `WORKER_PLAN_INVALID` | Plan ID 格式或 Worker Plan 内容无效 |
| 403 | `ACCESS_DENIED` | 当前 actor 无权读取或应用 Plan |
| 404 | `WORKER_PLAN_NOT_FOUND` | Plan 不存在于当前组织 |
| 409 | `WORKER_PLAN_STATE_CHANGED` | Plan stale、expired、consumed 冲突或资源基线变化 |
| 422 | `NO_RUNNER_FOR_AGENT` | snapshot 没有可用运行目标 |
| 429 | `QUEUE_FULL` | 已解析 Runner 的持久命令队列已满 |
| 503 | `WORKER_APPLY_UNAVAILABLE` | Worker Apply 控制面未就绪 |

错误不会回显 Worker YAML、Prompt 内容、Secret 值或解析后的凭证。
