# Hive 审查修复（T1–T5）核验矩阵

> 角色：执行 Agent 产出证据 → 审核 Agent 独立复核 → 决策（父会话）终判
> 环境：backend http://localhost:10015 · web-user http://localhost:5173 ·
> postgres 容器 agentsmesh-main-postgres-1 · runner 容器 agentsmesh-main-runner-e2e-echo-1
> 登录：POST /proto.auth.v1.AuthService/Login {"username":"devuser","password":"AdminAb123456"}
> （加 header Connect-Protocol-Version: 1；后续请求带 Authorization: Bearer + X-Organization-Slug: dev-org）
> API 配方参考：output/hive-s1-smoke.mjs、output/hive-s3-smoke.mjs

| # | 功能 | 核验方法（真实接口/页面） | 预期证据 | 判定 |
|---|------|--------------------------|----------|------|
| V1 | T1 schema_migrations 一致性 | psql 查 `SELECT count(*), bool_or(dirty), max(version) FROM schema_migrations` | `1 | f | 166` | 待核 |
| V1b | T1 preflight 防线（负面测试） | 插入假行 `(999,false)` → `bash deploy/dev/hive_smoke.sh` → 应 exit 1 并打印 inconsistent → 删假行恢复 | fail-fast 生效且可恢复 | 待核 |
| V2 | T2 deny 热推送行为 | API：建 `scenario=permission_request_edit` session → 等 running → POST /v1/policies（tool_pattern=Edit, verdict=deny）→ 发消息 → 轮询 | pending_elicitations 始终为空 且 items≥2（runner 自动拒绝） | 待核 |
| V2b | T2 对照组（证明 V2 非空转） | 同场景、不推任何 deny 规则 → 发消息 → 轮询 | elicitation 必须出现；approve resolve 后 turn 正常完成 | 待核 |
| V3 | T3 真实 usage 透传 | API：e2e-echo 默认 echo 场景发消息 → GET session | `usage_by_model` 含 `gpt-4o`（tokens>0），`total_cost_usd>0` | 待核 |
| V3b | T3 定价交叉核验 | 用 V3 的 tokens 数 × model_prices 里 gpt-4o 单价（in 2.5 / out 10 per 1M）手算 | 手算值 == total_cost_usd（浮点容差 1e-9） | 待核 |
| V4 | T3 估算 fallback 保护 | V2b 的 session（mock 该场景不上报 usage）turn 完成后查 DB `pod_session_usage` | 该 pod 行 model=`gpt-4o-mini`（fallback），而 V3 的 pod 行 model=`gpt-4o` | 待核 |
| V5 | T5 CI job 等价执行 | 本地执行 CI 同款命令 `bash deploy/dev/hive_smoke.sh`（含 S0–S3 全部套件） | `hive smoke: all suites passed` | 待核 |
| V6 | 整体页面闭环 | 浏览器：web-user 登录 → 选 agent → 建会话 → 发消息 → 气泡渲染 echo 回复 | 对话闭环成功、无阻断性 console error | 待核 |

## 证据存放
- 执行 Agent：output/verify/evidence-api.md（每项贴原始请求/响应/SQL 输出）
- 浏览器 Agent：output/verify/evidence-browser.md + 截图
- 审核 Agent：output/verify/audit-report.md
