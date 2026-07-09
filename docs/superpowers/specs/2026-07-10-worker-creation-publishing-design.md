# Worker Creation and Publishing Redesign

## Status

Approved for implementation on 2026-07-10. The implementation uses these product meanings:

- Codex, DoAgent, Claude Code, and similar executors are Worker types.
- A runtime image is an immutable OCI image reference, not a Worker type label.
- A model is a visible LLM configuration or virtual-key binding.
- Publishing an Expert saves a reusable template and does not launch a Worker.

## Goal

Give users one truthful workflow for configuring and creating a Worker, while making the same immutable configuration publishable as an Expert. A running Worker must also expose explicit actions for publishing authored Skills and publishing its configuration as an Expert.

## Current Problems

The current page exposes natural-language creation and a full wizard as separate primary creation paths. The wizard calls an Agent definition an image, calls an individual Runner a deployment target, allows submission from step one, and mixes credentials, models, repositories, Skills, Experts, knowledge bases, budgets, lifecycle, and AgentFile source across unrelated steps.

The current create contract cannot express runtime image, compute target, deployment mode, or resource profile. Lifecycle controls contain values that are not sent to the backend. Data loaders frequently convert failures into empty collections. Publishing from a Pod sends only Expert name and slug, so runtime configuration is lost.

## Product Model

| Object | Responsibility |
| --- | --- |
| Model binding | Visible provider, model, credential reference, and enforceable usage policy |
| Worker type | Executable harness definition such as Codex or DoAgent |
| Runtime image | Immutable image identifier and digest compatible with a Worker type |
| Compute target | Cluster or managed execution environment available to the organization |
| Deployment mode | Pooled Runner or dedicated Worker Runner Pod |
| Resource profile | Server-defined CPU, memory, and optional GPU requests and limits |
| WorkerSpec | Versioned, immutable description of one Worker configuration |
| Worker | Running or retained instance created from a WorkerSpec snapshot |
| Expert | Reusable template backed by a WorkerSpec snapshot |
| Skill | Versioned authored capability package that can be mounted or published |

The existing internal `Pod` name may remain during migration. User-facing copy uses `Worker` consistently.

## WorkerSpec V1

```text
runtime
  model_binding (ai_model_id | virtual_api_key_id | agent_managed)
  worker_type_slug
  runtime_image_id

placement
  policy (explicit | automatic)
  compute_target_id
  deployment_mode
  resource_profile_id

type_config
  schema_version
  values
  secret_refs
  interaction_mode
  automation_level

workspace
  repository_id
  branch
  skill_ids
  knowledge_mounts
  env_bundle_ids
  instructions
  initial_task

lifecycle
  termination_policy
  idle_timeout_minutes

metadata
  alias
  source_expert_id
```

Pod creation persists an immutable WorkerSpec snapshot. Expert creation persists the same structure. `source_expert_id` is audit metadata; importing an Expert copies its values into the draft and does not create live coupling.

The requested spec contains controlled identifiers. The persisted snapshot additionally resolves immutable provider/model, Worker-type definition hash, image digest, compute-target kind, and resource requests/limits. Physical `runner_id` is a scheduling result on the Pod, not part of the requested WorkerSpec.

AgentFile remains the runtime delivery format. The backend compiles it once from WorkerSpec, user preferences, and system policy. It is not an independent source of business truth.

Secret values never enter WorkerSpec. Only scoped references are persisted or published.

## Create Workflow

The canonical route is `/{org}/workers/new`. Ticket, repository, knowledge-base, command-palette, and workspace entry points navigate to this route with typed prefill context. They do not host another complete form state machine.

### Step 1: Runtime and Compute

The order is model, Worker type, runtime image, compute target, deployment mode, and resource profile.

Every option comes from a scoped compatibility response. Unsupported options remain visible but disabled with a concrete reason. An explicit `automatic` placement policy may be selected; it is not a fallback after invalid input.

Resource profiles are server-owned. The first version does not accept arbitrary CPU, memory, or GPU strings.

### Step 2: Worker-Type Configuration

The selected Worker type exposes a versioned schema with groups, field types, required state, validation, conditional visibility, option sources, and secret-reference semantics. Switching type clears only incompatible values after confirmation.

Raw AgentFile editing moves into an advanced source panel. It cannot silently diverge from structured configuration.

### Step 3: Workspace and Capabilities

The user selects repository and branch, compatible Skills, knowledge mounts, environment bundles, initial task, durable instructions, automation level, and lifecycle.

Loading an Expert is an explicit import action. Clearing the Expert resets imported values rather than leaving a partially applied template.

### Step 4: Preflight and Destination

The server validates authorization, compatibility, target capacity, required configuration, secret references, lifecycle, and enforceable quotas. Blocking errors and advisory warnings are rendered separately.

The actions are `Create Worker` and `Publish as Expert`. A future combined action must be named `Publish and Run`; neither current action has hidden side effects.

Natural-language creation becomes `Fill with AI`. It updates the same draft and never submits through a separate endpoint or remounts the form.

## Runtime Publishing

The active Worker header exposes a `Publish` menu with `Publish Skill` and `Publish as Expert`. The sidebar context menu may mirror these actions but is not the only discoverable entry point.

Expert publishing reads the server-owned WorkerSpec snapshot and captures model binding, Worker type, image, placement semantics, resource profile, type configuration, repository, Skills, knowledge, environment references, instructions, automation, and lifecycle. Invalid or missing references block publication.

Skill publishing follows two operations:

1. Discover valid Skill candidates below allowlisted sandbox roots.
2. Publish only user-selected candidates after preview and validation.

Candidate discovery returns relative path, slug, name, content summary, hash, validation state, and whether the candidate is new, modified, or already published. The server rejects absolute paths, traversal, escaping symlinks, oversized packages, invalid identifiers, invalid frontmatter, and unauthorized Worker access.

## Authorization and Failure Semantics

Every external Runner, repository, model, virtual key, image, target, profile, Skill, knowledge base, and Expert identifier is resolved through organization- or user-scoped policy before decryption, persistence, or dispatch.

Explicit Runner selection uses the same eligibility validator as automatic placement. Custom Worker types are either fully resolvable and runnable or absent from the runnable catalog.

Unknown automation values, incompatible model bindings, invalid schema values, Git backing errors, and malformed WorkerSpec fail explicitly. Load failures are never rendered as valid empty states.

## State Ownership

The create draft uses a reducer with explicit section validity and async states. It is not stored as business truth in Zustand. Created Workers and Experts continue through Rust Core canonical state.

Submission is disabled while required dependencies are loading, failed, or stale. A submitted request is either locked until completion or explicitly cancelled through an abort-capable API.

## Existing Architecture Alignment

Dedicated deployment builds on `2026-07-08-worker-per-runner-pod-design.md`: a dedicated Kubernetes Runner Pod has `MAX_CONCURRENT_PODS=1`, while pooled deployment remains a distinct user-selected mode. The deployment choice must be represented in WorkerSpec and validated against compute-target capabilities.

The uncommitted `feat/worker-config-lifecycle` worktree overlaps lifecycle and Runner ACP behavior. This branch must not copy those dirty files. Integration happens only after that work is committed and its contract can be reviewed.

## Non-Goals

- Replacing Runner, relay, PTY, or ACP protocols wholesale.
- Allowing arbitrary untrusted image strings or raw Kubernetes manifests.
- Adding dynamic per-Worker ingress in the first release.
- Making Expert publication implicitly launch a Worker.
- Preserving misleading controls through compatibility fallbacks.

## Acceptance Conditions

- The page has one creation state machine and one primary create action.
- Incompatible runtime combinations cannot advance and always explain why.
- Required async data must resolve successfully before submission.
- WorkerSpec round-trips through Proto, backend persistence, Rust Core, and web state.
- A published Expert reproduces the source WorkerSpec except for identity metadata.
- Only selected, valid sandbox Skills are published.
- Cross-organization identifiers are rejected before secrets or commands are resolved.
- Lifecycle UI, request, snapshot, and runtime behavior agree.
- Unit, contract, migration, frontend, and browser E2E checks pass without weakening validators.
