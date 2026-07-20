import { createE2EEchoSession } from "./e2e-echo-session-plan.mjs";
import {
  cleanupSmokeSessions,
  trackSmokeSession,
} from "./session-fixture-cleanup.mjs";

const API = process.env.SESSION_COMPAT_API_URL || "http://localhost:10015";
const ORG = "dev-org";
const USER = { username: "devuser", password: "AdminAb123456" };

const report = { steps: [] };

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
  return (await res.json()).token;
}

function headers(token, extra = {}) {
  return {
    Authorization: `Bearer ${token}`,
    "X-Organization-Slug": ORG,
    "Content-Type": "application/json",
    ...extra,
  };
}

async function createSession(token, sessionIDs, title) {
  return trackSmokeSession(sessionIDs, await createE2EEchoSession(token, { title }));
}

async function testReadStatePersist(token, sessionIDs) {
  const sid = (await createSession(token, sessionIDs, "S4 read-state fixture")).id;
  const marker = Math.floor(Date.now() / 1000);
  const putRes = await fetch(`${API}/v1/sessions/${sid}/read-state`, {
    method: "PUT",
    headers: headers(token),
    body: JSON.stringify({ last_seen: marker, unread: true }),
  });
  step("S4.1 PUT read-state", putRes.ok, `sid=${sid} marker=${marker}`);

  const listRes = await fetch(`${API}/v1/sessions?limit=50`, { headers: headers(token) });
  const listBody = await listRes.json();
  const row = (listBody.data ?? []).find((c) => c.id === sid);
  const wireOk =
    row?.viewer_last_seen === marker && row?.viewer_unread === true;
  step("S4.1 GET sessions reflects read-state", wireOk,
    `seen=${row?.viewer_last_seen} unread=${row?.viewer_unread}`);
  return wireOk;
}

async function testOrgUsageSummary(token) {
  const res = await fetch(`${API}/v1/org/usage/summary`, { headers: headers(token) });
  const body = await res.json();
  const ok =
    res.ok &&
    body.object === "org_usage_summary" &&
    (typeof body.total_cost_usd === "number" || body.usage_by_model);
  step("S4.3 GET /v1/org/usage/summary", ok,
    ok ? `cost=${body.total_cost_usd ?? "n/a"} models=${Object.keys(body.usage_by_model ?? {}).join(",")}` : JSON.stringify(body));
  return ok;
}

async function testCostBudgetGate(token, sessionIDs) {
  const session = await createSession(token, sessionIDs, "S4 cost budget");
  step("S4.2 session created", true, session.id);

  await fetch(`${API}/v1/sessions/${session.id}/events`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify({
      type: "message",
      data: { role: "user", content: [{ type: "input_text", text: "charge usage" }] },
    }),
  });

  let spent = 0;
  for (let i = 0; i < 45; i++) {
    const sRes = await fetch(`${API}/v1/sessions/${session.id}`, { headers: headers(token) });
    const sBody = await sRes.json();
    if (typeof sBody.total_cost_usd === "number" && sBody.total_cost_usd > 0) {
      spent = sBody.total_cost_usd;
      break;
    }
    await new Promise((r) => setTimeout(r, 1000));
  }
  step("S4.2 usage accumulated", spent > 0, spent > 0 ? `$${spent.toFixed(8)}` : "timeout");

  const maxUsd = spent > 0 ? spent * 0.5 : 0.000001;
  const polRes = await fetch(`${API}/v1/policies`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify({
      name: "s4_cost_cap",
      type: "python",
      handler: "session_cost_budget",
      factory_params: { max_usd: maxUsd, priority: 50 },
    }),
  });
  const policy = await polRes.json();
  step("S4.2 cost budget policy created", polRes.ok && policy.handler === "session_cost_budget", policy.id);

  const blockedRes = await fetch(`${API}/v1/sessions/${session.id}/events`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify({
      type: "message",
      data: { role: "user", content: [{ type: "input_text", text: "should be blocked" }] },
    }),
  });
  const blockedBody = await blockedRes.json().catch(() => ({}));
  const blocked =
    blockedRes.status === 402 &&
    blockedBody.code === "cost_budget_exceeded";
  step("S4.2 turn blocked at budget (402)", blocked, `HTTP ${blockedRes.status}`);

  if (policy.id) {
    await fetch(`${API}/v1/policies/${encodeURIComponent(policy.id)}`, {
      method: "DELETE",
      headers: headers(token),
    });
  }
  return blocked;
}

async function testCostBudgetRegistry(token) {
  const res = await fetch(`${API}/v1/policy-registry`, { headers: headers(token) });
  const body = await res.json();
  const ok = body.data?.some((r) => r.handler === "session_cost_budget");
  step("S4.2 policy-registry lists session_cost_budget", ok);
  return ok;
}

async function testResumeExternalSession(token, sessionIDs) {
  const source = await createSession(token, sessionIDs, "S4 resume source");
  step("S4.5 source session", true, source.id);

  await fetch(`${API}/v1/sessions/${source.id}/events`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify({
      type: "message",
      data: { role: "user", content: [{ type: "input_text", text: "prime session" }] },
    }),
  });
  const primed = await pollItems(token, source.id, 2, 40);
  step("S4.5 source primed", primed.length >= 2, `items=${primed.length}`);

  const forkRes = await fetch(`${API}/v1/sessions/${source.id}/fork`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify({ title: "S4 resume fork" }),
  });
  const fork = await forkRes.json();
  step("S4.5 fork created", forkRes.ok && fork.id, fork.id);
  if (!forkRes.ok || !fork.id) return false;
  trackSmokeSession(sessionIDs, fork);

  for (let i = 0; i < 25; i++) {
    const sRes = await fetch(`${API}/v1/sessions/${fork.id}`, { headers: headers(token) });
    const sBody = await sRes.json();
    if (sBody.status && sBody.status !== "launching") break;
    await new Promise((r) => setTimeout(r, 1000));
  }

  await fetch(`${API}/v1/sessions/${fork.id}/events`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify({
      type: "message",
      data: { role: "user", content: [{ type: "input_text", text: "resume check" }] },
    }),
  });

  let resumed = false;
  for (let i = 0; i < 30; i++) {
    const itemsRes = await fetch(`${API}/v1/sessions/${fork.id}/items`, { headers: headers(token) });
    const itemsBody = await itemsRes.json();
    const text = JSON.stringify(itemsBody.data ?? []);
    if (text.includes("RESUMED_OK")) {
      resumed = true;
      break;
    }
    await new Promise((r) => setTimeout(r, 1000));
  }
  step("S4.5 fork used session/resume", resumed, resumed ? "RESUMED_OK in items" : "timeout");
  return resumed;
}

async function pollItems(token, sid, min, tries = 40) {
  for (let i = 0; i < tries; i++) {
    const res = await fetch(`${API}/v1/sessions/${sid}/items`, { headers: headers(token) });
    const data = await res.json();
    const n = data.data?.length ?? 0;
    if (n >= min) return data.data;
    await new Promise((r) => setTimeout(r, 1000));
  }
  return [];
}

async function main() {
  const sessionIDs = new Set();
  let token;
  try {
    token = await login();
    step("API login", true);

    await testReadStatePersist(token, sessionIDs);
    await testCostBudgetRegistry(token);
    await testOrgUsageSummary(token);
    await testCostBudgetGate(token, sessionIDs);
    await testResumeExternalSession(token, sessionIDs);
  } catch (error) {
    step("S4 smoke run", false, String(error));
  } finally {
    if (token) {
      try {
        step("S4 fixture cleanup", true, `deleted=${await cleanupSmokeSessions(token, sessionIDs)}`);
      } catch (error) {
        step("S4 fixture cleanup", false, String(error));
      }
    }
  }
  const failed = report.steps.filter((s) => !s.ok);
  if (failed.length) {
    console.error("\nFAILED:", failed);
    process.exitCode = 1;
  } else {
    console.log("\nS4 smoke: all steps passed");
  }
}

await main();
