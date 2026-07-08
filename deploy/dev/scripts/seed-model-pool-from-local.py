#!/usr/bin/env python3
"""Seed dev-org model pool via REST API (same path as web-user Settings / Models).

Reads local developer credentials and POSTs missing pool rows; never touches
backend startup or deploy .env for API keys.
"""

from __future__ import annotations

import json
import os
import sys
import urllib.error
import urllib.request

REPO = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", ".."))
DEV_ENV = os.path.join(REPO, "deploy", "dev", ".env")
ORG = "dev-org"
USER = "devuser"
PASS = "AdminAb123456"


def load_backend_port() -> int:
    if os.path.isfile(DEV_ENV):
        for line in open(DEV_ENV):
            if line.startswith("BACKEND_HTTP_PORT="):
                return int(line.strip().split("=", 1)[1])
    return 10015


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
        "is_default": True,
        "scope": "org",
    }


def load_minimax_cred() -> dict | None:
    path = os.path.expanduser("~/.config/sub2api/minimax.env")
    if not os.path.isfile(path):
        return None
    vals = {}
    for line in open(path):
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        k, v = line.split("=", 1)
        vals[k.strip()] = v.strip().strip('"\'')
    key = vals.get("MINIMAX_API_KEY", "").strip()
    if not key:
        return None
    return {
        "name": "MiniMax",
        "provider_type": "minimax",
        "model": vals.get("MINIMAX_MODEL", "MiniMax-M3"),
        "base_url": vals.get("MINIMAX_BASE_URL", "https://api.minimax.chat/anthropic"),
        "credentials": {"api_key": key},
        "is_default": False,
        "scope": "org",
    }


def has_provider(models: list, provider: str) -> bool:
    return any(m.get("provider_type") == provider for m in models)


def main() -> int:
    try:
        token = api("POST", "/auth/login", None, {"username": USER, "password": PASS})["token"]
    except urllib.error.URLError as e:
        print(f"backend not reachable: {e}", file=sys.stderr)
        return 1

    models = api("GET", "/v1/model-configs", token).get("data", [])
    created = 0
    for loader in (load_codex_cred, load_minimax_cred):
        spec = loader()
        if not spec:
            continue
        if has_provider(models, spec["provider_type"]):
            print(f"skip {spec['provider_type']}: already in pool")
            continue
        api("POST", "/v1/model-configs", token, spec)
        print(f"created {spec['provider_type']} -> {spec['name']}")
        created += 1

    if created == 0 and not models:
        print("no local credentials found and pool empty", file=sys.stderr)
        return 1
    print("done")
    return 0


if __name__ == "__main__":
    sys.exit(main())
