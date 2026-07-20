# PostgreSQL Rules

## Read Operations

Read-only commands include:

- `SELECT`;
- `EXPLAIN`;
- catalog inspection queries;
- `SHOW`.

## Oilan Production Read-Only Adapter

AgentsMesh production PostgreSQL is fixed to the registered asset
`db_agentsmesh_prod_postgres`. Use the dedicated entrypoint, not a local
connection URI, local `psql`, or a locally invoked Kubernetes database command:

```bash
node .agents/skills/dosql/scripts/oilan-postgres-doops-readonly.mjs probe \
  --input /path/to/request.json \
  --output /path/to/response.json
```

The request only accepts a unique DoOps `session` and an `operationId`.
`query` additionally accepts one allow-listed `queryName`:

```json
{
  "operationId": "dbop_oilan_probe_001",
  "session": "oilan-read-20260720-001",
  "queryName": "migration-version"
}
```

The adapter invokes only `doops -session <session> exec --target
gw-oilan-node`, checks the fixed `agentsmesh/postgres` service and
`agentsmesh-secrets#DB_PASSWORD` binding remotely, and emits a redacted,
hash-verifiable evidence document. It never records a URI, secret value, or
raw result. Its fixed remote command remains inside the DoOps execution path;
there is no local Kubernetes or PostgreSQL execution path. If the target is
offline, the command fails and produces no
successful verification.

This adapter is read-only. It must not execute production migrations, including
`000230 -> 000231`; that requires a separate user-confirmed DoSql change.

## Change Operations

PostgreSQL changes must become migration scripts:

- `CREATE`, `ALTER`, `DROP`, `COMMENT`, `TRUNCATE`;
- `INSERT`, `UPDATE`, `DELETE`, `MERGE`;
- role and grant changes.

## Validation

Before execution:

- inspect current schema with catalog queries;
- render migration SQL;
- prefer transaction-wrapped migration when safe;
- run dry-run or parse validation when available;
- define verification queries.

After execution:

- capture notices and affected rows;
- verify schema diff;
- attach output and artifact checksums.
