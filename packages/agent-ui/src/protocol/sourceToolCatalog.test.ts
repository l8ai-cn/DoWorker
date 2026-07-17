import { describe, expect, it } from "vitest";

import { resolveSourceTool } from "./index";

describe("source tool catalog", () => {
  it.each([
    ["acp", "Bash", "shell.execute"],
    ["acp", "Read", "filesystem.read"],
    ["acp", "Write", "filesystem.write"],
    ["acp", "Edit", "filesystem.edit"],
    ["acp", "Grep", "filesystem.search"],
    ["acp", "WebFetch", "web.fetch"],
    ["acp", "AskUserQuestion", "interaction.question"],
    ["acp", "shell", "shell.execute"],
    ["acp", "fileChange", "filesystem.change"],
    ["acp", "image_generation", "media.image.generate"],
    ["codex", "Bash", "shell.execute"],
    ["codex", "Read", "filesystem.read"],
    ["codex", "Write", "filesystem.write"],
    ["codex", "shell", "shell.execute"],
    ["codex", "fileChange", "filesystem.change"],
    ["codex", "image_generation", "media.image.generate"],
    ["claude", "Bash", "shell.execute"],
    ["claude", "Read", "filesystem.read"],
    ["claude", "Write", "filesystem.write"],
    ["claude", "Edit", "filesystem.edit"],
    ["claude", "Grep", "filesystem.search"],
    ["claude", "WebFetch", "web.fetch"],
    ["claude", "AskUserQuestion", "interaction.question"],
    ["web-user", "web_search_call", "web.search"],
    ["web-user", "file_search_call", "filesystem.search"],
    ["web-user", "code_interpreter_call", "code.interpret"],
    ["web-user", "computer_call", "computer.use"],
    ["web-user", "image_generation_call", "media.image.generate"],
    ["web-user", "mcp_call", "mcp.call"],
    ["web-user", "mcp_list_tools", "mcp.list-tools"],
  ])(
    "maps %s tool %s to its reviewed identity",
    (sourceProtocol, sourceToolName, semanticKey) => {
      expect(resolveSourceTool(sourceProtocol, sourceToolName)).toMatchObject({
        namespace: `agentsmesh.${sourceProtocol}`,
        semanticKey,
        schemaVersion: "1",
        sourceToolName,
      });
    },
  );

  it.each([
    ["acp", "shell_exec"],
    ["acp", "bash"],
    ["acp", "ReadFile"],
    ["codex", "shell_exec"],
    ["codex", "bash"],
    ["codex", "ReadFile"],
    ["claude", "bash"],
    ["claude", "shell_exec"],
    ["claude", "ReadFile"],
    ["claude", "Read file"],
    ["claude", "Bash "],
    ["claude", "*"],
    ["claude", "shell"],
    ["claude", "fileChange"],
    ["codex", "Edit"],
    ["codex", "Image generation"],
    ["codex", "prefix-image_generation"],
    ["web-user", "Bash"],
    ["web-user", "web_search"],
    ["web-user", "file_search"],
    ["web-user", "code_interpreter"],
    ["web-user", "computer"],
    ["web-user", "image_generation"],
    ["web-user", "mcp_list_tool"],
    ["web-user", "MCP_call"],
    ["unknown", "Bash"],
  ])("does not guess %s tool %s", (sourceProtocol, sourceToolName) => {
    expect(resolveSourceTool(sourceProtocol, sourceToolName)).toBeUndefined();
  });
});
