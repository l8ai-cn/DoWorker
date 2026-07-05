import { chromium } from "@playwright/test";
import { mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const OUT = join(process.cwd(), "output", "s1-smoke");
mkdirSync(OUT, { recursive: true });

const API = process.env.HIVE_API_URL || "http://localhost:10015";
const ORG = "dev-org";
const USER = { username: "devuser", password: "devpass123" };

const report = { steps: [], errors: [] };

function step(name, ok, detail = "") {
  report.steps.push({ name, ok, detail });
  console.log(`${ok ? "✓" : "✗"} ${name}${detail ? ` — ${detail}` : ""}`);
}

async function login() {
  const res = await fetch(`${API}/proto.auth.v1.AuthService/Login`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "Connect-Protocol-Version": "1" },
    body: JSON.stringify(USER),
  });
  if (!res.ok) throw new Error(`login ${res.status}`);
  const data = await res.json();
  return data.token;
}

function headers(token) {
  return {
    Authorization: `Bearer ${token}`,
    "X-Organization-Slug": ORG,
    "Content-Type": "application/json",
  };
}

async function createSession(token, body) {
  const res = await fetch(`${API}/v1/sessions`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`create session ${res.status}: ${await res.text()}`);
  return res.json();
}

async function postEvent(token, sid, text) {
  const res = await fetch(`${API}/v1/sessions/${sid}/events`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify({
      type: "message",
      data: { role: "user", content: [{ type: "input_text", text }] },
    }),
  });
  return { ok: res.ok || res.status === 202, status: res.status, body: await res.json().catch(() => ({})) };
}

async function pollItems(token, sid, min, tries = 30) {
  for (let i = 0; i < tries; i++) {
    const res = await fetch(`${API}/v1/sessions/${sid}/items`, { headers: headers(token) });
    const data = await res.json();
    const n = data.data?.length ?? 0;
    if (n >= min) return data.data;
    await new Promise((r) => setTimeout(r, 1000));
  }
  return [];
}

async function testListWire(token) {
  const res = await fetch(`${API}/v1/sessions?limit=5`, { headers: headers(token) });
  const body = await res.json();
  const row = body.data?.[0];
  const ok =
    res.ok &&
    body.data?.length > 0 &&
    row?.object === "conversation" &&
    typeof row?.updated_at === "number" &&
    row.updated_at > 0;
  step("S1.2 GET /v1/sessions list wire", ok, row ? `status=${row.status} updated_at=${row.updated_at}` : "empty");
  return ok;
}

function wsUpdatesUrl(token) {
  const u = new URL(`${API.replace("http", "ws")}/v1/sessions/updates`);
  u.searchParams.set("token", token);
  u.searchParams.set("org_slug", ORG);
  return u.toString();
}

async function testSessionUpdatesWS(token) {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  const ok = await page.evaluate(async (url) => {
    return new Promise((resolve) => {
      const ws = new WebSocket(url);
      const timer = setTimeout(() => {
        ws.close();
        resolve(false);
      }, 8000);
      ws.onopen = () => ws.send(JSON.stringify({ type: "watch", session_ids: [] }));
      ws.onmessage = () => {
        clearTimeout(timer);
        ws.close();
        resolve(true);
      };
      ws.onerror = () => {
        clearTimeout(timer);
        resolve(false);
      };
    });
  }, wsUpdatesUrl(token));
  await browser.close();
  step("S1.2 WS /v1/sessions/updates", ok, ok ? "connected" : "failed");
  return ok;
}

async function testElicitation(token) {
  const session = await createSession(token, {
    agent_id: "e2e-echo",
    title: "S1 elicitation smoke",
    scenario: "permission_request_edit",
  });
  await postEvent(token, session.id, "please edit the file");
  let elicitId = null;
  for (let i = 0; i < 25; i++) {
    const res = await fetch(`${API}/v1/sessions/${session.id}`, { headers: headers(token) });
    const body = await res.json();
    const pending = body.pending_elicitations?.[0]?.elicitation_id ?? body.pending_elicitations?.[0]?.id;
    if (pending) {
      elicitId = pending;
      break;
    }
    await new Promise((r) => setTimeout(r, 1000));
  }
  if (!elicitId) {
    step("S1.3 elicitation created", false, "no pending elicitation");
    return false;
  }
  step("S1.3 elicitation created", true, elicitId);
  const resolveRes = await fetch(
    `${API}/v1/sessions/${session.id}/elicitations/${elicitId}/resolve`,
    {
      method: "POST",
      headers: headers(token),
      body: JSON.stringify({ action: "accept", content: {} }),
    },
  );
  const resolved = resolveRes.ok || resolveRes.status === 202;
  step("S1.3 elicitation resolve", resolved, `HTTP ${resolveRes.status}`);
  if (resolved) {
    const items = await pollItems(token, session.id, 2, 20);
    step("S1.3 assistant after resolve", items.length >= 2, `items=${items.length}`);
  }
  return resolved;
}

async function testTerminalResources(token) {
  const session = await createSession(token, { agent_id: "e2e-echo", title: "S1 terminal smoke", pty_only: true });
  await postEvent(token, session.id, "hello terminal");
  await new Promise((r) => setTimeout(r, 5000));
  const res = await fetch(`${API}/v1/sessions/${session.id}/resources/terminals`, {
    headers: headers(token),
  });
  const body = await res.json();
  const terminalId = body.data?.[0]?.id;
  const listed = res.ok && terminalId != null;
  step("S1.4 list terminals", listed, terminalId ?? "none");
  if (!listed) return false;

  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  const attachUrl = (() => {
    const u = new URL(`${API.replace("http", "ws")}/v1/sessions/${session.id}/resources/terminals/${terminalId}/attach`);
    u.searchParams.set("token", token);
    u.searchParams.set("org_slug", ORG);
    return u.toString();
  })();
  const bytes = await page.evaluate(async (url) => {
    return new Promise((resolve) => {
      const ws = new WebSocket(url);
      let count = 0;
      const timer = setTimeout(() => {
        ws.close();
        resolve(count);
      }, 12000);
      ws.onmessage = () => {
        count += 1;
      };
      ws.onerror = () => {
        clearTimeout(timer);
        resolve(0);
      };
    });
  }, attachUrl);
  await browser.close();
  const ok = bytes > 0;
  step("S1.4 terminal attach bytes", ok, `messages=${bytes}`);
  return ok;
}

async function main() {
  try {
    const token = await login();
    step("API login", true);
    await testListWire(token);
    await testSessionUpdatesWS(token);
    await testElicitation(token);
    await testTerminalResources(token);
  } catch (err) {
    report.errors.push(String(err?.stack ?? err));
    step("S1 smoke run", false, String(err));
  }
  writeFileSync(join(OUT, "report.json"), JSON.stringify(report, null, 2));
  if (report.steps.some((s) => !s.ok)) process.exit(1);
}

main();
