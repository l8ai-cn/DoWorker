#!/usr/bin/env python3
"""Real Do-agent gomoku on admin-workspace (MiniMax model pool id=2)."""

from __future__ import annotations

import json
import sys
import time
import urllib.error
import urllib.request

API = "http://127.0.0.1:10015"
ORG = "admin-workspace"
TIMEOUT = int(sys.argv[1]) if len(sys.argv) > 1 else 900

PROMPT = """请作为多 Agent 团队负责人，用 Do-agent 的多 Agent 能力完成五子棋项目：

1. 先用 TodoWrite 创建 3-5 个可验收子目标（UI、棋盘逻辑、胜负判定、交互、README）
2. 对复杂子任务使用 Agent 工具派生子 Agent 分工实现
3. 在 workspace 生成可运行五子棋：gomoku2/index.html（15x15，黑白轮流，五连珠判胜，悔棋，中文标题「五子棋」）+ gomoku2/README.md

先输出目标清单，再写文件，中文简要汇报进度。"""


def req(method: str, path: str, token: str | None = None, body: dict | None = None) -> dict:
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
        headers["X-Organization-Slug"] = ORG
    data = json.dumps(body).encode() if body is not None else None
    r = urllib.request.Request(f"{API}{path}", data=data, headers=headers, method=method)
    with urllib.request.urlopen(r, timeout=120) as resp:
        raw = resp.read()
        return json.loads(raw) if raw else {}


def login() -> str:
    return req("POST", "/auth/login", None, {"username": "admin", "password": "Ab123456"})["token"]


def read_gomoku(token: str, sid: str) -> bool:
    for path in ("gomoku2/index.html", "workspace/gomoku2/index.html"):
        try:
            body = req("GET", f"/v1/sessions/{sid}/resources/environments/default/filesystem/{path}", token)
            html = body.get("content") or body.get("text") or ""
            if html.strip() and ("五子棋" in html or "canvas" in html.lower()):
                return True
        except urllib.error.HTTPError:
            continue
    return False


def main() -> int:
    token = login()
    sid = req(
        "POST",
        "/v1/sessions",
        token,
        {
            "agent_id": "do-agent",
            "host_id": "host_admin-workspace-do-agent",
            "title": "Gomoku Do-agent REAL",
            "workspace": "/workspace",
            "model_config_id": 2,
        },
    ).get("id")
    if not sid:
        print("FAIL: create session", file=sys.stderr)
        return 1
    preview = f"http://127.0.0.1:10020/c/{sid}?file=gomoku2%2Findex.html"
    print(f"session: {sid}")
    print(f"preview: {preview}")
    req(
        "POST",
        f"/v1/sessions/{sid}/events",
        token,
        {"type": "message", "data": {"content": [{"type": "input_text", "text": PROMPT}]}},
    )
    deadline = time.time() + TIMEOUT
    while time.time() < deadline:
        sess = req("GET", f"/v1/sessions/{sid}", token)
        status = sess.get("status")
        items = req("GET", f"/v1/sessions/{sid}/items", token).get("data", [])
        n = len(items)
        if status == "failed":
            err = next((i.get("message") for i in reversed(items) if i.get("type") == "error"), None)
            print(f"FAIL: {err}", file=sys.stderr)
            return 1
        if read_gomoku(token, sid) and status == "idle" and n > 2:
            print(f"PASS: gomoku ready ({n} items)")
            print(f"PREVIEW: {preview}")
            return 0
        print(f"  status={status} items={n} usage={sess.get('usage_by_model')}")
        time.sleep(20)
    print("FAIL: timeout", file=sys.stderr)
    return 1


if __name__ == "__main__":
    sys.exit(main())
