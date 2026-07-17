#!/usr/bin/env bash

release_platform_images() {
  printf '%s\n' backend marketplace marketplace-web relay web web-admin mobile
}

release_runtime_images() {
  printf '%s\n' runner-do-agent runner-video-studio
}

release_images() {
  release_platform_images
  release_runtime_images
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

release_runtime_digest() {
  local repo_root="${1:?repository root is required}"
  local image="${2:?image is required}"
  local digest

  case "${image}" in
    runner-do-agent)
      digest="$(
        jq -er '
          select(.image.repository == "repo.aiedulab.cn:8443/agentsmesh/runner-do-agent")
          | .image.digest
        ' "${repo_root}/docker/agent-runtime/do-agent-release.json"
      )"
      ;;
    runner-video-studio)
      digest="$(
        jq -er '
          .images[]
          | select(.slug == "video-studio-stable" and .enabled == true)
          | .digest as $digest
          | select(.reference == ("repo.aiedulab.cn:8443/agentsmesh/runner-video-studio@" + $digest))
          | $digest
        ' "${repo_root}/backend/internal/domain/workerruntime/runtime_catalog.lock.json"
      )"
      ;;
    *)
      echo "unknown release runtime image: ${image}" >&2
      return 1
      ;;
  esac
  [[ "${digest}" =~ ^sha256:[a-f0-9]{64}$ ]] || {
    echo "release digest missing for ${image}" >&2
    return 1
  }
  printf '%s' "${digest}"
}

release_image_digest() {
  local repo_root="${1:?repository root is required}"
  local image="${2:?image is required}"

  case "${image}" in
    runner-do-agent|runner-video-studio)
      release_runtime_digest "${repo_root}" "${image}"
      ;;
    *)
      release_platform_digest "${repo_root}" "${image}"
      ;;
  esac
}

release_remote_image_revision() {
  local repo_root="${1:?repository root is required}"
  local image="${2:?image is required}"
  local digest reference revision platform

  command -v docker >/dev/null || {
    echo "release requires docker for image provenance verification" >&2
    return 1
  }
  digest="$(release_image_digest "${repo_root}" "${image}")"
  reference="repo.aiedulab.cn:8443/agentsmesh/${image}@${digest}"
  docker pull "${reference}" >/dev/null || return 1
  platform="$(
    docker image inspect "${reference}" --format '{{.Os}}/{{.Architecture}}'
  )" || return 1
  [[ "${platform}" == "linux/amd64" ]] || {
    echo "release image platform mismatch for ${image}: ${platform}" >&2
    return 1
  }
  revision="$(
    docker image inspect "${reference}" \
      --format '{{ index .Config.Labels "org.opencontainers.image.revision" }}'
  )" || return 1
  [[ "${revision}" =~ ^[a-f0-9]{40}$ ]] || {
    echo "release image revision missing for ${image}" >&2
    return 1
  }
  printf '%s' "${revision}"
}

release_collect_image_revisions() {
  local repo_root="${1:?repository root is required}"
  local image revision revisions

  revisions='{}'
  for image in $(release_images); do
    revision="$(release_remote_image_revision "${repo_root}" "${image}")" || return 1
    revisions="$(
      jq -c --arg image "${image}" --arg revision "${revision}" \
        '. + {($image): $revision}' <<< "${revisions}"
    )" || return 1
  done
  printf '%s' "${revisions}"
}

release_verify_image_provenance() {
  local repo_root="${1:?repository root is required}"
  shift
  local metadata image expected actual

  metadata="${repo_root}/deploy/kubernetes/cluster-oilan/release/source.json"
  if [[ "$#" -eq 0 ]]; then
    set -- $(release_images)
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

release_verify_gitea_provenance() {
  local repo_root="${1:?repository root is required}"
  local lock="${repo_root}/deploy/kubernetes/cluster-oilan/release/kustomization.yaml"
  local digest reference platform version user_commands

  digest="$(
    awk '
      $1 == "-" && $2 == "name:" &&
        $3 == "repo.aiedulab.cn:8443/library/gitea" { found=1; next }
      found && $1 == "digest:" { print $2; exit }
    ' "${lock}"
  )"
  [[ "${digest}" =~ ^sha256:[a-f0-9]{64}$ ]] || {
    echo "release digest missing for library/gitea" >&2
    return 1
  }
  reference="repo.aiedulab.cn:8443/library/gitea@${digest}"
  docker pull "${reference}" >/dev/null || return 1
  platform="$(
    docker image inspect "${reference}" --format '{{.Os}}/{{.Architecture}}'
  )" || return 1
  [[ "${platform}" == "linux/amd64" ]] || {
    echo "release image platform mismatch for library/gitea: ${platform}" >&2
    return 1
  }
  version="$(
    docker run --rm --platform linux/amd64 --entrypoint /usr/local/bin/gitea \
      "${reference}" --version
  )" || return 1
  [[ "${version}" == "Gitea version 1.21.0 "* ]] || {
    echo "release Gitea version mismatch: ${version}" >&2
    return 1
  }
  user_commands="$(
    docker run --rm --platform linux/amd64 --entrypoint /usr/local/bin/gitea \
      "${reference}" admin user --help
  )" || return 1
  grep -Fq 'generate-access-token' <<<"${user_commands}" || {
    echo "release Gitea image lacks admin token generation support" >&2
    return 1
  }
}
