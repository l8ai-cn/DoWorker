// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { E2E_ECHO_AGENT_SLUG, resolveE2EPodCreateTargets } from "../../helpers/e2e-echo-runner";
import { terminateAllPods } from "../../helpers/pod-cleanup";
import { CreatePodModal } from "../../pages/modals/create-pod.modal";

test.describe("CreatePod UI", () => {
  test.afterEach(async () => {
    await terminateAllPods();
  });

  test("create flow leaves /workers/new and new pod appears", async ({ page, api }) => {
    const cc = await api.connect();
    await resolveE2EPodCreateTargets(cc);

    await terminateAllPods();

    await page.goto(`/${TEST_ORG_SLUG}/workspace`);
    await page.waitForLoadState("load");

    const podsBefore = await cc.pod.listPods({ orgSlug: TEST_ORG_SLUG }) as { total: bigint | number };
    const beforeTotal = Number(podsBefore.total);

    const newPodBtn = page
      .getByRole("button", { name: /new pod|create new pod|新建 pod|新建 worker|新建环境/i })
      .first();
    await newPodBtn.click();
    await page.waitForURL(new RegExp(`/${TEST_ORG_SLUG}/workers/new`), {
      timeout: 15_000,
    });

    const modal = new CreatePodModal(page);
    await modal.waitForOpen();
    await modal.selectAgent(E2E_ECHO_AGENT_SLUG);
    await modal.submit();

    await modal.waitForClosed(15_000);

    await expect
      .poll(async () => {
        const after = await cc.pod.listPods({ orgSlug: TEST_ORG_SLUG }) as { total: bigint | number };
        return Number(after.total);
      }, { timeout: 10_000 })
      .toBeGreaterThan(beforeTotal);
  });
});
