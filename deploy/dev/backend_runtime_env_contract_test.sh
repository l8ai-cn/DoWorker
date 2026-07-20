#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

grep -Fq 'load_env_file "$DEV_ENV" || true' run-backend-with-runtime-env.sh
grep -Fq 'load_env_file "$RUNTIME_ENV"' run-backend-with-runtime-env.sh
grep -Fq 'KB_GITEA_REPOSITORY_BASE_URLS=http://gitea:3000' lib/config_gen.sh
