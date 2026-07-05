import { chromium } from "@playwright/test";
import { mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const OUT = join(process.cwd(), "output", "browser-integration");
mkdirSync(OUT, { recursive: true });

const WEB = process.env.WEB_URL || "http://127.0.0.1:10007";
const WEB_USER = process.env.WEB_USER_URL || "http://127.0.0.1:5173";
const TRAEFIK_API = process.env.TRAEFIK_API_URL || "http://127.0.0.1:10000";
// web-user reads import.meta.env.VITE_AGENTSMESH_API_URL ?? "http://localhost:10000"
// for localStorage key — must match hostname, not 127.0.0.1.
const WEB_USER_AUTH_BASE = process.env.WEB_USER_AUTH_URL || "http://localhost:10000";
const API_DIRECT = process.env.HIVE_API_URL || "http://localhost:10015";
const ORG = "dev-org";
const USER = { username: "devuser", password: "AdminAb123456" };

function authKey(baseUrl) {
  const u = new URL(baseUrl);
  const port = u.port ? `_${u.port}` : "";
  const raw = `${u.protocol.replace(":", "")}_${u.hostname.toLowerCase()}${port}`;
  return `do-worker-auth/${raw.replace(/[^a-zA-Z0-9]/g, "_").slice(0, 64)}/session`;
}

async function setReactTextarea(page, testId, value) {
  await page.evaluate(
    ({ id, text }) => {
      const el = document.querySelector(`[data-testid="${id}"]`);
      if (!el || !(el instanceof HTMLTextAreaElement)) {
        throw new Error(`textarea ${id} not found`);
      }
      const setter = Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype, "value")?.set;
      setter?.call(el, text);
      el.dispatchEvent(new Event("input", { bubbles: true }));
    },
    { id: testId, text: value },
  );
}

async function login() {
  for (const base of [TRAEFIK_API, API_DIRECT]) {
    const res = await fetch(`${base}/proto.auth.v1.AuthService/Login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Connect-Protocol-Version": "1",
      },
      body: JSON.stringify(USER),
    });
    if (res.ok) {
      const data = await res.json();
      return {
        token: data.token,
        refreshToken: data.refreshToken ?? data.refresh_token ?? "",
        expiresAt: Math.floor(Date.now() / 1000) + Number(data.expiresIn ?? data.expires_in ?? 3600),
      };
    }
  }
  throw new Error("login failed on traefik and backend direct");
}

async function injectSession(context, storageBaseUrl, { token, refreshToken, expiresAt }, orgSlug) {
  const key = authKey(storageBaseUrl);
  await context.addInitScript(
    ({ key, blob }) => localStorage.setItem(key, JSON.stringify(blob)),
    {
      key,
      blob: {
        access_token: token,
        refresh_token: refreshToken,
        expires_at: expiresAt,
        base_url: storageBaseUrl,
        current_org_slug: orgSlug,
        schema_version: 1,
      },
    },
  );
}

async function listAgents(token) {
  const res = await fetch(`${API_DIRECT}/v1/agents`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Organization-Slug": ORG,
    },
  });
  const body = await res.json();
  return body.data ?? [];
}

const report = { steps: [], errors: [] };

function step(name, ok, detail = "") {
  report.steps.push({ name, ok, detail });
  console.log(`${ok ? "✓" : "✗"} ${name}${detail ? ` — ${detail}` : ""}`);
}

async function part1AgentsMesh(browser, auth) {
  const ctx = await browser.newContext();
  await injectSession(ctx, WEB, auth, ORG);
  const page = await ctx.newPage();
  page.setDefaultTimeout(60_000);

  await page.goto(`${WEB}/${ORG}/workspace`, { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(5000);
  await page.screenshot({ path: join(OUT, "01-web-workspace.png"), fullPage: true });

  const newPodBtn = page
    .getByRole("button", { name: /new pod|create new pod|新建 pod|create pod|创建/i })
    .first();
  await newPodBtn.waitFor({ state: "visible", timeout: 30_000 });
  await newPodBtn.click();

  const dialog = page.locator('[role="dialog"]').first();
  await dialog.waitFor({ state: "visible", timeout: 30_000 });
  await page.locator('[role="dialog"] .animate-spin').waitFor({ state: "hidden", timeout: 120_000 }).catch(() => {});
  await page.waitForTimeout(2000);
  await page.screenshot({ path: join(OUT, "02-web-create-pod-dialog.png") });

  const noAgents = await page
    .locator("text=/does not support any agents|暂不支持任何智能体|no online runners/i")
    .isVisible()
    .catch(() => false);
  if (noAgents) throw new Error("create pod dialog: no agents available");

  const agentSelect = dialog.locator("select#agent-select");
  await agentSelect.waitFor({ state: "visible", timeout: 90_000 });

  const agents = await listAgents(auth.token);
  const agent =
    agents.find((a) => a.id === "e2e-echo")?.id ??
    agents.find((a) => a.builtin)?.id ??
    agents[0]?.id;
  if (!agent) throw new Error("no agents from /v1/agents");
  await agentSelect.selectOption(agent);

  const prompt = dialog.locator("textarea").first();
  if (await prompt.isVisible().catch(() => false)) {
    await prompt.fill("Browser integration smoke — say hello briefly.");
  }

  await page.screenshot({ path: join(OUT, "02-web-create-pod-dialog.png") });
  const submit = dialog.getByRole("button", { name: /create|创建/i }).last();
  await submit.click();
  await dialog.waitFor({ state: "hidden", timeout: 120_000 });
  await page.waitForTimeout(3000);
  await page.screenshot({ path: join(OUT, "03-web-pod-created.png"), fullPage: true });

  step("AgentsMesh: create pod via browser", true, `agent=${agent}`);
  await ctx.close();
  return agent;
}

async function part2WebUser(browser, auth, agent) {
  const ctx = await browser.newContext();
  // Vite bakes VITE_AGENTSMESH_API_URL at build time (defaults to :10000).
  await injectSession(ctx, WEB_USER_AUTH_BASE, auth, ORG);
  const page = await ctx.newPage();
  page.setDefaultTimeout(90_000);

  await page.goto(WEB_USER, { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(6000);
  if (page.url().includes("/login")) {
    throw new Error(`web-user redirected to login — check JWT localStorage key`);
  }
  await page.screenshot({ path: join(OUT, "04-web-user-landing.png"), fullPage: true });

  const input = page.getByTestId("new-chat-landing-input");
  await input.waitFor({ state: "visible", timeout: 30_000 });

  await page.waitForFunction(
    () => {
      const chip = document.querySelector('[data-testid="new-chat-landing-workspace-chip"]');
      return chip && !chip.textContent?.includes("Working directory");
    },
    { timeout: 30_000 },
  );

  const agents = await listAgents(auth.token);
  if (agents.some((a) => a.id === "e2e-echo")) {
    await page.getByTestId("new-chat-landing-agent-select").click();
    await page.getByTestId("new-chat-landing-agent-e2e-echo").click();
    await page.getByRole("heading", { name: "What should we do?" }).click();
    await page.waitForTimeout(500);
  }

  const prompt = "Integration test: reply with one short greeting sentence.";
  await setReactTextarea(page, "new-chat-landing-input", prompt);
  await page.waitForFunction(
    () => !document.querySelector('[data-testid="new-chat-landing-submit"]')?.disabled,
    { timeout: 60_000 },
  );

  await page.getByTestId("new-chat-landing-input").press("Enter");
  await page.waitForURL(/\/c\//, { timeout: 90_000 });
  const createFailed = await page.locator('[data-testid="new-chat-landing-error"], text=/Couldn\'t create the session/i').isVisible().catch(() => false);
  if (createFailed) {
    const errText = await page.locator("body").innerText();
    throw new Error(`session create failed on landing: ${errText.slice(0, 300)}`);
  }
  await page.waitForTimeout(15000);
  await page.screenshot({ path: join(OUT, "05-web-user-after-send.png"), fullPage: true });

  const bodyText = await page.locator("body").innerText();
  const sessionId = page.url().match(/\/c\/([^/?#]+)/)?.[1] ?? "";
  const hasAssistant =
    sessionId.length > 0 &&
    /assistant|hello|hi|greeting|pong|echo/i.test(bodyText) &&
    !/Couldn't create|HTTP 403|400 Bad Request|unauthorized|login|failed to fetch/i.test(bodyText);

  step("Web User: send message in browser", hasAssistant, hasAssistant ? `session=${sessionId}` : bodyText.slice(0, 200));

  const token = auth.token;
  const sessionsRes = await fetch(`${API_DIRECT}/v1/sessions`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Organization-Slug": ORG,
    },
  });
  const sessions = await sessionsRes.json();
  const created = sessions.data?.find((s) => s.id === sessionId);
  step("API: session created from browser", !!created, sessionId || "no session id in URL");

  if (created) {
    let itemCount = 0;
    for (let i = 0; i < 20; i++) {
      const itemsRes = await fetch(`${API_DIRECT}/v1/sessions/${sessionId}/items`, {
        headers: {
          Authorization: `Bearer ${token}`,
          "X-Organization-Slug": ORG,
        },
      });
      const items = await itemsRes.json();
      itemCount = items.data?.length ?? 0;
      if (itemCount >= 2) break;
      await new Promise((r) => setTimeout(r, 2000));
    }
    step("API: conversation items persisted", itemCount >= 2, `items=${itemCount}`);
  }

  await ctx.close();
}

async function main() {
  const auth = await login();
  step("API login", true);

  const browser = await chromium.launch({ headless: true });
  let agent = "e2e-echo";
  try {
    try {
      agent = await part1AgentsMesh(browser, auth);
    } catch (err) {
      report.errors.push(`part1: ${String(err?.stack ?? err)}`);
      step("AgentsMesh: create pod via browser", false, String(err).split("\n")[0]);
    }
    await part2WebUser(browser, auth, agent);
  } catch (err) {
    report.errors.push(String(err?.stack ?? err));
    step("integration run", false, String(err));
  } finally {
    await browser.close();
    writeFileSync(join(OUT, "report.json"), JSON.stringify(report, null, 2));
  }

  const failed = report.steps.some((s) => !s.ok);
  if (failed) process.exit(1);
}

main();
