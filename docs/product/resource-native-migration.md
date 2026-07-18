# 资源原生迁移说明

## 适用版本

资源原生控制面由以下迁移组成，必须按编号执行：

| 迁移 | 作用 |
| --- | --- |
| `000211_orchestration_resources` | 创建 resource、revision 和 plan 表 |
| `000212_orchestration_resource_integrity` | 增加不可变性、版本和 Plan 消费约束 |
| `000215_orchestration_domain_links` | 把 Expert、Workflow 和运行记录固定到资源 revision |
| `000216_orchestration_worker_launches` | 增加 Worker Apply 的持久 launch 与 Pod 关联 |
| `000217_orchestration_goal_loop_link` | 把 GoalLoop 固定到资源 revision 与 WorkerSpec 快照 |
| `000226_enforce_orchestration_domain_snapshot_consistency` | 强制领域对象引用同一 resource revision 的准确 WorkerSpec 快照 |
| `000227_workflow_run_execution_manifest` | 为活跃资源原生 Workflow run 保存并校验不可变执行清单 |
| `000228_worker_spec_optional_model_binding` | 允许不需要模型的 Worker 类型保存空 model binding |

不要调整这些迁移的编号。当前正式主线前置迁移是
`000225_agent_workbench_stream`，本模块必须按 `000226`、`000227`、`000228`
顺序执行。后续迁移需要从 `000229` 之后重新与发布主线 owner 确认，不能只根据
本地空闲文件名占号。

## 升级行为

- 迁移不把历史 Expert、Workflow 或 Pod 自动转换成资源。
- Expert 和 Workflow 的资源关联列允许为空，历史记录继续按原领域数据运行。
- 新的资源 Apply 会同时写入资源 revision、WorkerSpec 快照和领域关联。
- 已写入的资源 revision 不允许更新或删除。
- Worker Apply 使用唯一 plan/launch 关系，避免重试时创建多个 Pod。
- GoalLoop Apply 写入 draft 领域对象，并通过组织复合外键固定 resource revision
  与 WorkerSpec 快照；历史 GoalLoop 的关联列保持为空。
- Expert、Workflow、Workflow run、GoalLoop 和 Worker launch 的数据库外键同时
  校验组织、resource、revision 和 WorkerSpec snapshot，不能拼接不属于该
  revision 的快照。
- 新建 Workflow run 会保存 `execution_manifest`。已完成的历史 run 可以保持
  manifest 为空；所有未完成 run 在 `000227` 执行前必须排空。
- Worker Definition 声明 `model_requirement.required=false` 时，WorkerSpec
  可以保存规范化空 model binding；必需模型的 Worker 类型仍必须固定模型资源。

这是显式双模式数据合同，不是运行时 fallback：历史对象保持历史来源，新对象
只有在通过资源 Apply 创建时才进入 resource-native 模式。系统不会在资源 Apply
失败后改走旧创建路径。

## 部署顺序

1. 备份 PostgreSQL。
2. 确认当前 migration version 为 `000225` 且 `dirty=false`。
3. 检查领域 revision/snapshot 一致性；发现不一致时先修复数据，不执行
   `000226`。
4. 排空所有 `finished_at IS NULL` 的 Workflow run；`000227` 不猜测活跃运行的
   历史执行参数。
5. 部署包含 `000226` 至 `000228` 与新后端的版本。
6. 执行 migration up，并确认 version 为 `000228`、`dirty=false`。
7. 启动后端并检查资源 Connect RPC 已挂载。
8. 通过 UI 或 API 创建 Prompt、绑定资源和 WorkerTemplate。
9. 生成并 Apply 一个 Expert 或 Workflow，检查领域对象的资源关联。
10. 触发 Workflow，检查 run 保存 execution manifest。
11. 创建一个 Worker，检查 launch、Pod 和 WorkerSpec 快照关联。
12. Apply 一个 GoalLoop，检查其状态为 draft、资源关联完整且没有创建 Pod。

开发环境继续使用：

```bash
./deploy/dev/dev.sh
```

生产环境使用仓库既有的 `golang-migrate` 流程，不手工修改表结构。

## 上线前检查

`000226` 会对每个已绑定领域对象执行同样的一致性检查。以下查询结果都必须为
0：

```sql
SELECT 'experts' AS relation, count(*) AS mismatches
FROM experts domain_row
WHERE domain_row.orchestration_resource_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM orchestration_resource_revisions revision
    WHERE revision.organization_id = domain_row.organization_id
      AND revision.resource_id = domain_row.orchestration_resource_id
      AND revision.revision = domain_row.orchestration_resource_revision
      AND revision.worker_spec_snapshot_id = domain_row.worker_spec_snapshot_id
  )
UNION ALL
SELECT 'workflows', count(*)
FROM workflows domain_row
WHERE domain_row.orchestration_resource_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM orchestration_resource_revisions revision
    WHERE revision.organization_id = domain_row.organization_id
      AND revision.resource_id = domain_row.orchestration_resource_id
      AND revision.revision = domain_row.orchestration_resource_revision
      AND revision.worker_spec_snapshot_id = domain_row.worker_spec_snapshot_id
  )
UNION ALL
SELECT 'workflow_runs', count(*)
FROM workflow_runs domain_row
WHERE domain_row.orchestration_resource_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM orchestration_resource_revisions revision
    WHERE revision.organization_id = domain_row.organization_id
      AND revision.resource_id = domain_row.orchestration_resource_id
      AND revision.revision = domain_row.orchestration_resource_revision
      AND revision.worker_spec_snapshot_id = domain_row.worker_spec_snapshot_id
  )
UNION ALL
SELECT 'goal_loops', count(*)
FROM goal_loops domain_row
WHERE domain_row.orchestration_resource_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM orchestration_resource_revisions revision
    WHERE revision.organization_id = domain_row.organization_id
      AND revision.resource_id = domain_row.orchestration_resource_id
      AND revision.revision = domain_row.orchestration_resource_revision
      AND revision.worker_spec_snapshot_id = domain_row.worker_spec_snapshot_id
  )
UNION ALL
SELECT 'orchestration_worker_launches', count(*)
FROM orchestration_worker_launches domain_row
WHERE NOT EXISTS (
  SELECT 1
  FROM orchestration_resource_revisions revision
  WHERE revision.organization_id = domain_row.organization_id
    AND revision.resource_id = domain_row.resource_id
    AND revision.revision = domain_row.resource_revision
    AND revision.worker_spec_snapshot_id = domain_row.worker_spec_snapshot_id
);
```

`000227` 前必须没有活跃 run：

```sql
SELECT count(*) AS active_workflow_runs
FROM workflow_runs
WHERE finished_at IS NULL;
```

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

SELECT id, orchestration_resource_id, orchestration_resource_revision,
       worker_spec_snapshot_id, execution_manifest
FROM workflow_runs
WHERE finished_at IS NULL;
```

## 回滚与验收

生产回滚边界和交付验收清单见
[资源原生迁移回滚与验收](resource-native-migration-release-checklist.md)。
