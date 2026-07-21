#!/usr/bin/env bash
# Cross-compile Agent Cloud runner CLI for 6 platforms and pack archives.
#
# Outputs (under dist/runner-release/ by default):
#   agent-cloud-runner_<goos>_<goarch>.{tar.gz,zip}
#   checksums.txt
#
# Env:
#   VERSION     optional; when set, also stamps -X main.version
#   OUT_DIR     default: dist/runner-release
#   BUILD_TIME  default: UTC now
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

OUT_DIR="${OUT_DIR:-$ROOT/dist/runner-release}"
VERSION="${VERSION:-dev}"
BUILD_TIME="${BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"
BINARY_NAME="agent-cloud-runner"
ARCHIVE_BASENAME="agent-cloud-runner"

# Generated Go stubs are gitignored; ensure they exist before cross-compile.
if [[ ! -f "$ROOT/proto/gen/go/runner/v1/runner.pb.go" ]]; then
  bash "$ROOT/scripts/proto-gen-go.sh" --force
fi

LDFLAGS="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

mkdir -p "$OUT_DIR"
rm -f "$OUT_DIR"/${ARCHIVE_BASENAME}_*.tar.gz "$OUT_DIR"/${ARCHIVE_BASENAME}_*.zip "$OUT_DIR"/checksums.txt

PLATFORMS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
)

stage="$(mktemp -d)"
trap 'rm -rf "$stage"' EXIT

for plat in "${PLATFORMS[@]}"; do
  goos="${plat%/*}"
  goarch="${plat#*/}"
  ext=""
  [[ "$goos" == "windows" ]] && ext=".exe"
  bin_name="${BINARY_NAME}${ext}"
  out_bin="$stage/${goos}_${goarch}/${bin_name}"
  mkdir -p "$(dirname "$out_bin")"

  echo "==> building ${goos}/${goarch}"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -trimpath -ldflags "$LDFLAGS" \
    -o "$out_bin" ./runner/cmd/runner

  cp README.md LICENSE "$stage/${goos}_${goarch}/"

  if [[ "$goos" == "windows" ]]; then
    archive="${OUT_DIR}/${ARCHIVE_BASENAME}_${goos}_${goarch}.zip"
    (
      cd "$stage/${goos}_${goarch}"
      zip -q "$archive" "$bin_name" README.md LICENSE
    )
  else
    archive="${OUT_DIR}/${ARCHIVE_BASENAME}_${goos}_${goarch}.tar.gz"
    tar -C "$stage/${goos}_${goarch}" -czf "$archive" "$bin_name" README.md LICENSE
  fi
  echo "    -> $archive"
done

(
  cd "$OUT_DIR"
  sha256sum ${ARCHIVE_BASENAME}_*.tar.gz ${ARCHIVE_BASENAME}_*.zip > checksums.txt
)
echo "==> wrote $OUT_DIR/checksums.txt"
cat "$OUT_DIR/checksums.txt"
