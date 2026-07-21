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
  "operationId": "dbop-oilan-probe-001",
  "session": "oilan-read-20260720-001",
  "queryName": "migration-version"
}
```

The adapter invokes only `doops -session <session> exec --target
gw-oilan-node`, checks the fixed `agentsmesh/postgres` service and
`agentsmesh-secrets#DB_PASSWORD` binding remotely, and emits a redacted,
hash-verifiable evidence document. It fixes the canonical DoOps binary and
config paths, validates the gateway/cluster/instance targeting line, and forces
`default_transaction_read_only=on` plus a 15-second statement timeout. It never
records a URI, secret value, or raw result. Its fixed remote command remains
inside the DoOps execution path; there is no local Kubernetes or PostgreSQL
execution path. If the target is offline or the result does not prove the
registered asset, the command fails and produces no successful verification.

Corroborate the tracked registration against DoOps Gateway events and target
session audit logs:

```bash
node .agents/skills/dosql/scripts/oilan-postgres-registration-verify.mjs
```

The July 21, 2026 registration observed PostgreSQL 16.14. A post-change
read-only check observed migration state `231`, `dirty=false`. Corroboration
fails closed unless the Gateway returns one
unique matching `exec` event with the exact fixed command summary and result.
It also verifies the SHA-256 of each target-managed `.doops-audit-log` against
the complete fixed command and exit record, covering the command prefix that
the Gateway summary truncates. The target log is mutable by the same execution
authority, so the command reports `releaseAuthority=false`; neither it nor local
read-only `.dosql` evidence is accepted by the release gate.
The verifier requires the canonical DoOps config and a Gateway login with
audit-read permission; missing remote audit access fails closed. Release-grade
verification requires a future immutable full-command digest or signed receipt
recorded by Gateway at execution time.

This adapter is read-only. It must not execute production migrations, including
`000225 -> 000231`; that requires a separate user-confirmed DoSql change.

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
