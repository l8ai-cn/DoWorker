# 终审决策（决策角色：父会话）

日期：2026-07-05 · 参与角色：执行 Agent（API 实测）→ 浏览器实测（chromium）→ 审核 Agent（独立复核+复跑）→ 决策

## 判定：开发完成，验证通过（全部 10 项 PASS）

| # | 功能 | 核验方式 | 判定 | 复核 |
|---|------|----------|------|------|
| V1 | schema_migrations 一致性 | psql 实查 `1\|f\|166` | PASS | 审核复跑一致 |
| V1b | preflight 防线 | 注入假行→smoke exit 1 fail-fast→恢复现场 | PASS | 审核确认 999 行已删 |
| V2 | policy deny 热推送 | 在跑 Pod 推 deny→无 elicitation、turn 完成 | PASS | runner 日志 rules=1 佐证 |
| V2+ | deny 方向断言（审核修复项） | assistant 文本 "Edit denied: skipped." | PASS | 修复后 smoke 全绿 |
| V2b | 对照组 | 无规则时 elicitation 1 秒内必现，resolve 后完成 | PASS | 证明 V2 非空转 |
| V3 | 真实 usage 透传 | usage_by_model 含 gpt-4o（14/5 tokens），cost>0 | PASS | 审核独立登录复查一致 |
| V3b | 定价交叉核验 | 手算 14×2.5/1e6+5×10/1e6=0.000085，diff=0 | PASS | 审核补验 fallback 定价也吻合 |
| V4 | 估算 fallback 保护 | pod_session_usage 同时存在 gpt-4o 与 gpt-4o-mini 行 | PASS | 审核复跑 SQL 一致 |
| V5 | CI 等价执行 | `bash deploy/dev/hive_smoke.sh` S0–S3 全绿 exit 0 | PASS | 修复 pod 泄漏后可重复执行 |
| V6 | 浏览器页面闭环 | 真 chromium：登录→建会话→发消息→echo 气泡渲染+API 双向核对 | PASS | 审核目检截图+复查 items |

## 审核发现 → 当场修复（已回归验证）

1. **V2 缺 deny 方向断言**（deny/allow 在 items 层表现相同）→ mock 场景在
   verdict 分支各发一条方向文本（"Edit denied: skipped." / "Edit approved:
   applied."），smoke 新增 `S3.2 verdict direction is deny` 断言 → 复跑全绿。
2. **测试 pod 泄漏导致套件非幂等**（审核复跑 V5 时 503 no available runner，
   runner 内存 slots 被 DB UPDATE 绕过）→ 新增 `output/terminate-all-pods.mjs`
   走真实 `PodService/TerminatePod`，`hive_smoke.sh` preflight 先 API 终止再
   DB 兜底 → 验证一次终止 6 个 live pod、DB 活跃数归零、连续重跑全绿。
3. **flaky 点（审核定级"中"）**：elicitation resolve 传播实测最坏 ~15s，
   mock 等待窗口 10s → 已升到 30s。产品层建议（未在本轮范围）：给 resolve
   链路加传播耗时观测。

## 证据链
- output/verify/verification-matrix.md（矩阵）
- output/verify/evidence-api.md（执行 Agent：8 项原始命令+输出）
- output/verify/evidence-browser.md + output/browser-integration/*.png（浏览器）
- output/verify/audit-report.md（审核 Agent 独立复核）
