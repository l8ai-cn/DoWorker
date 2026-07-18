import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { pollUntil } from "../../helpers/retry";
import { terminateAllPods } from "../../helpers/pod-cleanup";
import {
  buildE2EEchoWorkerSpec,
  type E2EWorkerSpecDraft,
} from "../../helpers/e2e-worker-spec";

type ConnectClient = Awaited<
  ReturnType<import("../../fixtures/api.fixture").ApiFixture["connect"]>
>;
type Runner = {
  id: bigint;
  currentPods: number;
  maxConcurrentPods: number;
  isEnabled: boolean;
};
type Pod = { podKey: string; runnerId: bigint };

test.describe("Journey: Runner Scaling", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
    await terminateAllPods();
  });

  test("enforces capacity and schedules onto an enabled runner", async ({ api }) => {
    const cc = await api.connect();
    const { items: available } = await cc.runner.listAvailableRunners({
      orgSlug: TEST_ORG_SLUG,
    }) as { items: Runner[] };
    expect(available.length, "dev env must expose two e2e runners").toBeGreaterThanOrEqual(2);

    const [primary, secondary] = available;
    const originals = available.map((runner) => ({
      id: runner.id,
      isEnabled: runner.isEnabled,
      maxConcurrentPods: runner.maxConcurrentPods,
    }));
    const createdPods: Pod[] = [];

    try {
      for (const runner of available) {
        await cc.runner.updateRunner({
          orgSlug: TEST_ORG_SLUG,
          id: runner.id,
          isEnabled: runner.id === primary.id,
          maxConcurrentPods: runner.id === primary.id || runner.id === secondary.id
            ? 1
            : runner.maxConcurrentPods,
        });
      }
      await expectAvailableRunnerIds(cc, [primary.id]);
      const primarySpec = await buildE2EEchoWorkerSpec(cc, {
        mode: "acp",
        scenario: "echo",
      });

      const first = await createPod(cc, primarySpec, "E2E capacity primary");
      createdPods.push(first);
      expect(first.runnerId).toBe(primary.id);
      await expectRunnerPods(cc, primary.id, 1);
      await expectAvailableRunnerIds(cc, []);
      await expect(createPod(cc, primarySpec, "E2E over capacity")).rejects.toThrow(
        /no available runner/i,
      );

      await cc.runner.updateRunner({
        orgSlug: TEST_ORG_SLUG,
        id: secondary.id,
        isEnabled: true,
        maxConcurrentPods: 1,
      });
      await expectAvailableRunnerIds(cc, [secondary.id]);
      const secondarySpec = await buildE2EEchoWorkerSpec(cc, {
        mode: "acp",
        scenario: "echo",
      });

      const second = await createPod(cc, secondarySpec, "E2E capacity secondary");
      createdPods.push(second);
      expect(second.runnerId).toBe(secondary.id);
      await expectRunnerPods(cc, secondary.id, 1);
      await expectAvailableRunnerIds(cc, []);
      await expect(createPod(cc, secondarySpec, "E2E fully saturated")).rejects.toThrow(
        /no available runner/i,
      );

      await terminatePods(cc, createdPods);
      await expectRunnerPods(cc, primary.id, 0);
      await expectRunnerPods(cc, secondary.id, 0);
      await expectAvailableRunnerIds(cc, [primary.id, secondary.id]);
      createdPods.length = 0;
    } finally {
      await terminatePods(cc, createdPods);
      for (const runner of originals) {
        await cc.runner.updateRunner({
          orgSlug: TEST_ORG_SLUG,
          id: runner.id,
          isEnabled: runner.isEnabled,
          maxConcurrentPods: runner.maxConcurrentPods,
        });
      }
    }
  });
});

async function createPod(
  client: ConnectClient,
  workerSpec: E2EWorkerSpecDraft,
  alias: string,
): Promise<Pod> {
  const created = await client.pod.createPod({
    orgSlug: TEST_ORG_SLUG,
    cols: 80,
    rows: 24,
    workerSpec: { ...workerSpec, alias },
  }) as { pod?: Pod };
  if (!created.pod?.podKey || !created.pod.runnerId) {
    throw new Error(`CreatePod returned incomplete placement for ${alias}`);
  }
  return created.pod;
}

async function expectRunnerPods(
  client: ConnectClient,
  runnerId: bigint,
  expected: number,
) {
  await pollUntil(async () => {
    const result = await client.runner.getRunner({
      orgSlug: TEST_ORG_SLUG,
      id: runnerId,
    }) as { runner?: Runner };
    return result.runner?.currentPods === expected;
  }, {
    maxAttempts: 10,
    intervalMs: 1_000,
    label: `runner-${runnerId}-pods-${expected}`,
  });
}

async function expectAvailableRunnerIds(
  client: ConnectClient,
  expected: bigint[],
) {
  await pollUntil(async () => {
    const { items } = await client.runner.listAvailableRunners({
      orgSlug: TEST_ORG_SLUG,
    }) as { items: Runner[] };
    const actual = items.map((runner) => runner.id);
    return actual.length === expected.length &&
      expected.every((id) => actual.includes(id));
  }, {
    maxAttempts: 10,
    intervalMs: 500,
    label: `available-runners-${expected.join(",") || "none"}`,
  });
}

async function terminatePods(client: ConnectClient, pods: Pod[]) {
  for (const pod of pods) {
    await client.pod.terminatePod({
      orgSlug: TEST_ORG_SLUG,
      podKey: pod.podKey,
    }).catch(() => undefined);
  }
}
