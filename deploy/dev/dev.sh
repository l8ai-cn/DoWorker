#!/bin/bash
# =============================================================================
# AgentsMesh dev environment — entry point.
# =============================================================================
#
# Functionality lives in lib/. This file is only:
#   - global path / env var setup
#   - lib loader (order matters: leaves before composites)
#   - arg parsing + main orchestration
#
# 一键启动开发环境：
#   ./dev.sh                # docker infra + host backend/relay + frontend
#   ./dev.sh --frontends      # 仅重启 web + web-admin + web-user
#   ./dev.sh --coordinator-runners # 平台托管：Coordinator 按需起 runner
#   ./dev.sh --runners-k8s       # runner 部署到本地 K8s 集群 (Docker Desktop)
#   ./dev.sh --rebuild-runner   # 重 build runner binary + 重启 runner 容器
#   ./dev.sh --clean        # 清理所有服务
#   ./dev.sh --help         # 帮助
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$SCRIPT_DIR/.env"
MIGRATIONS_DIR="$SCRIPT_DIR/../../backend/migrations"
SEED_FILE="$SCRIPT_DIR/seed/seed.sql"
LEMONSQUEEZY_SEED_FILE="$SCRIPT_DIR/seed/seed_lemonsqueezy.sql"
E2E_ECHO_SEED_FILE="$SCRIPT_DIR/seed/e2e_echo.sql"

# Source order: leaves (no deps) first, composites last.
# log → worktree/doctor → config_gen/host_services/bootstrap → lifecycle
# (lifecycle composes host_services + the rest).
# shellcheck source=lib/log.sh
source "$SCRIPT_DIR/lib/log.sh"
# shellcheck source=lib/worktree.sh
source "$SCRIPT_DIR/lib/worktree.sh"
# shellcheck source=lib/doctor.sh
source "$SCRIPT_DIR/lib/doctor.sh"
# shellcheck source=lib/config_gen.sh
source "$SCRIPT_DIR/lib/config_gen.sh"
# shellcheck source=lib/worker_runtime_catalog.sh
source "$SCRIPT_DIR/lib/worker_runtime_catalog.sh"
# shellcheck source=lib/coordinator_runners.sh
source "$SCRIPT_DIR/lib/coordinator_runners.sh"
# shellcheck source=lib/host_services.sh
source "$SCRIPT_DIR/lib/host_services.sh"
# shellcheck source=lib/bootstrap.sh
source "$SCRIPT_DIR/lib/bootstrap.sh"
# shellcheck source=lib/lifecycle.sh
source "$SCRIPT_DIR/lib/lifecycle.sh"

main() {
    cd "$SCRIPT_DIR"

    # Default: air + plain next (no Bazel / ibazel).
    export DEV_NO_BAZEL="${DEV_NO_BAZEL:-1}"

    case "${1:-}" in
        --clean|-c)
            clean
            exit 0
            ;;
        --reset-runners|--kill-runners|--rebuild-runner)
            reset_runners
            exit 0
            ;;
        --help|-h)
            print_usage
            exit 0
            ;;
        --frontends|-f)
            cd "$SCRIPT_DIR"
            [[ -f "$ENV_FILE" ]] || { error "缺少 $ENV_FILE，请先运行 ./dev.sh"; exit 1; }
            source "$ENV_FILE"
            generate_web_env
            generate_web_admin_env
            export DEV_FORCE_FRONTEND=1
            print_banner
            warn_loopback_port_conflict
            start_all_frontends
            show_result
            exit 0
            ;;
    esac

    local backend_only=false
    local requested_runner_launcher
    requested_runner_launcher="$(resolve_requested_runners_launcher "${RUNNERS_LAUNCHER:-}" "$@")"
    for arg in "$@"; do
        case "$arg" in
            --backend-only) backend_only=true ;;
            --lite) export DEV_LITE=1; export DEV_NO_BAZEL=1; export WEB_USER_SKIP=1 ;;
            --frontends|-f) ;;
            --coordinator-runners|--runners-k8s) ;;
        esac
    done

    print_banner

    # Phase 1: configs (deterministic, no docker yet).
    generate_ssl_certs
    generate_access_token_keys
    generate_ai_cli_configs
    generate_env
    source "$ENV_FILE"
    local lite_runner_mode=false
    if [[ "${DEV_LITE:-}" == "1" ]]; then
        lite_runner_mode=true
    fi
    local effective_runner_launcher
    effective_runner_launcher="$(
        resolve_effective_runners_launcher \
            "$requested_runner_launcher" "${RUNNERS_LAUNCHER:-}" "$lite_runner_mode"
    )"
    check_ibazel_doctor
    generate_traefik_config
    generate_web_env
    generate_web_admin_env
    generate_runner_ssh_key
    warn_loopback_port_conflict
    prepare_local_worker_runtime_catalog

    if [[ -n "$effective_runner_launcher" ]]; then
        persist_runners_launcher_mode "$effective_runner_launcher"
        case "$effective_runner_launcher" in
            coordinator)
                info "Runner 模式: Coordinator 平台托管 (不预起 runner 容器)"
                stop_compose_runners
                ;;
            k8s)
                info "Runner 模式: Kubernetes 集群 (跳过 docker-compose.runners.yml)"
                ;;
        esac
        source "$ENV_FILE"
    fi

    # Phase 2: bazel-build the runner binary so docker compose's runner
    build_runner_binary
    # Cross-compile the e2e-mock-agent alongside the runner — same build
    # context, same image. Required for mcp-e2e / envbundle-e2e / acp-ui-e2e
    # which depend on the `e2e-echo` AgentFile resolving `EXECUTABLE
    # e2e-mock-agent` to a real binary on the runner's PATH.
    build_mock_agent_binary
    build_do_agent_binary || return 1

    # Phase 3: docker infrastructure + DB bootstrap.
    docker_compose_up
    wait_for_postgres
    run_migrations
    run_marketplace_migrations
    init_seed "${COMPOSE_PROJECT_NAME}-postgres-1"
    sync_worker_definition_projections
    init_gitea
    setup_gitea_ssh_config

    # Phase 4: host services. backend must come up before runners can
    # complete their mTLS handshakes — runner containers connect via
    # traefik:9443, traefik passthroughs to host backend.
    start_backend_host
    start_marketplace_host
    start_relay_host

    if runners_k8s_enabled; then
        deploy_runners_k8s || warn "K8s runner 部署失败 — 检查 Docker Desktop Kubernetes 是否启用"
    fi

    # Phase 5: frontends (skipped in CI).
    if [[ "$backend_only" == "true" ]]; then
        info "--backend-only: skipping frontend startup"
    else
        start_all_frontends
    fi

    show_result
}

main "$@"
