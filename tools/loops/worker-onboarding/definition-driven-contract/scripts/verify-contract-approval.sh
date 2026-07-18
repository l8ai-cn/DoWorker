#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"
evidence="$repo_root/tools/loops/worker-onboarding/definition-driven-contract/evidence/named-binding-approval.json"

jq -e '
  .schema_version == 1 and
  .status == "approved" and
  .selection == "named_definition_document_bindings" and
  (.approved_by | type == "string" and length > 0) and
  .legacy_positional_read == false
' "$evidence" >/dev/null
