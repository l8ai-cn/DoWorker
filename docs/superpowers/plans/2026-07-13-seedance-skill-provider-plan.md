# Seedance Skill and Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish an adapted Seedance Expert skill and provide a deterministic Volcengine generation tool.

**Architecture:** Fork the MIT upstream repository, retain attribution, replace custom routing tokens with native progressive disclosure, and add a tested asynchronous API client. Credentials enter only through `SEEDANCE_*` environment variables.

**Tech Stack:** Markdown skills, Python standard library, unittest, GitHub.

---

## Current Verification Status

Verified on 2026-07-15:

- The adapted skill repository is published at commit
  `21fe97e4628b71adba4722c1e4971fca47003115`.
- Skill validation, prompt lint tests, and mocked provider client tests passed.
- `--check-credentials` succeeds with the configured Volcengine connection and
  does not create a generation task.
- The chat model `doubao-seed-1-8-251228` is usable. The official video model
  `doubao-seedance-2-0-260128` is not enabled for the account and returns
  `ModelNotOpen`.
- Real generation remains pending provider entitlement and explicit approval of
  the exact billable generation fingerprint.

---

### Task 1: Fork and Initialize the Skill Repository

**Files:**
- Modify: `SKILL.md`
- Modify: `agents/openai.yaml`
- Preserve: `LICENSE`

- [ ] Fork `Emily2040/seedance-2.0` as `l8ai-cn/seedance-expert-skill`.
- [ ] Record the upstream commit in the adapted skill body and keep MIT attribution.
- [ ] Run the skill creator validator:

```bash
python /Users/wwyz/.codex/skills/.system/skill-creator/scripts/quick_validate.py .
```

- [ ] Commit:

```bash
git add SKILL.md agents/openai.yaml LICENSE
git commit -m "feat: adapt Seedance expert skill for Codex"
```

### Task 2: Curate Progressive References

**Files:**
- Modify: `SKILL.md`
- Retain selected files under: `references/`
- Delete only unused upstream generated presentation assets and migrated legacy copies.

- [ ] Keep prompt, reference workflow, camera, continuity, safety, API, and retake references.
- [ ] Replace `[skill:name]` and `[ref:name]` navigation with relative Markdown links.
- [ ] Verify no broken local references:

```bash
rg -n '\[(skill|ref):' SKILL.md references
```

Expected: no matches.

- [ ] Commit:

```bash
git add SKILL.md references
git commit -m "refactor: make Seedance references self-contained"
```

### Task 3: Add Prompt Validation

**Files:**
- Create: `scripts/seedance_prompt_lint.py`
- Create: `tests/test_seedance_prompt_lint.py`

- [ ] Write failing tests for empty prompts, unresolved `@Image` references,
  conflicting camera moves, and excessive timestamp events.
- [ ] Run:

```bash
python -m unittest tests.test_seedance_prompt_lint -v
```

Expected: FAIL before implementation.

- [ ] Implement deterministic validation with actionable diagnostics.
- [ ] Re-run the test and expect PASS.
- [ ] Commit:

```bash
git add scripts/seedance_prompt_lint.py tests/test_seedance_prompt_lint.py
git commit -m "feat: lint Seedance production prompts"
```

### Task 4: Add the Volcengine Task Client

**Files:**
- Create: `scripts/seedance_generate.py`
- Create: `tests/test_seedance_generate.py`

- [ ] Write mocked tests for create, queued/running/succeeded polling, provider
  failure, timeout, MP4 download, and metadata output.
- [ ] Run:

```bash
python -m unittest tests.test_seedance_generate -v
```

Expected: FAIL before implementation.

- [ ] Implement create/query/download with bounded polling and no retries that
  change provider, model, or request semantics.
- [ ] Re-run both script test modules and expect PASS.
- [ ] Commit:

```bash
git add scripts/seedance_generate.py tests/test_seedance_generate.py
git commit -m "feat: generate Seedance videos through Volcengine"
```

### Task 5: Publish and Verify

- [ ] Run skill validation and all repository tests.
- [ ] Push `main` to `l8ai-cn/seedance-expert-skill`.
- [ ] Verify:

```bash
git fetch origin main
git branch -r --contains HEAD
git show --no-patch --oneline HEAD
```

Expected: `origin/main` contains the final commit.
