# shellcheck shell=bash
# host_services_lite.sh — Go services via air + go build (no ibazel / no Bazel daemon).
#
# Enabled when DEV_NO_BAZEL=1 or DEV_LITE=1 (see dev-lite.sh).

dev_lite_enabled() {
    [[ "${DEV_NO_BAZEL:-}" == "1" || "${DEV_LITE:-}" == "1" ]]
}

ensure_go_protos() {
    local repo_root="$SCRIPT_DIR/../.."
    bash "$repo_root/scripts/proto-gen-go.sh" || {
        error "Go proto 生成失败 — brew install protobuf && ./scripts/proto-gen-go.sh"
        return 1
    }
}

ensure_go_codegen() {
    local repo_root="$SCRIPT_DIR/../.."
    ensure_go_protos || return 1
    bash "$repo_root/scripts/sync-amesh-convert.sh" || {
        error "amesh convert 生成失败 — ./scripts/sync-amesh-convert.sh"
        return 1
    }
}

ensure_air() {
    local gobin
    gobin="$(go env GOPATH)/bin"
    if [[ -x "$gobin/air" ]]; then
        export PATH="$gobin:$PATH"
        return 0
    fi
    info "安装 air (Go 热重载)..."
    go install github.com/air-verse/air@v1.61.7 || {
        error "air 安装失败"
        return 1
    }
    export PATH="$gobin:$PATH"
}

# Background-launch air. Args: name toml_relative log_file
_launch_air() {
    local name="$1" toml="$2" log_file="$3"
    local rt_dir repo_root
    rt_dir="$(_runtime_dir)/$name"
    repo_root="$SCRIPT_DIR/../.."
    mkdir -p "$rt_dir"
    local pid_file="$rt_dir/$name.pid"
    local pgid_file="$rt_dir/$name.pgid"

    if [[ -f "$pgid_file" ]]; then
        local old_pgid
        old_pgid=$(cat "$pgid_file" 2>/dev/null || true)
        if [[ -n "$old_pgid" ]]; then
            kill -TERM -- "-$old_pgid" 2>/dev/null || true
            sleep 1
            kill -KILL -- "-$old_pgid" 2>/dev/null || true
        fi
        rm -f "$pgid_file"
    fi
    rm -f "$pid_file"

    ensure_air || return 1

    info "启动 host service (air): $name"
    (
        cd "$repo_root"
        python3 -c "import os, sys; os.setsid(); os.execvp(sys.argv[1], sys.argv[1:])" \
            air -c "$toml" > "$log_file" 2>&1 &
        echo $! > "$pid_file"
        echo $! > "$pgid_file"
    )
}

build_runner_binary_go() {
    local repo_root="$SCRIPT_DIR/../.."
    ensure_go_codegen || return 1
    info "Go cross-compile runner (linux/amd64)..."
    (
        cd "$repo_root"
        GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
            go build -o "$SCRIPT_DIR/runner-binary" ./runner/cmd/runner
    ) || {
        error "go build runner 失败"
        return 1
    }
    chmod +x "$SCRIPT_DIR/runner-binary"
    success "Runner binary 已编译 (go build → deploy/dev/runner-binary)"
}

build_mock_agent_binary_go() {
    local repo_root="$SCRIPT_DIR/../.."
    ensure_go_codegen || return 1
    info "Go cross-compile e2e-mock-agent (linux/amd64)..."
    (
        cd "$repo_root"
        GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
            go build -o "$SCRIPT_DIR/e2e-mock-agent-binary" \
            ./runner/internal/agents/mockagent/cmd/e2e-mock-agent
    ) || {
        error "go build e2e-mock-agent 失败"
        return 1
    }
    chmod +x "$SCRIPT_DIR/e2e-mock-agent-binary"
    success "e2e-mock-agent binary 已编译 (go build)"

    [ -f "$SCRIPT_DIR/loopal-binary" ] || {
        cp "$SCRIPT_DIR/e2e-mock-agent-binary" "$SCRIPT_DIR/loopal-binary"
        chmod +x "$SCRIPT_DIR/loopal-binary"
    }
}

start_backend_host_lite() {
    source "$ENV_FILE"
    local repo_root="$SCRIPT_DIR/../.."
    mkdir -p "$repo_root/backend/logs"
    mkdir -p "$(_runtime_dir)/backend"

    # Re-use the same env exports as start_backend_host (call parent setup via
    # extracting shared block would be ideal; for now duplicate is avoided by
    # calling start_backend_host's export section — we invoke the ibazel path's
    # env by sourcing ENV and duplicating only launch logic).
    export DB_HOST=localhost
    export DB_PORT="$POSTGRES_PORT"
    export DB_USER=agentsmesh
    export DB_PASSWORD="${POSTGRES_PASSWORD:-agentsmesh_dev}"
    export DB_NAME=agentsmesh
    export DB_SSLMODE=disable
    export REDIS_URL="redis://localhost:${REDIS_PORT}"
    export JWT_SECRET="${JWT_SECRET:-dev-jwt-secret-change-in-production}"
    export INTERNAL_API_SECRET="${INTERNAL_API_SECRET:-dev-internal-secret}"
    export SERVER_ADDRESS=":${BACKEND_HTTP_PORT}"
    export GRPC_ADDRESS=":${BACKEND_GRPC_PORT}"
    export GRPC_PUBLIC_ENDPOINT="grpcs://127.0.0.1:${GRPC_PORT}"
    export DEBUG=true
    export RATE_LIMIT_DISABLED=true
    export PRIMARY_DOMAIN="${PRIMARY_DOMAIN}"
    export USE_HTTPS="${USE_HTTPS:-false}"
    export BLOCKSTORE_WEBHOOK_ALLOW_HOSTS="host.docker.internal,host.lan,localhost"
    export CORS_ALLOWED_ORIGINS="http://localhost:${HTTP_PORT},http://127.0.0.1:${HTTP_PORT},http://localhost:${WEB_PORT},http://127.0.0.1:${WEB_PORT},http://localhost:${WEB_ADMIN_PORT},http://127.0.0.1:${WEB_ADMIN_PORT},http://localhost:${WEB_USER_PORT:-10020},http://127.0.0.1:${WEB_USER_PORT:-10020},http://localhost:${MOBILE_LOVABLE_PORT:-10021},http://127.0.0.1:${MOBILE_LOVABLE_PORT:-10021}"
    export LOG_LEVEL=debug
    export LOG_FORMAT=text
    export LOG_FILE="$repo_root/backend/logs/agentsmesh.log"
    export EMAIL_PROVIDER=console
    export STORAGE_ENDPOINT="localhost:${MINIO_API_PORT}"
    export STORAGE_PUBLIC_ENDPOINT="localhost:${MINIO_API_PORT}"
    export STORAGE_RUNNER_ENDPOINT="host.lan:${MINIO_API_PORT}"
    export STORAGE_REGION=us-east-1
    export STORAGE_BUCKET=agentsmesh
    export STORAGE_ACCESS_KEY="${MINIO_ROOT_USER:-minioadmin}"
    export STORAGE_SECRET_KEY="${MINIO_ROOT_PASSWORD:-minioadmin}"
    export STORAGE_USE_SSL=false
    export STORAGE_USE_PATH_STYLE=true
    export STORAGE_MAX_FILE_SIZE=10
    export STORAGE_ALLOWED_TYPES="image/jpeg,image/png,image/gif,image/webp,application/pdf"
    export DEPLOYMENT_TYPE="${DEPLOYMENT_TYPE:-global}"
    export PAYMENT_MOCK="${PAYMENT_MOCK:-false}"
    export PKI_CA_CERT_FILE="$SCRIPT_DIR/ssl/ca.crt"
    export PKI_CA_KEY_FILE="$SCRIPT_DIR/ssl/ca.key"
    export PKI_SERVER_CERT_FILE="$SCRIPT_DIR/ssl/server.crt"
    export PKI_SERVER_KEY_FILE="$SCRIPT_DIR/ssl/server.key"
    export PKI_VALIDITY_DAYS=365
    export GEO_MMDB_PATH="${GEO_MMDB_PATH:-}"
    export OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:${OTEL_GRPC_PORT}"
    export OTEL_SERVICE_NAME=do-worker-backend
    export OTEL_TRACES_SAMPLER_ARG=1.0
    export AGENTSMESH_INCLUDE_INTERNAL_AGENTS=true
    local kb_token_file="$SCRIPT_DIR/runtime/gitea/backend-token"
    if [[ -f "$kb_token_file" ]]; then
        export KB_GITEA_URL="http://localhost:${GITEA_HTTP_PORT}"
        export KB_GITEA_TOKEN="$(cat "$kb_token_file")"
        export KB_GITEA_CLONE_URL="http://host.lan:${GITEA_HTTP_PORT}"
    fi
    export COORDINATOR_RUNNER_LAUNCHER=docker
    export COORDINATOR_RUNNER_DOCKER_COMPOSE_DIR="$SCRIPT_DIR"
    export COORDINATOR_RUNNER_DOCKER_COMPOSE_FILES=docker-compose.yml,docker-compose.runners.yml
    export COORDINATOR_RUNNER_DOCKER_COMPOSE_SERVICES="claude-code=runner-claude-code,codex-cli=runner-codex-cli,gemini-cli=runner-gemini-cli,e2e-echo=runner-e2e-echo,loopal=runner-loopal,do-agent=runner-do-agent,aider=runner-aider,opencode=runner-opencode"
    if coordinator_runners_enabled; then
        export_coordinator_runner_env
    fi

    ensure_go_codegen || return 1
    info "预编译 backend (go build)..."
    (cd "$repo_root" && go build -o "$(_runtime_dir)/backend/air/main" ./backend/cmd/server) || {
        error "backend go build 失败"
        return 1
    }

    _reap_port "$BACKEND_HTTP_PORT" "backend HTTP"
    _reap_port "$BACKEND_GRPC_PORT" "backend gRPC"

    local log_file="$(_runtime_dir)/backend/backend.log"
    _launch_air backend "deploy/dev/air/backend.toml" "$log_file" || return 1

    if ! _wait_http "http://localhost:${BACKEND_HTTP_PORT}/health" backend 120; then
        error "Backend 启动失败，查看日志: $log_file"
        tail -80 "$log_file" >&2 || true
        return 1
    fi
    success "Backend 已就绪 (air, host :${BACKEND_HTTP_PORT}, gRPC :${BACKEND_GRPC_PORT})"
}

start_relay_host_lite() {
    source "$ENV_FILE"
    local repo_root="$SCRIPT_DIR/../.."
    export SERVER_HOST="0.0.0.0"
    export SERVER_PORT="${RELAY_HTTP_PORT}"
    export WS_READ_BUFFER_SIZE=4096
    export WS_WRITE_BUFFER_SIZE=4096
    export JWT_SECRET="${JWT_SECRET:-dev-jwt-secret-change-in-production}"
    export BACKEND_URL="http://localhost:${HTTP_PORT}"
    export INTERNAL_API_SECRET="${INTERNAL_API_SECRET:-dev-internal-secret}"
    export RELAY_ID=dev-relay-1
    export RELAY_REGION=local
    export RELAY_CAPACITY=1000
    export PRIMARY_DOMAIN="${PRIMARY_DOMAIN}"
    export USE_HTTPS="${USE_HTTPS:-false}"
    export ALLOWED_ORIGINS="http://localhost:${HTTP_PORT},http://127.0.0.1:${HTTP_PORT},http://localhost:${WEB_PORT},http://127.0.0.1:${WEB_PORT},http://localhost:${WEB_ADMIN_PORT},http://127.0.0.1:${WEB_ADMIN_PORT},http://localhost:${WEB_USER_PORT:-10020},http://127.0.0.1:${WEB_USER_PORT:-10020}"
    export SESSION_KEEP_ALIVE_DURATION=30s
    export DEBUG=true
    export OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:${OTEL_GRPC_PORT}"
    export OTEL_SERVICE_NAME=do-worker-relay
    export OTEL_TRACES_SAMPLER_ARG=1.0

    ensure_go_codegen || return 1
    info "预编译 relay (go build)..."
    (cd "$repo_root" && go build -o "$(_runtime_dir)/relay/air/main" ./relay/cmd/relay) || {
        error "relay go build 失败"
        return 1
    }

    _reap_port "$RELAY_HTTP_PORT" "relay"
    local log_file="$(_runtime_dir)/relay/relay.log"
    _launch_air relay "deploy/dev/air/relay.toml" "$log_file" || return 1

    if ! _wait_http "http://localhost:${RELAY_HTTP_PORT}/health" relay 60; then
        error "Relay 启动失败，查看日志: $log_file"
        tail -80 "$log_file" >&2 || true
        return 1
    fi
    success "Relay 已就绪 (air, host :${RELAY_HTTP_PORT})"
}

stop_host_services_lite_extra() {
    pkill -f "deploy/dev/runtime/backend/air/main" 2>/dev/null || true
    pkill -f "deploy/dev/runtime/relay/air/main" 2>/dev/null || true
    pkill -f "air -c deploy/dev/air/backend.toml" 2>/dev/null || true
    pkill -f "air -c deploy/dev/air/relay.toml" 2>/dev/null || true
}
