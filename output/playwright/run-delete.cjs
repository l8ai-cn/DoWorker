const { chromium } = require('@playwright/test');
const BASE = 'http://localhost:10007';
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport:{ width:1440, height:900 } });
  const caps=[]; page.on('response',r=>{ const u=r.url(); if(u.includes('/coordinator/projects')) caps.push(`${r.request().method()} ${u.split('/coordinator/projects')[1]||'/'} -> ${r.status()}`); });
  page.on('dialog', d => d.accept());
  await page.goto(`${BASE}/login`,{waitUntil:'domcontentloaded'}); await page.waitForTimeout(2000);
  await page.fill('#username','dev@agentsmesh.local'); await page.fill('#password','devpass123');
  await page.click('button:has-text("SIGN IN")');
  await page.waitForURL(u=>!u.toString().includes('/login'),{timeout:30000}).catch(()=>{});
  await page.waitForTimeout(2500);
  await page.goto(`${BASE}/dev-org/automation`,{waitUntil:'domcontentloaded'}); await page.waitForTimeout(3000);

  // Run now
  if (await page.locator('button:has-text("Run now")').count()) {
    await page.locator('button:has-text("Run now")').first().click();
    await page.waitForTimeout(3000);
    console.log('Run now clicked');
  }
  // Delete (cleanup)
  if (await page.locator('button:has-text("Delete")').count()) {
    await page.locator('button:has-text("Delete")').first().click();
    await page.waitForTimeout(3000);
    console.log('Delete clicked');
  }
  const txt = await page.locator('body').innerText();
  console.log('empty state after delete =', /No automation projects yet/i.test(txt));
  await page.screenshot({ path:'output/playwright/07-after-delete.png', fullPage:true });
  console.log('API calls:\n  '+caps.join('\n  '));
  await browser.close();
})().catch(e=>{console.error('FAIL',e.stack);process.exit(1)});
