# shellcheck shell=bash
# lifecycle_web_user.sh — Vite dev server for clients/web-user (end-user workbench).

start_web_user() {
    source "$ENV_FILE"
    local web_user_dir="$SCRIPT_DIR/../../clients/web-user"
    local web_user_port="${WEB_USER_PORT:-10020}"

    _prepare_next_port "web-user" "$web_user_dir" "$web_user_port" || return 0

    if ! command -v npm &>/dev/null; then
        warn "未找到 npm，跳过 web-user 启动"
        return 0
    fi

    if [[ ! -d "$web_user_dir/node_modules" ]]; then
        info "安装 web-user 依赖 (clients/web-user)..."
        if ! (cd "$web_user_dir" && npm install); then
            error "web-user 依赖安装失败"
            return 1
        fi
    fi

    local log_file="$SCRIPT_DIR/web-user.log"
    info "启动 web-user (端口: $web_user_port, Vite → traefik :$HTTP_PORT)..."
    (
        cd "$web_user_dir"
        DO_WORKER_API_URL="http://127.0.0.1:${HTTP_PORT}" \
        AGENTSMESH_API_URL="http://127.0.0.1:${HTTP_PORT}" \
            _launch_setsid web-user "$log_file" \
            npm run dev -- --port "$web_user_port" --host 127.0.0.1
    )

    local max_wait=60
    for ((i = 1; i <= max_wait; i++)); do
        if curl -s "http://127.0.0.1:$web_user_port" &>/dev/null; then
            success "web-user 已启动 (http://localhost:$web_user_port)"
            return 0
        fi
        sleep 1
    done

    warn "web-user 启动中，请稍后访问 http://localhost:$web_user_port"
    echo "  查看日志: tail -f $log_file"
}

stop_web_user() {
    _stop_setsid web-user
    local web_user_port="${WEB_USER_PORT:-10020}"
    if lsof -i :"$web_user_port" &>/dev/null; then
        info "停止 web-user (端口: $web_user_port)..."
        lsof -ti :"$web_user_port" | xargs kill -9 2>/dev/null || true
    fi
    rm -f "$SCRIPT_DIR/web-user.log" "$SCRIPT_DIR/web-user.pid"
}
