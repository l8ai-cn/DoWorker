# Worker Phase 0D Connect Generated-Source Repair Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Restore the backend server build by removing duplicate literal declarations of Connect generated Go files without changing generated targets or application behavior.

**Architecture:** `amesh_proto_convert` owns each `*.amesh.go` output. The consuming `go_library` references only the generated target label; listing the output filename as an ordinary source asks Bazel for a nonexistent source-tree file.

**Tech Stack:** Bazel, rules_go, Starlark BUILD files.

### Task 1: Remove Duplicate Generated Source Literals

**Files:** Modify only the affected `BUILD.bazel` files below.

- `backend/internal/api/connect/admin/BUILD.bazel`
- `backend/internal/api/connect/admin/promocode/BUILD.bazel`
- `backend/internal/api/connect/admin/subscription/BUILD.bazel`
- `backend/internal/api/connect/admin/support_ticket/BUILD.bazel`
- `backend/internal/api/connect/agent/BUILD.bazel`
- `backend/internal/api/connect/agentpod_settings/BUILD.bazel`
- `backend/internal/api/connect/binding/BUILD.bazel`
- `backend/internal/api/connect/channel/BUILD.bazel`
- `backend/internal/api/connect/env_bundle/BUILD.bazel`
- `backend/internal/api/connect/license/BUILD.bazel`
- `backend/internal/api/connect/mesh/BUILD.bazel`
- `backend/internal/api/connect/runner/BUILD.bazel`
- `backend/internal/api/connect/support_ticket/BUILD.bazel`
- `backend/internal/api/connect/user/BUILD.bazel`
- `backend/internal/api/connect/user_credential/BUILD.bazel`

- [ ] **Step 1: Preserve the failing build evidence**

Run `bazel build //backend/cmd/server:server --verbose_failures`. Expected RED: each affected package reports a missing input such as `user_convert.amesh.go` even though its `amesh_proto_convert` target exists.

- [ ] **Step 2: Remove only duplicate literals**

In each `go_library.srcs`, remove the literal `"*_convert.amesh.go"`. Keep the `amesh_proto_convert(output = ...)` declaration and its `:..._amesh` source label unchanged. Do not regenerate, reformat, or edit Go code.

- [ ] **Step 3: Verify generated-source invariants**

Confirm every affected BUILD file contains exactly one output declaration and the existing generated target label, with no generated filename literal remaining in `go_library.srcs`.

- [ ] **Step 4: Verify GREEN**

Run:

```bash
bazel build //backend/cmd/server:server --verbose_failures
bazel run //:buildifier_check
git diff --check
```

Expected: all pass, and the diff consists only of the 15 literal-line deletions.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/api/connect/**/BUILD.bazel
git commit -m "fix(bazel): restore Connect generated sources"
```
