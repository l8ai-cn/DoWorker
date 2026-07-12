#!/usr/bin/env bash
set -euo pipefail

rendered="${1:?rendered manifest path is required}"
registry="repo.aiedulab.cn:8443/agentsmesh"

for image in backend marketplace marketplace-web relay web web-admin; do
  references="$(awk -v prefix="${registry}/${image}@" \
    '$1 == "image:" && index($2, prefix) == 1 {print $2}' "${rendered}")"
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
