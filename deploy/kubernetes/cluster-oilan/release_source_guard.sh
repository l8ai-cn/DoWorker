#!/usr/bin/env bash

release_require_pushed_clean_tree() {
  local repo_root="${1:?repository root is required}"
  local branch head remote_head

  [[ -z "$(git -C "${repo_root}" status --porcelain --untracked-files=normal)" ]] || {
    echo "release requires a clean working tree: ${repo_root}" >&2
    return 1
  }

  branch="$(git -C "${repo_root}" branch --show-current)"
  [[ -n "${branch}" ]] || {
    echo "release requires a named branch" >&2
    return 1
  }

  git -C "${repo_root}" fetch --quiet origin "${branch}"
  head="$(git -C "${repo_root}" rev-parse HEAD)"
  remote_head="$(git -C "${repo_root}" rev-parse "refs/remotes/origin/${branch}")"
  [[ "${head}" == "${remote_head}" ]] || {
    echo "release requires HEAD ${head} to be visible at origin/${branch}" >&2
    return 1
  }
}
