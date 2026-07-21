#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"
loop_dir="$repo_root/tools/loops/worker-onboarding/definition-driven-contract"
evidence="$loop_dir/evidence/current-contract-baseline.json"
env_file="$repo_root/deploy/dev/.env"

jq empty "$evidence"
jq -e '
  .schema_version == 1 and
  .formal_worker_count == 13 and
  .online_worker_slug_count == 12 and
  .formal_support_state == "none" and
  .config_document_contract.current_wire_field == "config_bundle_ids" and
  .config_document_contract.named_bindings_present == false and
  .config_document_contract.agentfile_selection == "last_selected_bundle_wins"
' "$evidence" >/dev/null

docker info --format 'server={{.ServerVersion}}' >/dev/null
project_name="$(awk -F= '$1 == "COMPOSE_PROJECT_NAME" {print $2}' "$env_file")"
backend_port="$(awk -F= '$1 == "BACKEND_HTTP_PORT" {print $2}' "$env_file")"
curl --fail --silent --show-error "http://127.0.0.1:${backend_port}/health" >/dev/null

catalog_path="$repo_root/deploy/dev/runtime/worker-runtime-catalog.local.json"
image_count="$(jq '[.images[]] | length' "$catalog_path")"
catalog_slug_count="$(jq -r '[.images[] | .worker_type_slugs[]] | unique | length' "$catalog_path")"
catalog_slugs="$(jq -r '[.images[] | .worker_type_slugs[]] | unique | .[]' "$catalog_path")"

test "$image_count" = "11"
test "$catalog_slug_count" = "12"
runner_slugs="$(
  docker exec "${project_name}-postgres-1" psql -U agentcloud -d agentcloud -Atc "
    SELECT DISTINCT slug
    FROM runners
    CROSS JOIN LATERAL jsonb_array_elements_text(available_agents) AS slug
    WHERE status = 'online' AND node_id LIKE 'dev-runner-%'
    ORDER BY slug;
  "
)"
[[ "$runner_slugs" == "$catalog_slugs" ]]

test -f "$repo_root/docs/superpowers/plans/2026-07-16-worker-definition-driven-create-contract.md"
test "$(
  jq -s '[.[] | select((.config_documents | length) > 1)] | length' \
    "$repo_root"/config/worker-types/*/definition.json
)" = "0"

bash "$repo_root/tools/loops/worker-onboarding/catalog-loop/scripts/verify-definition-chain.sh"
bash "$repo_root/tools/loops/worker-onboarding/catalog-loop/scripts/verify-worker-documentation.sh"
bash "$repo_root/tools/loops/worker-onboarding/catalog-loop/scripts/verify-rebuild-state.sh"
