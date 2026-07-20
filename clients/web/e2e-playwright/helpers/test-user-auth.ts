import type { Browser } from "@playwright/test";
import { TEST_USER, getApiBaseUrl, getWebBaseUrl } from "./env";
import { terminateStaleMarkedE2EPods } from "./pod-cleanup";

export async function authenticateE2ETestUser(browser: Browser): Promise<void> {
  await terminateStaleMarkedE2EPods();
  const apiBaseUrl = getApiBaseUrl();
  const loginRes = await fetch(`${apiBaseUrl}/proto.auth.v1.AuthService/Login`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "Connect-Protocol-Version": "1" },
    body: JSON.stringify({ username: TEST_USER.username, password: TEST_USER.password }),
  });
  if (!loginRes.ok) throw new Error(`login failed: ${loginRes.status}`);
  const data = await loginRes.json();
  const token = data.token;
  const refreshToken = data.refreshToken ?? data.refresh_token;
  const expiresIn = Number(data.expiresIn ?? data.expires_in ?? 3600);
  const baseUrl = getWebBaseUrl();
  const expiresAt = Math.floor(Date.now() / 1000) + expiresIn;

  const context = await browser.newContext();
  await context.addInitScript(
    ({ accessToken, refreshToken: nextRefreshToken, nextExpiresAt, nextBaseUrl }) => {
      const url = new URL(nextBaseUrl);
      const port = url.port ? `_${url.port}` : "";
      const raw = `${url.protocol.replace(":", "")}_${url.hostname.toLowerCase()}${port}`;
      const slug = raw.replace(/[^a-zA-Z0-9]/g, "_").slice(0, 64);
      localStorage.setItem(`do-worker-auth/${slug}/session`, JSON.stringify({
        access_token: accessToken,
        refresh_token: nextRefreshToken,
        expires_at: nextExpiresAt,
        base_url: nextBaseUrl,
        current_org_slug: null,
        schema_version: 1,
      }));
    },
    {
      accessToken: token,
      refreshToken,
      nextExpiresAt: expiresAt,
      nextBaseUrl: baseUrl,
    },
  );
  const page = await context.newPage();
  await page.goto(baseUrl, { waitUntil: "domcontentloaded" });
  await context.storageState({ path: ".auth/user.json" });
  await context.close();
}
