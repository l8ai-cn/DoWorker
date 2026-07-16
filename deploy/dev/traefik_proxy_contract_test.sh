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

temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT

success() { :; }
SCRIPT_DIR="$temp_dir"
WORKTREE_NAME="traefik-contract"
BACKEND_HTTP_PORT=11015
BACKEND_GRPC_PORT=11016
RELAY_HTTP_PORT=11017
source lib/config_gen.sh
generate_traefik_config

node - "$temp_dir/traefik/dynamic/http.yml" <<'NODE'
const fs = require("node:fs");
const YAML = require("yaml");
const config = YAML.parse(fs.readFileSync(process.argv[2], "utf8"));
if (!config?.http?.routers?.["backend-api"]?.rule) process.exit(1);
NODE
