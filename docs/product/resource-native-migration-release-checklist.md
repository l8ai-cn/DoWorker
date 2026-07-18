# 资源原生迁移回滚与验收

本文档继续
[资源原生迁移说明](resource-native-migration.md)，定义生产回滚边界和交付验收清单。

## 回滚边界

下迁移会先删除 GoalLoop、其他领域外键和 Worker launch 关联，再删除资源表。
已经通过资源 Apply 创建的 Expert、Workflow、GoalLoop 或 Pod 可能依赖这些
关联，因此生产回滚前必须先评估新数据，而不是直接执行全部 down。

`000218_normalize_agent_capability_heading` 是不可逆的文本归一化，down 会
fail closed。需要回退到 `000217` 或更早版本时，必须恢复执行迁移前的数据库备份。

WorkerSpec 快照不由资源迁移删除。资源 revision 对快照使用 `ON DELETE RESTRICT`，
避免回滚或误操作破坏仍被领域对象引用的运行事实。

- `000228` down 在存在空 model binding 快照时 fail closed。
- `000227` down 在存在未完成的资源原生 Workflow run 时 fail closed。
- `000226` down 会把领域外键从 revision/snapshot 四元组退回 revision 三元组，
  因而会削弱数据库一致性保证；必须先确认回退版本不会写入错误快照组合。

## 验收

- migration version 正确且 `dirty=false`
- 旧 Expert 和 Workflow 仍可读取
- 新资源的 revision、digest 和固定引用可导出
- Expert/Workflow 指向同组织 resource revision 与 WorkerSpec 快照
- 活跃 Workflow run 有 version 1 execution manifest，且组织、执行模式和
  sandbox 字段与行数据一致
- GoalLoop 指向同组织 resource revision 与 WorkerSpec 快照，状态为 draft 且
  Apply 后没有 Pod
- 不需要模型的 Worker 类型可 Plan/Apply，需要模型的类型缺少 ModelBinding 时
  仍然阻断
- Worker 重试只产生一个 launch 和一个 Pod 关联
- Apply 失败时没有旧路径或静默降级
