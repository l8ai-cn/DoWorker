# 证据索引

本章记录设计结论对应的源码证据。路径是分析时的当前工作树位置。

## 1. lulu 启动与端口

| 证据 | 路径 |
|---|---|
| Gateway 组合和端口 | `/Users/wwyz/Desktop/lulu-codex-web-mac/pc-gateway/src/service.ts` |
| codexapp 启动参数 | `/Users/wwyz/Desktop/lulu-codex-web-mac/pc-gateway/src/codexapp/CodexAppServer.ts` |
| 可选独立 app-server | `/Users/wwyz/Desktop/lulu-codex-web-mac/pc-gateway/src/index.ts` |
| 启动脚本 | `/Users/wwyz/Desktop/lulu-codex-web-mac/scripts/start-web.sh` |
| 历史运行日志 | `/Users/wwyz/Desktop/lulu-codex-web-mac/logs/lulu-codex-web-20260710-175006.log` |

确认端口口径：

- codexapp 源码默认 `5900`
- 2026-07-10 历史日志中的运行实例通过配置覆盖为 `15900`
- Gateway WS `17631`
- 可选 app-server WS `17632`
- UDP discovery `17630`
- QR Portal `17633`
- Cloud Relay `18080`

因此 `5900` 是默认配置，不代表所有运行实例固定使用该端口。

## 2. lulu 二维码和认证

| 结论 | 路径 |
|---|---|
| QR 生成 LAN URL | `pc-gateway/src/qr/QrPortalServer.ts` |
| 桌面注入扫码入口 | `pc-gateway/src/codexapp/CodexAppLogo.ts` |
| 主路径默认无密码 | `pc-gateway/src/codexapp/CodexAppServer.ts` |
| 遗留配对码和 Token | `pc-gateway/src/pairing/PairingManager.ts` |
| Mobile Gateway auth | `pc-gateway/src/mobile-api/MobileGatewayServer.ts` |
| codexapp session auth | `pc-gateway/vendor/codexapp/dist-cli/index.js` |

主二维码只包含 URL，不使用 `17631` 配对 Token。

## 3. lulu 协议和云隧道

| 结论 | 路径 |
|---|---|
| Cloud Client 出站 WS | `pc-gateway/src/cloud/CloudClient.ts` |
| JSON/Base64 消息类型 | `pc-gateway/src/cloud/RelayProtocol.ts` |
| Cloud Relay 转发 | `cloud-relay/src/index.ts` |
| 遗留 Gateway 类型 | `pc-gateway/src/types.ts` |
| JSON-RPC WS Client | `pc-gateway/src/codex-adapter/JsonRpcWsClient.ts` |
| codexapp adapter | `pc-gateway/src/codex-adapter/CodexAppServerAdapter.ts` |

Cloud Relay 代理完整 HTTP/WS；PC 端主动建立持久 WebSocket，再转发到
loopback codexapp。

## 4. lulu 前端和主机能力

| 结论 | 路径 |
|---|---|
| Vue/Vite 包信息 | `pc-gateway/vendor/codexapp/package.json` |
| SPA 构建产物 | `pc-gateway/vendor/codexapp/dist/assets/` |
| HTTP RPC、WS、SSE | `pc-gateway/vendor/codexapp/dist-cli/index.js` |
| Browser Preview | `pc-gateway/src/browser/BrowserPreviewService.ts` |
| 文件、终端和 worktree | `pc-gateway/vendor/codexapp/dist-cli/index.js` |

文件接口、主机 PTY 和 Browser Preview 证明 lulu 暴露的是主机级服务，
不能直接作为 Agent Cloud Pod 安全模型。

## 5. Agent Cloud 移动原型

| 结论 | 路径 |
|---|---|
| 单 Pod 移动工作区 | `clients/web/src/components/mobile/MobilePodWorkspace.tsx` |
| QR 弹窗 | `clients/web/src/components/mobile/PodMobileAccessDialog.tsx` |
| URL 生成 | `clients/web/src/lib/pod-mobile-access.ts` |
| 移动路由 | `clients/web/src/app/(dashboard)/[org]/mobile/pods/[podKey]/` |
| 入口接线 | `clients/web/src/components/ide/sidebar/WorkspaceSidebarContent.tsx` |
| 入口菜单 | `clients/web/src/components/ide/sidebar/SidebarPodContextMenu.tsx` |

## 6. PTY 和 ACP 数据面

| 结论 | 路径 |
|---|---|
| Browser connection API | `backend/internal/api/connect/pod/connection.go` |
| Browser TS facade | `clients/web/src/lib/api/connect/podConnect.ts` |
| TS connection owner | `clients/web/src/stores/relayConnection.ts` |
| Rust pool | `clients/core/crates/relay/src/pool.rs` |
| Rust driver | `clients/core/crates/relay/src/driver/` |
| Relay browser/runner handlers | `relay/internal/server/handler.go` |
| 帧类型 | `relay/internal/protocol/message.go` |
| Runner publisher | `runner/internal/relay/` |

## 7. Preview Tunnel

| 结论 | 路径 |
|---|---|
| Preview API | `backend/internal/api/rest/v1/pod_preview.go` |
| Preview route | `backend/internal/service/relay/preview.go` |
| Token issuer | `backend/internal/service/relay/token.go` |
| Session cookie | `relay/internal/server/handler_preview_session.go` |
| HTTP/WS proxy | `relay/internal/server/handler_preview.go` |
| Tunnel handler | `relay/internal/server/handler_tunnel.go` |
| Tunnel frames | `relay/internal/protocol/tunnelframe/frame.go` |
| Relay multiplexer | `relay/internal/tunnel/tunnel.go` |
| Runner tunnel client | `runner/internal/tunnel/client.go` |
| Runner loopback proxy | `runner/internal/tunnel/local_http.go` |

## 8. 协议与已删除 LAN 路径

| 结论 | 路径 |
|---|---|
| Pod connection proto | `proto/pod/v1/pod.proto` |
| Runner commands | `proto/runner/v1/runner.proto` |

`PodConnectionInfo` 已 reserved 旧 `local_relay_url/local_token` 字段；
Runner heartbeat 和 SubscribePod 也 reserved 本地 Relay 字段。这证明重新
加入 Runner 直连不是页面增量，而是公共协议和安全架构变更。

## 9. 测试证据

使用 Web 正式 Vitest 配置执行：

```text
clients/web/src/components/mobile/__tests__/MobilePodWorkspace.test.tsx
clients/web/src/components/mobile/__tests__/PodMobileAccessDialog.test.tsx
clients/web/src/app/(dashboard)/[org]/mobile/pods/[podKey]/preview/page.test.tsx
clients/web/src/lib/api/__tests__/podPreview.test.ts
```

结果：4 个文件、8 个测试通过。该证据不覆盖真机或真实网络链路。
