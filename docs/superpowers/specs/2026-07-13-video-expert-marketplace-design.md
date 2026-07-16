# Video Expert Marketplace Design

## Goal

Deliver the first operator-owned video expert suite in AgentsMesh:

- flexible Skill grouping through tags;
- video production, video editing, and directing experts;
- expert marketplace submission, review, publication, installation, and explicit upgrade;
- an executable video runtime rather than instruction-only experts.

## Product Roles

- **Publisher** creates an expert and submits an immutable market release.
- **Platform reviewer** approves or rejects submissions with an auditable reason.
- **Consumer** installs an approved release and chooses when to upgrade.
- **Skill curator** tags Skills and ensures market experts use approved platform Skills.

## Decisions

### Skill Tags

Tags are catalog metadata, not runtime selectors.

- A Skill may have multiple normalized tags.
- Tags drive filtering, grouping, and display in Skill management.
- Experts continue to persist explicit `skill_slugs`.
- Changing tags never changes an existing expert or WorkerSpec.
- Skill Git metadata is authoritative; the database is the query index.

No `skill_groups` table is introduced. A visible group is the set of Skills sharing a tag.

### Marketplace Releases

The hard-coded `marketApplications` list is replaced by persisted immutable releases.

```text
draft -> pending_review -> published
                        -> rejected -> pending_review
published -> withdrawn
```

Each submission snapshots the source and publisher identities, expert configuration,
WorkerSpec, exact approved Skill dependencies, presentation metadata, version, and
review history. Editing the source expert does not mutate submitted or installed releases.

### Installation And Upgrade

Installation copies one published release into the target organization.

- Installation is idempotent by market application and target organization.
- The installed expert records its source application and release.
- A newer release is exposed as an available upgrade.
- Upgrade is explicit and replaces configuration from the selected release.
- Consumer-local edits are never silently overwritten.

### Dependency Rule

A release may reference only active platform-level Skills. Submission fails with the
unavailable dependency list so cross-organization installations cannot appear successful
and then fail at runtime.

## Data Model

### Skills

Add `tags TEXT[] NOT NULL DEFAULT '{}'` to `skills` with a GIN index.

Skill metadata schema advances to version 2 and stores `tags`. Authored create/update and
imported metadata commits keep Git and the catalog synchronized. Upstream sync preserves
curator-managed tags.

### Expert Market Applications

`expert_market_applications` owns stable identity: `id`, slug, publisher organization and
user, latest published release, and timestamps.

### Expert Market Releases

`expert_market_releases` owns immutable versions: application and source expert IDs,
version, lifecycle status, presentation metadata, expert snapshot, WorkerSpec snapshot,
review metadata, and lifecycle timestamps.

Public queries return only the latest published, non-withdrawn release.

### Installed Experts

Experts gain nullable source market application and release IDs for installation
idempotency, provenance, and upgrade discovery.

## API

- Publisher: submit, list statuses, resubmit a rejection, and withdraw.
- Reviewer: list pending releases, inspect dependencies, approve, and reject with a reason.
- Consumer: list and inspect published applications, install, inspect upgrades, and explicitly upgrade.

## User Experience

### Skill Management

- Add and remove tags from a Skill.
- Filter by one or more tags.
- Switch between flat and tag-grouped views.
- Show untagged Skills explicitly.
- Represent loading, empty, saving, error, and disabled states.

### Expert Publishing

Expert details expose `Submit to marketplace`. The form captures summary, category, tags,
outcomes, and icon. The page shows review status and rejection reason.

### Admin Review

The admin console adds Expert Reviews with pending, published, rejected, and withdrawn
views. Review detail displays the snapshot, Worker type, automation level, and dependencies.

### Public Marketplace

Only published releases appear. Installation creates an organization-owned expert.
Installed experts show source version and upgrade availability.

## Initial Video Skill Set

The first release uses reviewed, redistributable Skills: `remotion-best-practices`,
`video-use`, `image2`, `motion-layer-animation`, operator-authored `video-delivery-qa`,
operator-authored `video-storyboard-director`, and `remotion-video-director` where its
license permits redistribution.

Unlicensed GitHub Skills may inform requirements but are not copied or redistributed.

## Initial Experts

### Video Production Expert

Skills: `remotion-best-practices`, `video-use`, `image2`, `motion-layer-animation`,
`video-delivery-qa`.

Produces assets, motion compositions, narration-aware timelines, and verified renders.

### Video Editing Expert

Skills: `video-use`, `remotion-best-practices`, `video-delivery-qa`.

Performs rough cut, captions, audio treatment, pacing, reframing, transitions, and QA.

### Directing Expert

Skills: `video-storyboard-director`, `remotion-video-director`.

Produces a creative brief, script, shot list, storyboard, pacing plan, and handoff package.

## Video Runtime

Add a dedicated `video-studio` Worker type and image based on the Codex runtime:

- Node.js and Codex CLI;
- FFmpeg with libass;
- Chromium and Remotion runtime libraries;
- Python and basic media tooling;
- Noto CJK fonts.

Runtime contract tests execute an ASS subtitle burn and a minimal Remotion render. Video
dependencies are not added to unrelated Worker images.

## Acceptance Scenarios

1. A curator assigns multiple tags and filters or groups Skills by them.
2. Tag changes do not alter expert `skill_slugs` or existing WorkerSpecs.
3. A publisher submits an expert and sees `pending_review`.
4. Submission fails when any dependency is not an active platform Skill.
5. A reviewer rejects with a reason; the publisher edits and resubmits.
6. Approval publishes an immutable release visible in the public market.
7. Another organization installs the release and launches an independent expert.
8. Source expert edits do not change the installed copy.
9. A new version offers an explicit upgrade and never auto-upgrades.
10. All three video experts use `video-studio` and resolve their Skills.
11. The runtime renders a playable 9:16 MP4 with burned subtitles.
12. Browser tests cover tagging, submission, review, install, and upgrade.

## Delivery

Delivery requires backend, Rust/TypeScript contract, web, web-admin, migration,
runtime-image, unit, integration, and browser tests. Release completes only after commit
and push, CI/GitOps confirmation, service health checks, and browser verification in the
user-confirmed target environment.
