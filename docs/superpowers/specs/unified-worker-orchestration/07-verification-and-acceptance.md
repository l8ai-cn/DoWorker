# Verification and Acceptance

## Contract Scenarios

### Worker Plan

Given a valid draft and current options revision, when the user plans, then the
response contains one canonical proposed spec, dependency manifest, compiled
layer hash, plan hash, and no secret values.

Given a stale options or policy revision, when apply is attempted, then apply
fails and creates no snapshot, revision, Pod, or outbox event.

### Snapshot Replay

Given a V2 or exact-classified V1 snapshot, when the current WorkerDefinition is
upgraded, then the snapshot still resolves its retained definition revision.

Given that exact revision is revoked, when execution is requested, then it
fails with `definition-revoked` and does not use the current definition.

Given a legacy V1 snapshot without immutable dependency evidence, when audit
runs, then it is not labeled exact and cannot receive a fabricated manifest.

### Expert

Given an Expert metadata edit, when it is saved, then the active runtime
revision does not change.

Given an Expert runtime edit, when its plan is applied, then a new immutable
revision is created and the active pointer changes atomically.

Given a legacy Expert without a snapshot, when it is run, then it returns
`migration-required` and does not reconstruct AgentFile.

### Workflow

Given a WorkflowRun created from revision 3, when revision 4 is published
before dispatch, then the run still uses revision 3 and its pinned snapshot.

Given the same Cron occurrence or event ID is delivered twice, then exactly one
WorkflowRun exists and cursor advancement is not lost.

Given malformed config or an unavailable bundle during migration, then the
Workflow becomes `migration_required`; no warning-only migration is allowed.

### Mesh and GoalLoop

Given a Ticket bound from an Expert, when the Expert later changes revision,
then the Ticket keeps its originally resolved snapshot.

Given an active GoalLoop, when a caller tries to change its snapshot or
verification command, then the update is rejected.

## Test Layers

### Domain Tests

- canonical WorkerSpec encoding and hash stability;
- immutable revision validation;
- plan status and expiration state machine;
- Expert and Workflow active-revision compare-and-swap;
- allowed invocation field validation;
- policy overlay monotonicity.

### Repository and Migration Tests

- PostgreSQL immutability triggers;
- organization-scoped composite foreign keys;
- content-addressed snapshot deduplication;
- plan idempotency and concurrent apply;
- launch intent lease recovery and request-hash conflicts;
- pending Runner command crash recovery;
- legacy classification and `migration_required`;
- up and down behavior within the supported rollback window.

### Service Contract Tests

- plan resolution across model, repository, Skill, knowledge, environment,
  image, target, and resource profile;
- definition revision retention and revocation;
- AgentFile deterministic compilation;
- run manifest materialization;
- command and manifest hash validation on Runner retries;
- Workflow trigger occurrence idempotency and persistent predecessor CAS;
- explicit resource, policy, capacity, and quota failures;
- absence of decrypted values in persisted JSON and logs.

### Cross-Stack Tests

- Proto round-trip through Go, Rust Core, WASM, and TypeScript;
- Worker wizard plan and apply;
- Expert create, revise, run, and publish from Pod;
- Workflow create, revise, manual trigger, Cron trigger, and persistent resume;
- GoalLoop start, verify, pause, resume, and completion;
- Ticket assignment and Pod launch;
- marketplace install success and unresolved dependency failure.
- session fork/switch/MCP rebuild, host binding, and Runner MCP plan enforcement.

### Browser E2E

Real browser tests cover:

- loading, empty, failed, stale, and disabled option states;
- plan diff review and stale-plan recovery;
- separate Expert metadata and runtime revision editing;
- Workflow revision visibility;
- migration-required remediation;
- authorization and cross-organization rejection;
- console and network errors;
- screenshots for primary success and blocking states.

## Determinism Gates

For the same snapshot and invocation, before live policy and placement:

- normalized spec bytes are identical;
- compiled WorkerSpec AgentFile bytes are identical;
- dependency manifest bytes are identical;
- hashes are identical.

For one materialization, the effective AgentFile and run manifest are persisted
before dispatch and match the command sent to Runner.

## Security Gates

- tenant scope is derived from authentication context;
- all referenced IDs are checked against organization ownership;
- plan apply verifies creator or explicit delegated permission;
- plan and idempotency keys are unguessable;
- secret values do not appear in API responses, events, audit logs, snapshots,
  manifests, pending command templates, Git projections, or test fixtures;
- system policy cannot increase permissions relative to WorkerSpec;
- revoked resources cannot be revived by snapshot replay.
- callback DNS, egress, redirect, signing, and log-redaction rules are enforced.

## Operational Signals

Metrics:

- plan success, blocking issue, expiration, and stale-apply counts;
- snapshot creation and deduplication;
- materialization failures by stable code;
- definition and resource revocation blocks;
- migration-required objects;
- projection lag and failure;
- WorkflowRun revision distribution.

Structured logs include organization ID, plan ID, snapshot ID, target revision,
manifest ID, Pod key, and stable error code. They exclude prompts when policy
marks them sensitive and always exclude secrets.

## Release Gates

The cutover cannot ship until:

1. focused Go, migration, Rust, TypeScript, and browser tests pass;
2. every fresh Pod entry point is in the execution inventory;
3. no fresh path accepts conflicting legacy runtime fields;
4. every runnable Expert has an active revision;
5. every enabled Workflow has an active revision;
6. unresolved legacy objects are explicitly disabled and reported;
7. definition revision retention and revocation are proven;
8. run manifests match Runner dispatch evidence;
9. rollback rehearsal succeeds within the supported window;
10. user-visible Worker, Expert, Workflow, Loop, and migration documentation is
    updated.

## Definition of Done

The design is implemented when a single WorkerSpec snapshot can be selected by
every product domain, every resulting Pod has an auditable run manifest, and
removing all Expert, Workflow, and Mesh legacy AgentFile builders does not
remove any supported runtime behavior.
