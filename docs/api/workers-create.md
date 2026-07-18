# Worker Creation Contracts

`Worker` 是产品中的一次性 AI 执行单元。`Pod` 仍是后端和 Runner 的运行生命周期
对象。当前存在三条不同合同，不能做字段级互译或失败降级。

## 创建路径

| 路径 | 调用方 | 合同 | 结果 |
| --- | --- | --- | --- |
| Resource-native Worker | 产品 UI、typed Connect 客户端 | YAML/表单 -> Validate -> Plan -> `CreateWorkerFromPlan` | Resource revision、launch、WorkerSpec snapshot、Pod |
| Direct WorkerSpec | 内部产品客户端 | `ListWorkerCreateOptions`、`PreflightWorker`、`CreatePod` | WorkerSpec snapshot、Pod |
| External REST | 现有 API-key 集成 | legacy `agent_slug` 与 `agentfile_layer` | Legacy Pod |

Resource Apply 失败时不会调用 Direct WorkerSpec 或 External REST。

## Resource-native Worker

产品入口：

```text
/{org}/workers/new
```

“立即运行”编辑 `Worker`，“Worker 模板”编辑 `WorkerTemplate`，“引用资源”编辑
ModelBinding、Prompt 等依赖。

```yaml
apiVersion: agentsmesh.io/v1alpha1
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
- Repository、Skill、KnowledgeBase、运行时环境包和配置文档绑定
- 交互模式、自动化级别和生命周期

`optionsRevision` 必须来自当前 Worker 创建选项。Plan 会重新执行正式 preflight，
生成 canonical WorkerSpec artifact；Apply 把它保存为不可变
`worker_spec_snapshots` 记录。

Worker 创建选项中的 `config_document_requirements` 声明 `document_id`、格式、
目标路径和 `required`。WorkerTemplate 通过
`workspace.configDocumentBindings` 把文档 ID 绑定到 config 类型的
EnvironmentBundle。必需文档必须绑定；可选文档未配置时应省略，不发送空引用。

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
config_document_bindings
automation_level / interaction_mode
instructions / initial_task / alias
termination_policy / idle_timeout_minutes
options_revision
```

`PreflightWorker` 只有在没有 blocking issue 时才返回 resolved spec。`CreatePod`
会重新解析同一 draft，不接受客户端伪造或复用其他 revision 的 resolved spec。
`config_document_bindings` 的 `document_id` 必须由当前 Worker Definition 声明，
`config_bundle_id` 必须指向当前 actor 可读取、与 Worker 类型兼容的 config
EnvironmentBundle。

Wire definitions：

```text
proto/pod/v1/worker_creation.proto
proto/pod/v1/pod.proto
```

## Secret

Worker 表单和 YAML 不传 API key：

- 模型凭据来自 ModelBinding 指向的 AI model resource connection。
- Worker 类型 Secret 通过 EnvironmentBundle 引用。
- 环境包不能覆盖模型资源管理的 credential 字段。
- Plan、Diff、导出和错误不回显 Secret。

## External REST

现有 API-key 集成继续使用：

```http
POST /api/v1/ext/orgs/{org_slug}/workers
Authorization: Bearer amk_...
Content-Type: application/json
```

```json
{
  "agent_slug": "codex-cli",
  "repository_id": 7,
  "model_resource_id": 42,
  "automation_level": "autonomous",
  "agentfile_layer": "PROMPT \"Implement JWT refresh and add tests\""
}
```

该请求不提交 Resource Draft，不生成 Resource Plan，也不能表达完整
WorkerTemplate 引用、不可变 revision 和 typed Apply 结果。`/pods` 是该历史
handler 的兼容别名。

| Action | Method | Endpoint |
| --- | --- | --- |
| List | `GET` | `/workers` |
| Get | `GET` | `/workers/{pod_key}` |
| Create | `POST` | `/workers` |
| Send prompt | `POST` | `/workers/{pod_key}/prompt` |
| Terminate | `POST` | `/workers/{pod_key}/terminate` |

## 选择建议

- 产品 UI 和新 typed 客户端使用 Resource-native Worker。
- 需要构造完整 WorkerSpec 的内部服务使用 Direct WorkerSpec。
- 只有既有 API-key 集成继续使用 External REST。

不要把 External REST 描述为 Resource-native Worker，也不要在 typed Apply
失败后静默改用 legacy 路径。
