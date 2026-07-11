# Do Worker 整改计划

基于 10 批次 bug-hunter 结论，分阶段修复 P0/P1 缺陷并对齐前端架构演进方向。

## 架构定位对齐

| 组件 | 目标角色 | 本轮关系 |
|------|----------|----------|
| **web** | 主控制台（逐步吸收 web-admin） | P1 PersistedSession 语义对齐 |
| **web-user** | 终端用户产品面 | 与 Rust SSOT / light-session 契约一致 |
| **web-admin** | 内部管理（过渡） | 记录并入 web 的 TODO，本轮不迁移 |
| **mobile** | 移动端（未来） | Phase 4 文档化 |
| **relay** | 独立 Gateway 隧道（演进） | 终端数据面已分离，控制面继续走 gRPC |
| **backend** | API + 编排 SSOT | P0 启动顺序、P1 pending drainer |
| **runner** | 自托管执行面 | P0 ACP 子进程泄漏 |
| **clients/core (Rust)** | 前端业务 SSOT | P1 protojson int64 解析 |

---

## Phase 1 — 稳定性（P0，本轮）

**目标**：消除资源泄漏与启动竞态，保证 Runner 重连后 ACP/Usage 事件不丢。

| # | 项 | Owner | 修复方向 | 验收标准 | 验证命令 |
|---|-----|-------|----------|----------|----------|
| 1 | ACP 子进程泄漏 | `runner/` | `Start()`/`NewSession()` 失败时调 `acpClient.Stop()`；sandbox 用 `removePodSandbox` | 失败路径无 orphan 子进程；worktree 正确移除 | `bazel test //runner/internal/runner:message_handler_acp_test //runner/internal/runner:pod_io_acp_test` |
| 2 | Backend 启动竞态 | `backend/cmd/server`, `backend/internal/api/rest` | gRPC `Start()` 延后至 `SetPodEventSink` 之后 | Runner 早连时 ACP/Usage/ExternalSession 事件进入 session stream | `bazel test //backend/internal/api/grpc:...` + 手动：backend 重启后立即连 runner，观察 session 事件 |

---

## Phase 2 — 数据正确性（P1，本轮部分）

**目标**：消除 nil panic 与 protojson 解析遗漏。

| # | 项 | Owner | 修复方向 | 验收标准 | 验证命令 |
|---|-----|-------|----------|----------|----------|
| 3 | Pending drainer nil sender | `backend/internal/service/runner` | `msgSender == nil` 时 skip dispatch | 启动窗口内 drain 不 panic | `bazel test //backend/internal/service/runner:pending_drain_test` |
| 4 | Rust `ji64` for `pod_id` | `clients/core/crates/state` | MR/Pipeline 事件用 `ji64()` 替代 `as_i64()` | 字符串编码 `pod_id` 触发 refetch | `bazel test //clients/core/crates/state:state_test` |
| 5 | PersistedSession 语义 | `clients/web-user`, `clients/web` | auth-session 与 light-session/Rust 对齐 | 刷新后 org/user 一致 | 手动 E2E |
| 6 | Expert Zustand | `clients/web-user` | **TODO**：范围过大，暂不迁 Rust | 文档记录 | — |

---

## Phase 3 — 前端整合（架构，下轮）

**目标**：减少重复前端，统一 wasm 边界。

- web-admin 路由/Connect handler 逐步迁入 `clients/web`（`/admin` 已存在 basePath）
- web-user 成为默认用户入口；web 保留 power-user / org 管理
- Expert 状态最终迁入 Rust（依赖 Phase 2 #5 会话契约）

**验收**：单一 `pnpm install` 树；营销页仍 0 wasm（`check-no-wasm-in-marketing.sh` 通过）。

---

## Phase 4 — 基础设施演进（架构，长期）

**目标**：relay 独立为 Gateway 隧道服务；mobile 客户端对接同一 Rust core。

- relay：控制面注册保留 backend，数据面可独立部署与扩缩
- mobile：复用 `agentsmesh-wasm` 或 native FFI over Rust core
- 每步保持 backend ↔ runner gRPC 契约不变
- [x] **隧道 Gateway 已落地**（2026-07-08，`feat/gateway-module`）：relay 新增 HTTP 数据面
  `/runner/tunnel`（Runner 出站注册，Runner 粒度隧道，`tunnel.Registry` 负责重连/迁移/离线清理）
  + `/preview/{podKey}/*`（HTTP/WS/SSE/Range 代理到 Runner 本地回环服务，JWT claim 路由，
  逐 stream credit 流控与错误隔离）。控制面 `connect_tunnel`（`runner.proto` command=14）由
  backend 在 Runner `initialized` 回调下发；心跳上报隧道统计（`active_tunnels`/`active_streams`）。
  详见 `docs/superpowers/specs/2026-07-08-gateway-module-design.md` 与
  `docs/superpowers/plans/2026-07-08-gateway-module.md`。

**验收**：relay 可独立 `bazel build //relay/...` 部署；mobile 原型可读 org pod 列表。

---

## 本轮执行范围

- [x] Phase 0：本文档
- [x] Phase 1：P0 #1 #2
- [x] Phase 2：P1 #3 #4
- [ ] Phase 2 剩余：#5 #6
- [ ] Phase 3–4：下轮规划
