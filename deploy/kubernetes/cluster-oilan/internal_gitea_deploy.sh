#!/usr/bin/env bash

backup_internal_gitea() {
  if ! dexec "kubectl -n ${NS} get deploy/gitea -o name >/dev/null"; then
    return
  fi
  local replicas
  replicas="$(
    dexec "kubectl -n ${NS} get deploy/gitea -o jsonpath='{.spec.replicas}'" |
      tail -n 1 | tr -d '\r'
  )"
  [[ "${replicas}" =~ ^[1-9][0-9]*$ ]] || {
    echo "cannot back up internal Gitea with replicas=${replicas}" >&2
    return 1
  }
  dexec "kubectl -n ${NS} scale deploy/gitea --replicas=0" || return 1
  if run_internal_gitea_backup; then
    return
  fi
  echo "internal Gitea backup failed; restoring existing deployment" >&2
  restore_internal_gitea "${replicas}" || {
    echo "failed to restore internal Gitea after backup failure" >&2
    return 1
  }
  return 1
}

run_internal_gitea_backup() {
  dexec "kubectl -n ${NS} wait --for=delete pod -l app=gitea --timeout=180s" || return 1
  dexec "kubectl -n ${NS} delete pod gitea-backup --ignore-not-found --wait=true" || return 1
  apply_pinned_manifest "15-gitea-backup-pod.yaml" gitea "${REG}/library/gitea" || return 1
  dexec "kubectl -n ${NS} wait --for=condition=Ready pod/gitea-backup --timeout=180s" || return 1
  dexec "set -eu; test \"\$(kubectl -n ${NS} exec pod/gitea-backup -- sqlite3 /data/gitea/gitea.db 'PRAGMA quick_check;' | tr -d '\r')\" = ok; backup_dir=/root/backups/agentsmesh; timestamp=\$(date -u +%Y%m%dT%H%M%SZ); backup=\${backup_dir}/gitea-${RELEASE_DEPLOY_COMMIT:0:12}-\${timestamp}.tar.gz; umask 077; mkdir -p \"\${backup_dir}\"; kubectl -n ${NS} exec pod/gitea-backup -- tar -C /data -czf - . > \"\${backup}.tmp\"; test -s \"\${backup}.tmp\"; mv \"\${backup}.tmp\" \"\${backup}\"; sha256sum \"\${backup}\" > \"\${backup}.sha256\"; sha256sum -c \"\${backup}.sha256\"; echo \"Gitea backup: \${backup}\"" || return 1
  dexec "kubectl -n ${NS} delete pod gitea-backup --wait=true"
}

restore_internal_gitea() {
  local replicas="$1" result=0
  dexec "kubectl -n ${NS} delete pod gitea-backup --ignore-not-found --wait=true" || result=1
  dexec "kubectl -n ${NS} scale deploy/gitea --replicas=${replicas}" || result=1
  dexec "kubectl -n ${NS} rollout status deploy/gitea --timeout=300s" || result=1
  return "${result}"
}

ensure_internal_gitea() {
  backup_internal_gitea
  apply_pinned_manifest "14-gitea.yaml" gitea "${REG}/library/gitea"
  dexec "kubectl -n ${NS} rollout status deploy/gitea --timeout=300s"
  dexec "bash bootstrap_internal_gitea.sh ${NS}"
}
