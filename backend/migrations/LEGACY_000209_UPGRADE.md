# Legacy 000209 Upgrade

The historical production line used these versions:

- `000207`: capability heading normalization
- `000208`: GoalLoop iteration state
- `000209`: GoalLoop retry prompt state

The current line reuses `000207-000209` for adapter and WorkerSpec contracts,
so a database that is still on the historical `000209` line cannot be upgraded
with ordinary `migrate up`. The immutable migration chain will fail closed
before any forward-only repair migration can run. Do not edit the released
`000210` bridge to hide this; databases that have already executed it would not
receive the edit.

Normal `migrate up` is supported for:

- a fresh database
- clean current-line databases at `000222` or later
- production deployments that have already reached `000229`

`000231_repair_post_229_resource_lineage` is a forward-only repair for the
`000229+` deployment baseline. Its down migration intentionally retains the
repair because removing it would reintroduce dispatch-breaking lineage gaps.

`000230_coordinator_worker_spec_snapshot` adds a nullable snapshot binding for
coordinator projects. Existing rows remain `NULL` so ordinary migration remains
clean and repeatable; production dispatch fails closed until an audited
operator/API path creates immutable worker spec snapshots and binds each project
in the owning row transaction. The migration must not infer a snapshot from
mutable `agent_slug` or model defaults.

Historical `000209` recovery is manual: stop writers, restore or snapshot the
database, prove the exact historical schema fingerprint, and run an approved
pre-migration bridge in a controlled recovery window before resuming the normal
line. If that proof is unavailable, restore from the pre-upgrade backup rather
than forcing the migration version.

A dirty `000222` is intentionally rejected by golang-migrate before any later
migration runs. Do not use `force` as a repair. Stop writers, retain the
pre-migration backup, inspect the failed `000222` transaction and schema, then
either restore the backup or complete an approved repair that proves the schema
matches a clean version. Only after that evidence exists may the migration
version be changed under the database recovery runbook.
