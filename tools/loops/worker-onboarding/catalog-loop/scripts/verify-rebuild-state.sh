#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$LOOP_ROOT"

test -s evidence/revocations/2026-07-12-invalid-shared-contract.md
expected_runs="$(sort catalog/formal-worker-slugs.txt)"
actual_runs="$(
  find runs -mindepth 1 -maxdepth 1 -type d -exec basename {} \; | sort
)"
[[ "$actual_runs" == "$expected_runs" ]]
loop_id="$(jq -r '.metadata.id' loop.json)"
max_tokens="$(jq -r '.termination_policy.budget_exits.max_tokens' loop.json)"
status="$(jq -r '.status' state.json)"
active_task="$(jq -r '.active_task_id // empty' state.json)"

jq -e '
  .metadata.id == "worker-integration-evidence-rebuild" and
  .metadata.name == "Worker Integration Evidence Rebuild" and
  .loop_nodes[0].id == "rebuild-goal" and
  (.termination_policy.budget_exits.max_iterations | type == "number" and . > 0) and
  (.termination_policy.budget_exits.max_tokens | type == "number" and . > 0)
' loop.json >/dev/null
jq -e --arg loop_id "$loop_id" --argjson max_tokens "$max_tokens" '
  .loop_id == $loop_id and
  .loop_name == "Worker Integration Evidence Rebuild" and
  (
    (.status == "running" and
      (.active_loop_node_id | type == "string" and length > 0) and
      (.active_task_id | type == "string" and length > 0) and
      .terminal_reason == null and
      .token_estimate < $max_tokens
    ) or
    (.status == "budget_exhausted" and
      .active_loop_node_id == null and
      .active_task_id == null and
      (.terminal_reason | startswith("token_budget_exhausted:")) and
      .token_estimate >= $max_tokens
    )
  )
' state.json >/dev/null
[[ "$(jq -cS '.atomic_tasks' loop.json)" == "$(jq -cS '.' tasks.json)" ]]
[[ "$(jq -cS '.loop_nodes' loop.json)" == "$(jq -cS '.' loops.json)" ]]
jq -e --arg loop_id "$loop_id" '
  .loop_id == $loop_id and
  (.drift_checks | index("state loop id matches the canonical manifest")) and
  (.drift_checks | index("active task maps to a declared atomic task"))
' monitoring-plan.json >/dev/null
if [[ "$status" == "running" ]]; then
  jq -e --arg active_task "$active_task" 'any(.[]; .id == $active_task)' \
    tasks.json >/dev/null
fi
jq -e '
  (.last_verifier_id | type == "string" and length > 0)
' state.json >/dev/null
jq -e '
  .schema_version == 1 and
  (.workers | length == 14) and
  ([.workers[] | select(.support_status == "not_supported")] | length == 13) and
  ([.workers[] | select(
    .support_status == "verified_local_dev" and
    .product_path == "created_acp_prompt_and_cleanup_verified" and
    .browser == "created_acp_prompt_and_cleanup_verified" and
    .runner == "create_and_acp_prompt_verified" and
    .repeatable_e2e == "passed"
  )] | length == 1) and
  ([.workers[] | select(
    .product_path == "blocked_missing_model_resource_verified" and
    .browser == "missing_model_resource_guard_verified"
  )] | length == 2) and
  ([.workers[] | select(
    .product_path == "unverified" and
    .browser == "missing_runtime_image_guard_verified"
  )] | length == 8) and
  ([.workers[] | select(
    .product_path == "unverified" and
    .browser == "unverified"
  )] | length == 3)
' catalog/worker-evidence-matrix.json >/dev/null
grep -Fq 'Worker Integration Evidence Rebuild Plan' PROGRESS.md
