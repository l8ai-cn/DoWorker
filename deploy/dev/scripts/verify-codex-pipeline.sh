#!/usr/bin/env bash
# Codex ACP pipeline — step-by-step health check + smoke test.
# Usage: deploy/dev/scripts/verify-codex-pipeline.sh [--run-task]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEV_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="$DEV_DIR/.env"

if [[ -f "$ENV_FILE" ]]; then
  # shellcheck source=/dev/null
  source "$ENV_FILE"
fi

API_PORT="${BACKEND_HTTP_PORT:-10015}"
HTTP_PORT="${HTTP_PORT:-10000}"
PRIMARY="${PRIMARY_DOMAIN:-localhost:${HTTP_PORT}}"
ORG="dev-org"
USER="dev@agentcloud.local"
PASS="AdminAb123456"
RUN_TASK=false
[[ "${1:-}" == "--run-task" ]] && RUN_TASK=true

pass() { echo "  ✅ $1"; }
fail() { echo "  ❌ $1"; FAILED=1; }

FAILED=0
step() { echo ""; echo "== Step $1: $2 =="; }

pick_base() {
  for url in "http://localhost:${API_PORT}" "http://${PRIMARY}"; do
    if curl -sf --max-time 5 "${url}/health" >/dev/null 2>&1; then
      echo "$url"
      return 0
    fi
  done
  return 1
}

step 0 "Backend 可达 + 清理陈旧 Worker（避免 quota_exceeded）"
BASE=""
for attempt in 1 2 3 4 5 6 7 8 9 10; do
  if BASE=$(pick_base); then
    pass "backend health OK at $BASE"
    break
  fi
  echo "  … waiting for backend (attempt $attempt)"
  sleep 3
done
if [[ -z "$BASE" ]]; then
  fail "backend not reachable on :${API_PORT} or ${PRIMARY}"
  echo "  Hint: go build ./backend/cmd/server && start host backend (see deploy/dev/lib/host_services.sh)"
  exit 1
fi

step 1 "登录 (/auth/login)"
TOKEN=""
for attempt in 1 2 3 4 5; do
  RESP=$(curl -sS --max-time 30 -X POST "$BASE/auth/login" \
    -H 'Content-Type: application/json' \
    -d "{\"username\":\"$USER\",\"password\":\"$PASS\"}" || true)
  TOKEN=$(echo "$RESP" | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d.get("token",""))' 2>/dev/null || true)
  if [[ -n "$TOKEN" ]]; then
    pass "login OK (attempt $attempt)"
    break
  fi
  echo "  … retry $attempt: $(echo "$RESP" | head -c 120)"
  sleep 2
done
[[ -z "$TOKEN" ]] && fail "login failed on $BASE/auth/login" && exit 1

HDR=(-H "Authorization: Bearer $TOKEN" -H "X-Organization-Slug: $ORG" -H "Content-Type: application/json")

RUNNING=$(curl -sS --max-time 15 "${HDR[@]}" "$BASE/v1/sessions?limit=50" 2>/dev/null \
  | python3 -c 'import sys,json; print(sum(1 for s in json.load(sys.stdin).get("data",[]) if s.get("status") in ("running","launching")))' 2>/dev/null || echo 0)
if [[ "${RUNNING:-0}" -ge 8 ]]; then
  echo "  … cleaning $RUNNING active sessions (pod quota headroom)"
  for sid in $(curl -sS --max-time 15 "${HDR[@]}" "$BASE/v1/sessions?limit=50" 2>/dev/null \
    | python3 -c 'import sys,json; print(" ".join(s["id"] for s in json.load(sys.stdin).get("data",[])))' 2>/dev/null); do
    curl -sS -o /dev/null -X DELETE "${HDR[@]}" "$BASE/v1/sessions/$sid" 2>/dev/null || true
  done
  pass "freed pod quota (deleted stale sessions)"
fi

step 2 "Runner / codex-cli 在线"
AGENTS=$(curl -sS --max-time 15 "${HDR[@]}" "$BASE/v1/agents" || echo '{}')
echo "$AGENTS" | python3 -c '
import sys,json
d=json.load(sys.stdin)
agents=d if isinstance(d,list) else d.get("data",d.get("agents",[]))
slugs=[a.get("id") or a.get("slug") for a in agents if isinstance(a,dict)]
print("agents:", slugs[:8])
sys.exit(0 if "codex-cli" in slugs else 1)
' && pass "codex-cli registered" || fail "codex-cli not in /v1/agents"

RUNNER_UP=$(docker ps --format '{{.Names}}' | grep -c 'runner-codex-cli' || true)
[[ "$RUNNER_UP" -ge 1 ]] && pass "$RUNNER_UP codex runner container(s) up" || fail "no codex runner containers"

step 3 "模型池（ai_models — Worker 创建时引用并注入配置）"
MODEL_ROW=$(docker exec agentcloud-main-postgres-1 psql -U agentcloud -d agentcloud -tAc \
  "SELECT id||'|'||name||'|'||provider_type FROM ai_models WHERE organization_id=1 AND provider_type='openai' AND is_enabled=true ORDER BY is_default DESC, id LIMIT 1" 2>/dev/null | tr -d ' ' || true)
[[ -n "$MODEL_ROW" ]] && pass "openai model in pool: $MODEL_ROW" || fail "no openai model in ai_models - run deploy/dev/scripts/seed-model-pool-from-local.py"

step "3b" "模型连接测试（pool 源 → provider base_url）"
API_CHECK=0
CONN=$(python3 <<'PY'
import json, os, re, urllib.request
home = os.path.expanduser("~")
key = ""
try:
    key = json.load(open(os.path.join(home, ".codex", "auth.json"))).get("OPENAI_API_KEY", "").strip()
except OSError:
    pass
base = "https://api.openai.com"
cfg_path = os.path.join(home, ".codex", "config.toml")
if os.path.isfile(cfg_path):
    m = re.search(r'base_url\s*=\s*"([^"]+)"', open(cfg_path).read())
    if m:
        base = m.group(1).rstrip("/")
if not key:
    print("MISSING|no ~/.codex/auth.json key (run deploy/dev/scripts/seed-model-pool-from-local.py)")
    raise SystemExit(0)
url = f"{base}/v1/models"
req = urllib.request.Request(url, headers={"Authorization": f"Bearer {key}"})
try:
    with urllib.request.urlopen(req, timeout=20) as r:
        print(f"OK|{r.status}|{base}")
except urllib.error.HTTPError as e:
    print(f"FAIL|{e.code}|{base}")
except Exception as e:
    print(f"ERR|{type(e).__name__}|{base}")
PY
)
IFS='|' read -r CONN_STATUS CONN_CODE CONN_BASE <<< "$CONN"
case "$CONN_STATUS" in
  OK)
    pass "provider accepted key (HTTP $CONN_CODE @ $CONN_BASE) — Worker 创建时从 ai_models 注入"
    API_CHECK=1
    ;;
  FAIL)
    fail "provider rejected key (HTTP $CONN_CODE @ $CONN_BASE). Add/update openai model via POST /v1/model-configs or ~/.codex + one-time backend seed"
    ;;
  MISSING)
    fail "$CONN_CODE"
    ;;
  *)
    fail "connection test error: $CONN ($CONN_BASE)"
    ;;
esac
[[ "$API_CHECK" -eq 1 ]] || true

step 4 "创建 Worker（codex-cli + 模型池）"
OPENAI_MODEL_ID=$(docker exec agentcloud-main-postgres-1 psql -U agentcloud -d agentcloud -tAc \
  "SELECT id FROM ai_models WHERE organization_id=1 AND provider_type='openai' AND is_enabled=true ORDER BY is_default DESC, id LIMIT 1" 2>/dev/null | tr -d ' ' || true)
CREATE_BODY='{"agent_id":"codex-cli","title":"Codex pipeline verify"}'
if [[ -n "$OPENAI_MODEL_ID" ]]; then
  CREATE_BODY=$(OPENAI_MODEL_ID="$OPENAI_MODEL_ID" python3 -c 'import json,os; print(json.dumps({"agent_id":"codex-cli","title":"Codex pipeline verify","model_config_id":int(os.environ["OPENAI_MODEL_ID"])}))')
fi
CREATE=$(curl -sS --max-time 60 "${HDR[@]}" -X POST "$BASE/v1/sessions" -d "$CREATE_BODY")
SID=$(echo "$CREATE" | python3 -c 'import sys,json; print(json.load(sys.stdin).get("id",""))' 2>/dev/null || true)
[[ -n "$SID" ]] && pass "session $SID" || { fail "create session: $CREATE"; exit 1; }

POD=$(docker exec agentcloud-main-postgres-1 psql -U agentcloud -d agentcloud -tAc \
  "SELECT pod_key FROM agent_sessions WHERE id='$SID'" 2>/dev/null | tr -d ' ' || true)
[[ -n "$POD" ]] && pass "pod $POD" || fail "no pod_key in DB"

step 5 "发送用户消息 (POST /events)"
if [[ "$RUN_TASK" == true ]]; then
  pass "skip hello — step 6 sends gomoku task only"
else
MSG='你好 Codex，请用中文回复一句话确认在线。'
EVENT_BODY=$(MSG="$MSG" python3 <<'PY'
import json, os
print(json.dumps({
    "type": "message",
    "data": {
        "role": "user",
        "content": [{"type": "input_text", "text": os.environ["MSG"]}],
    },
}))
PY
)
curl -sS --max-time 30 "${HDR[@]}" -X POST "$BASE/v1/sessions/$SID/events" \
  -d "$EVENT_BODY" \
  | python3 -c 'import sys,json; d=json.load(sys.stdin); assert d.get("item_id") or d.get("queued"); print(d)' \
  && pass "user message queued" || fail "post message failed"

ITEMS=$(curl -sS --max-time 15 "${HDR[@]}" "$BASE/v1/sessions/$SID/items?limit=5&order=asc")
echo "$ITEMS" | python3 -c '
import sys,json
n=len(json.load(sys.stdin).get("data",[]))
print("user items:", n)
sys.exit(0 if n>=1 else 1)
' && pass "user item persisted" || fail "user item missing in DB"
fi

if [[ "$RUN_TASK" == false ]]; then
  step 6 "等待 assistant 回复（最多 120s）"
else
  step 6 "派发五子棋任务并等待 index.html（最多 180s）"
  PROMPT='在 workspace 根目录创建 index.html，实现可浏览器运行的双人对战五子棋（15x15）。完成后回复：已创建 index.html'
  TASK_BODY=$(PROMPT="$PROMPT" python3 <<'PY'
import json, os
print(json.dumps({
    "type": "message",
    "data": {
        "role": "user",
        "content": [{"type": "input_text", "text": os.environ["PROMPT"]}],
    },
}))
PY
)
  curl -sS --max-time 30 "${HDR[@]}" -X POST "$BASE/v1/sessions/$SID/events" \
    -d "$TASK_BODY" \
    | python3 -c 'import sys,json; d=json.load(sys.stdin); assert d.get("item_id") or d.get("queued"); print(d)' \
    && pass "gomoku task queued" || fail "post gomoku task failed"
fi

MAX_POLLS=$([[ "$RUN_TASK" == true ]] && echo 72 || echo 24)
ASSIST=0
HTML=0
for i in $(seq 1 "$MAX_POLLS"); do
  sleep 5
  ST=$(curl -sS --max-time 15 "${HDR[@]}" "$BASE/v1/sessions/$SID" \
    | python3 -c 'import sys,json; print(json.load(sys.stdin).get("status",""))' 2>/dev/null || echo "?")
  OUT=$(curl -sS --max-time 15 "${HDR[@]}" "$BASE/v1/sessions/$SID/items?limit=20&order=asc" 2>/dev/null || echo '{}')
  echo "$OUT" | python3 -c "
import sys,json
for it in json.load(sys.stdin).get('data',[]):
  role=it.get('role')
  if role=='assistant':
    t=''.join(c.get('text','') for c in it.get('content',[]))
    if t: print('ASSIST:', t[:300])
  if it.get('type')=='error':
    print('ERROR:', it.get('message','')[:300])
" 2>/dev/null || true
  HAS_ERR=$(echo "$OUT" | python3 -c 'import sys,json; print(1 if any(i.get("type")=="error" for i in json.load(sys.stdin).get("data",[])) else 0)' 2>/dev/null || echo 0)
  if [[ "$HAS_ERR" -eq 1 && "$ASSIST" -eq 0 ]]; then
    echo "  ⚠ Codex returned error items (check OPENAI_API_KEY or network)"
  fi
  ASSIST=$(echo "$OUT" | python3 -c '
import sys,json
items=json.load(sys.stdin).get("data",[])
for i in items:
  if i.get("role")=="assistant":
    t="".join(c.get("text","") for c in i.get("content",[]))
    if t: sys.exit(0)
  if i.get("type") in ("function_call","function_call_output"):
    sys.exit(0)
sys.exit(1)
' 2>/dev/null && echo 1 || echo 0)
  for c in $(docker ps --format '{{.Names}}' | grep 'runner-codex-cli'); do
    if [[ -n "$POD" ]] && docker exec "$c" test -f "/workspace/repos/sandboxes/$POD/workspace/index.html" 2>/dev/null; then
      HTML=1
      wc=$(docker exec "$c" wc -c "/workspace/repos/sandboxes/$POD/workspace/index.html" | awk '{print $1}')
      pass "index.html found ($wc bytes) on $c"
      break 2
    fi
  done
  echo "  poll $i status=$ST assist=$ASSIST html=$HTML"
  if [[ "$RUN_TASK" == true ]]; then
    [[ "$HTML" -eq 1 ]] && break
  else
    [[ "$ASSIST" -eq 1 ]] && ASSIST=2 && break
  fi
done

[[ "$ASSIST" -ge 1 ]] && pass "assistant message in items" || fail "no assistant reply (Codex turn empty or events not bridged)"

if [[ "$RUN_TASK" == true ]]; then
  [[ "$HTML" -eq 1 ]] && pass "gomoku index.html created" || fail "index.html not found in sandbox"
fi

echo ""
if [[ "$FAILED" -eq 0 ]]; then
  echo "All steps passed. SID=$SID POD=$POD BASE=$BASE"
  exit 0
fi
echo "Pipeline check failed. SID=$SID POD=$POD BASE=$BASE"
echo "Hints: tail -f deploy/dev/runtime/backend/backend.log"
echo "       docker logs agentcloud-main-runner-codex-cli-2-1 --since 5m"
exit 1
