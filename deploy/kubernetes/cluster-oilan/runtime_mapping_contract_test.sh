#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"

python3 - "${REPO_ROOT}" <<'PY'
import json
import os
import pathlib
import re
import sys

root = pathlib.Path(sys.argv[1])
release = json.loads(
    (root / "docker/agent-runtime/do-agent-release.json").read_text()
)
catalog = json.loads(
    (root / "backend/internal/domain/workerruntime/runtime_catalog.lock.json").read_text()
)
manifest = (root / "deploy/kubernetes/cluster-oilan/30-backend.yaml").read_text()
seed = (root / "deploy/kubernetes/cluster-oilan/21-seed-configmap.yaml").read_text()

assert release["schema_version"] == 1
source = release["source"]
build = release["build"]
artifact = release["artifact"]
release_image = release["image"]
assert source["repository"] == "https://cnb.cool/l8ai/dowork"
assert source["branch"] == "codex/doagent-standard-acp-session-new"
assert re.fullmatch(r"[a-f0-9]{40}", source["commit"])
assert source["subdirectory"] == "doagent"
assert re.fullmatch(r"sha256:[a-f0-9]{64}", source["cargo_lock_sha256"])
assert re.fullmatch(r"sha256:[a-f0-9]{64}", source["cargo_toml_sha256"])
assert build["platform"] == "linux/amd64"
assert re.fullmatch(r"rust@sha256:[a-f0-9]{64}", build["builder_image"])
assert re.fullmatch(r"\d+\.\d+\.\d+", build["rust_toolchain"])
assert re.fullmatch(r"\d+\.\d+\.\d+", artifact["version"])
assert re.fullmatch(r"sha256:[a-f0-9]{64}", artifact["binary_sha256"])
assert release_image["repository"] == (
    "repo.aiedulab.cn:8443/agentcloud/runner-do-agent"
)
assert re.fullmatch(r"sha256:[a-f0-9]{64}", release_image["digest"])
expected_remote = os.environ.get("EXPECTED_REMOTE_DIGEST")
if expected_remote:
    assert expected_remote == release_image["digest"], (
        "remote Harbor digest does not match the trusted release manifest"
    )

matches = [
    image
    for image in catalog["images"]
    if set(image["worker_type_slugs"]) == {"do-agent", "seedance-expert"}
]
assert len(matches) == 1, "runtime catalog must publish one shared do-agent/seedance-expert image"
image = matches[0]
assert image["enabled"] is True, "do-agent runtime image must be enabled"
reference = image["reference"]
digest = image["digest"]
assert digest == release_image["digest"]
assert reference == f"{release_image['repository']}@{digest}"
assert re.fullmatch(r"sha256:[a-f0-9]{64}", digest)

mapping_line = re.search(
    r'name: COORDINATOR_RUNNER_IMAGES\s+value: "([^"]+)"',
    manifest,
)
assert mapping_line, "COORDINATOR_RUNNER_IMAGES is missing"
mappings = dict(item.split("=", 1) for item in mapping_line.group(1).split(","))
assert mappings.get("do-agent") == reference
assert mappings.get("seedance-expert") == reference
for worker_type in ("do-agent", "seedance-expert"):
    assert f"('{worker_type}')" in seed, (
        f"{worker_type} coordinator runner node_id must be pre-registered"
    )
PY
