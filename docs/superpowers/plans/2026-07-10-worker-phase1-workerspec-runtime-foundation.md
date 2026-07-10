# Worker Phase 1 WorkerSpec Runtime Foundation

## Scope

Land the first main-branch atomic commit for Worker creation redesign without
touching Pod, Proto, Web creation, credential EnvBundle, mobile access, parser,
ACP, or AgentPod lifecycle files.

## Commit 1 Deliverable

- WorkerSpec V1 domain contract with immutable runtime image digest, positive
  model resource ID, Worker type definition hash, placement, type config,
  workspace, lifecycle, metadata, codec, summary, and snapshot.
- Runtime resource resolver that returns canonical image and placement from
  scoped repository lookups and compatibility checks.
- Immutable `worker_spec_snapshots` migration `000194`.
- Focused contract tests for WorkerSpec, runtime resolution, snapshot
  repository, and migration shape.

## Verification

- `go test ./backend/internal/domain/workerspec`
- `go test ./backend/internal/service/workerruntime`
- `go test ./backend/internal/infra -run TestWorkerSpecSnapshotRepository`
- `go test ./backend/migrations -run TestMigration000194`
- Real PostgreSQL migrate up/down contract for `000194`.
- `go build ./backend/cmd/server`
- `git diff --check`
- Diff review confirms no Bazel files, no Task 7 files, and no unrelated dirty
  files are staged.
