import { expect, test } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { FormalWorkerWizardPage } from "../../pages/formal-worker-wizard.page";

let podKey: string | undefined;

test.describe("Formal Worker wizard: Codex", () => {
  test.afterEach(async ({ api }) => {
    if (!podKey) return;
    const client = await api.connect();
    await client.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });
    podKey = undefined;
  });

  test("creates a Codex Worker and receives an ACP response", async ({ page }) => {
    const wizard = new FormalWorkerWizardPage(page, TEST_ORG_SLUG);

    await wizard.goto();
    await wizard.configureCodex();
    await wizard.preflight();
    podKey = await wizard.create();

    await expect(page.getByRole("textbox", {
      name: /^(Send instruction(?:…|\.{3})|发送指令(?:…|\.{3}))$/i,
    })).toBeVisible();
    await wizard.promptReady();
  });
});
