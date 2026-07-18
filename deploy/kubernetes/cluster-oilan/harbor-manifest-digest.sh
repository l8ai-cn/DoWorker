#!/usr/bin/env bash

platform_manifest_digest() {
  local image="$1" manifest="" digest="" attempt=1
  local os="${PLATFORM%%/*}" platform_tail="${PLATFORM#*/}" architecture variant
  architecture="${platform_tail%%/*}"
  variant=""
  [[ "${platform_tail}" == */* ]] && variant="${platform_tail#*/}"
  until manifest="$(docker buildx imagetools inspect "${image}" --format '{{json .Manifest}}')" &&
    digest="$(
      jq -er \
        --arg os "${os}" \
        --arg architecture "${architecture}" \
        --arg variant "${variant}" '
        if has("manifests") then
          [
            .manifests[]
            | select(
                .platform.os == $os and
                .platform.architecture == $architecture and
                ($variant == "" or .platform.variant == $variant)
              )
          ][0].digest
        else
          .digest
        end
      ' <<<"${manifest}"
    )" &&
    [[ "${digest}" =~ ^sha256:[a-f0-9]{64}$ ]]; do
    [[ "${attempt}" -ge 4 ]] && {
      echo "invalid registry digest for ${image}: ${digest}" >&2
      return 1
    }
    echo "  registry manifest query failed; retry ${attempt}/4..." >&2
    sleep 3
    attempt=$((attempt + 1))
  done
  printf '%s' "${digest}"
}

infra_manifest_digest() {
  local image="$1" manifest="" digest="" attempt=1
  until manifest="$(docker buildx imagetools inspect "${image}" --format '{{json .Manifest}}')" &&
    digest="$(jq -er '.digest' <<<"${manifest}")" &&
    [[ "${digest}" =~ ^sha256:[a-f0-9]{64}$ ]]; do
    [[ "${attempt}" -ge 4 ]] && {
      echo "invalid registry index digest for ${image}: ${digest}" >&2
      return 1
    }
    echo "  registry index query failed; retry ${attempt}/4..." >&2
    sleep 3
    attempt=$((attempt + 1))
  done
  printf '%s' "${digest}"
}

promote_platform_manifest() {
  local tag="$1" source="$2" attempt=1
  until docker buildx imagetools create \
    --prefer-index=false \
    --tag "${tag}" \
    "${source}"; do
    [[ "${attempt}" -ge 4 ]] && return 1
    echo "  platform manifest promotion failed; retry ${attempt}/4 in 5s..." >&2
    sleep 5
    attempt=$((attempt + 1))
  done
}
