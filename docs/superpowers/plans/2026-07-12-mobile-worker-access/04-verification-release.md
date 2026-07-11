# 验证、发布和回滚

## Task 9: Automated Verification

按顺序运行：

```bash
cd backend
go test ./internal/api/connect/pod ./internal/api/rest/v1 \
  ./internal/service/relay -count=1

cd ../relay
go test ./internal/channel ./internal/server ./internal/protocol \
  ./internal/tunnel -count=1

cd ../runner
go test ./internal/tunnel ./internal/relay ./internal/runner -count=1

cd ../clients/core
cargo test -p agentsmesh-protocol -p agentsmesh-relay

cd ../../
pnpm run build:wasm
pnpm run web:test
pnpm run web:typecheck
pnpm run web:lint
```

任何失败必须修根因，不增加兼容 fallback。

## Browser E2E

启动：

```bash
./deploy/dev/dev.sh
```

用浏览器执行：

1. 登录 `dev@agentsmesh.local`。
2. 创建 PTY Worker。
3. 从可见手机入口打开二维码。
4. 用移动 viewport 打开 canonical URL。
5. 验证 Worker 列表、连接状态和 Snapshot。
6. 申请控制，发送输入。
7. 同时打开桌面，验证 observer/controller。
8. 创建 ACP Worker，验证 prompt 和 permission。
9. 配置 Preview，验证 HTTP/WS。
10. 重启 Relay、Runner，验证状态与恢复。

必须检查 console、network 和截图。

## 2026-07-12 本地真实验证记录

环境：

- Web：`http://192.168.1.169:10007`
- API/Relay：`http://192.168.1.169:10000`
- Worker：`1-standalone-c0c45043`
- Runner：Docker `e2e-echo`

已验证：

1. 手机 viewport `390x844` 通过 LAN 地址登录并打开 canonical Worker route。
2. 页面显示 Web 和 Relay 均已连接，初始状态为 observer。
3. 点击“接管控制”后终端输入启用。
4. 输入 `lan-mobile-e2e` 并回车后，Runner PTY `total_reads` 从 4 增至 7。
5. Relay 在 `2026-07-12 21:45:38 +08:00` 向该手机订阅者广播 38 字节响应。
6. 第二个桌面连接在手机持有租约时显示“另一台设备正在控制”；释放后可接管。
7. Mobile Access Descriptor 返回
   `http://192.168.1.169:10007/dev-org/mobile/workers/1-standalone-c0c45043`，
   URL 不含 token。
8. 二维码、旧 `/mobile/pods/*` 重定向、Preview `replace` 跳转已做浏览器和
   Vitest 验证。

本记录不能替代 iPhone Safari、Android Chrome 和生产环境 smoke。正式发布
仍按下方发布门禁执行。

## 真机

至少：

- iPhone Safari
- Android Chrome

验证扫码、登录回跳、中文 IME、横竖屏、后台 30 秒、Wi-Fi/蜂窝切换。

## Review

每个任务执行：

1. Spec reviewer 对照设计和任务。
2. Quality reviewer 检查并发、安全、文件行数和测试质量。
3. Reviewer 问题修复后重新审查。

最终再做跨模块 review。

## Git

只暂存本任务文件，保留 `clients/web/next.config.ts` 他人改动。

最终：

```bash
git status --short
git log --oneline origin/main..HEAD
git push origin main
git fetch origin main
git branch -r --contains HEAD
```

发布使用仓库 GitOps/CI。生产 smoke 未通过时回滚版本，不手工热修。

本地 `dev.sh --backend-only` 曾因 Docker Hub IPv6 鉴权超时无法拉取无关的
Runner base image；Backend、Relay 和现有 Runner 随后使用仓库开发脚本的
host service 路径重启并完成上述验证。该基础设施事件不能用来跳过生产镜像和
CI 验证。
