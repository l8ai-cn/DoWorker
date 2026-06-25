const { chromium } = require('@playwright/test');
const BASE = 'http://localhost:10007';
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  await page.goto(`${BASE}/login`, { waitUntil:'domcontentloaded' });
  await page.waitForTimeout(2000);
  await page.fill('#username','dev@agentsmesh.local');
  await page.fill('#password','devpass123');
  await page.click('button:has-text("SIGN IN")');
  await page.waitForURL(u=>!u.toString().includes('/login'),{timeout:30000}).catch(()=>{});
  await page.waitForTimeout(2000);
  const caps = [];
  page.on('response', async (resp) => {
    if (resp.url().includes('/coordinator/projects')) {
      let body=''; try{ body=await resp.text(); }catch{}
      caps.push({ url:resp.url(), status:resp.status(), body: body.slice(0,600) });
    }
  });
  await page.goto(`${BASE}/dev-org/automation`, { waitUntil:'domcontentloaded' });
  await page.waitForTimeout(4000);
  console.log(JSON.stringify(caps, null, 2));
  await browser.close();
})().catch(e=>{console.error('FAIL',e.message);process.exit(1)});
