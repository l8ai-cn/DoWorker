#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
SCRIPT="push-images.sh"
RUNNER_SCRIPT="push-runner-images.sh"

grep -q 'bash "${SCRIPT_DIR}/push-runner-images.sh" all' "$SCRIPT"
grep -q 'bash "${SCRIPT_DIR}/push-runner-images.sh" do-agent' "$SCRIPT"
grep -q 'FORCE_REBUILD=1' "$RUNNER_SCRIPT"
grep -q 'REQUIRE_DO_AGENT_BINARY=1' "$RUNNER_SCRIPT"
grep -q 'DO_AGENT_BINARY_SHA256' "$RUNNER_SCRIPT"
grep -q 'publish_do_agent' "$RUNNER_SCRIPT"
grep -q 'candidate_tag="candidate-${release_tag}"' "$RUNNER_SCRIPT"
grep -q 'do_agent_release_value image.tag' "$RUNNER_SCRIPT"
grep -q 'do_agent_release_value build.source_date_epoch' "$RUNNER_SCRIPT"
grep -q 'harbor_ensure_immutable_tag' "$RUNNER_SCRIPT"
grep -q 'docker buildx imagetools create' "$RUNNER_SCRIPT"
grep -q -- '--prefer-index=false' "$RUNNER_SCRIPT"
grep -q 'do_agent_release_value artifact.binary_sha256' "$RUNNER_SCRIPT"
grep -q 'do_agent_release_value image.digest' "$RUNNER_SCRIPT"
grep -q 'EXPECTED_REMOTE_DIGEST=' "$RUNNER_SCRIPT"
grep -q 'runtime_mapping_contract_test.sh' "$RUNNER_SCRIPT"
! grep -q 'docker push "${repository}:latest"' "$RUNNER_SCRIPT"

candidate_line="$(grep -n 'docker push "${repository}:${candidate_tag}"' "$RUNNER_SCRIPT" | cut -d: -f1)"
immutable_line="$(grep -n 'harbor_ensure_immutable_tag' "$RUNNER_SCRIPT" | tail -1 | cut -d: -f1)"
latest_line="$(grep -n -- '--tag "${repository}:latest"' "$RUNNER_SCRIPT" | cut -d: -f1)"
[[ "${candidate_line}" -lt "${immutable_line}" ]]
[[ "${immutable_line}" -lt "${latest_line}" ]]
