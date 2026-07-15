# Do Worker agent runtime image: runner daemon + one pre-installed AI CLI.
#
#   bash docker/agent-runtime/build.sh all
#   bash docker/agent-runtime/build.sh claude-code
#
# Shared `base` stage is built once as do-worker/runner-base:latest so apt/npm
# layers are not re-run for every agent variant.

FROM node:24-bookworm-slim AS base

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
ARG HERMES_AGENT_VERSION=0.18.2
ARG REMOTION_VERSION=4.0.489
ENV AGENT_RUNTIME=${AGENT_RUNTIME}

RUN install_apt() { \
      attempt=1; \
      while [ "$attempt" -le 3 ]; do \
        if apt-get update \
          && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends --fix-missing "$@"; then \
          return 0; \
        fi; \
        echo "runtime apt install attempt ${attempt}/3 failed" >&2; \
        attempt=$((attempt + 1)); \
        sleep 10; \
      done; \
      return 1; \
    }; \
    case "${AGENT_RUNTIME}" in \
      claude-code) npm install -g @anthropic-ai/claude-code ;; \
      codex-cli) npm install -g @openai/codex ;; \
      video-studio) \
        install_apt chromium ffmpeg fontconfig fonts-noto-cjk python3 python3-pip \
        && npm install -g @openai/codex \
        && mkdir -p /opt/video-studio \
        && npm install --prefix /opt/video-studio --omit=dev \
          "@remotion/cli@${REMOTION_VERSION}" \
          "remotion@${REMOTION_VERSION}" \
          react@19.2.4 \
          react-dom@19.2.4 \
        && ln -s "$(command -v codex)" /usr/local/bin/video-studio-codex \
        && ln -s /opt/video-studio/node_modules/.bin/remotion /usr/local/bin/remotion \
        && rm -rf /var/lib/apt/lists/* ;; \
      gemini-cli) npm install -g @google/gemini-cli ;; \
      opencode) npm install -g opencode-ai ;; \
      grok-build) npm install -g @xai-official/grok ;; \
      openclaw) npm install -g openclaw ;; \
      hermes) \
        install_apt python3-pip \
        && PIP_BREAK_SYSTEM_PACKAGES=1 npm install -g "hermes-agent@${HERMES_AGENT_VERSION}" \
        && rm -rf /var/lib/apt/lists/* ;; \
      aider) \
        install_apt python3-pip \
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
         codex-cli|video-studio) mkdir -p /home/runner/.codex ;; \
         gemini-cli) mkdir -p /home/runner/.gemini ;; \
         loopal) mkdir -p /home/runner/.loopal ;; \
         do-agent) mkdir -p /home/runner/.agent ;; \
         grok-build) mkdir -p /home/runner/.grok ;; \
         openclaw) mkdir -p /home/runner/.openclaw ;; \
         hermes) mkdir -p /home/runner/.hermes ;; \
         e2e-echo) true ;; \
       esac \
    && chown -R runner:runner /workspace /app /home/runner \
    && echo 'runner ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/runner

COPY --chmod=0755 . /tmp/agent-runtime-context/

# ARG from a prior stage is not inherited; without this redeclare
# AGENT_RUNTIME is empty and the *) branch deletes e2e-mock-agent.
ARG AGENT_RUNTIME=e2e-echo

RUN install -m 0755 /tmp/agent-runtime-context/runner-binary /usr/local/bin/do-worker-runner \
    && ln -sf do-worker-runner /usr/local/bin/agentsmesh-runner \
    && install -m 0755 /tmp/agent-runtime-context/runner-entrypoint.sh /usr/local/bin/runner-entrypoint.sh \
    && case "${AGENT_RUNTIME}" in \
         e2e-echo) install -m 0755 /tmp/agent-runtime-context/e2e-mock-agent-binary /usr/local/bin/e2e-mock-agent ;; \
         loopal) install -m 0755 /tmp/agent-runtime-context/loopal-binary /usr/local/bin/loopal ;; \
         do-agent) install -m 0755 /tmp/agent-runtime-context/do-agent-binary /usr/local/bin/do-agent ;; \
         *) true ;; \
       esac \
    && rm -rf /tmp/agent-runtime-context

ENV CHROME_BIN=/usr/bin/chromium \
    NODE_PATH=/opt/video-studio/node_modules

USER runner
WORKDIR /app
ENTRYPOINT ["/usr/local/bin/runner-entrypoint.sh"]
