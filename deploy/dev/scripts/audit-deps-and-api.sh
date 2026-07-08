#!/usr/bin/env bash
# Dependency + mobile API alignment audit for dev environment.
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEV_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="$DEV_DIR/.env"
[[ -f "$ENV_FILE" ]] && source "$ENV_FILE"

HTTP_PORT="${HTTP_PORT:-10000}"
BACKEND_PORT="${BACKEND_HTTP_PORT:-10015}"
GRPC_PORT="${BACKEND_GRPC_PORT:-10016}"
RELAY_PORT="${RELAY_HTTP_PORT:-10017}"
POSTGRES_PORT="${POSTGRES_PORT:-10002}"
REDIS_PORT="${REDIS_PORT:-10003}"
MOBILE_PORT="${MOBILE_LOVABLE_PORT:-10021}"
ORG="dev-org"
USER="dev@agentsmesh.local"
PASS="AdminAb123456"
BASE="http://localhost:${HTTP_PORT}"

pass() { echo "  ✅ $1"; OK=$((OK + 1)); }
fail() { echo "  ❌ $1"; FAIL=$((FAIL + 1)); }
warn() { echo "  ⚠️  $1"; WARN=$((WARN + 1)); }
section() { echo ""; echo "== $1 =="; }

OK=0
FAIL=0
WARN=0

section "1. 基础设施 (Docker)"
for svc in postgres redis minio traefik; do
  if docker ps --format '{{.Names}}' | grep -q "agentsmesh.*${svc}"; then
    pass "docker: $svc"
  else
    fail "docker: $svc not running"
  fi
done

section "2. Host 服务 (Bazel / 手动)"
if curl -sf --max-time 3 "http://localhost:${BACKEND_PORT}/health" >/dev/null; then
  pass "host backend :${BACKEND_PORT}/health"
else
  fail "host backend :${BACKEND_PORT} down (Traefik /v1 /auth 依赖此端口)"
fi
if curl -sf --max-time 3 "http://localhost:${RELAY_PORT}/health" >/dev/null; then
  pass "host relay :${RELAY_PORT}/health"
else
  fail "host relay :${RELAY_PORT} down (终端 WebSocket 依赖此端口)"
fi
if lsof -i ":${GRPC_PORT}" -sTCP:LISTEN >/dev/null 2>&1; then
  pass "host gRPC :${GRPC_PORT} listening (runner 控制面)"
else
  fail "host gRPC :${GRPC_PORT} not listening"
fi
if nc -z localhost "$POSTGRES_PORT" 2>/dev/null; then
  pass "postgres :${POSTGRES_PORT}"
else
  fail "postgres :${POSTGRES_PORT}"
fi
if nc -z localhost "$REDIS_PORT" 2>/dev/null; then
  pass "redis :${REDIS_PORT}"
else
  fail "redis :${REDIS_PORT}"
fi

section "3. Traefik 路由 (统一入口 :${HTTP_PORT})"
if curl -sf --max-time 3 "${BASE}/health" >/dev/null; then
  pass "Traefik → backend /health"
else
  fail "Traefik :${HTTP_PORT}/health (Bad Gateway = host backend 未就绪)"
fi

section "4. Runner 容器 (数据面 pod 执行)"
CODEX_RUNNERS=$(docker ps --format '{{.Names}}' | grep -c 'runner-codex-cli' || true)
[[ "$CODEX_RUNNERS" -ge 1 ]] && pass "$CODEX_RUNNERS codex-cli runner(s)" || fail "no codex-cli runner"

ONLINE=$(docker exec agentsmesh-main-postgres-1 psql -U agentsmesh -d agentsmesh -tAc \
  "SELECT count(*) FROM runners WHERE status='online' AND last_heartbeat > now()-interval '60 seconds'" 2>/dev/null | tr -d ' ' || echo 0)
[[ "${ONLINE:-0}" -ge 1 ]] && pass "$ONLINE runner(s) gRPC online (DB heartbeat)" || fail "no runner gRPC heartbeat in last 60s"

section "5. Mobile dev (:${MOBILE_PORT})"
if curl -sf --max-time 3 -o /dev/null "http://localhost:${MOBILE_PORT}/" 2>/dev/null; then
  pass "mobile-lovable :${MOBILE_PORT}"
else
  warn "mobile-lovable :${MOBILE_PORT} not running (bazel run //clients/mobile-lovable:dev 或 dev:up)"
fi

section "6. Mobile API 对齐 (经 Traefik ${BASE})"
TOKEN=$(curl -sS --max-time 15 -X POST "${BASE}/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER\",\"password\":\"$PASS\"}" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin).get("token",""))' 2>/dev/null || true)
[[ -n "$TOKEN" ]] && pass "POST /auth/login" || { fail "POST /auth/login"; TOKEN=""; }

api_check() {
  local name=$1 method=$2 path=$3 expect=$4
  local body=${5:-}
  local code
  if [[ -n "$body" ]]; then
    code=$(curl -sS -o /dev/null -w '%{http_code}' --max-time 30 -X "$method" "${BASE}${path}" \
      -H "Authorization: Bearer $TOKEN" -H "X-Organization-Slug: $ORG" \
      -H 'Content-Type: application/json' -d "$body" 2>/dev/null || echo 000)
  else
    code=$(curl -sS -o /dev/null -w '%{http_code}' --max-time 30 -X "$method" "${BASE}${path}" \
      -H "Authorization: Bearer $TOKEN" -H "X-Organization-Slug: $ORG" 2>/dev/null || echo 000)
  fi
  [[ "$code" == "$expect" ]] && pass "$name ($code)" || fail "$name ($code, expect $expect)"
}

if [[ -n "$TOKEN" ]]; then
  api_check "GET /v1/agents" GET "/v1/agents" 200
  api_check "GET /v1/sessions" GET "/v1/sessions?limit=5" 200
  api_check "GET /v1/sessions/projects" GET "/v1/sessions/projects" 200
  api_check "GET /api/v1/orgs/:slug/experts" GET "/api/v1/orgs/${ORG}/experts?limit=5" 200

  CREATE=$(curl -sS --max-time 60 -X POST "${BASE}/v1/sessions" \
    -H "Authorization: Bearer $TOKEN" -H "X-Organization-Slug: $ORG" \
    -H 'Content-Type: application/json' \
    -d '{"agent_id":"e2e-echo","title":"api-audit"}' 2>/dev/null || echo '{}')
  SID=$(echo "$CREATE" | python3 -c 'import sys,json; print(json.load(sys.stdin).get("id",""))' 2>/dev/null || true)
  if [[ -n "$SID" ]]; then
    pass "POST /v1/sessions (e2e-echo) sid=$SID"
    api_check "GET /v1/sessions/:id" GET "/v1/sessions/${SID}" 200
    api_check "GET /v1/sessions/:id/items" GET "/v1/sessions/${SID}/items?limit=5&order=desc" 200
    api_check "GET /v1/sessions/:id/resources/terminals" GET "/v1/sessions/${SID}/resources/terminals?order=asc&limit=100" 200
    EVENT='{"type":"message","data":{"role":"user","content":[{"type":"input_text","text":"audit"}]}}'
    api_check "POST /v1/sessions/:id/events" POST "/v1/sessions/${SID}/events" 202 "$EVENT"
    SC=$(curl -sS -o /dev/null -w '%{http_code}' --max-time 3 -N \
      -H "Authorization: Bearer $TOKEN" -H "X-Organization-Slug: $ORG" \
      -H 'Accept: text/event-stream' "${BASE}/v1/sessions/${SID}/stream" 2>/dev/null || echo 000)
    SC="${SC:0:3}"
    [[ "$SC" == "200" ]] && pass "GET /v1/sessions/:id/stream SSE ($SC)" || warn "GET stream ($SC, may timeout quickly)"
  else
    fail "POST /v1/sessions: $CREATE"
  fi
fi

section "7. 双 Backend 检测"
DOCKER_BACKEND=$(docker ps --format '{{.Names}}' | grep -c '^agentsmesh-backend-1$' || true)
if [[ "$DOCKER_BACKEND" -ge 1 ]]; then
  warn "agentsmesh-backend-1 (docker) 仍在运行 — dev 模式应以 host :${BACKEND_PORT} 为 SSOT；docker backend 不参与 Traefik 路由但占资源"
fi

section "8. OpenAI 凭据 (Codex 真实执行)"
KEY_OK=0
for c in $(docker ps --format '{{.Names}}' | grep 'runner-codex-cli' | head -1); do
  if docker exec "$c" test -s /home/runner/.codex/auth.json 2>/dev/null; then
    pass "codex auth.json in $c"
    KEY_OK=1
  fi
done
BUNDLE=$(docker exec agentsmesh-main-postgres-1 psql -U agentsmesh -d agentsmesh -tAc \
  "SELECT id FROM ai_models WHERE organization_id=1 AND provider_type='openai' AND is_enabled=true LIMIT 1" 2>/dev/null | tr -d ' ')
[[ -n "$BUNDLE" ]] && pass "openai model in ai_models pool (id=$BUNDLE)" \
  || fail "missing openai ai_models row — run backend once for dev seed or add model via API"

echo ""
echo "Summary: ✅ $OK  ❌ $FAIL  ⚠️  $WARN"
[[ "$FAIL" -eq 0 ]] && exit 0 || exit 1
