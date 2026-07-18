# Aider Loop Decisions

## Active Human Gate

- Blocked reason: `approval-missing`
- Required decision: name one non-production credential bundle and resource scope
  for a disposable Aider launch, then authorize its cleanup.
- Prohibited before approval: reading credential values, injecting credential
  references, creating a Pod, terminal interaction, and provider requests.
- Evidence: `evidence/aider-precreation-2026-07-16.json`

## Delegated Low-Risk Authority

The decision proxy may order read-only checks and record evidence. It may not
approve a launch, credential use, deployment, push, merge, or deletion.
