#!/usr/bin/env bash
# Cross-compile do-agent (linux/amd64) into deploy/dev/do-agent-binary for
# runner.Dockerfile COPY. Uses a one-shot rust container so macOS hosts
# don't need a working openssl cross toolchain.

build_do_agent_binary() {
    if [[ -f "$SCRIPT_DIR/do-agent-binary" ]] \
        && file -b "$SCRIPT_DIR/do-agent-binary" | grep -q 'ELF.*x86-64'; then
        info "do-agent binary 已存在，跳过 Docker 编译"
        return 0
    fi

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
        # COPY's do-agent-binary into every runner image, so emit a tiny ELF
        # stub that exits 127 — enough for image build; do-agent pods won't run.
        if [[ "${CI:-}" == "true" || "${SKIP_DOAGENT_BUILD:-}" == "1" ]]; then
            info "doagent 源码未找到 — CI 写入 do-agent stub (设 DOAGENT_DIR 可启用真编译)"
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

# Minimal linux/amd64 ELF that exits 127. Satisfies Dockerfile COPY without
# needing the AgentForge/doagent source tree in CI.
_write_do_agent_stub() {
    local out="$1"
    python3 - "$out" <<'PY'
import pathlib, struct, sys
out = pathlib.Path(sys.argv[1])
# ELF64 LE ET_EXEC x86_64, one PT_LOAD + exit(127) stub at entry 0x400078
ehdr = bytearray(64)
ehdr[0:4] = b"\x7fELF"
ehdr[4:8] = bytes([2, 1, 1, 0])
struct.pack_into("<HHIQQQIHHHHHH", ehdr, 16,
                 2, 0x3E, 1, 0x400078, 64, 0, 0, 64, 56, 1, 0, 0)
phdr = struct.pack("<IIQQQQQQ", 1, 5, 0, 0x400000, 0x400000, 0x80, 0x80, 0x1000)
code = bytes([
    0x48, 0xc7, 0xc0, 0x3c, 0x00, 0x00, 0x00,  # mov rax, 60 (exit)
    0x48, 0xc7, 0xc7, 0x7f, 0x00, 0x00, 0x00,  # mov rdi, 127
    0x0f, 0x05,                                # syscall
])
blob = bytes(ehdr) + phdr + code
out.write_bytes(blob.ljust(0x80, b"\x00"))
out.chmod(0o755)
PY
}
