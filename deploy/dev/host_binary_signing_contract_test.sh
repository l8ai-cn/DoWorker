#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")"

bash -n build-host-go-binary.sh
grep -q 'codesign --force --sign -' build-host-go-binary.sh

for service in backend marketplace relay; do
    grep -q "bash deploy/dev/build-host-go-binary.sh" "air/${service}.toml"
done
