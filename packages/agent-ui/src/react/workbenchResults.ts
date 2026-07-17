import type { ComponentType } from "react";

import type { AgentArtifactItem } from "../agentArtifactContracts";
import type { AgentToolActivityItem } from "../agentToolContracts";
import type { ToolRendererRegistry } from "../registry/ToolRendererRegistry";
import type {
  AgentToolRendererRegistration,
  AgentToolWorkbenchRendererProps,
} from "./rendererTypes";

export type WorkbenchResult =
  | {
      id: string;
      item: AgentArtifactItem;
      kind: "artifact";
    }
  | {
      id: string;
      item: AgentToolActivityItem;
      kind: "tool";
      Renderer: ComponentType<AgentToolWorkbenchRendererProps>;
    };

export function collectWorkbenchResults(
  artifacts: readonly AgentArtifactItem[],
  tools: readonly AgentToolActivityItem[],
  toolRenderers?: ToolRendererRegistry<AgentToolRendererRegistration>,
): WorkbenchResult[] {
  const results: WorkbenchResult[] = artifacts.map((item) => ({
    id: `artifact:${item.id}`,
    item,
    kind: "artifact",
  }));
  for (const item of tools) {
    const Renderer = toolRenderers?.lookup(item.identity)?.workbench;
    if (Renderer) {
      results.push({
        id: `tool:${item.id}`,
        item,
        kind: "tool",
        Renderer,
      });
    }
  }
  return results;
}
