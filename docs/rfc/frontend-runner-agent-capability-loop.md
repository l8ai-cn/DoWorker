# Frontend Runner-Agent Capability Loop

## Goal

Make the frontend enforce the same runner-agent capability contract as the
backend: an agent can only be selected, scheduled, or submitted when at least one
online runner reports that agent in `available_agents`; manually selected runners
must support the chosen agent.

## Scope

- Create Pod form
- Loop create/edit dialog
- Coordinator automation project create dialog

## Non-Goals

- No backend fallback.
- No synthetic mapping from one agent to another.
- No broad UI restyle.

## Success Conditions

- Unit tests fail before implementation and pass after implementation for:
  - incompatible runners are not offered for a selected agent;
  - Create Pod cannot submit with no compatible runner;
  - Loop cannot submit with no compatible runner;
  - Coordinator creation submits an explicit compatible `agent_slug`.
- Real browser verification exercises at least one create flow with rendered
  selects, disabled states, and no relevant console errors.

## Failure Conditions

- Same test or browser validation fails three consecutive times for the same
  reason.
- Required backend/API data is absent and cannot be inferred safely.
- The change would require weakening tests, hiding errors, or adding fallback.

## Budget

- Maximum 6 implementation/verification cycles.
- Stop before commit/push unless explicitly requested.

## Current Cycle

1. Completed failing tests for frontend capability matching.
2. Implemented shared compatibility helpers and wired them into the three flows.
3. Fixed the shared web `Select` control so disabled state and listbox options
   are reflected in browser-visible semantics.
4. Verified targeted unit tests, full web unit tests, type-check, lint, and diff check.
5. Completed real browser validation on the local dev environment.

## Verification Record

- `bazel test //clients/web:unit --test_output=errors --test_arg=src/components/pod/CreatePodForm/__tests__/CreatePodForm.test.tsx --test_arg=src/components/pod/CreatePodForm/__tests__/CreatePodForm-submit.test.tsx --test_arg=src/components/loops/__tests__/LoopCreateDialog.test.tsx --test_arg=src/components/coordinator/__tests__/CreateProjectDialog.test.tsx`
- `bazel test //clients/web:unit --test_output=errors`
- `bazel build //clients/web:src`
- `bazel test //clients/web:lint --test_output=errors`
- `git diff --check`
- Real Chrome validation against `http://localhost:10007`:
  - Coordinator create dialog renders agent selection and blocks submit until an agent is selected.
  - Create Pod dialog renders real agent options and blocks create until an agent is selected.
  - Loop create dialog renders real agent options and blocks create until an agent is selected.
  - Screenshots:
    - local screenshots under `output/` (gitignored scratch)
