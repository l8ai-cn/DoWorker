#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
ROOT="$(cd ../.. && pwd)"
DOCKERFILE="${ROOT}/docker/agent-runtime/Dockerfile"

if grep -q "^RUN npm install -g" "$DOCKERFILE" \
  && grep -A5 "^RUN npm install -g" "$DOCKERFILE" | grep -q "@anthropic-ai/claude-code" \
  && grep -A5 "^RUN npm install -g" "$DOCKERFILE" | grep -q "@openai/codex"; then
  echo "docker/agent-runtime/Dockerfile installs claude-code and codex in the same layer" >&2
  exit 1
fi

grep -q "ARG AGENT_RUNTIME" "$DOCKERFILE"
grep -q "AGENT_RUNTIME: claude-code" docker-compose.runners.yml
grep -q "AGENT_RUNTIME: codex-cli" docker-compose.runners.yml
grep -q "AGENT_RUNTIME: do-agent" docker-compose.runners.yml
grep -q "runner-do-agent" docker-compose.runners.yml
grep -q "AGENT_RUNTIME: aider" docker-compose.runners.yml
grep -q "AGENT_RUNTIME: opencode" docker-compose.runners.yml
grep -q "do-agent-binary" "$DOCKERFILE"
grep -q "runner-claude-code" docker-compose.runners.yml
grep -q "runner-codex-cli" docker-compose.runners.yml
grep -q "docker/agent-runtime/Dockerfile" docker-compose.runners.yml
grep -q "COORDINATOR_RUNNER_DOCKER_COMPOSE_SERVICES" lib/host_services.sh
grep -q "case \"\${AGENT_RUNTIME}\"" runner-entrypoint.sh
grep -q "default_agent: \"\${DEFAULT_AGENT}\"" runner-entrypoint.sh
grep -q "RUNNER_SSH_SOURCE_DIR" runner-entrypoint.sh
grep -q "init_runner_ssh" runner-entrypoint.sh
grep -q "'e2e-mock-agent'," seed/e2e_echo.sql

if grep -q '/home/runner/.ssh:ro' docker-compose.runners.yml; then
  echo "runner SSH material must not be mounted over the runner home directory" >&2
  exit 1
fi
grep -q './runner-ssh:/run/runner-ssh-source:ro' docker-compose.runners.yml

if awk '/runner-claude-code:/{flag=1; next} /runner-codex-cli:/{flag=0} flag' docker-compose.runners.yml \
  | grep -q "/home/runner/.codex"; then
  echo "runner-claude-code mounts codex config" >&2
  exit 1
fi

if awk '/runner-codex-cli:/{flag=1; next} /runner-gemini-cli:/{flag=0} flag' docker-compose.runners.yml \
  | grep -q "/home/runner/.claude"; then
  echo "runner-codex-cli mounts claude config" >&2
  exit 1
fi
