# Resource-Native Contract

## Decision

Agent Cloud exposes a Kubernetes-inspired resource document for authoring,
inspection, API automation, and GitOps. It does not copy Kubernetes
reconciliation semantics or replace the existing domain services.

The first API group is `agentcloud.io/v1alpha1`. Splitting resources into more
groups before the contract stabilizes would add versioning and reference
complexity without improving the first delivery.

## Resource Envelope

```yaml
apiVersion: agentcloud.io/v1alpha1
kind: Expert
metadata:
  name: code-reviewer
  namespace: acme
  displayName: Code Reviewer
  labels:
    department: engineering
spec:
  workerTemplateRef:
    kind: WorkerTemplate
    name: codex-reviewer
    revision: 12
  promptRef:
    kind: Prompt
    name: code-review-system
    revision: 4
status:
  observedGeneration: 3
  activeRevision: 9
  conditions:
    - type: Ready
      status: "True"
```

`metadata.name` and `metadata.namespace` follow `slugkit`. `displayName` is
presentation metadata and may contain Unicode. `uid`, `resourceVersion`,
`generation`, and `status` are server-owned.

Submitted manifests containing server-owned fields fail validation. Exported
documents may include them as read-only evidence.

## Kind Taxonomy

| Kind | Ownership | Runtime meaning |
| --- | --- | --- |
| WorkerDefinition | Platform, read-only | Executable harness contract |
| WorkerTemplate | User, versioned | Reusable Worker authoring resource |
| Worker | User invocation | One Worker creation intent |
| Expert | User, versioned | Business capability whose revision pins a snapshot |
| Workflow | User, versioned | Reusable task whose revision pins a snapshot |
| GoalLoop | User invocation | One goal execution with fixed verifier and snapshot |
| WorkerRun | Server, read-only | Observed execution and manifest identity |
| WorkflowRun | Server, read-only | Observed Workflow occurrence |
| ModelBinding | User or organization | Model and provider connection selection |
| Prompt | User or organization | Versioned prompt content |
| Skill | User or organization | Versioned capability artifact |
| KnowledgeBase | User or organization | Authorized knowledge source |
| ToolBinding | User or organization | Tool-specific model and credential binding |
| EnvironmentBundle | User or organization | Versioned non-secret values and secret refs |
| Repository | Organization | Authorized source repository |
| ComputeTarget | Organization | Runner pool or cluster placement target |
| ResourceProfile | Organization | CPU, memory, and GPU request contract |

Applying a `WorkerTemplate` creates or selects an immutable
`WorkerSpecSnapshot`. The template is the user-facing reusable resource; the
snapshot remains the runtime SSOT.

## Reference Contract

Draft references contain:

```yaml
kind: ModelBinding
namespace: acme
name: coding-primary
revision: 7
```

`apiVersion` defaults only when the referenced Kind belongs to the same API
version. Cross-namespace references are denied unless a typed policy explicitly
allows them.

Plan resolves every reference to:

```text
apiVersion + kind + namespace + name + uid + revision + digest
```

Apply persists the resolved identity. Runtime never resolves an applied
reference by mutable name or active revision. Secret values are never part of a
resource document, reference, plan, diff, snapshot, or status.

Selectors may support listing and placement. They cannot select models,
credentials, Secrets, or authorization-sensitive dependencies.

## Schema Registry

The registry is keyed by `apiVersion + kind` and provides:

- strict Spec decoding with unknown-field rejection;
- deterministic normalization and validation;
- safe summary and semantic diff projection;
- typed conversion into an existing domain draft;
- reference extraction and sensitivity metadata.

The registry does not contain persistence or business actions. Domain services
remain responsible for ownership, revisions, transactions, and lifecycle.

## Control Plane API

The generic resource service supports:

```text
GetResource
ListResources
ExportResource
ValidateResource
PlanResource
WatchResources
```

There is no generic mutation that hides domain semantics. Plans are consumed by
typed operations:

```text
CreateWorkerFromPlan
ApplyWorkerTemplatePlan
CreateExpertFromPlan
CreateExpertRevisionFromPlan
CreateWorkflowFromPlan
CreateWorkflowRevisionFromPlan
CreateGoalLoopFromPlan
BindTicketWorkerFromPlan
```

Run, trigger, pause, resume, publish, archive, cancel, and terminate remain
typed commands. Applying a Worker or GoalLoop does not imply automatic
recreation after termination.

## Security And Failure Rules

- Tenant scope comes from authentication, never manifest metadata alone.
- Unknown API versions, Kinds, fields, and references are errors.
- Status and server metadata cannot be written by manifest Apply.
- YAML is limited to 256 KiB source/output, 64 KiB per physical line, 10,000
  nodes, 64 container levels, and one document. Aliases, anchors, merge keys,
  duplicate keys, custom tags, YAML-only scalar types, and ambiguous unquoted
  strings are rejected or quoted by the codec.
- Diffs show Secret reference identity only.
- A stale or changed security policy requires a new plan.
- No resource path falls back to legacy AgentFile or mutable domain fields.

## Acceptance

The resource-native surface is accepted when:

1. JSON and YAML round-trip to the same normalized typed draft;
2. every reference is pinned by Plan and preserved by Apply;
3. every runtime Pod still resolves from `WorkerSpecSnapshot`;
4. invalid, stale, unauthorized, or unknown resources fail explicitly;
5. the frontend contract in `11-resource-editor-frontend.md` passes.

User-facing behavior is documented in
`docs/product/resource-native-orchestration.md` and
`docs/product/resource-yaml-manual.md`.
