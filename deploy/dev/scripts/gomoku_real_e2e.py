#!/usr/bin/env python3
"""Real-agent gomoku E2E: Codex-cli + Do-agent (MiniMax), no e2e-echo mock."""

from __future__ import annotations

import json
import sys
import time
import urllib.error
import urllib.request

API = "http://127.0.0.1:10015"
ORG = "dev-org"
TIMEOUT = int(sys.argv[1]) if len(sys.argv) > 1 else 900

GOMOKU_PROMPT = """请在 workspace 目录创建一个可浏览器运行的五子棋小游戏：
- gomoku/index.html：15×15 棋盘，黑白轮流，五连珠判胜，悔棋，单文件 HTML+CSS+JS，中文界面标题「五子棋」
- gomoku/README.md：简要说明规则与如何打开
完成后用一两句话总结。"""

DO_AGENT_PROMPT = """请作为多 Agent 团队负责人，用 Do-agent 的多 Agent 能力完成五子棋项目：

1. 先用 TodoWrite 创建 3-5 个可验收子目标（UI、棋盘逻辑、胜负判定、交互、README）
2. 对复杂子任务使用 Agent 工具派生子 Agent 分工实现
3. 在 workspace 生成可运行五子棋：gomoku/index.html（15x15，黑白轮流，五连珠判胜，悔棋）+ gomoku/README.md

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
    return req("POST", "/auth/login", None, {"username": "devuser", "password": "AdminAb123456"})["token"]


def create_session(token: str, agent_id: str, host_id: str, title: str, extra: dict | None = None) -> str:
    body = {"agent_id": agent_id, "host_id": host_id, "title": title, "workspace": "/workspace"}
    if extra:
        body.update(extra)
    sid = req("POST", "/v1/sessions", token, body).get("id")
    if not sid:
        raise RuntimeError("create session failed")
    return sid


def send_message(token: str, sid: str, text: str) -> None:
    req(
        "POST",
        f"/v1/sessions/{sid}/events",
        token,
        {"type": "message", "data": {"content": [{"type": "input_text", "text": text}]}},
    )


def read_gomoku_html(token: str, sid: str) -> tuple[bool, str]:
    for path in ("gomoku/index.html", "workspace/gomoku/index.html"):
        try:
            body = req("GET", f"/v1/sessions/{sid}/resources/environments/default/filesystem/{path}", token)
            html = body.get("content") or body.get("text") or ""
            if html.strip() and ("五子棋" in html or "gomoku" in html.lower() or "canvas" in html.lower()):
                return True, html[:300]
        except urllib.error.HTTPError:
            continue
    return False, ""


def wait_done(token: str, sid: str, label: str) -> None:
    deadline = time.time() + TIMEOUT
    while time.time() < deadline:
        sess = req("GET", f"/v1/sessions/{sid}", token)
        status = sess.get("status")
        err = sess.get("last_task_error")
        if status == "failed":
            raise RuntimeError(f"{label} failed: {err}")
        ok, snippet = read_gomoku_html(token, sid)
        if ok and status == "idle":
            print(f"PASS {label}: gomoku ready — {snippet[:80]}...")
            return
        if ok:
            print(f"  {label}: html exists, status={status}, waiting idle...")
        else:
            print(f"  {label}: status={status}, waiting gomoku/index.html...")
        time.sleep(10)
    raise TimeoutError(f"{label}: timeout {TIMEOUT}s")


def run_case(token: str, label: str, agent_id: str, host_id: str, prompt: str, extra: dict | None = None) -> str:
    print(f"\n=== {label} ===")
    sid = create_session(token, agent_id, host_id, f"Gomoku {label}", extra)
    url = f"http://127.0.0.1:10020/c/{sid}?file=gomoku%2Findex.html"
    print(f"session: {sid}")
    print(f"preview: {url}")
    send_message(token, sid, prompt)
    wait_done(token, sid, label)
    return sid


def main() -> int:
    token = login()
    codex_sid = run_case(token, "Codex-cli", "codex-cli", "host_dev-runner-codex", GOMOKU_PROMPT)
    do_sid = run_case(
        token,
        "Do-agent multi",
        "do-agent",
        "host_dev-runner-do-agent",
        DO_AGENT_PROMPT,
        {"model_config_id": 3},
    )
    print("\n=== SUMMARY ===")
    print(f"Codex:   http://127.0.0.1:10020/c/{codex_sid}?file=gomoku%2Findex.html")
    print(f"Do-agent: http://127.0.0.1:10020/c/{do_sid}?file=gomoku%2Findex.html")
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception as e:
        print(f"FAIL: {e}", file=sys.stderr)
        sys.exit(1)
