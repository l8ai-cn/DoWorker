#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

traefik_block="$(
  awk '
    /^  traefik:$/ { in_traefik = 1 }
    in_traefik && /^  [a-zA-Z0-9_-]+:$/ && $1 != "traefik:" { exit }
    in_traefik { print }
  ' docker-compose.yml
)"

for variable in HTTP_PROXY HTTPS_PROXY http_proxy https_proxy; do
  grep -q "${variable}: \"\"" <<<"$traefik_block"
done
grep -q 'NO_PROXY: "\*"' <<<"$traefik_block"
grep -q 'no_proxy: "\*"' <<<"$traefik_block"
