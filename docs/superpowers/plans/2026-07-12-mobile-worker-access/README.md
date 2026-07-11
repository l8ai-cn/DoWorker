# Mobile Worker Access Phase 0/1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> `superpowers:subagent-driven-development`. Every production change follows
> RED -> GREEN -> targeted regression -> review.

**Goal:** 在不增加第二套移动协议的前提下，完成可靠的移动 Worker 云端接入。

**Architecture:** Backend 继续负责鉴权和连接编排；Browser 与 Runner 继续使用
现有 Relay 二进制协议；Preview 使用现有 multiplexed tunnel。Phase 1 不实现
Edge Relay，只保留协议扩展边界。

**Tech Stack:** Go、Connect RPC、Gin、WebSocket、Rust/WASM、Next.js、Vitest、
Playwright。

---

## 基线与所有权

- Branch: `main`
- Baseline: `2b6897565`
- 保留他人修改：`clients/web/next.config.ts`
- 禁止修改：AI Resource、Marketplace、Workflow 和部署外的无关文件
- 设计：`docs/superpowers/specs/mobile-worker-access-b/`

## 执行顺序

1. [Phase 0 Backend 与 Preview](01-phase0-backend-preview.md)
2. [Relay And Tunnel Ready](01a-relay-subscription-ready.md)
3. [PreviewConfig Revision](01b-preview-config-revision.md)
4. [Tunnel 重连与控制租约](02-tunnel-and-control-lease.md)
5. [Mobile Access API 与 Web](03-mobile-access-web.md)
6. [验证、发布和回滚](04-verification-release.md)

## 原子任务

| Task | 契约 | 完成证据 |
|---|---|---|
| 1 | Pod subscription fail-closed | Connect tests |
| 1b | Runner publisher/tunnel ready acknowledgement | Runner + Backend protocol tests |
| 2 | Preview session fail-closed、无裸 Token | Gin tests |
| 3 | Preview path/config contract | service/relay tests |
| 4 | PreviewConfig revision 写入与应用 | migration + service + proxy tests |
| 5 | Runner tunnel 自动重连 | runner tunnel tests |
| 6 | Relay 单写者控制租约 | Relay + Rust tests |
| 7 | Mobile Access Descriptor | proto + Connect tests |
| 8 | 移动 Worker 列表和可见入口 | Vitest |
| 9 | PTY/ACP/Preview 移动状态 | Vitest + browser |
| 10 | 全链路验证与发布 | Go/Rust/Web/E2E |

## 提交策略

每个任务：

1. 写一个最小失败测试。
2. 运行并确认因缺失行为失败。
3. 写最小实现。
4. 运行定向测试。
5. 运行相邻回归测试。
6. Spec review。
7. Code quality review。
8. 原子 commit。

最终提交默认 push `main`，但仅在所有验证通过后执行。
