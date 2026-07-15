# 资源 Kind 声明参考

所有资源使用 `agentsmesh.io/v1alpha1`。`metadata.namespace` 必须等于当前组织
slug；同组织 ResourceRef 可以省略 namespace。

## 基础引用资源

[绑定资源、Prompt 与 WorkerTemplate](resource-build-blocks-reference.md)说明：

- `ModelBinding`、`ToolBinding`
- `Repository`、`Skill`、`KnowledgeBase`
- `EnvironmentBundle`、`ComputeTarget`、`ResourceProfile`
- `Prompt`
- `WorkerTemplate`

这些资源先独立 Apply，再由 Worker、Expert、Workflow 或 GoalLoop 通过
ResourceRef 引用。上层资源不会复制凭据或可变配置。

## 执行资源

[Worker、Expert、Workflow 与 GoalLoop](resource-execution-reference.md)说明不同
执行意图的字段和约束：

- `Worker`：一次性启动声明，创建后不能更新。
- `Expert`：可复用专家定义，固定 WorkerTemplate 与可选 Prompt。
- `Workflow`：可重复或定时执行，固定 WorkerTemplate 与必填 Prompt。
- `GoalLoop`：目标驱动循环；当前支持 Validate 和 Plan，尚无 typed Apply。

## 共同约束

- `metadata.name`、`metadata.namespace` 和 ResourceRef identifier 必须匹配
  `^[a-z0-9]+(-[a-z0-9]+)*$`，长度为 2 到 100。
- 省略 ResourceRef revision 只表示 Plan 时解析 active revision；Plan 会固定
  实际 revision 和 digest。
- Secret 只能通过受控资源引用，不能写入 YAML。
- Apply 必须消费当前、未过期且未被使用的 Plan。
