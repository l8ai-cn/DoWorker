#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CONFIG="${ROOT}/deploy/kubernetes/cluster-oilan/02-configmap.yaml"
RELAY="${ROOT}/deploy/kubernetes/cluster-oilan/31-relay.yaml"
INGRESS="${ROOT}/deploy/kubernetes/cluster-oilan/40-ingress.yaml"
PREVIEW_INGRESS="${ROOT}/deploy/kubernetes/cluster-oilan/44-preview-ingress.yaml"
KUSTOMIZATION="${ROOT}/deploy/kubernetes/cluster-oilan/kustomization.yaml"
RENDERED="$(mktemp)"
trap 'rm -f "$RENDERED"' EXIT

grep -Fq 'PREVIEW_PUBLIC_ORIGIN: "https://preview.dowork.l8ai.cn"' "$CONFIG"
grep -Fq 'PREVIEW_COOKIE_MODE: "partitioned"' "$CONFIG"
grep -A5 -F -- '- name: PREVIEW_PUBLIC_ORIGIN' "$RELAY" |
  grep -Fq 'key: PREVIEW_PUBLIC_ORIGIN'

grep -Fq -- '- host: "*.preview.dowork.l8ai.cn"' "$PREVIEW_INGRESS"
grep -Fq -- '- hosts: ["*.preview.dowork.l8ai.cn"]' "$PREVIEW_INGRESS"
grep -Fq -- '- path: /preview' "$PREVIEW_INGRESS"
grep -Fq 'nginx.ingress.kubernetes.io/ssl-redirect: "true"' "$PREVIEW_INGRESS"
grep -Fq -- '- 44-preview-ingress.yaml' "$KUSTOMIZATION"

if grep -Fq -- '- path: /preview' "$INGRESS"; then
  echo "preview route must not remain on the authenticated app origin" >&2
  exit 1
fi

kubectl kustomize "${ROOT}/deploy/kubernetes/cluster-oilan" >"$RENDERED"
test "$(grep -Fc 'path: /preview' "$RENDERED")" -eq 1
preview_block="$(awk 'BEGIN { RS = "---" } /name: agentsmesh-preview/ { print }' "$RENDERED")"
grep -Fq "host: '*.preview.dowork.l8ai.cn'" <<<"$preview_block"
grep -Fq 'name: relay' <<<"$preview_block"
grep -Fq 'number: 8090' <<<"$preview_block"
grep -Fq 'pathType: Prefix' <<<"$preview_block"
grep -Fq 'nginx.ingress.kubernetes.io/ssl-redirect: "true"' <<<"$preview_block"

echo "OILAN preview origin deployment contract passed"
