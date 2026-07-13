import { mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const OUT = join(process.cwd(), "output", "browser-integration");
mkdirSync(OUT, { recursive: true });
const API = process.env.SESSION_COMPAT_API_URL || "http://localhost:10015";
const ORG = "dev-org";

async function login() {
  const res = await fetch(`${API}/proto.auth.v1.AuthService/Login`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "Connect-Protocol-Version": "1" },
    body: JSON.stringify({ username: "devuser", password: "AdminAb123456" }),
  });
  const data = await res.json();
  return data.token;
}

const report = { steps: [] };
function step(name, ok, detail = "") {
  report.steps.push({ name, ok, detail });
  console.log(`${ok ? "✓" : "✗"} ${name}${detail ? ` — ${detail}` : ""}`);
}

async function main() {
  const token = await login();
  const headers = {
    Authorization: `Bearer ${token}`,
    "X-Organization-Slug": ORG,
    "Content-Type": "application/json",
  };

  const t0 = performance.now();
  const agents = await fetch(`${API}/v1/agents`, { headers }).then((r) => r.json());
  step("GET /v1/agents", (agents.data?.length ?? 0) > 0, `${agents.data?.length} agents in ${(performance.now() - t0).toFixed(0)}ms`);

  const create = await fetch(`${API}/v1/sessions`, {
    method: "POST",
    headers,
    body: JSON.stringify({ agent_id: "e2e-echo", title: "API integration smoke" }),
  }).then((r) => ({ status: r.status, body: r.json() })).then(async (x) => ({ status: x.status, body: await x.body }));
  const sid = create.body.id;
  step("POST /v1/sessions", create.status === 200, `${sid} status=${create.body.status}`);

  const evt = await fetch(`${API}/v1/sessions/${sid}/events`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      type: "message",
      data: { role: "user", content: [{ type: "text", text: "Reply with exactly: pong" }] },
    }),
  }).then((r) => r.json());
  step("POST /v1/sessions/.../events", evt.queued === true, `item=${evt.item_id}`);

  for (let i = 0; i < 20; i++) {
    await new Promise((r) => setTimeout(r, 2000));
    const items = await fetch(`${API}/v1/sessions/${sid}/items`, { headers }).then((r) => r.json());
    const assistant = (items.data ?? []).find((it) => it.role === "assistant");
    if (assistant) {
      const text = assistant.content?.map((c) => c.text).join(" ") ?? "";
      step("GET /v1/sessions/.../items assistant", true, text.slice(0, 80));
      break;
    }
    if (i === 19) step("GET /v1/sessions/.../items assistant", false, `only ${items.data?.length ?? 0} items after 40s`);
  }

  const health = await fetch(`${API}/health?session_ids=${encodeURIComponent(sid)}`, { headers }).then((r) => r.json());
  const live = health.sessions?.[sid]?.runner_online;
  step("GET /health?session_ids=", live === true, `runner_online=${live}`);

  writeFileSync(join(OUT, "api-smoke-report.json"), JSON.stringify(report, null, 2));
  if (report.steps.some((s) => !s.ok)) process.exit(1);
}

main();
