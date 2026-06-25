const { chromium } = require('@playwright/test');
const BASE = 'http://localhost:10007';
const results = [];
const rec = (n, ok, d='') => { results.push({n, ok}); console.log(`${ok?'✅ PASS':'❌ FAIL'}  ${n}${d?'  — '+d:''}`); };
const go = async (p, path) => { const r = await p.goto(BASE+path, { waitUntil:'domcontentloaded', timeout:45000 }); await p.waitForTimeout(2500); return r; };

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport:{ width:1440, height:900 } });
  const errs = [];
  page.on('pageerror', e => errs.push('PAGEERR: '+e.message));

  // T1 login
  await go(page, '/login');
  await page.fill('#username','dev@agentsmesh.local');
  await page.fill('#password','devpass123');
  await page.click('button:has-text("SIGN IN")');
  await page.waitForURL(u=>!u.toString().includes('/login'),{timeout:30000}).catch(()=>{});
  await page.waitForTimeout(3000);
  rec('T1 登录', !page.url().includes('/login'), page.url());
  const org='dev-org';

  // T2 nav
  for (const name of ['workspace','tickets','channels','mesh','loops','automation','infra','settings']) {
    const resp = await go(page, `/${org}/${name}`);
    const txt = await page.locator('body').innerText().catch(()=> '');
    const ok = resp && resp.status()<400 && txt.length>40 && !/page not found/i.test(txt);
    rec(`T2 ${name}`, ok, `status=${resp&&resp.status()}`);
  }

  // T3 automation page (banner should be gone)
  await go(page, `/${org}/automation`);
  let txt = await page.locator('body').innerText();
  const noErr = !/API Error|Internal Server Error/i.test(txt);
  const hasTitle = /Automation/i.test(txt);
  await page.screenshot({ path:'output/playwright/03-automation.png', fullPage:true });
  rec('T3 Automation 渲染(无错误横幅)', hasTitle && noErr, `title=${hasTitle} noErr=${noErr}`);

  // T4 open dialog + repo dropdown populated
  await page.locator('button:has-text("New project")').first().click();
  await page.waitForTimeout(1200);
  txt = await page.locator('body').innerText();
  const fields = /Repository|Name|Label|Interval|仓库|名称/i.test(txt);
  await page.screenshot({ path:'output/playwright/04-create-dialog.png' });
  rec('T4 创建对话框渲染', fields && errs.length===0, `fields=${fields} pageErrs=${errs.length}`);

  // open repo select and count options
  let optCount = 0;
  try {
    await page.locator('[role="combobox"], button:has([data-placeholder]), [data-slot="select-trigger"]').first().click({ timeout:5000 });
    await page.waitForTimeout(800);
    optCount = await page.locator('[role="option"]').count();
  } catch(e) {}
  await page.screenshot({ path:'output/playwright/05-repo-options.png' });
  rec('T4 仓库下拉填充', optCount>0, `options=${optCount}`);

  // T5 create a project end-to-end
  let created=false;
  try {
    const uniq = 'e2e-auto-'+Date.now().toString().slice(-6);
    // pick first repo option if open
    if (optCount>0) await page.locator('[role="option"]').first().click();
    await page.waitForTimeout(400);
    // fill name (first text input in dialog)
    const nameInput = page.locator('input[type="text"], input:not([type])').first();
    await nameInput.fill(uniq);
    await page.waitForTimeout(300);
    await page.locator('button:has-text("Create"), button:has-text("Submit"), button:has-text("创建")').last().click();
    await page.waitForTimeout(3500);
    txt = await page.locator('body').innerText();
    created = txt.includes(uniq);
    await page.screenshot({ path:'output/playwright/06-after-create.png', fullPage:true });
    rec('T5 创建项目并显示', created, `name=${uniq}`);
  } catch(e) { rec('T5 创建项目并显示', false, e.message); }

  console.log('\n--- PAGE ERRORS ---\n'+(errs.slice(0,8).join('\n')||'(none)'));
  const passed = results.filter(r=>r.ok).length;
  console.log(`\n=== RESULT: ${passed}/${results.length} PASS ===`);
  await browser.close();
  process.exit(0);
})().catch(e=>{console.error('SCRIPT_FAIL:',e.stack);process.exit(1)});
