import { expect, type Page } from "@playwright/test";

export class FormalWorkerWizardPage {
  constructor(
    private page: Page,
    private orgSlug: string,
  ) {}

  async goto(): Promise<void> {
    await this.page.goto(`/${this.orgSlug}/workers/new`);
    await expect(
      this.page.getByRole("heading", { name: /^(Create Worker|创建 Worker)$/i }),
    ).toBeVisible();
  }

  async configureCodex(): Promise<void> {
    const workerType = this.page.getByLabel(/^(Worker Type|Worker 类型)$/i);
    await workerType.click();
    await this.page.getByRole("option", { name: "OpenAI Codex", exact: true }).click();
    const changeType = this.page.getByRole("button", {
      name: /^(Change type|Switch type|切换类型)$/i,
    });
    if (await changeType.isVisible()) await changeType.click();
    await expect(this.page.getByTestId("worker-runtime-field-model")).toBeVisible();
    await expect(workerType).toContainText("OpenAI Codex");

    await this.next();
    await expect(this.page.getByText("Approval Mode", { exact: true })).toBeVisible();

    await this.next();
    await expect(
      this.page.getByRole("heading", { name: /^(Configure Workspace|配置工作区)$/i }),
    ).toBeVisible();
  }

  async preflight(): Promise<void> {
    await this.next();
    await expect(
      this.page.getByText(/^(Worker configuration is ready\.|Worker 配置已就绪。)$/i),
    ).toBeVisible();
  }

  async create(): Promise<string> {
    await Promise.all([
      this.page.waitForURL(new RegExp(`/${this.orgSlug}/workspace\\?pod=`)),
      this.page.getByRole("button", { name: /^(Create Worker|创建 Worker)$/i }).click(),
    ]);
    const podKey = new URL(this.page.url()).searchParams.get("pod");
    if (!podKey) throw new Error("Worker creation did not return a pod key");
    return podKey;
  }

  async promptReady(): Promise<void> {
    const control = this.page.getByRole("button", {
      name: /^(Take control|Acquire Control|取得控制权)$/i,
    });
    await expect(control).toBeEnabled();
    await control.click();

    const prompt = this.page.getByRole("textbox", {
      name: /^(Send instruction(?:…|\.{3})|发送指令(?:…|\.{3}))$/i,
    });
    await expect(prompt).toBeEnabled();
    await prompt.fill("Reply with exactly READY. Do not modify files.");
    await prompt.press("Enter");
    await expect(this.page.getByText("READY", { exact: true })).toBeVisible({ timeout: 30_000 });
  }

  private async next(): Promise<void> {
    await this.page.getByRole("button", { name: /^(Next|下一步)$/i }).click();
  }
}
