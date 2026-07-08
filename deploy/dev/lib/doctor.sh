# shellcheck shell=bash
# doctor.sh — fail-fast prereq check.
#
# Why ibazel is mandatory: dev.sh's contract is "hot-reload dev environment".
# Without ibazel, host services can't pick up `.go` edits without manual
# restart, breaking the contract. CI installs ibazel explicitly (see
# .github/workflows/bazel.yml) — no fallback in this script.

check_ibazel_doctor() {
    if [[ "${DEV_NO_BAZEL:-}" == "1" || "${DEV_LITE:-}" == "1" ]]; then
        check_lite_doctor
        return
    fi
    local missing=()
    command -v bazel  >/dev/null 2>&1 || missing+=("bazel (bazelisk)")
    command -v ibazel >/dev/null 2>&1 || missing+=("ibazel (bazel-watcher)")
    if (( ${#missing[@]} > 0 )); then
        error "缺少必需工具：${missing[*]}"
        echo "  bazel:   brew install bazelisk (macOS) | bazelisk releases (Linux)"
        echo "  ibazel:  https://github.com/bazelbuild/bazel-watcher/releases"
        echo "           (no homebrew formula — grab the darwin-arm64 / linux-amd64 binary)"
        echo ""
        echo "  低内存模式可跳过 Bazel/ibazel: ./dev-lite.sh"
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

# Lite dev: Go + air for backend/relay; Bazel only for Next.js frontend (one app).
check_lite_doctor() {
    local missing=()
    command -v go >/dev/null 2>&1 || missing+=("go")
    command -v docker >/dev/null 2>&1 || missing+=("docker")
    command -v pnpm >/dev/null 2>&1 || missing+=("pnpm")
    command -v bazel >/dev/null 2>&1 || missing+=("bazel (frontend next_dev only)")
    if (( ${#missing[@]} > 0 )); then
        error "dev-lite 缺少工具：${missing[*]}"
        echo "  go:     https://go.dev/dl/"
        echo "  docker: Docker Desktop"
        echo "  pnpm:   npm install -g pnpm"
        echo "  bazel:  brew install bazelisk  (仅启动 web 前端时需要)"
        exit 1
    fi
    if ! command -v air >/dev/null 2>&1 && [[ ! -x "$(go env GOPATH)/bin/air" ]]; then
        info "air 未安装 — dev-lite 启动时会自动 go install"
    fi
    if ! command -v protoc >/dev/null 2>&1 && [[ ! -f "$SCRIPT_DIR/../../proto/gen/go/loop/v1/loop.pb.go" ]]; then
        warn "protoc 未安装且 proto/gen/go 缺失 — 首次启动会用 bazel 一次性生成 Go proto"
        echo "  推荐: brew install protobuf && ./scripts/proto-gen-go.sh"
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
