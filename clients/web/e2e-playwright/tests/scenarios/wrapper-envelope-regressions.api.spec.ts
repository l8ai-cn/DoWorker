import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

const ORG_BASE = `/api/v1/orgs/${TEST_ORG_SLUG}`;

test.describe("Backend wrapper envelope contracts", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  test("ticket list keeps total/limit/offset", async ({ api }) => {
    const res = await api.get(`${ORG_BASE}/tickets?limit=5&offset=0`);
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body).toHaveProperty("tickets");
    expect(body).toHaveProperty("total");
    expect(body).toHaveProperty("limit");
    expect(body).toHaveProperty("offset");
    expect(body.limit).toBe(5);
    expect(body.offset).toBe(0);
  });

  test("dlq list keeps total alongside entries", async ({ api }) => {
    const res = await api.get(`${ORG_BASE}/messages/dlq?limit=10&offset=0`);
    // Messaging is feature-gated — dev/CI backends may run with svc.Message=nil
    // and short-circuit registerMessageRoutes. Treat 404 as "feature off" rather
    // than a wrapper-shape regression.
    if (res.status === 404) { test.skip(); return; }
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body).toHaveProperty("entries");
    expect(body).toHaveProperty("total");
    expect(typeof body.total).toBe("number");
  });

  test("pod create response carries pod envelope", async ({ api }) => {
    const runners = (await (await api.get(`${ORG_BASE}/runners/available`)).json()).runners;
    if (!runners?.length) { test.skip(); return; }
    const agents = (await (await api.get(`${ORG_BASE}/agents`)).json()).builtin_agents;
    if (!agents?.length) { test.skip(); return; }

    const res = await api.post(`${ORG_BASE}/pods`, {
      runner_id: runners[0].id,
      agent_slug: agents[0].slug,
    });
    expect([200, 201]).toContain(res.status);
    const body = await res.json();
    expect(body).toHaveProperty("pod");
    expect(body.pod?.pod_key || body.pod?.key).toBeTruthy();
    if (body.pod?.pod_key || body.pod?.key) {
      const key = body.pod.pod_key || body.pod.key;
      await api.post(`${ORG_BASE}/pods/${key}/terminate`, {});
    }
  });

  test("loop runs list keeps pagination shape", async ({ api }) => {
    const loopsRes = await api.get(`${ORG_BASE}/loops?limit=1`);
    if (loopsRes.status !== 200) { test.skip(); return; }
    const loops = (await loopsRes.json()).loops;
    if (!loops?.length) { test.skip(); return; }

    const res = await api.get(`${ORG_BASE}/loops/${loops[0].slug}/runs?limit=10&offset=0`);
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body).toHaveProperty("runs");
    expect(body).toHaveProperty("total");
    expect(body).toHaveProperty("limit");
    expect(body).toHaveProperty("offset");
  });
});
