# PreviewConfig Revision Contract

## Scope

Preview port/path 是 Pod 配置的一部分，必须进入不可变 config revision，并与
Pod 当前投影原子同步。不得只更新 `pods`，也不得由前端拼接或缓存配置。

Preview 配置只改变 Relay 到 Runner 本地 HTTP 服务的路由，不改变 Worker
进程、PTY 或 ACP 运行参数。因此该版本不重启 Worker；revision 在同一事务中
激活，下一次 Preview 请求通过带目标端口和路径的 tunnel stream 把新配置显式
交给 Runner。后续若 PreviewConfig 增加进程环境或启动命令，再引入 Runner
reinitialize 命令，不能预先伪造无实际运行时效果的状态机。

## Task 4.1: Persist PreviewConfig In Revisions

**Files**

- Create: `backend/migrations/000204_add_preview_config_to_pod_revisions.up.sql`
- Create: `backend/migrations/000204_add_preview_config_to_pod_revisions.down.sql`
- Modify: `backend/migrations/migrations_test.go`
- Modify: `backend/internal/domain/agentpod/pod_config_revision.go`
- Modify: `backend/internal/domain/agentpod/pod.go`
- Modify: `backend/internal/infra/agentpod_repo.go`
- Modify: focused repository tests

### RED

Migration test verifies:

- `preview_port` accepts `0` or `1024..65535`.
- `preview_path` is non-empty, starts with `/`, and rejects traversal.
- down migration removes both columns.

Repository test creates revision 1 and asserts Pod plus revision return the same
normalized preview values.

### GREEN

- Add typed `PreviewPort` and `PreviewPath` columns to config revisions.
- Normalize path before persistence.
- Create/update Pod and revision in one transaction.
- Returned domain objects must match committed values.

Run:

```bash
cd backend
go test ./migrations ./internal/infra -run 'PreviewConfig|Migration000204' -count=1
```

## Task 4.2: Create And Update API

**Files**

- Modify: `proto/pod/v1/pod.proto`
- Regenerate Pod Go/TS/amesh bindings only
- Modify: `backend/internal/service/agentpod/pod_service.go`
- Create: `backend/internal/service/agentpod/pod_preview_config.go`
- Create: `backend/internal/service/agentpod/pod_preview_config_test.go`
- Modify: `backend/internal/api/connect/pod/server.go`
- Create: `backend/internal/api/connect/pod/preview_config.go`
- Create: `backend/internal/api/connect/pod/preview_config_test.go`
- Modify: Rust Core Pod service/projection files selected by generated API

### RED

Given a running Pod at generation 1, when an authorized member updates preview
port/path, then:

- revision 2 is created and active;
- revision 1 remains immutable;
- Pod generation becomes 2;
- `active_config_revision_id` points to revision 2;
- response and subsequent GetPod return normalized values;
- invalid port/path and unauthorized user fail without DB writes.

### GREEN

Add `UpdatePodPreviewConfig` as a typed Connect RPC. Service validates tenant,
normalizes input, locks the Pod row, calculates the next revision, creates the
revision and updates the Pod projection atomically.

不发送进程重启命令。Runner 在每个 Preview stream 上消费由受信 token claim
解析出的目标端口；配置变更由下一次真实代理请求生效。Readiness 以该请求
成功为准，不以 revision 激活或 tunnel command dispatch 为准。

Run:

```bash
pnpm proto:gen-go-all
cd backend
go test ./internal/service/agentpod ./internal/api/connect/pod -run PreviewConfig -count=1
cd ../clients/core
cargo test --workspace preview_config
```

## Acceptance

- Config revision is the audit source of truth.
- Pod projection is the current read model.
- No duplicate front-end persistence.
- No fake reinitializing state.
- Any future runtime-dependent preview setting must add and verify a real
  Runner acknowledgement before activation.
