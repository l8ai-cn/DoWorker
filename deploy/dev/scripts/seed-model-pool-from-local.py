#!/usr/bin/env python3
"""Seed dev-org model pool via REST API (same path as web-user Settings / Models).

Reads credentials from deploy/dev/secrets/minimax.env (preferred) or local
developer paths; never commits API keys to git.
"""

from __future__ import annotations

import json
import os
import subprocess
import sys
import urllib.error
import urllib.request

REPO = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", ".."))
DEV_ENV = os.path.join(REPO, "deploy", "dev", ".env")
DEV_MINIMAX = os.path.join(REPO, "deploy", "dev", "secrets", "minimax.env")
ORG = "dev-org"
USER = "devuser"
PASS = "AdminAb123456"


def load_backend_port() -> int:
    if os.path.isfile(DEV_ENV):
        for line in open(DEV_ENV):
            if line.startswith("BACKEND_HTTP_PORT="):
                return int(line.strip().split("=", 1)[1])
    return 10015


def load_compose_project() -> str:
    if os.path.isfile(DEV_ENV):
        for line in open(DEV_ENV):
            if line.startswith("COMPOSE_PROJECT_NAME="):
                return line.strip().split("=", 1)[1]
    return "agentsmesh-main"


def api(method: str, path: str, token: str | None = None, body: dict | None = None) -> dict:
    port = load_backend_port()
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
        headers["X-Organization-Slug"] = ORG
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(f"http://127.0.0.1:{port}{path}", data=data, headers=headers, method=method)
    with urllib.request.urlopen(req, timeout=60) as resp:
        raw = resp.read()
        return json.loads(raw) if raw else {}


def parse_env_file(path: str) -> dict[str, str]:
    vals: dict[str, str] = {}
    if not os.path.isfile(path):
        return vals
    for line in open(path):
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        k, v = line.split("=", 1)
        vals[k.strip()] = v.strip().strip('"\'')
    return vals


def load_codex_cred() -> dict | None:
    home = os.path.expanduser("~")
    auth_path = os.path.join(home, ".codex", "auth.json")
    if not os.path.isfile(auth_path):
        return None
    doc = json.load(open(auth_path))
    key = (doc.get("OPENAI_API_KEY") or "").strip()
    if not key:
        return None
    cfg = os.path.join(home, ".codex", "config.toml")
    base_url, model = "", ""
    if os.path.isfile(cfg):
        for line in open(cfg):
            line = line.strip()
            if line.startswith("model ") and "=" in line:
                model = line.split("=", 1)[1].strip().strip('"\'')
            if "base_url" in line and "=" in line:
                base_url = line.split("=", 1)[1].strip().strip('"\'')
    return {
        "name": "OpenAI (Codex)",
        "provider_type": "openai",
        "model": model or "gpt-5.5",
        "base_url": base_url,
        "credentials": {"api_key": key},
        "is_default": False,
        "scope": "org",
    }


def load_minimax_cred() -> dict | None:
    for path in (DEV_MINIMAX, os.path.expanduser("~/.config/sub2api/minimax.env")):
        vals = parse_env_file(path)
        key = vals.get("MINIMAX_API_KEY", "").strip()
        if key.startswith("minimax "):
            key = key.split(" ", 1)[1].strip()
        if not key:
            continue
        is_default = vals.get("MINIMAX_IS_DEFAULT", "true").lower() in ("1", "true", "yes")
        return {
            "name": "MiniMax (System)",
            "provider_type": "minimax",
            "model": vals.get("MINIMAX_MODEL", "MiniMax-M3"),
            "base_url": vals.get("MINIMAX_BASE_URL", "https://api.minimax.chat/anthropic"),
            "credentials": {"api_key": key},
            "is_default": is_default,
            "scope": "org",
        }
    return None


def find_provider(models: list, provider: str) -> dict | None:
    for m in models:
        if m.get("provider_type") == provider:
            return m
    return None


def promote_minimax_default() -> None:
    project = load_compose_project()
    container = f"{project}-postgres-1"
    sql = """
UPDATE ai_models SET is_default = false WHERE organization_id = 1;
UPDATE ai_models SET is_default = true
WHERE organization_id = 1 AND provider_type = 'minimax' AND is_enabled = true;
"""
    try:
        subprocess.run(
            ["docker", "exec", container, "psql", "-U", "agentsmesh", "-d", "agentsmesh", "-c", sql],
            check=True,
            capture_output=True,
            text=True,
        )
        print("promoted minimax to org default in ai_models")
    except (subprocess.CalledProcessError, FileNotFoundError) as e:
        print(f"warn: could not promote minimax default via postgres: {e}", file=sys.stderr)


def main() -> int:
    try:
        token = api("POST", "/auth/login", None, {"username": USER, "password": PASS})["token"]
    except urllib.error.URLError as e:
        print(f"backend not reachable: {e}", file=sys.stderr)
        return 1

    models = api("GET", "/v1/model-configs", token).get("data", [])
    created = 0

    minimax = load_minimax_cred()
    if minimax:
        existing = find_provider(models, "minimax")
        if existing:
            print(f"skip minimax: already in pool (id={existing.get('id')})")
            if minimax.get("is_default") and not existing.get("is_default"):
                promote_minimax_default()
        else:
            api("POST", "/v1/model-configs", token, minimax)
            print(f"created minimax -> {minimax['name']} (is_default={minimax['is_default']})")
            created += 1
            models = api("GET", "/v1/model-configs", token).get("data", [])

    for loader in (load_codex_cred,):
        spec = loader()
        if not spec:
            continue
        if find_provider(models, spec["provider_type"]):
            print(f"skip {spec['provider_type']}: already in pool")
            continue
        api("POST", "/v1/model-configs", token, spec)
        print(f"created {spec['provider_type']} -> {spec['name']}")
        created += 1

    if minimax and minimax.get("is_default"):
        promote_minimax_default()

    if created == 0 and not models and not minimax:
        print("no local credentials found and pool empty", file=sys.stderr)
        return 1
    print("done")
    return 0


if __name__ == "__main__":
    sys.exit(main())
