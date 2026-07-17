# Worker Creation Experience Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the Worker creation flow understandable and usable for Chinese users, enable the real local MiniMax, OpenClaw, and DoAgent runtimes, and preserve the complete draft across related navigation.

**Architecture:** Keep the backend runtime catalog as the source of truth for selectable runtimes. Extend the existing Worker creation options/spec contract instead of adding a second form model. Keep draft state in the route-level controller and persist a sanitized draft snapshot in session storage so navigation restores the current step without storing plaintext secrets.

**Tech Stack:** Go, Connect/protobuf, Next.js App Router, React, TypeScript, Tailwind project tokens, Vitest, Go tests, in-app browser.

---

### Task 1: Make local runtime availability truthful

**Files:**
- Modify: `scripts/generate-local-worker-runtime-catalog.mjs`
- Modify: `deploy/dev/lib/worker_runtime_catalog.sh`
- Modify: `backend/internal/domain/workerruntime/runtime_catalog.lock.json` only if the release catalog contract requires a checked-in entry
- Test: `scripts/__tests__/generate-local-worker-runtime-catalog.test.mjs` or the existing script test location
- Test: `backend/internal/domain/workerruntime/catalog_test.go`
- Test: `backend/internal/service/workercreation/options_test.go`

- [ ] Add a failing generator test proving local catalog generation accepts `minimax-cli`, `openclaw`, and `do-agent`, preserves immutable Docker digests, and maps each runtime to exactly one worker type.
- [ ] Run the focused script test and confirm it fails because the current metadata map and service map only accept Codex and Gemini.
- [ ] Extend the local metadata map and `local_worker_runner_services` mapping for the three local runtimes.
- [ ] Ensure the generator still removes the catalog when no inspected image has a sha256 digest; do not mark an unverified image selectable.
- [ ] Run the generator test and the shell contract test.
- [ ] Start the three corresponding runner services from the active development compose project and verify they register as online through the existing runner/API path.
- [ ] Query Worker creation options and confirm all three worker types have a selectable runtime image and no “No runtime image” blocking reason.

### Task 2: Add a concrete resource contract

**Files:**
- Inspect and modify the existing Worker spec/resource types under `backend/internal/domain/workerspec/`
- Modify: `backend/internal/service/workercreation/options.go`
- Modify: `backend/internal/api/connect/pod/worker_creation_options.go`
- Modify: `clients/web/src/lib/api/facade/podConnect.ts`
- Modify: `clients/web/src/components/pod/CreatePodForm/WorkerRuntimeStep.tsx`
- Add or modify a focused resource editor component under `clients/web/src/components/pod/CreatePodForm/`
- Test: backend worker creation validation/options tests
- Test: frontend resource editor tests

- [ ] Add a failing validation test for CPU, memory, and storage values with positive units and reject zero, negative, or oversized values.
- [ ] Add a failing options test that exposes human-readable resource values for each preset.
- [ ] Implement the minimal typed resource request/limit fields needed by the existing pod creation compiler; retain preset profiles but allow a custom profile in the draft.
- [ ] Replace “compute target”, “deployment mode”, and “resource profile” as primary labels with Chinese user-facing concepts and short descriptions while retaining the backend enum values.
- [ ] Add explicit CPU, memory, and storage controls with stable unit labels and validation feedback.
- [ ] Run the backend and frontend focused tests.

### Task 3: Fix Worker type terminology and configuration schema

**Files:**
- Modify: `backend/internal/api/connect/pod/worker_creation_options.go`
- Modify: the worker type schema/domain files under `backend/internal/domain/workerspec/`
- Modify: `clients/web/src/components/pod/CreatePodForm/workerTypeConfigSchema.ts`
- Modify: `clients/web/src/components/pod/CreatePodForm/WorkerTypeConfigStep.tsx`
- Add: a focused visual variable field component under `clients/web/src/components/pod/CreatePodForm/`
- Test: worker type schema parsing tests
- Test: Worker type configuration component tests

- [ ] Add a failing schema test for field label, description, required/optional state, default, select options, and secret-reference fields.
- [ ] Implement backward-compatible decoding for existing basic schemas while making the new schema metadata explicit.
- [ ] Render the section as “Agent 配置变量”, with Chinese labels, descriptions, required markers, type-appropriate controls, and a clear empty state.
- [ ] Keep secret fields as references to credential records; never place plaintext API keys in the draft preview or AI fill result.
- [ ] Run schema and component tests.

### Task 4: Add credential customization and Worker config import

**Files:**
- Inspect existing credential/env-bundle APIs under `backend/internal/service/workercreation/` and `clients/web/src/lib/api/`
- Modify the existing credential creation API/client if it already owns the correct boundary
- Add a focused credential editor component under `clients/web/src/components/pod/CreatePodForm/`
- Add a config-file parsing/validation helper under `clients/web/src/components/pod/CreatePodForm/`
- Modify: `clients/web/src/components/pod/CreatePodForm/WorkerCredentialModelSection.tsx`
- Test: credential reference validation and config-file parsing tests

- [ ] Add a failing test proving an existing credential can be selected and a user-defined credential produces only a stored reference in the Worker spec.
- [ ] Add a failing file parsing test for supported Worker config formats, invalid content, and oversized files.
- [ ] Implement custom API key/provider/base URL/model fields with explicit sensitive-input handling and no plaintext persistence in session storage.
- [ ] Implement a file upload/import control that validates the file locally, sends it only to the approved local API, and stores a reference or validated values rather than raw secrets in the draft.
- [ ] Run focused tests and inspect network payloads to ensure secrets are not included in draft persistence.

### Task 5: Make Skill and knowledge base selection independent and persistent

**Files:**
- Modify: `clients/web/src/components/pod/hooks/useCreatePodFormEffects.ts`
- Modify: `clients/web/src/components/pod/hooks/useWorkerCreateDependencies.ts`
- Modify: `clients/web/src/components/pod/CreatePodForm/SkillMultiSelect.tsx`
- Modify: `clients/web/src/components/pod/CreatePodForm/KnowledgeBaseMountSelect.tsx`
- Modify: `clients/web/src/components/pod/CreatePodForm/WorkerWorkspaceCapabilities.tsx`
- Modify: `clients/web/src/components/pod/hooks/workerCreateDraft.ts`
- Modify: `clients/web/src/components/pod/hooks/workerCreateController.ts`
- Modify: `clients/web/src/components/pod/CreatePodForm/index.tsx`
- Test: draft reducer/persistence tests
- Test: Skill and knowledge base selector tests

- [ ] Add a failing test proving skills load and can be selected when no repository is selected.
- [ ] Add a failing reducer/persistence test proving the current step, selected worker type, resources, knowledge bases, skills, and non-sensitive fields survive route unmount/remount.
- [ ] Remove the repository gate from Skill selection; keep repository selection independent.
- [ ] Rename visible actions to “选择知识库” and “选择 Skill”; keep management links secondary.
- [ ] Persist a sanitized draft snapshot keyed by organization and creation-flow instance; restore it on mount and clear it after successful creation or explicit cancel.
- [ ] Preserve selected knowledge base slugs/modes in the local draft so the selector remains readable after returning from the knowledge base page.
- [ ] Run reducer and component tests.

### Task 6: Browser acceptance and runtime verification

**Files:**
- No production changes unless browser evidence reveals a defect.
- Evidence: screenshots and browser console/network logs from the in-app browser.

- [ ] Reload `/dev-org/workers/new` after the dev server rebuild and confirm no framework error overlay.
- [ ] Select MiniMax, OpenClaw, and DoAgent and confirm each has a selectable runtime image, a usable model/credential path, and clear Chinese labels.
- [ ] Create one instance for each runtime with a minimal safe task and verify the resulting Worker/pod reaches a ready or idle state.
- [ ] Navigate from the creation form to the knowledge base page and back; verify draft fields and current step remain.
- [ ] Select a Skill without selecting a repository and verify it remains selected through preflight.
- [ ] Verify console errors, failed network requests, disabled-state reasons, loading state, empty state, validation errors, and desktop/mobile layout screenshots.
- [ ] Run the final focused Go and frontend test commands and read their output before reporting completion.

## Acceptance Scenarios

- Given the local MiniMax, OpenClaw, and DoAgent runner services are online, when the user opens Worker creation, then those three agent modes are selectable and each has a real runtime image option.
- Given the user has no repository selected, when the user opens Skill selection, then available Skills can be selected without an error or repository warning.
- Given the user selects a knowledge base and enters non-sensitive Worker settings, when the user visits the knowledge base page and returns, then the selections, current step, and non-sensitive settings are restored.
- Given the user enters CPU, memory, and storage, when any value is invalid, then the form blocks preflight and explains the exact field error.
- Given the user configures an API key, when the draft is persisted or previewed, then the plaintext key is absent and only the credential reference is retained.
