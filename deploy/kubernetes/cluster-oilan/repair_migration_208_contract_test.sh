#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${DIR}/../../.." && pwd)"
SQL="${DIR}/24-repair-migration-208.sql"
PRECONDITIONS="${DIR}/24-repair-migration-208-preconditions.sql"
JOB="${DIR}/25-repair-migration-208-job.yaml"
DEPLOY="${DIR}/repair-migration-208.sh"
HOTFIX_BUILD="${DIR}/build-backend-migration-hotfix.sh"

extract_migration() {
  local version="$1"
  sed -n \
    "/^-- BEGIN ${version}\$/,/^-- END ${version}\$/p" "${SQL}" |
    sed '1d;$d'
}

diff -u \
  "${ROOT}/backend/migrations/000207_add_agent_adapter_id.up.sql" \
  <(extract_migration 000207)
diff -u \
  "${ROOT}/backend/migrations/000208_upgrade_cursor_cli_agent.up.sql" \
  <(extract_migration 000208)

grep -Fq '\ir repair-preconditions.sql' "${SQL}"
grep -Fq "migration_version NOT IN (208, 222)" "${PRECONDITIONS}"
grep -Fq "dirty migration 222 is not in the expected rolled-back state" \
  "${PRECONDITIONS}"
grep -Fq "migration 207 has no adapter mapping" "${PRECONDITIONS}"
grep -Fq "67b424efb3a5b844df0184388b3cf822" "${PRECONDITIONS}"
grep -Fq "8248044a68445136126905472f6fbc02" "${PRECONDITIONS}"
grep -Fq "faba46db825a34e91fe33398ba447ccd" "${PRECONDITIONS}"
grep -Fq "runner cluster mapping is incorrect" "${PRECONDITIONS}"
grep -Fq "pod cluster mapping is incorrect" "${PRECONDITIONS}"
grep -Fq "registration token mapping is incorrect" "${PRECONDITIONS}"
grep -Fq "pending auth mapping is incorrect" "${PRECONDITIONS}"
grep -Fq "6945d5c8ae2c98789d3768284673ec6d" "${SQL}"
grep -Fq "/app/server migrate force 208" "${JOB}"
grep -Fq "/app/server migrate force 221" "${JOB}"
grep -Fq "/app/server migrate up" "${JOB}"
grep -Fq 'test "$version" = "222 (dirty=false)"' "${JOB}"
grep -Fq "backoffLimit: 0" "${JOB}"
if grep -Fq "ttlSecondsAfterFinished" "${JOB}"; then
  echo "failed repair evidence must not expire automatically" >&2
  exit 1
fi
grep -Eq 'pgvector@sha256:[a-f0-9]{64}' "${JOB}"
grep -Eq 'backend@sha256:[a-f0-9]{64}' "${JOB}"
grep -Fq 'EXPECTED_GO_VERSION="go1.26.2"' "${HOTFIX_BUILD}"
grep -Fq 'EXPECTED_SERVER_SHA="ad4f5c8df61f08b98fee8a59732b3e10b73011f1f768acec047b46ca94f76d91"' \
  "${HOTFIX_BUILD}"
grep -Fq 'EXPECTED_IMAGE_DIGEST="sha256:fa58ff8756f5052ee48026f6fd20500e49ac0b464e655c24d34fc23fdba972e6"' \
  "${HOTFIX_BUILD}"
grep -Fq 'migration-208-${RUN_ID}' "${DEPLOY}"
grep -Fq "kubectl create -f -" "${DEPLOY}"
grep -Fq "create configmap migration-208-repair-lock" "${DEPLOY}"
grep -Fq "delete configmap migration-208-repair-lock" "${DEPLOY}"
grep -Fq 'test "$(cat /repair-lock/run-id)" = "__RUN_ID__"' "${JOB}"

if awk '
  $1 == "image:" && $2 !~ /@sha256:[a-f0-9]{64}$/ { invalid = 1 }
  END { exit invalid }
' "${JOB}"; then
  :
else
  echo "repair job images must be pinned by digest" >&2
  exit 1
fi

if grep -Fq "25-repair-migration-208-job.yaml" "${DIR}/kustomization.yaml"; then
  echo "one-time repair job must not be a persistent kustomize resource" >&2
  exit 1
fi
