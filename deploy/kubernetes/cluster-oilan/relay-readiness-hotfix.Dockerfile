FROM repo.aiedulab.cn:8443/agentcloud/relay@sha256:4e5992c1702cfc467d578ae4ff693cdece606c534c62c1c25dbd373498d4022d
ARG RELAY_SHA
COPY --chown=1000:1000 relay-${RELAY_SHA} /app/relay
