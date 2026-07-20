const API = process.env.SESSION_COMPAT_API_URL ||
  `http://localhost:${process.env.BACKEND_HTTP_PORT || "10015"}`;
const ORG = "dev-org";

async function login() {
  const res = await fetch(`${API}/proto.auth.v1.AuthService/Login`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "Connect-Protocol-Version": "1" },
    body: JSON.stringify({ username: "devuser", password: "AdminAb123456" }),
  });
  if (!res.ok) throw new Error(`login ${res.status}`);
  return (await res.json()).token;
}

function headers(token) {
  return {
    Authorization: `Bearer ${token}`,
    "X-Organization-Slug": ORG,
    "Content-Type": "application/json",
  };
}

async function listRunners(token) {
  const res = await fetch(`${API}/v1/runners`, { headers: headers(token) });
  if (!res.ok) throw new Error(`list runners ${res.status}: ${await res.text()}`);
  const body = await res.json();
  return Array.isArray(body.data) ? body.data : [];
}

async function main() {
  const token = await login();
  let last = "no runner data";
  for (let attempt = 0; attempt < 60; attempt += 1) {
    const runners = await listRunners(token);
    const e2e = runners.filter((runner) =>
      runner.online === true &&
      (runner.harnesses ?? []).includes("e2e-echo")
    );
    const ready = e2e.filter((runner) => runner.tunnel_state === "connected");
    if (ready.length > 0) {
      console.log(`e2e-echo runner(s) online with tunnel: ${ready.length}`);
      return;
    }
    last = e2e.map((runner) =>
      `${runner.runner_id}: tunnel=${runner.tunnel_state ?? "missing"} error=${runner.tunnel_last_error ?? ""}`
    ).join("; ") || "no online e2e-echo runner";
    await new Promise((resolve) => setTimeout(resolve, 3000));
  }
  throw new Error(`no tunnel-ready e2e-echo runner after wait: ${last}`);
}

main().catch((error) => {
  console.error(error.message);
  process.exit(1);
});
