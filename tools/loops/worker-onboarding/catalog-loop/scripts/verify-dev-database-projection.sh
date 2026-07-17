#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "${LOOP_ROOT}/../../../.." && pwd)"
POSTGRES_CONTAINER="${WORKER_DEFINITION_POSTGRES_CONTAINER:-agentsmesh-main-postgres-1}"
SLUGS=(
  aider claude-code codex-cli cursor-cli do-agent gemini-cli grok-build
  hermes loopal minimax-cli openclaw opencode
)

for slug in "${SLUGS[@]}"; do
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
done
