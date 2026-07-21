export const INSTALL_COMMANDS = {
  unix: "curl -fsSL https://agentcloud.ai/install.sh | sh",
  windows: "irm https://agentcloud.ai/install.ps1 | iex",
} as const;
