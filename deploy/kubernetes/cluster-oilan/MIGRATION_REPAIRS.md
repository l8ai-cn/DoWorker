# Oilan Migration Repairs

Normal Oilan deploys do not run DDL/DML. Schema changes and seed changes must be
applied first through an audited DoSql change, then `deploy.sh` receives:

```bash
DOSQL_RELEASE_DB_TARGET=<target> \
DOSQL_RELEASE_DB_MODE=production \
DOSQL_RELEASE_DB_SESSION=<dosql-session> \
DOSQL_RELEASE_CHANGE_ID=<change-id> \
DOSQL_RELEASE_OPERATION_ID=<dosql-operation-id> \
DOSQL_RELEASE_MIGRATION_VERSION=<latest-backend-migration> \
DOOPS_TARGET=gw-oilan-node \
./deploy.sh
```

The repository gate rejects caller-supplied evidence claims. It requires a
fixed production query binding to read a DoSql append-only journal and evidence
artifact, validate the verified lifecycle and artifact fingerprint, and prove
the target, mode, session, change, operation, and migration version. That
binding is not available yet, so Oilan deploys remain blocked.

Before a future release evidence can be created, complete a separate DoSql task
to register the canonical `gw-oilan-node` / `agentsmesh` PostgreSQL asset and
its read-only Postgres-over-DoOps adapter. That task must only audit current
state; mutation `000230 -> 000231` remains a user-confirmed DoSql change.

Production migration state `222 dirty=true` caused by the historical
`video-studio` insert must be repaired only after the corrected Backend image is
published and committed:

```bash
MIGRATION_REPAIR_ACK=repair-dirty-222-video-studio \
DOOPS_SESSION=<release-session> \
bash deploy/kubernetes/cluster-oilan/repair-migration-222.sh
```

The repair requires a clean, pushed `main` commit with successful GitHub checks,
verifies the exact dirty state and schema preconditions, stops application
writes, creates a checksummed backup, and reruns migration `000222` from version
`221`. Any failed precondition leaves the database untouched.
