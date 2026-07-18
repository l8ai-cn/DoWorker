# Experts API

Expert 有两个明确合同：

| 路径 | 用途 | 资源关联 |
| --- | --- | --- |
| Resource Connect API | 产品 UI 与 typed 客户端 | 固定 Expert revision、WorkerTemplate、Prompt 和 WorkerSpec 快照 |
| External REST API | 现有 API-key 集成 | 直接操作历史 Expert 领域合同 |

两条路径不是字段兼容层。Resource Apply 失败时不会改走 External REST。

## Resource-native Expert

声明：

```yaml
apiVersion: agentsmesh.io/v1alpha1
kind: Expert
metadata:
  name: code-review-expert
  namespace: acme
  displayName: Code Review Expert
spec:
  workerTemplateRef:
    kind: WorkerTemplate
    name: codex-reviewer
  promptRef:
    kind: Prompt
    name: code-review
  description: Reviews changes and returns findings with evidence.
  category: engineering
  releaseNotes: Initial resource-native revision.
```

调用顺序：

```text
ValidateResource
PlanResource
ApplyExpertPlan
```

`ApplyExpertPlan` 返回 resource、`expert_id`、`worker_spec_snapshot_id` 和
`resource_revision`。Plan 固定 WorkerTemplate 与 Prompt 的 uid、revision 和
digest；Apply 后 Expert 投影使用固定 WorkerSpec 快照。

完整协议见[资源编排 API](orchestration-resources.md)。

## External REST

Base path：

```text
/api/v1/ext/orgs/{org_slug}/experts
```

### Scopes

| Scope | Access |
| --- | --- |
| `experts:read` | List and get experts |
| `experts:write` | Create, update, delete, run experts |
| `pods:read` / `pods:write` | Existing compatibility scopes accepted by this API |

### List

```http
GET /experts?limit=50&offset=0
```

### Get

```http
GET /experts/{slug}
```

### Create

```http
POST /experts
Content-Type: application/json
```

```json
{
  "name": "Code review assistant",
  "slug": "code-review-assistant",
  "agent_slug": "codex-cli",
  "runner_id": 1,
  "repository_id": 42,
  "branch_name": "main",
  "prompt": "Review pull requests for security issues.",
  "interaction_mode": "pty",
  "automation_level": "autonomous",
  "perpetual": false,
  "used_env_bundles": ["openai-default"],
  "skill_slugs": ["pdf-tool"],
  "knowledge_mounts": [{ "slug": "team-docs", "mode": "ro" }]
}
```

该请求不创建 `orchestration_resources` revision，也不经过 Resource Plan。

### Update

```http
PATCH /experts/{slug}
```

只提交需要修改的字段。若 Expert 已绑定
`orchestration_resource_id/revision`，接口返回 `409
EXPERT_MANAGED_BY_RESOURCE_APPLY`；应编辑 Expert 资源并执行
`Validate -> Plan -> ApplyExpertPlan`。

### Delete

```http
DELETE /experts/{slug}
```

资源托管 Expert 同样返回 `409 EXPERT_MANAGED_BY_RESOURCE_APPLY`，不能绕过
不可变 resource revision 删除领域投影。

### Run

```http
POST /experts/{slug}/run
Content-Type: application/json
```

```json
{
  "alias": "review-run-1",
  "prompt_override": "Focus on SQL injection this time.",
  "cols": 120,
  "rows": 40
}
```

只有绑定 WorkerSpec snapshot 的 Expert 可以运行。成功响应为 `201`，并返回
正在初始化的 Pod；缺少 snapshot 的历史草稿会要求重新发布或改用
resource-native Expert。Run 请求不能覆盖 Runner 或执行节点，实际执行目标由
该 Expert 绑定的 WorkerSpec snapshot 决定。

## Publish from Worker

Session-authenticated Worker 可以发布历史 Expert：

```http
POST /api/v1/orgs/{org_slug}/pods/{pod_key}/publish-expert
```

```json
{
  "name": "My expert",
  "slug": "my-expert",
  "skill_slugs": ["pdf-tool"],
  "knowledge_mounts": [{ "slug": "team-docs" }]
}
```

该接口复制 Worker 运行字段，不等同于 `ApplyExpertPlan`。需要 resource-native
Expert 时，应先声明并 Apply WorkerTemplate、Prompt 和 Expert。
