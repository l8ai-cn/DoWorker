#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
source harbor_immutable_release.sh

test_lifetime=30

harbor_load_credentials() {
  HARBOR_USERNAME=test
  HARBOR_PASSWORD=test
}

harbor_registry_token() {
  TEST_LIFETIME="${test_lifetime}" python3 - <<'PY'
import base64
import json
import os

payload = json.dumps({"iat": 1000, "exp": 1000 + int(os.environ["TEST_LIFETIME"]) * 60})
encoded = base64.urlsafe_b64encode(payload.encode()).decode().rstrip("=")
print(f"header.{encoded}.signature")
PY
}

if harbor_require_upload_token_expiration registry.example 120 2>/dev/null; then
  echo "expected a 30-minute token to fail" >&2
  exit 1
fi

test_lifetime=120
harbor_require_upload_token_expiration registry.example 120
[[ -z "${HARBOR_USERNAME+x}" ]]
[[ -z "${HARBOR_PASSWORD+x}" ]]
