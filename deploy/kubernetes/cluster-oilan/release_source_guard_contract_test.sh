#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

git init --bare "${TMP}/origin.git" >/dev/null
git init -b main "${TMP}/repo" >/dev/null
git -C "${TMP}/repo" config user.name "Release Contract"
git -C "${TMP}/repo" config user.email "release-contract@example.test"
git -C "${TMP}/repo" remote add origin "${TMP}/origin.git"
mkdir -p "${TMP}/bin"
cat > "${TMP}/bin/gh" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' '[{
  "total_count": 3,
  "check_runs": [
    {"name":"Runtime release contracts","status":"completed","conclusion":"success"},
    {"name":"Loop and sandbox security regressions","status":"completed","conclusion":"success"},
    {"name":"Web-user artifact preview","status":"completed","conclusion":"success"}
  ]
}]'
EOF
chmod +x "${TMP}/bin/gh"
cat > "${TMP}/bin/docker" <<'EOF'
#!/usr/bin/env bash
if [[ "${1:-}" == "pull" ]]; then
  exit 0
fi
if [[ "${1:-} ${2:-}" == "image inspect" ]]; then
  printf '%s\n' "${EXPECTED_REVISION:-${RELEASE_SOURCE_COMMIT}}"
  exit 0
fi
exit 1
EOF
chmod +x "${TMP}/bin/docker"
export PATH="${TMP}/bin:${PATH}"
printf 'release\n' > "${TMP}/repo/release.txt"
git -C "${TMP}/repo" add release.txt
git -C "${TMP}/repo" commit -m "release" >/dev/null
git -C "${TMP}/repo" push -u origin main >/dev/null

# shellcheck source=release_source_guard.sh
source "${ROOT}/release_source_guard.sh"
release_require_pushed_clean_tree "${TMP}/repo"
[[ "${RELEASE_SOURCE_COMMIT}" == "$(git -C "${TMP}/repo" rev-parse HEAD)" ]]
mkdir -p "${TMP}/repo/deploy/kubernetes/cluster-oilan/release"
{
  printf '%s\n' 'images:'
  for image in $(release_platform_images); do
    printf '%s\n' \
      "  - name: repo.aiedulab.cn:8443/agentsmesh/${image}" \
      '    digest: sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
  done
} > "${TMP}/repo/deploy/kubernetes/cluster-oilan/release/kustomization.yaml"
release_write_source_metadata "${TMP}/repo"
release_verify_source_metadata "${TMP}/repo"
git -C "${TMP}/repo" add deploy/kubernetes/cluster-oilan/release
git -C "${TMP}/repo" commit -m "release metadata" >/dev/null
git -C "${TMP}/repo" push >/dev/null
release_require_pushed_clean_tree "${TMP}/repo"
release_verify_source_metadata "${TMP}/repo"

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

git -C "${TMP}/repo" switch -c codex/not-main >/dev/null
git -C "${TMP}/repo" push -u origin codex/not-main >/dev/null
if release_require_pushed_clean_tree "${TMP}/repo" 2>/dev/null; then
  echo "non-main release branch was accepted" >&2
  exit 1
fi

git -C "${TMP}/repo" switch main >/dev/null
cat > "${TMP}/bin/gh" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' '[{
  "total_count": 5,
  "check_runs": [
    {"name":"Runtime release contracts","status":"completed","conclusion":"success"},
    {"name":"Loop and sandbox security regressions","status":"completed","conclusion":"success"},
    {"name":"Web-user artifact preview","status":"completed","conclusion":"success"},
    {"name":"Deploy US West","status":"queued","conclusion":null},
    {"name":"Deploy CN","status":"queued","conclusion":null}
  ]
}]'
EOF
chmod +x "${TMP}/bin/gh"
release_require_pushed_clean_tree "${TMP}/repo"

cat > "${TMP}/bin/gh" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' '[{
  "total_count": 4,
  "check_runs": [
    {"name":"Runtime release contracts","status":"completed","conclusion":"success"},
    {"name":"Loop and sandbox security regressions","status":"completed","conclusion":"success"},
    {"name":"Web-user artifact preview","status":"completed","conclusion":"success"},
    {"name":"Unknown pending check","status":"queued","conclusion":null}
  ]
}]'
EOF
chmod +x "${TMP}/bin/gh"
if release_require_pushed_clean_tree "${TMP}/repo" 2>/dev/null; then
  echo "release with an unknown pending check was accepted" >&2
  exit 1
fi

cat > "${TMP}/bin/gh" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' '[{
  "total_count": 3,
  "check_runs": [
    {"name":"Runtime release contracts","status":"completed","conclusion":"success"},
    {"name":"Loop and sandbox security regressions","status":"completed","conclusion":"failure"},
    {"name":"Web-user artifact preview","status":"completed","conclusion":"success"}
  ]
}]'
EOF
chmod +x "${TMP}/bin/gh"
if release_require_pushed_clean_tree "${TMP}/repo" 2>/dev/null; then
  echo "failed CI release was accepted" >&2
  exit 1
fi

cat > "${TMP}/bin/gh" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' '[{
  "total_count": 2,
  "check_runs": [
    {"name":"Runtime release contracts","status":"completed","conclusion":"success"},
    {"name":"Web-user artifact preview","status":"completed","conclusion":"success"}
  ]
}]'
EOF
chmod +x "${TMP}/bin/gh"
if release_require_pushed_clean_tree "${TMP}/repo" 2>/dev/null; then
  echo "release with a missing required check was accepted" >&2
  exit 1
fi

cat > "${TMP}/bin/gh" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' '[{
  "total_count": 3,
  "check_runs": [
    {"name":"Runtime release contracts","status":"completed","conclusion":"skipped"},
    {"name":"Loop and sandbox security regressions","status":"completed","conclusion":"success"},
    {"name":"Web-user artifact preview","status":"completed","conclusion":"success"}
  ]
}]'
EOF
chmod +x "${TMP}/bin/gh"
if release_require_pushed_clean_tree "${TMP}/repo" 2>/dev/null; then
  echo "release with a skipped required check was accepted" >&2
  exit 1
fi

cat > "${TMP}/bin/gh" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' '[{
  "total_count": 4,
  "check_runs": [
    {"name":"Runtime release contracts","status":"completed","conclusion":"success"},
    {"name":"Loop and sandbox security regressions","status":"completed","conclusion":"success"},
    {"name":"Web-user artifact preview","status":"completed","conclusion":"success"}
  ]
}]'
EOF
chmod +x "${TMP}/bin/gh"
if release_require_pushed_clean_tree "${TMP}/repo" 2>/dev/null; then
  echo "incomplete check pagination was accepted" >&2
  exit 1
fi

old_revision="$(git -C "${TMP}/repo" rev-parse HEAD^)"
current_revision="$(git -C "${TMP}/repo" rev-parse HEAD)"
jq \
  --arg commit "${current_revision}" \
  --arg backend "${old_revision}" \
  --arg web "${current_revision}" \
  '.commit = $commit | .images.backend = $backend | .images.web = $web' \
  "${TMP}/repo/deploy/kubernetes/cluster-oilan/release/source.json" \
  > "${TMP}/repo/deploy/kubernetes/cluster-oilan/release/source.json.tmp"
mv \
  "${TMP}/repo/deploy/kubernetes/cluster-oilan/release/source.json.tmp" \
  "${TMP}/repo/deploy/kubernetes/cluster-oilan/release/source.json"
EXPECTED_REVISION="${old_revision}" release_verify_platform_image_provenance \
  "${TMP}/repo" backend
EXPECTED_REVISION="${current_revision}" release_verify_platform_image_provenance \
  "${TMP}/repo" web
if EXPECTED_REVISION="${current_revision}" \
  release_verify_platform_image_provenance "${TMP}/repo" backend 2>/dev/null; then
  echo "mismatched remote image provenance was accepted" >&2
  exit 1
fi
