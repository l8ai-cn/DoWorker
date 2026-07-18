#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
result="$root/evidence/codex-live-launch-approval-and-result.json"

jq -e '
  .worker_slug == "codex-cli" and
  .approval.named_non_production_scope == true and
  .model_resource_credential_use == "verified" and
  .pod_lifecycle == "verified" and
  .acp_session == "verified" and
  .provider_smoke == "verified" and
  .cleanup == "verified"
' "$result" >/dev/null
