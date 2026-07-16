#!/usr/bin/env bash
set -euo pipefail

rendered="${1:?rendered manifest path is required}"
registry="repo.aiedulab.cn:8443/agentsmesh"

image_references() {
  awk '
    $1 == "image:" { print $2 }
    $1 == "-" && $2 == "image:" { print $3 }
  ' "${rendered}"
}

for image in backend marketplace marketplace-web relay web web-admin mobile pgvector redis minio kubectl; do
  references="$(image_references | awk -v prefix="${registry}/${image}@" \
    'index($1, prefix) == 1 { print $1 }')"
  [[ -n "${references}" ]] || {
    echo "missing platform image: ${image}" >&2
    exit 1
  }
  while IFS= read -r reference; do
    prefix="${registry}/${image}@sha256:"
    digest="${reference#"${prefix}"}"
    [[ "${reference}" == "${prefix}${digest}" &&
       "${#digest}" -eq 64 &&
       ! "${digest}" =~ [^a-f0-9] ]] || {
      echo "mutable platform image: ${reference}" >&2
      exit 1
    }
  done <<< "${references}"
  [[ "$(printf '%s\n' "${references}" | sort -u | wc -l | tr -d ' ')" == "1" ]] || {
    echo "platform image uses multiple digests: ${image}" >&2
    exit 1
  }
done

while IFS= read -r reference; do
  [[ "${reference}" =~ ^${registry}/[^@[:space:]]+@sha256:[a-f0-9]{64}$ ]] || {
    echo "mutable release image: ${reference}" >&2
    exit 1
  }
done < <(image_references)

backend_reference="$(image_references | awk -v prefix="${registry}/backend@sha256:" \
  'index($1, prefix) == 1 { print $1; exit }')"
backend_digest="${backend_reference##*@}"
annotations="$(awk '$1 == "agentsmesh.ai/verified-image-digest:" { gsub(/"/, "", $2); print $2 }' "${rendered}")"
[[ -n "${annotations}" ]] || {
  echo "missing verified backend digest annotation" >&2
  exit 1
}
while IFS= read -r annotation; do
  [[ "${annotation}" == "${backend_digest}" ]] || {
    echo "backend digest annotation drift: ${annotation} != ${backend_digest}" >&2
    exit 1
  }
done <<< "${annotations}"

! grep -q '__BACKEND_' "${rendered}" || {
  echo "unresolved backend release placeholder" >&2
  exit 1
}
