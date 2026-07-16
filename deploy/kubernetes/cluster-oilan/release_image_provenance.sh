#!/usr/bin/env bash

release_platform_images() {
  printf '%s\n' backend marketplace marketplace-web relay web web-admin mobile
}

release_platform_digest() {
  local repo_root="${1:?repository root is required}"
  local image="${2:?image is required}"
  local lock="${repo_root}/deploy/kubernetes/cluster-oilan/release/kustomization.yaml"
  local digest

  digest="$(
    awk -v name="repo.aiedulab.cn:8443/agentsmesh/${image}" '
      $1 == "-" && $2 == "name:" && $3 == name { found=1; next }
      found && $1 == "digest:" { print $2; exit }
    ' "${lock}"
  )"
  [[ "${digest}" =~ ^sha256:[a-f0-9]{64}$ ]] || {
    echo "release digest missing for ${image}" >&2
    return 1
  }
  printf '%s' "${digest}"
}

release_remote_image_revision() {
  local repo_root="${1:?repository root is required}"
  local image="${2:?image is required}"
  local digest reference revision

  command -v docker >/dev/null || {
    echo "release requires docker for image provenance verification" >&2
    return 1
  }
  digest="$(release_platform_digest "${repo_root}" "${image}")"
  reference="repo.aiedulab.cn:8443/agentsmesh/${image}@${digest}"
  docker pull --platform linux/amd64 "${reference}" >/dev/null
  revision="$(
    docker image inspect "${reference}" \
      --format '{{ index .Config.Labels "org.opencontainers.image.revision" }}'
  )"
  [[ "${revision}" =~ ^[a-f0-9]{40}$ ]] || {
    echo "release image revision missing for ${image}" >&2
    return 1
  }
  printf '%s' "${revision}"
}

release_collect_platform_image_revisions() {
  local repo_root="${1:?repository root is required}"
  local image revision revisions

  revisions='{}'
  for image in $(release_platform_images); do
    revision="$(release_remote_image_revision "${repo_root}" "${image}")"
    revisions="$(
      jq -c --arg image "${image}" --arg revision "${revision}" \
        '. + {($image): $revision}' <<< "${revisions}"
    )"
  done
  printf '%s' "${revisions}"
}

release_verify_platform_image_provenance() {
  local repo_root="${1:?repository root is required}"
  shift
  local metadata image expected actual

  metadata="${repo_root}/deploy/kubernetes/cluster-oilan/release/source.json"
  if [[ "$#" -eq 0 ]]; then
    set -- $(release_platform_images)
  fi
  for image in "$@"; do
    expected="$(jq -er --arg image "${image}" '.images[$image]' "${metadata}")"
    [[ "${expected}" =~ ^[a-f0-9]{40}$ ]] || {
      echo "release metadata revision missing for ${image}" >&2
      return 1
    }
    actual="$(release_remote_image_revision "${repo_root}" "${image}")"
    [[ "${actual}" == "${expected}" ]] || {
      echo "release image provenance mismatch for ${image}: ${actual}" >&2
      return 1
    }
  done
}
