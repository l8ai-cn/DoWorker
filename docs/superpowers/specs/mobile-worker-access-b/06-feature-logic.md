# 功能清单与逐项逻辑

## 1. 桌面生成手机入口

**Given** 用户可读一个 Worker

**When** 点击标题栏手机图标

**Then** Backend 返回 canonical URL 和能力，前端生成无 Token QR。

失败路径：

- 无权限：隐藏入口或显示 403
- Worker 不存在：关闭弹窗并刷新列表
- 没有公网地址：显示配置错误，不生成 localhost QR

## 2. 手机扫码登录

1. 扫码打开 canonical URL。
2. 未登录跳转登录页并保存 redirect。
3. 登录完成恢复 Worker 深链。
4. Org layout 切换组织。
5. Descriptor API 再次做 Pod RBAC。

二维码本身不能授予权限。

## 3. 手机 Worker 列表

列表从 Rust Core Pod state 派生，不维护 mobile 副本。

排序：

1. controlling
2. running
3. initializing/queued
4. paused
5. terminal states

点击 Worker 进入单 Worker 工作区，不自动申请控制租约。

## 4. 连接 PTY

1. 用户选择 Cloud 或 Edge。
2. 前端请求 `GetPodConnection(relay_id)`。
3. Backend 确保 Runner publisher ready。
4. Backend 返回短期 Browser Token。
5. Rust Relay driver 建立 subscriber。
6. 收到 Snapshot 后标记 connected。
7. 用户点击“取得控制”后申请租约。
8. 获租约后才能发送 Input/Resize。

## 5. 连接 ACP

前五步与 PTY 相同。收到 ACP Snapshot 后：

- 渲染历史、计划、模式和 pending permission
- 取得控制租约后允许 prompt、permission、cancel、interrupt
- observer 只能查看

## 6. 前后台恢复

**When** 浏览器从后台回到前台：

1. 检查 WebSocket 状态。
2. 断线则由 Rust driver 重连。
3. 重连后发送 Resync 和最后尺寸。
4. 等待 Snapshot。
5. 租约失效则回到 observer。

不得在页面层新建备用 socket。

## 7. 网络切换

Wi-Fi 切到蜂窝时：

- Cloud 模式：同 Relay 重连。
- Edge 模式：明确显示 Edge 不可达。
- 不自动切换 Cloud。
- 用户可主动返回连接方式选择 Cloud。

## 8. 打开 Preview

1. Descriptor 声明 Preview capability。
2. 用户点击 Preview。
3. Backend 确认 tunnel ready。
4. 返回一次性 `session_url`。
5. 页面用 `window.location.replace` 跳转。
6. Relay 换取 HttpOnly Cookie。
7. HTTP/WS 请求通过 tunnel 到 Pod loopback 服务。

## 9. Preview 配置

Worker 创建/设置页面提供：

- 开关
- port 数字输入
- path 输入
- 配置校验

保存后创建新的 config revision。运行中 Worker 是否支持热更新应由现有
Pod config revision 状态机决定；不允许只改 DB 而 Runner 不知情。

## 10. Tunnel 断线

Runner 检测 read/write/heartbeat 失败后进入 Backoff 并自动重连。

Relay 对现有 Preview stream：

- 关闭 WebSocket
- 未完成 HTTP 返回 `502 tunnel_disconnected`
- 清理 stream credit 和 pending request

重连完成后新请求可用，旧请求不重放。

## 11. 多端观察

桌面和手机可同时收到：

- PTY Output/Snapshot
- ACP Event/Snapshot
- Runner disconnect/reconnect
- 控制租约状态

Relay 不为每个客户端复制 Runner publisher。

## 12. 多端接管

默认已有控制者时，新客户端：

- 显示控制者设备和租约剩余时间
- 可请求接管
- 原控制者确认，或管理员强制接管
- 租约到期自动释放

Phase 1 可先实现“先到先得 + 显式释放 + 超时”，确认接管作为后续增量。

## 13. Runner 离线

打开已有 Worker：

- Console 显示 `runner_offline`
- 不签发无法工作的 Browser Token
- 提供刷新 Runner 状态

创建新任务可复用 RFC-006 的持久队列，但它是“任务下发”，不是实时控制
连接的 fallback。

## 14. Relay 不可用

Backend 返回：

```json
{
  "code": "relay_unavailable",
  "relay_id": "relay-cloud-1",
  "retryable": true
}
```

前端显示所选模式失败。不能自动选择另一个 Relay。

## 15. 权限变化

用户权限被撤销后：

- 新 Token 签发失败
- 已有短期 Token 到期后不能续签
- 高风险场景可通过 Relay revoke 通道提前断开
- 控制租约立即释放

## 16. Worker 终止

Runner 上报 terminated 后：

- Relay 广播终态
- Browser 停止重连
- 控制租约释放
- Preview 新请求返回 410
- 手机页面保留日志/状态，不显示输入控件

## 17. 诊断信息

移动连接诊断展示：

- Backend reachable
- Authenticated
- Worker active
- Runner connected
- Relay selected
- Runner publisher ready
- Browser WebSocket connected
- Snapshot received

诊断信息不得展示 Token、内部 IP 或敏感配置。
