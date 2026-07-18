#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

if grep -Eq 'echo "\$keys_json" \| python3' gitea/init-gitea.sh; then
    echo "Gitea deploy-key parsing must not use a JSON pipe" >&2
    exit 1
fi

grep -Fq 'map(select(.title == $title))[0].key // empty' gitea/init-gitea.sh
grep -Fq 'ssh-keygen -lf "$registered_key_file"' gitea/init-gitea.sh
