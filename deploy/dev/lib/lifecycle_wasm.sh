# shellcheck shell=bash

_do_worker_wasm_needs_build() {
    local root_dir="$1"
    local out_js="$root_dir/packages/do-worker-wasm/wasm_pkg.js"
    local core_dir="$root_dir/clients/core"
    local input

    [[ ! -f "$out_js" ]] && return 0

    for input in \
        "$root_dir/proto" \
        "$core_dir" \
        "$root_dir/scripts/build-wasm.sh" \
        "$root_dir/scripts/seed-rust-proto-stubs.sh"; do
        [[ -e "$input" ]] || continue
        if [[ -f "$input" ]]; then
            [[ "$input" -nt "$out_js" ]] && return 0
            continue
        fi
        if find "$input" \
            -path "$core_dir/target" -prune -o \
            -path "$core_dir/target/*" -prune -o \
            -path "$core_dir/crates/proto/*/src" -prune -o \
            -path "$core_dir/crates/proto/*/src/*" -prune -o \
            -type f -newer "$out_js" -print -quit | grep -q .; then
            return 0
        fi
    done
    return 1
}

_ensure_do_worker_wasm() {
    local root_dir="$SCRIPT_DIR/../.."

    if ! _do_worker_wasm_needs_build "$root_dir"; then
        return 0
    fi

    info "构建 do-worker-wasm (pnpm run build:wasm)..."
    if ! (cd "$root_dir" && pnpm run build:wasm); then
        error "do-worker-wasm 构建失败 — 纯 Next 无法解析 wasm"
        return 1
    fi
    success "do-worker-wasm 已就绪"
}
