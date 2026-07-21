# Agent Supply and OPC Marketing Design

## Product Definition

Agent Cloud is an enterprise Agent supply and AI-native organization incubation platform. It turns models, skills, knowledge, tools, prompts, and execution environments into governable Agents that can be discovered, installed, operated, and improved across an organization.

The platform supports two connected loops:

1. Enterprise Agent supply: build, verify, release, install, run, and evolve reusable Agents.
2. Organization incubation: combine supplied Agents into AI partners and operating teams that help an individual or a small core team establish an OPC delivery loop.

Higher-education digital employees are a solution built on the same platform. They are not the global product category.

## Product Concept Hierarchy

- Agent: the smallest reusable unit of supplied capability.
- AI partner: a business-facing Agent or coordinated Agent group that keeps context around ongoing goals.
- OPC: a human-led operating unit supported by an AI partner team.
- Agent market: the internal or public distribution surface for approved Agents, skills, and applications.
- Worker: one execution unit used by an Agent.
- Runner and cluster: infrastructure that executes Workers.
- Pod and WorkerSpec: implementation concepts hidden from normal marketing content.

## Audiences

- Enterprise platform owners need to establish a governed internal Agent supply.
- Agent builders need to package and publish reusable capabilities.
- Business teams need to discover and apply trusted Agents without rebuilding them.
- OPC founders need to assemble an AI-native operating team around a business goal.
- Universities need to build digital employees for teaching, research, administration, and industry collaboration.
- Technical administrators need infrastructure, identity, permission, credential, and audit controls.

## Information Architecture

The primary marketing navigation contains:

- Home
- Product
- Solutions
- Agent Market
- Documentation

Each item has a dedicated route. The current `/how-it-works` page becomes the Product page at `/product`. The old route redirects to `/product` so existing links remain valid.

The Solutions page contains three equal solution entries:

- Enterprise Agent Supply
- OPC Incubation
- Higher-Education Digital Employees

Marketplace remains a separate distribution destination and is no longer presented as a solution.

## Homepage

The homepage answers four questions in order:

1. What does the platform supply?
2. How does Agent supply become repeatable?
3. Which organization models does it support?
4. Why can the supplied Agents be trusted?

Homepage regions:

1. Supply-first hero with a visible Agent supply network.
2. Six-stage supply lifecycle: build, verify, release, install, run, evolve.
3. Three solution directions.
4. Product foundation: Agent factory, market, workspace, automation, governance.
5. Current market applications.
6. Trust and deployment controls.
7. Final action to build or acquire the first reusable Agent.

Pricing is excluded.

## Content Rules

- Lead with outcomes and organizational supply, not Agent count or orchestration topology.
- Use Agent for the supplied capability and AI partner for the business relationship.
- Use digital employee only within the higher-education solution.
- Do not expose Pod, AgentPod, WorkerSpec, ResourceRef, revision, or digest on marketing pages.
- Worker and Runner may appear only when explaining runtime compatibility or governance.
- Do not claim unimplemented university workflows. Describe them as supported solution directions built from current composable capabilities.

## Visual Direction

Retain the existing neutral dark, teal-accented product system and current light/dark section rhythm. Use squared operational surfaces, restrained motion, clear status labels, and real workflow language. Avoid decorative gradients, oversized marketing cards, nested cards, and architecture diagrams that require technical knowledge.

## Acceptance Scenarios

1. Given a new visitor, when the homepage loads, then the first viewport identifies Agent Cloud as an Agent supply and AI-native organization incubation platform.
2. Given an enterprise platform owner, when they open Solutions, then Enterprise Agent Supply is immediately visible with a concrete supply lifecycle and outcome.
3. Given an OPC founder, when they open Solutions, then they can understand how supplied Agents become an AI operating team.
4. Given a university visitor, when they open Solutions, then Higher-Education Digital Employees is visible as a first-class solution without redefining the whole platform.
5. Given any visitor, when they use the main navigation, then Home, Product, Solutions, Agent Market, and Documentation each open a dedicated page.
6. Given a mobile visitor, when they open the menu and browse each page, then text, controls, and content do not overlap or overflow.
7. Given a marketing route, when its production bundle is checked, then it does not import the WASM runtime.
