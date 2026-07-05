#!/bin/bash
# shellcheck shell=bash
# Generates K8s ConfigMaps/Secrets + patched workload manifest from .env ports.
# SSOT for runner env/entrypoint/CA/SSH: deploy/dev/* (same as docker-compose).

generate_runners_k8s_manifest() {
    local out_dir="$SCRIPT_DIR/runtime/runners-k8s"
    local base="$SCRIPT_DIR/../kubernetes/local/runners-workloads.yaml"
    mkdir -p "$out_dir"

    if [[ ! -f "$ENV_FILE" ]]; then
        error ".env 不存在 — 先运行 generate_env"
        return 1
    fi
    # shellcheck disable=SC1090
    source "$ENV_FILE"

    local http_port="${HTTP_PORT:?HTTP_PORT missing}"
    local grpc_port="${BACKEND_GRPC_PORT:?BACKEND_GRPC_PORT missing}"
    local otel_port="${OTEL_GRPC_PORT:-10012}"
    local mcp_port="${RUNNER_MCP_PORT:-10018}"
    local mcp2_port="${RUNNER_2_MCP_PORT:-10019}"
    local project="${COMPOSE_PROJECT_NAME:?COMPOSE_PROJECT_NAME missing}"

    kubectl create namespace agentsmesh --dry-run=client -o yaml > "$out_dir/00-namespace.yaml"

    kubectl create secret generic agentsmesh-dev-ca \
        --namespace=agentsmesh \
        --from-file=ca.crt="$SCRIPT_DIR/ssl/ca.crt" \
        --from-file=ca.key="$SCRIPT_DIR/ssl/ca.key" \
        --dry-run=client -o yaml > "$out_dir/01-secret-ca.yaml"

    local ssh_args=(--namespace=agentsmesh)
    for f in id_ed25519 id_ed25519.pub config known_hosts; do
        [[ -f "$SCRIPT_DIR/runner-ssh/$f" ]] && ssh_args+=(--from-file="$f=$SCRIPT_DIR/runner-ssh/$f")
    done
    kubectl create secret generic agentsmesh-runner-ssh \
        "${ssh_args[@]}" \
        --dry-run=client -o yaml > "$out_dir/02-secret-ssh.yaml"

    cat > "$out_dir/03-configmap-env.yaml" << EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: agentsmesh-runner-env
  namespace: agentsmesh
  labels:
    app.kubernetes.io/name: agentsmesh-runner
data:
  AGENTSMESH_MCP_BIND: "0.0.0.0"
  ANTHROPIC_BASE_URL: "${ANTHROPIC_BASE_URL:-https://api.anthropic.com}"
  BACKEND_URL: "http://host.docker.internal:${http_port}"
  DEBUG: "true"
  GRPC_ENDPOINT: "host.docker.internal:${grpc_port}"
  HOME: /home/runner
  HTTP_PROXY: ""
  HTTPS_PROXY: ""
  MAX_CONCURRENT_PODS: "10"
  NO_PROXY: "*"
  OTEL_EXPORTER_OTLP_ENDPOINT: "http://host.docker.internal:${otel_port}"
  OTEL_SERVICE_NAME: agentsmesh-runner
  OTEL_TRACES_SAMPLER_ARG: "1.0"
  RELAY_BASE_URL: "ws://host.docker.internal:${http_port}/relay"
  RUNNER_ORG_SLUG: dev-org
  SSL_DIR: /app/ssl
  WORKSPACE: /workspace
  http_proxy: ""
  https_proxy: ""
EOF

    kubectl create configmap runner-entrypoint \
        --namespace=agentsmesh \
        --from-file=runner-entrypoint.sh="$SCRIPT_DIR/runner-entrypoint.sh" \
        --dry-run=client -o yaml > "$out_dir/04-configmap-entrypoint.yaml"

    sed \
        -e "s/__COMPOSE_PROJECT_NAME__/${project}/g" \
        -e "s/__RUNNER_MCP_PORT__/${mcp_port}/g" \
        -e "s/__RUNNER_2_MCP_PORT__/${mcp2_port}/g" \
        "$base" > "$out_dir/05-workloads.yaml"

    : > "$out_dir/manifest.yaml"
    local f
    for f in "$out_dir"/0*.yaml; do
        [[ -f "$f" ]] || continue
        cat "$f" >> "$out_dir/manifest.yaml"
        printf '\n---\n' >> "$out_dir/manifest.yaml"
    done
    success "生成 K8s runner manifest → $out_dir/manifest.yaml"
}
