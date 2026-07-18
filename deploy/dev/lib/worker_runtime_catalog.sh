#!/bin/bash
# shellcheck shell=bash

local_worker_runner_services() {
    local catalog="$1"
    local worker_type service
    local services=()
    while IFS= read -r worker_type; do
        case "$worker_type" in
            codex-cli) service="runner-codex-cli" ;;
            gemini-cli) service="runner-gemini-cli" ;;
            minimax-cli) service="runner-minimax-cli" ;;
            openclaw) service="runner-openclaw" ;;
            do-agent|seedance-expert) service="runner-do-agent" ;;
            e2e-echo) service="runner-e2e-echo" ;;
            *) continue ;;
        esac
        if [[ " ${services[*]-} " != *" $service "* ]]; then
            services+=("$service")
        fi
    done < <(jq -r '.images[].worker_type_slugs[]' "$catalog")
    printf '%s' "${services[*]}"
}

local_worker_bootstrap_services() {
    local services
    services="$(local_worker_runner_services "$1")"
    if [[ " $services " != *" runner-e2e-echo "* ]]; then
        services="${services:+$services }runner-e2e-echo"
    fi
    printf '%s' "$services"
}

prepare_local_worker_runtime_catalog() {
    local repo_root="$SCRIPT_DIR/../.."
    local output="$SCRIPT_DIR/runtime/worker-runtime-catalog.local.json"
    local status=0

    node "$repo_root/scripts/generate-local-worker-runtime-catalog.mjs" \
        --output "$output" \
        --runtime "codex-cli=${COMPOSE_PROJECT_NAME}-runner-codex-cli:latest" \
        --runtime "gemini-cli=${COMPOSE_PROJECT_NAME}-runner-gemini-cli:latest" \
        --runtime "minimax-cli=${COMPOSE_PROJECT_NAME}-runner-minimax-cli:latest" \
        --runtime "openclaw=${COMPOSE_PROJECT_NAME}-runner-openclaw:latest" \
        --runtime "do-agent=${COMPOSE_PROJECT_NAME}-runner-do-agent:latest" \
        --runtime "e2e-echo=${COMPOSE_PROJECT_NAME}-runner-e2e-echo:latest" || status=$?

    if [[ "$status" -eq 0 ]]; then
        export WORKER_RUNTIME_CATALOG_FILE="$output"
        export DEV_LOCAL_WORKER_RUNTIME_SERVICES
        DEV_LOCAL_WORKER_RUNTIME_SERVICES="$(
            local_worker_bootstrap_services "$output"
        )"
        return 0
    fi
    if [[ "$status" -eq 2 ]]; then
        unset WORKER_RUNTIME_CATALOG_FILE
        unset DEV_LOCAL_WORKER_RUNTIME_SERVICES
        warn "未发现已验证的本地 Worker 运行时；Worker 向导保持正式发布门禁"
        return 0
    fi
    return "$status"
}
