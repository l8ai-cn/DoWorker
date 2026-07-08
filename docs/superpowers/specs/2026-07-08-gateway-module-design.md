# Gateway 模块详细设计

| 属性 | 值 |
|------|-----|
| **状态** | Draft（待评审） |
| **创建日期** | 2026-07-08 |
| **目标** | 将 relay 演进为统一 Gateway：内网隧道通道 + 所有 worker agent 的 HTTP/WS/媒体资源经 Gateway 提供给前端 |
| **参考** | doops.sh 网关隧道（`tunnel_hub.go` / `cmd/agent/main.go`）、RFC-006、2026-07-08 worker-per-runner-pod 设计 |

---

## 1. 背景与现状审核结论

### 1.1 已有能力（复用，不重写）

| 能力 | 位置 | 结论 |
|------|------|------|
| 终端/ACP WebSocket 隧道 | `relay/internal/channel/`、`relay/internal/protocol/message.go` | podKey 配对 + 双向转发，测试充分，**保持不动** |
| Relay 注册/心跳/ACME 证书 | `relay/internal/backend/client.go`、`backend/internal/api/rest/internal/relay_*.go` | Gateway 复用同一注册通道，**不新增注册协议** |
| Token 签发 | `backend/internal/service/relay/token.go`（HS256，claims 含 `pod_key/runner_id/user_id/org_id`） | 扩展 claims 增加 `token_type`，向后兼容 |
| Runner 出站 WS 客户端 | `runner/internal/relay/client*.go`（重连、token 刷新、backoff） | 隧道客户端按同一模式实现，复用 `safego`/重连骨架 |
| 命令下发 | `proto/runner/v1/runner.proto` `ServerMessage` oneof（`subscribe_pod = 8`，最新编号 13） | 新命令占用 `14` |
| 反代入口 | `deploy/dev/traefik/dynamic/http.yml`（`/relay` strip）、`deploy/kubernetes/cluster-oilan/40-ingress.yaml` | 新增 `/preview` 路由（不 strip） |

### 1.2 审核发现的问题（本设计一并修复）

| # | 问题 | 位置 | 处理 |
|---|------|------|------|
| A1 | WS `CheckOrigin` 恒真 | `relay/internal/server/handler.go:21`、`backend/.../terminal_attach.go:16` | Phase 1 增加 Origin 白名单（`PRIMARY_DOMAIN` 派生 + 可配置） |
| A2 | token 类型仅靠 `UserID==0` 区分 | `relay/internal/auth/token.go` | 增加显式 `token_type` claim；空值按旧规则回退 |
| A3 | `SendSubscribePod` 失败仅 Warn，客户端仍拿到 token | `backend/.../connect/pod/connection.go:60-62` | Preview 路径下发失败返回明确错误码，不发 token |
| A4 | doops `HandleAgentConnect` 无 token 校验（其 bug hunt P0-1） | 参考实现缺陷 | 隧道连接**必须**校验 backend 签发的 tunnel token |
| A5 | doops 背压满时静默丢帧（P1-4）、单操作超时关整条隧道（P1-3） | 参考实现缺陷 | 采用 credit 流控 + 按 stream 关闭，不关整条隧道 |

### 1.3 从 doops.sh 借鉴的机制

| doops 机制 | 本设计取舍 |
|---|---|
| agent 主动出站连公网 gateway，`cluster/instance` 注册 | **采用**：runner 主动连 `WS /runner/tunnel`，以 `runner_id` 注册，流按 `pod_key` 路由 |
| `waitForAgent` 100ms 轮询 + 10s 宽限 | **采用**（缩短为 5s）：隧道闪断时 preview 请求短暂等待而非立即 502 |
| `opSlot` 每目标串行 + `MaxQueuedPerTarget` 有限排队 | **改造**：preview 是读流量，不需要串行；改为每隧道最大并发 stream 数 + 每 pod 排队上限 |
| 请求/响应按 `id` 映射（`pending map[int64]chan`） | **改造**：升级为带 `stream_id` 的二进制多路复用帧协议（支持流式 body，doops 是 JSON 整包） |
| 审计（user×target×action，bytes in/out） | **采用**：结构化访问日志 + OTel 指标，暂不建独立审计库 |
| 独立用户/token/grants 体系 | **不采用**：鉴权 SSOT 保持在 backend（org/pod policy） |

---

## 2. 总体架构

```
                       ┌─────────────────────────────────────────────┐
  Browser              │ Gateway（relay 进程演进，同一二进制）          │
  ────────────────────►│                                             │
  WS  /browser/relay   │  channel/   终端+ACP（现状，不动）            │
  ANY /preview/:podKey │  tunnel/    隧道注册表 + 多路复用帧            │
       │               │  proxy/     HTTP/WS 反向代理 + 流控           │
       │               └──────┬──────────────────▲───────────────────┘
       │ REST/Connect         │ 内部API           │ WS /runner/tunnel（出站）
       ▼                      ▼ 注册/心跳          │ WS /runner/relay（出站，现状）
  ┌──────────┐  gRPC双向流  ┌─┴────────┐          │
  │ Backend  │─────────────►│  Runner  │──────────┘
  │ (SSOT)   │ connect_tunnel│          │ HTTP 127.0.0.1:port
  └──────────┘  subscribe_pod└──────────┘──────────► Worker 本地服务
                                                    (dev server / 静态资源 / 视频)
```

三个平面：

1. **控制面**（现状不变）：Backend ↔ Runner gRPC/mTLS；Backend 签发所有 token；Gateway 向 Backend 注册/心跳。
2. **终端数据面**（现状不变）：`/runner/relay` + `/browser/relay`，channel 按 podKey 配对。
3. **HTTP 数据面**（新增）：每个 Runner 一条 `/runner/tunnel` 长连接，内部按 `stream_id` 多路复用任意 HTTP/WS 请求；Gateway 对外暴露 `/preview/{podKey}/*`。

关键决策：

- **隧道粒度 = Runner**（不是 pod）。一条 WS 承载该 runner 上所有 pod 的 preview 流量，减少连接数；`pod_key` 在每个 stream 的头帧里。
- **路由依据 = JWT claims，不查表**。preview token / session cookie 内含 `pod_key + runner_id`（backend 签发），Gateway 无需回源 backend 就能定位隧道，与现有 relay token 模式一致。
- **目标端口由 Backend 决定**。`REQ_START` 帧携带 `target`（如 `127.0.0.1:3000`），来源是 pod 元数据；Runner 只校验 target 是 loopback，防 SSRF。

---

## 3. 代码目录设计

### 3.1 Gateway 服务（`relay/` 内演进，暂不改模块名）

Bazel、k8s manifests、traefik 均引用 `relay`，改名收益低风险高；目录内按职责新增包，未来整体 `git mv` 为 `gateway/` 是纯机械操作。

```
relay/
├── cmd/relay/main.go                     # 不变
├── internal/
│   ├── config/config.go                  # [改] 新增 tunnel/proxy/origin 配置段
│   ├── auth/
│   │   ├── token.go                      # [改] TokenClaims 增加 TokenType；ValidateTyped()
│   │   └── origin.go                     # [新] Origin 白名单校验（~60 行）
│   ├── server/
│   │   ├── server.go                     # [改] 挂载新路由：/runner/tunnel、/preview/
│   │   ├── handler.go                    # [改] 现终端 WS handler；接入 origin 校验
│   │   ├── handler_tunnel.go             # [新] Runner 隧道 WS 接入（~120 行）
│   │   ├── handler_preview.go            # [新] /preview/{podKey}/* HTTP 入口（~200 行）
│   │   └── handler_preview_session.go    # [新] token→cookie 交换端点（~80 行）
│   ├── channel/                          # 不变（终端/ACP）
│   ├── protocol/
│   │   ├── message.go                    # 不变（终端协议 0x01–0x0D）
│   │   └── tunnelframe/frame.go          # [新] 隧道帧编解码（独立子包，~150 行）
│   ├── tunnel/
│   │   ├── registry.go                   # [新] runnerID → *Tunnel；WaitForTunnel 宽限（~120 行）
│   │   ├── tunnel.go                     # [新] 单条隧道连接：读写循环、心跳、stream 表（~250 行）
│   │   ├── stream.go                     # [新] Stream 状态机 + credit 流控（~200 行）
│   │   └── limits.go                     # [新] 并发/排队/超时限制（~80 行）
│   ├── proxy/
│   │   ├── http.go                       # [新] HTTP 请求 ↔ stream 帧转换、流式回写（~250 行）
│   │   ├── websocket.go                  # [新] WS 升级透传（~150 行）
│   │   └── headers.go                    # [新] hop-by-hop 过滤、X-Forwarded-*（~80 行）
│   ├── backend/client.go                 # [改] 心跳上报增加 tunnel 统计字段
│   └── otel/                             # [改] 新增隧道/代理指标
```

### 3.2 Backend

```
backend/internal/service/relay/
│   ├── token.go                          # [改] GenerateTypedToken(..., tokenType)；TokenClaims.TokenType
│   └── preview.go                        # [新] ResolvePreviewRoute(podKey)：active/权限/target 校验（~100 行）
backend/internal/domain/agentpod/
│   └── pod.go                            # [改] 新增 PreviewPort int（0=禁用）、PreviewPath string
backend/migrations/
│   └── 0001XX_pod_preview.up.sql         # [新] pods 表加 preview_port/preview_path
backend/internal/api/rest/v1/
│   └── pod_preview.go                    # [新] GET /orgs/:slug/pods/:key/preview → preview_url + token（~120 行）
│                                         #      PATCH .../preview → 设置 preview_port
backend/internal/api/connect/pod/
│   └── preview.go                        # [新] GetPodPreviewInfo（Connect 对等实现，~80 行）
backend/internal/service/runner/
│   └── command_sender.go                 # [改] SendConnectTunnel(runnerID, gatewayURL, tunnelToken)
backend/cmd/server/services_init.go       # [改] Runner initialized 回调追加下发 connect_tunnel
```

### 3.3 Proto

```
proto/runner/v1/runner.proto
    ServerMessage oneof 新增:
        ConnectTunnelCommand connect_tunnel = 14;
    新消息:
        message ConnectTunnelCommand {
          string gateway_url = 1;   // 与 relay_url 同源，例如 wss://domain/relay
          string tunnel_token = 2;  // token_type=tunnel 的 JWT
        }
    RunnerMessage 复用 RequestRelayTokenEvent 语义新增:
        RequestTunnelTokenEvent request_tunnel_token = <next>;
```

### 3.4 Runner

```
runner/internal/tunnel/
│   ├── client.go                         # [新] 隧道 WS 客户端：复用 relay client 的重连/backoff 模式（~250 行）
│   ├── dispatcher.go                     # [新] 帧→stream 分发；stream 生命周期（~200 行）
│   └── local_http.go                     # [新] 请求本地服务：流式转发、loopback 校验、WS 拨号（~200 行）
runner/internal/runner/
│   └── message_handler_tunnel.go         # [新] OnConnectTunnel：建立/更新隧道（~100 行，对照 message_handler_relay.go）
runner/internal/config/config.go          # [改] RewriteRelayURL 同样作用于 gateway_url（dev 环境）
```

### 3.5 前端 / 部署

```
clients/web-user/src/hooks/usePodPreview.ts   # [新] 拉取 preview 信息，管理 session 建立
clients/web-user/src/components/PreviewPanel.tsx  # [新] iframe + sandbox + 刷新/新窗口打开
deploy/dev/traefik/dynamic/http.yml           # [改] 新增 router: PathPrefix(`/preview`) → relay（不 strip）
deploy/kubernetes/cluster-oilan/40-ingress.yaml  # [改] 同上
```

---

## 4. 功能详细设计

### 4.1 隧道帧协议（`protocol/tunnelframe`）

单条 WS 二进制帧格式：`[1B frame_type][4B stream_id BE][payload]`。`stream_id=0` 保留给连接级消息。stream_id 由 Gateway 侧分配（单调递增 uint32），Runner 侧只回应。

| type | 名称 | 方向 | payload |
|------|------|------|---------|
| 0x01 | TUNNEL_HELLO | R→G | JSON `{runner_id, org_id, version, capabilities:[...]}` |
| 0x02 | TUNNEL_PING / 0x03 PONG | 双向 | 空 |
| 0x10 | REQ_START | G→R | JSON `{method, path, query, headers, pod_key, target, content_length, is_websocket}` |
| 0x11 | REQ_BODY | G→R | 原始 chunk（≤256KB，对齐 doops `gitTunnelChunkBytes`） |
| 0x12 | REQ_END | G→R | 空 |
| 0x13 | STREAM_CANCEL | 双向 | JSON `{code, reason}`，只关本 stream |
| 0x20 | RESP_START | R→G | JSON `{status, headers}`（Range 场景即 206 + Content-Range 原样） |
| 0x21 | RESP_BODY | R→G | 原始 chunk |
| 0x22 | RESP_END | R→G | 空 |
| 0x23 | RESP_ERROR | R→G | JSON `{code: "target_unreachable"\|"target_forbidden"\|..., message}` |
| 0x30 | WS_DATA | 双向 | 升级成功后（RESP_START status=101）承载子 WS 帧 |
| 0x31 | WS_CLOSE | 双向 | JSON `{code, reason}` |
| 0x40 | CREDIT | 双向 | 4B uint32，接收方消费后追加发送窗口 |

**流控（修复 doops P1-4）**：每 stream 每方向初始窗口 1 MiB。发送方按 chunk 大小扣减，窗口耗尽即停读上游（Runner 停读本地 HTTP body / Gateway 停读客户端 body）；接收方把数据 flush 到目的端后发 CREDIT 补窗。内存上界 = `并发 stream 数 × 2 MiB`，与文件大小无关。

**错误隔离（修复 doops P1-3）**：任何 stream 超时/取消只发 `STREAM_CANCEL` 关闭该 stream；隧道连接仅在心跳超时（3×10s 无 PONG）或读写致命错误时整条断开。单个 preview 请求超时绝不影响同 runner 其他流量。

### 4.2 隧道注册表（`tunnel/registry.go`）

```go
type Registry struct {
    mu       sync.RWMutex
    tunnels  map[int64]*Tunnel        // runnerID -> tunnel
    waiters  map[int64][]chan *Tunnel // runnerID -> 等待宽限期内重连的请求
}

func (r *Registry) Register(t *Tunnel)                    // 同 runnerID 重连：关旧连接、迁移前唤醒 waiters
func (r *Registry) Get(runnerID int64) *Tunnel
func (r *Registry) WaitForTunnel(ctx, runnerID int64, grace time.Duration) *Tunnel // 借鉴 doops waitForAgent，grace=5s
func (r *Registry) Unregister(t *Tunnel)
func (r *Registry) Stats() RegistryStats                  // 供心跳/otel
```

- `Register` 幂等：新连接以 `connected_at` 为准，旧连接收 `TUNNEL_HELLO` 冲突后被关闭（对齐 doops `registerAgent` 接管语义，但显式关闭旧连接并 drain 其 stream）。
- `WaitForTunnel`：先查在线；未命中则 100ms 轮询直到 `grace` 或 ctx 取消。用于 runner 重连瞬间的 preview 请求。

### 4.3 Preview HTTP 入口（`server/handler_preview.go`）

路由 `ANY /preview/{podKey}/*rest`。处理链：

1. **鉴权**：优先读 `gw_preview` cookie（session 模式），回退 `?token=`（首帧或直链）。校验 `token_type=preview`、`pod_key` 与 path 一致、未过期。失败 401。
2. **定位隧道**：从 claim 取 `runner_id`，`registry.WaitForTunnel(ctx, runnerID, 5s)`。未命中 502 `{code:"target_offline"}`。
3. **配额**：`limits.Acquire(podKey)`——每 pod 最大并发 stream（默认 32），超限进队列（每 pod 上限 16，等待 5s），再满返回 429 `target_busy`。
4. **构造 REQ_START**：`target` 来自 token claim 的 `preview_target`（backend 从 pod 元数据签入，Gateway 不接受客户端指定 target，防 SSRF）；`path` = `/`+rest；透传经 `proxy/headers.go` 过滤后的 header，追加 `X-Forwarded-For/Proto/Host`。
5. **流式双向**：见 4.1 credit 流控。若 `RESP_START.status==101` 转 WS 透传（4.4）。
6. **收尾**：`RESP_END`/`STREAM_CANCEL` 后释放配额，记录访问日志（podKey、path、status、bytes、耗时）。

**Session 模式**（`handler_preview_session.go`）：`GET /preview/{podKey}/__session?token=<preview-jwt>` 校验后下发 `HttpOnly; Secure; SameSite=Lax; Path=/preview/{podKey}` 的短期 cookie（值为同一 JWT 或其派生），再 302 到 `previewBaseUrl`。目的：iframe/子资源请求不必在每个 URL 带长 token，避免 token 泄漏进 referer/日志。

### 4.4 WebSocket 透传（`proxy/websocket.go`）

- 客户端在 `/preview/{podKey}/ws/*` 发起 Upgrade → Gateway 升级客户端连接 → 发 `REQ_START{is_websocket:true}`。
- Runner 拨号本地 WS，成功回 `RESP_START{status:101}`。
- 之后两侧的 WS 帧封进 `WS_DATA`（保留 opcode，二进制/文本区分），任一侧关闭发 `WS_CLOSE`。
- 复用 credit 流控，防止慢消费者撑爆内存。

### 4.5 媒体与大文件

- **图片/HTML/CSS/JS**：普通流式，`RESP_START` 原样透传 `Content-Type`。
- **视频/Range**：Gateway 透传客户端 `Range` 头；Runner 的本地 HTTP 客户端不改写，`206 + Content-Range + Accept-Ranges` 原样回。分块 ≤256KB，配合 credit，实现边下边发。
- **压缩**：`Content-Encoding` 原样透传，Gateway/Runner 都不解压重压。
- **SSE**：识别 `text/event-stream`，禁用缓冲，每 `RESP_BODY` 立即 flush。

### 4.6 Backend 侧逻辑

- `ResolvePreviewRoute(ctx, podKey)`：校验 pod `IsActive()`、调用方对 pod 有读权限（复用 `policy.PodPolicy.AllowRead`）、`PreviewPort>0`；返回 `{runnerID, target:"127.0.0.1:"+port, path}`。任一不满足返回对应错误码。
- `GET /orgs/:slug/pods/:key/preview`：调用 `ResolvePreviewRoute` → `GenerateTypedToken(podKey, runnerID, userID, orgID, tokenType="preview", preview_target, 30min)` → 返回 `{previewBaseUrl:"https://domain/preview/"+podKey+"/", token, sessionUrl, expiresAt}`。**下发 connect_tunnel 失败则返回 503，不返回 token（修复 A3）**。
- Runner `initialized` 回调（`services_init.go`）：对支持隧道能力的 runner 下发 `ConnectTunnelCommand`，token_type=`tunnel`，claim 只含 `runner_id/org_id`（不绑定单一 pod）。

### 4.7 Runner 侧逻辑

- `OnConnectTunnel`（对照 `message_handler_relay.go` 的锁策略）：已连同一 gateway_url 则 `UpdateToken`；否则新建 `tunnel.Client`，`Connect()`→`Start()`，原子替换。
- `dispatcher`：收 `REQ_START` 建 stream 起 goroutine；`local_http.go` 校验 `target` 是 loopback（`127.0.0.0/8`/`::1`），拒绝其余（返回 `RESP_ERROR{target_forbidden}`），然后 `http.Client` 无重定向跟随地请求本地服务，流式回帧。
- token 过期：复用 `RequestTunnelTokenEvent` 向 backend 要新 token，与 relay 现有 `SetTokenExpiredHandler` 同构。

---

## 5. 端到端时序（preview 首屏）

```
Browser            Backend           Gateway            Runner        LocalSvc
  │ GET pods/:key/preview             │                  │
  ├──────────────►│ ResolvePreviewRoute                  │
  │               │ (active/perm/port)                   │
  │◄──────────────┤ {previewBaseUrl, token, sessionUrl}  │
  │ GET /preview/:key/__session?token │                  │
  ├──────────────────────────────────►│ 校验→Set-Cookie 302
  │◄──────────────────────────────────┤                  │
  │ GET /preview/:key/  (cookie)      │                  │
  ├──────────────────────────────────►│ WaitForTunnel    │
  │                                    ├─REQ_START/BODY/END►│ GET 127.0.0.1:port/
  │                                    │                  ├──────────────►│
  │                                    │◄RESP_START/BODY◄─┤◄──────────────┤
  │◄════ 流式 200 + body ═════════════┤ (credit 流控)     │
```

runner 未连隧道时：Backend `initialized` 回调已下发 `connect_tunnel`；若请求早于隧道建立，`WaitForTunnel` 宽限 5s 覆盖重连窗口，仍失败则 502。

---

## 6. 错误处理

| 场景 | 返回 | 说明 |
|------|------|------|
| token 无效/过期/pod 不匹配 | 401 | 不泄漏 pod 是否存在 |
| 无 pod 读权限 | 403 | backend 决策 |
| pod 未启用 preview / port=0 | 404 `preview_disabled` | |
| runner 隧道离线（含宽限后） | 502 `target_offline` | |
| 本地服务拒连/超时 | 502 `target_unreachable` | Runner 回 RESP_ERROR |
| target 非 loopback | 502 `target_forbidden` | Runner 侧防 SSRF |
| 每 pod 并发/队列打满 | 429 `target_busy` | 借鉴 doops 队列，明确错误不静默 |
| 单 stream 超时（默认 300s，视频可配长） | STREAM_CANCEL | 只关本流，不断隧道 |
| 心跳超时 / 隧道断开 | 该 runner 全部 in-flight stream 收 502 | registry 清理 + 唤醒 waiters |

---

## 7. 安全

- **三类 token 分离**：`token_type ∈ {browser, runner, tunnel, preview}`；各端点只接受对应类型（修复 A2）。tunnel token 不绑 pod、preview token 强绑 pod+target。
- **Origin 白名单**（修复 A1）：`auth/origin.go` 依据 `PRIMARY_DOMAIN` + 配置项校验 WS `Origin`，终端与隧道端点统一接入。
- **SSRF 防护**：target 只能由 backend 经 token 注入且必须 loopback；Runner 二次校验。
- **Header 卫生**：`proxy/headers.go` 剥离 hop-by-hop（`Connection/Upgrade/Keep-Alive/Proxy-*/Transfer-Encoding/TE/Trailer`）与入站 `X-Forwarded-*` 后重建。
- **iframe 隔离**：`PreviewPanel` 用 `sandbox`；响应注入 `Content-Security-Policy`、`X-Frame-Options` 由 preview 子域策略统一（后续可迁独立子域）。
- **限额**：请求头 ≤64KB、body 无上限但受 credit 背压、单 stream 时长可配、每 org 隧道并发上限。

---

## 8. 配置

Gateway（`RELAY_` 前缀，沿用 viper 映射）：

| 变量 | 默认 | 说明 |
|------|------|------|
| `TUNNEL_ENABLED` | true | 是否开启 HTTP 隧道平面 |
| `TUNNEL_MAX_STREAMS_PER_POD` | 32 | 每 pod 并发 stream |
| `TUNNEL_QUEUE_PER_POD` | 16 | 每 pod 排队上限 |
| `TUNNEL_QUEUE_TIMEOUT` | 5s | 排队等待超时 |
| `TUNNEL_RECONNECT_GRACE` | 5s | WaitForTunnel 宽限 |
| `TUNNEL_STREAM_TIMEOUT` | 300s | 单 stream 超时 |
| `TUNNEL_STREAM_WINDOW` | 1MiB | credit 初始窗口 |
| `ALLOWED_ORIGINS` | 派生自 PRIMARY_DOMAIN | Origin 白名单，逗号分隔 |

Backend：`ConnectTunnelCommand` 下发开关、preview token TTL（默认 30min）。

---

## 9. 测试策略

| 层 | 测试 | 关注点 |
|----|------|--------|
| `protocol/tunnelframe` | 编解码单测 | 帧边界、stream_id、大 payload 分块 |
| `tunnel/stream` | credit 流控单测 | 窗口耗尽阻塞、补窗恢复、内存上界 |
| `tunnel/registry` | 单测 | 重连接管、WaitForTunnel 宽限、waiter 唤醒 |
| `proxy/http` | httptest 双向 | Range/206、SSE flush、header 过滤、大文件不驻留 |
| `proxy/websocket` | 集成 | 升级、双向帧、关闭传播 |
| `handler_preview` | 集成 | 401/403/404/502/429 各错误码、cookie 交换 |
| runner `local_http` | 单测 | loopback 校验、target_forbidden、流式转发 |
| backend `preview` | 单测 | ResolvePreviewRoute 各分支、下发失败不发 token |
| 端到端 | fake runner + 内存 local svc | 首屏、runner 重连恢复、视频 Range |
| 回归 | 现有 relay 全套 | 终端/ACP 行为零变更 |

---

## 10. 分阶段落地

1. **Phase 1 — 加固现状**：Origin 白名单（A1）、token_type claim（A2）、preview 下发失败不发 token（A3）。不引入隧道。可独立发布。
2. **Phase 2 — 隧道骨架**：`tunnelframe` + `tunnel/` + runner `tunnel client` + `connect_tunnel` proto/下发。仅打通 HELLO/PING、空转，无 preview 对外路由。
3. **Phase 3 — HTTP preview**：`handler_preview` + `proxy/http` + backend preview API + pod 元数据迁移。支持 HTML/JS/CSS/图片。
4. **Phase 4 — 媒体与 WS**：Range/视频、SSE、`proxy/websocket`、前端 `PreviewPanel` + session cookie。
5. **Phase 5 — 运维化**：隧道/代理 OTel 指标、心跳上报隧道统计、管理端 targets 视图、多 Gateway 实例。

每个 Phase 产出可测试、可回归的软件；Phase 1 与现网完全兼容，Phase 2 起对 runner 增量下发。

---

## 11. 未决问题

- preview 是否需要独立子域（`{podKey}.preview.domain`）以获得更强 cookie/CSP 隔离？当前设计走 path 前缀，子域为后续增强。
- 多 Gateway 实例时，同一 runner 的隧道落在哪个实例：需 backend 在 `connect_tunnel` 指定 Gateway URL，并保证 preview token 的 gateway 亲和（Phase 5 决策）。
- 是否需要对 preview 响应做内容级审计/DLP（当前只记录元数据）。
