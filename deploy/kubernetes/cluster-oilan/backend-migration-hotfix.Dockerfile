FROM repo.aiedulab.cn:8443/agentsmesh/backend@sha256:c160059c60e2b9d6b8d99b4919ca90ec6fae570d69787c49fce9a9aa498e7a42

ARG SERVER_SHA
COPY --chown=1000:1000 server-${SERVER_SHA} /app/server
