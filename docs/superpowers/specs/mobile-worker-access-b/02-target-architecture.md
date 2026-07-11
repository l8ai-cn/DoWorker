# 目标架构与时序

## 1. 总体架构

```mermaid
flowchart LR
  Desktop["Desktop Web"] --> Backend["Backend"]
  Phone["Mobile Web / PWA"] --> Backend
  Desktop --> Relay["Cloud or Edge Relay"]
  Phone --> Relay
  Backend <-->|"gRPC bidi + mTLS"| Runner["Runner"]
  Relay <-->|"PTY / ACP binary WS"| Runner
  Relay <-->|"Multiplexed preview tunnel"| Runner
  Runner --> Pod["Worker Pod"]
  Pod --> PTY["PTY / ACP"]
  Pod --> Preview["Loopback preview service"]
```

控制面和数据面必须保持分离：

- Backend 负责认证、授权、连接编排和审计。
- Relay 负责 PTY/ACP 字节与 Preview 流量。
- Runner 只建立出站连接，不暴露公共管理端口。
- 手机不直接访问 Runner gRPC。

## 2. Phase 1 云端接入

```mermaid
sequenceDiagram
  actor User
  participant Desktop
  participant Phone
  participant Backend
  participant Relay
  participant Runner
  participant Pod

  User->>Desktop: 打开“手机接入”
  Desktop->>Backend: GetMobileAccessDescriptor(podKey)
  Backend-->>Desktop: token-free deep link + capabilities
  Desktop-->>User: 展示二维码
  User->>Phone: 扫码
  Phone->>Backend: 登录并恢复 redirect
  Backend-->>Phone: Worker summary
  Phone->>Backend: GetPodConnection
  Backend->>Runner: SubscribePod
  Runner->>Relay: publisher connect
  Backend-->>Phone: relay URL + short-lived browser JWT
  Phone->>Relay: subscriber connect
  Relay-->>Phone: Snapshot / AcpSnapshot
  Phone->>Relay: Input / Resize / AcpCommand
  Relay->>Runner: binary frames
  Runner->>Pod: PTY or ACP operation
```

`SubscribePod` 失败时 Backend 必须返回 `Unavailable`，不得签发 Browser
Token。手机页面显示 Runner 未连接或 Relay 建链失败。

## 3. Preview 时序

```mermaid
sequenceDiagram
  participant Phone
  participant Backend
  participant Relay
  participant Runner
  participant App as Pod Preview App

  Phone->>Backend: CreatePreviewSession(podKey)
  Backend->>Runner: EnsureTunnelConnected
  Runner->>Relay: /runner/tunnel typed JWT
  Relay-->>Backend: tunnel ready acknowledgement
  Backend-->>Phone: session_url only
  Phone->>Relay: GET __session?token=...
  Relay-->>Phone: Set-Cookie HttpOnly + 302
  Phone->>Relay: GET /preview/{podKey}/...
  Relay->>Runner: REQ_START/BODY/END
  Runner->>App: loopback HTTP/WS
  App-->>Runner: response
  Runner-->>Relay: RESP_START/BODY/END
  Relay-->>Phone: response
```

设计要求：

- Backend 只有在 tunnel ready 后才签发 Preview session。
- API 响应不返回独立 `token` 字段。
- 手机只跳转 `session_url`。
- Preview Cookie 限定 Pod path。
- Runner tunnel 必须维护重连状态机。

## 4. Phase 2 局域网 Edge Relay

```mermaid
flowchart LR
  Phone["Phone"] -->|"HTTPS / WSS"| Edge["Edge Relay"]
  Backend["Backend"] -->|"注册、授权、健康状态"| Edge
  Runner["Runner"] -->|"同一 Relay protocol"| Edge
  Edge --> Phone
```

Edge Relay 不是 Runner 内置 Gateway。它必须：

- 运行标准 Relay 实现
- 使用 Backend 签发的 typed JWT
- 使用 HTTPS/WSS 和可信证书
- 注册健康状态与可达地址
- 使用相同 PTY/ACP 和 Tunnel 帧
- 由用户显式选择“局域网”

Backend 返回连接候选，但不静默切换：

```json
{
  "mode": "cloud",
  "candidates": [
    {"id": "relay-cloud-1", "kind": "cloud", "status": "ready"},
    {"id": "relay-edge-7", "kind": "edge", "status": "ready"}
  ]
}
```

用户选择 Edge 后，如果手机无法到达，页面显示：

- DNS/TLS 失败
- Relay 不健康
- 手机不在允许网络
- Runner 未连接到该 Edge

## 5. 多端控制

默认采用单写者租约：

```mermaid
stateDiagram-v2
  [*] --> Observe
  Observe --> Requesting: 请求控制
  Requesting --> Controller: 获得租约
  Requesting --> Observe: 被拒绝
  Controller --> Observe: 释放或超时
  Controller --> Observe: 连接断开
```

- 所有客户端可观察输出。
- 只有租约持有者可发送 Input、Resize 和 ACP Command。
- 手机接管时桌面显示明确提示。
- 危险 ACP permission 仍按现有 RBAC 和审批逻辑执行。
