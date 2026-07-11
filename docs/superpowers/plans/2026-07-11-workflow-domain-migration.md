# Workflow Domain Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` or `executing-plans` task by task.

**Goal:** Replace the current scheduled Loop contract with a clean-cut Workflow contract across storage, backend, Runner MCP, Rust Core, Web, and documentation.

**Architecture:** The current scheduled execution engine keeps its behavior but moves from `loop` to `workflow` names. A database migration renames the live tables and database objects. All generated and handwritten API contracts use `workflow`; no `/loops` scheduling alias remains.

**Tech Stack:** PostgreSQL migrations, Go, Connect RPC, protobuf, Rust/WASM, Next.js, TypeScript, Vitest, Playwright.

---

### Task 1: Database Workflow Storage

**Files:**
- Create: `backend/migrations/000200_rename_loops_to_workflows.up.sql`
- Create: `backend/migrations/000200_rename_loops_to_workflows.down.sql`
- Create: `backend/migrations/workflow_rename_test.go`

- [ ] Write a failing migration test that requires `workflows`, `workflow_runs`,
  workflow-named indexes, and renamed sequences after the up migration.
- [ ] Run `go test ./backend/migrations -run TestMigration000200`.
- [ ] Rename tables, indexes, constraints, and sequences in up/down SQL.
- [ ] Re-run the migration test and `go test ./backend/migrations`.

### Task 2: Workflow Backend Contract

**Files:**
- Move: `backend/internal/domain/loop/` -> `backend/internal/domain/workflow/`
- Move: `backend/internal/service/loop/` -> `backend/internal/service/workflow/`
- Move: `backend/internal/api/connect/loop/` -> `backend/internal/api/connect/workflow/`
- Move: `proto/loop/v1/loop.proto` -> `proto/workflow/v1/workflow.proto`
- Move: `proto/loop_state/v1/loop_state.proto` -> `proto/workflow_state/v1/workflow_state.proto`
- Modify: `backend/internal/infra/loop*_repo*.go`
- Modify: server initialization, REST handlers, event names, API scopes, and tests

- [ ] Write failing compile-oriented tests for `WorkflowService`, `Workflow`,
  `WorkflowRun`, `workflow_run:*`, and `workflows:*`.
- [ ] Move and rename the scheduled domain, preserving cron claim, atomic trigger,
  Pod-status SSOT, cancellation, callback, and run statistics behavior.
- [ ] Update SQL table references and server wiring.
- [ ] Regenerate proto outputs with `pnpm proto:gen-go-all`.
- [ ] Run focused Go tests for workflow domain, service, infra, Connect, REST, and MCP adapters.

### Task 3: Runner and MCP Workflow Tools

**Files:**
- Move: Runner loop MCP client/tool files to workflow names
- Move: backend runner MCP loop adapters to workflow names
- Modify: MCP tool registry, protobuf imports, and MCP E2E suites

- [ ] Write failing MCP tool assertions for `list_workflows`, `create_workflow`,
  and `trigger_workflow`.
- [ ] Rename scheduled MCP tools and their backend adapters.
- [ ] Update callers and generated Workflow service references.
- [ ] Run focused Runner and `tests/mcp-e2e` workflow suites.

### Task 4: Rust and Web Workflow Clients

**Files:**
- Move: Rust loop state/service/proto modules to workflow names
- Move: `clients/web/src/components/loops/` -> `components/workflows/`
- Move: dashboard and docs `/loops` routes -> `/workflows`
- Move: web loop store, Connect adapter, projections, and view models
- Modify: navigation, API-key scopes, i18n, E2E tests, public docs, and sitemap

- [ ] Write failing TypeScript and Rust contract tests for Workflow labels, paths,
  scopes, and generated service calls.
- [ ] Rename scheduled client state and routes without touching `loopal`.
- [ ] Regenerate proto output and build WASM.
- [ ] Run Web unit tests, Rust workspace tests, lint, typecheck, build, and targeted Playwright workflow paths.

### Task 5: Documentation and Release Validation

**Files:**
- Modify: API docs, feature docs, tutorials, i18n descriptions, blog references,
  and the domain-split design document where implementation detail changed

- [ ] Update visible documentation to define Worker, Loop, Workflow, and Run.
- [ ] Remove scheduled-task claims from Loop documentation and old `/loops` API examples.
- [ ] Verify no scheduled `LoopService`, `loops:*`, `/loops`, or `loop_run:*`
  references remain outside the new goal Loop implementation and historical migrations.
- [ ] Run `git diff --check`, targeted test suites, and browser verification.
