#!/bin/bash
# shellcheck shell=bash
# Deploy dev runners on local Kubernetes (Docker Desktop / kind) instead of compose.

# shellcheck source=generate_runners_k8s_manifest.sh
source "$SCRIPT_DIR/lib/generate_runners_k8s_manifest.sh"

runners_k8s_enabled() {
    [[ "${RUNNERS_LAUNCHER:-docker}" == "k8s" ]]
}

ensure_k8s_cluster() {
    if ! command -v kubectl &>/dev/null; then
        error "kubectl 未安装 — 无法部署 K8s runner"
        return 1
    fi
    if ! kubectl cluster-info &>/dev/null; then
        error "Kubernetes 集群不可用 — 请在 Docker Desktop 启用 Kubernetes"
        return 1
    fi
    success "Kubernetes 集群就绪: $(kubectl config current-context)"
}

build_runner_compose_images() {
    local node_arch platform
    node_arch="$(kubectl get nodes -o jsonpath='{.items[0].status.nodeInfo.architecture}')"
    platform="linux/${node_arch:-amd64}"
    local node_base_image="${NODE_BASE_IMAGE:-node:24-bookworm-slim}"
    local python_base_image="${PYTHON_BASE_IMAGE:-python:3.11-slim-bookworm}"

    info "构建 runner 镜像 (K8s 节点架构: ${platform})..."
    cd "$SCRIPT_DIR"
    local rt
    for rt in e2e-echo claude-code codex-cli cursor-cli gemini-cli loopal minimax-cli openclaw hermes; do
        docker build --platform "$platform" \
            -f ../../docker/agent-runtime/Dockerfile \
            --build-arg "AGENT_RUNTIME=${rt}" \
            --build-arg "NODE_BASE_IMAGE=${node_base_image}" \
            --build-arg "PYTHON_BASE_IMAGE=${python_base_image}" \
            --build-arg "HTTP_PROXY=" \
            --build-arg "HTTPS_PROXY=" \
            -t "do-worker/runner-${rt}:latest" \
            -t "${COMPOSE_PROJECT_NAME}-runner-${rt}:latest" \
            . || {
            error "runner 镜像构建失败: ${rt}"
            return 1
        }
    done
    success "runner 镜像已构建 (do-worker/runner-* + ${COMPOSE_PROJECT_NAME}-runner-*)"
}

wait_for_runner_pods() {
    info "等待 runner Deployment 就绪..."
    local dep
    for dep in runner-claude-code runner-codex-cli runner-cursor-cli runner-e2e-echo runner-e2e-echo-2 \
               runner-gemini-cli runner-loopal runner-minimax-cli runner-openclaw runner-hermes; do
        kubectl rollout status "deployment/${dep}" -n agentsmesh --timeout=300s || {
            warn "${dep} rollout 超时"
            kubectl get pods -n agentsmesh -l "app=${dep}" 2>/dev/null || true
            return 1
        }
    done
    success "runner 集群 Deployment 已就绪"
}

deploy_runners_k8s() {
    ensure_k8s_cluster || return 1
    build_runner_compose_images || return 1
    generate_runners_k8s_manifest || return 1

    info "kubectl apply runner 集群..."
    kubectl apply -f "$SCRIPT_DIR/runtime/runners-k8s/manifest.yaml" || {
        error "kubectl apply 失败"
        return 1
    }
    wait_for_runner_pods
}

teardown_runners_k8s() {
    if ! command -v kubectl &>/dev/null; then
        return 0
    fi
    if kubectl get namespace agentsmesh &>/dev/null; then
        info "删除 K8s runner 命名空间 agentsmesh..."
        kubectl delete namespace agentsmesh --wait=false 2>/dev/null || true
    fi
}

hot_swap_runner_k8s_binary() {
    local dep pod
    for dep in runner-claude-code runner-codex-cli runner-cursor-cli runner-e2e-echo runner-e2e-echo-2 \
               runner-gemini-cli runner-loopal runner-minimax-cli runner-openclaw runner-hermes; do
        pod="$(kubectl get pods -n agentsmesh -l "app=${dep}" \
            -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)"
        if [[ -z "$pod" ]]; then
            info "跳过 ${dep} (Pod 不存在)"
            continue
        fi
        kubectl cp "$SCRIPT_DIR/runner-binary" "agentsmesh/${pod}:/usr/local/bin/agentsmesh-runner"
        case "$dep" in
            runner-e2e-echo|runner-e2e-echo-2)
                kubectl cp "$SCRIPT_DIR/e2e-mock-agent-binary" \
                    "agentsmesh/${pod}:/usr/local/bin/e2e-mock-agent"
                ;;
        esac
        kubectl delete pod -n agentsmesh "$pod" --wait=false
        info "已热更新 ${dep} (${pod})"
    done
    wait_for_runner_pods || true
}
