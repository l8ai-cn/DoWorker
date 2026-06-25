const { chromium } = require('@playwright/test');
const fs = require('fs');
const BASE = 'http://localhost:10007';
const ORG = 'dev-org';
const POD = process.env.POD_KEY || '1-standalone-dbab661d';
const OUT = 'output/playwright';
const log = (...a) => console.log(...a);
const shot = async (p, n) => { await p.screenshot({ path: `${OUT}/${n}`, fullPage: false }).catch(() => {}); };

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport: { width: 1500, height: 950 } });
  page.on('pageerror', e => log('PAGEERR:', e.message));

  await page.goto(BASE + '/login', { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.waitForTimeout(1500);
  await page.fill('#username', 'dev@agentsmesh.local');
  await page.fill('#password', 'devpass123');
  await page.click('button:has-text("SIGN IN")');
  await page.waitForURL(u => !u.toString().includes('/login'), { timeout: 30000 }).catch(() => {});
  await page.waitForTimeout(3000);
  log('LOGIN url=', page.url());

  const url = `${BASE}/${ORG}/do-agent/${POD}`;
  await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.waitForTimeout(5000);
  await shot(page, 'dac-01-console.png');

  const input = page.getByPlaceholder(/Send instruction|发送指令|message|prompt/i).first();
  let ready = false;
  try { await input.waitFor({ state: 'visible', timeout: 60000 }); ready = true; }
  catch (e) { log('input not visible in 60s'); }
  log('INPUT ready =', ready);
  await shot(page, 'dac-02-input.png');
  if (!ready) {
    fs.writeFileSync(`${OUT}/dac-body.txt`, await page.locator('body').innerText().catch(() => ''));
    await browser.close(); process.exit(2);
  }

  // Turn 1: deterministic round-trip
  await input.fill('Reply with exactly this token and nothing else: PONG-42');
  await input.press('Enter');
  log('SENT turn1');
  let got1 = false;
  for (let i = 0; i < 40; i++) {
    const t = await page.locator('body').innerText().catch(() => '');
    if ((t.match(/PONG-42/g) || []).length >= 2) { got1 = true; break; }
    await page.waitForTimeout(2000);
  }
  fs.writeFileSync(`${OUT}/dac-turn1.txt`, await page.locator('body').innerText().catch(() => ''));
  await shot(page, 'dac-03-turn1.png');
  log('TURN1 got PONG-42 =', got1);

  // Turn 2: real task — create a file
  await input.fill('Create a file named hello.txt in the current directory containing exactly: Hello from do-agent via AgentsMesh. Then run `ls` and report the result.');
  await input.press('Enter');
  log('SENT turn2 (file task)');
  let got2 = false;
  for (let i = 0; i < 75; i++) {
    const t = await page.locator('body').innerText().catch(() => '');
    if (/hello\.txt/i.test(t)) got2 = true;
    if (i === 12) await shot(page, 'dac-04-turn2-mid.png');
    await page.waitForTimeout(2000);
    if (got2 && i > 18) break;
  }
  fs.writeFileSync(`${OUT}/dac-turn2.txt`, await page.locator('body').innerText().catch(() => ''));
  await shot(page, 'dac-05-turn2.png');
  log('TURN2 mentions hello.txt =', got2);

  await browser.close();
  process.exit(0);
})().catch(e => { console.error('SCRIPT_FAIL:', e.stack); process.exit(1); });
