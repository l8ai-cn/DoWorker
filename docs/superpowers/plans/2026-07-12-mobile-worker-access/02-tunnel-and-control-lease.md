# Tunnel 重连与控制租约

## Task 4: Runner Tunnel Reconnect

**Files**

- Modify: `runner/internal/tunnel/client.go`
- Create: `runner/internal/tunnel/client_reconnect.go`
- Modify: `runner/internal/tunnel/client_test.go`
- Create: `runner/internal/tunnel/client_reconnect_test.go`

### RED

测试服务器首次连接后主动关闭，第二次连接记录 HELLO：

```go
func TestClient_ReconnectsAfterReadFailure(t *testing.T) {
    server, connections := reconnectingTunnelServer(t)
    client := NewClient(context.Background(), server.URL, "tok", 7, 3, nil)
    require.NoError(t, client.Connect())
    client.Start()
    require.Eventually(t, func() bool {
        return connections.Load() >= 2 && client.IsConnected()
    }, 3*time.Second, 20*time.Millisecond)
}
```

Run:

```bash
cd runner
go test ./internal/tunnel -run 'TestClient_(Reconnects|Stop|Connect)' -count=1
```

Expected: reconnect 用例 FAIL。

### GREEN

- `Start` 启动单一 reconnect loop。
- read failure 清理当前 generation 后进入 backoff。
- 退避指数增长并带 jitter，测试注入短 backoff。
- `Stop` 必须中断 dial/backoff。
- 新 generation 不得被旧 read loop 标记 disconnected。

Run 同一命令，Expected: PASS。

## Task 5: Relay Single-Writer Lease

**Files**

- Create: `relay/internal/channel/control_lease.go`
- Create: `relay/internal/channel/control_lease_test.go`
- Modify: `relay/internal/channel/channel.go`
- Modify: `relay/internal/channel/channel_io.go`
- Modify: `relay/internal/protocol/message.go`
- Modify: `clients/core/crates/relay/src/dispatch.rs`
- Modify: `clients/core/crates/relay/src/types.rs`

### RED: Relay

```go
func TestChannel_OnlyControllerForwardsInput(t *testing.T) {
    ch, publisher, first, second := channelWithTwoSubscribers(t)
    acquireControl(t, first, "first")
    writeInput(t, second, "blocked")
    assertNoPublisherMessage(t, publisher)
    writeInput(t, first, "allowed")
    assertPublisherInput(t, publisher, "allowed")
}
```

再测：

- 首个 acquire 成功。
- 第二个 acquire 收到 `control_busy`。
- owner 可 renew，非 owner renew 被拒绝。
- lease TTL 到期后其他 subscriber 可 acquire。
- owner release 后立即广播 observer 状态。
- controller disconnect 自动释放。
- Output/Snapshot 仍广播给所有 subscriber。

Run:

```bash
cd relay
go test ./internal/channel ./internal/protocol -run 'Control|Controller' -count=1
```

Expected: FAIL，当前任意 subscriber 都能写。

### GREEN: Relay

Control payload：

```json
{"type":"control_lease","action":"acquire","client_label":"mobile"}
```

Reply：

```json
{"type":"control_lease","status":"granted","lease_id":"opaque","expires_at":1234567890}
```

规则：

- `Input/Resize/AcpCommand` 需要有效租约。
- `Ping/Pong/Resync/Control` 无需租约。
- 无控制者时不自动授予，客户端必须显式 acquire。
- 被拒绝输入返回 Control event，不静默丢弃。
- 所有权绑定 Relay 生成的 subscriber ID，不信任客户端标签。
- acquire 后返回 opaque lease ID；renew/release 必须同时匹配 subscriber 和
  lease ID。
- TTL 使用 Relay 单调时钟判定；客户端在到期前续租，断线不保留租约。
- granted/busy/released/expired 状态广播给所有观察者，但不暴露用户凭证。

### RED/GREEN: Rust

在 `dispatch_tests.rs` 增加：

```rust
assert_eq!(
    dispatch_message(MsgType::Control, lease_payload, &[]),
    DispatchAction::ControlLease(ControlLeaseStatus::Granted)
);
```

Run:

```bash
cd clients/core
cargo test -p agentsmesh-relay
```

Expected: RED 后增加 typed projection，GREEN 全通过。

Rust pool 增加 acquire/renew/release 命令和 lease 状态 mirror；WASM 只导出
该 typed API。Web hook 负责页面可见时续租，失焦或卸载时显式 release，
续租失败立即回到 observer。不得在 React 中自行推断 lease 所有权。
