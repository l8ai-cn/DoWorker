# Seedance Worker and Artifact Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bind a separate Seedance video resource to a dedicated Worker, render its MP4 output, publish an Expert, and prove the browser workflow.

**Architecture:** Extend Worker Definitions and WorkerSpec with exact, role-based tool model bindings. Use the existing do-agent runtime for conversation and inject the verified video binding through dedicated environment names. Render output-file events through the existing session file endpoint.

**Tech Stack:** Go, Protobuf, PostgreSQL, Next.js, React, TypeScript, Vitest, Playwright.

---

## Current Verification Status

Verified on 2026-07-15:

- The `seedance-expert` Worker type resolves the do-agent runtime, chat resource
  `4`, video resource `2`, runtime image `3`, and skill `1`.
- Credential-only Workers completed without creating a video task. Pod
  `7-standalone-26ec3d80` persisted the initial user task at conversation
  position `1` before tool events.
- Video resource `2` remains disabled at revision `7`. Volcengine accepts the
  credential, but the official Seedance model returns `ModelNotOpen`; no
  billable generation request is authorized while that entitlement is absent.
- Pod `7-standalone-87a81496` created
  `output/seedance-platform-visual-proof.mp4` as a non-provider display-chain
  fixture. The workspace API returned `video/mp4`, Base64 content, and `42928`
  bytes.
- Browser verification at `http://localhost:10007/dev-org/workspace` loaded the
  640x360, 4-second MP4, played it to completion, and reported zero console
  errors. Evidence:
  `output/playwright/seedance-platform-visual-proof-browser.png`.
- Focused backend tests, the web-user video artifact tests, and the complete Web
  Vitest suite passed (`264` files, `2115` tests).
- Expert Run now provisions a session and persists the effective initial user
  message before Runner dispatch. A nonblank prompt override wins; otherwise
  the validated WorkerSpec `workspace.initial_task` is used. Snapshot mismatch
  and snapshot-service failures return HTTP `409` and `503` respectively.
- The complete isolated backend suite passed after the Expert Run fix:
  `go test ./backend/... -count=1`.
- A transactional PostgreSQL probe executed `000207`, `000208`, and `000210`
  in order against a temporary schema. Cursor resolved to
  `cursor-acp` / `agent` / `pty,acp`; Seedance resolved to
  `do-agent-acp` / `do-agent` / `pty,acp`. The transaction was rolled back.
- Platform commit and push remain gated on the uncommitted prerequisite
  migrations `000207` through `000209`. Do not renumber `000210`.

---

### Task 1: Define Tool Model Requirements

**Files:**
- Modify: `config/worker-types/schema/definition.schema.json`
- Modify: `backend/internal/service/workerdefinition/`
- Test: `backend/internal/service/workerdefinition/*_test.go`

- [ ] Add failing loader tests for required role, provider, modality, capability,
  protocol adapter, and unique environment targets.
- [ ] Run:

```bash
go test ./backend/internal/service/workerdefinition/... -run ToolModel -count=1
```

- [ ] Implement `tool_model_requirements` without weakening primary model rules.
- [ ] Re-run and expect PASS.

### Task 2: Snapshot Tool Model Bindings

**Files:**
- Modify: `proto/pod/v1/worker_creation.proto`
- Modify: `backend/internal/domain/workerspec/`
- Modify: `backend/internal/service/workercreation/`
- Test: corresponding Go and proto contract tests.

- [ ] Add failing tests for exact resolution, duplicate roles, incompatible video
  resources, and snapshot identity/revision drift.
- [ ] Generate proto stubs with `pnpm proto:gen-go-all`.
- [ ] Implement role-based draft references and immutable bindings.
- [ ] Run focused WorkerSpec and Worker creation tests.

### Task 3: Add the Dedicated Worker Definition

**Files:**
- Create: `config/worker-types/seedance-expert/AgentFile`
- Create: `config/worker-types/seedance-expert/definition.json`
- Modify: `config/worker-types/catalog.json`
- Modify: runtime catalog fixtures and generated catalogs.

- [ ] Add a contract test proving the worker uses do-agent, requires a primary
  chat model, requires one Doubao video-generation resource, and exposes no
  plaintext credential field.
- [ ] Regenerate catalogs using repository scripts.
- [ ] Run worker-definition and runtime-catalog contract tests.

### Task 4: Inject and Revalidate Video Credentials

**Files:**
- Modify: `backend/internal/service/agentpod/`
- Test: focused environment, create, and resume tests.

- [ ] Add failing tests for `SEEDANCE_API_KEY`, `SEEDANCE_BASE_URL`, and
  `SEEDANCE_MODEL`, plus conflict and revision-drift failures.
- [ ] Resolve every exact tool resource during create and resume.
- [ ] Inject only definition-declared environment names.
- [ ] Run:

```bash
go test ./backend/internal/service/agentpod/... -run 'ToolModel|Seedance' -count=1
```

### Task 5: Add Worker Form Selection

**Files:**
- Modify: `clients/web/src/components/pod/CreatePodForm/`
- Modify: `clients/web/src/lib/api/connect/podWorkerDraftProto.ts`
- Test: focused Vitest files.

- [ ] Add failing tests for required video picker, loading, empty, invalid,
  disabled, and exact submitted resource ID states.
- [ ] Implement controls with existing design tokens and worker-definition data.
- [ ] Run `pnpm run web:test -- --run` for focused tests and typecheck.

### Task 6: Render Output Files

**Files:**
- Modify: `clients/web-user/src/lib/renderItems.ts`
- Create: `clients/web-user/src/components/OutputFileArtifact.tsx`
- Test: focused web-user tests.

- [ ] Add failing tests for MP4 player, generic download, missing filename,
  unsafe path, loading, and playback error.
- [ ] Resolve output files through the session resource endpoint.
- [ ] Render native video controls without nested cards or new theme tokens.
- [ ] Run focused Vitest and typecheck.

### Task 7: Publish and Run the Expert

- [ ] Import `l8ai-cn/seedance-expert-skill` through the platform.
- [ ] Securely configure a rotated Doubao key and separate chat/video resources.
- [ ] Create a `seedance-expert` Worker and run mocked generation first.
- [ ] Run one approved low-duration real Seedance generation.
- [ ] Confirm the MP4 plays and browser console/network show no relevant errors.
- [ ] Publish the verified Worker as `Seedance Expert`.
- [ ] Create a second Worker from the Expert and repeat a non-billing preflight.
- [ ] Commit and push only Seedance task files, then verify remote containment.
