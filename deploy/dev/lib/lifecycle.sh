# shellcheck shell=bash
# Lifecycle helpers for deploy/dev.
# Start / stop / wait / health-check for host services + docker infra.

# shellcheck source=runners_k8s.sh
source "$(dirname "${BASH_SOURCE[0]}")/runners_k8s.sh"
# shellcheck source=coordinator_runners.sh
source "$(dirname "${BASH_SOURCE[0]}")/coordinator_runners.sh"
# shellcheck source=lifecycle_wasm.sh
source "$(dirname "${BASH_SOURCE[0]}")/lifecycle_wasm.sh"

# Banner / usage / docker-compose-up are factored out of the original main()
# so the entry point is just orchestration.

print_banner() {
    echo ""
    echo "=========================================="
    echo "  Agent Cloud 开发环境初始化"
    echo "=========================================="
    echo ""
}

print_usage() {
    cat << 'EOF'
用法:
  ./dev.sh                                  # 一键启动完整开发环境 (air + plain next)
  ./dev.sh --backend-only                   # 仅启动 docker + host backend/relay (CI)
  ./dev.sh --coordinator-runners            # 平台托管：Coordinator 按需起 runner
  ./dev.sh --runners-k8s                    # runner 部署到本地 K8s 集群
  ./dev.sh --rebuild-runner                 # 重 build runner binary + 重启容器
  ./dev.sh --reset-runners                  # 重启 host runner+relay (backend 不动)
  ./dev.sh --clean                          # 停止并清理所有服务
  ./dev.sh --lite                           # 仅 web 主前端 + Coordinator runners
  ./dev-lite.sh                             # 同 --lite

  改动 backend / relay 源码: air 自动重编译
  仅重启前端:               ./dev.sh --frontends
  改动 runner 源码:         ./dev.sh --rebuild-runner
  会话兼容验收 (dev 栈已起): bash deploy/dev/session_compat_smoke.sh

前端日志: tail -f deploy/dev/web.log
web-user 日志: tail -f deploy/dev/web-user.log
EOF
}

# `docker compose up -d --build` with a 3-attempt retry loop. The build
# context is small but the npm registry / Docker Hub fetch is flaky on
# fresh CI runners, so retries beat hard-fail every time.
docker_compose_up() {
    if coordinator_runners_enabled; then
        info "启动 Docker 基础设施 (Runner 由 Coordinator 按需创建)..."
    elif runners_k8s_enabled; then
        info "启动 Docker 基础设施 (runner 由 K8s 集群托管)..."
    elif [[ "${DEV_E2E_RUNNERS_ONLY:-}" == "1" ]]; then
        info "启动 Docker 基础设施 + e2e-echo runners only..."
    else
        info "启动 Docker 基础设施 + runner (首次可能需要几分钟)..."
    fi
    local up_attempt=0
    local up_max=3
    # Playwright CI only needs runners that ship e2e-mock-agent. Starting
    # claude/gemini/… images makes ListAvailableRunners return them first
    # and CreatePod then fails with "e2e-mock-agent: not found in $PATH".
    local -a up_services=()
    if [[ "${DEV_E2E_RUNNERS_ONLY:-}" == "1" ]] \
        && ! coordinator_runners_enabled \
        && ! runners_k8s_enabled; then
        up_services=(runner-e2e-echo runner-e2e-echo-2 runner-admin-workspace)
    elif [[ -n "${DEV_LOCAL_WORKER_RUNTIME_SERVICES:-}" ]] \
        && ! coordinator_runners_enabled \
        && ! runners_k8s_enabled; then
        read -r -a up_services <<<"$DEV_LOCAL_WORKER_RUNTIME_SERVICES"
    fi
    while [ $up_attempt -lt $up_max ]; do
        up_attempt=$((up_attempt + 1))
        # set -o pipefail so docker compose's non-zero exit (auth.docker.io
        # token timeouts, build failures) actually fails the pipe — without
        # it grep returns 0 even if compose crashed and the loop exits
        # success'fully' while postgres is missing.
        if [[ ${#up_services[@]} -gt 0 ]]; then
            # Base stack first (no runners file), then only e2e-echo services.
            if (set -o pipefail; COMPOSE_FILE=docker-compose.yml \
                docker compose up -d --build --quiet-pull 2>&1 \
                | grep -v "^#" | grep -v "^\[" | grep -v "^$") \
                && (set -o pipefail; docker compose up -d --build --quiet-pull \
                    "${up_services[@]}" 2>&1 \
                | grep -v "^#" | grep -v "^\[" | grep -v "^$"); then
                break
            fi
        elif (set -o pipefail; docker compose up -d --build --quiet-pull 2>&1 | grep -v "^#" | grep -v "^\[" | grep -v "^$"); then
            break
        fi
        if [ $up_attempt -eq $up_max ]; then
            error "Docker compose up failed after $up_max attempts"
            exit 1
        fi
        warn "compose up failed (attempt $up_attempt/$up_max) — retrying in 10s"
        sleep 10
    done
    success "Docker 基础设施已启动"
}

wait_for_postgres() {
    local pg_container="${COMPOSE_PROJECT_NAME}-postgres-1"
    info "等待 PostgreSQL 就绪..."
    if ! wait_for_service "$pg_container" "pg_isready -U agentcloud"; then
        error "PostgreSQL 启动超时"
        exit 1
    fi
    success "PostgreSQL 已就绪"
}

runner_compose_services() {
    echo runner-e2e-echo runner-e2e-echo-2 runner-claude-code runner-codex-cli runner-codex-cli-2 runner-cursor-cli runner-gemini-cli runner-loopal runner-do-agent runner-grok-build runner-minimax-cli runner-openclaw runner-hermes runner-aider runner-opencode runner-admin-workspace runner-admin-workspace-do-agent
}

# Kill stale runner CLI processes (in case anyone installed agentcloud-runner
# from `cargo install` or similar), rebuild the binary, then restart every
# agent-specific runner service so each image picks up runner-binary.
reset_runners() {
    if [[ -f "$ENV_FILE" ]]; then
        source "$ENV_FILE"
    fi

    echo ""
    echo "=========================================="
    echo "  Reset Runner (rebuild binary + restart)"
    echo "=========================================="
    echo ""

    if pgrep -f "agentcloud-runner" &>/dev/null; then
        info "停止本地 agentcloud-runner 进程..."
        pkill -f "agentcloud-runner" 2>/dev/null || true
        sleep 1
        pkill -9 -f "agentcloud-runner" 2>/dev/null || true
    fi

    build_runner_binary || return 1
    build_mock_agent_binary || return 1
    build_do_agent_binary || return 1

    if runners_k8s_enabled; then
        hot_swap_runner_k8s_binary || return 1
        success "K8s runner Pod 已热更新"
        echo ""
        return 0
    fi

    cd "$SCRIPT_DIR"
    # docker cp hot-swap instead of `up -d --build`: the image rebuild path
    # re-runs apt in runner.Dockerfile, which hangs behind Docker Desktop's
    # broken proxy on some hosts. Containers that were never created are
    # skipped — first-time creation is owned by ./dev.sh.
    local svc container updated=0
    for svc in $(runner_compose_services); do
        container="$(docker compose ps -aq "$svc" 2>/dev/null | head -1)"
        if [[ -z "$container" ]]; then
            info "跳过 ${svc} (容器不存在 — 先运行 ./dev.sh 创建)"
            continue
        fi
        docker cp "$SCRIPT_DIR/runner-binary" "${container}:/usr/local/bin/agent-cloud-runner"
        docker exec "$container" ln -sf agent-cloud-runner /usr/local/bin/agentcloud-runner 2>/dev/null || true
        case "$svc" in
            runner-e2e-echo*)
                docker cp "$SCRIPT_DIR/e2e-mock-agent-binary" "${container}:/usr/local/bin/e2e-mock-agent"
                ;;
        esac
        docker restart "$container" >/dev/null
        updated=$((updated + 1))
        info "已热更新 $svc"
    done
    if [[ "$updated" -eq 0 ]]; then
        error "没有可更新的 runner 容器 — 先运行 ./dev.sh"
        return 1
    fi
    success "Runner 容器已重启 (binary via docker cp，跳过 apt rebuild)"

    echo ""
}

# Tear down everything dev.sh creates: host service pids, frontend port
# squatters, docker volumes, .env. Safe to re-run.
clean() {
    if [[ -f "$ENV_FILE" ]]; then
        source "$ENV_FILE"
    fi
    local web_port="${WEB_PORT:-3000}"
    local web_admin_port="${WEB_ADMIN_PORT:-3001}"
    local web_user_port="${WEB_USER_PORT:-10020}"

    info "停止 host-side 服务 (air)..."
    stop_host_services
    success "host-side 服务已停止"

    stop_web_user

    _stop_setsid web
    _stop_setsid web-admin
    _stop_setsid web-user

    if lsof -i :"$web_port" &>/dev/null; then
        info "停止前端服务 (端口: $web_port)..."
        lsof -ti :"$web_port" | xargs kill -9 2>/dev/null || true
        success "前端服务已停止"
    fi

    if lsof -i :"$web_admin_port" &>/dev/null; then
        info "停止 Admin Console (端口: $web_admin_port)..."
        lsof -ti :"$web_admin_port" | xargs kill -9 2>/dev/null || true
        success "Admin Console 已停止"
    fi

    rm -f "$SCRIPT_DIR/web.log"
    rm -f "$SCRIPT_DIR/web-admin.log"
    rm -rf "$(_runtime_dir)"

    teardown_runners_k8s

    if [[ -f "$ENV_FILE" ]]; then
        info "清理 Docker 环境: ${COMPOSE_PROJECT_NAME:-agentcloud}..."
        cd "$SCRIPT_DIR"
        docker compose down -v --remove-orphans 2>/dev/null || true
        rm -f "$ENV_FILE"
        success "清理完成"
    else
        warn "Docker 环境未初始化"
    fi
}

show_result() {
    source "$ENV_FILE"

    echo ""
    echo "=========================================="
    echo "  Agent Cloud 开发环境已就绪!"
    echo "=========================================="
    echo ""
    echo "  前端:       http://localhost:$WEB_PORT"
    echo "  Admin:      http://localhost:$WEB_ADMIN_PORT"
    echo "  web-user:   http://localhost:${WEB_USER_PORT:-10020}"
    echo "  API:        http://localhost:$HTTP_PORT/api  (→ host backend :$BACKEND_HTTP_PORT)"
    echo "  Marketplace:http://localhost:$HTTP_PORT/api/marketplace  (→ host :$MARKETPLACE_HTTP_PORT)"
    echo "  Relay:      ws://localhost:$HTTP_PORT/relay  (→ host relay :$RELAY_HTTP_PORT)"
    echo "  gRPC mTLS:  grpcs://localhost:$GRPC_PORT      (→ host backend :$BACKEND_GRPC_PORT)"
    echo ""
    echo "  Host services (air hot-reload):"
    echo "    backend  日志: tail -f deploy/dev/runtime/backend/backend.log"
    echo "    market   日志: tail -f deploy/dev/runtime/marketplace/marketplace.log"
    echo "    relay    日志: tail -f deploy/dev/runtime/relay/relay.log"
    echo ""
    if runners_k8s_enabled; then
        echo "  K8s runners (namespace agentcloud):"
        echo "    状态: kubectl get pods -n agentcloud"
        echo "    日志: kubectl logs -n agentcloud -l app=runner-e2e-echo -f"
        echo "    MCP:  localhost:${RUNNER_MCP_PORT:-10018} (dev-runner)"
    elif coordinator_runners_enabled; then
        echo "  Coordinator 平台托管 Runner (按需 docker compose up):"
        echo "    日志: tail -f deploy/dev/runtime/backend/backend.log | grep -i runner"
        echo "    容器: docker compose ps | grep runner"
    else
        echo "  Docker runners (agent-specific images, no hot reload):"
        echo "    日志: docker compose logs -f $(runner_compose_services)"
    fi
    echo "    重 build: ./dev.sh --rebuild-runner"
    echo ""
    echo "  测试账号:   dev@agentcloud.local / AdminAb123456"
    echo "  管理员:     admin@agentcloud.local / Ab123456"
    echo ""
    echo "  其他服务:"
    echo "    Gitea:    http://localhost:$GITEA_HTTP_PORT (gitea-admin / gitea-admin-123)"
    echo "    Traefik:  http://localhost:$TRAEFIK_DASHBOARD_PORT (Dashboard)"
    echo "    Adminer:  http://localhost:$ADMINER_PORT"
    echo "    MinIO:    http://localhost:$MINIO_CONSOLE_PORT"
    echo "    Jaeger:   http://localhost:$JAEGER_UI_PORT (Tracing UI)"
    echo ""
    echo "  停止: ./dev.sh --clean"
    echo "  仅重 build runner: ./dev.sh --rebuild-runner"
    echo ""
}

# Reusable lockfile-driven pnpm install: skips if node_modules is in sync
# with pnpm-lock.yaml (md5 fingerprint), reinstalls otherwise. Returns
# non-zero on install failure so callers can decide fail-vs-skip.
_install_root_deps_if_needed() {
    local context="$1"            # human label for logs ("前端依赖" / "Admin Console 依赖")
    local stale_cache_dir="$2"    # .next/cache to wipe on reinstall
    local root_dir="$SCRIPT_DIR/../.."
    local lockfile="$root_dir/pnpm-lock.yaml"
    local lockfile_hash_file="$root_dir/node_modules/.pnpm-lock-hash"
    local current_hash="" cached_hash=""
    [[ -f "$lockfile" ]] && current_hash=$(md5 -q "$lockfile" 2>/dev/null || md5sum "$lockfile" | cut -d' ' -f1)
    [[ -f "$lockfile_hash_file" ]] && cached_hash=$(cat "$lockfile_hash_file")

    if [[ -d "$root_dir/node_modules" && "$current_hash" == "$cached_hash" ]]; then
        return 0
    fi

    info "安装 ${context}（根 workspace）..."
    if ! (cd "$root_dir" && pnpm install --frozen-lockfile); then
        error "${context} 安装失败"
        return 1
    fi
    echo "$current_hash" > "$lockfile_hash_file"
    rm -rf "$stale_cache_dir"
    success "${context} 安装完成"
}

# Common pre-flight for both Next.js dev servers: clear stale lockfile +
# port squatters. Returns 1 if the port is held by something we can't
# safely kick (i.e., not our own stale Next.js process).
_prepare_next_port() {
    local label="$1"      # "前端" / "Admin Console"
    local web_dir="$2"    # absolute path to clients/web or web-admin
    local web_port="$3"
    local stale_lock=false

    local lock_file="$web_dir/.next/dev/lock"
    if [[ -f "$lock_file" ]]; then
        warn "检测到残留的 ${label}锁文件，清理中..."
        # Only kill `next dev` process for the web frontend — admin keeps
        # using the lsof fallback because both frontends share the same
        # `next dev` process name and we don't want one cleanup to kill
        # the other.
        if [[ "$label" == "前端" ]]; then
            pkill -f "next dev" 2>/dev/null || true
        fi
        lsof -ti :"$web_port" 2>/dev/null | xargs kill -9 2>/dev/null || true
        sleep 1
        rm -f "$lock_file"
        rm -rf "$web_dir/.next/cache"
        success "${label}锁文件和缓存已清理"
        stale_lock=true
    fi

    if [[ "$stale_lock" == false ]] && lsof -i :"$web_port" &>/dev/null; then
        if _frontend_port_up "$web_port" && [[ "${DEV_FORCE_FRONTEND:-}" != "1" ]] && [[ "$label" != "web-user" ]]; then
            info "${label} 已在端口 $web_port 正常运行，跳过启动"
            return 1
        fi
        if [[ "$label" == "web-user" ]] || [[ "${DEV_FORCE_FRONTEND:-}" == "1" ]] || ! _frontend_port_up "$web_port"; then
            if ! _frontend_port_up "$web_port"; then
                warn "${label} 端口 $web_port 响应异常 (非 2xx/3xx)，清理并重启..."
            else
                warn "重启 ${label}：释放端口 $web_port..."
            fi
            if [[ "$label" == "前端" ]]; then
                _stop_setsid web
            elif [[ "$label" == "Admin Console" ]]; then
                _stop_setsid web-admin
            fi
            lsof -ti :"$web_port" 2>/dev/null | xargs kill -9 2>/dev/null || true
            sleep 1
            rm -f "$web_dir/.next/dev/lock"
            rm -rf "$web_dir/.next/cache"
            return 0
        fi
        warn "端口 $web_port 已被占用，跳过${label}启动"
        return 1
    fi
    return 0
}


# Launch the Next.js web frontend via plain `next dev` (pnpm workspace).
start_frontend() {
    source "$ENV_FILE"
    local web_dir="$SCRIPT_DIR/../../clients/web"
    local web_port="${WEB_PORT:-3000}"
    local root_dir="$SCRIPT_DIR/../.."

    _prepare_next_port "前端" "$web_dir" "$web_port" || {
        warn "主前端 (端口 $web_port) 未能启动 — 端口被占用"
        return 1
    }

    if ! command -v pnpm &>/dev/null; then
        error "未找到 pnpm，请先安装: npm install -g pnpm"
        return 1
    fi

    _install_root_deps_if_needed "前端依赖" "$web_dir/.next/cache" || return 1
    _ensure_agent_cloud_wasm || return 1

    local log_file="$SCRIPT_DIR/web.log"
    local public_web_host
    public_web_host=$(python3 -c \
        'import sys, urllib.parse; host = urllib.parse.urlparse(sys.argv[1]).hostname; sys.exit(1) if not host else print(host)' \
        "${PUBLIC_WEB_URL:-}") || {
        error "PUBLIC_WEB_URL 必须是包含 hostname 的绝对 URL"
        return 1
    }
    info "启动前端服务 (端口: $web_port, plain next)..."
    local saved_dir="$PWD"
    cd "$web_dir"
    # API_PROXY_TARGET drives next.config.ts rewrites: /api/* + /proto.* →
    # host backend (:BACKEND_HTTP_PORT). Bypass traefik so macOS apps that
    # squat 127.0.0.1:$HTTP_PORT (e.g. netdisk) can't break login/API proxy.
    #
    # NEXT_PUBLIC_E2E=true enables build-time conditional registration of
    # test-only UI surfaces (e.g. the e2e-echo credential form). Production
    # builds never see this flag, so the e2e form is tree-shaken out. See
    # clients/web/src/components/settings/envBundleCredentialForms/index.ts.
    API_PROXY_TARGET="http://127.0.0.1:${BACKEND_HTTP_PORT}" \
    ALLOWED_DEV_ORIGINS="$public_web_host" \
    NEXT_PUBLIC_E2E="true" \
        _launch_setsid web "$log_file" \
        node ../../node_modules/next/dist/bin/next dev --turbopack --port "$web_port"
    cd "$saved_dir"

    local max_wait=90
    for ((i=1; i<=max_wait; i++)); do
        if _frontend_port_up "$web_port"; then
            success "前端服务已启动 (http://localhost:$web_port)"
            return 0
        fi
        sleep 1
    done

    warn "前端服务启动中，请稍后访问 http://localhost:$web_port"
    echo "  查看日志: tail -f $log_file"
}


start_admin_frontend() {
    source "$ENV_FILE"
    local web_admin_dir="$SCRIPT_DIR/../../clients/web-admin"
    local web_admin_port="${WEB_ADMIN_PORT:-3001}"
    local root_dir="$SCRIPT_DIR/../.."

    _prepare_next_port "Admin Console" "$web_admin_dir" "$web_admin_port" || {
        warn "Admin Console (端口 $web_admin_port) 未能启动 — 端口被占用"
        return 1
    }

    if ! command -v pnpm &>/dev/null; then
        error "未找到 pnpm，无法启动 Admin Console"
        return 1
    fi

    _install_root_deps_if_needed "Admin Console 依赖" "$web_admin_dir/.next/cache" || return 1

    local log_file="$SCRIPT_DIR/web-admin.log"
    info "启动 Admin Console (端口: $web_admin_port, plain next)..."
    local saved_dir="$PWD"
    cd "$web_admin_dir"
    # web-admin's next.config rewrites use PRIMARY_DOMAIN to compute the
    # backend URL (its fallback is the prod-only localhost:10000, which
    # never matches a worktree). Pin it to traefik so /api/* proxies.
    PRIMARY_DOMAIN="localhost:$HTTP_PORT" \
        _launch_setsid web-admin "$log_file" \
        node ../../node_modules/next/dist/bin/next dev --turbopack --port "$web_admin_port"
    cd "$saved_dir"

    local max_wait=60
    for ((i=1; i<=max_wait; i++)); do
        if curl -s "http://localhost:$web_admin_port" &>/dev/null; then
            success "Admin Console 已启动 (http://localhost:$web_admin_port)"
            return 0
        fi
        sleep 1
    done

    warn "Admin Console 启动中，请稍后访问 http://localhost:$web_admin_port"
    echo "  查看日志: tail -f $log_file"
}

# shellcheck source=lifecycle_launch.sh
source "$SCRIPT_DIR/lib/lifecycle_launch.sh"
# shellcheck source=lifecycle_web_user.sh
source "$SCRIPT_DIR/lib/lifecycle_web_user.sh"
# shellcheck source=lifecycle_frontends.sh
source "$SCRIPT_DIR/lib/lifecycle_frontends.sh"
