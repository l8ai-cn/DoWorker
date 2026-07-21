# Oilan Migration Repairs

Normal Oilan deploys do not run DDL/DML. Schema changes and seed changes must be
applied first through an audited DoSql change, then `deploy.sh` receives:

```bash
DOSQL_RELEASE_DB_TARGET=db_agentcloud_prod_postgres \
DOSQL_RELEASE_DB_MODE=production \
DOSQL_RELEASE_DB_SESSION=<dosql-session> \
DOSQL_RELEASE_CHANGE_ID=<change-id> \
DOSQL_RELEASE_OPERATION_ID=<dosql-operation-id> \
DOSQL_RELEASE_MIGRATION_VERSION=<latest-backend-migration> \
DOOPS_TARGET=gw-oilan-node \
./deploy.sh
```

The repository gate rejects caller-supplied evidence claims. It accepts only a
change-specific hash-chained journal and an immutable evidence artifact under
`.dosql`, with matching target, environment, session, operation, release
namespace, migration version, and artifact fingerprint.

The canonical asset was verified through `gw-oilan-node` on July 21, 2026:

- PostgreSQL `server_version_num=160014` (16.14);
- database `agentcloud`;
- `schema_migrations` exists;
- migration version `231`, `dirty=false` after the approved direct update.

Corroborate the registration against Gateway events and target session audit
logs without querying PostgreSQL:

```bash
node .agents/skills/dosql/scripts/oilan-postgres-registration-verify.mjs
```

This requires the canonical `~/.agent/skills/doops/config.json` target and a
Gateway login with audit-read permission. Gateway and target audit records are
corroborating evidence; they do not replace the release journal and evidence
artifact consumed by `dosql_release_gate.sh`.

The approved direct production update from `000225` through `000231` was
executed on July 21, 2026. Its release evidence is:

- change `chg-oilan-direct-schema-226-231-20260721`;
- operation `dbop-oilan-direct-schema-226-231-20260721`;
- session `oilan-db-direct-226-231-20260721`;
- resulting migration `231`, `dirty=false`.

The historical `222 dirty=true` state caused by the `video-studio` insert is not
the current production state. Do not run its repair while the verified state is
`231 dirty=false`. If a future read-only check again proves that exact dirty
state, repair it only after the corrected Backend image is published and
committed:

```bash
MIGRATION_REPAIR_ACK=repair-dirty-222-video-studio \
DOOPS_SESSION=<release-session> \
bash deploy/kubernetes/cluster-oilan/repair-migration-222.sh
```

The repair requires a clean, pushed `main` commit with successful GitHub checks,
verifies the exact dirty state and schema preconditions, stops application
writes, creates a checksummed backup, and reruns migration `000222` from version
`221`. Any failed precondition leaves the database untouched.
