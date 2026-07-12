#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
ROOT="$(cd ../.. && pwd)"
DOCKERFILE="${ROOT}/docker/agent-runtime/Dockerfile"
COMPOSE="${ROOT}/deploy/dev/docker-compose.runners.yml"
PREPARE="${ROOT}/docker/agent-runtime/prepare_binaries.sh"
DOAGENT_BUILD="${ROOT}/deploy/dev/lib/build_do_agent_binary.sh"

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
  /^COPY --chmod=0755 runner-entrypoint\.sh/ { after_copy=1; next }
  after_copy && /^ARG AGENT_RUNTIME/ { found=1; exit }
  END { exit found ? 0 : 1 }
' "$DOCKERFILE" || {
  echo "Dockerfile must redeclare ARG AGENT_RUNTIME after binary COPY" >&2
  exit 1
}
grep -q "@anthropic-ai/claude-code" "$DOCKERFILE"
grep -q "@openai/codex" "$DOCKERFILE"
grep -q "@google/gemini-cli" "$DOCKERFILE"
grep -q "npm install -g mmx-cli" "$DOCKERFILE"
grep -q "@xai-official/grok" "$DOCKERFILE"
grep -q "npm install -g openclaw" "$DOCKERFILE"
grep -q "hermes-agent" "$DOCKERFILE"
grep -q "HERMES_AGENT_VERSION" "$DOCKERFILE"
grep -q "do-agent-binary" "$DOCKERFILE"
grep -q "FROM runtime AS do-agent-runtime" "$DOCKERFILE"
grep -q "runner-entrypoint.sh" "$DOCKERFILE"
grep -q "minimax-cli) init_minimax_config" "${ROOT}/deploy/dev/runner-entrypoint.sh"

if grep -Eq 'e2e-mock-agent.*do-agent-binary' "$PREPARE" \
  || grep -q "_write_do_agent_stub" "$DOAGENT_BUILD"; then
  echo "do-agent build must not substitute a mock or stub binary" >&2
  exit 1
fi

grep -q "AGENT_RUNTIME: claude-code" "$COMPOSE"
grep -q "AGENT_RUNTIME: codex-cli" "$COMPOSE"
grep -q "AGENT_RUNTIME: gemini-cli" "$COMPOSE"
grep -q "AGENT_RUNTIME: do-agent" "$COMPOSE"
grep -q "AGENT_RUNTIME: grok-build" "$COMPOSE"
grep -q "AGENT_RUNTIME: openclaw" "$COMPOSE"
grep -q "AGENT_RUNTIME: hermes" "$COMPOSE"
grep -q "AGENT_RUNTIME: aider" "$COMPOSE"
grep -q "AGENT_RUNTIME: opencode" "$COMPOSE"
grep -q "runner-claude-code" "$COMPOSE"
grep -q "runner-codex-cli" "$COMPOSE"
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
