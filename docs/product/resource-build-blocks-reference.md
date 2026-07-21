# 基础引用资源声明

## 绑定资源

绑定资源把已有平台对象纳入资源 revision。数字 ID 只保存在绑定资源中，上层
资源只引用绑定资源名称。

| Kind | Spec |
| --- | --- |
| `ModelBinding` | `resourceId` |
| `Repository` | `repositoryId` |
| `Skill` | `skillId` |
| `KnowledgeBase` | `knowledgeBaseId` |
| `EnvironmentBundle` | `environmentBundleId` |
| `ComputeTarget` | `computeTargetId` |
| `ResourceProfile` | `resourceProfileId` |

```yaml
apiVersion: agentcloud.io/v1alpha1
kind: ModelBinding
metadata:
  name: coding-primary
  namespace: acme
spec:
  resourceId: 101
```

`ToolBinding` 固定一个 ModelBinding：

```yaml
apiVersion: agentcloud.io/v1alpha1
kind: ToolBinding
metadata:
  name: web-search
  namespace: acme
spec:
  modelRef:
    kind: ModelBinding
    name: coding-primary
```

## Prompt

Prompt 内容最大 65,536 字符，最多 128 个变量。变量名使用 identifier 规则，
默认值最大 8,192 字符。

```yaml
apiVersion: agentcloud.io/v1alpha1
kind: Prompt
metadata:
  name: delivery-review
  namespace: acme
spec:
  content: Review {{topic}} and return a concise delivery plan.
  variables:
    topic:
      required: true
    audience:
      required: false
      default: engineering
```

## WorkerTemplate

WorkerTemplate 是可复用运行配置。`optionsRevision` 来自 Worker 创建选项接口，
不能自行递增。`modelRef` 是否必填由 Worker 类型决定。

```yaml
apiVersion: agentcloud.io/v1alpha1
kind: WorkerTemplate
metadata:
  name: do-agent-reviewer
  namespace: acme
spec:
  optionsRevision: runtime-catalog-2026-07-13-release-gated
  workerType: do-agent
  modelRef:
    kind: ModelBinding
    name: coding-primary
  toolRefs: {}
  runtime:
    runtimeImageId: 1
    placementPolicy: automatic
    computeTargetRef:
      kind: ComputeTarget
      name: primary-pool
    deploymentMode: pooled
    customResources:
      cpuRequestMilliCPU: 500
      cpuLimitMilliCPU: 1000
      memoryRequestBytes: 536870912
      memoryLimitBytes: 1073741824
      storageRequestBytes: 1073741824
      storageLimitBytes: 10737418240
  typeConfig:
    schemaVersion: 1
    values: {}
    secretRefs: {}
    interactionMode: acp
    automationLevel: autonomous
  workspace:
    branch: ""
    skillRefs: []
    knowledgeMounts: []
    environmentBundleRefs: []
    configDocumentBindings:
      - documentId: settings
        configBundleRef:
          kind: EnvironmentBundle
          name: do-agent-settings
    instructions: Review before editing.
  lifecycle:
    terminationPolicy: manual
    idleTimeoutMinutes: 0
  metadata:
    alias: Reviewer
```

`runtime.resourceProfileRef` 与 `customResources` 互斥。仓库、Skill、知识库和
环境包分别通过 `repositoryRef`、`skillRefs`、`knowledgeMounts` 和
`environmentBundleRefs` 引用。Worker 类型要求的配置文档通过
`configDocumentBindings[].documentId` 对应声明，并由 `configBundleRef`
引用一个 `EnvironmentBundle`；表单会按 Worker 类型目录保留同名绑定并移除
不再适用的文档声明。EnvironmentBundle 候选按 Worker 类型和用途查询：
runtime 字段接受 `runtime/shared`，配置文档只接受 `config`，Secret 引用只
接受 `credential`。runtime 候选会排除包含模型资源托管字段的包，每个 Secret
字段只显示包含该 Worker Definition `target_name` 的包；不可访问、inactive
或 `agentSlug` 不兼容的资源也不会进入对应候选目录。YAML 可以手工声明其他
名称，但 Validate/Plan 仍会按当前事实阻止不兼容引用。
