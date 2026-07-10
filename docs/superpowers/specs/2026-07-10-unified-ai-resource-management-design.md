# Unified AI Resource Management Design

## Objective

Replace the split EnvBundle credential, `ai_models`, and implicit primary-credential paths with one explicit resource system. Users manage provider connections once, attach compatible models to them, and select a concrete model resource when creating a Worker. The first release supports chat, image, audio, video, embedding, and multimodal resource metadata; quota and usage surfaces are truthful extension points until provider accounting is implemented.

The approved UI direction is **A: Unified Resource Center**.

## Product Contract

- A provider connection represents one account or Token plus its API endpoint.
- A model resource represents one usable model or capability exposed by a provider connection.
- One connection may own many model resources; credentials are never copied per model.
- Personal connections are managed by their owner. Organization connections are managed by organization owners/admins and are visible as safe metadata to members.
- Worker creation selects one explicit, visible, enabled, compatible model resource.
- Defaults may initialize the UI selection, but the client always submits the selected resource ID. The backend never silently chooses or appends a different credential.
- Missing, inaccessible, incompatible, or unhealthy resources fail Worker validation. They never degrade to image-local authentication or an empty environment.
- EnvBundle remains available for non-credential runtime/config environment bundles. `kind=credential` leaves the active Worker path after migration.

## Domain Model

### ProviderConnection

`provider_connections` owns encrypted credentials and connection lifecycle:

```text
id, owner_scope, owner_id, provider_key, name, base_url
credentials_encrypted, configured_fields
status, is_enabled, last_validated_at, validation_error
revision, created_by, created_at, updated_at
```

`owner_scope` is `user` or `org`. `status` is `unchecked`, `valid`, or `invalid`. Only configured field names leave the backend. Token values are write-only.

### ModelResource

`model_resources` owns selectable capabilities:

```text
id, provider_connection_id, model_id, display_name
modalities[], capabilities[], is_enabled
revision, created_at, updated_at
```

Supported modalities are `chat`, `image`, `audio`, `video`, `embedding`, and `multimodal`. Capabilities are registry-defined identifiers such as `text-generation`, `vision-input`, `image-generation`, `speech-to-text`, `text-to-speech`, and `video-generation`.

Defaults are normalized in `model_resource_defaults`:

```text
owner_scope, owner_id, modality, model_resource_id, created_at, updated_at
```

There is at most one default per `(owner_scope, owner_id, modality)`. A model may be the default for several modalities without becoming the default for every capability it exposes. Effective reads prefer a personal default over an organization default for the same modality. Defaults only initialize the client selection; the client still submits the exact selected resource ID.

### UsageSummary

List responses include an optional usage summary:

```text
quota_total?, usage_total?, remaining?, unit?, period?, measured_at?
```

Until accounting exists, this object is absent and the UI shows `未接入`, never `0` or invented usage.

## Provider Registry

A code-owned registry defines provider key, display name, modalities, credential field schema, default endpoint, protocol adapter, model discovery support, and connectivity check policy. Database rows reference registry keys instead of duplicating provider definitions.

The initial catalog covers:

- Chat/multimodal: OpenAI, Anthropic, Google Gemini/Vertex AI, Azure OpenAI, OpenRouter, DeepSeek, Alibaba DashScope/Qwen, Volcengine/Doubao, MiniMax, Zhipu, Moonshot, xAI, and Mistral.
- Image: OpenAI, Stability AI, Black Forest Labs, Replicate, fal.ai, and Ideogram.
- Audio: OpenAI, ElevenLabs, MiniMax, Google, and Azure Speech.
- Video: OpenAI/Sora, Google/Veo, Runway, Kling, MiniMax/Hailuo, Luma, Replicate, and fal.ai.
- Custom OpenAI-compatible and Anthropic-compatible connections.

The registry makes support additive. Provider-specific execution adapters may be added without changing persistence or UI contracts.

## API Boundaries

Add typed Connect services for:

- list provider catalog;
- list effective connections/resources for the current user and organization;
- create/update/disable/delete/validate provider connections;
- create/update/disable/delete model resources and set a default for one modality;
- get safe resource metadata and optional usage summary.

Mutation handlers enforce owner scope. Effective reads merge user and organization resources once and return explicit scope/permission metadata. Custom endpoint validation blocks loopback, link-local, private network, and metadata-service targets unless a separately authorized deployment policy allows them.

`CreatePodRequest` replaces the old credential-profile field and ambiguous model/credential combination with `model_resource_id`. The backend validates visibility, enabled state, modality compatibility, and connection health before resolving credentials.

## Worker Runtime

The backend resolves exactly the submitted `model_resource_id`, decrypts its provider connection, and generates the harness-specific environment or config bundle. Model credentials do not travel through user-authored AgentFile text.

Remove implicit primary credential bundle mounting. Remove credential-kind EnvBundle selection from Worker creation. Runtime EnvBundles remain ordered overlays, but cannot provide model authentication fields after the migration gate.

Worker compatibility is capability-based:

- Codex/Claude/Gemini/do-agent require a resource with `chat` and a supported protocol adapter.
- Image/audio/video resources remain available to tools and future creation workflows but are excluded from incompatible Worker selectors.
- No compatible resources produces a blocking empty state with a link to AI Resources.

## Unified Resource Center UI

Add `AI Resources` to personal and organization settings. The page contains:

- capability filters: all, chat, image, audio, video, embedding;
- truthful summary cards: resource count, enabled count, quota, usage;
- provider connection groups showing scope, status, configured fields, model count, and validation time;
- model resource rows showing model ID, modalities, default modalities, enabled state, and optional usage;
- a provider onboarding dialog for provider, scope, endpoint, credentials, validation, and model selection;
- explicit loading, empty, error, permission-denied, validating, invalid, disabled, and saved states.

The Worker form replaces `API credential` and virtual-key ambiguity with one `Model resource` selector grouped by personal and organization scope. A manage-resources link opens the new settings module. Quota attribution remains a secondary property of the selected model resource/virtual key, not another credential picker.

## Migration

Migration is explicit and fail-closed:

1. Create provider/resource tables and APIs without switching Worker reads.
2. Run an application migrator with the production encryptor. It decrypts existing `ai_models` and credential EnvBundles, writes canonical connections/resources, and records source-to-target mappings.
3. Verify source/target counts, configured-field parity, owner scope, and decryptability. Any failed row blocks cutover and produces an operator-visible report.
4. Switch settings and Worker creation to model resources in one release gate. No runtime dual-read or fallback is allowed.
5. Remove implicit primary injection, old AgentCredential clients/proto, the credential-profile request field, credential EnvBundle UI, and obsolete tests/docs.
6. Drop old credential storage only after the verification report is clean and rollback snapshots exist.

Existing virtual API keys are remapped to the migrated model resource IDs before the old `ai_models` table is retired.

## Audit and Errors

Create, rotate, validate, enable/disable, default, and delete operations emit structured audit events with actor, owner scope, resource ID, provider key, result, and request correlation ID. Secret values never enter logs or audit payloads.

Connection and resource mutations use revision-based compare-and-swap so stale updates cannot restore old credentials or overwrite a newer validation result. Validation is two-phase: an audited transaction first marks the connection and its resources `unchecked`, then the provider probe runs with deterministic network timeouts, and a second audited compare-and-swap commits the result. A crash, concurrent rotation, or final audit failure therefore leaves the resource non-selectable.

List, validation, decrypt, and runtime resolution errors remain distinct typed errors. UI error states provide retry or management actions. Worker creation does not proceed after a model-resource error.

## Acceptance Scenarios

- Given a user and organization connection, effective listing returns both with correct scope and mutation permissions.
- Given one connection with chat/image/video models, the resource center shows all while a Codex Worker only offers compatible chat resources.
- Given a default resource, the form initializes it and submits its ID; the backend resolves that exact resource.
- Given a different explicit resource, no default or primary resource is appended or allowed to override it.
- Given no compatible resource, Worker creation is disabled and links to AI Resources.
- Given a list or decrypt failure, the page/Worker shows an error and does not present image authentication as a successful alternative.
- Given missing accounting, quota and usage display `未接入` rather than fabricated numbers.
- Given an organization member, organization connection secrets and mutations remain unavailable while safe resource metadata is selectable.
- Given a migration failure, cutover is blocked and old data remains untouched.

## Verification

- Backend repository/service/API tests cover ownership, effective visibility, compatibility, exact resolution, validation errors, audit emission, and migration parity.
- Web unit tests cover navigation, onboarding, scope permissions, all UI states, resource filtering, exact Worker submission, and absence of the default-auth option.
- Bazel build, lint, backend tests, web tests, and schema checks pass.
- Browser QA covers personal/org resource management and Worker success, empty, loading, invalid, permission, and error paths; console and network logs contain no relevant errors.

## Non-Goals

- Content publishing to social/video platforms.
- Provider billing aggregation, automatic recharge, or currency conversion in this release.
- Executing image/audio/video generation directly from the resource center.
- A runtime compatibility fallback to EnvBundle, legacy AgentCredential, or image-local authentication.
