# Resource Editor Frontend

## User And Job

The primary user is a developer or operator configuring a Worker, Expert,
Workflow, or GoalLoop. The primary job is to create or revise one resource,
understand the exact runtime impact, and apply it without losing draft state.

The experience is a domain tool, not a generic Kubernetes console.

## YAML And Form State

The default experience is a domain form. YAML is an advanced source view:

- users with read access may view redacted YAML;
- users with edit permission may switch to source mode;
- form and YAML operate on one typed `ResourceDraft`;
- raw YAML text exists separately only while parsing fails;
- switching back to the form is blocked while YAML is invalid;
- no previous valid draft is silently applied after a parse failure;
- semantic validation and Plan remain server-authoritative.
- the editor displays the 256 KiB document and 64 KiB line limits before the
  user reaches them;
- strings that resemble numbers, booleans, null, dates, or YAML control
  syntax are emitted with an unambiguous quoted representation.

Persisted resource data and API caches stay in Rust Core. An unsaved editor
draft uses a local typed reducer because it is transient UI state. No Zustand
business SSOT is introduced.

The YAML editor is lazy-loaded so normal form users do not pay its bundle cost.
Server canonical formatting is authoritative; the first release does not
promise preservation of YAML comments.

The help entry links to `docs/product/resource-yaml-manual.md`; error messages
must not echo unknown or duplicate input keys because a user may have pasted a
Secret into the wrong position.

## Resource Editor Layout

```text
Kind icon  Display name  Ready  Revision

Overview | Configuration | YAML | Plan/Diff | Revisions | Runs | Events
```

The primary create and revise flow is:

1. choose Kind or a product template;
2. edit through a domain form or YAML;
3. validate locally for syntax and required fields;
4. request a server plan;
5. review blocking issues, warnings, and semantic diff;
6. apply through the typed target service;
7. display revision, status conditions, events, and resulting runs.

Reference fields use a picker showing Kind, name, status, scope, and resolved
revision. A draft may select an active revision, but Plan must show and pin the
exact resolved revision before Apply.

## Component Boundaries

The frontend adds focused resource-editor components:

```text
resource-editor/
  resource-draft-reducer.ts
  resource-editor-controller.ts
  ResourceEditorShell.tsx
  ResourceConfigurationPanel.tsx
  ResourceYamlPanel.tsx
  ResourceReferencePicker.tsx
  ResourcePlanReview.tsx
  ResourceSemanticDiff.tsx
  ResourceRevisionHistory.tsx
  ResourceStatusConditions.tsx
```

Existing Worker selectors for model, image, repository, Skill, knowledge, and
environment resources are reused through typed props. Workflow and Expert no
longer own parallel Worker configuration state after their cutover.

Domain-only fields remain outside the shared Worker editor:

- Expert product metadata and release notes;
- Workflow prompt, trigger, concurrency, timeout, callback, and retention;
- GoalLoop objective, verifier, budgets, and stopping conditions;
- Worker invocation prompt, alias, terminal size, and idempotency key.

## State Contract

The controller tracks:

```text
draft
source buffer and parse error
dirty revision
validation result
plan request and plan identity
apply request and result
active resource revision
permission and reference availability
```

Changing the draft invalidates the current plan. Concurrent async results carry
request IDs and cannot overwrite a newer draft. Apply is disabled until the
current draft hash matches the plan hash.

## UI States

Every resource editor covers:

- loading, empty, unavailable, and permission-denied references;
- invalid YAML and field-addressed validation errors;
- dirty, planning, plan-ready, stale-plan, applying, applied, and failed states;
- active revision conflict and recovery by re-plan;
- read-only status, migration-required, and revoked dependency states;
- desktop and mobile layouts without horizontal form overflow.

Errors preserve the draft and identify the failed field, reference, or action.
Destructive actions remain separate from Apply and require confirmation.

## Responsive Behavior

- Desktop uses a stable content area with a bounded review side panel.
- Mobile uses one vertical flow; Plan review becomes a full-height sheet.
- YAML and diff views may scroll horizontally inside their own region only.
- Selectors and action controls keep at least 44px touch targets.
- Long translated labels wrap without resizing fixed tool controls.

## Verification

Vitest covers reducer transitions, request races, form/YAML round trips,
plan invalidation, status protection, secret redaction, and permission states.

Playwright covers:

- create Worker from the form and review Plan;
- edit YAML and observe the same form values;
- invalid YAML blocking form switch and Apply;
- stale plan recovery;
- Expert and Workflow revision creation;
- migration-required remediation;
- cross-organization and read-only rejection;
- desktop and mobile screenshots, console, and network errors.

The frontend is accepted only after browser evidence confirms the primary path
and blocking states without clipping, overlap, or horizontal page overflow.
