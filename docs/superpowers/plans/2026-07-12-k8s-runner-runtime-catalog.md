# K8s Runner Runtime Catalog

## Acceptance

1. The Worker creation API exposes a selectable image for each deployed runtime:
   `do-agent`, `grok-build`, `minimax-cli`, `openclaw`, and `hermes`.
2. The catalog and coordinator launcher read one immutable image mapping and reject
   mutable tags, duplicate entries, and unknown runtime slugs.
3. DoAgent accepts its supported model resource protocols during Worker preflight.
4. Oillan K8s manifests define and pre-register every standing Runner and only
   reference image digests.
5. The MiniMax Agentfile row is updated only through the approved DoSQL data-change
   channel; no migration is added for that production correction.

## Verification

1. Focused Go tests cover image mapping validation, selectable Worker options, and
   DoAgent model resolution.
2. Runtime and Kubernetes manifest contract tests pass.
3. Images are built remotely, pushed, pinned by digest, then deployed through DoOps.
4. K8s rollout, Runner registration, API options, and browser Worker creation are
   verified against Oillan.

## Current Release State

- The code baseline rejects mutable runtime images and keeps unverified Worker
  types out of the selectable catalog.
- MiniMax activation requires its existing Agentfile row to be corrected through
  an approved DoSQL data change.
- Grok, OpenClaw, Hermes, and DoAgent require verified Docker Hub digests before
  their catalog entries and K8s Runner manifests are added.
- The xAI credential must be created through the encrypted model-resource path;
  it is not represented in Git or a migration.
