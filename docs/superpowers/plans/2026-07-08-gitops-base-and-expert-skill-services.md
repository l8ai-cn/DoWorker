# Implementation Plan — Gitops Base Service + Git-backed Expert & Skill Services

- **Date:** 2026-07-08
- **Repo:** `/Users/wwyz/Documents/智能体空间/doworker`
- **Design spec:** `docs/superpowers/specs/2026-07-08-gitops-base-and-expert-skill-services-design.md`
- **Status:** Implementation-ready plan. This document is the only file this task creates; no source is modified.

---

## Decisions (locked)

These are user decisions. Do **not** re-open them during implementation.

1. **Source of truth = Git.** `agent.md` (prompt/AgentFile source) and `expert.json` (structured config) in the expert's repo are authoritative. The `experts` DB columns become a derived index/run-config cache.
2. **Avatar (形象) is committed into the repo `assets/` directory** and versioned with the expert. `expert.json` / `metadata` holds the **relative path** (e.g. `assets/avatar.png`). Avatar bytes are **NOT** stored in S3.
3. **Expert type (类型) and avatar live in `metadata` jsonb + `expert.json`**, NOT in new dedicated DB columns.
4. **Skill Git-backing is ADDITIVE.** A new "author-in-platform" source that coexists with the existing external-import / marketplace flow. It is not a replacement.
5. **Knowledge base is NOT migrated onto gitops this phase.** `knowledgebase` keeps its current direct `*gitea.Client` usage. Migrating it onto gitops is explicitly deferred (future work).
6. **One repo per expert is fixed.** Reuse the knowledgebase provisioner pattern: `org<ID>-<slug>` naming inside a shared namespace org, with compensating cleanup on failure.
7. **Legacy backfill = lazy.** Experts created before this feature (no `git_repo_path`) are provisioned a repo on their **next update / run** (see Rollout section). No one-time backfill job in this phase.

---

## Findings from code re-read (deltas vs. the design spec)

The spec is largely accurate. The following concrete details were verified/corrected and are baked into the phases below:

1. **Latest migration on disk is `000183_pod_preview` — CONFIRMED.** `000181_im_channel_bridges`, `000182_im_weixin_support`, and `000183_pod_preview` all exist, so the current max is **000183**. The next free number is **000184** (verified free by listing `backend/migrations/`). The experts migration is therefore **`000184_experts_git_backing`**. (Files: `backend/migrations/000183_pod_preview.{up,down}.sql`; experts table created in `000178_experts.up.sql`.)
2. **`gitea.FileChange.Content` is a `string`, not `[]byte`** (`backend/internal/infra/gitea/contents.go`). The gitops interface intentionally uses `[]byte` (for binary avatars). `CommitFiles` already does `base64.StdEncoding.EncodeToString([]byte(ch.Content))`, so the gitops→gitea conversion `string(fileChange.Content)` is lossless for binary. **No infra change required for binary content.**
3. **There is NO `gitea.ErrNotFound` sentinel.** `gitea.Client.do()` returns a *formatted string* error for any status ≥ 300 (e.g. `"gitea: GET /... → 404: ..."`); only `gitea.ErrNotConfigured` exists (`client.go`). To expose `gitops.ErrNotFound` cleanly, Phase 1 adds a **typed error to the gitea client** (`gitea.HTTPError{StatusCode int, ...}`) so gitops can map 404 → `ErrNotFound` without string-matching. This is a small, self-contained infra addition (the only change to `internal/infra/gitea`).
4. **Namespace is baked into `gitea.Config`** (`Config.Namespace`, default `am-kb`); every client method uses `c.cfg.Namespace`. Therefore the simplest path is **one `*gitea.Client` per namespace** (no per-call namespace change). gitops wraps one client and is bound to one namespace — matches the spec's recommended option.
5. **The expert service is wired in `backend/cmd/server/main.go`** (around line 208, `svc.Expert = expertSvc.NewService(expertSvc.Deps{...})`), **not** in `services_init_helpers.go`. The KB gitea client is built inside `services_init_helpers.go::initializeKnowledgeBaseService` and is **not returned** to callers. Phase 1/3 therefore adds a small helper that constructs a per-namespace `*gitea.Client` (reusing `cfg.KnowledgeBase` URL/token/clone-base) and builds gitops instances, invoked from `main.go`.
6. **The skill packager operates on a local filesystem directory**, not on gitea content APIs: `parseSkillDir(dir)`, `computeDirSHA(dir)`, `packageSkillDir(dir)` in `backend/internal/service/extension/skill_packager.go` + `skill_importer_*.go`. So the gitops→packager bridge (Phase 4) must **materialize** the repo tree into a temp dir (`gitops.ListTree` + `gitops.ReadFile` → write files) before calling the existing `packageDir`. Existing `InstallSource` values are `"market"`, `"github"`, `"upload"`; add `"gitops"`.
7. **`KnowledgeBaseMountSelect` lives in `clients/web/src/components/pod/CreatePodForm/KnowledgeBaseMountSelect.tsx`** (not under `components/experts/`). `ExpertSkillSlugsField` is at `clients/web/src/components/experts/ExpertSkillSlugsField.tsx`. The frontend expert API/types are in `clients/web/src/lib/api/expertApi.ts`; form model in `clients/web/src/components/experts/expertFormModel.ts`; store in `clients/web/src/stores/expert`. Experts routes live under `clients/web/src/app/(dashboard)/[org]/experts/`.
8. **KB does not expose REST file/tree endpoints today** (KB content is reached via MCP / connect paths, not `routes_*.go`). So the expert content endpoints in Phase 3 are net-new REST routes (registered in `routes_expert.go`), modeled on the KB *service* methods `ReadFile`/`ListDir` (`backend/internal/service/knowledgebase/contents.go`), not on an existing REST handler.

---

## Architecture target

```
REST v1 handlers
   ├── Expert svc  (namespace am-experts)  ─┐
   ├── Skill svc   (namespace am-skills)   ─┤─ import gitops.Service (interface, fakeable)
   └── KnowledgeBase svc (am-kb)  *unchanged this phase; keeps *gitea.Client*
                                             │
                              gitops base svc  (backend/internal/service/gitops)
                              naming • namespaces • provision+seed+cleanup • content r/w • error mapping
                                             │ wraps
                              gitea HTTP client (backend/internal/infra/gitea)  — thin transport
```

Rule: only `gitops` (and, for now, `knowledgebase`) import `internal/infra/gitea`. New `expert`/`skill` code imports `gitops`, never the raw client.

---

## Phase 1 — Gitops base service (new package, no behavior change)

**Goal:** ship the shared choke point + fake + tests. No existing service consumes it yet, so nothing can break.

### Files to create

- `backend/internal/service/gitops/gitops.go` — interface + types.
- `backend/internal/service/gitops/service.go` — gitea-backed implementation.
- `backend/internal/service/gitops/naming.go` — repo naming helpers.
- `backend/internal/service/gitops/errors.go` — `ErrNotConfigured`, `ErrNotFound`.
- `backend/internal/service/gitops/fake.go` — in-memory `Fake` for consumer tests.
- `backend/internal/service/gitops/service_test.go` — httptest-server tests of the real wrapper (mirror `knowledgebase/sync_test.go` / `provisioner_test.go`).
- `backend/internal/service/gitops/fake_test.go` — round-trip tests for the fake.
- `backend/internal/service/gitops/BUILD.bazel` — mirror `knowledgebase/BUILD.bazel` (`go_library` `gitops` with `srcs` = the new `.go` files, `importpath = github.com/l8ai-cn/agentcloud/backend/internal/service/gitops`, `deps = ["//backend/internal/infra/gitea"]`; `go_test` `gitops_test` embedding `:gitops` with `deps` = `//backend/internal/infra/gitea` + `@com_github_stretchr_testify//assert` + `//require`; and the `go_default_library` alias).

### Bazel wiring (Phase 1)

The repo is Bazel-managed with hand-maintained `BUILD.bazel` files, but gazelle is configured (root `# gazelle:prefix github.com/l8ai-cn/agentcloud`; target `//:gazelle`). Rather than hand-writing the new `BUILD.bazel`, prefer regenerating:

```bash
bazel run //:gazelle
```

This creates `backend/internal/service/gitops/BUILD.bazel` and updates `//backend/internal/infra/gitea:BUILD.bazel` if the new `HTTPError` type pulls in any new import (it does not — `fmt`/`strings` are already imported). Review the generated file against `knowledgebase/BUILD.bazel` for parity, then verify with `bazel test //backend/internal/service/gitops/... //backend/internal/infra/gitea/...`.

### Files to modify (only infra addition)

- `backend/internal/infra/gitea/client.go` — add a typed HTTP error so gitops can detect 404:

```go
// HTTPError carries the Gitea HTTP status so higher layers can branch
// (e.g. 404 -> gitops.ErrNotFound) without string matching.
type HTTPError struct {
	StatusCode int
	Method     string
	Path       string
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("gitea: %s %s → %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}
```

In `do()`, replace the `fmt.Errorf(...)` on `resp.StatusCode >= 300` with `return &HTTPError{StatusCode: resp.StatusCode, Method: method, Path: path, Body: strings.TrimSpace(string(data))}`. (Same message format, now type-inspectable. This is the ONLY change to the infra client.)

### Interface (`gitops.go`)

```go
package gitops

type Author struct{ Name, Email string }

type FileChange struct {
	Path    string
	Content []byte // bytes so binary assets (avatars) are first-class
}

type Entry struct {
	Name string
	Path string
	Type string // "file" | "dir"
	Size int64
	SHA  string
}

type Repo struct {
	Namespace     string
	Name          string // org<ID>-<slug>
	Path          string // namespace/name -> stored as git_repo_path
	DefaultBranch string
	HTTPCloneURL  string
}

type ProvisionParams struct {
	OrgID         int64
	Slug          string
	DefaultBranch string       // "" -> "main"
	CommitMessage string       // seed commit message
	Author        Author       // zero -> platform default
	Seed          []FileChange // initial files; empty repo if nil
}

type Service interface {
	Namespace() string
	EnsureNamespace(ctx context.Context) error
	Provision(ctx context.Context, p ProvisionParams) (*Repo, error)
	Commit(ctx context.Context, repoName, branch, message string, a Author, changes []FileChange) error
	ReadFile(ctx context.Context, repoName, branch, path string) ([]byte, *Entry, error)
	ListDir(ctx context.Context, repoName, branch, path string) ([]Entry, error)
	ListTree(ctx context.Context, repoName, ref string) ([]Entry, error)
	DeleteRepo(ctx context.Context, repoName string) error

	RepoName(orgID int64, slug string) string     // "org<ID>-<slug>"
	RepoPath(orgID int64, slug string) string      // "<namespace>/org<ID>-<slug>"
	RepoNameFromPath(path string) string           // strip namespace prefix

	CloneURL(repoName string) string
	CloneToken() string
}
```

### Implementation notes (`service.go`)

```go
type service struct {
	git *gitea.Client
	ns  string
	log *slog.Logger
}

// NewService returns nil when git is nil (feature-disabled convention, matches KB).
func NewService(git *gitea.Client, log *slog.Logger) Service {
	if git == nil {
		return nil
	}
	if log == nil {
		log = slog.Default()
	}
	return &service{git: git, ns: git.Namespace(), log: log.With("component", "gitops", "namespace", git.Namespace())}
}
```

- `Provision`: `EnsureNamespace` → `RepoName(orgID, slug)` → `git.CreateRepo(name, branch)` → `git.CommitFiles(name, branch, msg, author, seed→gitea.FileChange{Path, Content: string(fc.Content)}, nil)`; **on commit failure → `git.DeleteRepo(name)` and return wrapped error** (this folds `knowledgebase/provisioner.go::provisionRepo`). Return `*Repo{Namespace, Name, Path: ns+"/"+name, DefaultBranch: repo.DefaultBranch (fallback branch), HTTPCloneURL: git.CloneURL(name)}`.
- `Commit`: fold KB's SHA-probe (`knowledgebase/contents.go::CommitFile`): for each change, `git.GetFile(...)`; if found, add `isUpdate[path]=sha`; then one `git.CommitFiles(...)`. Default author `{"Agent Cloud", "platform@agentcloud.local"}` when zero.
- `ReadFile`: `git.GetFile` → `DecodedContent()`; map `*gitea.HTTPError` with `StatusCode==404` → `ErrNotFound`.
- `ListDir`/`ListTree`: wrap `git.ListDir` / `git.ListTree`, translate `*gitea.ContentEntry` / `gitea.TreeEntry` → `[]Entry`. For tree, `Type` "blob"→"file", "tree"→"dir". Map 404 → `ErrNotFound`.
- Naming (`naming.go`): `RepoName = fmt.Sprintf("org%d-%s", orgID, slug)`; `RepoPath = ns + "/" + RepoName`; `RepoNameFromPath` = substring after last `/` (copy of KB's `repoNameFromPath`).
- `errors.go`: `var ErrNotConfigured = gitea.ErrNotConfigured`; `var ErrNotFound = errors.New("gitops: not found")`.

### Fake (`fake.go`)

```go
type Fake struct {
	NS    string
	Repos map[string]*fakeRepo // repoName -> repo
	// optional failure injections for tests:
	FailProvision, FailCommit bool
}
type fakeRepo struct {
	Branch string
	Files  map[string][]byte // path -> content
	SHAs   map[string]string
}
func NewFake(ns string) *Fake { ... } // implements gitops.Service in-memory
```

- Track files per repo, compute a stable pseudo-SHA (e.g. sha1 of content) so `Commit` update-vs-create and `ReadFile` SHA are testable.
- `Provision` records seed files; honors `FailProvision`. `DeleteRepo` removes the map entry. `CloneURL`/`CloneToken` return deterministic strings.

### Wiring (deferred to consumers)

Add `backend/cmd/server`-side helper (used by Phases 3/4), e.g. in `services_init_helpers.go`:

```go
func newGiteaClientForNamespace(cfg *config.Config, namespace string) *gitea.Client {
	if !cfg.KnowledgeBase.Enabled() { return nil }
	return gitea.NewClient(gitea.Config{
		BaseURL:      cfg.KnowledgeBase.GiteaURL,
		AdminToken:   cfg.KnowledgeBase.GiteaToken,
		Namespace:    namespace,
		CloneBaseURL: cfg.KnowledgeBase.CloneBaseURL,
	})
}
```

Not yet called in Phase 1 (only defined/tested where convenient), or add in Phase 3 — either is fine; keep Phase 1 zero-behavior-change.

### Verification

```bash
cd backend && go build ./... && go test ./internal/service/gitops/... ./internal/infra/gitea/...
# Bazel (matches CI): regenerate BUILD files, then build+test via Bazel
bazel run //:gazelle
bazel test //backend/internal/service/gitops/... //backend/internal/infra/gitea/...
```

Expected: new package builds, gitops unit tests pass, gitea tests still pass (typed error is format-compatible). The module root is `doworker/go.mod` (repo root), so `cd backend && go build/test ./...` works locally but `bazel test //backend/...` is the CI source of truth — run gazelle so the new package's `BUILD.bazel` deps are correct.

---

## Phase 2 — Expert migration 000184 + domain fields

**Goal:** add git-index columns + `metadata` jsonb; no service logic yet. Independently committable and reversible.

### Files to create

- `backend/migrations/000184_experts_git_backing.up.sql`
- `backend/migrations/000184_experts_git_backing.down.sql`

### up SQL

```sql
ALTER TABLE experts ADD COLUMN git_repo_path  VARCHAR(255);
ALTER TABLE experts ADD COLUMN default_branch VARCHAR(255) NOT NULL DEFAULT 'main';
ALTER TABLE experts ADD COLUMN http_clone_url VARCHAR(1000);
ALTER TABLE experts ADD COLUMN metadata       JSONB NOT NULL DEFAULT '{}'::jsonb;

COMMENT ON COLUMN experts.git_repo_path  IS 'am-experts/org<ID>-<slug>; NULL = legacy row not yet git-backed (lazy provision on next update).';
COMMENT ON COLUMN experts.metadata       IS 'Derived cache of expert.json extras: avatar (形象, repo-relative path) + expertType (类型) + future non-column config.';
```

### down SQL

```sql
ALTER TABLE experts DROP COLUMN IF EXISTS metadata;
ALTER TABLE experts DROP COLUMN IF EXISTS http_clone_url;
ALTER TABLE experts DROP COLUMN IF EXISTS default_branch;
ALTER TABLE experts DROP COLUMN IF EXISTS git_repo_path;
```

### Files to modify

- `backend/internal/domain/expert/expert.go` — extend the `Expert` struct (keep all existing columns as cache):

```go
GitRepoPath   *string         `gorm:"size:255;column:git_repo_path" json:"git_repo_path,omitempty"`
DefaultBranch string          `gorm:"size:255;not null;default:main;column:default_branch" json:"default_branch"`
HTTPCloneURL  *string         `gorm:"size:1000;column:http_clone_url" json:"http_clone_url,omitempty"`
Metadata      json.RawMessage `gorm:"type:jsonb;not null;default:'{}';column:metadata" json:"metadata"`
```

`GitRepoPath` is a pointer so NULL legacy rows are distinguishable from `""`.

### Verification

```bash
cd backend && go build ./...
# apply + rollback the migration against a scratch/test DB using the project's migrate tooling, e.g.:
#   make migrate-up && make migrate-down   (or the repo's documented migrate command)
go test ./internal/domain/expert/... ./internal/infra/...   # repository round-trip still compiles/passes
```

---

## Phase 3 — Expert repo-backing (Git-backed lifecycle)

**Goal:** provision one repo per expert; write `agent.md`/`expert.json`/`README.md`/`assets/` on create; commit on update; source the AgentFile layer from `agent.md` on run with DB cache fallback; cleanup on delete/failure; extend REST for avatar + expert_type.

### Files to create

- `backend/internal/service/expert/gitbacking.go` — repo layout renderers + gitops glue:
  - `type expertConfig struct { Schema int; Name, Description, Avatar, ExpertType, AgentSlug, InteractionMode string; Perpetual bool; SkillSlugs []string; KnowledgeMounts []expertdom.KnowledgeMount; UsedEnvBundles []string; ConfigOverrides map[string]any; Repository *struct{RepositoryID *int64; Branch string} }` — serialized to `expert.json`.
  - `func renderExpertSeed(e *expertdom.Expert, layer string, avatar *avatarUpload) ([]gitops.FileChange, error)` — builds `agent.md` (from `layer`), `expert.json`, `README.md`, and `assets/<file>` when an avatar is supplied.
  - `func (s *Service) provisionExpertRepo(ctx, e, layer, avatar) (*gitops.Repo, error)` — wraps `gitops.Provision`.
  - `func (s *Service) ensureExpertRepo(ctx, e, layer, avatar) error` — lazy backfill: if `e.GitRepoPath == nil` and gitops enabled, Provision + set columns (used by Update/Run).
  - `func (s *Service) commitExpertChanges(ctx, e, layer, avatar) error` — re-render changed files + `gitops.Commit`.
  - `func (s *Service) readAgentFileFromGit(ctx, e) (string, bool)` — `gitops.ReadFile(repoName, branch, "agent.md")`; returns `(content, true)` or `("", false)` on miss.
- `backend/internal/service/expert/gitbacking_test.go` — inject `gitops.NewFake("am-experts")`; assert seed content, `git_repo_path`/`metadata` persisted, DB-failure→`DeleteRepo`, lazy backfill, and `agent.md`-sourced run layer.

### Files to modify

- `backend/internal/service/expert/service.go`
  - Add `gitops gitops.Service` to `Service` and `Deps`. Constructor stays; `gitops` may be nil (DB-only mode / feature disabled), matching KB's nil convention.
  - Add helpers `avatarUpload{Filename string; Data []byte}` decode + `metadata` merge (`{"avatar": "...", "expertType": "..."}`).
- `backend/internal/service/expert/crud.go`
  - `CreateExpertRequest`/`UpdateExpertRequest`: add `Avatar *AvatarInput`, `ExpertType *string` (route into `expert.json` + `metadata`, NOT new columns).
  - `Create`: after slug resolve, build layer via `buildAgentfileLayer` (generator), then **if `s.gitops != nil`**: `provisionExpertRepo(...)`, set `row.GitRepoPath/DefaultBranch/HTTPCloneURL/Metadata`. Then `store.Create`. **If `store.Create` fails and a repo was provisioned → `s.gitops.DeleteRepo(repoName)`** (mirror KB compensating cleanup). If `gitops == nil`, behave exactly as today (DB-only).
  - `Update`: apply field changes; then `ensureExpertRepo` (lazy provision if legacy) and `commitExpertChanges`. **Ordering = commit-then-update-cache** (correct for Git-authoritative): write `agent.md`/`expert.json`/`assets/` to Git first via `gitops.Commit`, then `store.Update` to refresh the derived cache columns + `metadata`. **Git is the source of truth; the DB is an index/cache that may lag on partial failure.** If the commit succeeds but `store.Update` fails, log and surface the error: Git is now ahead of the DB cache, so the cache is stale — reconcile it on the next read/update (see `Run`/`Get` cache-refresh below), do **not** treat the DB as authoritative. All gitops steps skipped when `gitops == nil` (DB-only mode). (Do not model this on KB `Update`, which is DB-only and never re-commits.)
  - `Delete`: `store.Delete` first (authoritative), then best-effort `s.gitops.DeleteRepo(RepoNameFromPath(*row.GitRepoPath))`, log on failure (mirror KB `Delete`).
- `backend/internal/service/expert/run.go`
  - In `Run`, **read is Git-first with DB fallback** (Git is source of truth): before `buildAgentfileLayer`, if `gitops != nil` and `expert.GitRepoPath != nil`, try `readAgentFileFromGit`; on success use it and **refresh the `agentfile_layer` DB cache best-effort** (this is the cache-reconcile point that heals a stale cache left by a failed `store.Update`). On miss/error (e.g. transient Gitea outage), fall back to the DB `agentfile_layer` cache / `buildAgentfileLayer` so `Run` never hard-fails. `buildAgentfileLayer` is retained as the author-time generator + legacy fallback, not as the authoritative source.
- `backend/internal/api/rest/v1/expert_handler_types.go`
  - Add to `createExpertRequest`/`updateExpertRequest`: `Avatar *avatarInput \`json:"avatar"\`` (`{ filename, content_base64 }`) and `ExpertType *string \`json:"expert_type"\``. (Base64 JSON chosen over multipart to keep the existing `ShouldBindJSON` handler flow; note this decision.)
- `backend/internal/api/rest/v1/expert_handler.go`
  - Map new fields into the service requests. `GetExpert`/`ListExperts` responses already serialize the whole `Expert` row, so `git_repo_path`, `default_branch`, `http_clone_url`, `metadata` (avatar+type) flow to the frontend automatically once the domain struct has JSON tags (Phase 2).
  - **Avatar upload validation (concrete handler step).** When `Avatar` is present in `createExpertRequest`/`updateExpertRequest`, validate **in the REST handler before calling the service** (fail fast with `400 BadRequest`, never persist an unvalidated blob):
    1. **Base64 decode** `content_base64`; reject on decode error.
    2. **Max decoded size** ≤ **2 MB** (reject with a clear message; guards against decode-bomb / oversized-blob).
    3. **MIME / type allow-list**: sniff the decoded bytes with `http.DetectContentType` and allow **only** `image/png`, `image/jpeg`, `image/webp`, `image/gif`; reject anything else. Do **not** trust the client-supplied filename or a client MIME header.
    4. **Safe fixed filename**: derive `assets/avatar.<ext>` from the sniffed type (`.png`/`.jpg`/`.webp`/`.gif`) — ignore the client filename entirely, so the repo-relative path stored in `expert.json`/`metadata` is always platform-controlled.
    The validated `{ Data []byte, Ext string }` is what the handler forwards to the service (`AvatarInput` → `avatarUpload`); `renderExpertSeed` writes it to `assets/avatar.<ext>` and records `assets/avatar.<ext>` in `expert.json`/`metadata`.
  - Add handlers `GetExpertFile` (`gitops.ReadFile`) and `GetExpertTree` (`gitops.ListTree`) — thin pass-throughs; 404→`ResourceNotFound`.
  - **Path sanitization (concrete handler step, both read routes).** The `*path` (and `tree` `path` query) params flow into `gitops.ReadFile`/`gitops.ListTree`; `gitea.escapePath` uses per-segment `url.PathEscape`, which does **not** neutralize `..`. So sanitize in the handler before the gitops call: `TrimPrefix("/")` → `path.Clean` → **reject** any result that is absolute, starts with `/`, equals `..`, or contains a `..` segment (also reject NUL/control bytes), and constrain the cleaned path to the repo tree (no escaping the repo root). Return `400 BadRequest` on rejection. Apply the identical guard to both `GetExpertFile` and `GetExpertTree`.
- `backend/internal/api/rest/v1/routes_expert.go`
  - Register `experts.GET("/:expertSlug/files/*path", h.GetExpertFile)` and `experts.GET("/:expertSlug/tree", h.GetExpertTree)` (guarded so they no-op when gitops disabled).
- `backend/cmd/server/main.go`
  - Build the expert gitops instance and pass it into `Deps`:

```go
expertGitops := gitops.NewService(newGiteaClientForNamespace(cfg, "am-experts"), appLogger.Logger)
svc.Expert = expertSvc.NewService(expertSvc.Deps{
	Store:    infra.NewExpertRepository(db),
	Pods:     services.pod,
	Dispatch: podOrchestrator,
	Repos:    services.repository,
	Gitops:   expertGitops, // nil when KB gitea not configured -> DB-only experts
	Logger:   appLogger.Logger,
})
```

### Bazel wiring (Phase 3)

This phase adds new import edges that must be reflected in the affected `BUILD.bazel` `deps`:

- `backend/internal/service/expert/BUILD.bazel` — add `//backend/internal/service/gitops` to `deps`; add the new `gitbacking.go` to the `expert` `go_library` `srcs`, and `gitbacking_test.go` to a `go_test` target (add the `expert_test` target if none exists, embedding `:expert` with `//backend/internal/service/gitops` in `deps` for the `gitops.NewFake` usage).
- `backend/internal/api/rest/v1/BUILD.bazel` — add `//backend/internal/service/gitops` to `deps` if the new `GetExpertFile`/`GetExpertTree` handlers reference `gitops` types (e.g. `gitops.Entry`, `gitops.ErrNotFound`).
- `backend/cmd/server/BUILD.bazel` — add `//backend/internal/service/gitops` to `deps` (main.go now imports it to build the per-namespace instance).

Prefer regenerating rather than editing by hand:

```bash
bazel run //:gazelle
```

Gazelle reads the Go imports and rewrites each package's `deps`/`srcs`. Manually sanity-check the three files above against the diff, then verify with `bazel test //backend/...`.

### `expert.json` shape (home of 形象 + 类型)

```jsonc
{
  "schema": 1,
  "name": "Data Analyst",
  "description": "...",
  "avatar": "assets/avatar.png",   // 形象 (repo-relative path)
  "expertType": "analysis",         // 类型
  "agentSlug": "claude-code",
  "interactionMode": "pty",
  "perpetual": false,
  "skillSlugs": ["web-search"],
  "knowledgeMounts": [{ "slug": "team-docs", "mode": "ro" }],
  "usedEnvBundles": ["default"],
  "configOverrides": { "model": "opus" },
  "repository": { "repositoryId": 12, "branch": "main" }
}
```

Repo layout: `/agent.md`, `/expert.json`, `/README.md`, `/assets/`.

### Verification

```bash
cd backend && go build ./... && go test ./internal/service/expert/... ./internal/api/rest/v1/...
# Bazel (matches CI): regenerate deps, then build+test
bazel run //:gazelle
bazel test //backend/internal/service/expert/... //backend/internal/api/rest/v1/... //backend/cmd/server/...
```

Expected: expert tests (with `gitops.Fake`) assert seed files, metadata persistence, compensating delete, lazy backfill, agent.md-sourced run, and nil-gitops DB-only path. `bazel run //:gazelle` must leave the `expert`, `api/rest/v1`, and `cmd/server` `BUILD.bazel` `deps` referencing `//backend/internal/service/gitops`.

---

## Phase 4 — Skill service (Git-backed, additive)

**Goal:** a new independent, gitops-backed authoring source for skills (`am-skills`) that bridges into the existing packager→object-storage→`SkillMarketItem` install pipeline. The existing external-import/marketplace flow is untouched.

### Files to create

- `backend/internal/service/skill/service.go` — new package (peer to `expert`), holds `gitops gitops.Service` bound to `am-skills`, plus a bridge to the existing `extension` packager/install. Constructor returns nil when `gitops == nil`.
  - `Deps{ Gitops gitops.Service; Packager SkillPackagerBridge; Market extension.Repository; Logger *slog.Logger }`.
- `backend/internal/service/skill/authoring.go`
  - `Create(ctx, CreateSkillRequest) (*Skill, error)`: render `SKILL.md` (same frontmatter shape `parseSkillDir`/`parseFrontmatter` understands: `name`, `description`, `license`, ...) + `skill.json` → `gitops.Provision("am-skills", org<ID>-<slug>, seed)`; then bridge-package + index as `SkillMarketItem` with `InstallSource="gitops"`.
  - `Update(ctx, ...)`: `gitops.Commit` new `SKILL.md`/`skill.json`, re-package, bump market-item version.
  - `Delete`: best-effort `gitops.DeleteRepo`.
- `backend/internal/service/skill/materialize.go`
  - `func materializeRepo(ctx, g gitops.Service, repoName, ref string) (dir string, cleanup func(), err error)` — `gitops.ListTree` + `gitops.ReadFile` → write into `os.MkdirTemp`, so the existing local-dir packager (`packageDir`/`parseSkillDir`/`computeDirSHA`) consumes gitops content unchanged. **(This is the key bridge; the packager is filesystem-based, not gitea-API-based.)**
- `backend/internal/service/skill/*_test.go` — inject `gitops.Fake("am-skills")` + a fake packager bridge; assert `SKILL.md`/`skill.json` seed round-trips into the packager, `InstallSource="gitops"`, and nil-gitops disabled path.

### Files to modify (extension bridge)

- `backend/internal/service/extension/skill_packager.go` (or a new `skill_packager_dir.go`) — extract/expose a `PackageFromDir(ctx, dir string) (*PackagedSkill, error)` that just calls the existing private `packageDir`, so the skill service can package a materialized gitops dir without re-cloning. (Minimal, non-breaking; existing `PackageFromGitHub`/`PackageFromUpload` unchanged.)
- `backend/internal/service/extension/skill_packager_install.go` / `service_install.go` — accept/record `InstallSource="gitops"` (extend the existing `"market"|"github"|"upload"` set). No change to how experts reference skills.
- `backend/cmd/server/main.go` — construct `skill.NewService(skill.Deps{ Gitops: gitops.NewService(newGiteaClientForNamespace(cfg, "am-skills"), log), Packager: skillPkg (PackageFromDir), Market: extRepo, Logger: log })` and attach to the services container.
- `backend/internal/api/rest/v1/routes_skill.go` (new) + handler — CRUD for authored skills; register in the v1 router alongside expert routes. No-op when the skill service is nil.

### Bazel wiring (Phase 4)

New package + new import edges to reflect in `BUILD.bazel`:

- `backend/internal/service/skill/BUILD.bazel` — **new** `go_library` `skill` (`srcs` = `service.go`/`authoring.go`/`materialize.go`, `importpath = .../backend/internal/service/skill`, `deps` = `//backend/internal/service/gitops` + `//backend/internal/service/extension`) + `go_test` `skill_test` embedding `:skill` (deps include `//backend/internal/service/gitops` for `gitops.Fake` and testify) + `go_default_library` alias.
- `backend/internal/service/extension/BUILD.bazel` — add the new/exported `PackageFromDir` source file to `srcs` if a new file is introduced (`skill_packager_dir.go`); no new external dep expected.
- `backend/internal/api/rest/v1/BUILD.bazel` — add the new `routes_skill.go`/handler to `srcs` and `//backend/internal/service/skill` to `deps`.
- `backend/cmd/server/BUILD.bazel` — add `//backend/internal/service/skill` to `deps` (main.go builds the skill service).

Regenerate + verify:

```bash
bazel run //:gazelle
bazel test //backend/internal/service/skill/... //backend/internal/service/extension/... //backend/internal/api/rest/v1/... //backend/cmd/server/...
```

### How experts reference skills (unchanged)

Experts reference skills **by slug** (`expert.SkillSlugs` → `SKILLS` AgentFile directive in `buildAgentfileLayer`). Whether a skill's source is external-clone or gitops-backed is invisible to experts — resolved through the existing extension install layer at run time.

### Verification

```bash
cd backend && go build ./... && go test ./internal/service/skill/... ./internal/service/extension/...
# Bazel (matches CI): regenerate deps for the new package + edges, then build+test
bazel run //:gazelle
bazel test //backend/internal/service/skill/... //backend/internal/service/extension/... //backend/internal/api/rest/v1/... //backend/cmd/server/...
```

Expected: skill authoring tests pass with the fake; extension packager/install tests still pass; a gitops-authored skill packages + indexes identically to an external import. Confirm the new `skill` package `BUILD.bazel` exists after gazelle and that `extension`/`api/rest/v1`/`cmd/server` deps were updated.

---

## Phase 5 — Frontend: create-expert page

**Goal:** a dedicated create-expert page (keep the existing drawer for now), surfacing avatar (形象) + expert type (类型) and reusing existing field components.

### Files to create

- `clients/web/src/app/(dashboard)/[org]/experts/new/page.tsx` — the new create page, sectioned:
  - **Section 0 — Basics:** name, description, avatar upload (→ base64 into `avatar`), `expertType` (类型), `agentSlug`, slug (auto from name via existing `slugifyExpert`).
  - **Section 1 — agent.md editor:** a textarea/markdown editor bound to `agentfileLayer` (the authored `agent.md`).
  - **Section 2 — Skill mount:** reuse `clients/web/src/components/experts/ExpertSkillSlugsField.tsx`.
  - **Section 3 — Knowledge mount:** reuse `clients/web/src/components/pod/CreatePodForm/KnowledgeBaseMountSelect.tsx`.
- `clients/web/src/components/experts/CreateExpertForm.tsx` — the form component the page renders (extract shared bits from `ExpertEditDrawer.tsx`/`expertFormModel.ts` where reasonable).

### Files to modify

- `clients/web/src/lib/api/expertApi.ts`
  - Extend `Expert` with `git_repo_path?: string | null`, `default_branch?: string`, `http_clone_url?: string | null`, `metadata?: { avatar?: string; expertType?: string } & Record<string, unknown>`.
  - Extend `CreateExpertInput`/`UpdateExpertInput` with `avatar?: { filename: string; content_base64: string }` and `expert_type?: string`.
- `clients/web/src/components/experts/expertFormModel.ts` — add `avatar`/`expertType` to `ExpertFormState` + `buildExpertConfig`.
- `clients/web/src/app/(dashboard)/[org]/experts/page.tsx` — make the "Create" button link to `/[org]/experts/new` (keep the drawer available too, per decision "keep existing drawer for now").
- **i18n message keys (concrete step).** The expert UI uses `next-intl` (`useTranslations("experts.edit")` in `ExpertEditDrawer.tsx`); every user-facing string on the new create page must be a message key, not a literal. Add new keys under the existing `experts` object (follow the existing `experts.*` structure, e.g. a new `experts.create.*` group alongside `experts.edit.*`) for: page title/subtitle, the four section headings (Basics / agent.md editor / Skill mount / Knowledge mount), field labels + placeholders (name, description, avatar upload, expert type, agentSlug, slug, agent.md), avatar validation error messages (too large / unsupported type), and submit/cancel/loading actions. Add the keys to **all** locale files, not just English:
  - `clients/web/src/messages/en/experts.json` (author the canonical English copy)
  - and the same keys in `clients/web/src/messages/{zh,ja,ko,pt,fr,es,de}/experts.json` (8 locales total). Keep the key paths identical across locales so `next-intl` does not fall back / warn; translate values (or stage English placeholders if translations lag, tracked as follow-up).

### Verification

```bash
cd clients/web && pnpm typecheck && pnpm test -- experts   # or the repo's documented web test command
pnpm build
```

Expected: type-checks, expert-related unit tests pass, page builds and renders the four sections, avatar preview + expertType round-trip.

---

## Testing strategy

- **`gitops.Service` is an interface with `gitops.Fake`** (in-memory `repoName -> path -> []byte`, with branch + pseudo-SHA bookkeeping). Expert and Skill unit tests inject the fake, so they run with **no live Gitea and no network**. This is the primary payoff of the interface boundary.
- **gitops package tests** use the existing `httptest.Server` pattern already proven in `backend/internal/service/knowledgebase/sync_test.go` / `provisioner_test.go` (`gitea.NewClient(gitea.Config{BaseURL: srv.URL})`) to exercise the real gitea wrapper + Provision/cleanup against a fake Gitea HTTP server. Add one test that returns 404 to assert `ReadFile` maps to `gitops.ErrNotFound` (validates the new typed `gitea.HTTPError`).
- **Expert tests:** seed-file content (`agent.md`/`expert.json`/`assets/avatar.*`), `git_repo_path`/`metadata` persisted, DB-`Create` failure → `DeleteRepo` invoked, lazy backfill on `Update`, `Run` sources layer from `agent.md` (with DB-cache fallback), and **nil-gitops DB-only mode** behaves exactly as today.
- **Skill tests:** `SKILL.md`/`skill.json` seed materializes into a temp dir and round-trips through the existing packager; `InstallSource="gitops"`; nil-gitops disabled path.
- **No live-Gitea dependency** is introduced into expert/skill test suites.

---

## Rollout / backfill

- **Lazy provisioning (chosen).** Legacy experts (rows with `git_repo_path IS NULL`) are **not** touched by a migration job. On the next `Update` or `Run`, `ensureExpertRepo` provisions the repo from current DB columns (via `buildAgentfileLayer` generator → `agent.md` + `expert.json`) and backfills `git_repo_path`/`default_branch`/`http_clone_url`/`metadata`. Until then, `Run` uses the DB-column path (unchanged behavior). This avoids a big-bang backfill and keeps the change incremental and low-risk.
- **Feature-flag by configuration.** When the internal Gitea is not configured (`cfg.KnowledgeBase.Enabled() == false`), `gitops.NewService` returns nil and expert/skill services run in **DB-only mode** — identical to today. No forced dependency on Gitea.
- **A one-time backfill job is explicitly out of scope** for this phase (can be added later as an admin command if desired).

---

## Risks

1. **Git is source of truth; DB is an index/cache that can lag.** `Git` (`agent.md`/`expert.json`/`assets/`) is authoritative; the `experts` columns + `metadata` are a derived cache. Writes are **commit-then-update-cache**; reads are **Git-first with DB fallback**. On a partial failure (commit succeeds, `store.Update` fails), the DB cache is stale and Git is ahead — the next `Run`/`Get` uses fresh Git content and reconciles the cache, while a pure cache read (`List`/detail) may briefly render stale data. This is real, bounded drift (not "benign"). Mitigation: single write path in `commitExpertChanges` + `Update`; refresh `agentfile_layer`/cache columns on the next Git-first read; surface the DB-update error rather than silently trusting the cache. Do **not** treat the DB as authoritative on conflict (this differs from KB, whose `Update` is DB-only and never touches Git).
2. **Typed-error infra change.** Replacing the `fmt.Errorf` in `gitea.do()` with `*HTTPError` could surprise any code that string-parses gitea errors. Mitigation: `HTTPError.Error()` preserves the exact prior message format; grep for gitea error string-matching before merging Phase 1.
3. **Namespace/org bootstrapping.** `am-experts` / `am-skills` orgs must be creatable by the admin token. `Provision` calls `EnsureNamespace` (idempotent) first, so first-write creates them — but a token lacking org-create scope will fail. Mitigation: surface a clear provisioning error and document the token requirement.
4. **Avatar bytes in Git.** Committing binaries into per-expert repos grows repo size over time (each avatar change = a blob). Acceptable per locked decision. Mitigation is now a **concrete Phase 3 handler step** (see `expert_handler.go` → "Avatar upload validation"): ≤ 2 MB decoded size cap, MIME allow-list via magic-byte sniffing, and a platform-controlled `assets/avatar.<ext>` filename.
5. **Packager is filesystem-based.** The skill bridge must materialize gitops content to a temp dir; large skills mean disk I/O and temp-dir cleanup. Mitigation: `materializeRepo` returns a `cleanup func()` (deferred `os.RemoveAll`), bound tree size via `ListTree`.
6. **Run latency / availability.** Sourcing `agent.md` from Git on every `Run` adds a network hop and a Gitea-outage failure mode. Mitigation: cache into `agentfile_layer` and fall back to the DB cache on read error, so `Run` never hard-fails on a transient Gitea issue.
7. **Config coupling.** All three domains currently share `cfg.KnowledgeBase.*` (URL/token/clone-base). Mitigation acceptable for now (one Gitea instance, per-namespace clients); a dedicated `cfg.Gitea` block is future cleanup, not required here.
8. **Migration ordering.** If another branch also claims `000184`, there will be a collision. Mitigation: confirm `000184` is unused at merge time (currently the max on disk is `000183` — `000181`/`000182`/`000183` are already taken, so `000184` is the first free number).

---

## Commit sequence summary

| Phase | Scope | Independently shippable |
|-------|-------|-------------------------|
| 1 | `gitops` package + `Fake` + tests; typed `gitea.HTTPError` | Yes — no consumers yet |
| 2 | Migration 000184 + domain struct fields | Yes — additive columns, reversible |
| 3 | Expert repo-backing (create/update/run/delete) + REST + wiring | Yes — nil-gitops = today's behavior |
| 4 | Git-backed Skill service (additive) + packager bridge | Yes — coexists with existing skill flow |
| 5 | Frontend create-expert page (avatar/type/agent.md/skills/knowledge) | Yes — drawer retained |

---

## Review (2026-07-08)

Independent, read-only technical review against the actual codebase. Every code fact
below was checked against the real source files.

### Verdict

**Architecturally sound; the phasing is correct and the code facts are ~95% accurate.**
There is **one Blocker** (a wrong migration number that will collide on first run) and a
handful of Should-fix gaps around the Bazel build, source-of-truth wording, and unspecified
input validation. **Phase 1 (gitops base package + typed `gitea.HTTPError`) is safe to
start as written** — it touches no migrations and its stated infra facts are all confirmed.
Do **not** start Phase 2 until the migration number is corrected.

> **Applied review fixes on 2026-07-08** — the Blocker (migration → `000184`) and all 5 Should-fix items (Bazel wiring, source-of-truth wording, avatar validation, path traversal, i18n keys) are now reflected in the plan body above. The issue list below is kept intact for traceability.

### Confirmed code facts (plan is accurate on these)

- **`gitea.FileChange.Content` is `string`** — `backend/internal/infra/gitea/contents.go:11`.
  `CommitFiles` base64-encodes `[]byte(ch.Content)` (line 32), so the gitops `[]byte →
  string(fc.Content)` conversion is lossless for binary. ✔
- **No `gitea.ErrNotFound`; only `gitea.ErrNotConfigured`** (`client.go:20`). `do()` returns
  a formatted `fmt.Errorf("gitea: %s %s → %d: ...")` for any status ≥ 300 (`client.go:77-80`).
  The proposed typed `HTTPError{StatusCode, Method, Path, Body}` is correct and `do()` already
  has `method`/`path` in scope plus `fmt`/`strings` imported — the change compiles as described. ✔
- **All gitea method signatures match the spec**: `EnsureNamespace(ctx)`,
  `CreateRepo(ctx,name,defaultBranch)(*Repo,error)`, `DeleteRepo(ctx,name)`,
  `CommitFiles(ctx,repo,branch,message,CommitAuthor,[]FileChange,isUpdate map[string]string)`,
  `GetFile(ctx,repo,branch,path)`, `ListDir(ctx,repo,branch,path)`, `ListTree(ctx,repo,ref)`,
  `Namespace()`, `CloneURL(name)`, `CloneToken()`. Namespace is baked into `Config`
  (`client.go:22-27`), so one client per namespace is indeed the smaller change. ✔
- **KB provisioner pattern is exactly as described**: `provisioner.go::provisionRepo`
  (EnsureNamespace → `org%d-%s` → CreateRepo → CommitFiles → DeleteRepo-on-failure),
  `service.go::Create` compensating `DeleteRepo` on DB failure (line 103), `Delete`
  DB-row-first then best-effort `DeleteRepo` (lines 169-185), `repoNameFromPath` (line 187),
  `CommitFile` SHA-probe (`contents.go:58-76`). The gitops `Provision`/`Commit` folds are faithful. ✔
- **Expert service is pure-DB today** and wired in `backend/cmd/server/main.go`
  (`svc.Expert = expertSvc.NewService(...)` at **line 212**, plan said ~208 — close enough),
  **not** in `services_init_helpers.go`. The KB gitea client is built in
  `services_init_helpers.go::initializeKnowledgeBaseService` and is not returned, so a new
  per-namespace helper is genuinely needed. ✔
- **Config fields exist**: `cfg.KnowledgeBase.{GiteaURL,GiteaToken,GiteaOrg,CloneBaseURL}` +
  `Enabled()` (`config_infra.go:51-62`, `config.go:187-189`). The `newGiteaClientForNamespace`
  helper as written is valid (`config.Config` is the right type). ✔
- **Skill packager is filesystem-based**: `packageDir` → `parseSkillDir` / `computeDirSHA` /
  `packageSkillDir` (`skill_packager.go:103-136`); `PackageFromGitHub`/`PackageFromUpload`
  both materialize to `os.MkdirTemp` first. The gitops→temp-dir bridge is required, as claimed. ✔
- **`InstallSource` values are `market`/`github`/`upload`** and `install_source` is
  `VARCHAR(20) NOT NULL` with **only a comment, no CHECK constraint**
  (`000060_add_extension_market.up.sql:158`) — so adding `"gitops"` needs **no migration**. ✔
- **Frontend components exist and are reusable**: `components/experts/ExpertSkillSlugsField.tsx`,
  `components/pod/CreatePodForm/KnowledgeBaseMountSelect.tsx`, `components/experts/expertFormModel.ts`,
  `lib/api/expertApi.ts`, route dir `app/(dashboard)/[org]/experts/` (no `new/` yet). ✔
- **Proposed new expert columns do not exist** (`000178_experts.up.sql` has no
  `git_repo_path`/`default_branch`/`http_clone_url`/`metadata`; domain struct confirms). ✔

### Prioritized issues

#### Blocker

- **B1 — Migration `000181` is already taken (collision on migrate-up).**
  The plan's Findings #1, Phase 2, Risk #8, and the commit-summary table all assert "latest
  migration is `000180`, next is `000181`." **False.** On disk today:
  `000181_im_channel_bridges`, `000182_im_weixin_support`, `000183_pod_preview` all exist —
  the current max is **`000183`**. Using `000181` will collide / fail to apply.
  **Fix:** rename the new migration to **`000184_experts_git_backing.{up,down}.sql`** (verify
  `000184` is still free at merge time) and update every reference to `000181` in Phase 2, the
  down-SQL filenames, Risk #8, and the commit-sequence table. (Phase 1 is unaffected.)

#### Should-fix

- **S1 — Bazel BUILD wiring is under-specified.** The repo is Bazel-managed
  (`WORKSPACE.bazel` + `MODULE.bazel`, `rules_go`, hand-maintained `BUILD.bazel` with explicit
  `srcs`/`deps`, e.g. `knowledgebase/BUILD.bazel`). The plan creates `gitops/BUILD.bazel` but
  does **not** mention updating the dependent `BUILD.bazel` files that gain a new import edge:
  `expert`, `api/rest/v1`, `cmd/server` (Phase 3), and the new `skill` package + `extension`
  (Phase 4). Each needs its `deps` list updated (or `gazelle` run).
  **Fix:** add a per-phase step "update/regenerate affected `BUILD.bazel` deps (or run gazelle)"
  and state the canonical verify command (likely `bazel test //backend/...`). Note the module
  root is `doworker/go.mod` (repo root), so `cd backend && go build/test ./...` works locally
  but may not match CI.

- **S2 — Source-of-truth wording contradicts the locked decision.** The Error-Handling table
  and Update flow say "DB row is authoritative … Git has a benign extra commit" (copied from KB,
  where `Update` is DB-only and never re-commits). But here the decision is **Git-authoritative**
  and Phase 3's `run.go` reads `agent.md` from Git first. So a *commit-succeeds / DB-update-fails*
  outcome leaves Git ahead of the DB cache: the next `Run` uses new Git content while `List`/detail
  render the stale cache — real drift, not "benign."
  **Fix:** reframe as "Git is source of truth; DB is a cache that may lag on partial failure;
  reconcile the cache on next read/update." Keep commit-first ordering (correct for Git-authoritative),
  but describe cache-refresh/backfill on read rather than calling the DB authoritative. Also drop
  the "mirrors KB Update" claim (KB `Update` does not touch Git).

- **S3 — Avatar upload validation is only in Risks, not an actual step.** Risk #4 mentions a size
  cap and content-type constraint, but Phase 3's handler steps don't implement them. Base64 into
  `assets/` without limits is a decode-bomb / oversized-blob vector.
  **Fix:** make it a concrete Phase 3 handler step: enforce max decoded size (e.g. ≤ 1–2 MB),
  allow-list content types / sniff magic bytes, and derive a safe fixed filename (`assets/avatar.<ext>`)
  rather than trusting the client-supplied filename.

- **S4 — Path-traversal handling on the new read routes is unspecified.** Phase 3 adds
  `GET /experts/:expertSlug/files/*path` and `/tree` passing `*path` into `gitops.ReadFile`.
  `gitea.escapePath` uses `url.PathEscape` per segment, which does **not** neutralize `..`
  segments. The blast radius is bounded by the contents API (repo-scoped), but the plan should
  still `path.Clean` + reject `..`/absolute paths at the handler.
  **Fix:** add an explicit path-sanitization step for both new routes (and mirror it if KB-style
  `TrimPrefix("/")` is reused).

- **S5 — i18n strings for the new page are missing from Phase 5.** The expert UI uses
  `next-intl` (`useTranslations("experts.edit")` in `ExpertEditDrawer.tsx`). A new create page
  with avatar/type/agent.md sections needs new message keys across all locale files.
  **Fix:** add a Phase 5 step to add the `experts.*` message keys to every `messages/*.json`
  locale, not just the English strings.

#### Nice-to-have

- **N1 — Skill package storage key mismatch.** Spec §5.3 wants `skills/gitops/{slug}/{sha}.tar.gz`,
  but Phase 4 reuses `packageDir`, which hard-codes `skills/direct/%s/%s.tar.gz`
  (`skill_packager.go:119`). Either accept `skills/direct/` for gitops-authored skills or add a
  keyed variant; state which.
- **N2 — Update commit concurrency.** Two concurrent `Update`s SHA-probe then commit; the second
  can fail on a stale SHA (Gitea rejects). Not fatal (returns an error), but the plan should note
  retry-on-conflict or last-writer semantics.
- **N3 — `Run` Git-read fallback depends on a populated cache.** The fallback path
  (`agentfile_layer` DB cache) only works if it was written on create/update; ensure the seed/commit
  flow always refreshes `agentfile_layer` so the transient-Gitea-outage fallback in Risk #6 is real.
- **N4 — `expertFormModel.ts` vs `useExpertEditForm`.** The existing drawer composes state via
  `useExpertEditForm.ts` (in addition to `expertFormModel.ts`); Phase 5's "extract shared bits"
  step should account for both to avoid duplicating form logic.

### Corrections to the plan's stated code facts

| Plan statement | Reality | Action |
|----------------|---------|--------|
| "Latest migration is `000180`; next is `000181`" (Findings #1, Phase 2, Risk #8) | Max on disk is **`000183`** (`000181`/`182`/`183` exist) | Use **`000184`** (Blocker B1) |
| Expert wired "around line 208" of `main.go` | Actually **line 212** | Cosmetic only |
| Everything else checked (gitea signatures, `FileChange.Content string`, no `ErrNotFound`, KB pattern, packager filesystem-basis, `InstallSource` set, config fields, frontend components, absent expert columns) | **Accurate** | None |

