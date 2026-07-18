# Named Definition Config Binding Contract

Status: approved and implemented; release remains gated on historical-data repair
and real Worker lifecycle evidence.

## Root Cause

`config_bundle_ids` and `configBundleRefs` preserve only anonymous bundle
identity. They lose the Definition document ID before compilation.

The current AgentFile evaluator receives config documents by bundle name.
`USE_CONFIG_BUNDLE` sets one shared `config_json`, silently skips a missing
bundle, and the final declaration wins. This cannot prove that a selected
document belongs to `settings`, `openclaw-json`, or a future document.

Runner is not the source of the problem. Backend compilation must resolve
Definition ownership and produce exact `FilesToCreate`; Runner only materializes
those already-resolved files inside the sandbox.

## Decision

Replace positional configuration references atomically with:

```text
config_document_bindings[{ document_id, config_bundle_id }]
```

Do not read, write, or compile `config_bundle_ids` after the cutover. Do not
retain the legacy direct Pod form as a compatibility path.

## Public Contract

Add these protobuf messages:

```proto
message WorkerConfigDocumentBinding {
  string document_id = 1;
  int64 config_bundle_id = 2;
}

message WorkerCredentialRequirement {
  string id = 1;
  string source_kind = 2;
  string source_ref = 3;
  string target_kind = 4;
  string target_name = 5;
}

message WorkerConfigDocumentRequirement {
  string document_id = 1;
  string format = 2;
  string target_path = 3;
}
```

`WorkerSpecDraft` gains
`repeated WorkerConfigDocumentBinding config_document_bindings = 28`.
Field 27 remains reserved in the public schema and is never read as a
compatibility input.

`WorkerTypeOption` gains redacted repeated credential and configuration-document
requirements. The response contains identifiers, formats, targets, source
kinds, and target environment names only. It never contains a secret value.

`WorkerTemplateWorkspaceSpec` replaces `ConfigBundleRefs` with explicit
`ConfigDocumentBindings`. Rust Core remains binary transport SSOT: regenerate
the protobuf types and retain no parallel application-side state model.

## Backend Resolution

For every selected Worker Definition:

1. Require every declared configuration document exactly once.
2. Reject unknown document IDs, duplicate IDs, duplicate bundle IDs, wrong
   bundle kind, inaccessible bundles, malformed JSON, and missing bindings.
3. Resolve selected bundle IDs with an exact-ID loader. Do not load all visible
   config bundles and do not key compilation by a mutable bundle name.
4. Build the AgentFile config context as `document_id -> parsed JSON`.
5. Emit `USE_CONFIG_BUNDLE "<document_id>"` from the compiler.
6. Snapshot document ID, format, Definition hash, bundle ID, bundle revision,
   and content hash with the WorkerSpec.

The existing current Workers have at most one Definition-owned document. A
future Definition with multiple documents must be rejected until AgentFile has
explicit named-document evaluation semantics; `config_json` last-wins behavior
must not be presented as multi-document support.

## Target Ownership

`target_path` is Definition metadata and an ownership assertion, not a path
that Runner interprets. AgentFile is the only path-materialization authority.

- Do Agent and Seedance AgentFiles write `config_json` to their settings file,
  while `DO_AGENT_SETTINGS` points at that resolved file.
- OpenClaw AgentFile writes `config_json` to
  `openclaw-home/.openclaw/openclaw.json`.
- Compiler tests must evaluate the actual AgentFile and assert resulting
  `FilesToCreate` paths and JSON. Runner tests assert path confinement and
  file creation only.

## Migration

Migrate WorkerSpec snapshots and WorkerTemplate resources in one transaction.

- A snapshot with no declared Definition documents must have no config bundle
  IDs; otherwise abort the migration.
- A snapshot with exactly one declared document maps its one config bundle ID
  to that document ID.
- Any missing, duplicate, or ambiguous historical value aborts migration.
- Rewrite all rows before enabling the new schema. Remove old fields and old
  reads in the same release. No dual read, dual write, or fallback.

The legacy `CreatePodModal` entry points must route to the versioned
WorkerTemplate creation flow before this migration is released.

## Required Verification

1. Proto/Rust/Web binary wire tests preserve document and credential
   requirements without secret material.
2. Backend tests reject unknown, duplicate, wrong-kind, inaccessible, malformed,
   and stale document bindings with field-specific errors.
3. Compiler tests prove Do Agent, OpenClaw, and Seedance emit their actual
   AgentFile target files from the selected document IDs.
4. Migration tests use representative stored snapshots and fail closed on
   ambiguity.
5. Browser tests show one labelled document selector per Definition document,
   reset stale bindings on Worker type change, and expose no direct Pod form.
6. Authorized non-production lifecycle tests prove preflight, create,
   PTY/ACP connection, harmless prompt, termination, and cleanup for each
   eligible Worker.

## Approval Record

The maintainer approved this atomic cutover on July 16, 2026:

```text
config_bundle_ids/configBundleRefs
  -> config_document_bindings[{document_id, config_bundle_id}]
```

with migration failure on ambiguity, no compatibility read path, and removal of
the user-reachable direct Pod creation form.

The Web entry now routes the direct Workspace creation action to the versioned
WorkerTemplate flow. The release must still repair or regenerate every
historical record reported by the migration dry run before `--apply` is used.
