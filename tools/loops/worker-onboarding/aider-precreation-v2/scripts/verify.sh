#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
evidence="$root/evidence/aider-precreation-2026-07-16.json"
review="$root/evidence/aider-precreation-review-2026-07-16.json"

jq -e '
  .worker_slug == "aider" and
  .definition.adapter_id == "aider-pty" and
  .definition.interaction_mode == "pty" and
  .definition.model_required == false and
  .runtime.image_digest == "sha256:c886a5d43f246cdf5aecd159106081e400d4f1c741acbe26788e4d5b4503b5a6" and
  .runtime.version == "0.86.2" and
  .runner.node_id == "dev-runner-aider" and
  .runner.status == "online" and
  .runner.tunnel_state == "connected" and
  .browser.template_apply.status == "applied" and
  .database.worker_launch_count == 0 and
  .database.matching_pod_count == 0 and
  .safety_boundary.provider_request == false and
  .safety_boundary.pod_create_rpc == false
' "$evidence" >/dev/null

jq -e '
  .worker_slug == "aider" and
  .verdict == "precreation_verified_not_supported" and
  .plaintext_credential_present == false and
  .provider_response_present == false
' "$review" >/dev/null

jq -e '
  .status == "blocked_human_gate" and
  .active_task_id == "verify-live-launch" and
  .last_verifier_exit_code == 0
' "$root/state.json" >/dev/null

grep -Fq '[x] `accept-contract`' "$root/ACCEPTANCE.md"
grep -Fq '[x] `accept-runtime`' "$root/ACCEPTANCE.md"
grep -Fq '[x] `accept-precreation`' "$root/ACCEPTANCE.md"
grep -Fq '[x] `accept-review`' "$root/ACCEPTANCE.md"
grep -Fq '[ ] `accept-live-launch`' "$root/ACCEPTANCE.md"
