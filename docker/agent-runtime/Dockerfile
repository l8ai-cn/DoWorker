# Do Worker agent runtime image: runner daemon + one pre-installed AI CLI.
#
#   bazel run //docker/agent-runtime:build_all
#   bazel run //docker/agent-runtime:build_claude_code
#
# Shared `base` stage is built once as do-worker/runner-base:latest so apt/npm
# layers are not re-run for every agent variant.

ARG DO_AGENT_STAGE=without-do-agent

FROM node:20-bookworm-slim AS base

ARG HTTP_PROXY=
ARG HTTPS_PROXY=
ARG http_proxy=
ARG https_proxy=
ENV HTTP_PROXY=${HTTP_PROXY} \
    HTTPS_PROXY=${HTTPS_PROXY} \
    http_proxy=${http_proxy} \
    https_proxy=${https_proxy} \
    NO_PROXY=*

RUN printf 'Acquire::http::Proxy "false";\nAcquire::https::Proxy "false";\n' \
      > /etc/apt/apt.conf.d/99noproxy \
    && set -ux; \
    packages="git openssh-client sudo ca-certificates wget openssl"; \
    ok=0; \
    for attempt in 1 2 3 4 5 6 7 8; do \
      if apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends $packages; then \
        rm -rf /var/lib/apt/lists/*; \
        ok=1; \
        break; \
      fi; \
      echo "apt install attempt ${attempt} failed, retrying..." >&2; \
      sleep 20; \
    done; \
    test "$ok" = 1

FROM base AS runtime

ARG AGENT_RUNTIME=e2e-echo
ENV AGENT_RUNTIME=${AGENT_RUNTIME}

RUN case "${AGENT_RUNTIME}" in \
      claude-code) npm install -g @anthropic-ai/claude-code ;; \
      codex-cli) npm install -g @openai/codex ;; \
      gemini-cli) npm install -g @google/gemini-cli ;; \
      opencode) npm install -g opencode-ai ;; \
      aider) \
        apt-get update \
        && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends python3-pip \
        && pip3 install --break-system-packages aider-chat \
        && rm -rf /var/lib/apt/lists/* ;; \
      e2e-echo|loopal|do-agent) true ;; \
      *) echo "unsupported AGENT_RUNTIME=${AGENT_RUNTIME}" >&2; exit 1 ;; \
    esac \
    && npm cache clean --force

RUN groupmod -n runner node \
    && usermod -l runner -g runner node \
    && usermod -d /home/runner -m runner \
    && mkdir -p /workspace /app /home/runner/.do-worker \
    && case "${AGENT_RUNTIME}" in \
         claude-code) mkdir -p /home/runner/.claude ;; \
         codex-cli) mkdir -p /home/runner/.codex ;; \
         gemini-cli) mkdir -p /home/runner/.gemini ;; \
         loopal) mkdir -p /home/runner/.loopal ;; \
         do-agent) mkdir -p /home/runner/.agent ;; \
         e2e-echo) true ;; \
       esac \
    && chown -R runner:runner /workspace /app /home/runner \
    && echo 'runner ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/runner

COPY --chmod=0755 runner-binary /usr/local/bin/do-worker-runner
RUN ln -sf do-worker-runner /usr/local/bin/agentsmesh-runner
COPY --chmod=0755 e2e-mock-agent-binary /usr/local/bin/e2e-mock-agent
COPY --chmod=0755 loopal-binary /usr/local/bin/loopal
COPY --chmod=0755 runner-entrypoint.sh /usr/local/bin/runner-entrypoint.sh

FROM runtime AS runtime-with-do-agent
COPY --chmod=0755 do-agent-binary /usr/local/bin/do-agent

FROM runtime AS runtime-without-do-agent

FROM runtime-${DO_AGENT_STAGE} AS final

RUN case "${AGENT_RUNTIME}" in \
      e2e-echo) rm -f /usr/local/bin/loopal /usr/local/bin/do-agent ;; \
      loopal) rm -f /usr/local/bin/e2e-mock-agent /usr/local/bin/do-agent ;; \
      do-agent) rm -f /usr/local/bin/e2e-mock-agent /usr/local/bin/loopal ;; \
      *) rm -f /usr/local/bin/e2e-mock-agent /usr/local/bin/loopal /usr/local/bin/do-agent ;; \
    esac

USER runner
WORKDIR /app
ENTRYPOINT ["/usr/local/bin/runner-entrypoint.sh"]
