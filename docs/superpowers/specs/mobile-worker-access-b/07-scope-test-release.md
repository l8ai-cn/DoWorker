# 改动范围与测试

## 1. Phase 0 阻断修复

必须先完成：

- `GetPodConnection` 的 subscribe fail-closed
- Preview command sender 缺失 fail-closed
- Preview API 删除裸 `token`
- PreviewConfig 生产写入路径
- `preview_path` 真正参与代理
- Runner tunnel 自动重连
- canonical URL 不使用浏览器 localhost
- 修复/明确 Vitest 从仓库根执行方式

## 2. Phase 1 前端范围

预计新增：

```text
clients/web/src/app/(dashboard)/[org]/mobile/workers/page.tsx
clients/web/src/app/(dashboard)/[org]/mobile/workers/[podKey]/page.tsx
clients/web/src/app/(dashboard)/[org]/mobile/workers/[podKey]/preview/page.tsx
clients/web/src/components/mobile-worker/*
clients/web/src/hooks/useMobileAccessDescriptor.ts
clients/web/src/hooks/useWorkerControlLease.ts
```

预计修改：

```text
clients/web/src/components/workspace/*
clients/web/src/components/ide/sidebar/*
clients/web/src/stores/relayConnection.ts
clients/web/src/messages/{en,zh}/*
clients/web/src/lib/ide-chrome.ts
clients/web/src/lib/api/connect/podConnect.ts
```

Rust Core 继续持有业务状态；只有控制租约投影需要扩展 Relay crate。

## 3. Phase 1 Backend/Relay/Runner 范围

Backend：

```text
backend/internal/api/connect/pod/*
backend/internal/api/rest/v1/pod_preview.go
backend/internal/service/mobileaccess/*
backend/internal/service/relay/*
backend/internal/service/agentpod/*
```

Relay：

```text
relay/internal/server/handler.go
relay/internal/server/handler_preview*.go
relay/internal/channel/*
relay/internal/auth/*
relay/internal/tunnel/*
```

Runner：

```text
runner/internal/tunnel/*
runner/internal/runner/message_handler_tunnel.go
runner/internal/relay/*
```

协议：

```text
proto/pod/v1/pod.proto
proto/runner/v1/runner.proto
proto/gen/*
clients/core/crates/relay/*
```

实际实施时按单一职责拆分，每个提交只覆盖一个可验证契约。

## 4. 数据迁移

Phase 1 优先复用 Pod config revision，不新增 Mobile 专属表。

若 PreviewConfig 现有列不足：

- 新建当前 migration 序号
- up/down 完整
- port/path check constraint
- migration PostgreSQL 真实 up/down 测试

Phase 2 Edge Relay 元数据单独迁移，不能提前塞入 Phase 1。

## 5. 单元与协议测试

Backend：

- RBAC 和 Org scope
- canonical URL
- relay selection
- subscribe/tunnel fail-closed
- PreviewConfig 校验
- Token claims/TTL/type

Relay：

- typed token 拒绝交叉使用
- origin allowlist
- Preview path join
- HTTP/WS/SSE/Range
- tunnel credit/backpressure
- control lease acquire/renew/release
- 非控制者 Input/ACP Command 拒绝

Runner：

- tunnel reconnect/backoff/jitter
- heartbeat timeout
- token refresh
- generation race
- loopback target 限制

Rust/Web：

- Snapshot 前不标 connected
- 前后台恢复
- observer/controller 状态
- QR 无 Token
- localhost 不生成不可达 URL
- Preview 使用 `replace`

## 6. 浏览器与真机 E2E

必须在真实服务上执行：

| 场景 | iPhone Safari | Android Chrome | Desktop |
|---|---:|---:|---:|
| 扫码、登录、深链恢复 | 必测 | 必测 | 辅助 |
| Worker 列表和选择 | 必测 | 必测 | 必测 |
| PTY 中文 IME/快捷键 | 必测 | 必测 | 对照 |
| ACP prompt/permission | 必测 | 必测 | 对照 |
| 横竖屏 Resize | 必测 | 必测 | 不适用 |
| 后台 30 秒恢复 | 必测 | 必测 | 对照 |
| Wi-Fi/蜂窝切换 | 必测 | 必测 | 不适用 |
| Preview HTTP/WS/SSE | 必测 | 必测 | 必测 |
| Runner/Relay/Tunnel 断线 | 必测 | 必测 | 必测 |
| 桌面手机并发控制 | 必测 | 必测 | 必测 |

每个 E2E 保存：

- 截图
- Browser console
- 网络错误
- Backend/Relay/Runner 关联日志
- 请求和连接时序

## 7. 性能与稳定性门禁

建议验收指标：

- 扫码到移动页面可操作：P95 < 5 秒
- Relay 建链到首个 Snapshot：P95 < 2 秒
- 前台恢复到 Snapshot：P95 < 3 秒
- Tunnel 重连：P95 < 10 秒
- 终端输入回显：同区域 P95 < 250 ms
- 1 小时会话无未解释断线

指标未达标时必须定位根因，不通过静默降级绕过。

## 8. 安全测试

- QR 泄露不能绕过登录
- 用户跨 Org 访问返回 403
- Browser Token 不能连接 Runner endpoint
- Preview Token 不能连接 PTY endpoint
- 过期和撤销 Token 被拒绝
- Host/Origin 注入被拒绝
- Preview path traversal 被拒绝
- 非 loopback target 被拒绝
- 日志不出现 JWT/Cookie
- 控制租约不能被其他用户伪造
