# Multi-tenant Industry Marketplace Platform

- **Date:** 2026-07-11
- **Status:** Approved product design; implementation has not started
- **Product:** Do Worker Marketplace Platform
## 1. Purpose and Boundaries

Do Worker Marketplace Platform lets an enterprise, university, or industry
operator create and run a branded marketplace for AI applications and resources.
Each marketplace owns its storefront, spaces, catalog, users, publishing policy,
entitlements, quotas, and reports. This is a B2B2C marketplace platform, not a
single global Skill directory. The public market is one marketplace instance.

This document uses **Skill** as the capability-package term and does not define a
separate `Scale` resource type.
The marketplace platform owns marketplace provisioning, brand and domain,
Spaces, listings, publishing, membership, entitlement, quota, ledger, storefront,
and management experiences. The Do Worker runtime continues to own Worker,
WorkerSpec, Expert, Skill, model, Runner, repository, secrets, MCP connection
instances, dispatch, and execution evidence. Marketplace APIs call controlled
runtime APIs and do not duplicate runtime business state.

## 2. Participants

| Role | Responsibility |
| --- | --- |
| Platform operator | Infrastructure, security baseline, runtime integration, and platform governance |
| Marketplace owner | Creates a market and controls its business or internal-use policy |
| Marketplace administrator | Manages brand, Spaces, listings, users, quotas, approvals, and reports |
| Space maintainer | Curates one domain and maintains its evaluation standards |
| Publisher | Submits applications, Skills, connectors, or resources |
| Consumer | Discovers, acquires, configures, and uses listings |
| Quota administrator | Allocates quota to organizations, departments, teams, or users |

One user may hold different roles in different marketplaces.

## 3. Marketplace and Space

A `Marketplace` is an independently operated tenant with its own domain, brand,
membership, policy, catalog, and quota economy. A `Space` is a governed domain
inside a market. It organizes listings, collections, maintainers, and evaluation
standards around business outcomes.
Spaces do not own runtime instances. Consumers acquire listings from a Space and
create organization-scoped installations in the runtime platform.

## 4. Catalog and Listing Model

| Object | Meaning |
| --- | --- |
| `CatalogItem` | Reusable platform resource with stable identity |
| `CatalogItemVersion` | Immutable version and content digest |
| `Listing` | Market-specific offer referencing one catalog item |
| `ListingVersion` | Approved item version, presentation, policy, and quota rule |
| `Entitlement` | Permission for a user or organization to acquire or use a listing |
| `Installation` | Concrete organization-scoped installation |
| `QuotaPlan` | Grant, renewal, and consumption rules |
| `QuotaAccount` | Balance holder for a market, organization, department, or user |
| `UsageEvent` | Immutable evidence of metered consumption |
| `LedgerEntry` | Immutable reservation, debit, release, credit, or adjustment |

One catalog item may be listed in multiple markets with independent names,
visibility, quota costs, support terms, and approval requirements.
| Type | Consumer action | Runtime result |
| --- | --- | --- |
| Application | Acquire, configure, and run | Expert or managed application installation |
| Skill | Install or attach to an application | Version-locked Skill installation |
| MCP Connector | Authorize and configure | User- or organization-owned connection instance |
| Resource | Subscribe, allocate, or bind | Model, compute, knowledge, dataset, or template access |

An MCP listing publishes a connector definition, authorization requirements, tool
scope, and setup instructions. It never publishes credentials or a concrete
connection instance.

## 5. Marketplace Lifecycle

```text
draft -> configuring -> review -> published -> suspended -> archived
```
Creation requires selecting a blank or industry template; configuring identity,
brand, locale, domain, membership, SSO, Spaces, maintainers, resource types,
publisher policy, visibility, approval, and quota rules; publishing initial
listings; and passing domain, authentication, catalog, entitlement, quota, and
runtime checks. Any failed required check blocks publication.

## 6. Listing Lifecycle

```text
draft -> submitted -> validating -> needs_changes -> approved
      -> scheduled/published -> suspended -> deprecated -> removed
```
A publisher selects an immutable catalog item version, writes market-specific
content, declares outcomes and permissions, sets visibility and quota cost,
attaches verification evidence, and submits it.
Validation covers schema, identifiers, digest, dependencies, provenance, clean
organization installability, permissions, MCP scopes, secret references, external
accounts, smoke tests, acceptance tests, documentation, support, and change notes.
New releases never mutate existing listing versions. Permission expansion requires
a new approval and explicit consumer confirmation.

## 7. Consumer Journey

```text
discover -> compare -> acquire -> configure -> preflight
         -> reserve quota -> install/run -> verify -> settle usage
```

The storefront leads with outcomes and Spaces, not Skill names. Listing detail
shows outcomes, examples, publisher identity, verification status, required
accounts, permissions, resources, quota cost, maintenance state, and history.
Acquisition requires an explicit organization. The platform never silently selects
the first available organization. Installation has two phases:
1. `Plan` resolves entitlement, compatibility, dependencies, permissions, resource
   bindings, quota, and expected mutations without writing data.
2. `Apply` locks the plan, reserves quota, installs dependencies, creates and
   verifies the runtime installation, then settles or releases the reservation.
Partial installation is not a valid success state.

## 8. Quota and Metering

The product distinguishes API access tokens, provider-level model tokens, and
marketplace credits used for understandable allocation and consumption. Meters may
include model input/output tokens, Worker execution seconds, GPU seconds, MCP
calls, storage GB-days, and successful application runs.
```text
preflight -> reserve -> execute -> settle
                         |
                         +-> fail -> release
```

Usage events and ledger entries are append-only. Every event carries a globally
unique idempotency key. Balance is derived from the ledger and is never edited
directly. Insufficient quota reports exact required and available amounts.
Execution must not silently switch to a cheaper model or resource.

## 9. Applications and Deployment

| Deployable | Responsibility |
| --- | --- |
| `marketplace-storefront` | Branded public and authenticated consumer experience |
| `marketplace-console` | Owner, administrator, maintainer, publisher, and quota administration |
| `marketplace-api` | Market, Space, catalog, listing, entitlement, quota, ledger, and reporting APIs |
| Do Worker backend | Runtime resources, installation execution, and usage evidence |

All customers use a multi-tenant marketplace backend by default. Dedicated
deployment is an explicit enterprise isolation mode, not a fallback.
Every query, cache key, artifact path, usage event, and audit record is scoped by
`marketplace_id`. A custom domain resolves to exactly one active marketplace.

## 10. Security and Governance

- Marketplace roles do not implicitly grant runtime organization permissions.
- Runtime identifiers are resolved against the selected organization before use.
- Secrets stay in the runtime secret store; listings contain references only.
- High-risk permissions require market and runtime organization approval.
- Suspension blocks new acquisition while preserving status for installations.
- Emergency runtime disablement requires an audited security action.
- Quota adjustments, publication decisions, and permission approvals are audited.

## 11. Initial Delivery Scope

The first release includes marketplace creation, branding, domain binding, users,
roles, Spaces, collections, four listing types, publishing workflow, consumer
discovery, entitlement, configuration, installation, monthly quota grants,
reservation, settlement, release, ledger, reports, runtime integration, audit,
and critical loading, empty, blocked, success, and failure states.
Cash payment, publisher revenue sharing, public ratings, forums, and algorithmic
recommendations are excluded until one real vertical marketplace validates supply,
acquisition, successful usage, and quota operations.

## 12. Acceptance Scenarios

- Given a market owner, when brand, identity, policy, Space, and domain checks
  pass, then the owner can publish a reachable branded marketplace.
- Given one catalog item, when two markets list it, then presentation, visibility,
  quota cost, and approval state remain independent.
- Given a consumer without install permission, when acquisition is requested,
  then an approval request is created and no installation is written.
- Given incompatible runtime resources, when preflight runs, then exact blockers
  are returned without mutation or quota reservation.
- Given sufficient quota and compatible resources, when apply succeeds, then all
  dependencies and the installation exist, verification passes, and usage settles once.
- Given an install-stage failure, when compensation completes, then no partial
  installation remains and reserved quota is released.
- Given a version that expands permissions, when a consumer upgrades, then
  organization approval is required before mutation.
- Given duplicate usage delivery, when the same idempotency key is processed,
  then the ledger contains exactly one effective debit.

## 13. Existing Product Migration

The current `/marketplace` becomes the platform-operated default marketplace.
Hard-coded expert applications become seeded catalog items and listings.
The existing organization Skill library remains an internal runtime management
surface and must not be presented as the public marketplace. Existing Skill
runtime semantics remain unchanged when linked to marketplace installations.
Detailed design index:
`2026-07-11-marketplace-detailed-design-index.md`
