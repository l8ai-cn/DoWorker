# 资源 YAML 用户手册

## 适用范围

YAML 是资源编辑器的高级视图，适合代码审查、复制、版本管理和批量生成。普通
用户可以只使用领域表单；表单与 YAML 共享同一份 typed draft，不存在两套配置。

当前可完整 Apply 的 Kind：

- `WorkerTemplate`、`Worker`、`Prompt`、`Expert`、`Workflow`、`GoalLoop`
- `ModelBinding`、`ToolBinding`、`Repository`、`Skill`
- `KnowledgeBase`、`EnvironmentBundle`、`ComputeTarget`
- `ResourceProfile`

各 Kind 的字段和示例见[资源 Kind 声明参考](resource-kind-reference.md)。
GoalLoop Apply 创建草稿和固定 WorkerSpec 快照，启动仍是后续显式操作。

## 编辑流程

1. 在资源编辑器的“配置”页填写领域字段。
2. 切换到“YAML”审查同一草稿，或直接修改 YAML。
3. 可先执行“校验”，提前处理 schema、语义、权限和引用问题。
4. 执行“生成计划”，审查 CREATE/UPDATE、语义 Diff 和固定引用。
5. 确认 Plan 对应当前草稿后，执行该 Kind 的 Apply。
6. 检查返回的 revision、WorkerSpec 快照、领域对象或 Pod。

生成 Plan 时服务端也会执行完整校验，不能用 Plan 绕过“校验”按钮覆盖的检查。
任意草稿修改都会使旧 Plan 失效。当前 Plan 默认 15 分钟过期，且只能消费一次。
Apply 成功、失败或客户端未收到确定结果后，编辑器都会禁用旧 Plan；重新 Apply
前必须生成并审查新 Plan。WorkerTemplate YAML 还会在调用 Plan 前重新加载当前
组织的 Worker 选项和引用目录，失效事实不会沿用表单先前缓存。

## 单一 Draft 行为

- 表单变化会立即重新编码为 YAML。
- 有效 YAML 会解析回同一 typed draft。
- YAML 有错误时原始文本保留，不覆盖为旧内容。
- 错误修复前不能切回表单、生成 Plan 或 Apply。
- 系统不会在 YAML 错误后使用上一个有效版本。
- 服务端 canonical 格式是最终格式，不保证保留注释和原始字段顺序。

## 格式限制

| 限制 | 值 | 结果 |
| --- | --- | --- |
| YAML 源文件 | 最大 256 KiB | 超出时拒绝 |
| YAML 编码结果 | 最大 256 KiB | 超出时拒绝导出 |
| 单个物理行 | 最大 64 KiB | 拒绝解析或编码 |
| 节点数量 | 最大 10,000 | 拒绝文档 |
| 容器嵌套深度 | 最大 64 | 拒绝文档 |
| 文档数量 | 只能 1 个 | 拒绝多文档 |
| JSON 源文件 | 最大 1 MiB | 超出时拒绝 |

知识正文、Skill 包和大段配置不要压成超长 YAML。先保存到对应领域对象，再创建
绑定资源并通过 ResourceRef 引用。

## 不支持的 YAML 特性

以下输入会明确失败：

- 重复 mapping key
- anchor、alias 和 merge key `<<`
- 自定义 tag
- timestamp 和 binary 类型
- 十六进制、`.inf`、`.NaN` 等非 JSON 数字
- 超出 JavaScript 安全整数范围的整数
- 非字符串 mapping key
- 多文档输入
- 未知字段和大小写错误字段

如果业务值就是 `<<`，必须写成字符串：

```yaml
"<<": literal-value
```

## 字符串与标量

未加引号的标量按严格 JSON 语义解释：

```yaml
enabled: true
retryCount: 3
ratio: 0.75
optional: null
```

看起来像数字、布尔值、null、日期或 YAML 控制符的业务字符串必须加引号：

```yaml
modelId: "1e9999"
featureName: "true"
literalNull: "null"
releaseDate: "2026-07-15"
hexLabel: "0x10"
```

超过 `Number.MAX_SAFE_INTEGER` 的整数会被拒绝，避免浏览器在 Validate 前
静默舍入。需要更大值时，应使用后端 schema 明确声明的字符串字段，不能把数字
伪装成字符串绕过当前 Kind 的类型校验。

## 字段与 identifier

字段名区分大小写：

```yaml
apiVersion: agentsmesh.io/v1alpha1
kind: WorkerTemplate
metadata:
  displayName: Example
```

`ApiVersion`、`Kind`、`display_name` 不会被兼容或忽略。

`metadata.name`、`metadata.namespace`、ResourceRef 的 `name` 和 `namespace`
都必须满足：

```text
^[a-z0-9]+(-[a-z0-9]+)*$
```

长度为 2 到 100，并且不能使用平台保留字。展示文本应放在 `displayName`。

以下字段由服务器管理，不能放进提交草稿：

- `metadata.uid`
- `metadata.resourceVersion`
- `metadata.generation`
- `status`

## ResourceRef

普通引用：

```yaml
modelRef:
  kind: ModelBinding
  name: coding-primary
  revision: 4
```

省略 revision 时，Plan 解析当前 active revision。Plan 生成后会显示实际
revision 和 sha256 digest；Apply 后不会继续跟随名称对应的最新版。

`apiVersion` 和 `namespace` 在引用同版本、同组织资源时可以省略。引用的 Kind
必须与字段契约一致，例如 `workerTemplateRef` 不能指向 `Worker`。

## Worker 类型声明

`WorkerTemplate.spec.workerType` 选择 Worker Definition。运行时目录会为该类型
返回模型、credential 和配置文档要求；`optionsRevision` 固定本次声明所依据的
目录版本。

配置文档必须使用带 `documentId` 的命名绑定，不能提交匿名配置包数组：

```yaml
spec:
  optionsRevision: worker-catalog-2026-07-16
  workerType: do-agent
  workspace:
    configDocumentBindings:
      - documentId: settings
        configBundleRef:
          kind: EnvironmentBundle
          name: do-agent-settings
          revision: 2
```

`documentId` 必须与当前 Worker Definition 的 `config_documents` 声明完全一致；
`configBundleRef` 只能引用 `EnvironmentBundle`。领域表单会显示该文档的格式和
目标路径。切换 Worker 类型或刷新目录时，系统保留仍被新类型声明的同名绑定，
删除不再适用的绑定，并为新增必填文档留出空引用等待用户选择，不会把旧匿名
配置猜测为新文档。

credential 字段也由 Worker Definition 声明。领域表单只显示
`source_kind: credential_bundle` 的要求，并把 `target_name` 作为
`typeConfig.secretRefs` 的 key；未被当前 Worker 类型声明的 Secret 引用会被
移除。EnvironmentBundle 候选会按用途分开：`environmentBundleRefs` 只显示
runtime-safe 包，`configBundleRef` 只显示 config 包，Secret 引用只显示
包含当前 `target_name` 的 credential 包；三者都要求当前用户或组织可访问、
处于 active 状态，并与所选 Worker 类型的 `agentSlug` 兼容。runtime-safe
还意味着包中不能声明由模型资源管理的环境字段。字段兼容规则由服务端从 Worker
Definition 计算，前端不会复制字段清单，也不会读取 Secret 值。

YAML 高级视图允许手工填写 ResourceRef；领域表单只允许从当前组织目录选择。
目录加载、权限错误、没有候选或已保存引用无法解析时，表单保留引用并只读。
候选过滤不替代 Validate/Plan；EnvBundle 在加载后被停用、改变用途或兼容性、
撤销权限或修改字段时，Plan 会明确拒绝。

## Secret

ResourceRef 不能添加 `value`、`token`、`password` 或其他明文字段。
WorkerTemplate 通过 `EnvironmentBundle` 引用 Secret：

```yaml
typeConfig:
  secretRefs:
    CURSOR_API_KEY:
      kind: EnvironmentBundle
      name: cursor-credentials
      revision: 3
```

这里的 key 是 Worker Definition 的 credential `target_name`。例如
`cursor-cli` 声明 `CURSOR_API_KEY`；不能改写成小写 slug。AI 模型 API 先在
“设置 -> 组织 -> AI 资源”中创建，凭据加密保存；YAML 中的 `ModelBinding`
只声明对应模型资源 ID，不包含凭据。

## Plan 审查

Apply 前至少检查：

- 操作是预期的 CREATE 或 UPDATE
- 没有 blocking issue
- warning 已被理解
- 语义 Diff 路径与预期一致
- 每个引用的 Kind、name、revision 和 digest 正确
- WorkerTemplate 的 `optionsRevision` 仍对应当前运行时目录
- `configDocumentBindings[].documentId` 与当前 Worker 类型要求一致

如果草稿、权限、依赖、资源 head 或运行时目录发生变化，应重新生成 Plan。

## 常见错误

| 错误 | 处理 |
| --- | --- |
| `unknown field` | 检查 Kind 的字段和大小写 |
| `duplicate key` | 删除重复声明 |
| `mapping key must be string` | 为业务字符串 key 加引号 |
| `stale/expired/consumed plan` | 重新生成并审查 Plan；需要时先单独 Validate |
| `apply conflict or unknown result` | 不重放旧 Plan，重新生成 Plan 后再 Apply |
| `forbidden reference` | 选择当前组织有权读取的资源 |
| `reference not found` | 先 Apply 被引用资源或修正名称 |
| `options revision is stale` | 使用当前 Worker 创建选项返回的版本 |
| `worker-is-create-only` | 使用新名称创建新的 Worker |
| `goal-loop-is-create-only` | 使用新名称创建新的 GoalLoop |
| `goal-loop-name-already-exists` | 选择未被资源或历史 Loop 占用的名称 |
| `workflow execution requires a complete orchestration resource binding` | 用资源编辑器 Apply Workflow 后再 Trigger |
| `EXPERT_MANAGED_BY_RESOURCE_APPLY` | 在 Expert 资源草稿中修改并 Apply 新 revision |

错误消息有长度上限，重复 key 和未知 key 不回显用户输入，避免误粘贴的 Secret
进入日志。
