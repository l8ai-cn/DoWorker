#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
WORKFLOW="${ROOT}/.github/workflows/deploy.yml"

grep -Fq "runs-on: [self-hosted, deploy]" "${WORKFLOW}"
grep -Fq "checks: read" "${WORKFLOW}"
grep -Fq "GH_TOKEN: \${{ github.token }}" "${WORKFLOW}"
grep -Fq "fetch-depth: 0" "${WORKFLOW}"
grep -Fq "release_wait_for_ci_success \"\${GITHUB_SHA}\"" "${WORKFLOW}"
grep -Fq "bash deploy/kubernetes/cluster-oilan/deploy.sh" "${WORKFLOW}"
grep -Fq "predicate-quantifier: every" "${WORKFLOW}"
grep -Fq "'deploy/kubernetes/cluster-oilan/**'" "${WORKFLOW}"
grep -Fq "'!**/*_test.sh'" "${WORKFLOW}"
grep -Fq "'!**/*.md'" "${WORKFLOW}"
[[ "$(grep -c '^    name: Deploy test environment$' "${WORKFLOW}")" == "1" ]]

if grep -Fq "push-images.sh" "${WORKFLOW}"; then
  echo "deploy workflow must reconcile committed release state, not build it" >&2
  exit 1
fi

if grep -Eq "'(backend|marketplace|relay|clients|packages|docker)/\\*\\*'" "${WORKFLOW}"; then
  echo "source changes must not deploy before release locks are committed" >&2
  exit 1
fi
