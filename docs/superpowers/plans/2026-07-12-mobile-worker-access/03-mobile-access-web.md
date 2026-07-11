# Mobile Access API 与 Web

## Task 6: Mobile Access Descriptor

**Files**

- Modify: `proto/pod/v1/pod.proto`
- Modify: generated Go/TS files
- Create: `backend/internal/api/connect/pod/mobile_access.go`
- Create: `backend/internal/api/connect/pod/mobile_access_test.go`
- Modify: `backend/internal/api/connect/pod/server.go`
- Modify: `backend/internal/config/config_url.go`
- Modify: `clients/web/src/lib/api/connect/podConnect.ts`

### RED

Proto contract:

```protobuf
rpc GetMobileAccessDescriptor(GetMobileAccessDescriptorRequest)
    returns (MobileAccessDescriptor);
```

Backend test：

```go
func TestGetMobileAccessDescriptor_ReturnsTokenFreeCanonicalURL(t *testing.T) {
    resp, err := server.GetMobileAccessDescriptor(ctx, request("acme", "pod-1"))
    require.NoError(t, err)
    assert.Equal(t,
        "https://app.example/acme/mobile/workers/pod-1",
        resp.Msg.CanonicalUrl)
    assert.NotContains(t, resp.Msg.CanonicalUrl, "token=")
}
```

另测 Pod permission、inactive capability、缺失 public URL config。

Run:

```bash
cd backend
go test ./internal/api/connect/pod -run MobileAccess -count=1
```

Expected: FAIL，RPC 尚不存在。

### GREEN

- URL 只来自受控的绝对地址 `PUBLIC_WEB_URL`。未配置时 fail closed，不从
  API/Relay 的 `PRIMARY_DOMAIN` 或请求 Host 推导。
- 返回 Pod、console/preview capability 和 Cloud Relay readiness。
- 不返回 Browser/Preview JWT。
- 生成代码使用仓库 `pnpm proto:gen-go-all`，只保留目标生成物。

`PRIMARY_DOMAIN` 仍是浏览器访问 API、Relay WebSocket 和 Runner tunnel 的
统一入口。`PUBLIC_WEB_URL` 是 Next.js Web origin。生产环境二者可以是同域；
本地开发通常分别为 `:10000` 和 `:10007`。

## Task 7: Mobile Worker Routes and List

**Files**

- Create: `clients/web/src/app/(dashboard)/[org]/mobile/workers/page.tsx`
- Create: `clients/web/src/app/(dashboard)/[org]/mobile/workers/[podKey]/page.tsx`
- Create: `clients/web/src/components/mobile-worker/MobileWorkerList.tsx`
- Create: `clients/web/src/components/mobile-worker/MobileWorkerWorkspace.tsx`
- Create: `clients/web/src/components/mobile-worker/MobileConnectionStatus.tsx`
- Modify: `clients/web/src/components/mobile/PodMobileAccessDialog.tsx`
- Modify: `clients/web/src/components/mobile/MobileBottomNav.tsx`
- Modify: `clients/web/src/components/ide/sidebar/PodListItem.tsx`
- Modify: `clients/web/src/components/ide/sidebar/WorkspaceSidebarContent.tsx`
- Modify: `clients/web/src/components/workspace/TerminalPaneHeader.tsx`
- Modify: `clients/web/src/lib/pod-mobile-access.ts`
- Modify: `clients/web/src/lib/ide-chrome.ts`
- Modify: focused tests for every modified component

### RED

Vitest 覆盖：

- 列表 loading/empty/error/running/offline。
- canonical URL 来自 Descriptor，不读 `window.location.origin`。
- 点击 Worker 进入新 route。
- PTY/ACP 分支继续复用现有 Pane。
- 旧 `/mobile/pods/*` 页面 308/replace 到新路径。
- 手机底栏的 Worker 入口进入 `/mobile/workers`，不再依赖只渲染 PTY 的
  `TerminalSwiper`。
- Worker 行和已打开 Worker 标题栏均有可见手机图标；右键菜单只是补充入口。

Run:

```bash
cd clients/web
../../node_modules/.bin/vitest run --config vitest.config.ts \
  src/components/mobile-worker src/components/mobile/__tests__/PodMobileAccessDialog.test.tsx
```

Expected: 新组件测试 FAIL。

### GREEN

采用现有 Product Baseline：

- 无 hero、无嵌套卡片。
- 列表是密集可扫描行。
- 状态使用图标+文本，不只依赖颜色。
- 触控目标至少 44px。
- 使用现有 tokens、Button、EmptyState 和 Spinner。
- 新 `/mobile/workers/*` 路由隐藏桌面 IDE chrome，并有对应路由测试。

## Task 8: Mobile Workspace State

**Files**

- Create: `clients/web/src/hooks/useWorkerControlLease.ts`
- Create: `clients/web/src/components/mobile-worker/MobileTerminalToolbar.tsx`
- Create: `clients/web/src/components/mobile-worker/MobileAcpWorkspace.tsx`
- Modify: `clients/web/src/stores/relayConnection.ts`
- Modify: `clients/web/src/components/workspace/TerminalPane.tsx`
- Modify: `clients/web/src/components/workspace/AgentPanel.tsx`
- Modify: mobile Preview route/tests

### RED

测试：

- Snapshot 前显示“等待同步”。
- 无租约为 observer，输入控件 disabled。
- acquire 后启用输入。
- control_busy 显示当前被其他设备控制。
- Preview 仍使用 `window.location.replace(session_url)`。

Run:

```bash
cd clients/web
../../node_modules/.bin/vitest run --config vitest.config.ts \
  src/components/mobile-worker \
  src/components/mobile/__tests__/PodMobileAccessDialog.test.tsx \
  'src/app/(dashboard)/[org]/mobile/workers/[podKey]/preview/page.test.tsx'
```

Expected: 新状态和租约用例 FAIL。

### GREEN

控制租约状态来自 Rust relay projection，不建 mobile Zustand SSOT。
终端快捷键工具栏只组合现有 input API，不创建第二条 WebSocket。

## LAN 真机配置

手机不能访问电脑的 `localhost`。假设开发机 LAN IP 为 `192.168.1.169`：

```bash
PRIMARY_DOMAIN=192.168.1.169:10000
PUBLIC_WEB_URL=http://192.168.1.169:10007
USE_HTTPS=false
```

开发启动脚本必须把同一配置投影到四处：

1. Backend 使用 `PUBLIC_WEB_URL` 生成 token-free canonical URL，并把该 origin
   加入 CORS。
2. Web 使用 `PRIMARY_DOMAIN` 生成 `NEXT_PUBLIC_WS_URL`。
3. Relay 把 `PUBLIC_WEB_URL` 加入 WebSocket Origin allowlist。
4. Next dev 将 `PUBLIC_WEB_URL` 的 hostname 加入 `allowedDevOrigins`。

二维码只编码：

```text
http://192.168.1.169:10007/<org>/mobile/workers/<pod-key>
```

它不包含 Browser Token、Preview Token 或一次性 session URL。手机打开页面并
完成登录后，才通过既有 Connect API 获取短期连接凭证。Preview 页面也只调用
后端创建 session，然后用 `window.location.replace(session_url)` 离开，避免
返回页面时重复申请一次性 token。
