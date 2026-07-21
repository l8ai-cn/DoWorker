#!/bin/bash
# =============================================================================
# dev-lite.sh — 低内存开发环境
# =============================================================================
#
# 对比完整 ./dev.sh:
#   - backend / relay: air + go build（~几十 MB，改 .go 秒级重编译）
#   - runner: go cross-compile（一次性）
#   - runner 容器: Coordinator 按需创建（不预起 12 个）
#   - 前端: 仅 web 主站（跳过 web-admin / web-user）
#   - 前端: plain next dev + pnpm build:wasm
#
# 用法:
#   ./dev-lite.sh                 # docker + air backend/relay + web
#   ./dev-lite.sh --backend-only  # 不启前端
#   ./dev-lite.sh --frontends     # 栈已起时只重启 web
#   ./dev-lite.sh --clean         # 同 dev.sh --clean
#
# 可选: 无 Rust toolchain 时需预构建 packages/agent-cloud-wasm
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

export DEV_LITE=1
export WEB_USER_SKIP=1

# 默认 Coordinator 按需 runner；若 .env 已持久化 RUNNERS_LAUNCHER 则 dev.sh 会尊重。
if [[ -z "${RUNNERS_LAUNCHER:-}" ]]; then
    export RUNNERS_LAUNCHER=coordinator
fi

info_banner() {
    echo ""
    echo "=========================================="
    echo "  Agent Cloud dev-lite（低内存模式）"
    echo "  Go: air  |  Runner: 按需  |  Web: 仅主站"
    echo "=========================================="
    echo ""
}

case "${1:-}" in
    --help|-h)
        exec "$SCRIPT_DIR/dev.sh" --help
        ;;
    --clean|-c)
        exec "$SCRIPT_DIR/dev.sh" --clean
        ;;
esac

info_banner
exec "$SCRIPT_DIR/dev.sh" "$@"
