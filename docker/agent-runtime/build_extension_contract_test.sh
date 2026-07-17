#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

if ! grep -Fq 'FROM ${RUNTIME_EXTENSION_BASE} AS runtime-extension' Dockerfile; then
    echo "Dockerfile must define the prebuilt runtime extension stage" >&2
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

if ! awk '
  /FROM \$\{RUNTIME_EXTENSION_BASE\} AS runtime-extension/ { extension=1 }
  extension && /for attempt in 1 2 3 4 5 6 7 8/ { found=1 }
  END { exit found ? 0 : 1 }
' Dockerfile; then
    echo "Hermes extension install must use bounded Python dependency retries" >&2
    exit 1
fi

if ! awk '
  /FROM \$\{RUNTIME_EXTENSION_BASE\} AS runtime-extension/ { extension=1 }
  extension && /python3-pip/ { found=1 }
  END { exit found ? 0 : 1 }
' Dockerfile; then
    echo "Hermes extension must install Python for hermes-agent postinstall" >&2
    exit 1
fi

if ! awk '
  /FROM \$\{RUNTIME_EXTENSION_BASE\} AS runtime-extension/ { extension=1 }
  extension && /PIP_BREAK_SYSTEM_PACKAGES=1 npm install/ { found=1 }
  END { exit found ? 0 : 1 }
' Dockerfile; then
    echo "Hermes extension must allow its postinstall to install the Python bridge" >&2
    exit 1
fi
