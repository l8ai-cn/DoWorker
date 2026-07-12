# 发布与实施门禁

## 1. 发布流程

按 GitOps 闭环：

1. 合并协议和 migration。
2. 发布 Backend/Relay。
3. 发布 Runner。
4. 发布 Web/Rust WASM。
5. 执行 migration 状态检查。
6. 验证 Relay 和 tunnel health。
7. 执行桌面与真机 smoke。
8. 检查日志、指标和错误率。
9. 保留版本化回滚配置。
10. 运行 `deploy/kubernetes/cluster-oilan/verify-mobile-worker-access.sh`。

协议发布顺序必须保证新 Backend 不向旧 Runner 发送其无法理解的必需命令。
不允许手工修改线上数据库或临时注入 Preview 配置。

发布 smoke 必须先验证移动入口 HTTPS、认证、Codex `["acp","pty"]` 声明、
`requires_model_resource=true` 和唯一默认模型资源。测试组织中以
`MOBILE_SMOKE_RUN_INTERACTIONS=true` 运行时，还必须创建并清理 ACP/PTY
临时会话，验证 ACP 回复、PTY Relay token 和 ACP Relay 拒绝。

## 2. 灰度

建议 Feature Flag：

- `mobile_worker_access`
- `mobile_worker_preview`
- `worker_control_lease`
- `edge_relay_access`

Flag 只控制功能开放，不允许绕过协议校验、安全检查或 migration。

灰度顺序：

1. 内部组织
2. 测试组织
3. 5% 生产组织
4. 25%
5. 100%

每阶段观察连接成功率、Snapshot 延迟、Tunnel 错误和客户端异常。

## 3. 回滚

Web 回滚后新 API 可保留，但不能破坏旧客户端。

Backend/Relay/Runner 回滚必须满足：

- 新增 proto 字段保持向后兼容
- DB migration down 已真实验证
- 控制租约关闭后恢复为明确的旧并发行为
- PreviewConfig 不因回滚丢失
- Edge Relay 可独立从候选列表下线

严重安全问题应先关闭 Feature Flag，再执行版本回滚。

## 4. 原子提交

建议拆分：

1. `fix(relay): fail closed on pod subscription`
2. `fix(preview): complete config and path contract`
3. `fix(runner): reconnect preview tunnel`
4. `feat(relay): add worker control lease`
5. `feat(web): add mobile worker access flow`
6. `test(e2e): cover mobile worker access`
7. `docs: publish mobile access operations guide`

每个提交在进入下一提交前完成对应定向测试。

## 5. 设计批准后的实施门禁

开始实现前必须确认：

- 采用 Cloud-first + Edge Relay 方案
- 是否批准旧 `/mobile/pods` 路由的“一个发布周期且连续 30 天零命中后回收”
  兼容规则
- Phase 1 是否包含控制租约
- PreviewConfig 是否允许运行中更新
- 真机和线上测试环境

这些是架构、兼容、安全和发布决策，不能由实现者静默选择。

## 6. Phase 1 完成定义

只有以下证据全部成立才算完成：

- Phase 0 阻断项关闭
- 单元、协议和 Rust 测试通过
- iPhone Safari 与 Android Chrome 主路径通过
- PTY、ACP 和 Preview 真链路通过
- Runner/Relay/Tunnel 断线恢复通过
- 桌面与手机并发控制通过
- 生产灰度健康指标达标
- 文档、迁移和回滚说明已发布

仅有组件测试、页面截图或本地 localhost 演示不构成完成。
