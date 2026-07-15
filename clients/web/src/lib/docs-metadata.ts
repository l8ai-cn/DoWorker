import type { Metadata } from "next";

interface DocsMeta {
  title: string;
  description: string;
}

const docsMetadataMap: Record<string, DocsMeta> = {
  "/docs": {
    title: "Documentation",
    description:
      "Complete documentation for Do Worker — the agent workforce platform for ambitious teams.",
  },
  "/docs/getting-started": {
    title: "Quick Start",
    description:
      "Set up a Runner, configure a model resource, and create your first Worker through the current four-step wizard.",
  },
  "/docs/concepts": {
    title: "Core Concepts",
    description:
      "Understand Workers, Runners, model resources, credential bindings, workspaces, channels, and task workflows.",
  },
  "/docs/concepts/workers": {
    title: "Worker Types & Runtime",
    description:
      "Inspect every formal Worker Definition, its adapter, interaction modes, credentials, configuration fields, runtime image state, and verification status.",
  },
  "/docs/concepts/loop-and-workflow": {
    title: "Loop and Workflow",
    description:
      "Understand the boundary between Workers, one-time goal loops, and repeatable workflows.",
  },
  "/docs/concepts/resource-orchestration": {
    title: "Resource-native Orchestration",
    description:
      "Declare, validate, plan, and apply versioned Worker, Expert, Workflow, Prompt, and binding resources.",
  },
  "/docs/faq": {
    title: "FAQ",
    description:
      "Frequently asked questions about Do Worker — troubleshooting Runners, Pods, API keys, Git integration, and billing.",
  },
  "/docs/tutorials/runner-setup": {
    title: "Tutorial: Runner Setup",
    description:
      "Step-by-step guide to installing, registering, and verifying an Do Worker Runner for AI agent execution.",
  },
  "/docs/tutorials/mcp-and-skills": {
    title: "Tutorial: MCP Tools & Skills",
    description:
      "Extend AI agents with MCP servers and skills — install custom tools, configure built-in collaboration features, and add reusable workflows.",
  },
  "/docs/tutorials/git-setup": {
    title: "Tutorial: Connect Git Repositories",
    description:
      "Connect GitHub, GitLab, or Gitee to Do Worker and import repositories for AI agent workflows.",
  },
  "/docs/features/agentpod": {
    title: "Worker Types & Runtime",
    description:
      "Legacy path redirected to the Worker Types & Runtime documentation.",
  },
  "/docs/features/channels": {
    title: "Channels",
    description:
      "Multi-agent collaboration spaces where AI agents communicate, share context, and coordinate work in real time.",
  },
  "/docs/features/workflows": {
    title: "Workflows",
    description:
      "Automated feedback workflows for iterative agent-driven development — define triggers, conditions, and actions.",
  },
  "/docs/features/mesh": {
    title: "Core Concepts",
    description:
      "Legacy path redirected to the current Worker-centered documentation model.",
  },
  "/docs/features/repositories": {
    title: "Repositories",
    description:
      "Connect Git providers (GitHub, GitLab) and manage repository access for your AI agents with OAuth integration.",
  },
  "/docs/concepts/repositories-git": {
    title: "Repositories & Git Integration",
    description:
      "Connect Git providers, manage repository access, configure SSH keys, and use Git worktree isolation with AI agents.",
  },
  "/docs/concepts/agentfile": {
    title: "AgentFile Syntax Reference",
    description:
      "Complete reference for AgentFile — the DSL for configuring AI agent execution environments, similar to Dockerfile for containers.",
  },
  "/docs/concepts/agentfile-layer": {
    title: "AgentFile Layer",
    description:
      "Learn about AgentFile Layers — per-Pod override mechanism with a 3-tier merge model for customizing agent configuration at runtime.",
  },
  "/docs/features/tickets": {
    title: "Tickets",
    description:
      "Kanban-style task management integrated with AI agent workflows — create, assign, and track development tasks.",
  },
  "/docs/features/workspace": {
    title: "Workspace",
    description:
      "Git worktree-based workspace isolation ensuring each agent operates on its own branch without conflicts.",
  },
  "/docs/tutorials/first-pod": {
    title: "Tutorial: Your First Worker",
    description:
      "Legacy path redirected to the current Worker creation tutorial.",
  },
  "/docs/tutorials/first-worker": {
    title: "Tutorial: Your First Worker",
    description:
      "Create a Worker through the current runtime, type configuration, workspace, and preflight flow.",
  },
  "/docs/tutorials/ticket-workflow": {
    title: "Tutorial: Task Management with Tickets",
    description:
      "Learn how to use tickets and the Kanban board to organize work and track AI agent progress.",
  },
  "/docs/tutorials/multi-agent-collaboration": {
    title: "Tutorial: Multi-Agent Collaboration",
    description:
      "Set up multiple AI agents that communicate through channels to collaboratively build features.",
  },
  "/docs/tutorials/automated-workflows": {
    title: "Tutorial: Automated Workflows",
    description:
      "Configure scheduled workflows to automate recurring AI agent tasks like code reviews and dependency audits.",
  },
  "/docs/runners/setup": {
    title: "Runner Setup",
    description:
      "Install and configure the Do Worker Runner daemon — self-hosted agent execution with gRPC and mTLS security.",
  },
  "/docs/runners/mcp-tools": {
    title: "MCP Tools",
    description:
      "Model Context Protocol integration for Runners — extend agent capabilities with custom tools and resources.",
  },
  "/docs/guides/git-integration": {
    title: "Git Integration",
    description:
      "Set up Git provider connections, manage SSH keys, and configure repository access for your AI agent workflows.",
  },
  "/docs/guides/multi-agent-workflows": {
    title: "Multi-Agent Workflows",
    description:
      "Design and run multi-agent collaboration workflows — parallel development, code review, and coordinated shipping.",
  },
  "/docs/guides/team-management": {
    title: "Team Management",
    description:
      "Manage teams, roles, and permissions in Do Worker — organize your AI agent fleet for maximum productivity.",
  },
  "/docs/api": {
    title: "API Overview",
    description:
      "Do Worker REST API reference — authenticate, manage Pods, Tickets, Channels, and more programmatically.",
  },
  "/docs/api/authentication": {
    title: "API Authentication",
    description:
      "Authenticate with the Do Worker API using JWT tokens and OAuth — secure access to all platform endpoints.",
  },
  "/docs/api/channels": {
    title: "Channels API",
    description:
      "Create, list, and manage multi-agent collaboration channels via the Do Worker REST API.",
  },
  "/docs/api/workflows": {
    title: "Workflows API",
    description:
      "Manage automated feedback workflows programmatically — create triggers, monitor executions, and retrieve results.",
  },
  "/docs/api/pods": {
    title: "Worker API",
    description:
      "The PodService API used to create, monitor, and terminate Workers.",
  },
  "/docs/api/repositories": {
    title: "Repositories API",
    description:
      "Manage Git repository connections and access tokens via the Do Worker REST API.",
  },
  "/docs/api/runners": {
    title: "Runners API",
    description:
      "Register, monitor, and manage Runner daemons via the REST API — health checks, certificates, and configuration.",
  },
  "/docs/api/tickets": {
    title: "Tickets API",
    description:
      "Create, update, and query development tickets via the Do Worker REST API — integrate with your workflow.",
  },
};

const defaultMeta: DocsMeta = {
  title: "Documentation",
  description:
    "Do Worker documentation — orchestrate AI coding agents at scale.",
};

/**
 * Create Next.js Metadata for a docs page path.
 * Used by individual docs sub-page layout.tsx files.
 */
export function createDocsMetadata(path: string): Metadata {
  const meta = docsMetadataMap[path] ?? defaultMeta;
  return {
    title: meta.title,
    description: meta.description,
    alternates: {
      canonical: `https://agentsmesh.ai${path}`,
    },
    openGraph: {
      title: `${meta.title} | Do Worker Docs`,
      description: meta.description,
      url: `https://agentsmesh.ai${path}`,
    },
  };
}
