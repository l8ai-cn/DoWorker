# Seedance Expert Platform Design

## Goal

Create a dedicated Seedance Worker and Expert that can turn a conversation into
a verified Volcengine Seedance video task, persist the resulting MP4, and render
it inside the platform.

## User

The user is a creator who describes an idea in natural language, reviews a
production brief and prompt, approves generation, and watches the result without
leaving the Worker conversation.

## Acceptance Scenarios

1. Given a valid chat model and a valid Seedance video resource, when the user
   creates a Seedance Worker, then both exact resource snapshots are recorded.
2. Given the Seedance skill is installed, when the user requests a clip, then the
   Worker produces a concise brief before spending video-generation quota.
3. Given the user approves generation, when the Worker invokes the Seedance
   tool, then task creation, polling, failure reporting, MP4 download, and run
   metadata persistence are deterministic.
4. Given generation succeeds, when the output-file event reaches web-user, then
   an inline video player and download action are visible.
5. Given a verified Seedance Worker, when it is published as an Expert and used
   to create another Worker, then the skill and both model bindings survive.
6. Given an invalid or unauthorized video resource, Worker preflight or
   generation fails explicitly; no chat model, environment fallback, or provider
   substitution is attempted.

## Skill Repository

Fork `Emily2040/seedance-2.0` into `l8ai-cn/seedance-expert-skill` to preserve
history and MIT attribution. Adapt its root skill to native Codex progressive
disclosure:

- `SKILL.md`: intake, approval gate, mode selection, execution contract.
- `references/`: selected upstream prompt, camera, continuity, safety, API, and
  retake guidance.
- `scripts/seedance_generate.py`: create, poll, download, and write run metadata.
- `scripts/seedance_prompt_lint.py`: reject missing action, conflicting camera
  direction, unresolved reference tags, and overloaded timelines.
- `agents/openai.yaml`: user-facing metadata for the Seedance Expert skill.

The platform imports the repository root as one skill. References use relative
Markdown paths instead of the upstream repository's custom `[skill:]` and
`[ref:]` routing tokens.

## Worker Contract

Add a `seedance-expert` Worker Definition backed by the existing `do-agent`
runtime and adapter. Its primary model requirement remains
`chat + text-generation + openai-compatible`.

Add definition-driven tool model requirements. The Seedance Worker declares one
required role:

```json
{
  "id": "seedance-video",
  "provider_keys": ["doubao"],
  "modality": "video",
  "capability": "video-generation",
  "environment": {
    "api_key": "SEEDANCE_API_KEY",
    "base_url": "SEEDANCE_BASE_URL",
    "model_id": "SEEDANCE_MODEL"
  }
}
```

WorkerSpec stores each resolved tool model as a role plus the same immutable
resource, connection, provider, adapter, and model snapshot used by the primary
binding. Creation and resume re-resolve every exact resource and reject revision
or identity drift.

## Provider Validation

The existing Doubao connection probe is unsupported. Add a Doubao bearer probe
against the provider's model-list endpoint only if the official endpoint is
verified. Seedance entitlement validation uses a non-billing provider request
when available; otherwise the resource remains unchecked and cannot be selected.

A real generation smoke test is separate because it spends quota. It uses a
rotated key supplied through the platform credential store, never source files,
shell history, fixtures, logs, or committed configuration.

## Generation Flow

The Worker first compiles and lints a prompt. Generation requires an explicit
user approval. The script posts an asynchronous task, records the task ID,
polls with bounded intervals, downloads the MP4 before the signed URL expires,
and writes JSON metadata containing provider, model, prompt, parameters, task
status, output path, and timestamps.

The script exits non-zero on authentication, entitlement, moderation, provider,
timeout, download, or schema errors. It never changes model IDs or providers.

## Artifact Rendering

Render output-file blocks in web-user instead of dropping them. Resolve the file
through the existing session file-content endpoint and classify it by filename.
MP4/WebM/MOV use a native video element; other files retain a download action.
Loading, unavailable, unsafe URL, and playback-error states remain visible.

## Security

- Rotate the API key exposed in conversation before any network test.
- Keep chat and video resources distinct even when they share one provider key.
- Never expose decrypted credentials in WorkerSpec, events, logs, metadata, or UI.
- Require explicit authorization for real-person likeness, voice, and protected
  IP workflows.

## Verification

- Skill schema validation, prompt-lint fixtures, and mocked generation lifecycle.
- Go tests for definitions, draft resolution, WorkerSpec validation, resume drift,
  and environment injection.
- TypeScript tests for Worker form selection and output-file rendering.
- Browser E2E: configure both resources, create Worker, approve a test clip,
  observe status, play the MP4, publish Expert, and create a second Worker.
- Git verification for both repositories after push.
