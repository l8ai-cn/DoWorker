# Agent Cloud Rebrand Naming Matrix

## Product
| Role | Old | New |
|------|-----|-----|
| Display name | Agent Cloud / Agent Cloud | Agent Cloud |
| Identifier / slug | agentcloud, agent-cloud | agentcloud, agent-cloud |
| Go module | github.com/l8ai-cn/agentcloud | github.com/l8ai-cn/agentcloud |
| GitHub repo | github.com/l8ai-cn/AgentCloud, Agent Cloud/Agent Cloud | github.com/l8ai-cn/AgentCloud |
| Public domain | agentcloud.ai / agentcloud.io / agentcloud.cn | agentcloud.ai / agentcloud.io / agentcloud.cn |
| Dev email domain | agentcloud.local | agentcloud.local |
| Harbor/Docker project | agentcloud, agent-cloud | agentcloud, agent-cloud |
| K8s namespace | agentcloud | agentcloud |
| DB user / DB name | agentcloud | agentcloud |
| Compose project prefix | agentcloud- | agentcloud- |
| Config dirs | ~/.agent-cloud, ~/.agentcloud | ~/.agent-cloud (writes); read legacy |
| System config | /etc/agent-cloud, /etc/agentcloud | /etc/agent-cloud (writes); read legacy |
| Runner binary | agent-cloud-runner (+ agentcloud-runner symlink) | agent-cloud-runner (+ legacy symlinks) |
| JS packages | @agent-cloud/*, agent-cloud-wasm | @agent-cloud/*, agent-cloud-wasm |
| Rust crates | agentcloud_* | agentcloud_* |
| Resource apiVersion | agentcloud.io/v1alpha1 | agentcloud.io/v1alpha1 |
| JWT audience | agentcloud-api | agentcloud-api |
| JWT issuer default | agentcloud | agentcloud |
| OTEL service names | agent-cloud-*, agentcloud-* | agent-cloud-* |
| MCP server / plugin id | agentcloud | agentcloud |

## Compatibility (required)
1. Config dir resolution: prefer `~/.agent-cloud`, then `~/.agent-cloud`, then `~/.agentcloud`.
2. System config: search `/etc/agent-cloud`, `/etc/agent-cloud`, `/etc/agentcloud`.
3. Binary: install `agent-cloud-runner`; keep `agent-cloud-runner` and `agentcloud-runner` as symlinks in images/installers.
4. apiVersion: accept both `agentcloud.io/v1alpha1` and legacy `agentcloud.io/v1alpha1`; emit new by default.
5. JWT audiences: default includes both `agentcloud-api` and `agentcloud-api` during transition; core audience is `agentcloud-api`.
6. DB: new installs use `agentcloud` credentials/db; migration renames role/db where possible and rewrites stored brand identifiers (MCP name, agentfile plugin, apiVersion constraints, lineage tags).
7. Harbor: manifests point at `.../agentcloud/...`. Operators must create/mirror Harbor project; dual-pull is operational, not code fallback logic.

## Out of scope / non-goals
- Renaming historical git commit messages or tags.
- Renaming third-party products (Codex, Claude, Anthropic, etc.).
- Automatically renaming the remote GitHub repository object (requires org admin); code/docs/CI refer to AgentCloud.

## Verification
- `go test` on packages touching apiVersion/config/auth.
- `rg` gates: no remaining production identifiers for old brand except explicit legacy aliases.
- Migration up/down compiles and unit tests pass for orchestration resource apiVersion dual-accept.
