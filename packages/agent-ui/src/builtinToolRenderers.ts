import { FilePenLine } from "lucide-react";

import type { AgentToolRendererRegistration } from "./react/rendererTypes";
import { ToolRendererRegistry } from "./registry/ToolRendererRegistry";

export function createBuiltinToolRenderers() {
  const registry = new ToolRendererRegistry<AgentToolRendererRegistration>();
  registry.register(
    {
      namespace: "agentsmesh.acp",
      schemaVersion: "1",
      semanticKey: "filesystem.edit",
    },
    { presentation: { icon: FilePenLine, label: "File change" } },
    "builtin.filesystem.edit",
  );
  return registry;
}
