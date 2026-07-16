#!/usr/bin/env bash
# Cross-compile do-agent (linux/amd64) into deploy/dev/do-agent-binary for
# runner.Dockerfile COPY. Uses a one-shot rust container so macOS hosts
# don't need a working openssl cross toolchain.

build_do_agent_binary() {
    if [[ -x "$SCRIPT_DIR/do-agent-binary" ]] \
        && ! grep -aq "do-agent stub: source not built" "$SCRIPT_DIR/do-agent-binary"; then
        info "do-agent binary 已存在，跳过 Docker 编译"
        return 0
    fi
    rm -f "$SCRIPT_DIR/do-agent-binary"

    local doagent_dir="${DOAGENT_DIR:-}"
    if [[ -z "$doagent_dir" ]]; then
        for candidate in \
            "$SCRIPT_DIR/../../doagent" \
            "$HOME/Documents/code/doagent" \
            "$HOME/Documents/code/AgentForge/doagent"; do
            if [[ -d "$candidate" ]]; then
                doagent_dir="$candidate"
                break
            fi
        done
    fi
    if [[ ! -d "$doagent_dir" ]]; then
        # CI / fresh clones often lack the sibling doagent repo. Docker still
        # COPY's do-agent-binary into every runner image, so emit a /bin/sh
        # stub that exits 127 — enough for image build; do-agent pods won't run.
        if [[ "${CI:-}" == "true" || "${DEV_SKIP_DOAGENT:-}" == "1" || "${SKIP_DOAGENT_BUILD:-}" == "1" ]]; then
            info "doagent 源码未找到 — 写入 do-agent stub (设 DOAGENT_DIR 可启用真编译)"
            _write_do_agent_stub "$SCRIPT_DIR/do-agent-binary" || return 1
            return 0
        fi
        error "doagent 源码未找到 — 设置 DOAGENT_DIR 或 clone AgentForge/doagent"
        return 1
    fi

    info "Docker build do-agent (linux/amd64) from ${doagent_dir}..."
    docker run --rm --platform linux/amd64 \
        -e HTTP_PROXY= -e HTTPS_PROXY= -e http_proxy= -e https_proxy= -e NO_PROXY='*' \
        -v "${doagent_dir}:/src:ro" \
        -v "${SCRIPT_DIR}:/out" \
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
        error "do-agent 编译失败"
        return 1
    }
    success "do-agent binary 已复制到 deploy/dev/do-agent-binary"
}

# Shell stub that exits 127. Satisfies Dockerfile COPY without needing
# AgentForge/doagent source in CI. agent-runtime image has /bin/sh.
_write_do_agent_stub() {
    local out="$1"
    cat > "$out" <<'STUB'
#!/bin/sh
echo "do-agent stub: source not built in this environment" >&2
exit 127
STUB
    chmod +x "$out"
}
