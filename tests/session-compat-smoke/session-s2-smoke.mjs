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

async function main() {
  const sessionIDs = new Set();
  let token;
  try {
    token = await login();

    const harnessRes = await fetch(`${API}/v1/harnesses`, { headers: headers(token) });
    const harnessBody = await harnessRes.json();
    const harnessOk =
      harnessRes.ok && Array.isArray(harnessBody.data) && harnessBody.data.some((r) => r.id && r.label);
    step("S2.1 GET /v1/harnesses", harnessOk, `count=${harnessBody.data?.length ?? 0}`);

    const registryRes = await fetch(`${API}/v1/policy-registry`, { headers: headers(token) });
    const registryBody = await registryRes.json();
    const registryOk =
      registryRes.ok &&
      registryBody.data?.some((r) => r.handler === "acp_tool_rule" && r.kind === "factory");
    step("S2.2 GET /v1/policy-registry", registryOk);

    const listPolRes = await fetch(`${API}/v1/policies`, { headers: headers(token) });
    const listPolBody = await listPolRes.json();
    step("S2.2 GET /v1/policies", listPolRes.ok && Array.isArray(listPolBody.data), `count=${listPolBody.data?.length ?? 0}`);

    const createPolRes = await fetch(`${API}/v1/policies`, {
      method: "POST",
      headers: headers(token),
      body: JSON.stringify({
        name: "test_bash",
        type: "python",
        handler: "acp_tool_rule",
        factory_params: { tool_pattern: "Bash", verdict: "ask", priority: 10 },
      }),
    });
    const created = await createPolRes.json();
    const createOk = createPolRes.ok && created.id?.startsWith("pol_");
    step("S2.2 POST /v1/policies", createOk, created.id);

    if (created.id) {
      const delRes = await fetch(`${API}/v1/policies/${encodeURIComponent(created.id)}`, {
        method: "DELETE",
        headers: headers(token),
      });
      step("S2.2 DELETE /v1/policies/{id}", delRes.status === 204);
    }

    const fixture = trackSmokeSession(
      sessionIDs,
      await createE2EEchoSession(token, { title: "S2 API fixture" }),
    );
    const sid = fixture.id;
    const sessRes = await fetch(`${API}/v1/sessions?limit=50`, { headers: headers(token) });
    const sessBody = await sessRes.json();
    const session = sessBody.data?.find((item) => item.id === sid);
    if (!session) {
      step("S2.3 session fixture", false, "created fixture was not listed");
    } else {
      const listOk =
        typeof session.updated_at === "number" &&
        session.permission_level === 4;
      step("S2.3 list permission_level + updated_at", listOk);

      const readRes = await fetch(`${API}/v1/sessions/${sid}/read-state`, {
        method: "PUT",
        headers: headers(token),
        body: JSON.stringify({ last_seen: Math.floor(Date.now() / 1000), unread: false }),
      });
      step("S2.3 PUT read-state", readRes.ok);

      const permRes = await fetch(`${API}/v1/sessions/${sid}/permissions`, { headers: headers(token) });
      const permBody = await permRes.json();
      step("S2.3 GET permissions", permRes.ok && Array.isArray(permBody));

      const ownerRes = await fetch(`${API}/v1/sessions/${sid}/owner`, { headers: headers(token) });
      const ownerBody = await ownerRes.json();
      step("S2.3 GET owner", ownerRes.ok && typeof ownerBody.owner === "string");

      const switchRes = await fetch(`${API}/v1/sessions/${sid}/switch-agent`, {
        method: "POST",
        headers: headers(token),
        body: JSON.stringify({ agent_id: "e2e-echo" }),
      });
      step("S2.5 switch-agent implemented", switchRes.status !== 501, `HTTP ${switchRes.status}`);

      const hostsRes = await fetch(`${API}/v1/hosts`, { headers: headers(token) });
      const hostsBody = await hostsRes.json();
      const hostId = hostsBody.data?.[0]?.id;
      if (hostId) {
        const dirRes = await fetch(`${API}/v1/hosts/${hostId}/directories`, {
          method: "POST",
          headers: headers(token),
          body: JSON.stringify({ path: "/tmp/s2-smoke" }),
        });
        step("S2.5 POST directories implemented", dirRes.status !== 501, `HTTP ${dirRes.status}`);
      } else {
        step("S2.5 POST directories implemented", true, "skipped — no hosts");
      }
    }
  } catch (error) {
    step("S2 smoke run", false, String(error));
  } finally {
    if (token) {
      try {
        step("S2 fixture cleanup", true, `deleted=${await cleanupSmokeSessions(token, sessionIDs)}`);
      } catch (error) {
        step("S2 fixture cleanup", false, String(error));
      }
    }
  }

  const failed = report.steps.filter((s) => !s.ok);
  if (failed.length) {
    console.error("\nFAILED:", failed);
    process.exitCode = 1;
  }
  if (process.exitCode !== 1) console.log("\nS2 smoke: all steps passed");
}

await main();
