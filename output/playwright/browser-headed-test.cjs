const { chromium } = require('@playwright/test');

const BASE = 'http://localhost:10007';
const ORG = 'dev-org';

(async () => {
  const browser = await chromium.launch({ headless: false, slowMo: 200 });
  const page = await browser.newPage({ viewport: { width: 1440, height: 900 } });

  console.log('→ 打开登录页');
  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.fill('#username', 'dev@agentsmesh.local');
  await page.fill('#password', 'devpass123');
  await page.click('button:has-text("SIGN IN")');
  await page.waitForURL((u) => !u.toString().includes('/login'), { timeout: 45000 }).catch(() => {});
  await page.waitForTimeout(3000);
  console.log('✅ 登录成功:', page.url());

  console.log('→ 打开 Automation 页面');
  await page.goto(`${BASE}/${ORG}/automation`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(4000);
  await page.screenshot({ path: 'output/playwright/browser-automation.png', fullPage: true });

  const body = await page.locator('body').innerText();
  console.log('Automation 标题:', /Automation/i.test(body) ? 'OK' : 'MISSING');
  console.log('API 错误:', /API Error|Internal Server Error/i.test(body) ? 'YES' : 'none');

  console.log('→ 打开 Infra / Runners');
  await page.goto(`${BASE}/${ORG}/infra?tab=runners`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(4000);
  await page.screenshot({ path: 'output/playwright/browser-runners.png', fullPage: true });
  const infra = await page.locator('body').innerText();
  const onlineMatch = infra.match(/(\d+)\s*online/i);
  console.log('Runners online:', onlineMatch ? onlineMatch[0] : 'unknown');

  console.log('→ 保持浏览器 20s 供查看…');
  await page.waitForTimeout(20000);
  await browser.close();
})().catch((e) => {
  console.error('FAIL:', e.message);
  process.exit(1);
});
