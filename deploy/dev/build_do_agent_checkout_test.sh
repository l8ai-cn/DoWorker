#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
source "$ROOT/deploy/dev/lib/build_do_agent_binary.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

git -C "$tmp" init -q source
git -C "$tmp/source" config user.name test
git -C "$tmp/source" config user.email test@example.com
touch "$tmp/source/README"
git -C "$tmp/source" add README
git -C "$tmp/source" commit -qm initial
git -C "$tmp/source" worktree add -q "$tmp/linked"

is_do_agent_checkout "$tmp/source"
is_do_agent_checkout "$tmp/linked"
if is_do_agent_checkout "$tmp/plain"; then
    echo "plain directory must not be accepted as a checkout" >&2
    exit 1
fi
