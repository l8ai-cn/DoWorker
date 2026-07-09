#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
ROOT="$(cd ../.. && pwd)"
DOCKERFILE="${ROOT}/docker/agent-runtime/Dockerfile"
COMPOSE="${ROOT}/deploy/dev/docker-compose.runners.yml"
BUILD_SCRIPT="${ROOT}/docker/agent-runtime/build.sh"
PREPARE_SCRIPT="${ROOT}/docker/agent-runtime/prepare_binaries.sh"

if grep -q "^RUN npm install -g" "$DOCKERFILE" \
  && grep -A5 "^RUN npm install -g" "$DOCKERFILE" | grep -q "@anthropic-ai/claude-code" \
  && grep -A5 "^RUN npm install -g" "$DOCKERFILE" | grep -q "@openai/codex"; then
  echo "Dockerfile installs claude-code and codex in the same unconditional layer" >&2
  exit 1
fi

grep -q "ARG AGENT_RUNTIME" "$DOCKERFILE"
grep -q "case \"\${AGENT_RUNTIME}\"" "$DOCKERFILE"
grep -q "@anthropic-ai/claude-code" "$DOCKERFILE"
grep -q "@openai/codex" "$DOCKERFILE"
grep -q "@google/gemini-cli" "$DOCKERFILE"
grep -q "do-agent-binary" "$DOCKERFILE"
grep -q "runner-entrypoint.sh" "$DOCKERFILE"
grep -q -- "--target final" "$BUILD_SCRIPT"
grep -q "DO_AGENT_STAGE" "$BUILD_SCRIPT"
grep -q 'REQUESTED_RUNTIME.*do-agent' "$PREPARE_SCRIPT"

if grep -q -- "--target runtime" "$BUILD_SCRIPT"; then
  echo "agent-runtime builds must include the final binary-selection stage" >&2
  exit 1
fi

if grep -q "do-agent.*占位" "$PREPARE_SCRIPT"; then
  echo "prepare_binaries.sh must not substitute a fake do-agent binary" >&2
  exit 1
fi

grep -q "AGENT_RUNTIME: claude-code" "$COMPOSE"
grep -q "AGENT_RUNTIME: codex-cli" "$COMPOSE"
grep -q "AGENT_RUNTIME: gemini-cli" "$COMPOSE"
grep -q "AGENT_RUNTIME: do-agent" "$COMPOSE"
grep -q "AGENT_RUNTIME: aider" "$COMPOSE"
grep -q "AGENT_RUNTIME: opencode" "$COMPOSE"
grep -q "runner-claude-code" "$COMPOSE"
grep -q "runner-codex-cli" "$COMPOSE"
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
