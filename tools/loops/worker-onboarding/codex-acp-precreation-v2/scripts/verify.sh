#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
evidence="$root/evidence/codex-acp-precreation-2026-07-16.json"
review="$root/evidence/codex-acp-precreation-review-2026-07-16.json"

jq -e '
  .worker_slug == "codex-cli" and
  .definition.adapter_id == "codex-app-server" and
  .definition.interaction_mode == "acp" and
  .definition.model_required == true and
  .runtime.image_digest == "sha256:3e1a97a92f722bd362795aa7538a4c74844972821f8964b63c14cb6c0032871b" and
  .runtime.version == "0.144.5" and
  .runner.node_id == "dev-runner-codex" and
  .runner.status == "online" and
  .runner.tunnel_state == "connected" and
  .model_binding.protocol_adapter == "openai-compatible" and
  .browser.template_apply.status == "applied" and
  .database.worker_launch_count == 0 and
  .database.matching_pod_count == 0 and
  .safety_boundary.provider_request == false and
  .safety_boundary.pod_create_rpc == false
' "$evidence" >/dev/null

jq -e '
  .worker_slug == "codex-cli" and
  .verdict == "precreation_verified_not_supported" and
  .plaintext_credential_present == false and
  .provider_response_present == false
' "$review" >/dev/null

jq -e '.status == "blocked_human_gate" and .active_task_id == "verify-live-launch" and .last_verifier_exit_code == 0' \
  "$root/state.json" >/dev/null
grep -Fq '[ ] `accept-live-launch`' "$root/ACCEPTANCE.md"
