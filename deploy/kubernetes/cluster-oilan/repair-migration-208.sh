#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET="${DOOPS_TARGET:-gw-oilan-node}"
BASE_SESSION="${DOOPS_SESSION:-$(doops session | tr -d '[:space:]')}"
RUN_ID="$(date -u +%Y%m%d%H%M%S)-$$"
SESSION="${BASE_SESSION}-migration-208-${RUN_ID}"
WS="/root/ws/${SESSION}"
NS=agentcloud
RESOURCE="migration-208-repair-${RUN_ID}"
BUNDLE=""
REMOTE_WORKSPACE_MAY_EXIST=false

dexec() {
  doops -session "${SESSION}" exec --target "${TARGET}" \
    --cmd "cd ${WS} && $1"
}

cleanup() {
  local result=$?
  trap - EXIT
  [[ -z "${BUNDLE}" || ! -d "${BUNDLE}" ]] || rm -rf "${BUNDLE}"
  if [[ "${REMOTE_WORKSPACE_MAY_EXIST}" == true ]]; then
    doops -session "${SESSION}" clean \
      --target "${TARGET}" --workspace "${SESSION}" || result=1
  fi
  exit "${result}"
}
trap cleanup EXIT

BUNDLE="$(mktemp -d)"
cp "${DIR}/24-repair-migration-208-preconditions.sql" "${BUNDLE}/"
cp "${DIR}/24-repair-migration-208.sql" "${BUNDLE}/"
sed "s|__RUN_ID__|${RUN_ID}|g" \
  "${DIR}/25-repair-migration-208-job.yaml" \
  > "${BUNDLE}/25-repair-migration-208-job.yaml"

REMOTE_WORKSPACE_MAY_EXIST=true
doops -session "${SESSION}" push --target "${TARGET}" --src "${BUNDLE}"
rm -rf "${BUNDLE}"
BUNDLE=""

dexec "kubectl -n ${NS} create configmap migration-208-repair-lock \
  --from-literal=run-id=${RUN_ID}"
dexec "kubectl -n ${NS} create configmap ${RESOURCE} \
  --from-file=repair-preconditions.sql=24-repair-migration-208-preconditions.sql \
  --from-file=repair.sql=24-repair-migration-208.sql \
  --dry-run=client -o yaml | kubectl create -f -"
dexec "kubectl apply -f 25-repair-migration-208-job.yaml"

if ! dexec "kubectl -n ${NS} wait \
  --for=condition=complete job/${RESOURCE} --timeout=600s"; then
  dexec "kubectl -n ${NS} logs job/${RESOURCE} \
    --all-containers --prefix=true || true"
  exit 1
fi

dexec "kubectl -n ${NS} logs job/${RESOURCE} \
  -c finish-migrations"
dexec "kubectl -n ${NS} delete job ${RESOURCE} \
  --wait=true"
dexec "kubectl -n ${NS} delete configmap ${RESOURCE}"
dexec "kubectl -n ${NS} delete configmap migration-208-repair-lock"
