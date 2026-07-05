# AgentsMesh Runner — dev image
#
# Base: node:20-bookworm-slim ships Node 20 + npm, avoiding NodeSource/apt churn.
# Runner binary: bazel-built linux/amd64 copied to deploy/dev/runner-binary.
FROM node:20-bookworm-slim

# Docker Desktop injects a broken HTTP proxy (502). Disable for apt/npm inside build.
ARG AGENT_RUNTIME=e2e-echo
ARG HTTP_PROXY=
ARG HTTPS_PROXY=
ARG http_proxy=
ARG https_proxy=
ENV AGENT_RUNTIME=${AGENT_RUNTIME} \
    HTTP_PROXY=${HTTP_PROXY} \
    HTTPS_PROXY=${HTTPS_PROXY} \
    http_proxy=${http_proxy} \
    https_proxy=${https_proxy} \
    NO_PROXY=*

RUN printf 'Acquire::http::Proxy "false";\nAcquire::https::Proxy "false";\n' \
      > /etc/apt/apt.conf.d/99noproxy \
    && set -ux; \
    packages="git openssh-client sudo ca-certificates wget openssl"; \
    ok=0; \
    for attempt in 1 2 3 4 5; do \
      if apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends $packages; then \
        rm -rf /var/lib/apt/lists/*; \
        ok=1; \
        break; \
      fi; \
      echo "apt install attempt ${attempt} failed, retrying..." >&2; \
      sleep 15; \
    done; \
    test "$ok" = 1

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
COPY --chmod=0755 do-agent-binary /usr/local/bin/do-agent

RUN case "${AGENT_RUNTIME}" in \
      e2e-echo) rm -f /usr/local/bin/loopal /usr/local/bin/do-agent ;; \
      loopal) rm -f /usr/local/bin/e2e-mock-agent /usr/local/bin/do-agent ;; \
      do-agent) rm -f /usr/local/bin/e2e-mock-agent /usr/local/bin/loopal ;; \
      *) rm -f /usr/local/bin/e2e-mock-agent /usr/local/bin/loopal /usr/local/bin/do-agent ;; \
    esac

USER runner
WORKDIR /app
ENTRYPOINT ["/usr/local/bin/runner-entrypoint.sh"]
