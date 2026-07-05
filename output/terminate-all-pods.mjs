// Pre-test reset: terminate all live dev-org pods through the real
// Connect-RPC path so the runner frees its in-memory slots (a plain DB
// UPDATE leaves the runner at max_concurrent_pods and later creates 503).
// Mirrors clients/web/e2e-playwright/helpers/pod-cleanup.ts.
const API = process.env.HIVE_API_URL || "http://localhost:10015";
const ORG = "dev-org";
const HEADERS = { "Content-Type": "application/json", "Connect-Protocol-Version": "1" };

let loginRes;
try {
  loginRes = await fetch(`${API}/proto.auth.v1.AuthService/Login`, {
    method: "POST",
    headers: HEADERS,
    body: JSON.stringify({ username: "devuser", password: "devpass123" }),
  });
} catch {
  process.exit(0); // backend down — the caller's health check reports it
}
if (!loginRes.ok) process.exit(0);
const { token } = await loginRes.json();
const authed = { ...HEADERS, Authorization: `Bearer ${token}` };

const podsRes = await fetch(`${API}/proto.pod.v1.PodService/ListPods`, {
  method: "POST",
  headers: authed,
  body: JSON.stringify({ orgSlug: ORG }),
});
if (!podsRes.ok) process.exit(0);
const { items = [], pods = [] } = await podsRes.json();
const live = (items.length ? items : pods).filter((p) =>
  ["running", "initializing", "paused", "disconnected"].includes(p.status ?? ""),
);

for (const pod of live) {
  if (!pod.podKey) continue;
  await fetch(`${API}/proto.pod.v1.PodService/TerminatePod`, {
    method: "POST",
    headers: authed,
    body: JSON.stringify({ orgSlug: ORG, podKey: pod.podKey }),
  }).catch(() => {});
}
if (live.length > 0) {
  console.log(`terminated ${live.length} live pods`);
  await new Promise((r) => setTimeout(r, 5000));
}
