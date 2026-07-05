import { chromium } from "@playwright/test";
import { mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const OUT = join(process.cwd(), "output", "browser-integration");
mkdirSync(OUT, { recursive: true });
const WEB_USER = "http://127.0.0.1:5173";
const API = "http://localhost:10015";
const ORG = "dev-org";
const SESSION = process.env.SESSION_ID ?? "conv_14f73d11ca94ddf2";

function authKey() {
  return "agentsmesh-auth/http_localhost_10000/session";
}

async function login() {
  const res = await fetch(`${API}/proto.auth.v1.AuthService/Login`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "Connect-Protocol-Version": "1" },
    body: JSON.stringify({ username: "devuser", password: "devpass123" }),
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
        base_url: "http://localhost:10000",
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
    console.log("follow-up excerpt:", body2.slice(0, 500).replace(/\n/g, " | "));
  }

  writeFileSync(join(OUT, "web-user-session-report.txt"), body);
  await browser.close();
}

main();
