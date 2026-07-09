# Homepage AI Workforce Redesign

## Goal

Reposition the homepage from a developer-oriented agent infrastructure page to a
product experience for building and directing an AI workforce across research,
content, operations, sales, knowledge work, and product development.

The first viewport must answer three questions within five seconds:

1. What is Do Worker? An AI workforce platform.
2. What can it do? Assemble specialized Workers around a goal and deliver results.
3. What should I do next? Watch a real task unfold.

## Product Grounding

Every homepage claim maps to an existing product primitive:

- AgentPod provides an isolated workplace for each Worker.
- Tickets represent goals, tasks, ownership, and delivery status.
- Channels provide shared context and collaboration.
- Mesh exposes Worker relationships and live activity.
- Loop schedules repeatable work.
- Runners keep execution and data inside controlled infrastructure.
- Credentials and audit surfaces provide organizational governance.

The homepage may translate these primitives into non-technical language, but must
not imply unsupported autonomous business integrations or fully unattended
decision-making.

## Information Architecture

### 1. Navigation

Use outcome-oriented labels: Product, Scenarios, How it works, Pricing, Resources.
The primary action is “Watch demo”; sign-in and language controls remain available.

### 2. Hero: Give a Goal, Build a Team

Lead with “Give it a goal. Build an AI team.” The supporting copy explains that
specialized Workers research, create, operate, and deliver together while the user
keeps control.

The hero contains an interactive mission console based on actual product concepts:
a goal, Worker roster, activity stream, human checkpoint, and final deliverable.
It replaces the decorative mesh as the primary visual. The existing video remains
available as a secondary path.

### 3. Scenario Switcher

Present six scenarios: research, content, operations, sales, knowledge work, and
product development. Selecting a scenario updates the goal, Worker roles, workflow,
and deliverable without changing the underlying platform model.

This section communicates breadth without presenting Do Worker as a template
library. The shared message is that one workforce system supports many outcomes.

### 4. Work Lifecycle

Explain the product through one continuous lifecycle:

1. Describe the goal.
2. Workers split and coordinate the work.
3. The user reviews important decisions.
4. The team delivers an inspectable result.

Use Ticket, Channel, Mesh, and Pod-inspired interface fragments rather than generic
illustrations.

### 5. Platform Capabilities

Translate existing technical capabilities into four buyer outcomes:

- Organize: roles, tasks, ownership, and shared context.
- Observe: live activity, evidence, status, and deliverables.
- Control: permissions, checkpoints, credentials, and audit history.
- Operate: self-hosted execution, schedules, and reusable workflows.

### 6. Trust and Deployment

Keep self-hosting and data control, but position them as reasons organizations can
adopt the workforce safely. Move CLI compatibility and agent logos below the core
product story.

### 7. Pricing and Final Action

Preserve live public pricing. End with a choice between watching the product demo
and starting free. Enterprise contact remains available but is not the dominant
first action.

## Content Changes

Remove the homepage sections that frame the category around terminal agents,
coding-agent comparisons, installation commands, and development-only workflows.
Technical installation continues to live in Download and Docs.

Rewrite all visible homepage content around goals, roles, collaboration, oversight,
and outcomes. Maintain translation key parity across all eight locales.

Update structured data from `DeveloperApplication` and coding-agent keywords to a
general business `SoftwareApplication` positioned as an AI workforce platform.

## Visual Direction

Retain the recognizable dark Do Worker identity, mint accent, Space Grotesk display
type, and restrained glass surfaces. Replace the current star-field aesthetic with
a warmer operational canvas:

- deep graphite rather than blue-black space;
- mint for active work and successful handoffs;
- amber for human review checkpoints;
- off-white for high-priority content;
- fine grid and connective traces derived from the product interface;
- asymmetric editorial composition instead of centered stacked sections.

The page should feel like a premium work system, not a developer landing template
or science-fiction control panel.

## Interaction

- The mission console advances through a short deterministic task sequence.
- Users can pause, replay, or switch scenarios.
- Scenario controls are keyboard accessible and expose selected state.
- Motion respects `prefers-reduced-motion`.
- Mobile presents the same story as a vertical activity timeline.
- No interaction depends on WebAssembly or authenticated application state.

## Technical Boundaries

- Preserve `useLightSession`, redirect behavior, `LightAuthButtons`, and public
  pricing fetches.
- Do not import the Rust/WASM runtime or Zustand auth store into marketing routes.
- Keep homepage components below 200 lines and split data, visuals, and behavior by
  responsibility.
- Reuse existing design tokens where useful; scope new tokens to the landing page.
- Avoid changing dashboard behavior or backend contracts.

## Verification

- Add focused component tests for scenario selection, mission replay, reduced
  motion behavior, and primary CTA behavior.
- Add a homepage Playwright smoke test for desktop and mobile viewport structure.
- Run web unit tests, lint, type-check, and the no-WASM marketing guard.
- Validate all locale files contain the required landing keys.
- Perform browser review at desktop, tablet, and mobile widths.

## Completion Criteria

- The first viewport no longer identifies Do Worker as a coding or terminal tool.
- All six scenarios resolve to the same credible workforce product model.
- Product visuals are derived from existing capabilities rather than invented
  integrations.
- The old developer-first sections are removed from the homepage.
- The page remains fast, responsive, accessible, internationalized, and WASM-free.
