# 实施文件映射

本章是评审用预计范围。设计批准后应先同步当前 HEAD，再生成最终实施计划。

## 1. Phase 0

| 文件 | 动作 | 目标 |
|---|---|---|
| `backend/internal/api/connect/pod/connection.go` | 修改 | SubscribePod fail-closed |
| `backend/internal/api/connect/pod/connection_test.go` | 修改 | 缺失 sender、Runner offline、dispatch error |
| `backend/internal/api/rest/v1/pod_preview.go` | 修改 | tunnel fail-closed，删除裸 Token |
| `backend/internal/api/rest/v1/pod_preview_test.go` | 修改 | session response 和失败路径 |
| `backend/internal/service/relay/preview.go` | 修改 | path normalize 和 target contract |
| `relay/internal/server/handler_preview.go` | 修改 | 上游 path join |
| `relay/internal/server/handler_preview_test.go` | 修改 | HTTP/WS/path traversal |
| `runner/internal/tunnel/client.go` | 拆分修改 | 持久重连状态机 |
| `runner/internal/tunnel/client_reconnect.go` | 新增 | backoff、jitter、generation |
| `runner/internal/tunnel/client_reconnect_test.go` | 新增 | 断线和 Token 刷新 |

## 2. Access Descriptor

预计新增：

```text
backend/internal/service/mobileaccess/service.go
backend/internal/service/mobileaccess/descriptor.go
backend/internal/service/mobileaccess/service_test.go
backend/internal/api/connect/pod/mobile_access.go
backend/internal/api/connect/pod/mobile_access_test.go
```

预计协议：

```text
proto/pod/v1/pod.proto
proto/gen/go/pod/v1/*
proto/gen/ts/pod/v1/*
clients/web/src/lib/api/connect/podConnect.ts
```

职责：

- canonical URL
- Worker capability
- Cloud/Edge candidate
- RBAC 和 Org scope
- 不签发连接 Token

## 3. PreviewConfig

预计修改：

```text
backend/internal/domain/agentpod/pod.go
backend/internal/domain/agentpod/pod_config_revision.go
backend/internal/service/agentpod/pod_service.go
backend/internal/service/agentpod/settings_service.go
backend/internal/infra/agentpod_repo.go
proto/pod/v1/pod.proto
clients/web/src/components/pod/CreatePodForm/*
```

如需 schema 变更：

```text
backend/migrations/<next>_pod_preview_config.up.sql
backend/migrations/<next>_pod_preview_config.down.sql
backend/migrations/pod_preview_config_postgres_test.go
```

PreviewConfig 必须走 Pod config revision，不能新增一条只更新 Pod 表的旁路。

## 4. 控制租约

预计修改：

```text
relay/internal/channel/channel.go
relay/internal/channel/channel_io.go
relay/internal/channel/channel_manager.go
relay/internal/protocol/message.go
clients/core/crates/relay/src/dispatch.rs
clients/core/crates/relay/src/pool.rs
clients/core/crates/wasm/src/relay_manager.rs
clients/web/src/stores/relayConnection.ts
```

预计新增：

```text
relay/internal/channel/control_lease.go
relay/internal/channel/control_lease_test.go
clients/core/crates/relay/src/control_lease.rs
clients/web/src/hooks/useWorkerControlLease.ts
```

不新增第二个移动 Relay store。

## 5. 移动前端

预计新增：

```text
clients/web/src/app/(dashboard)/[org]/mobile/workers/page.tsx
clients/web/src/app/(dashboard)/[org]/mobile/workers/[podKey]/page.tsx
clients/web/src/app/(dashboard)/[org]/mobile/workers/[podKey]/preview/page.tsx
clients/web/src/components/mobile-worker/MobileWorkerList.tsx
clients/web/src/components/mobile-worker/MobileWorkerWorkspace.tsx
clients/web/src/components/mobile-worker/MobileConnectionStatus.tsx
clients/web/src/components/mobile-worker/MobileTerminalToolbar.tsx
clients/web/src/components/mobile-worker/MobileAcpWorkspace.tsx
clients/web/src/components/mobile-worker/MobileAccessDialog.tsx
clients/web/src/components/mobile-worker/MobileAccessModeSelector.tsx
clients/web/src/hooks/useMobileAccessDescriptor.ts
```

预计修改：

```text
clients/web/src/components/ide/sidebar/PodListItem.tsx
clients/web/src/components/ide/sidebar/SidebarPodContextMenu.tsx
clients/web/src/components/ide/sidebar/WorkspaceSidebarContent.tsx
clients/web/src/components/workspace/TerminalPane.tsx
clients/web/src/components/workspace/AgentPanel.tsx
clients/web/src/lib/ide-chrome.ts
clients/web/src/messages/en/app.json
clients/web/src/messages/zh/app.json
```

实施时应评估是否迁移或删除现有 `components/mobile/MobilePodWorkspace.tsx`
和 `PodMobileAccessDialog.tsx`，禁止保留两套等价页面。

## 6. 旧路由迁移

推荐：

```text
/{org}/mobile/pods/{podKey}
  -> 308 /{org}/mobile/workers/{podKey}
```

- Console 和 Preview 旧路由都只做重定向。
- 不复制页面、状态或 API。
- 保留至少一个生产发布周期。
- 记录旧路径命中量，不记录 Token 或用户内容。
- 连续 30 天零命中并发布迁移公告后才能删除。

## 7. Preview API 去重

`clients/web` 和 `clients/web-user` 不能继续各自解释 wire response。

优先方案：

- 把 Preview response type 和 fetch contract 放到共享 API client。
- 两个前端只保留各自呈现 hook。

禁止复制 Token refresh、错误码映射和 session URL 解析。

## 8. Phase 2

预计范围：

```text
backend/internal/service/relay/*
backend/internal/api/rest/internal/relay_registration*
relay/internal/config/*
deploy/kubernetes/<environment>/*relay*
backend/migrations/<next>_edge_relay_metadata.*
```

不修改 Runner 为其增加本地 Mobile HTTP server。

## 9. 明确排除

- lulu 源码和 vendored codexapp
- 独立 `pc-gateway`
- 独立 MobileGateway 协议
- 手机直连 Runner gRPC
- 匿名分享 Token
- 原生 App
- 主机文件系统和主机 Shell API
- 与移动接入无关的 AI Resource、Marketplace 和 Workflow 代码
