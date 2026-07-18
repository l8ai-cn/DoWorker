#!/usr/bin/env bash
set -euo pipefail

ROOT="$(mktemp -d)"
trap 'rm -rf "${ROOT}"' EXIT
REPO="${ROOT}/repo"
SCRIPT_DIR="${REPO}/deploy/kubernetes/cluster-oilan"
LOG="${ROOT}/calls.log"
mkdir -p "${SCRIPT_DIR}" "${REPO}/docker/agent-runtime"

cp "$(dirname "$0")/push-runner-images.sh" "${SCRIPT_DIR}/"
cp "$(dirname "$0")/push-runner-video-studio.sh" "${SCRIPT_DIR}/"
cp "$(dirname "$0")/harbor-manifest-digest.sh" "${SCRIPT_DIR}/"
cp "$(dirname "$0")/runner-build-base.sh" "${SCRIPT_DIR}/"
VERIFY_LINE="$(grep -n '^verify_runner_build_base$' "${SCRIPT_DIR}/push-runner-images.sh" | cut -d: -f1)"
EXPORT_LINE="$(grep -n '^export RUNTIME_BUILD_BASE=' "${SCRIPT_DIR}/push-runner-images.sh" | cut -d: -f1)"
CASE_LINE="$(grep -n '^case "${TARGET}" in$' "${SCRIPT_DIR}/push-runner-images.sh" | cut -d: -f1)"
[[ "${VERIFY_LINE}" -lt "${EXPORT_LINE}" && "${EXPORT_LINE}" -lt "${CASE_LINE}" ]]
sed '/^verify_runner_build_base$/,$d' "${SCRIPT_DIR}/push-runner-images.sh" \
  > "${SCRIPT_DIR}/push-runner-library.sh"

cat > "${REPO}/docker/agent-runtime/do_agent_release_manifest.sh" <<'EOF'
#!/usr/bin/env bash
EOF
cat > "${SCRIPT_DIR}/harbor_immutable_release.sh" <<'EOF'
#!/usr/bin/env bash
EOF
cat > "${SCRIPT_DIR}/release_source_guard.sh" <<'EOF'
#!/usr/bin/env bash
release_require_pushed_clean_tree() { :; }
EOF

source "${SCRIPT_DIR}/push-runner-library.sh"
export RUNTIME_BUILD_BASE="${RUNNER_BUILD_BASE}"
export RELEASE_SOURCE_COMMIT="1111111111111111111111111111111111111111"

require_do_agent_artifact() { :; }
do_agent_release_value() {
  case "$1" in
    source.commit) printf '%040d' 1 ;;
    build.source_date_epoch) printf '1784246400' ;;
    artifact.binary_sha256) printf 'sha256:%064d' 1 ;;
    *) return 1 ;;
  esac
}
push_runtime() { :; }
publish_do_agent() { :; }
publish_video_runtime_metadata() { :; }
bash() {
  printf '%s|bash %s\n' "${RUNTIME_BUILD_BASE:-}" "$*" >> "${LOG}"
}
docker() {
  printf '%s|docker %s\n' "${RUNTIME_BUILD_BASE:-}" "$*" >> "${LOG}"
}

push_all
push_video_studio

EXPECTED="${PROJ}/runner-node-base@${RUNNER_BUILD_BASE_DIGEST}"
[[ "$(wc -l < "${LOG}" | tr -d ' ')" -eq 11 ]]
if grep -Fv "${EXPECTED}|" "${LOG}" >/dev/null; then
  echo "every Runner build entry must inherit the locked Harbor base" >&2
  exit 1
fi
for runtime in \
  claude-code codex-cli video-studio gemini-cli grok-build minimax-cli \
  openclaw hermes do-agent; do
  grep -Fq "bash docker/agent-runtime/build.sh ${runtime}" "${LOG}"
done
grep -Fq "docker build --platform linux/amd64 --target runtime" "${LOG}"
grep -Fq -- "--build-arg RUNTIME_BUILD_BASE=${EXPECTED}" "${LOG}"
