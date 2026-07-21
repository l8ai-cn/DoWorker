#!/usr/bin/env bash

release_ci_state() {
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
  )" || return 1

  jq -er '
    [
      "Runtime release contracts",
      "Loop and sandbox security regressions",
      "Web-user artifact preview"
    ] as $required
    | [
        "Deploy test environment",
        "Deploy US West",
        "Deploy US West Relay 01",
        "Deploy US West Relay Beijing 02",
        "Migrate US West",
        "Deploy CN",
        "Deploy CN Relay 01",
        "Migrate CN"
      ] as $nonBlockingDeployments
    | [.[] | .check_runs[]] as $checks
    | [
        $checks[]
        | select(.status == "completed" and .conclusion == "success")
        | .name
      ] as $successful
    | (
        .[0].total_count != ($checks | length)
        or any(
          $checks[];
          . as $check
          | ($required | index($check.name)) != null
            and $check.status == "completed"
            and $check.conclusion != "success"
        )
        or any(
          $checks[];
          . as $check
          | ($nonBlockingDeployments | index($check.name)) == null
            and $check.status == "completed"
            and (
              $check.conclusion != "success"
              and $check.conclusion != "neutral"
              and $check.conclusion != "skipped"
            )
        )
      ) as $failed
    | (
        (($required - $successful | length) == 0)
        and all(
          $checks[];
          . as $check
          | ($nonBlockingDeployments | index($check.name)) != null
            or (
              $check.status == "completed"
              and (
                $check.conclusion == "success"
                or $check.conclusion == "neutral"
                or $check.conclusion == "skipped"
              )
            )
        )
      ) as $ready
    | if $failed then "failed"
      elif $ready then "ready"
      else "pending"
      end
  ' <<< "${checks}"
}

release_require_ci_success() {
  local head="${1:?commit is required}"
  local state

  state="$(release_ci_state "${head}")" || return 1
  [[ "${state}" == "ready" ]] || {
    echo "release requires completed successful GitHub checks for ${head}" >&2
    return 1
  }
}

release_wait_for_ci_success() {
  local head="${1:?commit is required}"
  local timeout="${RELEASE_CI_WAIT_SECONDS:-3600}"
  local interval="${RELEASE_CI_POLL_SECONDS:-15}"
  local started now state

  started="$(date +%s)"
  while true; do
    state="$(release_ci_state "${head}")" || return 1
    case "${state}" in
      ready) return 0 ;;
      failed)
        echo "release checks failed for ${head}" >&2
        return 1
        ;;
    esac
    now="$(date +%s)"
    if (( now - started >= timeout )); then
      echo "timed out waiting for release checks for ${head}" >&2
      return 1
    fi
    sleep "${interval}"
  done
}
