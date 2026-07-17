# Agent Workbench Real Runner E2E Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prove programming, HTML, image editing, PPT, video, mobile, plain mount, and iframe behavior against a real connected Runner.

**Architecture:** A Python driver creates real sessions, sends deterministic task prompts, waits on V2 receipts and session state, and writes a machine-readable evidence manifest. Playwright opens those sessions in each host and validates rendering and interaction.

**Tech Stack:** dev.sh, REST/SSE, Python, Codex Runner, Playwright, ffprobe, PDF/PPTX inspection.

---

### Task 1: Add Deterministic Capability Preflight

**Files:**
- Modify: `deploy/dev/scripts/audit-deps-and-api.sh`
- Create: `deploy/dev/scripts/agent_workbench_e2e_preflight.sh`
- Create: `deploy/dev/scripts/agent_workbench_e2e_preflight_test.sh`

- [ ] **Step 1: Write failing preflight cases**

```bash
expect_fail "runner offline" env RUNNER_ONLINE=false "$SCRIPT"
expect_fail "Codex auth missing" env CODEX_AUTH_FILE=/missing "$SCRIPT"
expect_fail "video provider missing" env LOVART_API_KEY= "$SCRIPT" --capability video
```

- [ ] **Step 2: Implement exact checks**

Check PostgreSQL, Redis, Backend, Relay, gRPC, Web, Web User, one Runner heartbeat within 60 seconds, Codex binary/auth, Playwright Chromium, image-generation capability, PPT build dependencies, video provider credential, `ffprobe`, and preview origin. Each missing capability exits nonzero with one stable error code.

- [ ] **Step 3: Run and commit**

Run: `bash deploy/dev/scripts/agent_workbench_e2e_preflight_test.sh`
Expected: PASS for every explicit failure and a configured success fixture.

```bash
git add deploy/dev/scripts/audit-deps-and-api.sh deploy/dev/scripts/agent_workbench_e2e_preflight.sh deploy/dev/scripts/agent_workbench_e2e_preflight_test.sh
git commit -m "test(e2e): add workbench capability preflight"
```

### Task 2: Drive Real Programming And HTML Tasks

**Files:**
- Create: `deploy/dev/scripts/agent_workbench_real_e2e.py`
- Create: `deploy/dev/scripts/agent_workbench_prompts.json`
- Modify: `deploy/dev/scripts/verify-codex-pipeline.sh`
- Modify: `deploy/dev/scripts/gomoku_real_e2e.py`

- [ ] **Step 1: Define deterministic task output**

Programming prompt requires `artifacts/programming/index.html`, `app.js`, `styles.css`, `result.json`, and a follow-up code edit. `result.json` contains task ID, output files, assertions, and no timestamp-dependent value.

- [ ] **Step 2: Implement the driver**

The driver logs in, selects an online Runner, creates a `codex-cli` session, posts a generated V2 `send_prompt` command, waits for terminal receipt plus session `idle`, reads files through the session filesystem API, and records session ID, pod key, artifact IDs, revisions, receipt IDs, and assertions.

- [ ] **Step 3: Run the real task**

Run: `./deploy/dev/dev.sh && bash deploy/dev/scripts/agent_workbench_e2e_preflight.sh --capability programming && python3 deploy/dev/scripts/agent_workbench_real_e2e.py programming`
Expected: exit 0; real assistant response exists; HTML, JS, CSS, and result JSON exist; follow-up changes the file revision; no mock endpoint is called.

- [ ] **Step 4: Commit**

```bash
git add deploy/dev/scripts/agent_workbench_real_e2e.py deploy/dev/scripts/agent_workbench_prompts.json deploy/dev/scripts/verify-codex-pipeline.sh deploy/dev/scripts/gomoku_real_e2e.py
git commit -m "test(e2e): drive real programming workbench tasks"
```

### Task 3: Drive Real Image, PPT, And Video Tasks

**Files:**
- Modify: `deploy/dev/scripts/agent_workbench_prompts.json`
- Modify: `deploy/dev/scripts/agent_workbench_real_e2e.py`
- Create: `deploy/dev/scripts/validate_workbench_artifacts.py`
- Create: `deploy/dev/scripts/validate_workbench_artifacts_test.py`

- [ ] **Step 1: Define required outputs**

Image task produces source PNG, edited PNG, mask PNG, and metadata with dimensions, orientation, base revision, and edit action ID. PPT task produces PPTX, PDF preview, thumbnails, and slide manifest with stable IDs. Video task produces MP4, poster, thumbnails, and metadata with duration, dimensions, codec, and generation stages.

- [ ] **Step 2: Implement validators**

```py
assert image.size == (metadata["width"], metadata["height"])
assert pptx_slide_count(deck) == len(manifest["slides"])
assert pdf_page_count(preview) == len(manifest["slides"])
assert ffprobe(video)["duration"] > 0
assert metadata["stages"][-1]["state"] == "succeeded"
```

- [ ] **Step 3: Run real capabilities**

Run: `bash deploy/dev/scripts/agent_workbench_e2e_preflight.sh --capability image,presentation,video && python3 deploy/dev/scripts/agent_workbench_real_e2e.py image presentation video`
Expected: exit 0; each task has real receipts, artifact revisions, valid binaries, and no fixture substitution. Missing provider capability fails before session creation.

- [ ] **Step 4: Commit**

```bash
git add deploy/dev/scripts/agent_workbench_prompts.json deploy/dev/scripts/agent_workbench_real_e2e.py deploy/dev/scripts/validate_workbench_artifacts.py deploy/dev/scripts/validate_workbench_artifacts_test.py
git commit -m "test(e2e): validate real media and presentation tasks"
```

### Task 4: Add Browser E2E For All Hosts

**Files:**
- Create: `clients/web-user/playwright.config.ts`
- Create: `clients/web-user/e2e/real-runner-workbench.spec.ts`
- Modify: `clients/web-user/e2e/embed-host.html`
- Create: `clients/web-user/e2e/plain-mount-host.html`

- [ ] **Step 1: Add desktop and 390px scenarios**

```ts
await expect(page.getByRole("tab", { name: "结果" })).toBeVisible();
await page.getByRole("tab", { name: "结果" }).click();
await expect(page.locator("video")).toHaveJSProperty("duration", expect.any(Number));
await expect(page.getByRole("button", { name: "重新生成当前页" })).toBeEnabled();
```

- [ ] **Step 2: Cover interaction paths**

Desktop validates resizable conversation/results layout, collapsed commands, code diff, live preview, image slider/annotation/edit receipt, video playback/seek, PPT navigation/edit/export, and terminal lease. Repeat essential flows at 390px in Web User, plain mount, and iframe.

- [ ] **Step 3: Capture deterministic evidence**

For each host record screenshot, console errors, failed network requests, visible artifact ID/revision, and command receipt. Tests reject horizontal overflow, default browser dialogs, top-level untrusted HTML, and missing Chinese labels.

- [ ] **Step 4: Run and commit**

Run: `pnpm exec playwright test --config clients/web-user/playwright.config.ts`
Expected: PASS for desktop, mobile, plain mount, iframe, reconnect, read-only, error, and loading projects.

```bash
git add clients/web-user/playwright.config.ts clients/web-user/e2e/real-runner-workbench.spec.ts clients/web-user/e2e/embed-host.html clients/web-user/e2e/plain-mount-host.html
git commit -m "test(e2e): verify real workbench hosts and media"
```

### Task 5: Run The Full Acceptance Gate

**Files:**
- Create: `deploy/dev/scripts/verify-agent-workbench.sh`
- Modify: `README.md`

- [ ] **Step 1: Compose the gate**

Run security tests, proto generation checks, Go tests, Cargo workspace tests, WASM build, shared UI tests, Web/Web User typecheck and tests, real task driver, and Playwright in that order. Exit on the first failure.

- [ ] **Step 2: Execute**

Run: `bash deploy/dev/scripts/verify-agent-workbench.sh`
Expected: exit 0 and evidence manifest contains programming, image, presentation, and video session IDs plus desktop/mobile/plain/iframe screenshots.

- [ ] **Step 3: Document and commit**

Document required credentials, exact local URLs, artifact paths, known provider quotas, and rerun commands. Do not document a mock or degraded mode.

```bash
git add deploy/dev/scripts/verify-agent-workbench.sh README.md
git commit -m "docs: add agent workbench acceptance gate"
```
