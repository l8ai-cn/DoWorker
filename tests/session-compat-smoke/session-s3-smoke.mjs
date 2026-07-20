import { createE2EEchoSession } from "./e2e-echo-session-plan.mjs";
import {
  cleanupSmokeSessions,
  trackSmokeSession,
} from "./session-fixture-cleanup.mjs";

const API = process.env.SESSION_COMPAT_API_URL || "http://localhost:10015";
const ORG = "dev-org";
const USER = { username: "devuser", password: "AdminAb123456" };

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

async function createSession(token, sessionIDs, body) {
  if (body.agent_id === "e2e-echo") {
    return trackSmokeSession(sessionIDs, await createE2EEchoSession(token, body));
  }
  const res = await fetch(`${API}/v1/sessions`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`create session ${res.status}: ${await res.text()}`);
  return trackSmokeSession(sessionIDs, await res.json());
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
  return { ok: res.ok || res.status === 202, status: res.status };
}

async function getSession(token, sid) {
  const res = await fetch(`${API}/v1/sessions/${sid}`, { headers: headers(token) });
  return { ok: res.ok, body: await res.json() };
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

async function pollUsage(token, sid, tries = 45) {
  for (let i = 0; i < tries; i++) {
    const { ok, body } = await getSession(token, sid);
    const cost = body.total_cost_usd;
    if (ok && typeof cost === "number" && cost > 0) {
      return cost;
    }
    await new Promise((r) => setTimeout(r, 1000));
  }
  return null;
}

async function testUsageReporting(token, sessionIDs) {
  const session = await createSession(token, sessionIDs, {
    agent_id: "e2e-echo",
    title: "S3 usage smoke",
  });
  step("S3.3 session created", !!session.id, session.id);

  const sent = await postEvent(token, session.id, "count tokens please");
  step("S3.3 user message sent", sent.ok, `HTTP ${sent.status}`);

  const items = await pollItems(token, session.id, 2, 40);
  step("S3.3 assistant reply", items.length >= 2, `items=${items.length}`);

  const cost = await pollUsage(token, session.id, 45);
  step("S3.3 total_cost_usd > 0", cost != null, cost != null ? `$${cost.toFixed(6)}` : "timeout");

  // Model name must come from the agent's usage report: the mock emits
  // gpt-4o, which differs from the runner's estimate-fallback default
  // (gpt-4o-mini) — so this only passes via the real passthrough path.
  const { body } = await getSession(token, session.id);
  const models = Object.keys(body.usage_by_model ?? {});
  step("S3.3 usage_by_model from agent report", models.includes("gpt-4o"), models.join(",") || "empty");
  return { sessionId: session.id, cost, itemCount: items.length };
}

async function testPolicyHotPush(token, activeSessionId) {
  const createRes = await fetch(`${API}/v1/policies`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify({
      name: "s3_hot_push",
      type: "python",
      handler: "acp_tool_rule",
      factory_params: { tool_pattern: "Write", verdict: "deny", priority: 5 },
    }),
  });
  const created = await createRes.json();
  const createOk = createRes.ok && created.id?.startsWith("pol_");
  step(
    "S3.2 policy create + hot push",
    createOk,
    activeSessionId ? `active=${activeSessionId} id=${created.id ?? "none"}` : created.id,
  );

  if (created.id) {
    const delRes = await fetch(`${API}/v1/policies/${encodeURIComponent(created.id)}`, {
      method: "DELETE",
      headers: headers(token),
    });
    step("S3.2 policy delete + hot push", delRes.status === 204);
  }
  return createOk;
}

// Behavioral proof of hot push: a deny rule pushed AFTER the pod is running
// must make the runner auto-deny the mock's Edit permission request — no
// elicitation surfaces, yet the turn still completes (mock emits a failed
// tool call and ends the turn).
async function testPolicyDenyBehavior(token, sessionIDs) {
  const session = await createSession(token, sessionIDs, {
    agent_id: "e2e-echo",
    title: "S3 deny behavior",
    scenario: "permission_request_edit",
  });
  for (let i = 0; i < 20; i++) {
    const { body } = await getSession(token, session.id);
    if (body.status && body.status !== "launching") break;
    await new Promise((r) => setTimeout(r, 1000));
  }

  const createRes = await fetch(`${API}/v1/policies`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify({
      name: "s3_deny_edit",
      type: "python",
      handler: "acp_tool_rule",
      factory_params: { tool_pattern: "Edit", verdict: "deny", priority: 100 },
    }),
  });
  const created = await createRes.json();
  step("S3.2 deny rule pushed to running pod", createRes.ok, created.id);
  await new Promise((r) => setTimeout(r, 2000));

  await postEvent(token, session.id, "please edit the file");

  let sawElicitation = false;
  let itemCount = 0;
  let itemsText = "";
  for (let i = 0; i < 20; i++) {
    const { body } = await getSession(token, session.id);
    if ((body.pending_elicitations?.length ?? 0) > 0) {
      sawElicitation = true;
      break;
    }
    const itemsRes = await fetch(`${API}/v1/sessions/${session.id}/items`, { headers: headers(token) });
    const itemsBody = await itemsRes.json();
    itemCount = itemsBody.data?.length ?? 0;
    itemsText = JSON.stringify(itemsBody.data ?? []);
    if (itemCount >= 2 && itemsText.includes("Edit denied")) break;
    await new Promise((r) => setTimeout(r, 1000));
  }
  step("S3.2 runner auto-denied (no elicitation)", !sawElicitation && itemCount >= 2,
    sawElicitation ? "elicitation surfaced — hot push not applied" : `items=${itemCount}`);
  // Direction assertion: the mock reveals its verdict in the assistant text
  // ("Edit denied: skipped." vs "Edit approved: applied."), so a rule
  // misconfigured as allow cannot pass this step.
  step("S3.2 verdict direction is deny", itemsText.includes("Edit denied") && !itemsText.includes("Edit approved"),
    itemsText.includes("Edit denied") ? "assistant text confirms deny" : `no deny marker in items`);

  if (created.id) {
    await fetch(`${API}/v1/policies/${encodeURIComponent(created.id)}`, {
      method: "DELETE",
      headers: headers(token),
    });
  }
  return !sawElicitation;
}

async function testForkSession(token, sessionIDs, sourceSessionId) {
  await postEvent(token, sourceSessionId, "fork me");
  const sourceItems = await pollItems(token, sourceSessionId, 2, 40);
  step("S3.4 source conversation", sourceItems.length >= 2, `items=${sourceItems.length}`);

  const forkRes = await fetch(`${API}/v1/sessions/${sourceSessionId}/fork`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify({ title: "S3 fork child" }),
  });
  const forkBody = await forkRes.json();
  const forkOk = forkRes.ok && forkBody.id && forkBody.id !== sourceSessionId;
  step("S3.4 POST fork", forkOk, forkBody.id ?? `HTTP ${forkRes.status}`);

  if (!forkOk) return false;
  trackSmokeSession(sessionIDs, forkBody);

  const forkItems = await fetch(`${API}/v1/sessions/${forkBody.id}/items`, {
    headers: headers(token),
  }).then((r) => r.json());
  const copied = (forkItems.data?.length ?? 0) >= sourceItems.length;
  step("S3.4 fork copied items", copied, `fork=${forkItems.data?.length ?? 0} source=${sourceItems.length}`);

  const cont = await postEvent(token, forkBody.id, "continue on fork");
  const contItems = cont.ok ? await pollItems(token, forkBody.id, sourceItems.length + 2, 30) : [];
  step("S3.4 fork continues", contItems.length > sourceItems.length, `items=${contItems.length}`);
  return forkOk && copied;
}

async function main() {
  const sessionIDs = new Set();
  let token;
  try {
    token = await login();
    step("API login", true);

    const usage = await testUsageReporting(token, sessionIDs);
    await testPolicyHotPush(token, usage.sessionId);
    await testPolicyDenyBehavior(token, sessionIDs);
    await testForkSession(token, sessionIDs, usage.sessionId);
  } catch (error) {
    report.errors.push(String(error?.stack ?? error));
    step("S3 smoke run", false, String(error));
  } finally {
    if (token) {
      try {
        step("S3 fixture cleanup", true, `deleted=${await cleanupSmokeSessions(token, sessionIDs)}`);
      } catch (error) {
        step("S3 fixture cleanup", false, String(error));
      }
    }
  }
  const failed = report.steps.filter((s) => !s.ok);
  if (failed.length) {
    console.error("\nFAILED:", failed);
    process.exitCode = 1;
  } else {
    console.log("\nS3 smoke: all steps passed");
  }
}

await main();
