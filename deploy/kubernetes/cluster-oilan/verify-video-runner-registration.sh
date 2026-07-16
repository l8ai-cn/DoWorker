#!/usr/bin/env bash
set -euo pipefail

NS="${NS:-agentsmesh}"
NODE_ID="dev-runner-video-studio"
EXPECTED_AGENTS='["video-studio"]'

query_runner() {
  kubectl -n "${NS}" exec deploy/postgres -- \
    psql -U agentsmesh -d agentsmesh -Atc \
    "SELECT count(*)
     FROM runners r
     JOIN organizations o ON o.id = r.organization_id
     WHERE o.slug = 'dev-org'
       AND r.node_id = '${NODE_ID}'
       AND r.status = 'online'
       AND r.is_enabled
       AND r.last_heartbeat > NOW() - INTERVAL '90 seconds'
       AND r.max_concurrent_pods = 1
       AND r.available_agents @> '${EXPECTED_AGENTS}'::jsonb;"
}

for _ in $(seq 1 60); do
  [[ "$(query_runner)" == "1" ]] && exit 0
  sleep 2
done

kubectl -n "${NS}" exec deploy/postgres -- \
  psql -U agentsmesh -d agentsmesh -P pager=off -c \
  "SELECT o.slug, r.node_id, r.status, r.is_enabled, r.last_heartbeat,
          r.max_concurrent_pods, r.available_agents
   FROM runners r
   JOIN organizations o ON o.id = r.organization_id
   WHERE o.slug = 'dev-org' AND r.node_id = '${NODE_ID}';"
exit 1
