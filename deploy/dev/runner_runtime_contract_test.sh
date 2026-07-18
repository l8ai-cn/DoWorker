#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
ROOT="$(cd ../.. && pwd)"
DOCKERFILE="${ROOT}/docker/agent-runtime/Dockerfile"
K8S_RUNNER_MANIFEST="${ROOT}/deploy/dev/lib/generate_runners_k8s_manifest.sh"

if grep -q "^RUN npm install -g" "$DOCKERFILE" \
  && grep -A5 "^RUN npm install -g" "$DOCKERFILE" | grep -q "@anthropic-ai/claude-code" \
  && grep -A5 "^RUN npm install -g" "$DOCKERFILE" | grep -q "@openai/codex"; then
  echo "docker/agent-runtime/Dockerfile installs claude-code and codex in the same layer" >&2
  exit 1
fi

grep -q "ARG AGENT_RUNTIME" "$DOCKERFILE"
grep -q "AGENT_RUNTIME: claude-code" docker-compose.runners.yml
grep -q "AGENT_RUNTIME: codex-cli" docker-compose.runners.yml
grep -q "AGENT_RUNTIME: cursor-cli" docker-compose.runners.yml
grep -q "AGENT_RUNTIME: do-agent" docker-compose.runners.yml
grep -q "AGENT_RUNTIME: grok-build" docker-compose.runners.yml
grep -q "AGENT_RUNTIME: minimax-cli" docker-compose.runners.yml
grep -q "AGENT_RUNTIME: openclaw" docker-compose.runners.yml
grep -q "AGENT_RUNTIME: hermes" docker-compose.runners.yml
grep -q "runner-do-agent" docker-compose.runners.yml
grep -q "runner-grok-build" docker-compose.runners.yml
grep -q "runner-minimax-cli" docker-compose.runners.yml
grep -q "runner-openclaw" docker-compose.runners.yml
grep -q "runner-hermes" docker-compose.runners.yml
grep -q "runner-openclaw" ../kubernetes/local/runners-workloads.yaml
grep -q "runner-hermes" ../kubernetes/local/runners-workloads.yaml
grep -q "runner-minimax-cli" ../kubernetes/local/runners-workloads.yaml
grep -q "runner-cursor-cli" ../kubernetes/local/runners-workloads.yaml
grep -q "cursor-cli" lib/runners_k8s.sh
grep -q "AGENT_RUNTIME: aider" docker-compose.runners.yml
grep -q "AGENT_RUNTIME: opencode" docker-compose.runners.yml
grep -q "COPY --chmod=0755 binaries/" "$DOCKERFILE"
grep -q "@xai-official/grok" "$DOCKERFILE"
grep -q 'npm install -g "mmx-cli@${MINIMAX_CLI_VERSION}"' "$DOCKERFILE"
grep -q 'npm install -g "openclaw@${OPENCLAW_VERSION}"' "$DOCKERFILE"
grep -q "hermes-agent" "$DOCKERFILE"
grep -q "HERMES_AGENT_VERSION" "$DOCKERFILE"
grep -q "runner-claude-code" docker-compose.runners.yml
grep -q "runner-codex-cli" docker-compose.runners.yml
grep -q "runner-cursor-cli" docker-compose.runners.yml
grep -q "docker/agent-runtime/Dockerfile" docker-compose.runners.yml
grep -q "COORDINATOR_RUNNER_DOCKER_COMPOSE_SERVICES" lib/host_services.sh
grep -q "case \"\${AGENT_RUNTIME}\"" runner-entrypoint.sh
grep -q "claude-code|codex-cli|cursor-cli|" runner-entrypoint.sh
grep -q 'HTTP_PROXY: ${RUNNER_HTTP_PROXY:-}' docker-compose.runners.yml
grep -q 'HTTPS_PROXY: ${RUNNER_HTTPS_PROXY:-}' docker-compose.runners.yml
grep -q 'NO_PROXY: ${RUNNER_NO_PROXY:-traefik,host.lan,host.docker.internal,localhost,127.0.0.1,::1,postgres,redis,otel-collector}' docker-compose.runners.yml
grep -q 'HTTP_PROXY: "${RUNNER_HTTP_PROXY:-}"' "$K8S_RUNNER_MANIFEST"
grep -q 'HTTPS_PROXY: "${RUNNER_HTTPS_PROXY:-}"' "$K8S_RUNNER_MANIFEST"
grep -q 'NO_PROXY: "${RUNNER_NO_PROXY:-host.docker.internal,localhost,127.0.0.1,::1,otel-collector,.svc,.cluster.local}"' "$K8S_RUNNER_MANIFEST"
grep -q "default_agent: \"\${DEFAULT_AGENT}\"" runner-entrypoint.sh
grep -q "init_runner_ssh" runner-entrypoint.sh
grep -q 'runner-ssh:/run/runner-ssh-source:ro' docker-compose.runners.yml
grep -q 'sudo install -d -m 700 -o runner -g runner "$ssh_dir"' runner-entrypoint.sh
grep -q 'sudo install -m 600 -o runner -g runner "$RUNNER_SSH_SOURCE_DIR/id_ed25519" "$ssh_dir/id_ed25519"' runner-entrypoint.sh
if grep -q 'runner-ssh:/home/runner/.ssh:ro' docker-compose.runners.yml; then
  echo "runner SSH source must not be mounted directly at ~/.ssh" >&2
  exit 1
fi
grep -q "'e2e-mock-agent'," seed/e2e_echo.sql

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
