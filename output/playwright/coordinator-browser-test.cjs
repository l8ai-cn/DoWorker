const { chromium } = require('@playwright/test');

const BASE = 'http://localhost:10007';
const API = 'http://localhost:10000/api';
const ORG = 'dev-org';
const results = [];
const rec = (n, ok, d = '') => {
  results.push({ n, ok });
  console.log(`${ok ? '✅ PASS' : '❌ FAIL'}  ${n}${d ? '  — ' + d : ''}`);
};

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1440, height: 900 } });
  const errs = [];
  page.on('pageerror', (e) => errs.push(e.message));

  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.fill('#username', 'dev@agentsmesh.local');
  await page.fill('#password', 'AdminAb123456');
  await page.click('button:has-text("SIGN IN")');
  await page.waitForURL((u) => !u.toString().includes('/login'), { timeout: 30000 });
  await page.waitForTimeout(2000);
  rec('登录', !page.url().includes('/login'), page.url());

  await page.goto(`${BASE}/${ORG}/automation`, { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.waitForTimeout(3000);
  const body = await page.locator('body').innerText();
  const noApiErr = !/API Error|Internal Server Error|500/i.test(body);
  const hasAutomation = /Automation|自动化/i.test(body);
  await page.screenshot({ path: 'output/playwright/coord-01-automation.png', fullPage: true });
  rec('Automation 页面', hasAutomation && noApiErr, `title=${hasAutomation} noErr=${noApiErr}`);

  const newBtn = page.locator('button:has-text("New project"), button:has-text("新建")').first();
  if (await newBtn.isVisible().catch(() => false)) {
    await newBtn.click();
    await page.waitForTimeout(1200);
    const dlg = await page.locator('body').innerText();
    rec('创建项目对话框', /Repository|仓库|Name|名称/i.test(dlg));
    await page.screenshot({ path: 'output/playwright/coord-02-create-dialog.png' });
    await page.keyboard.press('Escape');
    await page.waitForTimeout(500);
  }

  const runBtns = page.locator('button:has-text("Run now"), button:has-text("立即运行")');
  const runCount = await runBtns.count();
  if (runCount > 0) {
    await runBtns.first().click();
    await page.waitForTimeout(4000);
    const afterRun = await page.locator('body').innerText();
    const runOk = !/API Error|Internal Server Error|failed to fetch/i.test(afterRun);
    await page.screenshot({ path: 'output/playwright/coord-03-run-now.png', fullPage: true });
    rec('Run now 触发', runOk, `projects=${runCount}`);
  } else {
    rec('Run now 触发', true, 'no projects yet (skipped)');
  }

  const token = await page.evaluate(() => localStorage.getItem('auth_token') || localStorage.getItem('token'));
  if (token) {
    const runners = await page.request.get(`${API}/v1/orgs/${ORG}/runners`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const runnersOk = runners.ok();
    let online = 0;
    if (runnersOk) {
      const data = await runners.json();
      const list = data.runners || data.data || data || [];
      online = list.filter((r) => r.status === 'online' || r.is_online).length;
    }
    rec('Runner API 在线检查', runnersOk, `online=${online}`);
  }

  console.log('\n--- PAGE ERRORS ---');
  console.log(errs.length ? errs.join('\n') : '(none)');
  const passed = results.filter((r) => r.ok).length;
  console.log(`\n=== RESULT: ${passed}/${results.length} PASS ===`);
  await browser.close();
  process.exit(passed === results.length ? 0 : 1);
})().catch((e) => {
  console.error('SCRIPT_FAIL:', e.stack);
  process.exit(1);
});
