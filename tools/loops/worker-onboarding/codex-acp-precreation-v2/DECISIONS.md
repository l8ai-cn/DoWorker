# Codex ACP Loop Decisions

## Active Human Gate

- Blocked reason: `approval-missing`
- Required decision: name one non-production organization model resource for a
  disposable Codex ACP Worker and authorize its cleanup.
- Prohibited before approval: reading credentials, validating the provider
  connection, creating a Pod, ACP interaction, and provider requests.
- Evidence: `evidence/codex-acp-precreation-2026-07-16.json`

The decision proxy may order read-only checks only.
