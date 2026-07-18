#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
source lib/doctor.sh

error() { :; }

lsof() {
  printf '%s\n' \
    'COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME' \
    'netdisk_s 1159 user 11u IPv4 0 0t0 TCP 127.0.0.1:10000 (LISTEN)'
}

if require_unshadowed_loopback_port; then
  echo "expected IPv4 loopback shadow to fail" >&2
  exit 1
fi

lsof() {
  printf '%s\n' \
    'COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME' \
    'com.docke 11227 user 250u IPv6 0 0t0 TCP *:10000 (LISTEN)'
}

require_unshadowed_loopback_port
grep -Fq 'require_unshadowed_loopback_port || exit 1' dev.sh
