#!/bin/bash

set -e
trap 'echo "✗ entrypoint failed at line $LINENO (exit=$?)" >&2' ERR

BACKEND_URL="${BACKEND_URL:-http://traefik:80}"
GRPC_ENDPOINT="${GRPC_ENDPOINT:-traefik:9443}"
RELAY_BASE_URL="${RELAY_BASE_URL:-ws://traefik:80}"
RUNNER_NODE_ID="${RUNNER_NODE_ID:-dev-runner}"
RUNNER_ORG_SLUG="${RUNNER_ORG_SLUG:-dev-org}"
MAX_CONCURRENT_PODS="${MAX_CONCURRENT_PODS:-10}"
SSL_DIR="${SSL_DIR:-/app/ssl}"
AGENT_RUNTIME="${AGENT_RUNTIME:-e2e-echo}"
DEFAULT_AGENT="${DEFAULT_AGENT:-${AGENT_RUNTIME}}"
RUNNER_SSH_SOURCE_DIR="${RUNNER_SSH_SOURCE_DIR:-/run/runner-ssh-source}"
RUNNER_USER="${RUNNER_USER:-runner}"
RUNNER_GROUP="${RUNNER_GROUP:-runner}"
RUNNER_PRIVILEGED_BOOTSTRAP_DONE="${RUNNER_PRIVILEGED_BOOTSTRAP_DONE:-0}"

CONFIG_DIR="${HOME}/.do-worker"
if [[ -d "${HOME}/.agentsmesh" && -w "${HOME}/.agentsmesh" ]]; then
    CONFIG_DIR="${HOME}/.agentsmesh"
elif [[ -d "${HOME}/.agentsmesh" && ! -w "${HOME}/.agentsmesh" ]]; then
    echo "▶ ${HOME}/.agentsmesh not writable (uid $(id -u)); using ${CONFIG_DIR}" >&2
fi
CERTS_DIR="${CONFIG_DIR}/certs"
CONFIG_FILE="${CONFIG_DIR}/config.yaml"

source /usr/local/lib/runner-ssh-bootstrap.sh
run_privileged_bootstrap_then_drop "$@"

case "${AGENT_RUNTIME}" in
    claude-code|codex-cli|video-studio|cursor-cli|gemini-cli|minimax-cli|e2e-echo|loopal|do-agent|aider|opencode|grok-build|openclaw|hermes) ;;
    *)
        echo "✗ Unsupported AGENT_RUNTIME=${AGENT_RUNTIME}" >&2
        exit 1
        ;;
esac

echo "========================================"
echo "  Do Worker Runner Entrypoint"
echo "========================================"
echo "  Backend URL:    $BACKEND_URL"
echo "  gRPC Endpoint:  $GRPC_ENDPOINT"
echo "  Relay Base URL: $RELAY_BASE_URL"
echo "  Node ID:        $RUNNER_NODE_ID"
echo "  Org Slug:       $RUNNER_ORG_SLUG"
echo "  Agent Runtime:  $AGENT_RUNTIME"
echo "  Default Agent:  $DEFAULT_AGENT"
echo "  Max Pods:       $MAX_CONCURRENT_PODS"
echo ""

wait_for_backend() {
    echo "等待 Backend 服务就绪..."
    local health_url="${BACKEND_URL}/health"
    for ((i=1; i<=240; i++)); do
        if wget -q -O /dev/null "$health_url" 2>/dev/null; then
            echo "✓ Backend 服务就绪"
            return 0
        fi
        echo "  等待 Backend... ($i/240)"
        sleep 2
    done
    echo "✗ Backend 服务启动超时" >&2
    exit 1
}

create_config() {
    mkdir -p "$CONFIG_DIR"
    cat > "$CONFIG_FILE" << EOF
server_url: "${BACKEND_URL}"
grpc_endpoint: "${GRPC_ENDPOINT}"
cert_file: "${CERTS_DIR}/runner.crt"
key_file: "${CERTS_DIR}/runner.key"
ca_file: "${CERTS_DIR}/ca.crt"
relay_base_url: "${RELAY_BASE_URL}"
node_id: "${RUNNER_NODE_ID}"
description: "Development Docker Runner"
org_slug: "${RUNNER_ORG_SLUG}"
max_concurrent_pods: ${MAX_CONCURRENT_PODS}
workspace: "/workspace"
workspace_root: "/workspace/repos"
worktrees_dir: "/workspace/worktrees"
base_branch: "main"
default_agent: "${DEFAULT_AGENT}"
default_shell: "/bin/bash"
log_level: "debug"
EOF
}
main() {
    wait_for_backend
    if [[ "$RUNNER_PRIVILEGED_BOOTSTRAP_DONE" != "1" ]]; then
        bootstrap_runner_ssh
    fi
    generate_runner_cert
    init_ai_cli_configs
    create_config
    echo "启动 Runner..."
    exec /usr/local/bin/do-worker-runner run --config "$CONFIG_FILE"
}
main "$@"
