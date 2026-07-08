# shellcheck shell=bash
# lifecycle_web_user.sh — Vite dev server for clients/web-user (end-user workbench).

_web_user_dir() {
    echo "$SCRIPT_DIR/../../clients/web-user"
}

_web_user_install_deps() {
    local web_user_dir
    web_user_dir="$(_web_user_dir)"
    if [[ -d "$web_user_dir/node_modules" ]]; then
        return 0
    fi
    info "安装 web-user 依赖 (clients/web-user)..."
    if command -v npm &>/dev/null; then
        (cd "$web_user_dir" && npm install) && return 0
    fi
    if command -v pnpm &>/dev/null; then
        (cd "$web_user_dir" && pnpm install) && return 0
    fi
    error "未找到 pnpm 或 npm，无法安装 web-user 依赖"
    return 1
}

_web_user_prepare_port() {
    local web_user_port="${WEB_USER_PORT:-10020}"
    _prepare_next_port "web-user" "$(_web_user_dir)" "$web_user_port"
}

web_user_launch_background() {
    source "$ENV_FILE"
    local web_user_dir web_user_port log_file

    web_user_dir="$(_web_user_dir)"
    web_user_port="${WEB_USER_PORT:-10020}"
    log_file="$SCRIPT_DIR/web-user.log"

    if ! _web_user_prepare_port; then
        error "web-user 端口 $web_user_port 无法释放，跳过启动"
        return 1
    fi

    if ! _web_user_install_deps; then
        return 1
    fi

    info "启动 web-user (端口: $web_user_port, Vite → backend localhost:${BACKEND_HTTP_PORT})..."
    (
        cd "$web_user_dir"
        DO_WORKER_API_URL="http://127.0.0.1:${BACKEND_HTTP_PORT}" \
        AGENTSMESH_API_URL="http://127.0.0.1:${BACKEND_HTTP_PORT}" \
            _launch_setsid web-user "$log_file" \
            npm run dev -- --port "$web_user_port" --host 127.0.0.1
    )
}

web_user_wait_ready() {
    source "$ENV_FILE"
    local web_user_port="${WEB_USER_PORT:-10020}"
    local log_file="$SCRIPT_DIR/web-user.log"
    local max_wait=90

    for ((i = 1; i <= max_wait; i++)); do
        if _frontend_port_up "$web_user_port"; then
            success "web-user 已启动 (http://localhost:$web_user_port)"
            return 0
        fi
        sleep 1
    done

    warn "web-user 仍在启动中 — 查看日志: tail -f $log_file"
    return 1
}

start_web_user() {
    web_user_launch_background || return 1
    web_user_wait_ready
}

stop_web_user() {
    _stop_setsid web-user
    local web_user_port="${WEB_USER_PORT:-10020}"
    if lsof -i :"$web_user_port" &>/dev/null; then
        info "停止 web-user (端口: $web_user_port)..."
        lsof -ti :"$web_user_port" | xargs kill -9 2>/dev/null || true
    fi
    rm -f "$SCRIPT_DIR/web-user.log"
}
