# Agent Cloud Rebrand Progress

## Status
- [x] Naming matrix
- [x] Bulk mechanical rename (module/packages/domains/images/k8s/docs)
- [x] Runtime compatibility code (config dirs, JWT audiences, apiVersion dual-accept, binary symlinks)
- [x] DB migration 000232
- [x] Proto Go regen for new module path
- [x] Focused compile/tests (server/runner/relay/config/type_meta)
- [ ] Push + CI confirmation

## Compatibility retained
- Config dirs: `~/.agent-cloud` preferred; read `~/.do-worker` and `~/.agentsmesh`
- System config: `/etc/agent-cloud`, `/etc/do-worker`, `/etc/agentsmesh`
- JWT audiences default includes `agentcloud-api` + legacy `agentsmesh-api`
- apiVersion accepts `agentcloud.io/v1alpha1` and legacy `agentsmesh.io/v1alpha1`
- Runner image keeps `do-worker-runner` and `agentsmesh-runner` symlinks
- Historical migration SQL left unchanged (checksums)

## CI fix preserved
- `clients/web/src/app/(dashboard)/DashboardShell.tsx` beforeunload abort for unread fetch
