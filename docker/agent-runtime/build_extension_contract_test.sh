#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

if ! grep -Fq 'FROM ${RUNTIME_EXTENSION_BASE} AS runtime-extension-base' Dockerfile; then
    echo "Dockerfile must define the prebuilt runtime extension stage" >&2
    exit 1
fi

if ! grep -Fq 'FROM ${RUNTIME_EXTENSION_RUNTIME_BASE} AS runtime-extension' Dockerfile; then
    echo "Dockerfile must select an explicit extension runtime base" >&2
    exit 1
fi

if ! grep -Fq 'RUNTIME_EXTENSION_BASE is only supported for openclaw and hermes' build.sh; then
    echo "build script must reject unsupported extension runtimes" >&2
    exit 1
fi

if ! grep -Fq 'runtime-extension' build.sh; then
    echo "build script must select the extension stage when requested" >&2
    exit 1
fi

if ! grep -Fq 'RUNTIME_EXTENSION_RUNTIME_BASE=runtime-extension-python' build.sh; then
    echo "Hermes extension must select the Python runtime base" >&2
    exit 1
fi

if ! grep -Fq 'build_cmd+=(--target runtime)' build.sh; then
    echo "standard builds must select the runtime stage" >&2
    exit 1
fi

if grep -Fq -- '--cache-from "$BASE_IMAGE"' build.sh; then
    echo "standard builds must not treat a local base image as a registry cache" >&2
    exit 1
fi

if ! grep -Fq 'ARG OPENCLAW_VERSION=2026.6.11' Dockerfile; then
    echo "OpenClaw must pin its verified package version" >&2
    exit 1
fi

if ! grep -Fq 'ARG OPENCLAW_NODE_VERSION=24.18.0' Dockerfile; then
    echo "OpenClaw extension must pin the required Node version" >&2
    exit 1
fi

if ! grep -Fq 'node@${OPENCLAW_NODE_VERSION}' Dockerfile; then
    echo "OpenClaw extension must install its required Node runtime" >&2
    exit 1
fi

if ! grep -Fq 'openclaw@${OPENCLAW_VERSION}' Dockerfile; then
    echo "OpenClaw must install the verified package version" >&2
    exit 1
fi

if ! grep -Fq '/usr/local/bin/openclaw' Dockerfile; then
    echo "OpenClaw extension must link its CLI into PATH" >&2
    exit 1
fi

if ! grep -Fq 'OPENCLAW_VERSION=${OPENCLAW_VERSION:-2026.6.11}' build.sh; then
    echo "build script must pass the OpenClaw package version" >&2
    exit 1
fi

if ! grep -Fq 'FROM runtime-extension-base AS runtime-extension-python' Dockerfile; then
    echo "Hermes extension must define a dedicated Python runtime base" >&2
    exit 1
fi

if ! grep -Fq 'COPY --from=python-runtime /usr/local /usr/local' Dockerfile; then
    echo "Hermes extension must provide Python for hermes-agent postinstall" >&2
    exit 1
fi

if ! grep -Fq 'libsqlite3.so.0.8.6' Dockerfile; then
    echo "Hermes extension must provide Python sqlite runtime dependencies" >&2
    exit 1
fi
