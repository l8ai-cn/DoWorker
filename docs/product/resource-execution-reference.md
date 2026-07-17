# 执行资源声明

## Worker

Worker 是一次性启动声明。Apply 会创建资源 revision、持久 launch 记录和 Pod；
同一 Worker 资源不能更新。

```yaml
apiVersion: agentsmesh.io/v1alpha1
kind: Worker
metadata:
  name: review-run-20260715
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

## Expert

Expert 固定 WorkerTemplate 和可选 Prompt，并创建现有 Expert 领域投影。

```yaml
apiVersion: agentsmesh.io/v1alpha1
kind: Expert
metadata:
  name: delivery-expert
  namespace: acme
  displayName: 交付专家
spec:
  workerTemplateRef:
    kind: WorkerTemplate
    name: codex-reviewer
  promptRef:
    kind: Prompt
    name: delivery-review
  description: Reviews a delivery and returns an actionable plan.
  category: engineering
  releaseNotes: Initial resource-native revision.
```

`category` 为空或使用 identifier；说明和发布说明分别最大 4,000 字符。

## Workflow

Workflow 的 WorkerTemplate 和 Prompt 都必填。

```yaml
apiVersion: agentsmesh.io/v1alpha1
kind: Workflow
metadata:
  name: nightly-review
  namespace: acme
spec:
  workerTemplateRef:
    kind: WorkerTemplate
    name: codex-reviewer
  promptRef:
    kind: Prompt
    name: delivery-review
  inputs: {}
  executionMode: direct
  cronExpression: "0 2 * * *"
  sandboxStrategy: fresh
  sessionPersistence: false
  concurrencyPolicy: skip
  maxConcurrentRuns: 1
  maxRetainedRuns: 30
  timeoutMinutes: 60
  idleTimeoutSeconds: 30
  callbackUrl: https://example.com/hooks/workflow
```

`executionMode` 为 `direct` 或 `autopilot`；`sandboxStrategy` 为 `fresh` 或
`persistent`。当前并发策略只支持 `skip`。会话保持要求 persistent sandbox
且最大并发为 1。回调不能指向 localhost、私网或 link-local 地址。

## GoalLoop

GoalLoop 是 create-only 资源。`CreateGoalLoopFromPlan` 会创建不可变 resource
revision、固定 WorkerSpec 快照和状态为 `draft` 的 GoalLoop 领域对象。Apply
不会创建 Pod，也不会自动启动。

```yaml
apiVersion: agentsmesh.io/v1alpha1
kind: GoalLoop
metadata:
  name: close-release-goal
  namespace: acme
spec:
  workerTemplateRef:
    kind: WorkerTemplate
    name: codex-reviewer
  objective: Complete the release and provide verification evidence.
  acceptanceCriteria:
    - All required tests pass.
    - Release notes are updated.
  verificationCommand: pnpm test
  maxIterations: 10
  tokenBudget: 200000
  timeoutMinutes: 60
  noProgressLimit: 3
  sameErrorLimit: 3
  escalationPolicy: pause
```

Apply 返回 `goal_loop_id`、`worker_spec_snapshot_id` 和
`resource_revision`。用户随后从 Loop 列表显式执行 Start；验证与取消也继续使用
GoalLoop 领域操作。

已有同名资源或历史 GoalLoop 会在 Plan 阶段产生 blocking issue。需要修改目标
或预算时，应使用新名称创建新的 GoalLoop，而不是更新既有声明。
