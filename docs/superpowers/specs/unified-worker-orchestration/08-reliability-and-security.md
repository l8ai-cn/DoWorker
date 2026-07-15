# Reliability and Security

## Worker Launch Intent

Applying a `create_worker` plan creates a durable `worker_launch_intent`.
Worker creation is not represented only by a consumed plan and an eventual Pod.

Required fields:

| Field | Meaning |
| --- | --- |
| `id`, `organization_id`, `requested_by_id` | Identity and scope |
| `plan_id`, `worker_spec_snapshot_id` | Applied configuration |
| `operation`, `target_key`, `idempotency_key`, `request_sha256` | Retry identity |
| `pod_key` | Preallocated Pod identity |
| `status` | Durable launch state |
| `manifest_id`, `command_sha256` | Effective execution identity |
| `attempt_count`, `lease_owner`, `lease_expires_at` | Recovery control |
| `last_error_code`, `last_error_summary` | Safe terminal evidence |
| timestamps | Audit and timeout handling |

The unique key is
`(organization_id, requested_by_id, operation, target_key, idempotency_key)`.
Reusing an idempotency key with another request hash is rejected.

State machine:

```text
pending -> materializing -> queued -> dispatching -> dispatched
                    |           |            |
                    +----------> failed <----+
pending | failed -> cancelled
```

Only transient failures explicitly classified as retryable may re-enter
`materializing`. Authorization, policy, definition, and request failures require
a new plan.

## Durable Runner Dispatch

Materialization writes the following in one transaction:

- WorkerRunManifest;
- queued Pod;
- exact pending Runner command;
- launch intent transition to `queued`;
- durable outbox event.

The pending command stores a canonical non-secret command template and secret
references, never decrypted values. Dispatch resolves secrets just in time and
verifies that recorded reference revisions remain current. A revision change
fails closed and requires rematerialization rather than mutating the manifest.

Online and offline Runners use the same pending command store. A successful
direct network send cannot be the only record that a command existed.

Dispatch workers claim commands with a lease and compare-and-swap status.
Crashes after commit are recovered by another worker after lease expiration.
Delivery is at least once; correctness comes from idempotency.

The command carries `pod_key`, `manifest_sha256`, and `command_sha256`. Runner
accepts a duplicate Pod key only when both hashes match the existing launch.
Different hashes for the same Pod key are a protocol conflict and fail closed.

## Workflow Trigger Identity

Every trigger creates a stable occurrence identity:

| Trigger | Stable key |
| --- | --- |
| Cron | normalized scheduled timestamp |
| API | caller idempotency key |
| Event | source event ID |
| Manual | generated request ID |

`workflow_runs` has a unique constraint on
`(workflow_id, trigger_type, trigger_key)`.

Cron executes one transaction that:

1. locks the Workflow scheduling cursor;
2. reads and pins the active revision;
3. validates the execution principal;
4. creates the WorkflowRun occurrence;
5. advances `next_run_at`;
6. writes a launch outbox event.

A crash before commit changes nothing. A crash after commit is recovered from
the outbox. Event and API retries return the existing WorkflowRun.

## Persistent Workflow Sessions

Workflow V1 does not implement session lanes. If
`session_persistence = true`, `max_concurrent_runs` must equal `1`.

Each persistent WorkflowRun stores `predecessor_run_id`. The predecessor is the
latest successfully completed run from the same Workflow revision and snapshot.
Completion updates the continuation pointer with run sequence compare-and-swap,
so a late older completion cannot replace a newer predecessor.

A revision or snapshot change starts a fresh session. Supporting concurrent
persistent lanes is a separate future design.

## Execution Principal

Workflow records distinguish:

- `initiated_by_id`: user or event that requested this occurrence;
- `execution_principal_id`: organization-scoped grant or service principal used
  for runtime access.

Scheduled and event Workflows require an explicit revocable execution
principal. `created_by_id` is audit metadata and is never an implicit fallback.
Manual execution also uses the configured principal unless the Workflow
explicitly declares caller-bound execution.

Every trigger and materialization revalidates the principal. Revocation blocks
new runs and resumes without changing historical revisions.

## Callback Security

Workflow callback configuration stores:

- HTTPS URL without userinfo;
- signing secret reference;
- allowed event types;
- timeout and bounded retry policy.

Delivery security requires:

- endpoint policy and egress allow rules;
- DNS resolution followed by destination IP validation on every connection;
- redirect validation using the same rules;
- unique delivery ID and idempotent retry;
- timestamped signature over delivery ID and body hash;
- sanitized logging of scheme, host, and error code only.

Query credentials, URL userinfo, resolved private addresses, and full callback
payloads are never logged.

## Recovery Jobs

Periodic deterministic jobs recover:

- expired launch leases;
- queued Pods missing a pending command;
- pending commands missing a Pod or manifest;
- WorkflowRun outbox events not yet dispatched;
- callbacks awaiting retry;
- Pods stuck before Runner acknowledgement.

Each recovery path has a maximum attempt count, terminal state, metric, and
operator-visible reason. Recovery cannot regenerate runtime configuration from
legacy fields or current defaults.
