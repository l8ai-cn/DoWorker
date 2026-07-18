#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")"

grep -q 'local http_port="${HTTP_PORT:-' lib/config_gen.sh
grep -q 'local backend_http_port="${BACKEND_HTTP_PORT:-' lib/config_gen.sh
