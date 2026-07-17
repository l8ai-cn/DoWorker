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
  const token = await login();

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

  const sessRes = await fetch(`${API}/v1/sessions?limit=1`, { headers: headers(token) });
  const sessBody = await sessRes.json();
  const sid = sessBody.data?.[0]?.id;
  if (!sid) {
    step("S2.3 session fixture", false, "no sessions");
  } else {
    const listOk =
      typeof sessBody.data[0].updated_at === "number" &&
      sessBody.data[0].permission_level === 4;
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

  const failed = report.steps.filter((s) => !s.ok);
  if (failed.length) {
    console.error("\nFAILED:", failed);
    process.exit(1);
  }
  console.log("\nS2 smoke: all steps passed");
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
