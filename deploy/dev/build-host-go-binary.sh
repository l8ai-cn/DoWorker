#!/bin/bash
set -euo pipefail

package="$1"
output="$2"

go build -o "$output" "$package"
if [[ "$(uname -s)" == "Darwin" ]]; then
    codesign --force --sign - "$output" >/dev/null
fi
