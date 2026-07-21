import { chromium } from "@playwright/test";
import { mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { devUrl } from "./dev-runtime-env.mjs";

const OUT = join(process.cwd(), "output", "browser-integration");
mkdirSync(OUT, { recursive: true });
const WEB_USER = devUrl("WEB_USER_URL", "http://127.0.0.1:5173");
const API = devUrl("SESSION_COMPAT_API_URL", "http://localhost:10015");
const ORG = "dev-org";
const SESSION = process.env.SESSION_ID ?? "conv_14f73d11ca94ddf2";
const WEB_USER_AUTH_BASE = devUrl("WEB_USER_AUTH_URL", "http://localhost:10000");

function authKey() {
  const u = new URL(WEB_USER_AUTH_BASE);
  const port = u.port ? `_${u.port}` : "";
  const raw = `${u.protocol.replace(":", "")}_${u.hostname.toLowerCase()}${port}`;
  return `agent-cloud-auth/${raw.replace(/[^a-zA-Z0-9]/g, "_").slice(0, 64)}/session`;
}

async function login() {
  const res = await fetch(`${API}/proto.auth.v1.AuthService/Login`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "Connect-Protocol-Version": "1" },
    body: JSON.stringify({ username: "devuser", password: "AdminAb123456" }),
  });
  const data = await res.json();
  return { token: data.token, refresh: data.refreshToken ?? "", exp: Math.floor(Date.now() / 1000) + 3600 };
}

async function main() {
  const auth = await login();
  const browser = await chromium.launch({ headless: true });
  const ctx = await browser.newContext();
  await ctx.addInitScript(
    ({ key, blob }) => localStorage.setItem(key, JSON.stringify(blob)),
    {
      key: authKey(),
      blob: {
        access_token: auth.token,
        refresh_token: auth.refresh,
        expires_at: auth.exp,
        base_url: WEB_USER_AUTH_BASE,
        current_org_slug: ORG,
        schema_version: 1,
      },
    },
  );
  const page = await ctx.newPage();
  await page.goto(`${WEB_USER}/c/${SESSION}`, { waitUntil: "domcontentloaded", timeout: 60_000 });
  await page.waitForTimeout(8000);
  await page.screenshot({ path: join(OUT, "06-web-user-session-chat.png"), fullPage: true });

  const body = await page.locator("body").innerText();
  if (page.url().includes("/login") || /Sign in|Welcome to Agent Cloud/i.test(body)) {
    throw new Error("web-user session smoke reached login instead of the session page");
  }
  const hasUserMsg = /Say hello in one sentence/i.test(body);
  console.log("user message visible:", hasUserMsg);
  console.log("page excerpt:", body.slice(0, 400).replace(/\n/g, " | "));

  const composer = page.locator("textarea, [contenteditable='true']").last();
  if (await composer.isVisible().catch(() => false)) {
    await composer.fill("Follow up: what is 2+2? Reply with just the number.");
    await page.keyboard.press("Enter");
    await page.waitForTimeout(12000);
    await page.screenshot({ path: join(OUT, "07-web-user-followup.png"), fullPage: true });
    const body2 = await page.locator("body").innerText();
    if (/401|unauthorized|failed to fetch|stream unavailable|couldn.t send/i.test(body2)) {
      throw new Error(`web-user follow-up failed: ${body2.slice(0, 300)}`);
    }
    console.log("follow-up excerpt:", body2.slice(0, 500).replace(/\n/g, " | "));
  }

  writeFileSync(join(OUT, "web-user-session-report.txt"), body);
  await browser.close();
}

main();
