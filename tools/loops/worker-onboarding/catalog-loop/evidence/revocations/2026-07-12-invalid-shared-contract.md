# Invalidated Prior Acceptance

Date: 2026-07-12

The following checklist items were reopened:

- `accept-discover-inventory`
- `accept-establish-shared-contract`

The old verifiers proved only that twelve catalog files and schema-shaped JSON
existed. They did not prove that `workercreation.Service` consumes the catalog,
that the database AgentFile projection matches it, or that the definition
metadata reaches Runner, Rust Core, Web, or a running image.

Observed root causes:

- Worker creation resolves types through the `agents` database table and
  AgentFile source, while the loaded `workerdefinition.Catalog` is not injected
  into that path.
- The definition loader discards credential-binding and configuration-document
  fields after validation.
- The runtime image catalog is a separate partial mapping; no target image had
  been built or probed.
- Runner ACP selection can infer a generic adapter from command names, and the
  image preparation script substitutes mock binaries for two targets.
- No authenticated E2E fixture existed, so the Web creation flow was not run.

This is an evidence correction, not a compatibility downgrade. The prior
artifacts remain available for historical audit. The active plan is
`docs/superpowers/plans/2026-07-12-worker-integration-evidence-rebuild.md`.
