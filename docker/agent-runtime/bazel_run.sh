#!/usr/bin/env bash
# Bazel entrypoint for agent-runtime image builds.
# Resolves paths in the live source tree (not runfiles).
set -euo pipefail

if [[ -z "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then
  echo "ERROR: invoke via 'bazel run //docker/agent-runtime:...', not direct exec." >&2
  exit 1
fi

exec "${BUILD_WORKSPACE_DIRECTORY}/docker/agent-runtime/build.sh" "$@"
