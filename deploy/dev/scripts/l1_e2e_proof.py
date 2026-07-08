#!/usr/bin/env python3
"""L1 real-agent E2E: create sessions, send prompts, verify file output."""

from __future__ import annotations

import json
import sys
import time
import urllib.error
import urllib.request

API = "http://localhost:10015"
ORG = "dev-org"
ENV = "default"
POLL_INTERVAL = 3
TIMEOUT = 180


def req(method: str, path: str, token: str | None = None, body: dict | None = None) -> dict:
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
        headers["X-Organization-Slug"] = ORG
    data = json.dumps(body).encode() if body is not None else None
    r = urllib.request.Request(f"{API}{path}", data=data, headers=headers, method=method)
    with urllib.request.urlopen(r, timeout=60) as resp:
        raw = resp.read()
        if not raw:
            return {}
        return json.loads(raw)


def login() -> str:
    r = req(
        "POST",
        "/auth/login",
        None,
        {"username": "devuser", "password": "AdminAb123456"},
    )
    return r["token"]


def create_session(token: str, agent_id: str, host_id: str, title: str, extra: dict | None = None) -> str:
    body = {"agent_id": agent_id, "host_id": host_id, "title": title}
    if extra:
        body.update(extra)
    r = req("POST", "/v1/sessions", token, body)
    sid = r.get("id") or r.get("session_id")
    if not sid:
        raise RuntimeError(f"create session failed: {r}")
    return sid


def send_message(token: str, sid: str, text: str) -> None:
    req(
        "POST",
        f"/v1/sessions/{sid}/events",
        token,
        {"type": "message", "data": {"content": [{"type": "input_text", "text": text}]}},
    )


def list_items(token: str, sid: str) -> list[dict]:
    r = req("GET", f"/v1/sessions/{sid}/items", token)
    return r.get("data") or r.get("items") or []


def wait_turn(token: str, sid: str) -> tuple[bool, str]:
    deadline = time.time() + TIMEOUT
    last_assistant = ""
    while time.time() < deadline:
        items = list_items(token, sid)
        for item in reversed(items):
            itype = item.get("type")
            status = item.get("status")
            if itype == "error":
                return False, item.get("message") or json.dumps(item)
            if itype in ("message", "assistant_message") and item.get("role") == "assistant":
                if status == "completed":
                    content = item.get("content") or []
                    parts = []
                    for block in content:
                        if isinstance(block, dict) and block.get("text"):
                            parts.append(block["text"])
                    last_assistant = "\n".join(parts)
                    return True, last_assistant
                if status == "in_progress":
                    break
        time.sleep(POLL_INTERVAL)
    return False, f"timeout after {TIMEOUT}s; last={last_assistant!r}"


def read_file(token: str, sid: str, path: str) -> str | None:
    try:
        r = req(
            "GET",
            f"/v1/sessions/{sid}/resources/environments/{ENV}/filesystem/{path}",
            token,
        )
        return r.get("content") or r.get("text")
    except urllib.error.HTTPError:
        return None


def find_file(token: str, sid: str, name: str) -> tuple[str | None, str | None]:
    for path in (name, f"workspace/{name}"):
        content = read_file(token, sid, path)
        if content and content.strip():
            return path, content.strip()
    listing = req(
        "GET",
        f"/v1/sessions/{sid}/resources/environments/{ENV}/filesystem?limit=200",
        token,
    )
    entries = listing.get("data") or listing.get("entries") or []
    for entry in entries:
        if isinstance(entry, dict) and entry.get("name") == name:
            p = entry.get("path") or name
            content = read_file(token, sid, p)
            if content:
                return p, content.strip()
    return None, None


def run_case(
    token: str,
    label: str,
    agent_id: str,
    host_id: str,
    prompt: str,
    filename: str,
    expected: str,
    extra: dict | None = None,
) -> dict:
    print(f"\n=== {label} ===")
    sid = create_session(token, agent_id, host_id, f"L1 {label}", extra)
    url = f"http://127.0.0.1:10020/c/{sid}"
    print(f"session: {sid}")
    print(f"url: {url}")
    send_message(token, sid, prompt)
    ok, detail = wait_turn(token, sid)
    print(f"turn: {'OK' if ok else 'FAIL'} — {detail[:200]}")
    path, content = find_file(token, sid, filename)
    file_ok = content == expected
    print(f"file: {path!r} content={content!r} expected={expected!r} -> {'OK' if file_ok else 'FAIL'}")
    return {
        "label": label,
        "session_id": sid,
        "url": url,
        "turn_ok": ok,
        "file_ok": file_ok,
        "path": path,
        "content": content,
        "detail": detail,
    }


def main() -> int:
    token = login()
    results = []

    results.append(
        run_case(
            token,
            "Codex",
            "codex-cli",
            "host_dev-runner-codex",
            "Create a file named codex-e2e-proof.txt in the workspace root with exactly this single line of content: E2E codex real task ok. Do not add anything else.",
            "codex-e2e-proof.txt",
            "E2E codex real task ok.",
        )
    )

    results.append(
        run_case(
            token,
            "Do-agent",
            "do-agent",
            "host_dev-runner-do-agent",
            "Run this exact shell command in the workspace: printf 'E2E do-agent real task ok.' > do-agent-e2e-proof.txt",
            "do-agent-e2e-proof.txt",
            "E2E do-agent real task ok.",
            {"model_config_id": 3},
        )
    )

    print("\n=== SUMMARY ===")
    all_ok = True
    for r in results:
        ok = r["turn_ok"] and r["file_ok"]
        all_ok = all_ok and ok
        print(f"{r['label']}: {'PASS' if ok else 'FAIL'} — {r['url']}")
    return 0 if all_ok else 1


if __name__ == "__main__":
    sys.exit(main())
