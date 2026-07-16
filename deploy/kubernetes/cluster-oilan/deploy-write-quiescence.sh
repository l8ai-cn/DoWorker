#!/usr/bin/env bash

APP_WRITES_STOPPED=false
BACKEND_REPLICAS=""
MARKETPLACE_REPLICAS=""

deployment_replicas() {
  local name="${1:?deployment name is required}"
  dexec "kubectl -n ${NS} get deploy ${name} -o jsonpath='{.spec.replicas}'" |
    tail -n 1 | tr -d '\r'
}

stop_application_writes() {
  local backend marketplace
  if backend="$(deployment_replicas backend 2>/dev/null)"; then
    marketplace="$(deployment_replicas marketplace)"
  else
    if deployment_replicas marketplace >/dev/null 2>&1; then
      echo "backend deployment missing while marketplace exists" >&2
      return 1
    fi
    echo "==> no existing application writers"
    return
  fi
  [[ "${backend}" =~ ^[0-9]+$ && "${marketplace}" =~ ^[0-9]+$ ]]
  BACKEND_REPLICAS="${backend}"
  MARKETPLACE_REPLICAS="${marketplace}"
  dexec "kubectl -n ${NS} scale deploy/backend deploy/marketplace --replicas=0"
  APP_WRITES_STOPPED=true
  dexec "set -eu; kubectl -n ${NS} wait --for=delete pod -l app=backend --timeout=180s; kubectl -n ${NS} wait --for=delete pod -l app=marketplace --timeout=180s"
}

mark_application_writes_restored() {
  APP_WRITES_STOPPED=false
}
