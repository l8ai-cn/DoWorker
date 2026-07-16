import type { AgentCommand } from "./contracts";

export const DEFAULT_AGENT_COMMANDS: AgentCommand[] = [
  {
    name: "compact",
    label: "/compact",
    description: "Compact the current context",
  },
  {
    name: "context",
    label: "/context",
    description: "Show context usage",
  },
  {
    name: "help",
    label: "/help",
    description: "Show available commands",
  },
  {
    name: "model",
    label: "/model",
    description: "Switch the active model",
    requiresArgument: true,
  },
  {
    name: "effort",
    label: "/effort",
    description: "Change reasoning effort",
    requiresArgument: true,
  },
];
