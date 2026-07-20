#!/usr/bin/env bash

# shellcheck source=dosql_release_gate.sh
source "${DIR}/dosql_release_gate.sh"

apply_stateful_prerequisites() {
  echo "==> apply stateful prerequisites"
  dexec "kubectl apply -f 02-configmap.yaml -f 30-backend-rbac.yaml"
  apply_pinned_manifest "10-postgres.yaml" pgvector
  apply_pinned_manifest "11-redis.yaml" redis
  apply_pinned_manifest "12-minio.yaml" minio
  dexec "kubectl -n ${NS} rollout status deploy/postgres --timeout=300s"
}
