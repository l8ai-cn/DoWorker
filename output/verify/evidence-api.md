# Hive 修复（T1–T5）真实接口核验证据 — 执行 Agent

> 执行时间：2026-07-05 02:20–02:30 (UTC+8)
> 环境：backend http://localhost:10015 · postgres 容器 agentsmesh-main-postgres-1 · runner 容器 agentsmesh-main-runner-e2e-echo-1
> 登录账号：devuser / dev-org（token 长度 300，正常获取）

## 汇总

| # | 项目 | 判定 |
|---|------|------|
| V1 | schema_migrations 一致性 | **PASS** |
| V1b | preflight 防线（负面测试） | **PASS** |
| V2 | deny 热推送行为 | **PASS** |
| V2b | 对照组 elicitation | **PASS**（首次尝试异常，重试通过，见异常记录） |
| V3 | 真实 usage 透传 | **PASS** |
| V3b | 定价交叉核验 | **PASS**（diff = 0.0） |
| V4 | 估算 fallback 保护 | **PASS** |
| V5 | CI 等价执行 | **PASS** |

---

## V1 迁移表一致性

【命令】
```bash
docker exec agentsmesh-main-postgres-1 psql -U agentsmesh -d agentsmesh -tAc \
  "SELECT count(*), bool_or(dirty), max(version) FROM schema_migrations;"
```

【原始输出】
```
1|f|166
```

【判定】**PASS** — 与预期 `1|f|166` 完全一致（单行、非 dirty、版本 166）。

---

## V1b preflight 防线（负面测试）

【命令 1：插入假行】
```bash
docker exec agentsmesh-main-postgres-1 psql -U agentsmesh -d agentsmesh -tAc \
  "INSERT INTO schema_migrations (version, dirty) VALUES (999, false);"
```
【原始输出 1】
```
INSERT 0 1
```

【命令 2：跑 hive_smoke】
```bash
bash deploy/dev/hive_smoke.sh; echo "exit=$?"
```
【原始输出 2】
```
exit=1
hive smoke: schema_migrations is inconsistent (count|dirty = 2|f)
Fix: docker compose run --rm --no-deps migrate ... force <version> && ... up
```
（在任何 suite（S0–S3）开跑之前即退出，无 suite 输出。）

【命令 3：恢复现场并复查】
```bash
docker exec agentsmesh-main-postgres-1 psql -U agentsmesh -d agentsmesh -tAc \
  "DELETE FROM schema_migrations WHERE version = 999;"
docker exec agentsmesh-main-postgres-1 psql -U agentsmesh -d agentsmesh -tAc \
  "SELECT count(*), bool_or(dirty), max(version) FROM schema_migrations;"
```
【原始输出 3】
```
DELETE 1
1|f|166
```

【判定】**PASS** — 非零退出（exit=1）、stdout 含 "schema_migrations is inconsistent"、fail-fast 于任何 suite 之前；现场已恢复到 `1|f|166`。

---

## V2 deny 热推送行为

session：`conv_02b4f170f92145ab`（pod `1-standalone-bab7685c`）

【命令 1：建 session】
```bash
curl -s -m 30 -X POST http://localhost:10015/v1/sessions \
  -H "Authorization: Bearer $TOKEN" -H "X-Organization-Slug: dev-org" -H "Content-Type: application/json" \
  -d '{"agent_id":"e2e-echo","title":"verify deny","scenario":"permission_request_edit"}'
```
【原始输出 1】
```json
{"id":"conv_02b4f170f92145ab","agent_id":"e2e-echo","agent_name":"e2e-echo","status":"launching","created_at":1783189293,"title":"verify deny","harness":"e2e-echo"}
```
轮询 status：`t=1s status=idle`（1 秒内离开 launching）。

【命令 2：建 deny policy（等 2 秒）】
```bash
curl -s -m 30 -X POST http://localhost:10015/v1/policies ... \
  -d '{"name":"verify-deny-edit","type":"python","handler":"acp_tool_rule","factory_params":{"tool_pattern":"Edit","verdict":"deny","priority":100}}'
```
【原始输出 2】
```json
{"created_at":1783189306,"created_by":null,"enabled":true,"factory_params":{"priority":100,"tool_pattern":"Edit","verdict":"deny"},"handler":"acp_tool_rule","id":"pol_18","name":"Edit","object":"default_policy","type":"python","updated_at":1783189306}
```
runner 日志佐证热推送到达：
```
time=18:21:46.264 level=INFO msg="Pod policy rules updated" module=pod pod_key=1-standalone-bab7685c rules=1
```

【命令 3：发消息】
```bash
curl ... -X POST .../v1/sessions/conv_02b4f170f92145ab/events \
  -d '{"type":"message","data":{"role":"user","content":[{"type":"input_text","text":"please edit"}]}}'
```
【原始输出 3】
```
{"item_id":"item_a77adde12024b655","queued":true}
HTTP=202
```

【命令 4：轮询 20 秒（每秒查 pending_elicitations + items 数）】
【原始输出 4】（20 行全部相同，节选首尾）
```
t=1s  pending=null items=2 status=idle
t=2s  pending=null items=2 status=idle
...
t=20s pending=null items=2 status=idle
```

【最终 items 内容】
```json
{
  "data": [
    {"id":"item_a77adde12024b655","role":"user","type":"message","status":"completed",
     "content":[{"text":"please edit","type":"text"}],"response_id":"resp_35f3d5f24dd789d1"},
    {"id":"item_7089c7f5dca99c19","role":"assistant","type":"message","status":"completed",
     "content":[{"text":"Will edit (with confirm): please edit","type":"text"}],"response_id":"resp_4b9cf348ad65fd00"}
  ],
  "has_more": false, "object": "list"
}
```

【命令 5：删除 policy】
```bash
curl -X DELETE .../v1/policies/pol_18   →   HTTP=204
```
runner 日志：`18:22:44.006 msg="Pod policy rules updated" pod_key=1-standalone-bab7685c rules=0`

【判定】**PASS** — 全程（20 秒 × 每秒）pending_elicitations 均为 null，未出现任何 elicitation；items=2（含 assistant 回复 "Will edit (with confirm): please edit"），说明 runner 收到热推送的 deny 规则后自动拒绝了 permission request（mock 场景在 deny 后照常 end_turn）。policy 已清理（204）。

---

## V2b 对照组（证明 V2 非空转）

【命令 1：确认无残留 deny 规则】
```bash
curl -s .../v1/policies | jq -c '[.data[]? | {id,name,handler,factory_params}] // .'
```
【原始输出 1】
```
[]
```

### 首次尝试（异常，已记录）— session `conv_33372829b3976d95`（pod `1-standalone-670cc0e0`）

- 建 session → `t=2s status=idle`；发消息 → `{"item_id":"item_dce717543a58d009","queued":true} HTTP=202`
- elicitation 立即出现（t=1s）：
```json
[{"elicitation_id":"elicit_378cc4ed08714e56","params":{"content_preview":"","message":"Tool: tc-mock-edit-perm-1","mode":"form","phase":"tool_call_approval","policy_name":"tool_call_approval"}}]
```
- resolve（`{"action":"accept","content":{}}`）→ `{"queued":true} HTTP=202`
- **异常**：其后轮询 35 秒 items 始终 =1（无 assistant 回复），session 终态 `status=idle, pending_elicitations=null, total_cost_usd=null`
- 原因分析（runner 日志）：
```
time=18:23:13.289 level=INFO msg="ACP responding to permission" module=pod pod_key=1-standalone-670cc0e0 request_id=9001 approved=true
```
  resolve 是异步 queued 的，本次从 API 202 到 runner 实际应答 permission 花了约 15 秒，而 mockagent 的 `permissionWaitTimeout` 只有 10 秒（`runner/internal/agents/mockagent/scenario_permission.go`），场景已按超时路径 end_turn，迟到的 approve 被丢弃，该 turn 的 assistant 文本未持久化、也未上报 usage（该 pod 在 pod_session_usage 中无行，与 V4 分析吻合）。

### 重试（按规则最多重试一次）— session `conv_514095c1d48ea379`（pod `1-standalone-3c530d12`）

【命令】完整脚本同流程（建 session → 发 "please edit" → 轮询 elicitation → resolve → 轮询 items）。
【原始输出】
```
CREATE: {"id":"conv_514095c1d48ea379","agent_id":"e2e-echo","agent_name":"e2e-echo","status":"launching","created_at":1783189551,"title":"verify elicitation control retry","harness":"e2e-echo"}
launch-poll t=1s status=launching
launch-poll t=2s status=idle
EVENT: {"item_id":"item_3cd74856cf05c69c","queued":true} HTTP=202
elicit-poll t=1s pending=[{"elicitation_id":"elicit_f201f45ac0f83095","params":{"content_preview":"","message":"Tool: tc-mock-edit-perm-1","mode":"form","phase":"tool_call_approval","policy_name":"tool_call_approval"}}]
EID=elicit_f201f45ac0f83095
RESOLVE: {"queued":true} HTTP=202
items-poll t=1s items=1
items-poll t=2s items=2
```
【最终 items】
```json
{
  "data": [
    {"id":"item_3cd74856cf05c69c","role":"user","content":[{"text":"please edit","type":"text"}], "status":"completed", "type":"message", "response_id":"resp_eda3f57a2f3e9485"},
    {"id":"item_be63615f8298184c","role":"assistant","content":[{"text":"Will edit (with confirm): please edit","type":"text"}], "status":"completed", "type":"message", "response_id":"resp_7aac99a5d9595d40"}
  ],
  "has_more": false, "object": "list"
}
```
【最终 session】
```json
{
  "status": "idle",
  "pending_elicitations": null,
  "total_cost_usd": 0.00000825,
  "usage_by_model": {
    "gpt-4o-mini": {"model":"gpt-4o-mini","input_tokens":19,"output_tokens":9,"cache_read_tokens":0,"cache_creation_tokens":0,"total_cost_usd":0.00000825}
  }
}
```

【判定】**PASS** — 无 deny 规则时 elicitation 必然出现（两次尝试都在 1 秒内出现，elicitation id 分别为 `elicit_378cc4ed08714e56` / `elicit_f201f45ac0f83095`），证明 V2 的"全程无 elicitation"确实是 deny 规则生效而非场景空转。重试会话 resolve(accept) 后 2 秒 turn 完成（items=2）。**为 V4 提供 fallback usage 数据的 session：`conv_514095c1d48ea379`（pod `1-standalone-3c530d12`）**。
⚠️ 异常记录：首次尝试暴露一个时序问题 — elicitation resolve 是异步投递，偶发传播延迟（本例 ~15s）超过 mock 场景 10s 的 permission 等待窗口时，approve 会被丢弃且该 turn 无 assistant 产出。属于测试基建/时序边界问题，不影响本项功能判定，建议审核 Agent 关注。

---

## V3 真实 usage 透传

session：`conv_70ac3fb30a15b821`（pod `1-standalone-224183dd`），默认 echo 场景。

【命令】建 session `{"agent_id":"e2e-echo","title":"verify usage"}` → 发 "usage check" → 轮询 items ≥2 → 轮询 total_cost_usd>0。
【原始输出】
```
CREATE: {"id":"conv_70ac3fb30a15b821","agent_id":"e2e-echo","agent_name":"e2e-echo","status":"launching","created_at":1783189583,"title":"verify usage","harness":"e2e-echo"}
launch-poll t=1s status=launching
launch-poll t=2s status=idle
EVENT: {"item_id":"item_e6ba32a0c861decc","queued":true} HTTP=202
items-poll t=1s items=1
items-poll t=2s items=2
cost-poll t=1s total_cost_usd=0.000085
```
【usage_by_model 完整 JSON】
```json
{
  "status": "idle",
  "total_cost_usd": 0.000085,
  "usage_by_model": {
    "gpt-4o": {
      "model": "gpt-4o",
      "input_tokens": 14,
      "output_tokens": 5,
      "cache_read_tokens": 0,
      "cache_creation_tokens": 0,
      "total_cost_usd": 0.000085
    }
  }
}
```

【判定】**PASS** — `usage_by_model` 含键 `gpt-4o`，input_tokens=14 > 0、output_tokens=5 > 0，`total_cost_usd=0.000085 > 0`（turn 完成后 1 秒内即可查到）。

---

## V3b 定价交叉核验

【命令 1：查单价】
```bash
docker exec agentsmesh-main-postgres-1 psql -U agentsmesh -d agentsmesh -tAc \
  "SELECT model, input_per_million, output_per_million FROM model_prices WHERE model='gpt-4o';"
```
【原始输出 1】
```
gpt-4o|2.500000|10.000000
```

【命令 2：手算（V3 tokens：input=14, output=5；cache_read/cache_creation 均为 0，不参与计费）】
```
cost = 14 × 2.5/1e6 + 5 × 10/1e6
     = 0.000035 + 0.000050
     = 8.5e-05
reported total_cost_usd = 0.000085
diff = 0.0
within 1e-9: True
```
（python3 实算输出：`manual = 14*2.5/1e6 + 5*10/1e6 = 8.5e-05`，`diff = 0.0`）

【判定】**PASS** — 手算值与 API 返回的 total_cost_usd 完全相等（diff=0.0，远小于 1e-9 容差），且单价 2.5/10 与预期一致。

---

## V4 估算 fallback 保护

【命令 1：最近 6 行 usage】
```bash
docker exec agentsmesh-main-postgres-1 psql -U agentsmesh -d agentsmesh -tAc \
  "SELECT pod_key, model, input_tokens, output_tokens, updated_at FROM pod_session_usage ORDER BY updated_at DESC LIMIT 6;"
```
（注：表实际主键列是 `pod_key` 而非任务书中写的 `pod_id`，无 `session_id` 列，已按实际 schema 调整。）

【原始输出 1】
```
1-standalone-224183dd|gpt-4o|14|5|2026-07-04 18:26:24.711999+00
1-standalone-3c530d12|gpt-4o-mini|19|9|2026-07-04 18:25:52.687445+00
1-standalone-bab7685c|gpt-4o-mini|19|9|2026-07-04 18:21:55.780257+00
1-standalone-a776f636|gpt-4o|16|6|2026-07-04 18:15:52.758264+00
1-standalone-5770751e|gpt-4o|29|11|2026-07-04 18:15:52.607732+00
1-standalone-eae3dcd2|gpt-4o-mini|21|11|2026-07-04 18:15:51.57091+00
```

【命令 2：session → pod 映射（确认行归属）】
```bash
docker exec ... psql -tAc "SELECT id, title, pod_key FROM agent_sessions WHERE id IN (...);"
```
【原始输出 2】
```
conv_02b4f170f92145ab|verify deny|1-standalone-bab7685c
conv_33372829b3976d95|verify elicitation control|1-standalone-670cc0e0
conv_514095c1d48ea379|verify elicitation control retry|1-standalone-3c530d12
conv_70ac3fb30a15b821|verify usage|1-standalone-224183dd
```
补充查询：四个 verify pod 中 `1-standalone-670cc0e0`（V2b 首次尝试）在 pod_session_usage 中**无行**。

【判定】**PASS** —
- V3 的 pod `224183dd`：model=`gpt-4o`，tokens 14/5 → **真实透传**路径 ✓
- V2 的 pod `bab7685c` 与 V2b 重试的 pod `3c530d12`：model=`gpt-4o-mini`，tokens 19/9 → permission 场景 mock 不上报 usage，走 **len/4 估算 fallback** ✓（两个 pod 消息文本相同故估算值相同，符合估算逻辑特征）
- V2b 首次尝试的 pod `670cc0e0` 无行：该 turn 因 resolve 迟到超时、assistant 文本为空未触发上报（与任务书预判的"turn 文本为空未触发上报"一致），已在 V2b 节分析；重试 pod 的行已补足证据链。

---

## V5 CI 等价执行

【命令】
```bash
bash deploy/dev/hive_smoke.sh; echo "exit=$?"
```

【原始输出】（完整，exit=0）
```
==========================================
  S0 message round-trip
==========================================
✓ GET /v1/agents — 9 agents in 3ms
✓ POST /v1/sessions — conv_e0480f5240806846 status=launching
✓ POST /v1/sessions/.../events — item=item_645422b827fda8d6
✓ GET /v1/sessions/.../items assistant — echo: Reply with exactly: pong
✓ GET /health?session_ids= — runner_online=true

==========================================
  S1 session wire + elicitation + terminal
==========================================
✓ API login
✓ S1.2 GET /v1/sessions list wire — status=idle updated_at=1783189629
✓ S1.2 WS /v1/sessions/updates — connected
✓ S1.3 elicitation created — elicit_71a45610e7896570
✓ S1.3 elicitation resolve — HTTP 202
✓ S1.3 assistant after resolve — items=2
✓ S1.4 list terminals — terminal_tui_main
✓ S1.4 terminal attach bytes — messages=2

==========================================
  S2 compat API
==========================================
✓ S2.1 GET /v1/harnesses — count=9
✓ S2.2 GET /v1/policy-registry
✓ S2.2 GET /v1/policies — count=0
✓ S2.2 POST /v1/policies — pol_19
✓ S2.2 DELETE /v1/policies/{id}
✓ S2.3 list permission_level + updated_at
✓ S2.3 PUT read-state
✓ S2.3 GET permissions
✓ S2.3 GET owner
✓ S2.5 switch-agent 501
✓ S2.5 POST directories 501

S2 smoke: all steps passed

==========================================
  S3 hive mechanism
==========================================
✓ API login
✓ S3.3 session created — conv_ec51a990e381d04b
✓ S3.3 user message sent — HTTP 202
✓ S3.3 assistant reply — items=2
✓ S3.3 total_cost_usd > 0 — $0.000110
✓ S3.3 usage_by_model from agent report — gpt-4o
✓ S3.2 policy create + hot push — active=conv_ec51a990e381d04b id=pol_20
✓ S3.2 policy delete + hot push
✓ S3.2 deny rule pushed to running pod — pol_21
✓ S3.2 runner auto-denied (no elicitation) — items=2
✓ S3.4 source conversation — items=3
✓ S3.4 POST fork — conv_ff9002967fef8889
✓ S3.4 fork copied items — fork=3 source=3
✓ S3.4 fork continues — items=5

hive smoke: all suites passed
```

【判定】**PASS** — exit 0，末行 "hive smoke: all suites passed"，S0/S1/S2/S3 全部套件通过（整个脚本约 29 秒完成）。注意 S1.3 elicitation 套件内 resolve 传播及时，同一机制在脚本内稳定通过，进一步佐证 V2b 首次尝试是偶发时序抖动。

---

## 异常发现汇总

1. **elicitation resolve 异步传播偶发延迟**（V2b 首次尝试）：resolve API 返回 202 queued 后，runner 实际应答 permission 延迟约 15 秒，超过 mockagent `permissionWaitTimeout`（10s），导致 approve 被丢弃、turn 以超时路径结束且 assistant 文本/usage 均未产出。重试即通过，hive_smoke S1.3 也稳定通过，判断为偶发时序抖动，但 10s 的 mock 等待窗口 vs 异步 resolve 无 SLA 是一个潜在 flaky 点。
2. **任务书 SQL 列名偏差**：`pod_session_usage` 实际列为 `pod_key`（无 `pod_id`/`session_id`），已按真实 schema 调整查询，不影响判定。
3. V1b 破坏性注入已完整恢复（复查 `1|f|166`），全部 verify 用 policy（pol_18）已删除，无环境残留。
