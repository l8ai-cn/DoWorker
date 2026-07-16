import { workspaceFileArtifacts, type AgentArtifactItem } from "@do-worker/agent-ui";
import type { AcpToolCall } from "@/stores/acpSessionTypes";

export function projectWebAcpWorkspaceArtifacts(
  tool: AcpToolCall,
): AgentArtifactItem[] {
  if (
    tool.toolName !== "fileChange" ||
    tool.status !== "completed" ||
    tool.success !== true
  ) {
    return [];
  }
  let argumentsValue: unknown;
  try {
    argumentsValue = JSON.parse(tool.argumentsJson);
  } catch {
    return [];
  }
  const changes =
    argumentsValue && typeof argumentsValue === "object"
      ? (argumentsValue as { changes?: unknown }).changes
      : undefined;
  return workspaceFileArtifacts(`tool-${tool.toolCallId}`, changes);
}
