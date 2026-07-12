// Migrated R5+: Connect-RPC only (no REST middle layer).
//
// This file historically asserted the REST JSON envelope shapes (e.g. `tickets`
// vs `items` key naming, `pod` wrapper, list pagination keys). The REST
// envelopes are gone, but the same shape contracts now ride on the Connect
// proto messages — list responses still expose `total`/`limit`/`offset`,
// CreatePod still returns the `pod` field, and goal-loop lists still paginate.
// Re-validate those at the Connect layer here so a regression on the proto
// descriptor (e.g., someone renaming `total` to `count` on the wire) gets
// caught at the e2e seam.
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { E2E_ECHO_AGENT_SLUG, pickE2EEchoRunner } from "../../helpers/e2e-echo-runner";

test.describe("Backend wrapper envelope contracts", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  test("ticket list keeps total/limit/offset", async ({ api }) => {
    const cc = await api.connect();
    const res = await cc.ticket.listTickets({ orgSlug: TEST_ORG_SLUG }) as {
      items: unknown[];
      total?: number | bigint;
      limit?: number;
      offset?: number;
    };
    expect(Array.isArray(res.items)).toBe(true);
    // Pagination fields exist on the proto descriptor even when zero. The
    // contract is: the response is shaped { items, total, limit, offset }.
    expect(typeof res.total).toMatch(/number|bigint/);
    expect(typeof res.limit).toBe("number");
    expect(typeof res.offset).toBe("number");
  });

  test("pod create response carries pod envelope", async ({ api }) => {
    const cc = await api.connect();
    const runners = await cc.runner.listAvailableRunners({ orgSlug: TEST_ORG_SLUG }) as {
      items: { id: bigint }[];
    };
    expect(runners.items.length, "dev env must have an online runner").toBeGreaterThan(0);
    const runnerId = pickE2EEchoRunner(runners.items).id;

    const resp = await cc.pod.createPod({
      orgSlug: TEST_ORG_SLUG,
      runnerId,
      agentSlug: E2E_ECHO_AGENT_SLUG,
    }) as { pod?: { podKey?: string } };
    // The legacy REST `{ "pod": {...} }` envelope is now the proto Pod field.
    expect(resp.pod).toBeDefined();
    expect(typeof resp.pod?.podKey).toBe("string");

    // Cleanup
    if (resp.pod?.podKey) {
      await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey: resp.pod.podKey }).catch(() => {});
    }
  });

  test("goal loop list keeps pagination shape", async ({ api }) => {
    const cc = await api.connect();
    const loops = await cc.goalLoop.listGoalLoops({
      orgSlug: TEST_ORG_SLUG,
    }) as { items: unknown[]; total?: number | bigint; limit?: number; offset?: number };
    expect(Array.isArray(loops.items)).toBe(true);
    expect(typeof loops.total).toMatch(/number|bigint/);
    expect(typeof loops.limit).toBe("number");
    expect(typeof loops.offset).toBe("number");
  });
});
