#!/usr/bin/env bash

RELEASE_GUARD_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=release_image_provenance.sh
source "${RELEASE_GUARD_DIR}/release_image_provenance.sh"

release_require_pushed_clean_tree() {
  local repo_root="${1:?repository root is required}"
  local branch head remote_head required_branch

  [[ -z "$(git -C "${repo_root}" status --porcelain --untracked-files=normal)" ]] || {
    echo "release requires a clean working tree: ${repo_root}" >&2
    return 1
  }

  branch="$(git -C "${repo_root}" branch --show-current)"
  [[ -n "${branch}" ]] || {
    echo "release requires a named branch" >&2
    return 1
  }
  required_branch="${RELEASE_BRANCH:-main}"
  [[ "${branch}" == "${required_branch}" ]] || {
    echo "release requires branch ${required_branch}, got ${branch}" >&2
    return 1
  }

  git -C "${repo_root}" fetch --quiet origin "${branch}"
  head="$(git -C "${repo_root}" rev-parse HEAD)"
  remote_head="$(git -C "${repo_root}" rev-parse "refs/remotes/origin/${branch}")"
  [[ "${head}" == "${remote_head}" ]] || {
    echo "release requires HEAD ${head} to be visible at origin/${branch}" >&2
    return 1
  }

  release_require_ci_success "${head}" || return 1
  export RELEASE_SOURCE_COMMIT="${head}"
}

release_require_ci_success() {
  local head="${1:?commit is required}"
  local repository="${RELEASE_REPOSITORY:-l8ai-cn/DoWorker}"
  local checks

  command -v gh >/dev/null || {
    echo "release requires gh for CI verification" >&2
    return 1
  }
  command -v jq >/dev/null || {
    echo "release requires jq for CI verification" >&2
    return 1
  }
  checks="$(
    gh api --paginate --slurp \
      "repos/${repository}/commits/${head}/check-runs?per_page=100"
  )"
  jq -e '
    [
      "Runtime release contracts",
      "Loop and sandbox security regressions",
      "Web-user artifact preview"
    ] as $required
    | [
        "Deploy US West",
        "Deploy US West Relay 01",
        "Deploy US West Relay Beijing 02",
        "Migrate US West",
        "Deploy CN",
        "Deploy CN Relay 01",
        "Migrate CN"
      ] as $externalDeployments
    | [.[] | .check_runs[]] as $checks
    | [
        $checks[]
        | select(
            .status == "completed"
            and .conclusion == "success"
          )
        | .name
      ] as $successful
    | (.[0].total_count == ($checks | length))
    and (($required - $successful | length) == 0)
    and all(
      $checks[];
      . as $check
      | ($externalDeployments | index($check.name)) != null
        or (
          $check.status == "completed"
          and (
            $check.conclusion == "success"
            or $check.conclusion == "neutral"
            or $check.conclusion == "skipped"
          )
        )
    )
  ' <<< "${checks}" >/dev/null || {
    echo "release requires completed successful GitHub checks for ${head}" >&2
    return 1
  }
}

release_write_source_metadata() {
  local repo_root="${1:?repository root is required}"
  local output="${repo_root}/deploy/kubernetes/cluster-oilan/release/source.json"
  local temporary="${output}.tmp"
  local source_commit="${RELEASE_SOURCE_COMMIT:?release source commit is required}"
  local image_revisions

  image_revisions="$(release_collect_platform_image_revisions "${repo_root}")"
  mkdir -p "$(dirname "${output}")"
  jq -n \
    --arg branch "${RELEASE_BRANCH:-main}" \
    --arg commit "${source_commit}" \
    --argjson images "${image_revisions}" \
    '{branch: $branch, commit: $commit, images: $images}' > "${temporary}"
  mv "${temporary}" "${output}"
}

release_verify_source_metadata() {
  local repo_root="${1:?repository root is required}"
  local metadata="${repo_root}/deploy/kubernetes/cluster-oilan/release/source.json"
  local branch commit image revision

  branch="$(jq -er '.branch' "${metadata}")"
  commit="$(jq -er '.commit' "${metadata}")"
  [[ "${branch}" == "${RELEASE_BRANCH:-main}" ]] || {
    echo "release metadata branch mismatch: ${branch}" >&2
    return 1
  }
  [[ "${commit}" =~ ^[a-f0-9]{40}$ ]] || {
    echo "release metadata contains an invalid source commit" >&2
    return 1
  }
  git -C "${repo_root}" merge-base --is-ancestor "${commit}" HEAD || {
    echo "release image source ${commit} is not an ancestor of HEAD" >&2
    return 1
  }
  for image in $(release_platform_images); do
    revision="$(jq -er --arg image "${image}" '.images[$image]' "${metadata}")"
    [[ "${revision}" =~ ^[a-f0-9]{40}$ ]] || {
      echo "release metadata revision missing for ${image}" >&2
      return 1
    }
    git -C "${repo_root}" merge-base --is-ancestor "${revision}" HEAD || {
      echo "release image source ${revision} for ${image} is not an ancestor of HEAD" >&2
      return 1
    }
  done
}
