# Mobile Worker Access B 级详细设计

| 属性 | 值 |
|---|---|
| 状态 | Approved |
| 日期 | 2026-07-12 |
| 决策 | 云 Relay 优先；局域网 Edge Relay 作为后续阶段 |
| 实施门禁 | Phase 0 全绿后进入 Phase 1 |

用户已批准按推荐决策开始实施。

## 1. 目标

为 Do Worker 提供可发现、可认证、可恢复、可审计的移动端 Worker
接入能力。手机访问的是 Worker 对应的 Pod，不直接进入 Runner 管理面。

本设计覆盖：

- 二维码和移动深链
- 移动 Worker 列表与单 Worker 工作区
- PTY、ACP 和 Preview
- Backend、Runner、Relay 与 Rust Core 协议对齐
- 云端接入和局域网接入
- 鉴权、Token、并发控制和审计
- 数据迁移、API、UI、测试与发布范围

## 2. 文档结构

1. [证据索引](00-evidence-map.md)
2. [现状逆向与差距](01-current-state-and-gap.md)
3. [目标架构与时序](02-target-architecture.md)
4. [协议与安全模型](03-protocol-and-security.md)
5. [前端详细设计](04-frontend-design.md)
6. [后端、Relay 与 Runner 设计](05-backend-relay-runner.md)
7. [Edge Relay 与数据设计](05-edge-relay-and-data.md)
8. [功能清单与逐项逻辑](06-feature-logic.md)
9. [实施文件映射](07-implementation-map.md)
10. [改动范围与测试](07-scope-test-release.md)
11. [发布与实施门禁](08-release-and-gates.md)

## 3. 核心决策

### 3.1 Worker、Pod 与 Runner

- 产品层使用 `Worker`。
- 运行时和协议层继续使用 `Pod`、`pod_key`。
- `Runner` 是承载 Pod 的可信节点，不是手机直接连接的业务资源。
- 手机与桌面必须共用 Pod、Relay 帧和 Rust Core 状态。

### 3.2 接入方式

推荐分两阶段：

- Phase 1：公网 HTTPS/WSS + Cloud Relay，完成生产闭环。
- Phase 2：受 Backend 管理的 Edge Relay，提供显式局域网模式。

不采用 Runner 内置独立 Mobile Gateway。该方案会重新引入已经删除的
`local_relay_url/local_token` 路径，并产生双协议和双鉴权。

### 3.3 二维码

二维码只能包含稳定、无密钥的 HTTPS 深链：

```text
https://<app-domain>/<org>/mobile/workers/<pod-key>
```

禁止写入：

- Browser Relay JWT
- Preview JWT
- Runner Token
- 登录 Session
- LAN 地址或未经验证的端口

### 3.4 无静默降级

云端和局域网是用户可见的连接模式，不互相静默 fallback。
选择的模式失败时必须展示具体错误、失败阶段和重试操作。

## 4. 交付阶段

| 阶段 | 交付内容 | 进入条件 |
|---|---|---|
| Design | 本文档评审 | 当前阶段 |
| Phase 0 | 修复现有阻断项 | 设计批准 |
| Phase 1 | 云端移动接入闭环 | Phase 0 全绿 |
| Phase 2 | Edge Relay 局域网接入 | Phase 1 线上验收 |
| Phase 3 | 离线任务、通知、设备管理 | 独立产品评审 |

## 5. 非目标

- 原生 iOS/Android App
- 匿名 Worker 分享
- 手机直接连接 Runner gRPC
- 复制 lulu 的 `pc-gateway` 或 `cloud-relay`
- 公开主机文件系统、主机 PTY 或任意 URL 浏览器控制
- 在本设计获批前实施任何产品代码
