#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
ROOT="$(cd ../.. && pwd)"
DOCKERFILE="${ROOT}/docker/agent-runtime/Dockerfile"
COMPOSE="${ROOT}/deploy/dev/docker-compose.runners.yml"
MINIMAX_WRAPPER="${ROOT}/docker/agent-runtime/minimax-cli-wrapper.sh"

if grep -q "^RUN npm install -g" "$DOCKERFILE" \
  && grep -A5 "^RUN npm install -g" "$DOCKERFILE" | grep -q "@anthropic-ai/claude-code" \
  && grep -A5 "^RUN npm install -g" "$DOCKERFILE" | grep -q "@openai/codex"; then
  echo "Dockerfile installs claude-code and codex in the same unconditional layer" >&2
  exit 1
fi

grep -q "ARG AGENT_RUNTIME" "$DOCKERFILE"
grep -q "case \"\${AGENT_RUNTIME}\"" "$DOCKERFILE"
# AGENT_RUNTIME must be redeclared after the build context is copied so the
# selected sidecar can be installed explicitly.
awk '
  /^COPY --chmod=0755 runner-entrypoint\.sh/ { after_copy=1; next }
  after_copy && /^ARG AGENT_RUNTIME/ { found=1; exit }
  END { exit found ? 0 : 1 }
' "$DOCKERFILE" || {
  echo "Dockerfile must redeclare ARG AGENT_RUNTIME after binary COPY" >&2
  exit 1
}
grep -q "@anthropic-ai/claude-code" "$DOCKERFILE"
grep -q "@openai/codex" "$DOCKERFILE"
grep -q "video-studio)" "$DOCKERFILE"
grep -q "REMOTION_VERSION" "$DOCKERFILE"
grep -q "fonts-noto-cjk" "$DOCKERFILE"
grep -q "CHROME_BIN=/usr/bin/chromium" "$DOCKERFILE"
grep -Fq "s|http://deb.debian.org|https://deb.debian.org|g" "$DOCKERFILE"
grep -q "@google/gemini-cli" "$DOCKERFILE"
grep -q "@xai-official/grok" "$DOCKERFILE"
grep -q "https://cursor.com/install" "$DOCKERFILE"
grep -q -- "--retry-all-errors" "$DOCKERFILE"
grep -q "/opt/cursor-agent" "$DOCKERFILE"
grep -q "/usr/local/bin/agent" "$DOCKERFILE"
grep -q 'npm install -g "mmx-cli@${MINIMAX_CLI_VERSION}"' "$DOCKERFILE"
grep -q "MINIMAX_CLI_VERSION" "$DOCKERFILE"
grep -q "minimax-cli-wrapper.sh" "$DOCKERFILE"
grep -q "MINIMAX_API_KEY" "$MINIMAX_WRAPPER"
grep -q "MMX_CONFIG_DIR" "$MINIMAX_WRAPPER"
grep -q "mmx-cli/dist/mmx.mjs" "$MINIMAX_WRAPPER"
grep -q 'npm install -g "openclaw@${OPENCLAW_VERSION}"' "$DOCKERFILE"
grep -q "hermes-agent" "$DOCKERFILE"
grep -q "HERMES_AGENT_VERSION" "$DOCKERFILE"
grep -q "COPY --chmod=0755 binaries/" "$DOCKERFILE"
grep -q "ARG RUNTIME_BASE=base" "$DOCKERFILE"
grep -q "install_python_pip()" "$DOCKERFILE"
grep -q "install -m 0755 /usr/local/lib/do-worker/do-agent-binary" "$DOCKERFILE"
grep -q "runner-entrypoint.sh" "$DOCKERFILE"

grep -q "AGENT_RUNTIME: claude-code" "$COMPOSE"
grep -q "AGENT_RUNTIME: codex-cli" "$COMPOSE"
grep -q "AGENT_RUNTIME: video-studio" "$COMPOSE"
grep -q "AGENT_RUNTIME: gemini-cli" "$COMPOSE"
grep -q "AGENT_RUNTIME: cursor-cli" "$COMPOSE"
grep -q "AGENT_RUNTIME: do-agent" "$COMPOSE"
grep -q "AGENT_RUNTIME: grok-build" "$COMPOSE"
grep -q "AGENT_RUNTIME: minimax-cli" "$COMPOSE"
grep -q "AGENT_RUNTIME: openclaw" "$COMPOSE"
grep -q "AGENT_RUNTIME: hermes" "$COMPOSE"
grep -q "AGENT_RUNTIME: aider" "$COMPOSE"
grep -q "AGENT_RUNTIME: opencode" "$COMPOSE"
grep -q "runner-claude-code" "$COMPOSE"
grep -q "runner-codex-cli" "$COMPOSE"
grep -q "runner-video-studio" "$COMPOSE"
grep -q "runner-cursor-cli" "$COMPOSE"
grep -q "runner-grok-build" "$COMPOSE"
grep -q "runner-minimax-cli" "$COMPOSE"
grep -q "runner-openclaw" "$COMPOSE"
grep -q "runner-hermes" "$COMPOSE"
grep -q "docker/agent-runtime/Dockerfile" "$COMPOSE"
grep -q "target: runtime" "$COMPOSE"

grep -q "case \"\${AGENT_RUNTIME}\"" "${ROOT}/deploy/dev/runner-entrypoint.sh"

if awk '/runner-claude-code:/{flag=1; next} /runner-codex-cli:/{flag=0} flag' "$COMPOSE" \
  | grep -q "/home/runner/.codex"; then
  echo "runner-claude-code mounts codex config" >&2
  exit 1
fi

if awk '/runner-codex-cli:/{flag=1; next} /runner-gemini-cli:/{flag=0} flag' "$COMPOSE" \
  | grep -q "/home/runner/.claude"; then
  echo "runner-codex-cli mounts claude config" >&2
  exit 1
fi
