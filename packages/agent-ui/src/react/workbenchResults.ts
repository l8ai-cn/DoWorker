import type { ComponentType } from "react";

import type { AgentArtifactItem } from "../agentArtifactContracts";
import type { AgentToolActivityItem } from "../agentToolContracts";
import type { ToolRendererRegistry } from "../registry/ToolRendererRegistry";
import type {
  AgentToolRendererRegistration,
  AgentToolWorkbenchRendererProps,
} from "./rendererTypes";
import { isUserVisibleArtifact } from "../artifactResultTrust";

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
  verifiedArtifactsOnly = false,
): WorkbenchResult[] {
  const results: WorkbenchResult[] = artifacts
    .filter((item) => !verifiedArtifactsOnly || isUserVisibleArtifact(item))
    .map((item) => ({
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
