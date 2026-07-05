const { chromium } = require('@playwright/test');
const fs = require('fs');
const BASE = 'http://localhost:10007';
const ORG = 'dev-org';
const OUT = 'output/playwright';
const log = (...a) => console.log(...a);
const shot = async (p, n) => { await p.screenshot({ path: `${OUT}/${n}`, fullPage: false }).catch(()=>{}); };

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport: { width: 1500, height: 950 } });
  const errs = [];
  page.on('pageerror', e => errs.push('PAGEERR: ' + e.message));

  // 1. Login
  await page.goto(BASE + '/login', { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.waitForTimeout(2000);
  await page.fill('#username', 'dev@agentsmesh.local');
  await page.fill('#password', 'AdminAb123456');
  await page.click('button:has-text("SIGN IN")');
  await page.waitForURL(u => !u.toString().includes('/login'), { timeout: 30000 }).catch(()=>{});
  await page.waitForTimeout(3000);
  log('LOGIN url=', page.url());

  // 2. Workspace
  await page.goto(`${BASE}/${ORG}/workspace`, { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.waitForTimeout(4000);
  await shot(page, 'da-01-workspace.png');

  // 3. Open New Pod modal
  await page.getByRole('button', { name: /new pod|create pod|新建/i }).first().click({ timeout: 15000 });
  await page.locator('[role="dialog"]').first().waitFor({ state: 'visible', timeout: 15000 });
  await page.waitForTimeout(1500);
  await shot(page, 'da-02-modal.png');

  // 4. Select do-agent + ACP
  const agentSel = page.locator('[role="dialog"] select#agent-select');
  await agentSel.waitFor({ state: 'visible', timeout: 20000 });
  await agentSel.selectOption('do-agent');
  await page.waitForTimeout(1200);
  log('AGENT selected = do-agent; options=', await agentSel.locator('option').allTextContents());
  // ACP mode toggle
  await page.getByRole('button', { name: /Conversational \(ACP\)|对话/i }).click({ timeout: 8000 }).catch(e=>log('ACP toggle err', e.message));
  await page.waitForTimeout(800);
  await shot(page, 'da-03-agent-acp.png');

  // 5. Create
  await page.locator('[role="dialog"]').getByRole('button', { name: /^Create Pod$|^创建/ }).click({ timeout: 8000 });
  log('CREATE clicked');
  await page.locator('[role="dialog"]').first().waitFor({ state: 'hidden', timeout: 20000 }).catch(()=>log('modal still visible'));
  await page.waitForTimeout(3000);
  await shot(page, 'da-04-after-create.png');

  // 6. Wait for ACP prompt input to appear (pod running + relay/ACP session)
  let ready = false;
  const input = page.getByPlaceholder(/Send instruction|发送指令/i).first();
  try {
    await input.waitFor({ state: 'visible', timeout: 120000 });
    ready = true;
  } catch (e) { log('ACP input not visible within 120s'); }
  log('ACP INPUT ready =', ready);
  await shot(page, 'da-05-agentpanel.png');

  const dump = async (tag) => {
    const txt = await page.locator('body').innerText().catch(()=> '');
    fs.writeFileSync(`${OUT}/da-stream-${tag}.txt`, txt);
    return txt;
  };

  // 7. Interactive turn 1 — deterministic round-trip
  if (ready) {
    await input.fill('Reply with exactly this token and nothing else: PONG-42');
    await input.press('Enter');
    log('SENT turn1');
    await page.waitForTimeout(2000);
    // wait up to 90s for response token
    let got1 = false;
    for (let i = 0; i < 45; i++) {
      const t = await page.locator('body').innerText().catch(()=> '');
      if (/PONG-42/.test(t) && (t.match(/PONG-42/g)||[]).length >= 2) { got1 = true; break; } // echoed in prompt + answer
      await page.waitForTimeout(2000);
    }
    await dump('turn1');
    await shot(page, 'da-06-turn1.png');
    log('TURN1 round-trip got PONG-42 =', got1);

    // 8. Interactive turn 2 — real task (file creation via tools)
    await input.fill('Create a file named hello.txt in the current working directory with the exact contents "Hello from do-agent via AgentsMesh", then run ls and tell me the files you see.');
    await input.press('Enter');
    log('SENT turn2 (file task)');
    let got2 = false;
    for (let i = 0; i < 75; i++) {
      const t = await page.locator('body').innerText().catch(()=> '');
      if (/hello\.txt/i.test(t)) { got2 = true; }
      // capture periodic screenshots of activity
      if (i === 10) await shot(page, 'da-07-turn2-mid.png');
      if (got2 && /hello\.txt/i.test(t)) {
        // give it a little more to finish streaming
      }
      await page.waitForTimeout(2000);
      if (i > 20 && got2) break;
    }
    await dump('turn2');
    await shot(page, 'da-08-turn2.png');
    log('TURN2 mentions hello.txt =', got2);
  }

  log('\n--- PAGE ERRORS ---\n' + (errs.slice(0, 6).join('\n') || '(none)'));
  await browser.close();
  process.exit(0);
})().catch(e => { console.error('SCRIPT_FAIL:', e.stack); process.exit(1); });
