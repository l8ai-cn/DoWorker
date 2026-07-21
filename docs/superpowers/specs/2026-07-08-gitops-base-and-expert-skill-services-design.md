# Gitops Base Service + Independent Git-backed Expert & Skill Services

- **Date:** 2026-07-08
- **Status:** Design / planning (no code changes beyond this document)
- **Repo:** `/Users/wwyz/Documents/智能体空间/doworker`
- **Author:** backend architecture

---

## 1. Goal / Non-Goals

### Goal

1. Extract a shared **gitops base service** that is the single choke point for all
   internal-Gitea repository lifecycle and content operations (create repo, ensure
   namespace, commit files, read file, list dir, list tree, delete repo, branch
   defaults, clone-URL resolution). Higher-level services stop importing
   `internal/infra/gitea` directly.
2. Make **Expert** an independent, Git-backed service: **one repo per expert**,
   storing `agent.md` (AgentFile/prompt source), `expert.json` (structured config),
   `README.md`, and `assets/` (avatar/images). The `experts` table becomes an
   index + run-config cache pointing at the repo.
3. Make **Skill** an independent, Git-backed / repo-oriented service that consumes
   the same gitops base service.
4. Establish the architectural rule: **gitops base at the bottom; expert, skill,
   and knowledgebase are independent peers on top of it**, none re-implementing Git.

### Non-Goals

- **Do NOT rewrite the knowledge base service now.** It keeps working against the
  same underlying Gitea. It *may* be migrated onto gitops later (optional step),
  but that is not required for this work.
- **Do NOT rewrite the runner/pod runtime.** The AgentFile layer that `expert.Run`
  produces and the `agentpod` orchestration path stay as-is; only the *source* of
  the AgentFile layer changes (Git-backed `agent.md` instead of DB column).
- **Keep experts one-repo-per-expert** (already decided). No mono-repo.
- **Avatar (形象) and expert type (类型) live in metadata/config JSON**, not new DB
  columns. New DB columns are limited to git-index fields (`git_repo_path`,
  `default_branch`, `metadata` jsonb).
- No change to external git-provider import (`internal/infra/git`, the
  `repository` service, webhooks). Those model *user* repos, not platform-owned
  repos, and are out of scope.

---

## 2. Current State

### 2.1 Internal Gitea client — `backend/internal/infra/gitea/`

The low-level HTTP client for the platform-owned Gitea instance. It is admin-token
scoped and provisions repos under one namespace org.

- `client.go`
  - `type Config struct { BaseURL, AdminToken, Namespace, CloneBaseURL string }`,
    `Config.Enabled()`, default namespace `am-kb`.
  - `func NewClient(cfg Config) *Client`
  - `func (c *Client) Namespace() string`
  - `type Repo struct { Name, FullName, CloneURL, DefaultBranch string }`
  - `func (c *Client) EnsureNamespace(ctx) error` — GET org, else POST `/orgs`.
  - `func (c *Client) CreateRepo(ctx, name, defaultBranch string) (*Repo, error)` —
    POST `/orgs/{ns}/repos`, `auto_init:false`, `private:true`.
  - `func (c *Client) DeleteRepo(ctx, name string) error`
  - `func (c *Client) CloneURL(name string) string` — runner-facing HTTPS URL.
  - `func (c *Client) CloneToken() string` — admin token used as basic-auth password.
  - `escapePath(p string) string` helper.
- `contents.go`
  - `type FileChange struct { Path, Content string }`
  - `type CommitAuthor struct { Name, Email string }`
  - `func (c *Client) CommitFiles(ctx, repo, branch, message string, author CommitAuthor, changes []FileChange, isUpdate map[string]string) error` —
    batch create/update in one commit; `isUpdate[path] = sha` switches an entry
    from `create` to `update`. Works on empty repos (seeds initial commit).
  - `type ContentEntry struct { Name, Path, SHA, Type, Size, Content }` +
    `DecodedContent()`.
  - `func (c *Client) GetFile(ctx, repo, branch, path) (*ContentEntry, error)`
  - `func (c *Client) ListDir(ctx, repo, branch, path) ([]*ContentEntry, error)`
- `tree.go`
  - `type TreeEntry struct { Path, Type, Size, SHA }`
  - `func (c *Client) ListTree(ctx, repo, ref) ([]TreeEntry, error)` — recursive
    tree in one call (used by KB search).

Wiring: `backend/cmd/server/services_init_helpers.go::initializeKnowledgeBaseService`
builds the client from `cfg.KnowledgeBase` (`KB_GITEA_URL` / `KB_GITEA_TOKEN` /
`GiteaOrg` / `CloneBaseURL`) and passes it into the KB service. If not configured,
the whole KB feature is nil ("disabled").

### 2.2 Knowledge base service (best existing reference) — `backend/internal/service/knowledgebase/`

This is the pattern to generalize. It owns a `*gitea.Client` directly.

- `service.go`
  - `type Service struct { repo knowledgebase.Repository; git *gitea.Client; log; secrets }`.
  - `NewService(repo, git, log)` returns `nil` when `git == nil` (feature disabled).
  - `Create`: derive unique slug (`slugkit.GenerateUnique` + `SlugExists`), call
    `provisionRepo`, then persist a `knowledge_bases` row with `GitRepoPath =
    Namespace()+"/"+repoName`, `HTTPCloneURL = git.CloneURL(repoName)`,
    `DefaultBranch`. **If DB `Create` fails → `git.DeleteRepo(repoName)` and return
    error** (compensating cleanup).
  - `Delete`: delete DB row first (authoritative), then best-effort `DeleteRepo`
    (an orphan repo is "just garbage").
  - `repoNameFromPath(path)` — strips namespace prefix off `git_repo_path`.
- `provisioner.go`
  - `//go:embed scaffold/*.tmpl`, `scaffoldFiles` map (`llms.txt`, `AGENTS.md`,
    `wiki/index.md`, `wiki/log.md`, `raw/README.md`), `renderScaffold(name, desc)`.
  - `provisionRepo(ctx, orgID, slug, name, desc, branch)`:
    `EnsureNamespace` → `repoName := fmt.Sprintf("org%d-%s", orgID, slug)` →
    `CreateRepo` → `CommitFiles(...seed scaffold...)`; **if commit fails →
    `DeleteRepo` and return error.**
- `contents.go` — `ReadFile`, `ListDir`, `CommitFile` (probe SHA to decide
  create-vs-update), all thin wrappers over the gitea client keyed by
  `repoNameFromPath(kb.GitRepoPath)` + `kb.DefaultBranch`.
- `sync.go` / `sync_worker.go` / `search.go` — consumers of `ListTree` / `GetFile`.
- Domain `backend/internal/domain/knowledgebase/knowledgebase.go`: the
  `KnowledgeBase` row already carries `GitRepoPath`, `HTTPCloneURL`,
  `DefaultBranch`, `SourceType`, `SourceConfig jsonb` — the shape we want experts
  to converge toward.

**Repo naming convention already in use:** `org<ID>-<slug>` inside a single shared
namespace org (default `am-kb`). KB slugs are unique per Agent Cloud org; the org
prefix disambiguates across orgs in one namespace.

### 2.3 Expert service (currently pure-DB) — `backend/internal/service/expert/`

No Git at all today. Everything is DB columns.

- Domain `backend/internal/domain/expert/expert.go` + migration
  `backend/migrations/000178_experts.up.sql`: `experts` table with `slug`, `name`,
  `description`, `agent_slug`, `runner_id`, `repository_id`, `branch_name`,
  `prompt TEXT`, `interaction_mode`, `perpetual`, `used_env_bundles text[]`,
  `skill_slugs text[]`, `knowledge_mounts jsonb`, `config_overrides jsonb`,
  `agentfile_layer TEXT`, `source_pod_key`, run-count/last-run. **No avatar, no
  expert type, no git columns today.**
- `service.go` — `Service{ store expertdom.Repository; pods; dispatch; repos RepoResolver; logger }`.
- `crud.go` — `Create`/`Update` build the row entirely from request fields and call
  `store.Create/Update`. Slug via `slugkit.GenerateUnique` + `store.SlugExists`.
- `run.go` — `buildAgentfileLayer(ctx, expert)`: if `expert.AgentfileLayer` is set,
  use it verbatim; otherwise **synthesize** AgentFile directives from DB columns
  (`MODE`, `PROMPT`, `USE_ENV_BUNDLE`, `CONFIG`, `REPO`/`BRANCH` via
  `repos.GetByID`, `SKILLS`, `KNOWLEDGE`). `Run` passes the layer + knowledge
  mounts into `agentpodSvc.OrchestrateCreatePodRequest`.
- `publish.go` — `PublishFromPod` builds a `CreateExpertRequest` from a running pod.
- REST `backend/internal/api/rest/v1/expert_handler.go` + `_types.go` + `routes_expert.go`:
  CRUD + `POST /experts/:slug/run` + `POST /pods/:pod_key/publish-expert`; request
  DTOs mirror the DB columns 1:1.

### 2.4 Skill / extension service — `backend/internal/service/extension/`

Skills are **imported from external Git and packaged into object storage**, not
provisioned into internal Gitea.

- `skill_importer.go` / `skill_importer_git.go`: clone external repo (`gitClone`,
  `gitCloneWithAuth`, SSH/PAT), `detectRepoType` (single/collection), `parseSkillDir`
  reads `SKILL.md`, `computeDirSHA`, `packageSkillDir` → tar.gz →
  `storage.Upload("skills/{registryID}/{slug}/{sha}.tar.gz")` → upsert
  `SkillMarketItem` DB rows.
- `skill_packager.go`: `PackageFromGitHub` / `PackageFromUpload` → `packageDir` →
  upload to `skills/direct/{slug}/{sha}.tar.gz`.
- `service_install.go`: `InstallSkillFromMarket/GitHub/Upload/UploadedKey` create
  `InstalledSkill` rows scoped to an org/repo; experts reference skills **by slug**
  (`expert.SkillSlugs pq.StringArray` → `SKILLS` AgentFile directive).

Key point: the skill subsystem uses `os/exec git clone` + object storage + tar.gz,
**not** the internal-Gitea contents API. The gitops base service targets the
internal-Gitea path (provisioned platform repos). Skills that we want to be
*Git-backed platform artifacts* (authored/edited in-platform) will use gitops; the
existing external-import + marketplace flow stays as an additional source.

---

## 3. Gitops Base Service Design

### 3.1 Package location

Create **`backend/internal/service/gitops`** (a service-layer package, peer to
`knowledgebase`/`expert`/`extension`). Rationale:

- It composes over `internal/infra/gitea` (the raw HTTP client stays as infra) and
  adds domain-agnostic policy: repo naming, namespaces-per-domain, seed-commit +
  compensating cleanup, error normalization. That policy is *service* behavior, not
  transport, so `internal/infra/gitops` would be the wrong layer.
- Keeps `internal/infra/gitea` a thin, dumb transport client (unchanged).

`gitops` imports `internal/infra/gitea`. `expert`, `skill`, (optionally)
`knowledgebase` import `gitops` and **must not** import `internal/infra/gitea`.

### 3.2 Namespaces per domain

Today all repos share one namespace org (`am-kb`). We keep one Gitea client but let
each *domain* declare its namespace so repos don't collide and are easy to browse:

- knowledge bases → `am-kb` (unchanged)
- experts → `am-experts`
- skills → `am-skills`

The gitops service is constructed with a **namespace** (per domain) or exposes a
`Namespaced(ns)` view. Concretely, one shared `*gitea.Client` per Gitea instance,
and the gitops layer parameterizes the namespace. Because the raw `gitea.Client`
bakes the namespace into `Config`, we either (a) construct one `gitea.Client` per
namespace, or (b) extend the client to accept a namespace per call. Recommended:
**one gitops service instance per domain namespace** built from a shared config
(mirrors how KB is wired today), so no change to `gitea.Client` is forced.

### 3.3 Repo naming convention

Reuse the KB convention, generalized:

```
<namespace>/org<ORG_ID>-<slug>
```

- namespace = domain namespace (`am-experts`, `am-skills`, `am-kb`).
- `org<ID>-` prefix disambiguates the org within one namespace (slugs are unique
  per org, not globally).
- `git_repo_path` stored on the row = `"<namespace>/org<ID>-<slug>"`; the bare repo
  name (`org<ID>-<slug>`) is what the gitea contents API needs. gitops owns the
  `RepoName(orgID, slug)` ↔ `RepoPath` conversion (generalizing KB's
  `repoNameFromPath`).

### 3.4 Interface sketch

```go
package gitops

// Author identifies the committer for seed/edit commits.
type Author struct {
	Name  string
	Email string
}

// FileChange is a create-or-update of one path in a commit.
type FileChange struct {
	Path    string
	Content []byte // bytes so binary assets (avatars) are first-class
}

// Entry is a directory/tree listing item.
type Entry struct {
	Name string
	Path string
	Type string // "file" | "dir"
	Size int64
	SHA  string
}

// Repo is the provisioned repository descriptor.
type Repo struct {
	Namespace     string
	Name          string // org<ID>-<slug>
	Path          string // namespace/name  -> stored as git_repo_path
	DefaultBranch string
	HTTPCloneURL  string
}

// ProvisionParams drives create-repo + seed-initial-commit atomically.
type ProvisionParams struct {
	OrgID         int64
	Slug          string
	DefaultBranch string       // "" -> "main"
	CommitMessage string       // seed commit message
	Author        Author       // "" -> platform default
	Seed          []FileChange // initial files; empty repo if nil
}

// Service is the single choke point for platform-owned repo operations.
// One instance is bound to one namespace (per domain).
type Service interface {
	// Namespace returns the domain namespace this instance manages.
	Namespace() string

	// EnsureNamespace makes the namespace org exist (idempotent).
	EnsureNamespace(ctx context.Context) error

	// Provision creates the repo and seeds the initial commit in one shot.
	// On seed-commit failure it deletes the repo and returns the error, so
	// callers never see a half-created repo.
	Provision(ctx context.Context, p ProvisionParams) (*Repo, error)

	// Commit creates/updates files in one commit. Create-vs-update is decided
	// per path by probing current SHA unless the caller passes known SHAs.
	Commit(ctx context.Context, repoName, branch, message string, a Author, changes []FileChange) error

	// ReadFile returns decoded file content.
	ReadFile(ctx context.Context, repoName, branch, path string) ([]byte, *Entry, error)

	// ListDir lists one directory level.
	ListDir(ctx context.Context, repoName, branch, path string) ([]Entry, error)

	// ListTree enumerates the whole tree recursively (one call).
	ListTree(ctx context.Context, repoName, ref string) ([]Entry, error)

	// DeleteRepo removes the repo (best-effort cleanup helper for callers).
	DeleteRepo(ctx context.Context, repoName string) error

	// RepoName maps (orgID, slug) -> "org<ID>-<slug>".
	RepoName(orgID int64, slug string) string
	// RepoPath maps (orgID, slug) -> "<namespace>/org<ID>-<slug>".
	RepoPath(orgID int64, slug string) string
	// RepoNameFromPath strips the namespace prefix off a stored git_repo_path.
	RepoNameFromPath(path string) string

	// CloneURL / CloneToken expose runner-facing clone credentials, unchanged
	// from the current gitea client behavior.
	CloneURL(repoName string) string
	CloneToken() string
}
```

Notes:

- `Commit` folds KB's `CommitFile` SHA-probe logic into the base service so no
  consumer re-implements create-vs-update detection.
- `Provision` folds `provisionRepo`'s "create → seed → delete-on-failure" into one
  method — the single most valuable reuse.
- `FileChange.Content []byte` (vs KB's `string`) so avatar/image assets in
  `assets/` are first-class (base64 handling stays inside the gitea client, which
  already base64-encodes on commit / decodes on read).

### 3.5 Construction / wiring

Add `gitops.NewService(git *gitea.Client, namespace string, log *slog.Logger)
Service`, returning `nil` when `git == nil` (same "feature disabled" convention as
KB). In `services_init_helpers.go`, build one shared `*gitea.Client` from
`cfg.KnowledgeBase` (or a renamed shared `cfg.Gitea`) and construct one gitops
instance per namespace:

```go
expertGitops := gitops.NewService(giteaClient, "am-experts", log)
skillGitops   := gitops.NewService(giteaClient, "am-skills", log)
// KB may keep its own path for now, or:
// kbGitops   := gitops.NewService(giteaClient, "am-kb", log)
```

(If we keep `gitea.Client.Config.Namespace` per-instance as it is today, construct
one `gitea.Client` per namespace instead — either is fine, one-client-per-namespace
is the smaller change.)

### 3.6 Error handling

- Wrap transport errors with a domain prefix (`gitops: create repo: %w`).
- Export sentinel `gitops.ErrNotConfigured = gitea.ErrNotConfigured` and
  `gitops.ErrNotFound` (map Gitea 404 → `ErrNotFound` in `ReadFile`/`ListDir`) so
  consumers can branch without string-matching.
- `Provision` is the only method with compensating cleanup; all other content
  methods are single API calls and return raw wrapped errors.

---

## 4. Expert Service (independent, Git-backed) Design

### 4.1 Repo layout (one repo per expert)

Namespace `am-experts`, repo `org<ID>-<slug>`:

```
/agent.md          # AgentFile / prompt source (authoritative run-time source)
/expert.json       # structured config (see below)
/README.md         # human-facing description
/assets/           # avatar/images (avatar.png, ...)
```

`expert.json` (the home of 形象 avatar + 类型 type, plus the config that today
lives in DB columns):

```jsonc
{
  "schema": 1,
  "name": "Data Analyst",
  "description": "...",
  "avatar": "assets/avatar.png",     // 形象 -> path within repo (or external URL)
  "expertType": "analysis",          // 类型
  "agentSlug": "claude-code",
  "interactionMode": "pty",          // pty | acp
  "perpetual": false,
  "skillSlugs": ["web-search", "sql"],
  "knowledgeMounts": [{ "slug": "team-docs", "mode": "ro" }],
  "usedEnvBundles": ["default"],
  "configOverrides": { "model": "opus" },
  "repository": { "repositoryId": 12, "branch": "main" }
}
```

`agent.md` becomes the AgentFile source. Today `buildAgentfileLayer` synthesizes a
layer from DB columns; going forward the **authored `agent.md` is the source of
truth**, and the DB-column synthesis becomes the fallback/generator used to seed
`agent.md` on first create (and on migration of existing experts).

### 4.2 DB table: index + run-config cache

Keep the existing `experts` columns as a **cache** (so listing, run, and the
existing REST payloads keep working without a Git round-trip). **Add only**:

- `git_repo_path VARCHAR(255)` — `am-experts/org<ID>-<slug>`
- `default_branch VARCHAR(255) NOT NULL DEFAULT 'main'`
- `http_clone_url VARCHAR(1000)` — runner-facing clone URL (parity with KB)
- `metadata JSONB NOT NULL DEFAULT '{}'` — holds **avatar + expert type** (and any
  other new config that must NOT become a column), the first-version snapshot.

**Explicitly NOT adding** avatar/type columns — they live in `metadata` (DB cache)
and `expert.json` (Git source of truth). Existing columns (`prompt`,
`interaction_mode`, `skill_slugs`, `knowledge_mounts`, `config_overrides`,
`agentfile_layer`, ...) remain as the cache/index so nothing downstream breaks; on
write they are also serialized into `expert.json` / `agent.md`.

Migration (next number is **000181**, after `000180_virtual_api_keys_quotas`):

```sql
-- 000181_experts_git_backing.up.sql
ALTER TABLE experts ADD COLUMN git_repo_path  VARCHAR(255);
ALTER TABLE experts ADD COLUMN default_branch VARCHAR(255) NOT NULL DEFAULT 'main';
ALTER TABLE experts ADD COLUMN http_clone_url VARCHAR(1000);
ALTER TABLE experts ADD COLUMN metadata       JSONB NOT NULL DEFAULT '{}';
```

Domain: extend `backend/internal/domain/expert/expert.go` `Expert` struct with
`GitRepoPath string`, `DefaultBranch string`, `HTTPCloneURL string`,
`Metadata json.RawMessage`. `git_repo_path` may be NULL for legacy rows → treat
NULL as "not yet git-backed" (lazy backfill on next update, or a backfill job).

### 4.3 Service composition

```go
type Service struct {
	store    expertdom.Repository
	gitops   gitops.Service   // NEW: namespace = am-experts
	pods     PodLoader
	dispatch PodDispatcher
	repos    RepoResolver
	logger   *slog.Logger
}
```

`gitops.Service` is fakeable (interface), so expert unit tests need no live Gitea.
When `gitops == nil`, the expert service still functions in DB-only mode (graceful
degradation / feature-flag), matching the KB "nil = disabled" convention.

### 4.4 Flow changes

**Create** (`crud.go`):
1. Validate + resolve unique slug (unchanged).
2. Build `expert.json` + `agent.md` (agent.md via existing `buildAgentfileLayer`
   generator, or the caller-supplied `AgentfileLayer`/`Prompt`) + `README.md` +
   optional `assets/avatar.*`.
3. `gitops.Provision(ProvisionParams{OrgID, Slug, DefaultBranch:"main",
   CommitMessage:"init: expert scaffold", Seed: [...]})`.
4. Persist the `experts` row with `git_repo_path`, `default_branch`,
   `http_clone_url`, `metadata` (avatar+type), plus the cache columns.
5. **If DB `Create` fails → `gitops.DeleteRepo(repoName)`** (mirror KB).

**Update** (`crud.go`):
1. Load row, apply field changes (unchanged cache columns).
2. Re-render changed files and `gitops.Commit(...)` (SHA-probe create-vs-update
   handled inside gitops).
3. Update cache columns + `metadata`. Commit-then-DB or DB-then-commit ordering:
   prefer commit first, then DB update; on DB failure the Git history simply has an
   extra commit (benign), consistent with KB treating the DB row as authoritative.

**Get / List**: served from DB cache (fast, no Git round-trip). Add
`GetContent`/`ReadFile`/`ListDir` methods that go through gitops for the authored
`agent.md` / `expert.json` / assets when the UI needs the source.

**Run** (`run.go`): unchanged orchestration. The AgentFile layer is sourced from
`agent.md` (read via gitops, cached in `agentfile_layer`) instead of synthesized
from columns. `buildAgentfileLayer` is retained as the generator that produces
`agent.md` at author time and as the fallback for legacy rows without a repo.

**PublishFromPod** (`publish.go`): unchanged entry point; it flows through
`Create`, which now also provisions the repo + seed commit.

### 4.5 REST / API impact

- `expert_handler_types.go`: add optional `avatar` / `expert_type` (and any new
  config) fields to `createExpertRequest` / `updateExpertRequest`; the handler
  routes them into `expert.json` + `metadata` (no new top-level DB columns).
- Response payload: expose `git_repo_path`, `default_branch`, `avatar`,
  `expert_type` (read from `metadata`) so the frontend
  `clients/web/src/components/experts/*` (e.g. `expertFormModel.ts`,
  `ExpertConfigFields.tsx`) can render 形象/类型.
- Add read endpoints for repo content if the UI needs to show/edit `agent.md`:
  `GET /experts/:expertSlug/files/*path` and `GET /experts/:expertSlug/tree`
  (thin pass-throughs to `gitops.ReadFile`/`ListTree`), mirroring the KB contents
  endpoints. Routes registered in `routes_expert.go`.

---

## 5. Skill Service (independent, Git-backed) Design

Two skill "sources" coexist:

1. **External/marketplace skills** (existing): cloned from external Git, packaged to
   object storage as tar.gz, indexed as `SkillMarketItem` / `InstalledSkill`. This
   flow is **unchanged** and does not use gitops.
2. **Platform-authored skills** (new, Git-backed via gitops): a skill can be created
   and edited in-platform, stored in its own internal-Gitea repo.

### 5.1 Package + composition

Add a `skill` service (either a new `backend/internal/service/skill` package or a
Git-backed submodule of `extension`) that holds a `gitops.Service` bound to
namespace `am-skills`. Like expert, it takes the interface so it is unit-testable
without Gitea.

### 5.2 Repo / dir layout

Namespace `am-skills`, one repo per authored skill `org<ID>-<slug>`:

```
/SKILL.md          # canonical skill definition (same format the importer parses)
/skill.json        # structured metadata (slug, displayName, description,
                   #   license, compatibility, allowedTools, category)
/scripts/          # optional helper scripts
/assets/           # optional
```

`SKILL.md` is deliberately the same shape `parseSkillDir` already understands, so
the packager can consume a gitops repo identically to an external clone.

### 5.3 Relationship to the existing install flow

- Publishing/authoring a skill: `gitops.Provision` the `am-skills/org<ID>-<slug>`
  repo with a `SKILL.md` + `skill.json` seed; index it as a `SkillMarketItem` with a
  source marker like `install_source = "gitops"` (extend `InstallSource`).
- Packaging for install/runtime reuses the existing pipeline: read the repo tree via
  `gitops.ListTree`/`ReadFile` (instead of `git clone`), then `packageSkillDir` →
  `storage.Upload("skills/gitops/{slug}/{sha}.tar.gz")` → upsert market item →
  `InstallSkillFromMarket`. This keeps runtime skill delivery (tar.gz in object
  storage) unchanged; only *authoring/storage of source* moves to gitops.
- Edits: `gitops.Commit` a new `SKILL.md`, re-package, bump `SkillMarketItem.Version`.

### 5.4 How experts reference skills

Unchanged: experts reference skills **by slug** via `expert.SkillSlugs` (and
`expert.json.skillSlugs`), which becomes the `SKILLS` AgentFile directive in
`buildAgentfileLayer`. Whether a skill's source is external-clone or gitops-backed
is invisible to the expert — it only sees slugs resolved through the extension
install layer at run time.

---

## 6. Service Boundaries Diagram

```
                      REST / API  (v1 handlers)
   ┌───────────────┬──────────────────┬────────────────────┐
   │               │                  │                    │
┌──▼───────────┐ ┌─▼──────────────┐ ┌─▼────────────────┐   │
│  Expert svc  │ │  Skill svc     │ │ KnowledgeBase svc│   │  (independent peers)
│ (am-experts) │ │ (am-skills)    │ │ (am-kb)          │   │
│              │ │                │ │  *migrate later* │   │
└──────┬───────┘ └───────┬────────┘ └────────┬─────────┘   │
       │                 │                   │             │
       │  gitops.Service (interface, fakeable in tests)    │
       └────────────────┬┴───────────────────┘             │
                        │                                   │
              ┌─────────▼──────────┐                        │
              │  gitops base svc   │  naming, namespaces,   │
              │ internal/service/  │  provision+seed+clean, │
              │      gitops        │  content read/write,   │
              │                    │  error normalization   │
              └─────────┬──────────┘                        │
                        │ wraps                              │
              ┌─────────▼──────────┐                        │
              │  gitea HTTP client │  transport only        │
              │ internal/infra/    │  (unchanged)           │
              │      gitea         │                        │
              └─────────┬──────────┘                        │
                        │                                   │
                 ┌──────▼───────┐                           │
                 │ internal     │                           │
                 │ Gitea server │                           │
                 └──────────────┘                           │

  (Out of scope, unchanged): external git-provider import
  internal/infra/git + service/repository + skill external-clone/tar.gz path.
```

Rule enforced: only `gitops` imports `internal/infra/gitea`. Expert/Skill (and
eventually KB) import `gitops`, never the raw client.

---

## 7. Data Flow — "Create Expert" end to end

```
Frontend new-expert page
  clients/web/src/components/experts/* (name, description, avatar upload,
  expertType, agentSlug, interactionMode, skillSlugs, knowledgeMounts, prompt)
        │  POST /api/v1/orgs/:org/experts  { ..., avatar, expert_type }
        ▼
REST  expert_handler.go::CreateExpert
        │  build expertSvc.CreateExpertRequest (+ avatar, expertType)
        ▼
expert.Service.Create (crud.go)
  1. validateExpertBasics + resolveSlug (slugkit.GenerateUnique / SlugExists)
  2. render seed files:
        agent.md     <- buildAgentfileLayer(generator) or supplied prompt/layer
        expert.json  <- name/desc/avatar(type)/agentSlug/mode/skills/mounts/...
        README.md
        assets/avatar.<ext>  (if uploaded)
  3. gitops.Provision(ProvisionParams{
        OrgID, Slug, DefaultBranch:"main",
        CommitMessage:"init: expert scaffold (agent.md, expert.json, README, assets/)",
        Author: platform default, Seed: [...] })
        └─ gitops: EnsureNamespace(am-experts)
                 -> CreateRepo(org<ID>-<slug>)
                 -> CommitFiles(seed)   [on failure: DeleteRepo + err]
  4. store.Create(experts row){
        slug,name,desc, cache columns (prompt, mode, skill_slugs, ...),
        git_repo_path="am-experts/org<ID>-<slug>",
        default_branch, http_clone_url,
        metadata={avatar, expertType} }
        └─ on DB failure: gitops.DeleteRepo(org<ID>-<slug>)   [mirror KB]
        ▼
  201 { expert: {..., git_repo_path, avatar, expert_type} }
```

Run-time (later): `expert.Run` → read/cache `agent.md` → AgentFile layer →
`agentpodSvc.OrchestrateCreatePodRequest` (unchanged).

---

## 8. Error Handling & Cleanup

Mirror the knowledgebase provisioner's compensating-transaction behavior, now
centralized in `gitops.Provision` plus caller-side DB cleanup:

| Failure point                          | Action                                                    |
|----------------------------------------|-----------------------------------------------------------|
| `EnsureNamespace` fails                | return error; nothing created                             |
| `CreateRepo` fails                     | return error; nothing to clean                            |
| Seed `CommitFiles` fails               | `gitops.Provision` calls `DeleteRepo`, returns error      |
| DB `store.Create` fails after Provision| service calls `gitops.DeleteRepo(repoName)`, returns error|
| `Update` commit fails                  | return error; DB unchanged (no partial cache write)       |
| DB `Update` fails after commit         | log; Git has an extra benign commit; DB row authoritative |
| `Delete`                               | delete DB row first (authoritative), best-effort DeleteRepo, log on failure |
| Gitea 404 on read                      | map to `gitops.ErrNotFound`                               |

Principle (from KB): the **DB row is authoritative**; an orphaned Gitea repo is
"just garbage," not a correctness bug. Repo cleanup is best-effort on delete and
strict (compensating) on create.

---

## 9. Testing Strategy

- **`gitops.Service` is an interface** → provide `gitops.Fake` (in-memory map of
  `repoName -> map[path][]byte`, with branch + SHA bookkeeping) so expert/skill
  service unit tests run with no live Gitea, no network. This is the main win of the
  interface boundary.
- **gitops package tests**: keep the existing `httptest.Server` approach that KB
  tests already use (`gitea.NewClient(gitea.Config{BaseURL: srv.URL})`, see
  `knowledgebase/sync_test.go`) to test the real gitea client wrapping + the
  Provision/cleanup logic against a fake Gitea HTTP server.
- **Expert service tests**: inject a `gitops.Fake`; assert seed files
  (`agent.md`/`expert.json`) content, `git_repo_path`/`metadata` persisted,
  DB-failure → `DeleteRepo` called, and `buildAgentfileLayer` still produces the
  correct AgentFile from `agent.md`.
- **Skill service tests**: inject `gitops.Fake`; assert `SKILL.md`/`skill.json`
  seed and that `parseSkillDir`-compatible content round-trips into the packager.
- **Nil-gitops path**: assert expert/skill services degrade to DB-only when
  `gitops == nil` (feature disabled), matching KB's nil-service convention.

---

## 10. Incremental Implementation Order

1. **gitops base service** — new `backend/internal/service/gitops` package +
   interface + `Fake`, wrapping the unchanged `gitea` client; wire one instance per
   namespace in `services_init_helpers.go`. Unit-test in isolation. No consumer
   changes yet. *(Ships value immediately; nothing breaks.)*
2. **(Optional) migrate knowledgebase onto gitops** — swap `s.git *gitea.Client`
   for `gitops.Service`, delete KB's `provisionRepo`/`repoNameFromPath` in favor of
   `gitops.Provision`/`RepoNameFromPath`. Pure refactor, behavior-preserving,
   guarded by the existing KB test suite. Skip if risk/time-boxed.
3. **Expert repo-backing** — migration 000181 (git columns + `metadata`); extend
   domain struct; add `gitops.Service` to `expert.Service`; implement
   Create/Update/Get content flows + seed rendering; backfill legacy rows lazily;
   extend REST DTOs/response for avatar+type; add content read endpoints. Keep DB
   cache columns so `Run` and existing clients are untouched.
4. **Skill service (Git-backed)** — new `skill` service (or `extension` submodule)
   with `gitops.Service` bound to `am-skills`; author/publish/edit flows; bridge to
   the existing packager (`ListTree`/`ReadFile` → `packageSkillDir` → object
   storage → `SkillMarketItem`); add `install_source="gitops"`.
5. **Frontend** — surface 形象/类型 and (optional) `agent.md` viewer/editor in
   `clients/web/src/components/experts/*`.

Each step is independently shippable and leaves the system working.

---

## 11. Open Questions (need user decision)

1. **Namespaces:** OK to introduce `am-experts` and `am-skills` (vs. reusing the
   single `am-kb` namespace)? And do we reuse the existing `KB_GITEA_*` config /
   admin token for all three domains, or introduce a shared `GITEA_*` config block?
2. **One gitea.Client per namespace vs. per-call namespace:** acceptable to
   construct one `gitea.Client` per namespace (zero change to the infra client), or
   do you want the client extended to take a namespace per call?
3. **`agent.md` source of truth vs. cache:** confirm the direction — authored
   `agent.md` in Git is authoritative and `experts.prompt` / `agentfile_layer`
   columns become a derived cache. Or should DB stay authoritative and Git be a
   mirror for the first version?
4. **Legacy expert backfill:** lazy (provision repo on next update) or a one-time
   backfill job that provisions repos for all existing experts?
5. **Skill scope:** should the Git-backed skill service fully replace authored-skill
   storage, or only add a new "author in platform" source alongside the existing
   external-import + marketplace flow (this design assumes the latter)?
6. **Avatar storage:** avatar bytes committed into the expert repo `assets/`, or
   kept in the existing S3/file service with only the URL in `expert.json`/`metadata`?
   (This design supports committing to `assets/` but either works.)
7. **Should knowledgebase migration (step 2) be in scope now**, or explicitly
   deferred?
```