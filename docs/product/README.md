# 产品文档

本目录面向使用 Do Worker 的开发者和平台操作人员，说明产品对象、
配置流程及用户可见约束。

## 资源原生编排

- [资源原生编排指南](resource-native-orchestration.md)：理解
  WorkerTemplate、Worker、Expert、Workflow、GoalLoop 及其引用关系。
- [资源 YAML 用户手册](resource-yaml-manual.md)：编写、校验、计划和应用
  YAML 资源，并处理格式限制和常见错误。
- [资源 Kind 声明参考](resource-kind-reference.md)：查看 WorkerTemplate、
  Worker、Expert、Workflow、GoalLoop 及绑定资源的字段与示例。
- [基础引用资源声明](resource-build-blocks-reference.md)：查看绑定资源、
  Prompt 与 WorkerTemplate 的完整示例。
- [执行资源声明](resource-execution-reference.md)：查看 Worker、Expert、
  Workflow 与 GoalLoop 的差异和字段约束。
- [资源原生迁移说明](resource-native-migration.md)：了解数据库升级顺序、
  历史数据边界和部署验收。
- [Loop 与 Workflow 产品边界](loop-and-workflow.md)：区分一次性目标闭环和
  可重复自动化任务。

WorkerTemplate、Worker、Prompt、Expert、Workflow 和绑定资源已经接入持久化
Validate/Plan/typed Apply、领域表单与 YAML 单一 Draft。GoalLoop 当前仅支持
schema、Validate 和 Plan；在 typed Apply 接入前，不应把成功 Plan 视为已创建
的 Loop，也不能通过旧 API 模拟资源 Apply。
