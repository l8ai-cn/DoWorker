#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

if rg -q 'echo "\$keys_json" \| python3' gitea/init-gitea.sh; then
    echo "Gitea deploy-key parsing must not use a JSON pipe" >&2
    exit 1
fi

grep -Fq "key.get('fingerprint', '')" gitea/init-gitea.sh
