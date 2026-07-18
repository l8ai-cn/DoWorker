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
apiVersion: agentsmesh.io/v1alpha1
kind: ModelBinding
metadata:
  name: coding-primary
  namespace: acme
spec:
  resourceId: 101
```

`ToolBinding` 固定一个 ModelBinding：

```yaml
apiVersion: agentsmesh.io/v1alpha1
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
apiVersion: agentsmesh.io/v1alpha1
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
apiVersion: agentsmesh.io/v1alpha1
kind: WorkerTemplate
metadata:
  name: codex-reviewer
  namespace: acme
spec:
  optionsRevision: runtime-catalog-2026-07-13-release-gated
  workerType: codex-cli
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
          name: codex-settings
    instructions: Review before editing.
  lifecycle:
    terminationPolicy: manual
    idleTimeoutMinutes: 0
  metadata:
    alias: Reviewer
```

`runtime.resourceProfileRef` 与 `customResources` 互斥。仓库、Skill、知识库和
环境包分别通过 `repositoryRef`、`skillRefs`、`knowledgeMounts` 和
`environmentBundleRefs` 引用。

Worker Definition 是配置文档契约的唯一来源。每个
`configDocumentBindings[].documentId` 必须由当前 Worker 类型声明，绑定值必须
是 `EnvironmentBundle` 的 config 资源。`required: true` 的配置文档必须绑定；
`required: false` 的配置文档可以完全省略，不能用空 ResourceRef 占位。
