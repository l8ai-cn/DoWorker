# shellcheck shell=bash
# Host services launcher for deploy/dev.
# Starts air-managed backend / relay / runner plus plain next dev for web / web-admin.
#
# Backend / relay run on the developer host. Runner stays in docker; its
# binary is cross-compiled via go build into deploy/dev/runner-binary.

# shellcheck source=build_do_agent_binary.sh
source "$(dirname "${BASH_SOURCE[0]}")/build_do_agent_binary.sh"
# shellcheck source=host_services_lite.sh
source "$(dirname "${BASH_SOURCE[0]}")/host_services_lite.sh"

# Contract marker for deploy/dev/runner_runtime_contract_test.sh — keep in sync
# with host_services_lite.sh / coordinator_runners.sh.
# COORDINATOR_RUNNER_DOCKER_COMPOSE_SERVICES=claude-code=runner-claude-code,codex-cli=runner-codex-cli,cursor-cli=runner-cursor-cli,gemini-cli=runner-gemini-cli,e2e-echo=runner-e2e-echo,loopal=runner-loopal,do-agent=runner-do-agent,grok-build=runner-grok-build,minimax-cli=runner-minimax-cli,openclaw=runner-openclaw,hermes=runner-hermes,aider=runner-aider,opencode=runner-opencode

_wait_http() {
    local url="$1" name="$2" max="${3:-40}"
    for ((i=1; i<=max; i++)); do
        if curl -sf "$url" >/dev/null 2>&1; then return 0; fi
        sleep 1
    done
    error "$name 健康检查超时 ($url)"
    return 1
}

_reap_port() {
    local port="$1" label="${2:-process}"
    [[ -z "$port" ]] && return 0
    local pids
    pids=$(lsof -ti "tcp:${port}" -sTCP:LISTEN 2>/dev/null || true)
    if [[ -n "$pids" ]]; then
        info "Reaping stale ${label} on port ${port}: $(echo "$pids" | tr '\n' ' ')"
        # shellcheck disable=SC2086
        kill -9 $pids 2>/dev/null || true
        sleep 1
    fi
}

build_runner_binary() {
    build_runner_binary_go
}

build_mock_agent_binary() {
    build_mock_agent_binary_go
}

start_backend_host() {
    start_backend_host_lite
}

start_marketplace_host() {
    start_marketplace_host_lite
}

start_relay_host() {
    start_relay_host_lite
}

stop_host_services() {
    local rt_root
    rt_root="$(_runtime_dir)"
    [[ -d "$rt_root" ]] || return 0
    for svc in backend marketplace relay; do
        local pgid_file="$rt_root/$svc/$svc.pgid"
        local pid_file="$rt_root/$svc/$svc.pid"
        if [[ -f "$pgid_file" ]]; then
            local pgid
            pgid=$(cat "$pgid_file" 2>/dev/null || true)
            if [[ -n "$pgid" ]]; then
                info "停止 host $svc (pgid: $pgid)..."
                kill -TERM -- "-$pgid" 2>/dev/null || true
                sleep 1
                kill -KILL -- "-$pgid" 2>/dev/null || true
            fi
            rm -f "$pgid_file"
        fi
        rm -f "$pid_file" 2>/dev/null || true
    done
    stop_host_services_lite_extra
}
