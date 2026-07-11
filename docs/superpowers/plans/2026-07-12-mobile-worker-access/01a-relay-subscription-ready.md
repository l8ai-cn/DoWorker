# Relay And Tunnel Ready Acknowledgement

## Contract

`SendSubscribePod`/`SendConnectTunnel` 写入 gRPC stream 不等于 Runner 已
连接 Relay。Browser/Preview Token 只能在 Runner 明确回报 publisher/tunnel
ready 后签发。失败、超时、Runner 断线都返回 unavailable，不使用 heartbeat
轮询或固定 sleep。

## Task 1b: Command-Correlated Result

**Files**

- Modify: `proto/runner/v1/runner.proto`
- Regenerate Runner Go bindings
- Modify: `runner/internal/client/protocol.go`
- Modify: `runner/internal/client/grpc_handler_dispatch.go`
- Modify: `runner/internal/client/grpc_sender_control.go`
- Modify: `runner/internal/tunnelframe/*`
- Modify: `runner/internal/tunnel/client.go`
- Modify: focused Runner dispatch/sender tests
- Modify: `relay/internal/protocol/tunnelframe/*`
- Modify: `relay/internal/server/handler_tunnel.go`
- Modify: focused Relay tunnel tests
- Create: `backend/internal/service/runner/relay_subscription_tracker.go`
- Create: `backend/internal/service/runner/relay_subscription_tracker_test.go`
- Modify: `backend/internal/service/runner/connection_manager.go`
- Modify: connection manager callback/handler files and tests
- Modify: `backend/internal/api/grpc/runner_adapter_types.go`
- Modify: `backend/internal/api/grpc/runner_adapter_send.go`
- Modify: `backend/internal/api/grpc/runner_adapter_message.go`
- Modify: focused adapter tests
- Modify: `backend/internal/api/connect/pod/connection.go`
- Modify: `backend/internal/api/connect/pod/connection_test.go`

### Protocol

```protobuf
message SubscribePodCommand {
  // existing fields
  string command_id = 8;
}

message RelaySubscriptionResultEvent {
  string command_id = 1;
  string pod_key = 2;
  bool ready = 3;
  string error_code = 4;
}

message ConnectTunnelCommand {
  // existing fields
  string command_id = 3;
}

message TunnelConnectionResultEvent {
  string command_id = 1;
  int64 runner_id = 2;
  bool ready = 3;
  string error_code = 4;
}
```

`RunnerMessage` adds both result events on unused field numbers.

### RED

Tests cover:

- Runner success sends `ready=true` only after `OnSubscribePod` returns nil.
- Runner failure sends `ready=false` with stable error code.
- Relay 在注册 tunnel 后发送 `HELLO_ACK`。
- Tunnel success is emitted only after Runner receives `HELLO_ACK`.
- Tunnel dial/handshake failure emits `ready=false`.
- Backend correlates concurrent commands by `command_id`.
- timeout and Runner disconnect remove pending waiters.
- GetPodConnection returns unavailable on negative result or timeout.
- GetPodPreview returns unavailable on negative tunnel result or timeout.
- Browser token is generated only after positive result.
- Preview token is generated only after positive tunnel result.

### GREEN

1. Backend registers a waiter before sending the command.
2. Command carries a UUID.
3. Runner executes the existing synchronous `OnSubscribePod`.
4. Runner executes `OnConnectTunnel` and waits for its initial connection result.
5. Relay validates HELLO, registers tunnel, then returns `HELLO_ACK`.
6. Runner sends exactly one correlated result event per command.
7. Backend resolves the waiter and only then returns connection/session data.
8. Context cancellation and timeout remove the waiter.

Backend waits at most 25 seconds. Runner sends readiness events through a
dedicated bounded queue that the gRPC writer drains before ordinary control
messages. A saturated control queue therefore cannot silently lose the result
that unblocks a browser or Preview request.

The readiness contract requires Runner protocol version 3. No heartbeat
polling, fixed sleep, fire-and-forget result, or compatibility path without
`command_id` is allowed.

Run:

```bash
pnpm proto:gen-go-all
cd runner
go test ./internal/client ./internal/runner -run 'RelaySubscription|SubscribePod' -count=1
cd ../relay
go test ./internal/protocol/tunnelframe ./internal/server -run 'HelloAck|Tunnel' -count=1
cd ../backend
go test ./internal/service/runner ./internal/api/grpc \
  ./internal/api/connect/pod ./internal/api/rest/v1 \
  -run 'RelaySubscription|TunnelConnection|GetPodConnection|GetPodPreview' -count=1
```

## Acceptance

- Dispatch success alone cannot produce a Browser Token.
- Tunnel command enqueue success alone cannot produce a Preview Token.
- Result belongs to the exact command, Pod and Runner.
- Late result after timeout is ignored and cleaned up.
- A full control queue cannot drop a readiness result.
- A Runner below protocol version 3 is rejected before a connection URL or
  session URL can be returned.
- No new persistent table is introduced.
