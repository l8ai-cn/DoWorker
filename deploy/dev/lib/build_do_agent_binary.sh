#!/usr/bin/env bash
# Cross-compile do-agent (linux/amd64) into deploy/dev/do-agent-binary.
# Uses a one-shot rust container so macOS hosts
# don't need a working openssl cross toolchain.

doagent_build_log() {
    local level="$1"
    shift
    if declare -F "$level" >/dev/null; then
        "$level" "$@"
    else
        printf '%s\n' "$*"
    fi
}

stage_doagent_binary() {
    local source="$1" destination="$2"
    if declare -F stage_runner_sidecar_binary >/dev/null; then
        stage_runner_sidecar_binary "$source" "$destination" "DoAgent"
        return
    fi
    [[ -x "$source" ]] || {
        doagent_build_log error "缺少 DoAgent 二进制: $source"
        return 1
    }
    mkdir -p "$(dirname "$destination")"
    cp "$source" "$destination"
    chmod +x "$destination"
}

build_do_agent_binary() {
    local dev_dir
    dev_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
    if [[ "${DEV_SKIP_DOAGENT:-0}" == "1" ]]; then
        doagent_build_log info "DEV_SKIP_DOAGENT=1，跳过 do-agent 构建"
        return 0
    fi
    if [[ -x "$dev_dir/do-agent-binary" ]]; then
        stage_doagent_binary \
            "$dev_dir/do-agent-binary" \
            "$dev_dir/binaries/do-agent-binary" || return 1
        doagent_build_log info "do-agent binary 已存在，跳过 Docker 编译"
        return 0
    fi

    local doagent_dir="${DOAGENT_DIR:-}"
    if [[ -z "$doagent_dir" ]]; then
        for candidate in \
            "$dev_dir/../../doagent" \
            "$HOME/Documents/code/doagent" \
            "$HOME/Documents/code/AgentForge/doagent"; do
            if [[ -d "$candidate" ]]; then
                doagent_dir="$candidate"
                break
            fi
        done
    fi
    if [[ ! -d "$doagent_dir" ]]; then
        doagent_build_log error "doagent 源码未找到 — 设置 DOAGENT_DIR 或 clone AgentForge/doagent"
        return 1
    fi

    doagent_build_log info "Docker build do-agent (linux/amd64) from ${doagent_dir}..."
    docker run --rm --platform linux/amd64 \
        -e HTTP_PROXY= -e HTTPS_PROXY= -e http_proxy= -e https_proxy= -e NO_PROXY='*' \
        -v "${doagent_dir}:/src:ro" \
        -v "${dev_dir}:/out" \
        rust:1.85-bookworm \
        bash -c 'set -e
            rustup update stable
            rustup default stable
            apt-get update -qq \
                && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq pkg-config libssl-dev file >/dev/null \
                || true
            mkdir -p /build
            tar -C /src -cf - --exclude=target --exclude=.git . | tar -C /build -xf -
            cd /build
            cargo build --release
            cp target/release/do-agent /out/do-agent-binary
            chmod +x /out/do-agent-binary
            file /out/do-agent-binary
            /out/do-agent-binary --version' || {
        doagent_build_log error "do-agent 编译失败"
        return 1
    }
    stage_doagent_binary \
        "$dev_dir/do-agent-binary" \
        "$dev_dir/binaries/do-agent-binary" || return 1
    doagent_build_log success "do-agent binary 已复制到 deploy/dev/do-agent-binary"
}
