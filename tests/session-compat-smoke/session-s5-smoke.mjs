import { createE2EEchoSession } from "./e2e-echo-session-plan.mjs";

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

async function createSession(token, body) {
  if (body.agent_id === "e2e-echo") {
    return createE2EEchoSession(token, body);
  }
  const res = await fetch(`${API}/v1/sessions`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`create session ${res.status}: ${await res.text()}`);
  return res.json();
}

async function main() {
  const token = await login();
  step("S5 login", true);

  const projectsRes = await fetch(`${API}/v1/sessions/projects`, { headers: headers(token) });
  const projectsBody = await projectsRes.json();
  // Wire is a bare string[] (web-user / mobile parse the body as string[]).
  step("S5 W1 GET /sessions/projects", projectsRes.ok && Array.isArray(projectsBody));

  const session = await createSession(token, { agent_id: "e2e-echo", title: "S5 parity" });
  const sid = session.id;
  step("S5 fixture session", !!sid, sid);

  const agentRes = await fetch(`${API}/v1/sessions/${sid}/agent`, { headers: headers(token) });
  const agentBody = await agentRes.json();
  step("S5 W1 GET /sessions/{id}/agent", agentRes.ok && !!agentBody.id, agentBody.id);

  const readRes = await fetch(`${API}/v1/sessions/${sid}/read-state`, { headers: headers(token) });
  step("S5 W1 GET /sessions/{id}/read-state", readRes.ok);

  const fsRes = await fetch(`${API}/v1/sessions/${sid}/resources/environments/default/changes`, {
    headers: headers(token),
  });
  step("S5 W2 GET environments/.../changes", fsRes.status !== 501, `HTTP ${fsRes.status}`);

  const fileRes = await fetch(`${API}/v1/sessions/${sid}/resources/files`, {
    method: "POST",
    headers: headers(token),
    body: "",
  });
  step("S5 W3 POST resources/files routed", fileRes.status !== 404 && fileRes.status !== 501, `HTTP ${fileRes.status}`);

  const commentRes = await fetch(`${API}/v1/sessions/${sid}/comments`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify({
      path: "/README.md",
      start_index: 0,
      end_index: 1,
      body: "S5 smoke",
    }),
  });
  step("S5 W4 POST /sessions/{id}/comments", commentRes.ok, `HTTP ${commentRes.status}`);

  const childRes = await fetch(`${API}/v1/sessions/${sid}/child_sessions`, { headers: headers(token) });
  const childBody = await childRes.json();
  step("S5 W4 GET /child_sessions", childRes.ok && Array.isArray(childBody.data));

  const mcpRes = await fetch(`${API}/v1/sessions/${sid}/agent/mcp-servers`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify({ name: "s5-test", transport: "stdio", command: "echo", args: [] }),
  });
  step("S5 W5 POST agent/mcp-servers", mcpRes.status !== 404 && mcpRes.status !== 501, `HTTP ${mcpRes.status}`);

  const polListRes = await fetch(`${API}/v1/sessions/${sid}/policies`, { headers: headers(token) });
  step("S5 W6 GET /sessions/{id}/policies", polListRes.ok);

  const polRes = await fetch(`${API}/v1/sessions/${sid}/policies`, {
    method: "POST",
    headers: headers(token),
    body: JSON.stringify({
      name: "s5_bash",
      type: "python",
      handler: "acp_tool_rule",
      factory_params: { tool_pattern: "Bash", verdict: "ask", priority: 1 },
    }),
  });
  const polBody = await polRes.json();
  const polOk = polRes.ok && polBody.source === "session";
  step("S5 W6 POST /sessions/{id}/policies", polOk, polBody.id);
  if (polBody.id) {
    const delPol = await fetch(
      `${API}/v1/sessions/${sid}/policies/${encodeURIComponent(polBody.id)}`,
      { method: "DELETE", headers: headers(token) },
    );
    step("S5 W6 DELETE session policy", delPol.status === 204);
  }

  const hostsRes = await fetch(`${API}/v1/hosts`, { headers: headers(token) });
  const hostsBody = await hostsRes.json();
  const hostId = hostsBody.hosts?.[0]?.host_id;
  if (hostId) {
    const dirRes = await fetch(`${API}/v1/hosts/${hostId}/directories`, {
      method: "POST",
      headers: headers(token),
      body: JSON.stringify({ path: "/tmp/s5-smoke-dir" }),
    });
    step("S5 W7 POST /hosts/{id}/directories", dirRes.status !== 501, `HTTP ${dirRes.status}`);
  } else {
    step("S5 W7 POST /hosts/{id}/directories", true, "skipped — no hosts");
  }

  const agentsRes = await fetch(`${API}/v1/agents`, { headers: headers(token) });
  const agentsBody = await agentsRes.json();
  const firstId = agentsBody.data?.[0]?.id;
  const page2Res = firstId
    ? await fetch(`${API}/v1/agents?after=${encodeURIComponent(firstId)}`, { headers: headers(token) })
    : null;
  const cursorOk =
    agentsRes.ok &&
    Array.isArray(agentsBody.data) &&
    typeof agentsBody.has_more === "boolean" &&
    (!page2Res || page2Res.ok);
  step("S5 W8 GET /v1/agents cursor", cursorOk, `count=${agentsBody.data?.length ?? 0}`);

  const delRes = await fetch(`${API}/v1/sessions/${sid}?delete_branch=false`, {
    method: "DELETE",
    headers: headers(token),
  });
  step("S5 W1 DELETE /sessions/{id}", delRes.status === 204 || delRes.ok, `HTTP ${delRes.status}`);

  const failed = report.steps.filter((s) => !s.ok);
  if (failed.length) {
    console.error("\nFAILED:", failed);
    process.exit(1);
  }
  console.log("\nS5 smoke: all steps passed");
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
