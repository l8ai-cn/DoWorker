# Worker Creation Contracts

`Worker` 是产品中的一次性 AI 执行单元。`Pod` 仍是后端和 Runner 的运行生命周期
对象。新建与恢复使用不同合同，不能做字段级互译或失败降级。

## 创建路径

| 路径 | 调用方 | 合同 | 结果 |
| --- | --- | --- | --- |
| Resource-native Worker | 产品 UI、typed Connect 客户端 | YAML/表单 -> Validate -> Plan -> `CreateWorkerFromPlan` | Resource revision、launch、WorkerSpec snapshot、Pod |
| Direct WorkerSpec | 内部 typed 客户端 | `ListWorkerCreateOptions`、`PreflightWorker`、`CreatePod` | WorkerSpec snapshot、Pod |
| External REST resume | API-key 集成 | `source_pod_key` lineage | 从不可变来源恢复的 Pod |

Resource Apply 失败时不会调用 Direct WorkerSpec 或 External REST。External REST
不提供 fresh Worker creation。

## Resource-native Worker

产品入口：

```text
/{org}/workers/new
```

“立即运行”编辑 `Worker`，“Worker 模板”编辑 `WorkerTemplate`，“引用资源”编辑
ModelBinding、Prompt 等依赖。

```yaml
apiVersion: agentcloud.io/v1alpha1
kind: Worker
metadata:
  name: release-review-20260715
  namespace: acme
spec:
  workerTemplateRef:
    kind: WorkerTemplate
    name: codex-reviewer
  promptRef:
    kind: Prompt
    name: delivery-review
  inputs:
    topic: release-42
  alias: Release Reviewer
```

Worker 是 create-only。再次执行应使用新的 `metadata.name`，不能更新已有 Worker
资源。

调用顺序：

```text
OrchestrationResourceService.ValidateResource
OrchestrationResourceService.PlanResource
OrchestrationResourceService.CreateWorkerFromPlan
```

Plan 固定 WorkerTemplate 和 Prompt revision，并把 Prompt 变量与 `inputs`
匹配。`CreateWorkerFromPlan` 返回：

```text
resource
launch_id
pod_id
pod_key
worker_spec_snapshot_id
resource_revision
runner_id
```

Worker launch 先持久化，再通过带 lease 的 claim 创建 Pod。plan、resource 和 Pod
存在唯一关联，调用重试不会创建第二个 Worker 资源或第二个 Pod 关联。

## WorkerTemplate

WorkerTemplate 保存可复用执行配置：

- Worker 类型和不可变 runtime image
- ModelBinding 与 ToolBinding
- ComputeTarget 和 ResourceProfile 或 custom resources
- Worker 类型 schema、配置值和 EnvironmentBundle Secret 引用
- Repository、Skill、KnowledgeBase 和配置包
- 交互模式、自动化级别和生命周期

`optionsRevision` 必须来自当前 Worker 创建选项。Plan 会重新执行正式 preflight，
生成 canonical WorkerSpec artifact；Apply 把它保存为不可变
`worker_spec_snapshots` 记录。

## Direct WorkerSpec

需要直接使用 typed WorkerSpec 的内部客户端可以调用：

```text
PodService.ListWorkerCreateOptions
PodService.PreflightWorker
PodService.CreatePod
```

选项请求返回 Worker 类型、runtime image、compute target、resource profile 和
schema revision。草稿包括：

```text
model_resource_id
worker_type_slug
runtime_image_id
compute_target_id
deployment_mode
resource_profile_id
type_schema_version
type_config_values
secret_refs
repository_id / branch
skill_ids / knowledge_mounts / env_bundle_ids
automation_level / interaction_mode
instructions / initial_task / alias
termination_policy / idle_timeout_minutes
options_revision
```

`PreflightWorker` 只有在没有 blocking issue 时才返回 resolved spec。`CreatePod`
会重新解析同一 draft，不接受客户端伪造或复用其他 revision 的 resolved spec。

Wire definitions：

```text
proto/pod/v1/worker_creation.proto
proto/pod/v1/pod.proto
```

## Runner MCP

Runner 暴露的 `create_pod` 工具只消费控制面已经生成的 Worker Plan：

```json
{
  "plan_id": "11111111-1111-4111-8111-111111111111"
}
```

工具 schema 不接受其他属性。Backend 会按发起 Pod 的组织和创建者身份重新检查
Plan 及其全部 ResourceRef 权限，再调用 Worker Apply。Runner 不能提交 snapshot
ID、Resource revision、Worker 类型、模型、Prompt、Ticket、仓库、权限、Secret、
AgentFile 或 placement 覆盖。创建成功后，Runner 仍会为发起 Pod 请求新 Pod 的
`pod:read` 和 `pod:write` 绑定。

## Secret

Worker 表单和 YAML 不传 API key：

- 模型凭据来自 ModelBinding 指向的 AI model resource connection。
- Worker 类型 Secret 通过 EnvironmentBundle 引用。
- 环境包不能覆盖模型资源管理的 credential 字段。
- Plan、Diff、导出和错误不回显 Secret。

## External REST

API-key 集成只能从已有 Pod 的不可变来源恢复：

```http
POST /api/v1/ext/orgs/{org_slug}/workers
X-API-Key: amk_...
Content-Type: application/json
```

```json
{
  "source_pod_key": "pod-abc123",
  "resume_agent_session": true,
  "cols": 120,
  "rows": 36
}
```

`source_pod_key` 必填。可选字段为 `resume_agent_session`、`ticket_slug`、
`cols`、`rows`、`queue_if_offline` 和 `queue_ttl_minutes`。Runner、Worker 类型、
模型、仓库、AgentFile、自动化级别、知识库或其他运行时覆盖会返回
`409 WORKER_RESUME_LINEAGE_ONLY`。

不带 `source_pod_key` 的请求会返回
`409 WORKER_RESOURCE_APPLY_REQUIRED`。fresh Worker 必须通过 typed Connect
执行 Validate、Plan 和 `CreateWorkerFromPlan`。`/pods` 与 `/workers` 共享相同
resume handler。

| Action | Method | Endpoint |
| --- | --- | --- |
| List | `GET` | `/workers` |
| Get | `GET` | `/workers/{pod_key}` |
| Resume | `POST` | `/workers` |
| Send prompt | `POST` | `/workers/{pod_key}/prompt` |
| Terminate | `POST` | `/workers/{pod_key}/terminate` |

## 选择建议

- 产品 UI 和新 typed 客户端使用 Resource-native Worker。
- 需要构造完整 WorkerSpec 的内部服务使用 Direct WorkerSpec。
- API-key 集成只使用 External REST 恢复已有 Worker lineage。

不要把 External REST resume 描述为 fresh Worker creation，也不要在 typed
Apply 失败后静默改用其他路径。
