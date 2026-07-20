# Oilan Migration Repairs

Normal Oilan deploys do not run DDL/DML. Schema changes and seed changes must be
applied first through an audited DoSql change, then `deploy.sh` receives:

```bash
DOSQL_RELEASE_DB_TARGET=<target> \
DOSQL_RELEASE_DB_SESSION=<dosql-session> \
DOSQL_RELEASE_CHANGE_ID=<change-id> \
DOSQL_RELEASE_MIGRATION_VERSION=<latest-backend-migration> \
DOOPS_TARGET=gw-oilan-node \
./deploy.sh
```

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
