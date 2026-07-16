# Legacy 000209 Upgrade

The historical production line used these versions:

- `000207`: capability heading normalization
- `000208`: GoalLoop iteration state
- `000209`: GoalLoop retry prompt state

The current line reuses `000207-000209` for adapter and WorkerSpec contracts.
A clean historical `000209` database therefore enters the compatibility bridge
inside `000210`. The bridge accepts only the historical schema fingerprint,
adds the skipped adapter contracts, and marks the column lineage so
`000213/000214` preserve the GoalLoop columns already owned by the old line.

Normal `migrate up` is supported for:

- a fresh database
- a clean historical `000209`
- a clean current `000222`

A dirty `000222` is intentionally rejected by golang-migrate before any later
migration runs. Do not use `force` as a repair. Stop writers, retain the
pre-migration backup, inspect the failed `000222` transaction and schema, then
either restore the backup or complete an approved repair that proves the schema
matches a clean version. Only after that evidence exists may the migration
version be changed under the database recovery runbook.
