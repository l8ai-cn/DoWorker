#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
result="$root/evidence/aider-live-launch-approval-and-result.json"

jq -e '
  .worker_slug == "aider" and
  .approval.named_non_production_scope == true and
  .credential_reference_injection == "verified" and
  .pod_lifecycle == "verified" and
  .pty_session == "verified" and
  .provider_smoke == "verified" and
  .cleanup == "verified"
' "$result" >/dev/null
