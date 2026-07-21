# 现状逆向与差距

## 1. lulu 的真实实现

`/Users/wwyz/Desktop/lulu-codex-web-mac` 存在三条路径：

| 路径 | 数据流 | 结论 |
|---|---|---|
| LAN 主路径 | Phone -> codexapp HTTP/WS -> Codex app-server stdio | 手机是同一 Web 服务的新客户端 |
| Mobile Gateway | Phone -> `17631` 自定义 WS API | 遗留旁路，主二维码未使用 |
| Cloud Relay | Phone -> Relay -> PC 出站 WS -> codexapp | 整站 HTTP/WS 反向代理 |

主二维码由 `QrPortalServer` 生成，只包含 `http://LAN-IP:5900/`。
它不创建独立远程会话，也不执行设备配对。

Cloud Relay 使用 JSON + Base64 转发完整 HTTP/WS 消息，存在：

- 无流窗口和背压
- HTTP 整包缓存
- 静态团队 Token
- 控制端无认证
- 设备掉线后 pending HTTP 只能等超时
- 默认 WebUI 无密码

lulu 的 “sandbox” 主要是 Codex 权限参数、Git worktree 和主机 PTY，
不是 Agent Cloud 的持久 Pod 和隔离生命周期。

## 2. lulu 前端功能

`codexapp@0.1.90` 是 Vue 3/Vite SPA，主要页面：

- 线程列表、新建线程和线程详情
- Skills
- Automations
- 账号、模型和配额

已确认的工作区功能：

- 线程搜索、归档、分叉、压缩和回滚
- 模型与推理强度
- 审批和 Agent 运行状态
- Git 状态、Review 和 worktree
- Skills、Plugins 和 MCP
- 文件上传、本机文件浏览和编辑
- 语音输入
- xterm 主机终端
- PWA、移动抽屉和通知

Gateway 另提供二维码页、Playwright 截图式 Browser Preview、远程设备页，
但这些不是 `codexapp` 的统一业务状态。

## 3. 实现方式对比

| 方案 | UI | 控制面 | 数据面 | 重复风险 |
|---|---|---|---|---|
| lulu LAN | codexapp | codexapp HTTP RPC | codexapp WS/SSE | 绕开 Agent Cloud |
| lulu Cloud | codexapp | Cloud Relay JSON | 整站代理 | 双 Relay、双鉴权 |
| 现有原型 | Web 单 Pod 页 | Backend | Cloud Relay | UI 不完整 |
| 推荐 Phase 1 | Mobile Worker UI | Backend | 现有 Cloud Relay | 最少 |
| 推荐 Phase 2 | 同一 Mobile UI | Backend | 标准 Edge Relay | 只扩展部署拓扑 |

如果复制 lulu Gateway，会重复：

- Pod/线程查询
- WebSocket 生命周期
- Token 和设备认证
- PTY/ACP 消息适配
- Preview HTTP/WS 代理
- Runner 在线状态
- 前端业务状态

## 4. 可借鉴内容

- 扫码直接进入目标会话
- 手机页面优先展示当前工作内容
- 明确区分控制台和 Web Preview
- 断线、加载、失败状态可见
- 终端快照恢复
- 移动抽屉、权限请求和任务进度适配

## 5. 禁止复制内容

- 默认无密码
- 静态、不可撤销的设备 Token
- JSON + Base64 整站代理
- 独立 Mobile Gateway 协议
- 未认证的设备列表和控制路由
- 任意主机文件读写和主机 Shell
- UDP/LAN IP 作为安全边界
- 修改 vendored 构建产物注入 UI

## 6. Agent Cloud 已有能力

| 能力 | 当前实现 |
|---|---|
| 控制面 | Backend <-> Runner gRPC bidi + mTLS |
| PTY 数据面 | Browser <-> Relay <-> Runner 二进制 WS |
| ACP 数据面 | 同一 Relay 的 ACP Snapshot/Event/Command |
| Browser 恢复 | Rust RelayConnectionPool 重连、Resync、Resize |
| Runner 恢复 | Relay publisher 重连和 Token 刷新 |
| Preview | Relay HTTP/WS proxy + Runner outbound tunnel |
| Auth | Org scope、Pod RBAC、typed JWT |
| Web 状态 | Rust Core/WASM 是业务状态 SSOT |

现有移动提交增加了：

- `/{org}/mobile/pods/{podKey}`
- `/{org}/mobile/pods/{podKey}/preview`
- `MobilePodWorkspace`
- `PodMobileAccessDialog`
- `preview_port/preview_path` 投影

## 7. 已确认缺口

### P0：二维码在本地不可达

`buildPodMobileConsoleUrl` 使用 `window.location.origin`。桌面打开
`localhost` 时，二维码中的 `localhost` 指向手机自身。

### P0：订阅命令失败仍签发 Token

`GetPodConnection` 对 `SendSubscribePod` 错误只记 warning，仍返回
Browser Token。浏览器会连接到没有 Runner publisher 的 Relay channel。

### P0：Preview dispatch 可被跳过

`GetPodPreview` 只有 `commandSender != nil` 时才发送 tunnel 命令。
依赖缺失时仍可能签发 Preview Token。

### P1：Preview 没有配置闭环

`preview_port` 默认是 0，正常 Worker 创建和设置流程没有写入入口。

### P1：`preview_path` 未生效

路由解析得到 `Path`，但 Preview URL 和 Token 只绑定 target，
请求不会自动加上配置路径。

### P1：Tunnel 不自动重连

Runner Tunnel Client 的 read loop 退出后只标记 disconnected，
没有持久重连循环。

### P1：入口不可发现

桌面唯一入口位于 Worker 右键菜单；触屏用户没有等价操作。

### P1：缺少移动信息架构

当前页面只复用 `TerminalPane` 或 `AgentPanel`，没有：

- Worker 选择
- Runner/Worker 在线状态
- 接入方式
- 连接诊断
- Preview 配置
- 多端控制状态
- 离线任务入口

### P1：多端输入无仲裁

多个 Browser subscriber 都能发送 PTY Input、Resize 和 ACP Command。
桌面与手机可同时写入同一 Worker。

### P2：Preview API 重复

`clients/web` 与 `clients/web-user` 各自实现 Preview fetch 和刷新，
存在响应映射和生命周期漂移风险。

## 8. 当前测试证据

使用 `clients/web/vitest.config.ts` 定向执行：

- 4 个移动测试文件通过
- 8 个测试通过

这只证明组件级行为，不证明以下场景：

- 真机扫码
- 跨设备登录回跳
- PTY/ACP 实际输入
- Wi-Fi/蜂窝切换
- Tunnel 断线恢复
- 桌面和手机并发控制
- 生产域名、TLS、Cookie 和反向代理
