#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
ROOT="$(cd ../.. && pwd)"
DOCKERFILE="${ROOT}/docker/agent-runtime/Dockerfile"
COMPOSE="${ROOT}/deploy/dev/docker-compose.runners.yml"

if grep -q "^RUN npm install -g" "$DOCKERFILE" \
  && grep -A5 "^RUN npm install -g" "$DOCKERFILE" | grep -q "@anthropic-ai/claude-code" \
  && grep -A5 "^RUN npm install -g" "$DOCKERFILE" | grep -q "@openai/codex"; then
  echo "Dockerfile installs claude-code and codex in the same unconditional layer" >&2
  exit 1
fi

grep -q "ARG AGENT_RUNTIME" "$DOCKERFILE"
grep -q "case \"\${AGENT_RUNTIME}\"" "$DOCKERFILE"
# AGENT_RUNTIME must be redeclared after COPY layers (ARG from a prior
# stage is not inherited). Without it the prune case sees "" and deletes
# e2e-mock-agent from every image.
awk '
  /^COPY --chmod=0755 \. \/tmp\/agent-runtime-context\// { after_copy=1; next }
  after_copy && /^ARG AGENT_RUNTIME/ { found=1; exit }
  END { exit found ? 0 : 1 }
' "$DOCKERFILE" || {
  echo "Dockerfile must redeclare ARG AGENT_RUNTIME after binary COPY" >&2
  exit 1
}
grep -q "@anthropic-ai/claude-code" "$DOCKERFILE"
grep -q "@openai/codex" "$DOCKERFILE"
grep -q "video-studio)" "$DOCKERFILE"
grep -q "chromium ffmpeg fontconfig fonts-noto-cjk python3 python3-pip" "$DOCKERFILE"
grep -q "/usr/local/bin/video-studio-codex" "$DOCKERFILE"
grep -q "@remotion/cli@" "$DOCKERFILE"
grep -q "REMOTION_VERSION" "$DOCKERFILE"
! grep -q 'NO_PROXY=\*' "$DOCKERFILE"
! grep -q 'Acquire::http::Proxy "false"' "$DOCKERFILE"
! grep -q '^ENV HTTP_PROXY=' "$DOCKERFILE"
! grep -q -- '--build-arg "HTTP_PROXY="' "${ROOT}/docker/agent-runtime/build.sh"
grep -q 'GOARCH="\$TARGET_ARCH"' "${ROOT}/docker/agent-runtime/prepare_binaries.sh"
grep -q 'ARG DEBIAN_MIRROR=' "$DOCKERFILE"
grep -q 'DEBIAN_SECURITY_MIRROR' "${ROOT}/docker/agent-runtime/build.sh"
grep -q "@google/gemini-cli" "$DOCKERFILE"
grep -q "@xai-official/grok" "$DOCKERFILE"
grep -q "npm install -g openclaw" "$DOCKERFILE"
grep -q "hermes-agent" "$DOCKERFILE"
grep -q "HERMES_AGENT_VERSION" "$DOCKERFILE"
grep -q 'PIP_BREAK_SYSTEM_PACKAGES=1 npm install -g "hermes-agent@' "$DOCKERFILE"
grep -q "do-agent) install .*do-agent-binary" "$DOCKERFILE"
grep -q "runner-entrypoint.sh" "$DOCKERFILE"

grep -q "AGENT_RUNTIME: claude-code" "$COMPOSE"
grep -q "AGENT_RUNTIME: codex-cli" "$COMPOSE"
grep -q "AGENT_RUNTIME: video-studio" "$COMPOSE"
grep -q "AGENT_RUNTIME: gemini-cli" "$COMPOSE"
grep -q "AGENT_RUNTIME: do-agent" "$COMPOSE"
grep -q "AGENT_RUNTIME: grok-build" "$COMPOSE"
grep -q "AGENT_RUNTIME: openclaw" "$COMPOSE"
grep -q "AGENT_RUNTIME: hermes" "$COMPOSE"
grep -q "AGENT_RUNTIME: aider" "$COMPOSE"
grep -q "AGENT_RUNTIME: opencode" "$COMPOSE"
grep -q "runner-claude-code" "$COMPOSE"
grep -q "runner-codex-cli" "$COMPOSE"
grep -q "runner-video-studio" "$COMPOSE"
grep -q "runner-grok-build" "$COMPOSE"
grep -q "runner-openclaw" "$COMPOSE"
grep -q "runner-hermes" "$COMPOSE"
grep -q "docker/agent-runtime/Dockerfile" "$COMPOSE"

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
