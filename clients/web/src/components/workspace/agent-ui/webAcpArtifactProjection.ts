import type {
  AgentArtifactItem,
  AgentTimelineItem,
} from "@do-worker/agent-ui";

export function mergeWebAcpArtifacts(
  items: AgentTimelineItem[],
  discovered: AgentArtifactItem[],
): AgentTimelineItem[] {
  const artifactIds = new Set(
    items.flatMap((item) =>
      item.kind === "artifact" ? [item.artifactId] : [],
    ),
  );
  for (const artifact of discovered) {
    if (artifactIds.has(artifact.artifactId)) continue;
    artifactIds.add(artifact.artifactId);
    items.push(artifact);
  }
  return items;
}
