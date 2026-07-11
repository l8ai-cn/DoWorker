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
grep -q "do-agent-binary" "$DOCKERFILE"
grep -q "runner-entrypoint.sh" "$DOCKERFILE"

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
grep -q "RUNNER_SSH_SOURCE_DIR" "${ROOT}/deploy/dev/runner-entrypoint.sh"
grep -q "init_runner_ssh" "${ROOT}/deploy/dev/runner-entrypoint.sh"
grep -q 'user: "0:0"' "$COMPOSE"
grep -q "handoff_runner_state" "${ROOT}/deploy/dev/runner-entrypoint.sh"
grep -q "exec sudo -E -H -u runner -- /usr/local/bin/do-worker-runner run" "${ROOT}/deploy/dev/runner-entrypoint.sh"

if grep -q '/home/runner/.ssh:ro' "$COMPOSE"; then
  echo "runner SSH material must not be mounted over the runner home directory" >&2
  exit 1
fi
grep -q './runner-ssh:/run/runner-ssh-source:ro' "$COMPOSE"

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
