# Install And Management

DoSql is meant to be installed and managed as a project Skill.

## Current Manual Install

Install DoOps first so the project has `doops` and a gateway target:

```bash
bash backend/scripts/install.sh \
  --project /path/to/business-project \
  --target-name <target-name> \
  --target-gateway <gateway-url> \
  --target-cluster <cluster> \
  --target-instance <instance> \
  --target-token '<gateway-user-token>' \
  --use '<database-capable target>'
```

Then install DoSql:

```bash
mkdir -p /path/to/business-project/.agent/skills/dosql
cp -R DoSuite/DoSql/Skill/. /path/to/business-project/.agent/skills/dosql/
```

Validate:

```bash
/Users/wwyz/.codex/skills/bin/quick-validate-skill \
  /path/to/business-project/.agent/skills/dosql
```

The first CLI wrapper is available inside the installed Skill:

```bash
node /path/to/business-project/.agent/skills/dosql/scripts/dosql-agent.mjs \
  classify \
  --input request.json \
  --output response.json
```

The primary onboarding flow should then discover candidates, ask the user to
confirm names, and register the confirmed assets:

```bash
node /path/to/business-project/.agent/skills/dosql/scripts/dosql-agent.mjs \
  discover-databases \
  --input discovery-request.json \
  --output discovery.json
node /path/to/business-project/.agent/skills/dosql/scripts/dosql-agent.mjs \
  register-database \
  --input database.json \
  --output registration.json
```

## Target One-Click Install

The intended management interface is:

```bash
doops skill install \
  --project /path/to/business-project \
  --product DoSql
```

The installer should copy the Skill, install or verify the `dosql-agent` CLI,
and check that the target gateway can reach the selected database-capable
doagent node.

## Target One-Click Database Registration

The intended database registration interface is:

```bash
node /path/to/business-project/.agent/skills/dosql/scripts/dosql-agent.mjs \
  discover-databases \
  --input discovery-request.json \
  --output discovery.json
node /path/to/business-project/.agent/skills/dosql/scripts/dosql-agent.mjs \
  register-database \
  --input database.json \
  --output registration.json
```

Registration should create:

- project and environment mapping;
- database assets;
- user-visible display names and aliases;
- connection or secret references;
- initial version baseline;
- structure snapshot;
- maintenance checklist;
- audit event.

Registration must be read-only against the target database. Mutating operations
belong to the change lifecycle and require user confirmation.

Before query, scan or change planning, the Agent should resolve the database
name from the user sentence:

```bash
node /path/to/business-project/.agent/skills/dosql/scripts/dosql-agent.mjs \
  resolve-database \
  --input resolve-request.json \
  --output resolved-database.json
```

If resolution returns `ambiguous`, the Agent must ask the user to choose the
environment or database asset.
