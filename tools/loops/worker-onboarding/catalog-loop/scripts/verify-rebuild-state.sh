#!/usr/bin/env bash
set -euo pipefail

LOOP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(git -C "${LOOP_ROOT}" rev-parse --show-toplevel)"
WORKER_CATALOG="${REPO_ROOT}/config/worker-types/catalog.json"
cd "$LOOP_ROOT"

test -s evidence/revocations/2026-07-12-invalid-shared-contract.md
loop_id="$(jq -r '.metadata.id' loop.json)"
max_tokens="$(jq -r '.termination_policy.budget_exits.max_tokens' loop.json)"
max_iterations="$(jq -r '.termination_policy.budget_exits.max_iterations' loop.json)"
status="$(jq -r '.status' state.json)"
active_task="$(jq -r '.active_task_id // empty' state.json)"

jq -e '
  .metadata.id == "worker-integration-evidence-rebuild" and
  .metadata.name == "Worker Integration Evidence Rebuild" and
  .loop_nodes[0].id == "rebuild-goal" and
  (.termination_policy.budget_exits.max_iterations | type == "number" and . > 0) and
  (.termination_policy.budget_exits.max_tokens | type == "number" and . > 0)
' loop.json >/dev/null
jq -e --arg loop_id "$loop_id" --argjson max_tokens "$max_tokens" --argjson max_iterations "$max_iterations" '
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
      (
        ((.terminal_reason | startswith("token_budget_exhausted:")) and .token_estimate >= $max_tokens) or
        ((.terminal_reason | startswith("iteration_budget_exhausted:")) and .iteration >= $max_iterations)
      )
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
expected_worker_slugs="$(jq -c '[.worker_types[].slug] | sort' "$WORKER_CATALOG")"
jq -e --argjson expected_worker_slugs "$expected_worker_slugs" '
  .schema_version == 1 and
  ([.workers[].slug] | sort) == $expected_worker_slugs and
  all(.workers[];
    (.slug | type == "string" and length > 0) and
    (.support_status | IN("verified_local_dev", "not_supported")) and
    (.definition | type == "string" and length > 0) and
    (.database_registration | type == "string" and length > 0) and
    (.runtime_catalog | type == "string" and length > 0) and
    (.live_create_option | type == "string" and length > 0) and
    (.image_target | type == "string" and length > 0) and
    (.runner | type == "string" and length > 0) and
    (.preflight | type == "string" and length > 0) and
    (.product_path | type == "string" and length > 0) and
    (.browser | type == "string" and length > 0) and
    (.evidence_ref | type == "string" and length > 0)
  )
' catalog/worker-evidence-matrix.json >/dev/null
test -s evidence/preflight-metadata-2026-07-16.json
jq -e '
  .schema_version == 1 and
  .safety_boundary.provider_request == false and
  .safety_boundary.pod_create_rpc == false and
  (.results | length == 6) and
  ([.results[].worker_type] | sort) ==
    ["codex-cli", "do-agent", "gemini-cli", "minimax-cli", "openclaw", "seedance-expert"] and
  (all(.results[] | select(.worker_type == "codex-cli" or .worker_type == "do-agent" or .worker_type == "minimax-cli" or .worker_type == "openclaw"); .result == "resolved" and (.issues | length == 0))) and
  (first(.results[] | select(.worker_type == "gemini-cli")).result == "blocked") and
  (first(.results[] | select(.worker_type == "seedance-expert")).result == "blocked") and
  (.residual_active_pods | length == 5)
' evidence/preflight-metadata-2026-07-16.json >/dev/null
test -s evidence/runner-runtime-attestation-2026-07-16.json
jq -e '
  .schema_version == 1 and
  .safety_boundary.provider_request == false and
  .safety_boundary.pod_create_rpc == false and
  (.online_runners | length) == 5 and
  ([.online_runners[].runner_id] | sort) == [
    "dev-runner-codex",
    "dev-runner-do-agent",
    "dev-runner-gemini",
    "dev-runner-minimax",
    "dev-runner-openclaw"
  ] and
  (.runner_protocol_contract.reported_fields == [
    "available_agents",
    "agent_versions[].slug",
    "agent_versions[].version",
    "agent_versions[].path"
  ]) and
  (.runner_protocol_contract.missing_attestation_fields | sort) == [
    "adapter_id",
    "interaction_modes",
    "runtime_image_digest"
  ] and
  (.release_catalog.enabled_worker_type_slugs == [])
' evidence/runner-runtime-attestation-2026-07-16.json >/dev/null
grep -Fq 'Worker Integration Evidence Rebuild Plan' PROGRESS.md
