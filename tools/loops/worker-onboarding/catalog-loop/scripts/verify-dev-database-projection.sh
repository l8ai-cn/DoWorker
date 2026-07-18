#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "${LOOP_ROOT}/../../../.." && pwd)"
POSTGRES_CONTAINER="${WORKER_DEFINITION_POSTGRES_CONTAINER:-}"
WORKER_CATALOG="${REPO_ROOT}/config/worker-types/catalog.json"

if [[ -z "$POSTGRES_CONTAINER" ]]; then
  project_name="$(awk -F= '/^COMPOSE_PROJECT_NAME=/{print $2; exit}' \
    "${REPO_ROOT}/deploy/dev/.env")"
  [[ -n "$project_name" ]] || {
    echo "COMPOSE_PROJECT_NAME is required in deploy/dev/.env" >&2
    exit 1
  }
  POSTGRES_CONTAINER="${project_name}-postgres-1"
fi

while IFS= read -r slug; do
  definition_dir="${REPO_ROOT}/config/worker-types/${slug}"
  count="$(docker exec "$POSTGRES_CONTAINER" psql -U agentsmesh -d agentsmesh -Atc \
    "SELECT count(*) FROM agents WHERE slug = '${slug}'")"
  [[ "$count" == "1" ]] || {
    echo "expected one database registration for ${slug}, got ${count}" >&2
    exit 1
  }

  expected_executable="$(jq -r '.executable' "${definition_dir}/definition.json")"
  expected_adapter_id="$(jq -r '.adapter_id' "${definition_dir}/definition.json")"
  expected_modes="$(jq -r '.interaction_modes | join(",")' "${definition_dir}/definition.json")"
  read -r executable adapter_id supported_modes is_builtin is_active is_internal uses_legacy_columns \
    < <(docker exec "$POSTGRES_CONTAINER" psql -U agentsmesh -d agentsmesh -At -F $'\t' -c \
      "SELECT executable, adapter_id, supported_modes, is_builtin, is_active, is_internal, uses_legacy_columns
       FROM agents WHERE slug = '${slug}'")

  [[ "$executable" == "$expected_executable" ]] || {
    echo "${slug}: executable ${executable} does not match ${expected_executable}" >&2
    exit 1
  }
  [[ "$adapter_id" == "$expected_adapter_id" ]] || {
    echo "${slug}: adapter_id ${adapter_id} does not match ${expected_adapter_id}" >&2
    exit 1
  }
  [[ "$supported_modes" == "$expected_modes" ]] || {
    echo "${slug}: supported_modes ${supported_modes} does not match ${expected_modes}" >&2
    exit 1
  }
  [[ "$is_builtin" == "t" && "$is_active" == "t" && "$is_internal" == "f" && "$uses_legacy_columns" == "f" ]] || {
    echo "${slug}: projection flags are not selectable builtin defaults" >&2
    exit 1
  }

  diff -u \
    <(docker exec "$POSTGRES_CONTAINER" psql -U agentsmesh -d agentsmesh -Atc \
      "SELECT agentfile_source FROM agents WHERE slug = '${slug}'" |
      perl -0pe 's/\n\z//') \
    "${definition_dir}/AgentFile"
done < <(jq -r '.worker_types[].slug' "$WORKER_CATALOG")
