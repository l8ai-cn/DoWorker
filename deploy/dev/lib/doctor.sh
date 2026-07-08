# shellcheck shell=bash
# doctor.sh — fail-fast prereq check.
#
# Why ibazel is mandatory: dev.sh's contract is "hot-reload dev environment".
# Without ibazel, host services can't pick up `.go` edits without manual
# restart, breaking the contract. CI installs ibazel explicitly (see
# .github/workflows/bazel.yml) — no fallback in this script.

check_ibazel_doctor() {
    local missing=()
    command -v bazel  >/dev/null 2>&1 || missing+=("bazel (bazelisk)")
    command -v ibazel >/dev/null 2>&1 || missing+=("ibazel (bazel-watcher)")
    if (( ${#missing[@]} > 0 )); then
        error "缺少必需工具：${missing[*]}"
        echo "  bazel:   brew install bazelisk (macOS) | bazelisk releases (Linux)"
        echo "  ibazel:  https://github.com/bazelbuild/bazel-watcher/releases"
        echo "           (no homebrew formula — grab the darwin-arm64 / linux-amd64 binary)"
        exit 1
    fi

    # Host AI CLIs are only needed when running a non-Docker runner locally.
    # deploy/dev runner images install their selected runtime at build time.
    local ai_missing=()
    command -v claude >/dev/null 2>&1 || ai_missing+=("claude")
    command -v codex  >/dev/null 2>&1 || ai_missing+=("codex")
    command -v gemini >/dev/null 2>&1 || ai_missing+=("gemini")
    if (( ${#ai_missing[@]} > 0 )); then
        warn "宿主机 AI CLI 未全装（${ai_missing[*]}）— 仅影响非 Docker runner"
        echo "  npm i -g @anthropic-ai/claude-code @openai/codex @google/gemini-cli"
    fi
}

# macOS apps (e.g. Baidu netdisk) sometimes bind 127.0.0.1:HTTP_PORT while
# Traefik listens on *:HTTP_PORT. Web dev now proxies API to BACKEND_HTTP_PORT
# directly; this warns so operators know why 127.0.0.1:10000 may not reach Traefik.
warn_loopback_port_conflict() {
    local http_port="${HTTP_PORT:-10000}"
    local backend_port="${BACKEND_HTTP_PORT:-10015}"
    local occupier=""
    while IFS= read -r line; do
        local proc addr
        proc=$(awk '{print $1}' <<< "$line")
        addr=$(awk '{print $9}' <<< "$line")
        if [[ "$addr" == "127.0.0.1:$http_port" && "$proc" != "com.docke" ]]; then
            occupier="$proc"
            break
        fi
    done < <(lsof -nP -iTCP:"$http_port" -sTCP:LISTEN 2>/dev/null || true)

    if [[ -n "$occupier" ]]; then
        warn "127.0.0.1:$http_port 被 ${occupier} 占用（Traefik 仍可通过 localhost:$http_port 访问）"
        info "Web API 代理已配置为直连 backend :$backend_port，登录/API 不受影响"
    fi
}
