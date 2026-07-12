# Public App Marketplace Closed-Loop Design

- **Date:** 2026-07-12
- **Status:** Approved for implementation
- **Supersedes:** `2026-07-12-dowork-embedded-marketplace-design.md`

## Product Essence

Do Worker Marketplace helps a team discover a work outcome, assess a trusted
application package, enable it for an explicit organization, and start the
first real task without having to understand the underlying Skill, MCP, or
runtime model.

The public storefront and the organization application center are different
surfaces over one Marketplace API and one catalog. The storefront is public and
discovery-oriented. The application center is authenticated and
organization-scoped.

## Roles And Primary Jobs

| Role | Primary job |
| --- | --- |
| Visitor | Find an outcome-oriented application worth evaluating |
| Organization administrator | Confirm compatibility, permissions, quota, and enablement target |
| Team member | Start work with an enabled application and understand its result |
| Marketplace owner | Curate Spaces, tags, publishing policy, and availability |
| Publisher | Publish a verified package with clear outcomes and operating guidance |

## Information Architecture

| Surface | Canonical path | Responsibility |
| --- | --- | --- |
| Public marketplace | `https://market.l8ai.cn` | Discovery, search, filtering, detail, trust, and acquisition handoff |
| Public listing | `https://market.l8ai.cn/apps/{listingSlug}` | Outcome, package composition, requirements, examples, and acquisition intent |
| Acquisition | `https://dowork.l8ai.cn/marketplace/acquire?...` | Authentication, explicit organization selection, preflight, confirmation, and enablement |
| Organization application center | `https://dowork.l8ai.cn/{org}/applications` | Enabled application status, configuration, quota, updates, and recent use |
| Application start | `https://dowork.l8ai.cn/{org}/applications/{installationId}` | First-run task, operational instructions, and result handoff |

`market.l8ai.cn` is the public canonical host. It must not redirect to a
logged-in dashboard route. The prior `/{org}/marketplace` route becomes a
compatibility redirect to the public storefront; it is not a first-level
dashboard activity. The existing organization Skill library remains a runtime
management surface.

## Discovery Model

The first screen is organized by outcomes, not resource types:

1. Software delivery
2. Cross-border commerce operations
3. Teaching and academic operations
4. Customer and business operations
5. Analysis and knowledge work

Each outcome is a Marketplace Space. A listing may appear in multiple Spaces.
Resource type is a filter and package-composition attribute:

| Resource type | Consumer meaning |
| --- | --- |
| Application | A ready-to-enable expert application with an executable first task |
| Skill | A reusable capability included in, or attachable to, an application |
| MCP connector | A system connection definition that needs organization authorization |
| Resource | A model, knowledge, template, or infrastructure dependency |

The storefront does not expose a successful acquisition action for a resource
whose runtime bridge is unavailable. It shows the real availability state
instead of a disabled action with invented promises.

## Taxonomy And Filtering

Tags are governed market taxonomy, not an unstructured text array. Each tag has
an identifier, display name, kind, and optional parent. A listing can carry
multiple tags.

| Tag kind | Examples |
| --- | --- |
| `scene` | Feature development, product publishing, course building |
| `industry` | Cross-border commerce, higher education, enterprise services |
| `audience` | Operator, teacher, engineering lead, delivery engineer |
| `capability` | Testing, translation, research, code review |
| `integration` | GitHub, Shopify, knowledge base |
| `readiness` | Ready now, Runner required, authorization required, approval required |

The public catalog supports `q`, `scene`, `industry`, `audience`, `capability`,
`type`, `integration`, `readiness`, `space`, `sort`, and cursor pagination.
Each cursor is bound to the full query context. Supported sorts are `featured`,
`latest`, and `relevance`; popularity is not shown until the platform has
trustworthy usage data.

## Listing Evaluation

A listing detail page must answer, in order:

1. What result does this application produce?
2. Which team and business scenario is it for?
3. What application, Skills, connectors, and resources are included?
4. What must the organization prepare?
5. Which permissions, quota, and version are involved?
6. What is the first task after enablement?

The detail page shows verified publisher status, maintenance state, tagged
scenarios, package summary, requirements, permissions, version notes,
documentation, support, and a concrete first-run template. It does not use a
large generic hero or placeholder navigation.

## Acquisition And First Use

```text
visitor
  -> public listing
  -> acquire intent
  -> sign in when needed
  -> choose organization explicitly
  -> preflight
  -> confirm plan
  -> enable
  -> organization application page
  -> start first task
  -> result and continuing use
```

Preflight resolves membership, Runner readiness, required connections,
permissions, dependencies, and quota without writing an installation or
reserving quota. Confirmation shows the exact target organization, mutations,
permissions, estimated quota, plan expiry, and first-run result.

After a successful installation, the user lands on the application page, not an
unfiltered Experts list. The page provides one primary action, `Start first
task`, and links to the installed runtime only after the user understands what
will happen.

## Organization Application Center

The application center separates enabled applications into:

| State | User action |
| --- | --- |
| Active | Start work, view recent results, configure connections |
| Needs attention | Fix missing Runner, authorization, quota, or compatibility |
| Update available | Review changed permissions and upgrade deliberately |
| Suspended | Inspect reason and request restoration |

Every application shows its activation state, version, remaining market quota,
required connection status, recent usage, support route, and uninstall or
suspend control when the underlying runtime supports it.

## Data And API Contract

The Marketplace service remains the source of truth for catalog, listing,
taxonomy, entitlement, installation, quota, and lifecycle state. Do Worker
remains the source of truth for runtime execution and result evidence.

Required additions:

- Use market-scoped taxonomy tags and a listing-tag relation as the only public
  read model. Preserve the legacy text column only for existing domain
  compatibility; create and publish flows write normalized tags as the source
  of truth.
- Extend public list and detail responses with tags, package summary,
  activation requirements, first-run templates, documentation, and support.
- Add query filtering and cursor pagination to public listings.
- Add authenticated installation read APIs for an organization's application
  center. They return installation state and never expose credentials.
- Include application destination and first-run template in successful apply
  responses.

## Acceptance Scenarios

1. Given an anonymous visitor, when they open `market.l8ai.cn`, then they can
   filter and inspect listings without loading Do Worker WASM or signing in.
2. Given a listing tagged for a scene and industry, when matching filters are
   applied, then the server returns only matching public listings and the URL is
   shareable.
3. Given an anonymous visitor who selects enablement, when authentication is
   required, then login returns them to the same listing acquisition intent.
4. Given an authenticated user with multiple organizations, when they acquire a
   listing, then they explicitly select one target organization before
   preflight.
5. Given a failed preflight, when the user sees blockers, then no installation,
   entitlement, or quota reservation is written.
6. Given successful enablement, when the user lands in the application center,
   then the enabled application, first task, and next configuration action are
   visible in one screen.
7. Given an unsupported component runtime, when a user views the component,
   then the real availability reason is shown and no impossible acquisition
   action is presented.
