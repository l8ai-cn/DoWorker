const { chromium } = require('@playwright/test');
const BASE = 'http://localhost:10007';
const ORG = 'dev-org';
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport: { width: 1500, height: 950 } });
  page.on('console', m => { if (m.type()==='error') console.log('CONSOLE.ERR:', m.text().slice(0,500)); });
  page.on('pageerror', e => console.log('PAGEERR:', e.message.slice(0,500)));
  page.on('requestfailed', r => console.log('REQFAIL:', r.method(), r.url(), r.failure() && r.failure().errorText));
  page.on('response', async (r) => {
    if (/pods|agent|create/i.test(r.url()) && (r.request().method() === 'POST' || r.status()>=400)) {
      let body = '';
      try { body = await r.text(); } catch (e) {}
      console.log('RESP', r.request().method(), r.url(), '->', r.status(), '\nBODY:', body.slice(0, 1000));
    }
  });
  await page.goto(BASE + '/login', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(1500);
  await page.fill('#username', 'dev@agentsmesh.local');
  await page.fill('#password', 'devpass123');
  await page.click('button:has-text("SIGN IN")');
  await page.waitForURL(u => !u.toString().includes('/login'), { timeout: 30000 }).catch(()=>{});
  await page.waitForTimeout(3000);
  await page.goto(`${BASE}/${ORG}/workspace`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(4000);
  await page.getByRole('button', { name: /new pod|create pod/i }).first().click();
  await page.locator('[role="dialog"]').first().waitFor({ state: 'visible' });
  await page.locator('[role="dialog"] select#agent-select').selectOption('do-agent');
  await page.waitForTimeout(800);
  await page.getByRole('button', { name: /Conversational \(ACP\)/i }).click().catch(()=>{});
  await page.waitForTimeout(500);
  await page.locator('[role="dialog"]').getByRole('button', { name: /^Create Pod$/ }).click();
  await page.waitForTimeout(8000);
  await browser.close();
  process.exit(0);
})().catch(e => { console.error('FAIL', e.stack); process.exit(1); });
