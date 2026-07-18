# Do Worker agent runtime image: runner daemon + one pre-installed AI CLI.
#
#   bash docker/agent-runtime/build.sh all
#   bash docker/agent-runtime/build.sh claude-code
#
# Shared `base` stage is built once as do-worker/runner-base:latest so apt/npm
# layers are not re-run for every agent variant.

ARG NODE_BASE_IMAGE=node:24-bookworm-slim
ARG PYTHON_BASE_IMAGE=python:3.11-slim-bookworm
ARG RUNTIME_EXTENSION_BASE=do-worker/runner-base:latest
ARG RUNTIME_EXTENSION_RUNTIME_BASE=runtime-extension-base
ARG RUNTIME_SHARED_BASE=base
ARG RUNTIME_BASE=runtime-shared-base
ARG OPENCLAW_VERSION=2026.6.11
ARG OPENCLAW_NODE_VERSION=24.18.0

FROM ${NODE_BASE_IMAGE} AS base

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
    packages="git openssh-client sudo ca-certificates curl wget openssl"; \
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

FROM ${PYTHON_BASE_IMAGE} AS python-runtime

FROM ${RUNTIME_SHARED_BASE} AS runtime-shared-base

FROM runtime-shared-base AS python-runtime-base
COPY --from=python-runtime /usr/local /usr/local
COPY --from=python-runtime /usr/lib/x86_64-linux-gnu/libsqlite3.so.0.8.6 /usr/lib/x86_64-linux-gnu/
COPY --from=python-runtime /usr/lib/x86_64-linux-gnu/libsqlite3.so.0 /usr/lib/x86_64-linux-gnu/

FROM ${RUNTIME_BASE} AS runtime

ARG AGENT_RUNTIME=e2e-echo
ARG HERMES_AGENT_VERSION=0.18.2
ARG MINIMAX_CLI_VERSION=1.0.16
ARG OPENCLAW_VERSION=2026.6.11
ENV AGENT_RUNTIME=${AGENT_RUNTIME}

RUN case "${AGENT_RUNTIME}" in \
      claude-code) npm install -g @anthropic-ai/claude-code ;; \
      codex-cli) npm install -g @openai/codex ;; \
      cursor-cli) \
        curl --fail --show-error --silent --location https://cursor.com/install -o /tmp/cursor-install.sh \
        && bash -c 'curl() { command curl --retry 5 --retry-delay 2 --retry-all-errors "$@"; }; export -f curl; bash /tmp/cursor-install.sh' \
        && install -d -m 0755 /opt/cursor-agent \
        && cp -a /root/.local/share/cursor-agent/. /opt/cursor-agent/ \
        && cursor_agent="$(find /opt/cursor-agent/versions -mindepth 2 -maxdepth 2 -type f -name cursor-agent -print -quit)" \
        && test -n "${cursor_agent}" \
        && ln -sf "${cursor_agent}" /usr/local/bin/agent \
        && rm -rf /root/.local /tmp/cursor-install.sh ;; \
      gemini-cli) npm install -g @google/gemini-cli ;; \
      opencode) npm install -g opencode-ai ;; \
      grok-build) npm install -g @xai-official/grok ;; \
      minimax-cli) npm install -g "mmx-cli@${MINIMAX_CLI_VERSION}" ;; \
      openclaw) npm install -g "openclaw@${OPENCLAW_VERSION}" ;; \
      hermes) \
        npm install -g "hermes-agent@${HERMES_AGENT_VERSION}" ;; \
      aider) \
        pip3 install --no-cache-dir aider-chat ;; \
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
         cursor-cli) mkdir -p /home/runner/.cursor ;; \
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

COPY --chmod=0755 runner-binary /usr/local/bin/do-worker-runner
RUN ln -sf do-worker-runner /usr/local/bin/agentsmesh-runner
COPY --chmod=0755 binaries/ /usr/local/lib/do-worker/
COPY --chmod=0755 runner-entrypoint.sh /usr/local/bin/runner-entrypoint.sh
COPY --chmod=0755 minimax-cli-wrapper.sh /usr/local/bin/minimax-cli-wrapper.sh

ARG AGENT_RUNTIME=e2e-echo

RUN case "${AGENT_RUNTIME}" in \
      e2e-echo) install -m 0755 /usr/local/lib/do-worker/e2e-mock-agent-binary /usr/local/bin/e2e-mock-agent ;; \
      loopal) install -m 0755 /usr/local/lib/do-worker/loopal-binary /usr/local/bin/loopal ;; \
      do-agent) install -m 0755 /usr/local/lib/do-worker/do-agent-binary /usr/local/bin/do-agent ;; \
      minimax-cli) rm -f /usr/local/bin/mmx && install -m 0755 /usr/local/bin/minimax-cli-wrapper.sh /usr/local/bin/mmx ;; \
      *) true ;; \
    esac \
    && rm -f /usr/local/bin/minimax-cli-wrapper.sh \
    && rm -rf /usr/local/lib/do-worker

USER runner
WORKDIR /app
ENTRYPOINT ["/usr/local/bin/runner-entrypoint.sh"]

FROM ${RUNTIME_EXTENSION_BASE} AS runtime-extension-base

USER root

FROM runtime-extension-base AS runtime-extension-python

USER root
COPY --from=python-runtime /usr/local /usr/local
COPY --from=python-runtime /usr/lib/x86_64-linux-gnu/libsqlite3.so.0.8.6 /usr/lib/x86_64-linux-gnu/
COPY --from=python-runtime /usr/lib/x86_64-linux-gnu/libsqlite3.so.0 /usr/lib/x86_64-linux-gnu/

FROM ${RUNTIME_EXTENSION_RUNTIME_BASE} AS runtime-extension

USER root

ARG AGENT_RUNTIME
ARG HERMES_AGENT_VERSION=0.18.2
ARG OPENCLAW_VERSION=2026.6.11
ARG OPENCLAW_NODE_VERSION=24.18.0
ENV AGENT_RUNTIME=${AGENT_RUNTIME}

RUN case "${AGENT_RUNTIME}" in \
      openclaw) \
        npm install -g --omit=dev --force "node@${OPENCLAW_NODE_VERSION}" \
        && npm install -g --omit=dev "openclaw@${OPENCLAW_VERSION}" \
        && ln -sf /usr/local/lib/node_modules/node/lib/node_modules/openclaw/openclaw.mjs /usr/local/bin/openclaw ;; \
      hermes) \
        npm install -g --omit=dev "hermes-agent@${HERMES_AGENT_VERSION}" ;; \
      *) echo "runtime-extension only supports AGENT_RUNTIME=openclaw|hermes" >&2; exit 1 ;; \
    esac \
    && npm cache clean --force \
    && mkdir -p "/home/runner/.${AGENT_RUNTIME}" \
    && chown -R runner:runner "/home/runner/.${AGENT_RUNTIME}"

USER runner
WORKDIR /app
