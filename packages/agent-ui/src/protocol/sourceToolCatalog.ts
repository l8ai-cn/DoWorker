import { create } from "@bufbuild/protobuf";

import {
  ToolIdentitySchema,
  type ToolIdentity,
} from "@do-worker/proto/agent_workbench/v2/tool_pb";

interface SourceToolDescriptor {
  semanticKey: string;
  schemaVersion: string;
}

const sourceToolCatalog = new Map<string, ReadonlyMap<string, SourceToolDescriptor>>([
  [
    "acp",
    new Map([
      ["Bash", { semanticKey: "shell.execute", schemaVersion: "1" }],
      ["Read", { semanticKey: "filesystem.read", schemaVersion: "1" }],
      ["Write", { semanticKey: "filesystem.write", schemaVersion: "1" }],
      ["Edit", { semanticKey: "filesystem.edit", schemaVersion: "1" }],
      ["Grep", { semanticKey: "filesystem.search", schemaVersion: "1" }],
      ["WebFetch", { semanticKey: "web.fetch", schemaVersion: "1" }],
      [
        "AskUserQuestion",
        { semanticKey: "interaction.question", schemaVersion: "1" },
      ],
      ["shell", { semanticKey: "shell.execute", schemaVersion: "1" }],
      ["fileChange", { semanticKey: "filesystem.change", schemaVersion: "1" }],
      [
        "image_generation",
        { semanticKey: "media.image.generate", schemaVersion: "1" },
      ],
    ]),
  ],
  [
    "claude",
    new Map([
      ["Bash", { semanticKey: "shell.execute", schemaVersion: "1" }],
      ["Read", { semanticKey: "filesystem.read", schemaVersion: "1" }],
      ["Write", { semanticKey: "filesystem.write", schemaVersion: "1" }],
      ["Edit", { semanticKey: "filesystem.edit", schemaVersion: "1" }],
      ["Grep", { semanticKey: "filesystem.search", schemaVersion: "1" }],
      ["WebFetch", { semanticKey: "web.fetch", schemaVersion: "1" }],
      [
        "AskUserQuestion",
        { semanticKey: "interaction.question", schemaVersion: "1" },
      ],
    ]),
  ],
  [
    "codex",
    new Map([
      ["Bash", { semanticKey: "shell.execute", schemaVersion: "1" }],
      ["Read", { semanticKey: "filesystem.read", schemaVersion: "1" }],
      ["Write", { semanticKey: "filesystem.write", schemaVersion: "1" }],
      ["shell", { semanticKey: "shell.execute", schemaVersion: "1" }],
      ["fileChange", { semanticKey: "filesystem.change", schemaVersion: "1" }],
      [
        "image_generation",
        { semanticKey: "media.image.generate", schemaVersion: "1" },
      ],
    ]),
  ],
  [
    "web-user",
    new Map([
      ["web_search_call", { semanticKey: "web.search", schemaVersion: "1" }],
      [
        "file_search_call",
        { semanticKey: "filesystem.search", schemaVersion: "1" },
      ],
      [
        "code_interpreter_call",
        { semanticKey: "code.interpret", schemaVersion: "1" },
      ],
      ["computer_call", { semanticKey: "computer.use", schemaVersion: "1" }],
      [
        "image_generation_call",
        { semanticKey: "media.image.generate", schemaVersion: "1" },
      ],
      ["mcp_call", { semanticKey: "mcp.call", schemaVersion: "1" }],
      ["mcp_list_tools", { semanticKey: "mcp.list-tools", schemaVersion: "1" }],
    ]),
  ],
]);

export function resolveSourceTool(
  sourceProtocol: string,
  exactSourceToolName: string,
): ToolIdentity | undefined {
  const descriptor = sourceToolCatalog.get(sourceProtocol)?.get(exactSourceToolName);
  if (!descriptor) {
    return undefined;
  }

  return create(ToolIdentitySchema, {
    namespace: `agentsmesh.${sourceProtocol}`,
    semanticKey: descriptor.semanticKey,
    schemaVersion: descriptor.schemaVersion,
    sourceToolName: exactSourceToolName,
  });
}
