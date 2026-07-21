# shellcheck shell=bash
# config_gen.sh — file generators (idempotent unless inputs change).
#
# Reads:
#   - `WORKTREE_NAME` / `PORT_OFFSET` (set by `generate_env`)
#   - host service ports from `.env` (backend, marketplace, relay)
#   - `SCRIPT_DIR` / `ENV_FILE` (entry-point globals)
# Writes:
#   - .env (worktree-scoped ports)
#   - clients/web/.env.local, clients/web-admin/.env.local
#   - traefik/traefik.yml + traefik/dynamic/{http,grpc}.yml
#   - ssl/* (delegates to generate-dev-certs.sh)
#   - runtime/access-token/* (Core Auth RS256 key pair)
#   - ai-cli-configs/{claude,codex,gemini}/* (claude code config etc)
#   - runtime/runner/certs/* (runner mTLS cert; only the host-side helper)
#   - runner-ssh/id_ed25519{,.pub}
#   - LOCAL_RUNNER_HOME: host-mode runner $HOME isolation so ~/.claude/* writes
#     don't clobber the developer's real config

generate_ssl_certs() {
    local ssl_dir="$SCRIPT_DIR/ssl"
    local need_regen=false

    if [[ ! -f "$ssl_dir/ca.crt" || ! -f "$ssl_dir/server.crt" ]]; then
        need_regen=true
    elif ! openssl x509 -in "$ssl_dir/ca.crt" -noout >/dev/null 2>&1; then
        warn "CA 证书损坏，将重新生成"
        need_regen=true
    elif ! openssl x509 -in "$ssl_dir/server.crt" -noout -text 2>/dev/null | grep -q "host.docker.internal"; then
        warn "Server 证书缺少 host.docker.internal SAN，将重新生成"
        need_regen=true
    fi

    if [[ "$need_regen" == "true" ]]; then
        info "生成 SSL 证书 (gRPC + mTLS)..."
        "$SCRIPT_DIR/generate-dev-certs.sh" --force > /dev/null 2>&1
        reset_runner_mtls_certs
        success "SSL 证书生成完成"
        return 0
    fi

    info "SSL 证书已存在"
}

generate_access_token_keys() {
    local key_dir="$SCRIPT_DIR/runtime/access-token"
    local private_key="$key_dir/private.pem"
    local public_key="$key_dir/public.pem"
    mkdir -p "$key_dir"

    if [[ -f "$private_key" && -f "$public_key" ]] &&
        openssl pkey -in "$private_key" -noout >/dev/null 2>&1 &&
        openssl pkey -pubin -in "$public_key" -noout >/dev/null 2>&1; then
        info "Access Token RSA 密钥已存在"
        return 0
    fi

    info "生成 Access Token RSA 密钥..."
    openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out "$private_key" >/dev/null 2>&1
    openssl pkey -in "$private_key" -pubout -out "$public_key" >/dev/null 2>&1
    chmod 600 "$private_key"
    chmod 644 "$public_key"
    success "Access Token RSA 密钥生成完成"
}

# Drop stale runner client certs after CA/server cert rotation. Runners
# regenerate on next container start (runner-entrypoint.sh).
reset_runner_mtls_certs() {
    local cleared=0
    while IFS= read -r container; do
        [[ -z "$container" ]] && continue
        docker exec "$container" sh -c 'rm -rf "$HOME/.agent-cloud/certs"/* 2>/dev/null || true' && cleared=$((cleared + 1)) || true
    done < <(docker ps --filter "name=${COMPOSE_PROJECT_NAME:-agentcloud-main}-runner" --format '{{.Names}}' 2>/dev/null || true)
    if (( cleared > 0 )); then
        info "已清除 $cleared 个 runner 容器的旧 mTLS 证书"
        local runner_ids
        runner_ids=$(docker ps --filter "name=${COMPOSE_PROJECT_NAME:-agentcloud-main}-runner" -q 2>/dev/null || true)
        if [[ -n "$runner_ids" ]]; then
            echo "$runner_ids" | xargs docker restart >/dev/null 2>&1 || true
        fi
    fi
}

generate_ai_cli_configs() {
    local config_dir="$SCRIPT_DIR/ai-cli-configs"

    mkdir -p "$config_dir/claude"
    cat > "$config_dir/claude/settings.json" << 'EOF'
{
  "permissions": {
    "allow": [
      "Bash",
      "Read",
      "Write",
      "Edit",
      "Glob",
      "Grep"
    ],
    "deny": []
  },
  "api_key_helper": null
}
EOF

    # Codex: headless mode for runner pods (Rust CLI >= 0.100.0).
    # When the dev machine has ~/.codex/config.toml with a custom OpenAI
    # provider (proxy base_url), merge it so pods inherit the same routing.
    mkdir -p "$config_dir/codex"
    local host_codex="${HOME}/.codex/config.toml"
    if [[ -f "$host_codex" ]] && grep -q '^\[model_providers\.OpenAI\]' "$host_codex" 2>/dev/null; then
        {
            echo 'approval_policy = "never"'
            echo 'sandbox_mode = "danger-full-access"'
            awk '
                /^model = / || /^model_provider = / { print; next }
                /^\[model_providers\.OpenAI\]/ { show=1 }
                show { print }
                show && /^\[/ && $0 !~ /^\[model_providers\.OpenAI\]/ { exit }
            ' "$host_codex"
        } > "$config_dir/codex/config.toml"
    else
    cat > "$config_dir/codex/config.toml" << 'EOF'
# Codex CLI configuration for headless mode
# Reference: https://developers.openai.com/codex/config-reference/

# Approval policy: "on-request", "never", "untrusted"
approval_policy = "never"

# Sandbox mode: "read-only", "workspace-write", "danger-full-access"
sandbox_mode = "danger-full-access"
EOF
    fi

    mkdir -p "$config_dir/gemini"
    cat > "$config_dir/gemini/settings.json" << 'EOF'
{
  "sandboxMode": "none",
  "autoApprove": true
}
EOF

    success "生成 AI CLI 配置"
}

# Traefik: file-provider only (post host-side migration there are no
# routable services left in the docker network — backend/relay run on
# the host, runner is the only docker target and isn't proxied).
generate_traefik_config() {
    local worktree_name="${WORKTREE_NAME:-main}"
    local project_name="agentcloud-${worktree_name}"
    local traefik_yml="$SCRIPT_DIR/traefik/traefik.yml"
    local dynamic_dir="$SCRIPT_DIR/traefik/dynamic"
    mkdir -p "$dynamic_dir"

    cat > "$traefik_yml" << EOF
# Traefik Static Configuration for Development
# Auto-generated by dev.sh - project: $project_name
#
# Backend / marketplace / relay run on the developer host (not in docker).
# Traefik discovers them via the file provider only.

api:
  dashboard: true
  insecure: true

entryPoints:
  web:
    address: ":80"
    # Long-lived connections (runner tunnel WS, preview SSE/video streaming)
    # must not be killed by entrypoint timeouts; readTimeout/writeTimeout=0
    # disables them, idleTimeout is generous but still bounded.
    transport:
      respondingTimeouts:
        readTimeout: 0s
        writeTimeout: 0s
        idleTimeout: 3600s
  grpc:
    address: ":9443"

providers:
  file:
    directory: /etc/traefik/dynamic
    watch: true

log:
  level: DEBUG
  format: common

accessLog:
  format: common
EOF

    cat > "$dynamic_dir/http.yml" << EOF
# HTTP routing — Auto-generated by dev.sh
# host services listen on worktree-scoped ports; traefik proxies in
# from :HTTP_PORT and forwards via host.lan (host-gateway).

http:
  routers:
    backend-api:
      entryPoints:
        - web
      rule: '!Host(\`preview.localhost\`) && !HostRegexp(\`^[a-z0-9-]+\\.preview\\.localhost$\`) && (PathPrefix(\`/api\`) || PathPrefix(\`/health\`) || PathPrefix(\`/v1\`) || PathPrefix(\`/auth\`) || PathPrefix(\`/proto.\`) || PathPrefix(\`/.well-known\`))'
      service: backend-api
      priority: 100

    marketplace-api:
      entryPoints:
        - web
      rule: "!Host(\`preview.localhost\`) && PathPrefix(\`/api/marketplace\`)"
      service: marketplace-api
      priority: 110

    relay:
      entryPoints:
        - web
      rule: "!Host(\`preview.localhost\`) && PathPrefix(\`/relay\`)"
      service: relay
      middlewares:
        - relay-strip
      priority: 30

    # Gateway HTTP data plane (tunnel/preview). Prefix is intentionally NOT
    # stripped: the gateway's own mux routes on the full /preview/{podKey}/*
    # path. No buffering middleware is attached here so WebSocket Upgrade
    # (preview WS passthrough) and streamed SSE/video responses pass through
    # untouched.
    preview:
      entryPoints:
        - web
      rule: "Host(\`preview.localhost\`) && PathPrefix(\`/preview\`)"
      service: relay
      priority: 40

    # Runner outbound tunnel endpoint. backend's cfg.TunnelURL() /
    # tunnelURLFromRelay() both mint "{scheme}://{host}/runner/tunnel"
    # (a bare, top-level path — NOT nested under /relay), so it needs its
    # own router rather than riding the relay-strip rule above.
    runner-tunnel:
      entryPoints:
        - web
      rule: "!Host(\`preview.localhost\`) && PathPrefix(\`/runner/tunnel\`)"
      service: relay
      priority: 40

  middlewares:
    relay-strip:
      stripPrefix:
        prefixes:
          - "/relay"

  services:
    backend-api:
      loadBalancer:
        servers:
          - url: "http://host.lan:${BACKEND_HTTP_PORT}"

    marketplace-api:
      loadBalancer:
        servers:
          - url: "http://host.lan:${MARKETPLACE_HTTP_PORT}"

    relay:
      loadBalancer:
        servers:
          - url: "http://host.lan:${RELAY_HTTP_PORT}"
EOF

    cat > "$dynamic_dir/grpc.yml" << EOF
# gRPC mTLS passthrough — Auto-generated by dev.sh
# Backend handles mTLS verification; traefik just tunnels TCP+TLS.

tcp:
  routers:
    grpc-mtls:
      entryPoints:
        - grpc
      rule: "HostSNI(\`*\`)"
      service: backend-grpc
      tls:
        passthrough: true

  services:
    backend-grpc:
      loadBalancer:
        servers:
          - address: "host.lan:${BACKEND_GRPC_PORT}"
EOF
    success "生成 Traefik 配置 (backend=:${BACKEND_HTTP_PORT} marketplace=:${MARKETPLACE_HTTP_PORT} relay=:${RELAY_HTTP_PORT})"
}

# `.env` is the SSOT for ports across worktrees. Re-runs preserve existing
# allocations so external tools (registered runners, stored OAuth callback
# URLs) keep pointing at the same ports across `dev.sh` invocations.
# Slot layout: 0-14 docker-exposed, 15-17 host-side services.
generate_env() {
    local worktree_name=$(get_worktree_name)
    local project_name="agentcloud-${worktree_name}"

    if [[ -f "$ENV_FILE" ]] && grep -q "COMPOSE_PROJECT_NAME=$project_name" "$ENV_FILE"; then
        source "$ENV_FILE"
        WORKTREE_NAME="$worktree_name"
        PORT_OFFSET=$(( (HTTP_PORT - 10000) / 50 ))
        if ! grep -q "^COMPOSE_FILE=" "$ENV_FILE"; then
            echo "COMPOSE_FILE=docker-compose.yml:docker-compose.runners.yml" >> "$ENV_FILE"
            export COMPOSE_FILE="docker-compose.yml:docker-compose.runners.yml"
        fi
        if grep -q '^RUNNERS_LAUNCHER=k8s' "$ENV_FILE"; then
            sed -i.bak 's|^COMPOSE_FILE=.*|COMPOSE_FILE=docker-compose.yml|' "$ENV_FILE"
            rm -f "$ENV_FILE.bak"
            export COMPOSE_FILE=docker-compose.yml
            export RUNNERS_LAUNCHER=k8s
        elif grep -q '^RUNNERS_LAUNCHER=coordinator' "$ENV_FILE"; then
            sed -i.bak 's|^COMPOSE_FILE=.*|COMPOSE_FILE=docker-compose.yml|' "$ENV_FILE"
            rm -f "$ENV_FILE.bak"
            export COMPOSE_FILE=docker-compose.yml
            export RUNNERS_LAUNCHER=coordinator
        fi
        # Backfill WEB_ADMIN_PORT for .env files predating that field.
        if ! grep -q "WEB_ADMIN_PORT" "$ENV_FILE"; then
            local admin_port=$((10011 + PORT_OFFSET * 50))
            echo "WEB_ADMIN_PORT=$admin_port" >> "$ENV_FILE"
            export WEB_ADMIN_PORT="$admin_port"
            info "补充 WEB_ADMIN_PORT=$admin_port 到 .env"
        fi
        # Backfill host-side service ports for older .env files.
        if ! grep -q "BACKEND_HTTP_PORT" "$ENV_FILE"; then
            local backend_http=$((10015 + PORT_OFFSET * 50))
            local backend_grpc=$((10016 + PORT_OFFSET * 50))
            local relay_http=$((10017 + PORT_OFFSET * 50))
            {
                echo ""
                echo "# Host-side service ports (see lib/host_services.sh)"
                echo "BACKEND_HTTP_PORT=$backend_http"
                echo "BACKEND_GRPC_PORT=$backend_grpc"
                echo "RELAY_HTTP_PORT=$relay_http"
            } >> "$ENV_FILE"
            export BACKEND_HTTP_PORT="$backend_http"
            export BACKEND_GRPC_PORT="$backend_grpc"
            export RELAY_HTTP_PORT="$relay_http"
            info "补充 host service 端口: backend=$backend_http grpc=$backend_grpc relay=$relay_http"
        fi
        if ! grep -q "MARKETPLACE_HTTP_PORT" "$ENV_FILE"; then
            local marketplace_http=$((10022 + PORT_OFFSET * 50))
            echo "MARKETPLACE_HTTP_PORT=$marketplace_http" >> "$ENV_FILE"
            export MARKETPLACE_HTTP_PORT="$marketplace_http"
            info "补充 Marketplace 端口: $marketplace_http"
        fi
        # Backfill runner MCP port for legacy .env files generated before
        # tests/mcp-e2e/ existed.
        if ! grep -q "RUNNER_MCP_PORT" "$ENV_FILE"; then
            local runner_mcp=$((10018 + PORT_OFFSET * 50))
            local runner2_mcp=$((10019 + PORT_OFFSET * 50))
            {
                echo ""
                echo "# Runner MCP HTTP port (exposed from runner container for tests/mcp-e2e/)"
                echo "RUNNER_MCP_PORT=$runner_mcp"
                echo "RUNNER_2_MCP_PORT=$runner2_mcp"
            } >> "$ENV_FILE"
            export RUNNER_MCP_PORT="$runner_mcp"
            export RUNNER_2_MCP_PORT="$runner2_mcp"
            info "补充 runner MCP 端口: runner=$runner_mcp runner-2=$runner2_mcp"
        fi
        # Older .env files may have RUNNER_MCP_PORT but not RUNNER_2_MCP_PORT.
        if ! grep -q "RUNNER_2_MCP_PORT" "$ENV_FILE"; then
            local runner2_mcp=$((10019 + PORT_OFFSET * 50))
            echo "RUNNER_2_MCP_PORT=$runner2_mcp" >> "$ENV_FILE"
            export RUNNER_2_MCP_PORT="$runner2_mcp"
        fi
        if ! grep -q "WEB_USER_PORT" "$ENV_FILE"; then
            local web_user_port=$((10020 + PORT_OFFSET * 50))
            echo "WEB_USER_PORT=$web_user_port" >> "$ENV_FILE"
            export WEB_USER_PORT="$web_user_port"
            info "补充 WEB_USER_PORT=$web_user_port 到 .env"
        fi
        if ! grep -q "MOBILE_LOVABLE_PORT" "$ENV_FILE"; then
            local mobile_port=$((10021 + PORT_OFFSET * 50))
            echo "MOBILE_LOVABLE_PORT=$mobile_port" >> "$ENV_FILE"
            export MOBILE_LOVABLE_PORT="$mobile_port"
            info "补充 MOBILE_LOVABLE_PORT=$mobile_port 到 .env"
        fi
        if ! grep -q "PUBLIC_WEB_URL" "$ENV_FILE"; then
            echo "PUBLIC_WEB_URL=http://localhost:$WEB_PORT" >> "$ENV_FILE"
            export PUBLIC_WEB_URL="http://localhost:$WEB_PORT"
            info "补充 PUBLIC_WEB_URL=$PUBLIC_WEB_URL 到 .env"
        fi
        if ! grep -q "MOBILE_PUBLIC_BASE_URL" "$ENV_FILE"; then
            echo "MOBILE_PUBLIC_BASE_URL=http://localhost:$MOBILE_LOVABLE_PORT" >> "$ENV_FILE"
            export MOBILE_PUBLIC_BASE_URL="http://localhost:$MOBILE_LOVABLE_PORT"
            info "补充 MOBILE_PUBLIC_BASE_URL=$MOBILE_PUBLIC_BASE_URL 到 .env"
        fi
        if ! grep -q "PREVIEW_PUBLIC_ORIGIN" "$ENV_FILE"; then
            echo "PREVIEW_PUBLIC_ORIGIN=http://preview.localhost:$HTTP_PORT" >> "$ENV_FILE"
            export PREVIEW_PUBLIC_ORIGIN="http://preview.localhost:$HTTP_PORT"
            info "补充 PREVIEW_PUBLIC_ORIGIN=$PREVIEW_PUBLIC_ORIGIN 到 .env"
        fi
        if ! grep -q "PREVIEW_COOKIE_MODE" "$ENV_FILE"; then
            echo "PREVIEW_COOKIE_MODE=same-site" >> "$ENV_FILE"
            export PREVIEW_COOKIE_MODE="same-site"
            info "补充 PREVIEW_COOKIE_MODE=$PREVIEW_COOKIE_MODE 到 .env"
        fi
        if ! grep -q "^MCP_REGISTRY_ENABLED=" "$ENV_FILE"; then
            echo "MCP_REGISTRY_ENABLED=false" >> "$ENV_FILE"
            export MCP_REGISTRY_ENABLED=false
            info "开发环境已关闭 MCP Registry 全量同步"
        fi
        if ! grep -q "^KB_GITEA_REPOSITORY_BASE_URLS=" "$ENV_FILE"; then
            echo "KB_GITEA_REPOSITORY_BASE_URLS=http://gitea:3000" >> "$ENV_FILE"
            export KB_GITEA_REPOSITORY_BASE_URLS="http://gitea:3000"
            info "开发环境已声明内部 Gitea repository origin"
        fi
        success "保留现有端口配置 (worktree: $worktree_name, PRIMARY_DOMAIN: localhost:$HTTP_PORT)"
        return 0
    fi

    local offset
    offset=$(calculate_port_offset "$worktree_name")
    WORKTREE_NAME="$worktree_name"
    PORT_OFFSET="$offset"

    local http_port=$((10000 + offset * 50))
    local grpc_port=$((10001 + offset * 50))

    cat > "$ENV_FILE" << EOF
# Agent Cloud Dev Environment - Auto-generated
# Worktree: $worktree_name | Offset: $offset

COMPOSE_PROJECT_NAME=$project_name
COMPOSE_FILE=docker-compose.yml:docker-compose.runners.yml
RUNNERS_LAUNCHER=docker

# =============================================================================
# Unified Domain Configuration - Single Source of Truth
# All URLs (OAuth callbacks, webhooks, Relay, etc.) are derived from this
# =============================================================================
PRIMARY_DOMAIN=localhost:$http_port
PUBLIC_WEB_URL=http://localhost:$((10007 + offset * 50))
MOBILE_PUBLIC_BASE_URL=http://localhost:$((10021 + offset * 50))
PREVIEW_PUBLIC_ORIGIN=http://preview.localhost:$http_port
PREVIEW_COOKIE_MODE=same-site
USE_HTTPS=false

# =============================================================================
# Ports (步长 50，支持最多 500 个 worktree，端口范围 10000-35000)
# Slots 0-14: external (docker-exposed) ports
# Slots 15-17: host-side service ports (loopback only, behind traefik)
# Slot 20: host-side web-user Vite dev server
# =============================================================================
HTTP_PORT=$http_port
GRPC_PORT=$grpc_port
POSTGRES_PORT=$((10002 + offset * 50))
REDIS_PORT=$((10003 + offset * 50))
MINIO_API_PORT=$((10004 + offset * 50))
MINIO_CONSOLE_PORT=$((10005 + offset * 50))
ADMINER_PORT=$((10006 + offset * 50))
WEB_PORT=$((10007 + offset * 50))
TRAEFIK_DASHBOARD_PORT=$((10008 + offset * 50))
GITEA_HTTP_PORT=$((10009 + offset * 50))
GITEA_SSH_PORT=$((10010 + offset * 50))
WEB_ADMIN_PORT=$((10011 + offset * 50))
OTEL_GRPC_PORT=$((10012 + offset * 50))
OTEL_HTTP_PORT=$((10013 + offset * 50))
JAEGER_UI_PORT=$((10014 + offset * 50))

# Host-side service ports (see lib/host_services.sh)
BACKEND_HTTP_PORT=$((10015 + offset * 50))
BACKEND_GRPC_PORT=$((10016 + offset * 50))
RELAY_HTTP_PORT=$((10017 + offset * 50))
MARKETPLACE_HTTP_PORT=$((10022 + offset * 50))

# Runner MCP HTTP port — exposed from the runner docker container so
# tests/mcp-e2e/ (running on the host) can drive the agent-facing MCP
# JSON-RPC surface. Always 10018 within a worktree's port slot.
RUNNER_MCP_PORT=$((10018 + offset * 50))

# Second runner's MCP HTTP port — only the cross-runner pod_interaction
# spec needs direct host access here, but exposing it lets future specs
# target runner-2 without touching compose.
RUNNER_2_MCP_PORT=$((10019 + offset * 50))

# End-user workbench (clients/web-user, Vite — proxies /v1 to traefik)
WEB_USER_PORT=$((10020 + offset * 50))
MOBILE_LOVABLE_PORT=$((10021 + offset * 50))
KB_GITEA_REPOSITORY_BASE_URLS=http://gitea:3000

# =============================================================================
# Credentials
# =============================================================================
POSTGRES_PASSWORD=agentcloud_dev
JWT_SECRET=dev-jwt-secret-change-in-production
INTERNAL_API_SECRET=dev-internal-secret
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin

# =============================================================================
# OAuth (optional) - Only ClientID/Secret needed, RedirectURLs derived
# =============================================================================
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=

# =============================================================================
# AI CLI - Claude Code
# =============================================================================
ANTHROPIC_BASE_URL=https://api.anthropic.com
ANTHROPIC_AUTH_TOKEN=

# The registry imports thousands of marketplace entries and can starve local
# Runner heartbeats. Enable it explicitly when marketplace sync is under test.
MCP_REGISTRY_ENABLED=false
EOF
    success "生成 .env 配置 (worktree: $worktree_name, PRIMARY_DOMAIN: localhost:$http_port)"
}

generate_web_env() {
    local offset="${PORT_OFFSET:-0}"
    local worktree_name="${WORKTREE_NAME:-main}"
    local http_port="${HTTP_PORT:-$((10000 + offset * 50))}"
    local backend_http_port="${BACKEND_HTTP_PORT:-$((10015 + offset * 50))}"
    local web_env_file="$SCRIPT_DIR/../../clients/web/.env.local"
    local primary_domain="${PRIMARY_DOMAIN:-localhost:$http_port}"
    local websocket_protocol="ws"
    if [[ "${USE_HTTPS:-false}" == "true" ]]; then
        websocket_protocol="wss"
    fi

    cat > "$web_env_file" << EOF
# Auto-generated by dev.sh
# Worktree: $worktree_name | HTTP Port: $http_port
#
# Unified Domain Configuration - 所有 URL 从 PRIMARY_DOMAIN 派生
# 变量名与 Backend/Relay 完全统一

# =============================================================================
# Unified Domain Configuration - Single Source of Truth
# =============================================================================
PRIMARY_DOMAIN=$primary_domain
USE_HTTPS=${USE_HTTPS:-false}

# =============================================================================
# 本地开发特殊配置
# =============================================================================
# 前端 API 基础 URL（留空使用相对路径，由 Next.js rewrites 代理）
NEXT_PUBLIC_API_URL=

# WebSocket URL（浏览器直连 Traefik；用 localhost 而非 127.0.0.1，避免 macOS
# 上 netdisk 等应用独占 127.0.0.1:$http_port 导致 relay 连不上）
NEXT_PUBLIC_WS_URL=$websocket_protocol://$primary_domain

# Next.js rewrites 代理目标（仅服务端 SSR 使用，不暴露给浏览器）
# 直连 host-side backend，绕过 Traefik — 避免 127.0.0.1:$http_port 被其它进程占用
API_PROXY_TARGET=http://127.0.0.1:$backend_http_port

# OAuth (optional)
NEXT_PUBLIC_GITHUB_CLIENT_ID=
EOF
    success "生成 clients/web/.env.local (PRIMARY_DOMAIN: $primary_domain)"
}

generate_web_admin_env() {
    local offset="${PORT_OFFSET:-0}"
    local worktree_name="${WORKTREE_NAME:-main}"
    local http_port="${HTTP_PORT:-$((10000 + offset * 50))}"
    local web_admin_env_file="$SCRIPT_DIR/../../clients/web-admin/.env.local"

    cat > "$web_admin_env_file" << EOF
# Auto-generated by dev.sh
# Worktree: $worktree_name | HTTP Port: $http_port
#
# Unified Domain Configuration - Admin Console

# =============================================================================
# Unified Domain Configuration - Single Source of Truth
# =============================================================================
PRIMARY_DOMAIN=localhost:$http_port
USE_HTTPS=false
EOF
    success "生成 clients/web-admin/.env.local (PRIMARY_DOMAIN: localhost:$http_port)"
}

# Runner mTLS client cert under runtime/runner/certs/. The runner *container*
# has its own copy via runner-entrypoint.sh; this host-side variant exists
# only for tools that need to inspect the cert from outside the container.
# Not on the main() path anymore.
generate_runner_cert() {
    local rt_dir
    rt_dir="$(_runtime_dir)/runner"
    local certs_dir="$rt_dir/certs"
    local ssl_dir="$SCRIPT_DIR/ssl"
    local node_id="${RUNNER_NODE_ID:-dev-runner}"
    local org_slug="${RUNNER_ORG_SLUG:-dev-org}"

    mkdir -p "$certs_dir"

    if [[ -f "$certs_dir/runner.crt" && -f "$certs_dir/runner.key" ]]; then
        return 0
    fi
    if [[ ! -f "$ssl_dir/ca.crt" || ! -f "$ssl_dir/ca.key" ]]; then
        error "CA 证书未找到: $ssl_dir (先跑 generate_ssl_certs)"
        return 1
    fi

    info "生成 Runner mTLS 客户端证书..."
    openssl genrsa -out "$certs_dir/runner.key" 2048 2>/dev/null
    openssl req -new -key "$certs_dir/runner.key" \
        -out "$certs_dir/runner.csr" \
        -subj "/CN=${node_id}/O=${org_slug}/OU=Runner" 2>/dev/null
    cat > "$certs_dir/runner_ext.cnf" << 'EOF'
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
EOF
    openssl x509 -req -days 365 \
        -in "$certs_dir/runner.csr" \
        -CA "$ssl_dir/ca.crt" -CAkey "$ssl_dir/ca.key" \
        -CAserial "$certs_dir/ca.srl" -CAcreateserial \
        -out "$certs_dir/runner.crt" \
        -extfile "$certs_dir/runner_ext.cnf" 2>/dev/null
    cp "$ssl_dir/ca.crt" "$certs_dir/ca.crt"
    rm -f "$certs_dir/runner.csr" "$certs_dir/runner_ext.cnf"
    chmod 600 "$certs_dir/runner.key"
    chmod 644 "$certs_dir/runner.crt" "$certs_dir/ca.crt"
    success "Runner 证书生成完成 ($certs_dir)"
}

# Generate runner SSH keypair for Gitea repo access. Private key isn't
# committed; if a stale .pub doesn't match the .priv (e.g., one was
# regenerated alone), SSH errors with "identity_sign: private key contents
# do not match public" — detect via ssh-keygen -y and regen both halves.
generate_runner_ssh_key() {
    local ssh_dir="$SCRIPT_DIR/runner-ssh"
    local private_key="$ssh_dir/id_ed25519"
    local public_key="$ssh_dir/id_ed25519.pub"

    if [[ -f "$private_key" ]]; then
        local needs_regen=false

        if [[ ! -f "$public_key" ]]; then
            warn "Public key missing, will regenerate SSH key pair"
            needs_regen=true
        else
            local derived_key stored_key
            derived_key=$(ssh-keygen -y -f "$private_key" 2>/dev/null | awk '{print $1, $2}')
            stored_key=$(awk '{print $1, $2}' "$public_key" 2>/dev/null)
            if [[ -z "$derived_key" || "$derived_key" != "$stored_key" ]]; then
                warn "SSH key pair mismatch detected, will regenerate"
                needs_regen=true
            fi
        fi

        if [[ "$needs_regen" == true ]]; then
            rm -f "$private_key" "$public_key"
        else
            chmod 600 "$private_key"
            info "Runner SSH key already exists and is valid"
            return 0
        fi
    fi

    info "Generating runner SSH key (private key not committed)..."
    ssh-keygen -t ed25519 -C "agentcloud-dev-runner@local" -f "$private_key" -N "" > /dev/null
    chmod 600 "$private_key"
    success "Runner SSH key generated: $private_key"
}
