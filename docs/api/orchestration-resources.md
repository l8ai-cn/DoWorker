# 资源编排 API

## 协议

资源控制面通过 Connect unary RPC 暴露：

```text
proto.orchestration_resource.v1.OrchestrationResourceService
```

开发环境入口位于 `http://localhost:10000/proto.orchestration_resource.v1...`。
请求需要现有 JWT 身份，并通过 `org_slug` 选择组织。浏览器客户端使用生成的
protobuf 类型和 Rust/WASM service runtime，不应手写 JSON REST 请求。

协议定义：

- `proto/orchestration_resource/v1/orchestration_resource.proto`
- Web facade：`clients/web/src/lib/api/facade/orchestrationResource.ts`

## 生命周期 RPC

| RPC | 写入 | 用途 |
| --- | --- | --- |
| `ValidateResource` | 否 | 解析 YAML/JSON，返回 canonical JSON 和问题 |
| `PlanResource` | 是 | 保存 15 分钟有效的不可变 Plan |
| `GetResourcePlan` | 否 | 查询 Plan、Diff、固定引用和状态 |
| `ApplyBindingResourcePlan` | 是 | Apply 八类绑定资源 |
| `ApplyPromptPlan` | 是 | Apply Prompt |
| `ApplyWorkerTemplatePlan` | 是 | Apply 模板并返回 WorkerSpec 快照 ID |
| `CreateWorkerFromPlan` | 是 | 创建一次性 Worker launch 与 Pod |
| `CreateGoalLoopFromPlan` | 是 | 创建 GoalLoop 草稿与固定 WorkerSpec 快照 |
| `ApplyExpertPlan` | 是 | Apply Expert 并返回领域 Expert ID |
| `ApplyWorkflowPlan` | 是 | Apply Workflow 并返回领域 Workflow ID |

## 查询与导出

| RPC | 说明 |
| --- | --- |
| `GetResource` | 按 apiVersion、kind、namespace、name 读取 head |
| `GetResourceCapabilities` | 查询当前 actor 对目标的读取、引用和 Plan 权限 |
| `ListResources` | 按组织列出资源，可按 kind、offset、limit 和类型化引用条件过滤 |
| `ExportResource` | 导出 active 或指定 revision 的 YAML/JSON |

`GetResourceCapabilities` 返回 `exists`、`can_view_source`、`can_reference` 和
`can_plan`。目标不存在时，`can_plan` 使用创建权限；目标存在时使用当前资源和
实时组织角色计算引用、读取 source 与更新 Plan 权限。该响应只用于避免前端先
加载无权查看的 source，不替代 Validate、Plan 或 typed Apply 的服务端再授权。

`ExportResource.revision` 省略或为 0 时导出 active revision。提交草稿不能包含
导出结果中的服务器字段。

列出 `EnvironmentBundle` 引用候选时，可以同时提交
`environment_bundle_filter`：

```proto
environment_bundle_filter {
  purpose: ENVIRONMENT_BUNDLE_PURPOSE_CREDENTIAL
  worker_type: "do-agent"
  target_name: "DO_API_KEY"
}
```

该过滤器只能与 `kind: EnvironmentBundle` 一起使用。服务端从资源 active
revision 的 `environmentBundleId` 关联当前 EnvBundle，再按当前 actor 的用户/
组织 ownership、`is_active`、`agent_slug` 和用途过滤。`RUNTIME` 接受
`runtime` 与 `shared`，`CONFIG` 只接受 `config`，`CREDENTIAL` 只接受
`credential`。`CREDENTIAL` 必须同时提交当前 Worker Definition 已声明的
`target_name`；服务端只返回包含该目标 key 的包。`RUNTIME` 会排除包含
Worker Definition 模型资源托管字段的包，避免环境包覆盖模型绑定。筛选只检查
EnvBundle 的 key，不读取或返回 Secret 值。

响应必须通过 `applied_environment_bundle_filter` 回显实际应用的用途、Worker
类型和目标字段；客户端发现缺失或不一致时必须拒绝候选，不会接受旧服务端忽略
未知字段后返回的未过滤列表。不支持的组合、未知 Worker 类型或未声明的
`target_name` 返回 `invalid_argument`。
active revision、`environmentBundleId` 或目标 EnvBundle 绑定损坏时返回
`unavailable`，不会静默跳过损坏资源。完整性检查、总数和分页读取在同一一致性
快照中执行，避免 EnvBundle 并发删除造成检查结果与候选列表不一致。

## TypeScript 调用

Web 应用统一从 facade 调用：

```ts
import {
  planResource,
  applyPromptPlan,
} from "@/lib/api/facade/orchestrationResource";
import {
  IssueSeverity,
  SourceFormat,
} from "@proto/orchestration_resource/v1/orchestration_resource_pb";

const plan = await planResource("acme", {
  format: SourceFormat.YAML,
  content: yaml,
});

if (
  !plan.plan
  || plan.issues.some((issue) => issue.severity === IssueSeverity.BLOCKING)
) {
  throw new Error("resource plan is blocked");
}

const resource = await applyPromptPlan("acme", plan.plan.planId);
```

客户端必须根据 Kind 调用对应 typed Apply，不能通过一个通用 Apply 猜测目标：

```text
binding kinds -> ApplyBindingResourcePlan
Prompt        -> ApplyPromptPlan
WorkerTemplate-> ApplyWorkerTemplatePlan
Worker        -> CreateWorkerFromPlan
GoalLoop      -> CreateGoalLoopFromPlan
Expert        -> ApplyExpertPlan
Workflow      -> ApplyWorkflowPlan
```

## ResourceSource

```proto
message ResourceSource {
  SourceFormat format = 1;
  bytes content = 2;
}
```

支持 `SOURCE_FORMAT_JSON` 和 `SOURCE_FORMAT_YAML`。YAML 最大 256 KiB，JSON
最大 1 MiB；详细限制见
[资源 YAML 用户手册](../product/resource-yaml-manual.md)。

## Plan 结果

`ResourcePlan` 包含：

- `operation`：CREATE 或 UPDATE
- `base_resource_version`：更新计划的并发基线
- `draft_hash`、`plan_hash`、`artifact_digest`
- `resolved_references`：uid、revision 和 digest
- `semantic_diff`：ADD、REMOVE、REPLACE
- `issues`：blocking 或 warning
- `artifact_kind`、`options_revision`
- `created_at`、`expires_at`、`status`

有 blocking issue 时 `PlanResourceResponse.plan` 不可用于 Apply。草稿改变后必须
重新 Plan，不能复用旧 plan ID。

## Apply 结果

- 绑定资源和 Prompt 返回 `Resource`。
- WorkerTemplate 额外返回 `worker_spec_snapshot_id`。
- Worker 返回 `launch_id`、`pod_id`、`pod_key`、snapshot、revision 和 runner。
- GoalLoop 返回 `goal_loop_id`、snapshot 和 resource revision。
- Expert 返回 `expert_id`、snapshot 和 resource revision。
- Workflow 返回 `workflow_id`、snapshot 和 resource revision。

Apply 会重新校验 actor、组织权限、Plan 是否过期或已消费、目标 head 是否仍为
基线版本，以及固定引用是否仍可读取。它不会在失败后走旧 API。

`CreateGoalLoopFromPlan` 只创建状态为 `draft` 的领域对象，不创建 Pod、不启动
循环。Start、Verify 和 Cancel 由 GoalLoop 服务的显式领域 RPC 负责。
Workflow Trigger 创建 run 时会固定完整 execution manifest；运行启动、回调、
保留和 timeout/idle 扫描不会重新解析当前 Workflow head。

## 旧定义入口

以下入口不再创建资源定义：

- Workflow Connect `CreateWorkflow`
- GoalLoop Connect `CreateGoalLoop`
- GoalLoop Connect `RunLoopProgram`
- Runner MCP `create_workflow`

它们在完成组织身份解析后返回 `failed_precondition` 或 MCP `409`，并要求调用
资源 `ValidateResource`、`PlanResource` 和对应 typed Apply。

资源托管的 Expert/Workflow 通过旧 Update/Delete 修改时也会被拒绝。Workflow
Enable/Disable/Trigger、GoalLoop Start/Verify/Cancel 和 Expert Run 仍是领域
动作。Trigger 如果发现 Workflow 没有完整 resource revision 与 snapshot 绑定，
返回前置条件错误，不会创建 run。

## 错误代码

| Connect code | 场景 |
| --- | --- |
| `invalid_argument` | 请求、source 或 target 无效 |
| `permission_denied` | 组织或引用权限不足 |
| `not_found` | 资源、revision 或 Plan 不存在 |
| `failed_precondition` | 旧定义入口、缺少 resource/snapshot 绑定或领域状态不允许 |
| `aborted` | stale、expired、consumed、冲突或 options revision 变化 |
| `unavailable` | planner、repository 或依赖服务不可用 |
| `internal` | 未分类服务端错误 |

服务端错误使用稳定的通用消息，不回显 YAML 中可能包含的敏感输入。

## 幂等与并发

- Plan 只能原子消费一次。
- UPDATE 依赖 `base_resource_version`，head 变化后返回 `aborted`。
- CREATE 遇到同 identity 已存在时返回冲突。
- Apply 写事务重新锁定组织成员关系；撤销成员或 admin 降级会阻止写入。
- 已消费 Plan 的幂等重放也重新检查当前成员权限。
- Worker 使用持久 launch 记录和唯一 plan 约束；重试不会创建第二个 Worker
  资源或第二个 Pod 关联。
- GoalLoop 使用 create-only 计划；同名 resource 或历史 GoalLoop 会在 Plan
  阶段返回 blocking issue。
