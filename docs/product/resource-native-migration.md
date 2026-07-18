# 资源原生迁移说明

## 适用版本

以下是当前工作区的资源迁移草案，不是可直接执行的正式序列：

| 迁移 | 作用 |
| --- | --- |
| `000211_orchestration_resources` | 创建 resource、revision 和 plan 表 |
| `000212_orchestration_resource_integrity` | 增加不可变性、版本和 Plan 消费约束 |
| `000215_orchestration_domain_links` | 把 Expert、Workflow 和运行记录固定到资源 revision |
| `000216_orchestration_worker_launches` | 增加 Worker Apply 的持久 launch 与 Pod 关联 |
| `000217_orchestration_goal_loop_link` | 把 GoalLoop 固定到资源 revision 与 WorkerSpec 快照 |
| `000219_enforce_orchestration_domain_snapshot_consistency` | 强制领域投影、WorkflowRun 和 Worker launch 的 revision/snapshot 一致 |
| `000220_workflow_run_execution_manifest` | 固定 WorkflowRun 的完整执行配置并约束可安全查询的字段类型 |
| `000221_worker_spec_optional_model_binding` | 允许不要求模型协议的 Worker Definition 生成无 ModelBinding 的 WorkerSpec |

正式发布主线的迁移头在 2026-07-16 已确认到 `000224`，尾部依次为
`000221_add_expert_revision`、`000222_add_video_studio_agent`、
`000223_align_seedance_do_agent_home`、`000224_validate_migration_lineage`。

当前工作区只看到未提交迁移到 `000221`，其中
`000221_worker_spec_optional_model_binding` 与正式 `000221_add_expert_revision`
同号冲突。冻结期不得重编号或占用 `000222` 到 `000224`；解除后由发布负责人
确认从 `000225` 起的正式排序。Runner、GoalLoop 和 migration 继续冻结。
`000219` 必须排在领域关联和 Worker launch 迁移之后，`000220` 必须排在
WorkflowRun 的资源关联与快照一致性约束之后；冲突的 optional-model 迁移必须
排在 WorkerSpec 快照和 Worker Definition 驱动编译之后，但当前没有正式编号。

## 交付门禁：解析依赖快照

`000211` 到 `000221` 固定资源 revision、领域投影和 WorkerSpec 身份，但尚未
把所有底层运行事实物化成历史依赖快照。正式资源迁移序列应从候选 `000225`
开始，但具体迁移尚未排序或预占；创建文件前必须由发布负责人确认。解析依赖
迁移至少要为每个 WorkerSpec snapshot 固定：

- 主模型和工具模型的资源 ID/revision、Provider、协议适配器、Model ID 与
  BaseURL；
- Repository 的固定分支、解析 commit SHA 和 clone endpoint；
- Skill、KnowledgeBase、配置包、ComputeTarget 与 ResourceProfile 的 revision
  或内容 digest，以及运行所需的非 Secret 数据；
- Secret 只保存受权限保护的引用，不保存明文、密文或可导出的凭据值。

运行时必须只读取已经物化的非 Secret 依赖；缺少历史制品时 fail closed，不得
回读当前领域行、按名称查找最新版或自动重建。Secret 在每次启动时经过租户和
权限检查读取当前值，以支持凭据轮换；引用被禁用、删除或无权读取时必须明确
失败。只有 schema、确定性审计/重建、运行时硬切换、数据库约束和浏览器关键
路径完成后，才能宣称历史 WorkerSpec 可重放并解除本模块的部署冻结。完整设计
见 `docs/superpowers/plans/2026-07-16-worker-spec-resolved-dependency-artifact.md`。

## 升级行为

- 迁移不把历史 Expert、Workflow 或 Pod 自动转换成资源。
- Expert 和 Workflow 的资源关联列允许为空，历史记录继续可读取。
- snapshot-backed 历史 Expert 仍可运行；资源托管 Expert 的 Update/Delete 必须
  改走资源 Apply。
- 未绑定 resource revision/snapshot 的历史 Workflow 不能再创建新 run；Trigger
  会返回前置条件错误。应创建 resource-native 替代定义并切换调度或调用方。
- 新的资源 Apply 会同时写入资源 revision、WorkerSpec 快照和领域关联。
- 已写入的资源 revision 不允许更新或删除。
- Worker Apply 使用唯一 plan/launch 关系，避免重试时创建多个 Pod。
- GoalLoop Apply 写入 draft 领域对象，并通过组织复合外键固定 resource revision
  与 WorkerSpec 快照。`RunLoopProgram` 不再创建或启动 GoalLoop；调用方必须
  先通过资源 Apply 创建草稿，再显式执行 Start。
- `000219` 不做回填。发现 Expert、Workflow、WorkflowRun、GoalLoop 或 Worker
  launch 的 revision/snapshot 已错配时，迁移会抛出具体表名并整体回滚。
- `000220` 不猜测历史 Workflow 配置，也不回填 manifest。执行 up 前必须结束或
  取消所有活动 WorkflowRun；任何尚未结束且缺少 manifest 的 run 都会使迁移
  整体回滚。已结束的历史 run 可保留空 manifest。新活动 run 必须保存完整、
  类型安全的 execution manifest。
- `000221` 不为现有快照猜测或删除模型绑定。它只放宽新快照的数据库约束；
  Worker Definition 声明模型协议时，Plan 和编译器仍要求固定 ModelBinding。

这是显式来源合同，不是运行时 fallback：snapshot-only Expert 保持其不可变
来源；新的 GoalLoop 和 Workflow 定义只有通过资源 Apply 才能执行。系统不会
在资源 Apply 失败后改走旧创建或 RunLoopProgram 路径。

## 部署顺序

以下步骤只能在发布负责人确认从 `000225` 起的排序并解除冻结后执行。当前不得
运行 migration、启动依赖新 schema 的服务或部署。

1. 备份 PostgreSQL。
2. 确认当前 migration version 和 dirty 状态。
3. 执行 schema 阶段迁移，创建 resolved-dependency artifact 及其不可变约束。
4. 部署所有 WorkerSpec 写入端，使快照与 artifact 在同一事务内写入。
5. 审计历史快照；只物化能够证明的事实，其余活动定义必须重新 Plan/Apply。
6. 开启全局 fresh-launch 维护门禁，drain 在途启动并停止所有旧读取后端实例。
7. 整体启动 artifact-only 后端，门禁内验证后再一次性恢复全部启动入口。
8. 使用后续已确认迁移启用新快照的延迟完整性约束。
9. 检查资源 Connect RPC 已挂载，且旧读取实例数量为零。
10. 通过 UI 或 API 创建 Prompt、绑定资源和 WorkerTemplate。
11. 生成并 Apply 一个 Expert 或 Workflow，检查领域对象的资源关联。
12. 创建一个 Worker，检查 launch、Pod、WorkerSpec 快照和 artifact 关联。
13. Apply 一个 GoalLoop，检查其状态为 draft、资源关联完整且没有创建 Pod。
14. 触发一个 Workflow，检查 run 的 execution manifest 已固定沙箱、Autopilot、
    回调、保留、timeout 与 idle timeout 配置。

开发环境继续使用：

```bash
./deploy/dev/dev.sh
```

生产环境使用仓库既有的 `golang-migrate` 流程，不手工修改表结构。

## 数据检查

```sql
SELECT version, dirty FROM schema_migrations;

SELECT kind, namespace, name, active_revision, generation, resource_version
FROM orchestration_resources
ORDER BY id DESC;

SELECT resource_id, revision, digest, worker_spec_snapshot_id
FROM orchestration_resource_revisions
ORDER BY id DESC;

SELECT id, plan_id, state, pod_id, pod_key
FROM orchestration_worker_launches
ORDER BY id DESC;

SELECT id, workflow_id, orchestration_resource_revision,
       worker_spec_snapshot_id, execution_manifest
FROM workflow_runs
WHERE orchestration_resource_id IS NOT NULL
ORDER BY id DESC;
```

Expert 与 Workflow 资源关联：

```sql
SELECT id, slug, orchestration_resource_id,
       orchestration_resource_revision, worker_spec_snapshot_id
FROM experts
WHERE orchestration_resource_id IS NOT NULL;

SELECT id, slug, orchestration_resource_id,
       orchestration_resource_revision, worker_spec_snapshot_id
FROM workflows
WHERE orchestration_resource_id IS NOT NULL;

SELECT id, slug, status, orchestration_resource_id,
       orchestration_resource_revision, worker_spec_snapshot_id, pod_key
FROM goal_loops
WHERE orchestration_resource_id IS NOT NULL;
```

`000219` 前可用以下模式检查错配；应分别对 `experts`、`workflows`、
`workflow_runs`、`goal_loops` 和 `orchestration_worker_launches` 执行：

```sql
SELECT domain_row.id
FROM workflows domain_row
LEFT JOIN orchestration_resource_revisions revision
  ON revision.organization_id = domain_row.organization_id
 AND revision.resource_id = domain_row.orchestration_resource_id
 AND revision.revision = domain_row.orchestration_resource_revision
 AND revision.worker_spec_snapshot_id = domain_row.worker_spec_snapshot_id
WHERE domain_row.orchestration_resource_id IS NOT NULL
  AND revision.id IS NULL;
```

## 回滚边界

下迁移会先删除 GoalLoop、其他领域外键和 Worker launch 关联，再删除资源表。
已经通过资源 Apply 创建的 Expert、Workflow、GoalLoop 或 Pod 可能依赖这些
关联，因此生产回滚前必须先评估新数据，而不是直接执行全部 down。

`000220` down 在存在活动 resource-native WorkflowRun 时会 fail closed；必须先
停止触发并 drain run。通过后会删除 WorkflowRun execution manifest，回滚后的
运行时不能再保证历史 run 与后续 Workflow revision 隔离。`000219` down 会恢复
三列 revision FK，从而移除数据库层 snapshot 一致性保护；它不会修改已有数据。
`000218_normalize_agent_capability_heading` 是不可逆的文本归一化，down 会
fail closed。需要回退到 `000217` 或更早版本时，必须恢复执行迁移前的数据库
备份。

WorkerSpec 快照不由资源迁移删除。资源 revision 对快照使用 `ON DELETE RESTRICT`，
避免回滚或误操作破坏仍被领域对象引用的运行事实。

## 验收

- migration version 正确且 `dirty=false`
- 旧 Expert 和 Workflow 仍可读取，未绑定 Workflow 的 Trigger 明确失败
- 新资源的 revision、digest 和固定引用可导出
- Expert/Workflow/WorkflowRun 指向同一 resource revision 记录中的 WorkerSpec 快照
- 活动 WorkflowRun 的 execution manifest 完整且 timeout/idle 字段可安全转整数
- GoalLoop 指向同组织 resource revision 与 WorkerSpec 快照，状态为 draft 且
  Apply 后没有 Pod
- RunLoopProgram 返回前置条件错误，不创建 GoalLoop 或 Pod
- Worker launch 的 revision 与 WorkerSpec 快照来自同一 revision 记录，重试只
  产生一个 launch 和一个 Pod 关联
- 修改 Skill、Repository、Model ID 或 BaseURL 后，旧 WorkerSpec 仍读取已物化
  的历史非 Secret 事实；新 revision 使用新事实
- 凭据轮换不改变运行配置 revision，旧验证结果不能覆盖新凭据状态，快照和
  Diff 不包含 Secret 值
- 缺少解析依赖快照、Secret 引用失效或权限撤销时启动明确失败，不回读当前资源
- Apply 失败时没有旧路径或静默降级
