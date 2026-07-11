#!/bin/bash
# shellcheck shell=bash
# Platform-managed runner mode (Coordinator auto-provision, no dev pre-launch).

coordinator_runners_enabled() {
    [[ "${RUNNERS_LAUNCHER:-docker}" == "coordinator" ]]
}

export_coordinator_runner_env() {
    # shellcheck disable=SC1090
    source "$ENV_FILE"
    export COORDINATOR_RUNNER_LAUNCHER=docker
    export COORDINATOR_RUNNER_DOCKER_COMPOSE_DIR="$SCRIPT_DIR"
    export COORDINATOR_RUNNER_DOCKER_COMPOSE_FILES=docker-compose.yml,docker-compose.runners.yml
    export COORDINATOR_RUNNER_DOCKER_COMPOSE_SERVICES="claude-code=runner-claude-code,codex-cli=runner-codex-cli,gemini-cli=runner-gemini-cli,e2e-echo=runner-e2e-echo,loopal=runner-loopal,do-agent=runner-do-agent,grok-build=runner-grok-build,openclaw=runner-openclaw,hermes=runner-hermes,aider=runner-aider,opencode=runner-opencode"
    export COORDINATOR_RUNNER_BACKEND_URL=http://traefik:80
    export COORDINATOR_RUNNER_GRPC_ENDPOINT="host.lan:${BACKEND_GRPC_PORT}"
    export COORDINATOR_RUNNER_RELAY_BASE_URL=ws://traefik:80/relay
    export COORDINATOR_RUNNER_ORG_SLUG=dev-org
    export COORDINATOR_RUNNER_DOCKER_SSL_HOST_PATH="$SCRIPT_DIR/ssl"
    export COORDINATOR_RUNNER_DOCKER_ENTRYPOINT_HOST_PATH="$SCRIPT_DIR/runner-entrypoint.sh"
}

persist_runners_launcher_mode() {
    local mode="$1"
    [[ -f "$ENV_FILE" ]] || return 0

    if grep -q '^RUNNERS_LAUNCHER=' "$ENV_FILE" 2>/dev/null; then
        sed -i.bak "s/^RUNNERS_LAUNCHER=.*/RUNNERS_LAUNCHER=${mode}/" "$ENV_FILE"
        rm -f "$ENV_FILE.bak"
    else
        echo "RUNNERS_LAUNCHER=${mode}" >> "$ENV_FILE"
    fi
    export RUNNERS_LAUNCHER="$mode"

    case "$mode" in
        coordinator|k8s)
            if grep -q '^COMPOSE_FILE=' "$ENV_FILE"; then
                sed -i.bak 's|^COMPOSE_FILE=.*|COMPOSE_FILE=docker-compose.yml|' "$ENV_FILE"
                rm -f "$ENV_FILE.bak"
            fi
            export COMPOSE_FILE=docker-compose.yml
            ;;
        docker)
            if grep -q '^COMPOSE_FILE=' "$ENV_FILE"; then
                sed -i.bak 's|^COMPOSE_FILE=.*|COMPOSE_FILE=docker-compose.yml:docker-compose.runners.yml|' "$ENV_FILE"
                rm -f "$ENV_FILE.bak"
            fi
            export COMPOSE_FILE=docker-compose.yml:docker-compose.runners.yml
            ;;
    esac
}

stop_compose_runners() {
    # Coordinator mode: ensure no stale pre-launched runners shadow auto-provision.
    cd "$SCRIPT_DIR"
    export COMPOSE_FILE=docker-compose.yml:docker-compose.runners.yml
    docker compose stop \
        runner-e2e-echo runner-e2e-echo-2 runner-claude-code \
        runner-codex-cli runner-gemini-cli runner-loopal \
        runner-do-agent runner-grok-build runner-openclaw runner-hermes \
        runner-aider runner-opencode \
        runner-admin-workspace runner-admin-workspace-do-agent 2>/dev/null || true
}
