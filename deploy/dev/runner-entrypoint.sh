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

CONFIG_DIR="${HOME}/.do-worker"
if [[ -d "${HOME}/.agentsmesh" && -w "${HOME}/.agentsmesh" ]]; then
    CONFIG_DIR="${HOME}/.agentsmesh"
elif [[ -d "${HOME}/.agentsmesh" && ! -w "${HOME}/.agentsmesh" ]]; then
    echo "▶ ${HOME}/.agentsmesh not writable (uid $(id -u)); using ${CONFIG_DIR}" >&2
fi
CERTS_DIR="${CONFIG_DIR}/certs"
CONFIG_FILE="${CONFIG_DIR}/config.yaml"

case "${AGENT_RUNTIME}" in
    claude-code|codex-cli|gemini-cli|e2e-echo|loopal|do-agent|aider|opencode) ;;
    *)
        echo "✗ Unsupported AGENT_RUNTIME=${AGENT_RUNTIME}" >&2
        exit 1
        ;;
esac

echo "========================================"
echo "  Do Worker Runner Entrypoint (Bazel)"
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

generate_runner_cert() {
    echo "▶ generate_runner_cert: CERTS_DIR=$CERTS_DIR SSL_DIR=$SSL_DIR"
    mkdir -p "$CERTS_DIR"
    ls -la "$CERTS_DIR" >&2 || true
    if [[ ! -f "$SSL_DIR/ca.crt" || ! -f "$SSL_DIR/ca.key" ]]; then
        echo "✗ CA 证书未找到: $SSL_DIR" >&2
        ls -la "$SSL_DIR" >&2 || true
        exit 1
    fi
    if [[ -s "$CERTS_DIR/runner.crt" && -s "$CERTS_DIR/runner.key" ]] \
        && openssl verify -CAfile "$SSL_DIR/ca.crt" "$CERTS_DIR/runner.crt" >/dev/null 2>&1; then
        cp "$SSL_DIR/ca.crt" "$CERTS_DIR/ca.crt"
        echo "✓ Runner 证书已存在"
        return 0
    fi
    rm -f "$CERTS_DIR/runner.crt" "$CERTS_DIR/runner.key" "$CERTS_DIR/ca.crt" \
          "$CERTS_DIR/ca.srl" "$CERTS_DIR/runner.csr" "$CERTS_DIR/runner_ext.cnf"
    echo "生成 Runner 客户端证书..."
    openssl genrsa -out "$CERTS_DIR/runner.key" 2048
    openssl req -new -key "$CERTS_DIR/runner.key" \
        -out "$CERTS_DIR/runner.csr" \
        -subj "/CN=${RUNNER_NODE_ID}/O=${RUNNER_ORG_SLUG}/OU=Runner"
    cat > "$CERTS_DIR/runner_ext.cnf" << 'EOF'
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
EOF
    openssl x509 -req -days 365 \
        -in "$CERTS_DIR/runner.csr" \
        -CA "$SSL_DIR/ca.crt" -CAkey "$SSL_DIR/ca.key" \
        -CAserial "$CERTS_DIR/ca.srl" -CAcreateserial \
        -out "$CERTS_DIR/runner.crt" \
        -extfile "$CERTS_DIR/runner_ext.cnf"
    cp "$SSL_DIR/ca.crt" "$CERTS_DIR/ca.crt"
    rm -f "$CERTS_DIR/runner.csr" "$CERTS_DIR/runner_ext.cnf"
    chmod 600 "$CERTS_DIR/runner.key"
    chmod 644 "$CERTS_DIR/runner.crt" "$CERTS_DIR/ca.crt"
    echo "✓ Runner 证书生成完成"
}

init_claude_config() {
    local claude_dir="${HOME}/.claude"
    local claude_actual="${claude_dir}/claude.json"
    local claude_link="${HOME}/.claude.json"
    mkdir -p "$claude_dir"
    if [[ ! -f "$claude_actual" ]]; then
        cat > "$claude_actual" << 'EOF'
{
  "hasCompletedOnboarding": true,
  "theme": "dark",
  "autoUpdaterStatus": "disabled",
  "shiftEnterKeyBindingInstalled": true
}
EOF
    fi
    if [[ ! -L "$claude_link" ]]; then
        rm -f "$claude_link"
        ln -s "$claude_actual" "$claude_link"
    fi
    if [[ ! -f "$claude_dir/settings.json" ]]; then
        cat > "$claude_dir/settings.json" << 'EOF'
{
  "permissions": {
    "allow": ["Bash(*)", "Read(*)", "Write(*)", "Edit(*)", "Glob(*)", "Grep(*)", "WebFetch(*)", "WebSearch(*)"],
    "deny": []
  },
  "autoUpdaterStatus": "disabled",
  "spinnerTipsEnabled": false
}
EOF
    fi
}

init_codex_config() {
    mkdir -p "${HOME}/.codex"
    if [[ ! -f "${HOME}/.codex/config.toml" ]]; then
        cat > "${HOME}/.codex/config.toml" << 'EOF'
model = "gpt-4.1"
approval_policy = "never"
sandbox_mode = "danger-full-access"

[shell_environment_policy]
inherit = "all"
EOF
    fi
}

init_gemini_config() {
    mkdir -p "${HOME}/.gemini"
    if [[ ! -f "${HOME}/.gemini/settings.json" ]]; then
        cat > "${HOME}/.gemini/settings.json" << 'EOF'
{
  "coreTools": ["read_file", "edit_file", "write_file", "run_shell_command", "search_files", "list_directory", "web_search"],
  "excludeTools": [],
  "theme": "Default (Dark)",
  "checkForUpdates": false,
  "sandbox": false,
  "yolo": false
}
EOF
    fi
}

init_ai_cli_configs() {
    case "${AGENT_RUNTIME}" in
        claude-code) init_claude_config ;;
        codex-cli) init_codex_config ;;
        gemini-cli) init_gemini_config ;;
        do-agent) init_do_agent_config ;;
        e2e-echo|loopal|aider|opencode) ;;
    esac
}

init_do_agent_config() {
    local settings="${HOME}/.agent/settings.json"
    mkdir -p "${HOME}/.agent"
    if [[ ! -f "$settings" ]]; then
        cat > "$settings" << 'EOF'
{
  "model": "minimax/MiniMax-M3"
}
EOF
    fi
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
description: "Development Docker Runner (Bazel binary)"
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
    generate_runner_cert
    init_ai_cli_configs
    create_config
    echo "启动 Runner (bazel-built binary)..."
    exec /usr/local/bin/do-worker-runner run --config "$CONFIG_FILE"
}
main "$@"
