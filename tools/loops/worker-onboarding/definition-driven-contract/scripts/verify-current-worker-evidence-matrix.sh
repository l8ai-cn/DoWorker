#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(git -C "$LOOP_ROOT" rev-parse --show-toplevel)"
MATRIX="${LOOP_ROOT}/evidence/current-worker-evidence-matrix.json"
TEMPORARY="$(mktemp)"
trap 'rm -f "$TEMPORARY"' EXIT

node "${LOOP_ROOT}/scripts/build-current-worker-evidence-matrix.mjs" \
  --output "$TEMPORARY"
[[ "$(jq -cS . "$MATRIX")" == "$(jq -cS . "$TEMPORARY")" ]]

expected_slugs="$(jq -c '[.worker_types[].slug] | sort' \
  "${REPO_ROOT}/config/worker-types/catalog.json")"
jq -e --argjson expected_slugs "$expected_slugs" '
  .schema_version == 1 and
  .formal_support_status == "none" and
  .worker_count == ($expected_slugs | length) and
  ([.workers[].slug] | sort) == $expected_slugs and
  all(.workers[];
    .formal_support_status == "not_verified" and
    (.runtime.image_probe_status | IN("passed", "image_missing", "probe_failed")) and
    (.runtime.create_option_selectable | type == "boolean") and
    (.runtime.online_runners | type == "array") and
    (.integration_gates.credential == "not_verified") and
    (.integration_gates.credential_reference_api |
      IN("definition_projection_passed", "model_resource_targets_hidden", "not_required")) and
    (.integration_gates.config_document_api |
      IN("not_projected_by_current_wire_contract", "not_required")) and
    (.integration_gates.rust_core == "not_verified") and
    (.integration_gates.web == "browser_not_verified") and
    (.integration_gates.lifecycle | IN("not_run", "failed", "passed"))
  ) and
  (.workers[] | select(.slug == "loopal") |
    .runtime.evidence_state == "blocked" and
    .runtime.create_option_selectable == false)
' "$MATRIX" >/dev/null
