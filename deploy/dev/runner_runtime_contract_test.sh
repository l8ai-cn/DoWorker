#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
ROOT="$(cd ../.. && pwd)"
DOCKERFILE="${ROOT}/docker/agent-runtime/Dockerfile"
PREPARE="${ROOT}/docker/agent-runtime/prepare_binaries.sh"
DOAGENT_BUILD="${ROOT}/deploy/dev/lib/build_do_agent_binary.sh"

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
grep -q "AGENT_RUNTIME: grok-build" docker-compose.runners.yml
grep -q "AGENT_RUNTIME: openclaw" docker-compose.runners.yml
grep -q "AGENT_RUNTIME: hermes" docker-compose.runners.yml
grep -q "runner-do-agent" docker-compose.runners.yml
grep -q "runner-grok-build" docker-compose.runners.yml
grep -q "runner-openclaw" docker-compose.runners.yml
grep -q "runner-hermes" docker-compose.runners.yml
grep -q "runner-openclaw" ../kubernetes/local/runners-workloads.yaml
grep -q "runner-hermes" ../kubernetes/local/runners-workloads.yaml
grep -q "AGENT_RUNTIME: aider" docker-compose.runners.yml
grep -q "AGENT_RUNTIME: opencode" docker-compose.runners.yml
grep -q "do-agent-binary" "$DOCKERFILE"
grep -q "FROM runtime AS do-agent-runtime" "$DOCKERFILE"
grep -q "npm install -g mmx-cli" "$DOCKERFILE"
grep -q "minimax-cli) init_minimax_config" runner-entrypoint.sh
grep -q "@xai-official/grok" "$DOCKERFILE"
grep -q "npm install -g openclaw" "$DOCKERFILE"
grep -q "hermes-agent" "$DOCKERFILE"
grep -q "HERMES_AGENT_VERSION" "$DOCKERFILE"
grep -q "runner-claude-code" docker-compose.runners.yml
grep -q "runner-codex-cli" docker-compose.runners.yml
grep -q "docker/agent-runtime/Dockerfile" docker-compose.runners.yml
grep -q "COORDINATOR_RUNNER_DOCKER_COMPOSE_SERVICES" lib/host_services.sh
grep -q "case \"\${AGENT_RUNTIME}\"" runner-entrypoint.sh
grep -q "default_agent: \"\${DEFAULT_AGENT}\"" runner-entrypoint.sh
grep -q "'e2e-mock-agent'," seed/e2e_echo.sql
grep -q 'DEV_SKIP_DOAGENT:-}" != "1"' dev.sh
grep -q 'DEV_E2E_RUNNERS_ONLY:-}" != "1"' dev.sh

if grep -Eq 'e2e-mock-agent.*do-agent-binary' "$PREPARE" \
  || grep -q "_write_do_agent_stub" "$DOAGENT_BUILD"; then
  echo "do-agent build must not substitute a mock or stub binary" >&2
  exit 1
fi

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
