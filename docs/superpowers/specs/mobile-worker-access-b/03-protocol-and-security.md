# 协议与安全模型

## 1. 协议复用原则

移动端不得新增专属 WebSocket 或 JSON-RPC 协议。

| 链路 | 协议 | 移动端动作 |
|---|---|---|
| Phone -> Backend | Connect/REST over HTTPS | 复用鉴权和 Org scope |
| Backend -> Runner | gRPC bidi + mTLS | 复用控制命令 |
| Phone -> Relay | Binary WebSocket | 复用 Browser endpoint |
| Runner -> Relay | Binary WebSocket | 复用 publisher endpoint |
| Relay -> Runner Preview | Multiplexed tunnel | 复用 tunnel frame |

## 2. PTY 与 ACP 帧

现有一字节消息类型继续作为 SSOT：

| Type | Hex | 方向 |
|---|---:|---|
| Snapshot | `0x01` | Runner -> Browser |
| Output | `0x02` | Runner -> Browser |
| Input | `0x03` | Browser -> Runner |
| Resize | `0x04` | Browser -> Runner |
| Ping/Pong | `0x05/0x06` | 双向 |
| Control | `0x07` | 双向 |
| Runner disconnected/reconnected | `0x08/0x09` | Relay -> Browser |
| Resync | `0x0A` | Browser -> Runner |
| ACP Event | `0x0B` | Runner -> Browser |
| ACP Command | `0x0C` | Browser -> Runner |
| ACP Snapshot | `0x0D` | Runner -> Browser |

手机前后台切换后必须走共享 Relay driver：

1. WebSocket 恢复。
2. 请求 `Resync`。
3. 等待 Snapshot 或 ACP Snapshot。
4. 重新发送最后有效尺寸。
5. 收到数据前保持 `reconnecting`，不能假装 connected。

## 3. Preview Tunnel 帧

继续使用：

- `HELLO`
- `PING/PONG`
- `REQ_START/BODY/END`
- `RESP_START/BODY/END/ERROR`
- `WS_DATA/CLOSE`
- `STREAM_CANCEL`
- `CREDIT`

保留现有 256 KiB chunk 和 credit-based backpressure。禁止引入
lulu 的整包 JSON + Base64 代理。

## 4. Mobile Access API

建议增加 Connect API：

```protobuf
rpc GetMobileAccessDescriptor(GetMobileAccessDescriptorRequest)
    returns (MobileAccessDescriptor);

message GetMobileAccessDescriptorRequest {
  string org_slug = 1;
  string pod_key = 2;
}

message MobileAccessDescriptor {
  Pod pod = 1;
  repeated MobileCapability capabilities = 2;
  repeated MobileConnectionCandidate candidates = 3;
  string canonical_url = 4;
}
```

`canonical_url` 是无 Token HTTPS 深链。Descriptor 不返回 Relay JWT
或 Preview JWT。

```protobuf
message MobileConnectionCandidate {
  string relay_id = 1;
  string kind = 2;          // cloud | edge
  string display_name = 3;
  string status = 4;        // ready | unavailable
  optional string reason = 5;
}
```

真正连接时仍调用 `GetPodConnection`，增加显式 `relay_id`：

```protobuf
message GetPodConnectionRequest {
  string org_slug = 1;
  string pod_key = 2;
  optional string relay_id = 3;
}
```

Backend 必须验证：

- Relay 属于允许的服务注册表
- Relay healthy
- Runner 能连接该 Relay
- 用户有 Pod read 权限
- Pod active

## 5. Token 类型

| Token | 接收方 | 建议 TTL | 绑定 |
|---|---|---:|---|
| Browser Relay | Relay | 15 分钟 | user/org/pod/runner/relay |
| Runner Relay | Relay | 1 小时 | runner/pod/relay |
| Tunnel | Relay | 1 小时 | runner/org/relay |
| Preview bootstrap | Relay | 5 分钟 | user/org/pod/runner/target/path |
| Control lease | Relay | 最多 5 分钟 | user/pod/client instance |

Token 必须包含明确 `token_type`，禁止通过缺失字段推断旧类型。

## 6. 二维码与登录

二维码只承载 canonical URL。流程：

1. 手机打开深链。
2. 未登录则跳转登录。
3. 登录成功恢复完整深链。
4. Org layout 验证组织成员关系。
5. Descriptor API 验证 Pod read 权限。
6. 用户主动连接后才签发短期 Relay Token。

二维码泄露最多暴露 `org_slug` 和 `pod_key`，不能绕过登录和 RBAC。

## 7. Preview 安全

Preview target 必须：

- 由 Pod 配置生成
- 固定 loopback host
- port 在允许范围
- path 规范化，禁止 `..`
- 写入 Token claim
- Relay 校验 claim 与 URL pod key 一致

Preview API 只返回：

```json
{
  "preview_base_url": "https://edge/preview/pod-1/",
  "session_url": "https://edge/preview/pod-1/__session?token=...",
  "expires_at": "..."
}
```

删除响应中的裸 `token` 字段。

## 8. Origin、Cookie 与日志

- 生产必须 HTTPS/WSS。
- Relay Browser/Preview origin 使用显式 allowlist。
- Preview Cookie：`HttpOnly; Secure; SameSite=Lax`。
- Cookie Path：`/preview/{podKey}`。
- URL query 中的 bootstrap Token 必须从访问日志中脱敏。
- 302 后不得在 Referer 中继续传播 Token。

## 9. 控制租约

Relay channel 保存：

```go
type ControlLease struct {
    ClientID  string
    UserID    int64
    ExpiresAt time.Time
}
```

Input、Resize、ACP Command 处理前检查租约。无租约返回明确 Control
事件，不静默丢弃。输出、Snapshot 和状态事件仍广播给所有观察者。

## 10. 审计

Backend 记录：

- mobile_access_opened
- mobile_connection_requested
- mobile_connection_failed
- preview_session_created
- control_lease_acquired/released/revoked

禁止记录 JWT、Cookie、PTY 内容和 ACP prompt 正文。
