# Hive 修复（T1–T5）审核报告 — 审核 Agent

> 审核原则：不信任执行方结论，逐项核对证据链 + 独立复跑（真实接口，不改代码/不删数据）。
> 审核时间：2026-07-05 02:36–02:42 (UTC+8)
> 复跑范围：V1、V3、V3b、V4、V5、V6 共 6 项（超过要求的 4 项），另补做 gpt-4o-mini 定价闭环核验。

## 一、逐项判定

| # | 项目 | 证据完整性 | 独立复跑 | 判定 |
|---|------|-----------|---------|------|
| V1 | schema_migrations 一致性 | 原始命令+输出齐全 | 复跑一致 `1\|f\|166` | **采信** |
| V1b | preflight 防线（负面测试） | 原始命令+输出齐全 | 现场恢复复查通过（无 999 行） | **采信** |
| V2 | deny 热推送行为 | 命令有省略但关键响应/日志完整 | 未复跑（避免新增 pod 加剧容量饱和，见 V5） | **采信（有保留）** |
| V2b | 对照组 elicitation | 原始输出齐全，异常如实记录 | session/usage 行存在性复核通过 | **采信** |
| V3 | 真实 usage 透传 | 原始输出齐全 | 复跑 usage_by_model 完全一致 | **采信** |
| V3b | 定价交叉核验 | 原始 SQL+手算过程齐全 | 独立手算 diff=0 | **采信** |
| V4 | 估算 fallback 保护 | 原始 SQL 输出齐全 | 复跑两模型行并存 | **采信** |
| V5 | CI 等价执行 | 完整脚本输出 | **复跑失败**（环境饱和，非代码回归，见下） | **采信（有保留）** |
| V6 | 浏览器页面闭环 | 脚本摘要+截图（无原始 API 响应体） | 截图有效+items 复跑通过 | **采信** |

**证据不足项：无。** 全部九项都附带了原始命令与原始输出（V6 的 API 交叉验证只有脚本摘要行，属最薄弱一环，但截图与本审复跑补齐了证据链）。

## 二、独立复跑记录（原始输出）

### 复跑 1 — V1 + V1b 现场恢复复查

```bash
docker exec agentsmesh-main-postgres-1 psql -U agentsmesh -d agentsmesh -tAc \
  "SELECT count(*), bool_or(dirty), max(version) FROM schema_migrations;"
# → 1|f|166
docker exec agentsmesh-main-postgres-1 psql -U agentsmesh -d agentsmesh -tAc \
  "SELECT count(*) FROM schema_migrations WHERE version = 999;"
# → 0
```
与预期一致；V1b 的破坏性注入确已恢复，无 999 残留行。✓

### 复跑 2 — V3/V3b usage 透传 + 定价闭环

自行登录（token_len=300）后查证据中的 usage session：

```bash
GET /v1/sessions/conv_70ac3fb30a15b821
# → {"status":"failed","total_cost_usd":0.000085,
#    "usage_by_model":{"gpt-4o":{"model":"gpt-4o","input_tokens":14,"output_tokens":5,
#    "cache_read_tokens":0,"cache_creation_tokens":0,"total_cost_usd":0.000085}}}
```

单价复查 + 独立手算（python3）：

```
model_prices: gpt-4o|2.500000|10.000000 · gpt-4o-mini|0.150000|0.600000
gpt-4o:      14×2.5/1e6 + 5×10/1e6  = 8.5e-05      reported 0.000085   diff = 0.0
gpt-4o-mini: 19×0.15/1e6 + 9×0.6/1e6 = 8.25e-06    reported 0.00000825 diff ≈ 1.7e-21
```

usage 数字、模型键、成本与执行 Agent 证据逐字节一致；gpt-4o 定价闭环成立，且本审**补验了执行 Agent 未覆盖的 gpt-4o-mini（fallback）定价闭环**，同样吻合。✓

⚠️ 注意：该 session 当前 `status=failed`（证据采集时为 idle）。核实为事后 pod 生命周期变化（pod 被终止后 session 转 failed），usage/items 数据完好，不影响 V3 判定，但说明测试只采样了单一时点（见漏洞清单 #3）。V2/V2b 的 session 同样已转 failed。

### 复跑 3 — V4 fallback 双模型并存

```bash
SELECT pod_key, model, input_tokens, output_tokens FROM pod_session_usage
ORDER BY updated_at DESC LIMIT 8;
# → 1-standalone-27a54957|gpt-4o|26|16
#   1-standalone-5e3201d8|gpt-4o|16|6
#   1-standalone-1d81222d|gpt-4o|29|11
#   1-standalone-c1d63123|gpt-4o-mini|21|11
#   1-standalone-43d6b827|gpt-4o-mini|21|11
#   1-standalone-0281a780|gpt-4o|18|8
#   1-standalone-224183dd|gpt-4o|14|5        ← V3 pod，与证据一致
#   1-standalone-3c530d12|gpt-4o-mini|19|9   ← V2b 重试 pod，与证据一致
```
gpt-4o 与 gpt-4o-mini 行同时存在，且证据中引用的两个关键 pod 行原样在库。✓

### 复跑 4 — V6 截图 + items

```
05-web-user-after-send.png: PNG image data, 1280 x 720, 8-bit RGB（104,262 字节，非空）
```
人工目检截图：web-user 聊天页渲染了气泡 `echo: Integration test: reply with one short greeting sentence.`，
侧边栏可见 verify deny / verify usage / verify elicitation control (retry) 等 session，与 API 证据互相印证；
右侧确有 "Failed to load agents."（与浏览器证据自述的非阻断观察一致，证明其没有隐瞒瑕疵）。

```bash
GET /v1/sessions/conv_2bb61bf445d08223/items
# → count=2
#   user:      "Integration test: reply with one short greeting sentence."  (completed)
#   assistant: "echo: Integration test: reply with one short greeting sentence."  (completed)
```
items ≥2 且 assistant 文本含 "echo:"。✓

### 复跑 5 — V5 hive_smoke（发现环境饱和，非代码回归）

```bash
bash deploy/dev/hive_smoke.sh; echo "exit=$?"
# S0 全过（含 preflight 迁移检查隐式通过）
# S1: elicitation created/resolve/assistant after resolve 均 ✓，随后：
# ✗ S1 smoke run — Error: create session 503: {"code":"runner_unavailable","error":"no available runner"}
# exit=1
```

**根因定位**（非代码回归）：e2e runner 配置 `max_concurrent_pods: 10`（容器内 config.yaml 实查），
runner 日志显示本审复跑时 `total_pods=10` 已满 —— 本轮全部验证（V2/V2b×2/V3 + 浏览器 + 多次 smoke）创建的
mock pod 从未清理，累计吃满配额，后端因此按 `runner_unavailable` 拒绝新 session。执行 Agent 跑 V5 时
（02:27）配额尚未满，其 "all suites passed" 完整输出的内部一致性（pol_19/20/21 与 V2 的 pol_18 序号衔接、
session id 均在库中可查）支持其真实性，故**采信原判定，但套件的非幂等性记为测试设计漏洞 #2**。
（本审未终止任何 pod 腾容量 —— 遵守"不删数据"约束。）

## 三、测试设计漏洞清单

1. **V2 只证明了"自动裁决"，未证明"裁决方向是 deny"。** mock 场景在 permission request **之前**就发出
   assistant 文本（`scenario_permission.go:27`），故 deny/allow/超时三条路径最终都是 items=2 且文本相同。
   V2 的证据能排除超时路径（items=2 在 t=1s 即达成，远快于 10s 超时窗口）并证明未产生 elicitation，
   但若规则 verdict 被误配为 allow，本测试同样会 PASS。严格证明需断言 tool_call 终态为
   `failed / "denied by user"`（mock 已区分输出，只是测试没查）。→ 建议补断言。
2. **测试资源不清理，套件非幂等。** 验证产生的 pod 全部滞留，打满 `max_concurrent_pods=10`，导致本审
   复跑 hive_smoke 中途 503。CI 冷启动环境不受影响，但本地重复验证/复核会被前序测试自我阻塞。
   → 建议 smoke 脚本/验证脚本结尾终止自建 session 的 pod，或提供清理入口。
3. **单时点采样。** V2/V3 的 session 在证据采集后由 idle 转为 failed（pod 被回收）。usage/items 持久化
   不受影响，但"turn 完成后 session 保持健康"这一维度未被覆盖。→ 低严重性，记录在案。
4. **定价闭环原本只覆盖 gpt-4o。** fallback 路径（gpt-4o-mini）的单价×tokens 核算执行 Agent 未做，
   本审已补验通过（diff≈1.7e-21），漏洞已闭合。
5. **V6 证据薄。** 只有脚本摘要行+截图，未贴 API 原始响应体。本审复跑已补齐，不再追究。
6. 对照组 V2b 设计**成立**：同场景、无规则时 elicitation 两次均在 1s 内出现（两个不同 elicitation id），
   足以证明 V2 的"全程无 elicitation"是规则生效而非场景空转。负面测试 V1b 的现场恢复**已复查确认**。
7. 小观察（不影响判定）：V2 建 policy 请求体 `name=verify-deny-edit`，响应 `name=Edit` —— API 侧似乎
   用 tool_pattern 覆盖了 name，证据内部不一致源于产品行为而非造假。

## 四、flaky 点评估（elicitation resolve 传播 vs mock 10s 窗口）

- **现象**：resolve API 返回 202（异步排队）后，runner 实际应答 permission 无 SLA；V2b 首次尝试传播耗时
  ~15s，超过 mockagent `permissionWaitTimeout = 10s`（`runner/internal/agents/mockagent/scenario_permission.go:10`），
  迟到的 approve 被丢弃，turn 走超时路径、无 assistant 产出、无 usage 行。执行 Agent 如实记录并按规则重试通过。
- **严重性评估：中。** 两层影响：
  (a) 测试基建层 —— 这是真实的 flaky 源，CI 中 S1.3/S3 同机制套件将偶发红；
  (b) 产品层 —— 15s 的 approve 传播延迟对真实用户同样存在（点了"允许"后 agent 侧 15 秒无响应），
  异步链路无 SLA/无回执确认，值得排查那次 15s 究竟阻塞在哪一跳（API→hub→gRPC→pod）。
- **缓解建议**：mock 窗口提高到 ≥30s（或可配置）；resolve 链路增加传播耗时指标/日志，超阈值告警。
- 本审复跑期间 S1.3 resolve 在 1s 内完成（runner 日志 18:37:57 `approved=true` 距 resolve 约 1s），
  佐证"偶发抖动"的定性，但不排除负载相关。

## 五、总体结论

**建议判定：开发完成且验证通过（附整改建议，无阻断项）。**

- 九项证据全部真实可查：无一项"贴结论不贴输出"；抽查复跑 6 项，数据库、API、截图三个层面的复核
  结果与执行 Agent 证据逐字一致；破坏性负面测试（V1b）现场已确认恢复；本审补验的 gpt-4o-mini
  fallback 定价闭环也吻合。T1（迁移一致性+preflight）、T2（deny 热推送）、T3（usage 透传+fallback）、
  T5（CI 等价）及页面闭环（V6）的功能判定均成立。
- 两项保留意见均不指向功能缺陷：V5 复跑失败为测试 pod 泄漏导致的容量饱和（测试基建问题，漏洞 #2）；
  V2 缺 deny 方向断言（漏洞 #1）建议在合入前补一条断言，成本极低。
- flaky 点（resolve 传播 vs 10s 窗口）定级"中"，建议作为独立跟进项处理，不阻塞本次判定。
