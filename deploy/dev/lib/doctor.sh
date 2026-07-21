# shellcheck shell=bash
# doctor.sh — fail-fast prereq check (Go + Docker + pnpm).

check_dev_doctor() {
    local missing=()
    command -v go >/dev/null 2>&1 || missing+=("go")
    command -v docker >/dev/null 2>&1 || missing+=("docker")
    command -v pnpm >/dev/null 2>&1 || missing+=("pnpm")
    if (( ${#missing[@]} > 0 )); then
        error "缺少必需工具：${missing[*]}"
        echo "  go:     https://go.dev/dl/"
        echo "  docker: Docker Desktop"
        echo "  pnpm:   npm install -g pnpm"
        exit 1
    fi

    if ! command -v air >/dev/null 2>&1 && [[ ! -x "$(go env GOPATH)/bin/air" ]]; then
        info "air 未安装 — 启动时会自动 go install"
    fi

    if ! command -v protoc >/dev/null 2>&1 && [[ ! -f "$SCRIPT_DIR/../../proto/gen/go/workflow/v1/workflow.pb.go" ]]; then
        warn "protoc 未安装且 proto/gen/go 缺失 — 请先: brew install protobuf && pnpm proto:gen-go"
    fi

    local ai_missing=()
    command -v claude >/dev/null 2>&1 || ai_missing+=("claude")
    command -v codex  >/dev/null 2>&1 || ai_missing+=("codex")
    command -v gemini >/dev/null 2>&1 || ai_missing+=("gemini")
    if (( ${#ai_missing[@]} > 0 )); then
        warn "宿主机 AI CLI 未全装（${ai_missing[*]}）— 仅影响非 Docker runner"
        echo "  npm i -g @anthropic-ai/claude-code @openai/codex @google/gemini-cli"
    fi

    if ! command -v soffice >/dev/null 2>&1; then
        warn "LibreOffice soffice 未安装 — 非 Docker Runner 无法生成 DOCX/XLSX/PPTX 预览"
        echo "  macOS: brew install --cask libreoffice"
        echo "  Linux: install libreoffice-writer libreoffice-calc libreoffice-impress"
    fi
}

# Docker Desktop can bind the wildcard address while another app owns the
# IPv4 loopback address. In that state `localhost` reaches the wrong process.
require_unshadowed_loopback_port() {
    local http_port="${HTTP_PORT:-10000}"
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
        error "127.0.0.1:$http_port 被 ${occupier} 占用，localhost 不会到达 Agent Cloud Traefik"
        error "停止该进程或为当前 worktree 分配其他 HTTP_PORT 后再启动"
        return 1
    fi
}
