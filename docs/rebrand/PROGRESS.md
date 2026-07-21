# Agent Cloud Rebrand Progress

## Status
- [x] Naming matrix
- [x] Bulk mechanical rename
- [x] Runtime compatibility removed (latest-only Agent Cloud identifiers)
- [x] DB migration 000232 (migrate then enforce agentcloud apiVersion only)
- [x] Proto/catalog hash fixes
- [ ] Merge to main + oilan deploy

## Compatibility policy (final)
No dual-brand runtime paths. Only:
- Config dir: `~/.agent-cloud`, `/etc/agent-cloud`
- JWT audience: `agentcloud-api`
- apiVersion: `agentcloud.io/v1alpha1`
- Runner binary: `agent-cloud-runner`
- Auth storage namespace: `agent-cloud-auth`

Historical migration SQL remains immutable (`agentsmesh` strings inside old files).
