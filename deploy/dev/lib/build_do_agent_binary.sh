#!/usr/bin/env bash

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
    local dev_dir repo_root
    dev_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
    repo_root="$(cd "${dev_dir}/../.." && pwd)"
    source "${repo_root}/docker/agent-runtime/do_agent_release_manifest.sh"
    if [[ "${DEV_SKIP_DOAGENT:-0}" == "1" ]]; then
        doagent_build_log info "DEV_SKIP_DOAGENT=1，跳过 do-agent 构建"
        return 0
    fi
    local expected_hash
    expected_hash="$(do_agent_release_value artifact.binary_sha256)"
    if [[ -x "$dev_dir/do-agent-binary" ]] \
        && [[ "$(do_agent_sha256 "$dev_dir/do-agent-binary")" == "$expected_hash" ]]; then
        stage_doagent_binary \
            "$dev_dir/do-agent-binary" \
            "$dev_dir/binaries/do-agent-binary" || return 1
        doagent_build_log info "do-agent binary 已存在，跳过 Docker 编译"
        return 0
    fi

    local checkout="${DOAGENT_DIR:-}"
    if ! is_do_agent_checkout "$checkout"; then
        doagent_build_log error "DOAGENT_DIR 必须指向明确的 do-agent Git checkout"
        return 1
    fi

    local source_commit source_subdirectory cargo_lock_hash cargo_toml_hash
    local builder_image rust_toolchain source_dir source_date_epoch
    source_commit="$(do_agent_release_value source.commit)"
    source_subdirectory="$(do_agent_release_value source.subdirectory)"
    cargo_lock_hash="$(do_agent_release_value source.cargo_lock_sha256)"
    cargo_toml_hash="$(do_agent_release_value source.cargo_toml_sha256)"
    builder_image="$(do_agent_release_value build.builder_image)"
    rust_toolchain="$(do_agent_release_value build.rust_toolchain)"
    source_dir="${checkout}/${source_subdirectory}"
    [[ "$(git -C "$checkout" rev-parse HEAD)" == "$source_commit" ]] || {
        doagent_build_log error "do-agent checkout 必须位于发布提交 ${source_commit}"
        return 1
    }
    [[ -z "$(git -C "$checkout" status --porcelain)" ]] || {
        doagent_build_log error "do-agent checkout 必须保持干净"
        return 1
    }
    [[ "$(do_agent_sha256 "${source_dir}/Cargo.lock")" == "$cargo_lock_hash" ]] || {
        doagent_build_log error "do-agent Cargo.lock 与发布清单不一致"
        return 1
    }
    [[ "$(do_agent_sha256 "${source_dir}/Cargo.toml")" == "$cargo_toml_hash" ]] || {
        doagent_build_log error "do-agent Cargo.toml 与发布清单不一致"
        return 1
    }
    source_date_epoch="$(git -C "$checkout" show -s --format=%ct "$source_commit")"

    rm -f "$dev_dir/do-agent-binary"
    mkdir -p /tmp/doagent-linux-cargo-registry /tmp/doagent-linux-target
    doagent_build_log info "Docker build do-agent (linux/amd64) from ${source_commit}..."
    docker run --rm --platform linux/amd64 \
        -e HTTP_PROXY= -e HTTPS_PROXY= -e http_proxy= -e https_proxy= -e NO_PROXY='*' \
        -e CARGO_INCREMENTAL=0 -e CARGO_BUILD_JOBS=1 \
        -e SOURCE_DATE_EPOCH="${source_date_epoch}" \
        -v "${source_dir}:/src:ro" \
        -v "${dev_dir}:/out" \
        -v /tmp/doagent-linux-cargo-registry:/usr/local/cargo/registry \
        -v /tmp/doagent-linux-target:/build/target \
        "${builder_image}" \
        bash -c "set -euo pipefail
            rustup toolchain install '${rust_toolchain}' --profile minimal
            mkdir -p /build
            tar -C /src -cf - --exclude=target . | tar -C /build -xf -
            cd /build
            cargo +'${rust_toolchain}' build --locked --release --jobs 1
            install -m 0755 target/release/do-agent /out/do-agent-binary
            file /out/do-agent-binary
            /out/do-agent-binary --version" || {
        doagent_build_log error "do-agent 编译失败"
        return 1
    }
    [[ "$(do_agent_sha256 "$dev_dir/do-agent-binary")" == "$expected_hash" ]] || {
        doagent_build_log error "do-agent 构建产物与发布清单哈希不一致"
        return 1
    }
    stage_doagent_binary \
        "$dev_dir/do-agent-binary" \
        "$dev_dir/binaries/do-agent-binary" || return 1
    doagent_build_log success "do-agent binary 已复制到 deploy/dev/do-agent-binary"
}

is_do_agent_checkout() {
    local checkout="${1:-}"
    [[ -n "$checkout" ]] \
        && git -C "$checkout" rev-parse --is-inside-work-tree >/dev/null 2>&1
}
