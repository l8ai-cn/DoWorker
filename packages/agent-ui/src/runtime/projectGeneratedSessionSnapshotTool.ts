import type {
  ToolExecution,
  ToolProgress,
} from "@do-worker/proto/agent_workbench/v2/tool_pb";

import type { AgentTimelineItem, AgentToolResult } from "../contracts";
import {
  projectArtifactReference,
  type ArtifactCatalog,
} from "./projectGeneratedSessionSnapshotArtifacts";
import { projectToolResultBlocks } from "./projectGeneratedSessionSnapshotContent";
import {
  decodeStructuredPayload,
  formatStructuredPayload,
} from "./projectGeneratedSessionSnapshotPayload";
import { projectToolStatus } from "./projectGeneratedSessionSnapshotStatuses";

export function projectToolExecution(
  tool: ToolExecution,
  itemId: string,
  catalog: ArtifactCatalog,
): AgentTimelineItem[] {
  const identity = tool.identity;
  if (
    !identity?.namespace ||
    !identity.semanticKey ||
    !identity.schemaVersion
  ) {
    return [
      {
        id: itemId,
        kind: "system",
        title: "Unsupported tool execution",
        detail: `executionId=${tool.executionId || "missing"}; identity=missing`,
        status: "failed",
      },
    ];
  }
  const input = decodeStructuredPayload(tool.input);
  const results = projectToolResults(tool, itemId, catalog);
  const textResults = results
    .filter(
      (result): result is Extract<AgentToolResult, { kind: "text" }> =>
        result.kind === "text",
    )
    .map((result) => result.text);
  return [
    {
      id: itemId,
      kind: "tool",
      identity: {
        namespace: identity.namespace,
        semanticKey: identity.semanticKey,
        schemaVersion: identity.schemaVersion,
      },
      title: tool.title || identity.sourceToolName || identity.semanticKey,
      detail: toolDetail(tool),
      input: input?.text,
      inputValue: input?.value,
      output: textResults.length > 0 ? textResults.join("\n\n") : undefined,
      results,
      status: projectToolStatus(tool.phase),
    },
  ];
}

function projectToolResults(
  tool: ToolExecution,
  itemId: string,
  catalog: ArtifactCatalog,
): AgentToolResult[] {
  const results = tool.results.flatMap((result, resultIndex) => {
    const resultId = result.resultId || `${itemId}:result:${resultIndex}`;
    return [
      ...projectToolResultBlocks(result.blocks, resultId, catalog),
      ...result.artifacts.map((artifact, artifactIndex) =>
        artifactResult(artifact, `${resultId}:artifact:${artifactIndex}`, catalog),
      ),
    ];
  });
  tool.artifacts.forEach((artifact, index) => {
    results.push(artifactResult(artifact, `${itemId}:artifact:${index}`, catalog));
  });
  return results;
}

function artifactResult(
  reference: ToolExecution["artifacts"][number],
  id: string,
  catalog: ArtifactCatalog,
): AgentToolResult {
  const projected = projectArtifactReference(reference, id, catalog);
  if (projected.kind !== "artifact") {
    return { id, kind: "data", value: projected.detail };
  }
  return {
    id,
    kind: "artifact",
    artifactId: projected.artifactId,
    mediaType: projected.mimeType,
    representationId: projected.selectedRepresentationId,
    revision: projected.revision,
    role: projected.role,
    schemaVersion: projected.schemaVersion,
  };
}

function toolDetail(tool: ToolExecution): string | undefined {
  const failure = tool.failure
    ? [
        `[${tool.failure.code || "tool_failed"}] ${tool.failure.message}`,
        formatStructuredPayload(tool.failure.details),
      ]
        .filter(Boolean)
        .join("\n")
    : undefined;
  return [tool.detail, progressDetail(tool.progress), failure]
    .filter(Boolean)
    .join("\n");
}

function progressDetail(progress: ToolProgress | undefined): string | undefined {
  if (!progress) return undefined;
  const count =
    progress.current !== undefined || progress.total !== undefined
      ? `${progress.current?.toString() ?? "?"}/${progress.total?.toString() ?? "?"}${progress.unit ? ` ${progress.unit}` : ""}`
      : undefined;
  const fraction =
    progress.fraction === undefined
      ? undefined
      : `${Math.round(progress.fraction * 100)}%`;
  return [progress.stage, progress.message, count, fraction]
    .filter(Boolean)
    .join(" - ");
}
