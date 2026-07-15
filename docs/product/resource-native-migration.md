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

不要调整这些迁移的编号。它们排在共享工作区已有的 `000207` 至 `000210`
之后，并与 `000213`、`000214` 的 GoalLoop 状态迁移保持现有顺序。

## 升级行为

- 迁移不把历史 Expert、Workflow 或 Pod 自动转换成资源。
- Expert 和 Workflow 的资源关联列允许为空，历史记录继续按原领域数据运行。
- 新的资源 Apply 会同时写入资源 revision、WorkerSpec 快照和领域关联。
- 已写入的资源 revision 不允许更新或删除。
- Worker Apply 使用唯一 plan/launch 关系，避免重试时创建多个 Pod。
- GoalLoop Apply 写入 draft 领域对象，并通过组织复合外键固定 resource revision
  与 WorkerSpec 快照；历史 GoalLoop 的关联列保持为空。

这是显式双模式数据合同，不是运行时 fallback：历史对象保持历史来源，新对象
只有在通过资源 Apply 创建时才进入 resource-native 模式。系统不会在资源 Apply
失败后改走旧创建路径。

## 部署顺序

1. 备份 PostgreSQL。
2. 确认当前 migration version 和 dirty 状态。
3. 部署包含新 migration 与新后端的版本。
4. 执行 migration up。
5. 启动后端并检查资源 Connect RPC 已挂载。
6. 通过 UI 或 API 创建 Prompt、绑定资源和 WorkerTemplate。
7. 生成并 Apply 一个 Expert 或 Workflow，检查领域对象的资源关联。
8. 创建一个 Worker，检查 launch、Pod 和 WorkerSpec 快照关联。
9. Apply 一个 GoalLoop，检查其状态为 draft、资源关联完整且没有创建 Pod。

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

## 回滚边界

下迁移会先删除 GoalLoop、其他领域外键和 Worker launch 关联，再删除资源表。
已经通过资源 Apply 创建的 Expert、Workflow、GoalLoop 或 Pod 可能依赖这些
关联，因此生产回滚前必须先评估新数据，而不是直接执行全部 down。

WorkerSpec 快照不由资源迁移删除。资源 revision 对快照使用 `ON DELETE RESTRICT`，
避免回滚或误操作破坏仍被领域对象引用的运行事实。

## 验收

- migration version 正确且 `dirty=false`
- 旧 Expert 和 Workflow 仍可读取
- 新资源的 revision、digest 和固定引用可导出
- Expert/Workflow 指向同组织 resource revision 与 WorkerSpec 快照
- GoalLoop 指向同组织 resource revision 与 WorkerSpec 快照，状态为 draft 且
  Apply 后没有 Pod
- Worker 重试只产生一个 launch 和一个 Pod 关联
- Apply 失败时没有旧路径或静默降级
