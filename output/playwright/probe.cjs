const { chromium } = require('@playwright/test');
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport: { width: 1440, height: 900 } });
  const errors = [];
  page.on('console', m => { if (m.type()==='error') errors.push(m.text()); });
  await page.goto('http://localhost:10007/login', { waitUntil: 'networkidle', timeout: 60000 });
  await page.waitForTimeout(1500);
  await page.screenshot({ path: 'output/playwright/01-login.png' });
  // dump inputs and buttons
  const inputs = await page.$$eval('input', els => els.map(e => ({type:e.type, name:e.name, id:e.id, ph:e.placeholder})));
  const buttons = await page.$$eval('button', els => els.map(e => (e.innerText||'').trim()).filter(Boolean));
  console.log('URL:', page.url());
  console.log('INPUTS:', JSON.stringify(inputs));
  console.log('BUTTONS:', JSON.stringify(buttons));
  console.log('CONSOLE_ERRORS:', errors.slice(0,5));
  await browser.close();
})().catch(e => { console.error('PROBE_FAIL:', e.message); process.exit(1); });
