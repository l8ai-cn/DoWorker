#!/usr/bin/env bash

INFRA_RECOPY_REQUIRED=10
INFRA_VERIFICATION_ERROR=20

release_locked_infra_digest() {
  local image="$1"
  local release_file="${REPO_ROOT}/deploy/kubernetes/cluster-oilan/release/kustomization.yaml"
  local record kind digest
  [[ -r "${release_file}" ]] || {
    echo "  release digest file is unreadable: ${release_file}" >&2
    return "${INFRA_VERIFICATION_ERROR}"
  }
  record="$(
    awk -v name="${PROJ}/${image}" '
      $1 == "-" && $2 == "name:" {
        if (found) {
          print "missing-digest"
          done=1
          exit
        }
        found=($3 == name)
        next
      }
      found && $1 == "digest:" {
        print "digest " $2
        done=1
        exit
      }
      END {
        if (!done) {
          print found ? "missing-digest" : "missing-image"
        }
      }
    ' "${release_file}"
  )" || return "${INFRA_VERIFICATION_ERROR}"
  kind="${record%% *}"
  digest="${record#* }"
  case "${kind}" in
    digest)
      [[ "${digest}" =~ ^sha256:[a-f0-9]{64}$ ]] || {
        echo "  malformed immutable release digest for ${image}: ${digest}" >&2
        return "${INFRA_VERIFICATION_ERROR}"
      }
      printf '%s' "${digest}"
      ;;
    missing-image)
      echo "  no release entry exists for ${image}" >&2
      return "${INFRA_VERIFICATION_ERROR}"
      ;;
    *)
      echo "  release entry has no immutable digest for ${image}" >&2
      return "${INFRA_VERIFICATION_ERROR}"
      ;;
  esac
}

harbor_artifact_http_status() {
  local dest="$1" project="${PROJ#${REG}/}"
  local repository="${dest%%:*}" reference="${dest#*:}"
  local cred username password
  cred="$(harbor_creds)" || return 1
  username="$(jq -er '.Username' <<< "${cred}")" || return 1
  password="$(jq -er '.Secret' <<< "${cred}")" || return 1
  curl -sk -u "${username}:${password}" -o /dev/null -w '%{http_code}' \
    "https://${REG}/api/v2.0/projects/${project}/repositories/${repository}/artifacts/${reference}"
}

verify_infra_mirror_platforms() {
  local immutable_ref="$1" display="$2" raw status
  if ! raw="$(docker buildx imagetools inspect "${immutable_ref}" --raw)"; then
    echo "  unable to inspect immutable Harbor mirror: ${immutable_ref}" >&2
    return "${INFRA_VERIFICATION_ERROR}"
  fi
  if jq -e '
    def has_linux($arch):
      any(.manifests[]?; .platform.os == "linux" and .platform.architecture == $arch);
    (
      .mediaType == "application/vnd.oci.image.index.v1+json"
      or .mediaType == "application/vnd.docker.distribution.manifest.list.v2+json"
    )
    and has_linux("amd64")
    and has_linux("arm64")
  ' <<< "${raw}" >/dev/null; then
    return 0
  else
    status=$?
  fi
  [[ "${status}" -eq 1 ]] || {
    echo "  invalid Harbor manifest JSON for ${display}" >&2
    return "${INFRA_VERIFICATION_ERROR}"
  }
  echo "  Harbor mirror is not an amd64/arm64 image index: ${display}" >&2
  return "${INFRA_RECOPY_REQUIRED}"
}

verified_locked_infra_digest() {
  local dest="$1" image="${1%%:*}"
  local locked status remote
  if locked="$(release_locked_infra_digest "${image}")"; then
    :
  else
    status=$?
    return "${status}"
  fi

  if ! status="$(harbor_artifact_http_status "${dest}")"; then
    echo "  Harbor artifact query failed for ${dest}" >&2
    return "${INFRA_VERIFICATION_ERROR}"
  fi
  case "${status}" in
    200) ;;
    404)
      echo "  Harbor artifact is absent: ${dest}" >&2
      return "${INFRA_RECOPY_REQUIRED}"
      ;;
    *)
      echo "  Harbor artifact query returned HTTP ${status}: ${dest}" >&2
      return "${INFRA_VERIFICATION_ERROR}"
      ;;
  esac

  if ! remote="$(manifest_digest "${PROJ}/${dest}")"; then
    return "${INFRA_VERIFICATION_ERROR}"
  fi
  [[ "${remote}" == "${locked}" ]] || {
    echo "  Harbor digest changed for ${dest}: ${remote} != ${locked}" >&2
    return "${INFRA_RECOPY_REQUIRED}"
  }
  if verify_infra_mirror_platforms "${PROJ}/${image}@${remote}" "${dest}"; then
    :
  else
    status=$?
    return "${status}"
  fi
  printf '%s' "${remote}"
}

set_infra_digest() {
  local image="$1" digest="$2"
  case "${image}" in
    pgvector) PLATFORM_DIGEST_PGVECTOR="${digest}" ;;
    redis) PLATFORM_DIGEST_REDIS="${digest}" ;;
    minio) PLATFORM_DIGEST_MINIO="${digest}" ;;
    mc) PLATFORM_DIGEST_MC="${digest}" ;;
    kubectl) PLATFORM_DIGEST_KUBECTL="${digest}" ;;
    *)
      echo "unsupported infrastructure image: ${image}" >&2
      return 1
      ;;
  esac
}

mirror() {
  local src="$1" dest="$2" image="${2%%:*}" digest status
  if digest="$(verified_locked_infra_digest "${dest}")"; then
    echo "==> reuse locked Harbor mirror ${PROJ}/${dest}@${digest}"
    set_infra_digest "${image}" "${digest}"
    return
  else
    status=$?
  fi
  [[ "${status}" -eq "${INFRA_RECOPY_REQUIRED}" ]] || return "${status}"

  echo "==> mirror ${src} -> ${PROJ}/${dest} (multi-arch index)"
  local n=1
  until docker buildx imagetools create --tag "${PROJ}/${dest}" "${src}"; do
    [[ "${n}" -ge 4 ]] && {
      echo "  imagetools failed after 4 tries: ${src}" >&2
      return 1
    }
    echo "  retry ${n}/4 in 8s..." >&2
    sleep 8
    n=$((n + 1))
  done
  digest="$(manifest_digest "${PROJ}/${dest}")" || return 1
  verify_infra_mirror_platforms "${PROJ}/${image}@${digest}" "${dest}" || return 1
  set_infra_digest "${image}" "${digest}"
}
