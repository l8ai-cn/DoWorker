#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
SCRIPT="push-images.sh"
RUNNER_SCRIPT="push-runner-images.sh"
MANIFEST_SCRIPT="harbor-manifest-digest.sh"

grep -q 'bash "${SCRIPT_DIR}/push-runner-images.sh" all' "$SCRIPT"
grep -q 'bash "${SCRIPT_DIR}/push-runner-images.sh" do-agent' "$SCRIPT"
grep -q 'release_source_guard.sh' "$SCRIPT"
grep -q 'release_require_pushed_clean_tree' "$SCRIPT"
grep -q 'release_write_source_metadata' "harbor-image-publishing.sh"
grep -q 'harbor-manifest-digest.sh' "harbor-image-publishing.sh"
grep -q 'harbor-infra-mirror.sh' "$SCRIPT"
grep -q 'harbor_require_upload_token_expiration "${REG}" 120' "$SCRIPT"
grep -Fq -- '--provenance=false' "harbor-image-publishing.sh"
grep -Fq '.platform.architecture == $architecture' "$MANIFEST_SCRIPT"
grep -Fq 'infra_manifest_digest' "$MANIFEST_SCRIPT"
grep -Fq -- '--prefer-index=false' "$MANIFEST_SCRIPT"
grep -Fq 'infra_manifest_digest "${PROJ}/${dest}"' "harbor-infra-mirror.sh"
! grep -En '(^|[^[:alnum:]_])manifest_digest\(' \
  "$MANIFEST_SCRIPT" "harbor-image-publishing.sh" "harbor-infra-mirror.sh" "$RUNNER_SCRIPT"
grep -Fq 'platform release promotion digest mismatch' "harbor-image-publishing.sh"
grep -Fq 'platform manifest promotion failed; retry' "$MANIFEST_SCRIPT"
grep -q 'release_collect_image_revisions' "release_source_guard.sh"
grep -q 'runner-do-agent runner-video-studio' "release_image_provenance.sh"
grep -q 'release_source_guard.sh' "$RUNNER_SCRIPT"
grep -q 'release_require_pushed_clean_tree' "$RUNNER_SCRIPT"
grep -q 'harbor-manifest-digest.sh' "$RUNNER_SCRIPT"
grep -Fq 'PLATFORM="${PLATFORM:-linux/amd64}"' "$RUNNER_SCRIPT"
grep -Fq 'export PLATFORM' "$RUNNER_SCRIPT"
grep -q 'harbor_require_upload_token_expiration "${REG}" 120' "$RUNNER_SCRIPT"
grep -Fq 'RUNNER_SOURCE_METADATA_MODE="${2:-write}"' "$RUNNER_SCRIPT"
grep -q 'defer-platform-source-metadata' "$RUNNER_SCRIPT"
grep -q 'RUNNER_SOURCE_METADATA_MODE' "push-runner-video-studio.sh"
! grep -Rq 'DEFER_PLATFORM_SOURCE_METADATA' "$SCRIPT" "$RUNNER_SCRIPT" "push-runner-video-studio.sh"
grep -q 'verify_runner_build_base' "$RUNNER_SCRIPT"
grep -Fq 'export RUNTIME_BUILD_BASE="${RUNNER_BUILD_BASE}"' "$RUNNER_SCRIPT"
grep -Fq -- '--build-arg "RUNTIME_SHARED_BASE=${RUNTIME_BUILD_BASE}"' "$RUNNER_SCRIPT"
grep -q 'FORCE_REBUILD=1' "$RUNNER_SCRIPT"
grep -q 'REQUIRE_DO_AGENT_BINARY=1' "$RUNNER_SCRIPT"
grep -q 'DO_AGENT_BINARY_SHA256' "$RUNNER_SCRIPT"
grep -q 'publish_do_agent' "$RUNNER_SCRIPT"
grep -q 'candidate_tag="candidate-${RELEASE_SOURCE_COMMIT:0:12}"' "$RUNNER_SCRIPT"
grep -Fq 'candidate_digest="$(platform_manifest_digest "${repository}:${candidate_tag}")"' "$RUNNER_SCRIPT"
grep -q 'do_agent_release_value build.source_date_epoch' "$RUNNER_SCRIPT"
grep -q 'harbor_ensure_immutable_tag' "$RUNNER_SCRIPT"
grep -q 'docker buildx imagetools create' "$RUNNER_SCRIPT"
grep -q -- '--prefer-index=false' "$RUNNER_SCRIPT"
grep -q 'do_agent_release_value artifact.binary_sha256' "$RUNNER_SCRIPT"
grep -q 'update-do-agent-runtime-digest.mjs' "$RUNNER_SCRIPT"
grep -q 'verify_do_agent_labels' "$RUNNER_SCRIPT"
grep -q 'EXPECTED_REMOTE_DIGEST=' "$RUNNER_SCRIPT"
grep -q 'runtime_mapping_contract_test.sh' "$RUNNER_SCRIPT"
! grep -q 'docker push "${repository}:latest"' "$RUNNER_SCRIPT"

candidate_line="$(grep -n 'docker push "${repository}:${candidate_tag}"' "$RUNNER_SCRIPT" | cut -d: -f1)"
immutable_line="$(grep -n 'harbor_ensure_immutable_tag' "$RUNNER_SCRIPT" | tail -1 | cut -d: -f1)"
update_line="$(grep -n 'update-do-agent-runtime-digest.mjs' "$RUNNER_SCRIPT" | cut -d: -f1)"
latest_line="$(grep -n -- '--tag "${repository}:latest"' "$RUNNER_SCRIPT" | cut -d: -f1)"
[[ "${candidate_line}" -lt "${immutable_line}" ]]
[[ "${immutable_line}" -lt "${update_line}" ]]
[[ "${update_line}" -lt "${latest_line}" ]]

all_case="$(grep -F 'all)' "$SCRIPT")"
[[ "${all_case}" == *'push_infra'* ]]
[[ "${all_case}" == *'push-runner-images.sh" all'* ]]
[[ "${all_case}" == *'all defer-platform-source-metadata'* ]]
[[ "${all_case}" == *'push_platform'* ]]

bash harbor_infra_mirror_contract_test.sh
bash harbor_upload_token_contract_test.sh
bash configure_harbor_upload_token_contract_test.sh
bash runner_build_base_contract_test.sh
bash runner_build_base_passthrough_contract_test.sh
bash oilan_staging_promotion_contract_test.sh
bash harbor_credentials_contract_test.sh
