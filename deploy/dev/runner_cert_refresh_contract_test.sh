#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

grep -Fq 'openssl verify -CAfile "$SSL_DIR/ca.crt" "$CERTS_DIR/runner.crt"' \
  runner-entrypoint.sh
