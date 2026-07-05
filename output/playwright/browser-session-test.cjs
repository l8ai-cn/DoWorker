const { chromium } = require('@playwright/test');
const { execSync } = require('child_process');

const BASE = 'http://localhost:10007';
const API = 'http://localhost:10000';
const ORG = 'dev-org';

function loginViaConnect() {
  const raw = execSync(
    `curl -sf -X POST '${API}/proto.auth.v1.AuthService/Login' -H 'Content-Type: application/json' -d '{"username":"dev@agentsmesh.local","password":"AdminAb123456"}'`,
    { encoding: 'utf8' },
  );
  return JSON.parse(raw);
}

function sessionKey() {
  return 'agentsmesh-auth/http_localhost_10007/session';
}

(async () => {
  const login = loginViaConnect();
  const expiresAt = Math.floor(Date.now() / 1000) + Number(login.expiresIn || 86400);
  const session = {
    schema_version: 1,
    access_token: login.token,
    refresh_token: login.refreshToken,
    expires_at: expiresAt,
    base_url: BASE,
    current_org_slug: ORG,
  };

  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1440, height: 900 } });
  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded' });
  await page.evaluate(
    ({ key, blob }) => localStorage.setItem(key, JSON.stringify(blob)),
    { key: sessionKey(), blob: session },
  );

  const results = [];
  const rec = (n, ok, d = '') => {
    results.push(ok);
    console.log(`${ok ? '✅' : '❌'} ${n}${d ? ' — ' + d : ''}`);
  };

  await page.goto(`${BASE}/${ORG}/automation`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(3500);
  const auto = await page.locator('body').innerText();
  await page.screenshot({ path: 'output/playwright/browser-automation.png', fullPage: true });
  rec('Automation 页面', /Automation/i.test(auto) && !/API Error|Internal Server Error/i.test(auto));

  await page.goto(`${BASE}/${ORG}/infra?tab=runners`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(3500);
  const infra = await page.locator('body').innerText();
  await page.screenshot({ path: 'output/playwright/browser-runners.png', fullPage: true });
  const online = (infra.match(/(\d+)\s*online/i) || [])[0] || 'n/a';
  rec('Infra Runners 页', /Runners|Runner/i.test(infra), online);

  const projects = await page.request.get(`${API}/api/v1/orgs/${ORG}/coordinator/projects`, {
    headers: { Authorization: `Bearer ${login.token}` },
  });
  rec('Coordinator API', projects.ok(), `status=${projects.status()}`);

  console.log(`\n=== ${results.filter(Boolean).length}/${results.length} PASS ===`);
  await browser.close();
  process.exit(results.every(Boolean) ? 0 : 1);
})().catch((e) => {
  console.error('FAIL:', e.stack || e.message);
  process.exit(1);
});
