#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

git init --bare "${TMP}/origin.git" >/dev/null
git init -b codex/release-test "${TMP}/repo" >/dev/null
git -C "${TMP}/repo" config user.name "Release Contract"
git -C "${TMP}/repo" config user.email "release-contract@example.test"
git -C "${TMP}/repo" remote add origin "${TMP}/origin.git"
printf 'release\n' > "${TMP}/repo/release.txt"
git -C "${TMP}/repo" add release.txt
git -C "${TMP}/repo" commit -m "release" >/dev/null
git -C "${TMP}/repo" push -u origin codex/release-test >/dev/null

# shellcheck source=release_source_guard.sh
source "${ROOT}/release_source_guard.sh"
release_require_pushed_clean_tree "${TMP}/repo"

printf 'dirty\n' >> "${TMP}/repo/release.txt"
if release_require_pushed_clean_tree "${TMP}/repo" 2>/dev/null; then
  echo "dirty release tree was accepted" >&2
  exit 1
fi

git -C "${TMP}/repo" restore release.txt
printf 'unpushed\n' >> "${TMP}/repo/release.txt"
git -C "${TMP}/repo" add release.txt
git -C "${TMP}/repo" commit -m "unpushed" >/dev/null
if release_require_pushed_clean_tree "${TMP}/repo" 2>/dev/null; then
  echo "unpushed release commit was accepted" >&2
  exit 1
fi

git -C "${TMP}/repo" push >/dev/null
release_require_pushed_clean_tree "${TMP}/repo"
