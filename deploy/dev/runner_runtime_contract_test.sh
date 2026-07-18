#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
ROOT="$(cd ../.. && pwd)"
DOCKERFILE="${ROOT}/docker/agent-runtime/Dockerfile"
K8S_RUNNER_MANIFEST="${ROOT}/deploy/dev/lib/generate_runners_k8s_manifest.sh"
PREPARE="${ROOT}/docker/agent-runtime/prepare_binaries.sh"
DOAGENT_BUILD="${ROOT}/deploy/dev/lib/build_do_agent_binary.sh"
CI_WORKFLOW="${ROOT}/.github/workflows/ci.yml"
SSH_BOOTSTRAP="${ROOT}/deploy/dev/runner-ssh-bootstrap.sh"

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
grep -q "target: runtime" docker-compose.runners.yml
grep -q "COORDINATOR_RUNNER_DOCKER_COMPOSE_SERVICES" lib/host_services.sh
grep -A15 "start_marketplace_host_lite()" lib/host_services_lite.sh \
  | grep -q 'export INTERNAL_API_SECRET='
grep -q "case \"\${AGENT_RUNTIME}\"" runner-entrypoint.sh
grep -q "claude-code|codex-cli|video-studio|cursor-cli|" runner-entrypoint.sh
grep -q 'HTTP_PROXY: ${RUNNER_HTTP_PROXY:-}' docker-compose.runners.yml
grep -q 'HTTPS_PROXY: ${RUNNER_HTTPS_PROXY:-}' docker-compose.runners.yml
grep -q 'NO_PROXY: ${RUNNER_NO_PROXY:-traefik,host.lan,host.docker.internal,localhost,127.0.0.1,::1,postgres,redis,otel-collector}' docker-compose.runners.yml
grep -q 'HTTP_PROXY: "${RUNNER_HTTP_PROXY:-}"' "$K8S_RUNNER_MANIFEST"
grep -q 'HTTPS_PROXY: "${RUNNER_HTTPS_PROXY:-}"' "$K8S_RUNNER_MANIFEST"
grep -q 'NO_PROXY: "${RUNNER_NO_PROXY:-host.docker.internal,localhost,127.0.0.1,::1,otel-collector,.svc,.cluster.local}"' "$K8S_RUNNER_MANIFEST"
grep -q "default_agent: \"\${DEFAULT_AGENT}\"" runner-entrypoint.sh
grep -q "bootstrap_runner_ssh" runner-entrypoint.sh
grep -q 'runner-ssh:/run/runner-ssh-source:ro' docker-compose.runners.yml
grep -q "runner-ssh-bootstrap.sh" "$DOCKERFILE"
grep -q "runner-ssh-bootstrap.sh" "$PREPARE"
if grep -q './runner-ssh:/home/runner/.ssh' docker-compose.runners.yml; then
  echo "runner SSH source must not be mounted over the runner home directory" >&2
  exit 1
fi
grep -q "'e2e-mock-agent'," seed/e2e_echo.sql
grep -q 'DEV_SKIP_DOAGENT:-}" != "1"' dev.sh
grep -q 'DEV_E2E_RUNNERS_ONLY:-}" != "1"' dev.sh

if grep -Eq 'e2e-mock-agent.*do-agent-binary' "$PREPARE" \
  || grep -q "_write_do_agent_stub" "$DOAGENT_BUILD"; then
  echo "do-agent build must not substitute a mock or stub binary" >&2
  exit 1
fi

for job in web-e2e session-compat-e2e mcp-e2e; do
  if ! awk "/^  ${job}:/{inside=1; next} inside && /^  [a-z0-9-]+:/{exit} inside" "$CI_WORKFLOW" \
    | grep -q 'DEV_E2E_RUNNERS_ONLY: "1"'; then
    echo "${job} must start only e2e-echo runners" >&2
    exit 1
  fi
done

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

ssh_fixture="$(mktemp -d)"
trap 'rm -rf "$ssh_fixture"' EXIT
mkdir -p "$ssh_fixture/source" "$ssh_fixture/home"
printf 'Host test\n' > "$ssh_fixture/source/config"
printf 'private-key\n' > "$ssh_fixture/source/id_ed25519"
HOME="$ssh_fixture/home" RUNNER_SSH_SOURCE_DIR="$ssh_fixture/source" \
  bash -c 'source "$1"; bootstrap_runner_ssh' _ "$SSH_BOOTSTRAP"

file_mode() {
  stat -c '%a' "$1" 2>/dev/null || stat -f '%Lp' "$1"
}

test -d "$ssh_fixture/home/.ssh"
test "$(file_mode "$ssh_fixture/home/.ssh")" = "700"
test "$(file_mode "$ssh_fixture/home/.ssh/config")" = "600"
test "$(file_mode "$ssh_fixture/home/.ssh/id_ed25519")" = "600"
