#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")"

if grep -q "python3" gitea/init-gitea.sh; then
    exit 1
fi
grep -q "command -v jq" gitea/init-gitea.sh
grep -q "jq -r" gitea/init-gitea.sh
