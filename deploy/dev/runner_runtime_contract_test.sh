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
grep -q "runtime-with-do-agent" "$DOCKERFILE"
grep -q "DO_AGENT_STAGE: with-do-agent" docker-compose.runners.yml
test "$(grep -c 'profiles: \["do-agent"\]' docker-compose.runners.yml)" -eq 2
grep -q "runner-claude-code" docker-compose.runners.yml
grep -q "runner-codex-cli" docker-compose.runners.yml
grep -q "docker/agent-runtime/Dockerfile" docker-compose.runners.yml
grep -q "COORDINATOR_RUNNER_DOCKER_COMPOSE_SERVICES" lib/host_services.sh
grep -q "case \"\${AGENT_RUNTIME}\"" runner-entrypoint.sh
grep -q "default_agent: \"\${DEFAULT_AGENT}\"" runner-entrypoint.sh
grep -q "'e2e-mock-agent'," seed/e2e_echo.sql

# Local/dev must still fail closed when doagent source is missing.
# CI may write a gated /bin/sh stub so Dockerfile COPY succeeds without
# cloning AgentForge/doagent — that path must stay behind CI/SKIP_DOAGENT_BUILD.
if ! grep -q 'doagent 源码未找到' lib/build_do_agent_binary.sh; then
  echo "build_do_agent_binary.sh must fail closed without real do-agent source" >&2
  exit 1
fi
if grep -q "_write_do_agent_stub\|do-agent stub" lib/build_do_agent_binary.sh \
  && ! grep -Eq 'CI:-|SKIP_DOAGENT_BUILD' lib/build_do_agent_binary.sh; then
  echo "do-agent stub must be gated on CI or SKIP_DOAGENT_BUILD" >&2
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
