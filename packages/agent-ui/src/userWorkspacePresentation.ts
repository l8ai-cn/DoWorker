import type {
  AgentArtifactItem,
  AgentSessionSnapshot,
  AgentTimelineItem,
} from "./contracts";
import {
  isUserVisibleArtifact,
  isVerifiedReadyVideoArtifact,
  isVideoArtifact,
} from "./artifactResultTrust";

export type AgentWorkspacePresentation = "developer" | "user";

export function userConversationItems(
  items: readonly AgentTimelineItem[],
): AgentTimelineItem[] {
  return items.filter(
    (item) =>
      item.kind === "message" &&
      (item.role === "user" || item.role === "assistant"),
  );
}

export function userVisibleArtifacts(
  items: readonly AgentArtifactItem[],
): AgentArtifactItem[] {
  return items.filter(isUserVisibleArtifact);
}

export type UserVideoTaskState =
  | "failed"
  | "partial"
  | "processing"
  | "task_failed"
  | "verified"
  | "verification_failed";

export function userVideoTaskState(
  snapshot: AgentSessionSnapshot,
  artifacts: readonly AgentArtifactItem[],
): UserVideoTaskState | null {
  const allVideos = artifacts.filter(isVideoArtifact);
  const videos = allVideos.filter(
    (artifact) =>
      (!snapshot.latestUserCommandId ||
        artifact.provenance?.commandId === snapshot.latestUserCommandId),
  );
  if (videos.length === 0) {
    if (
      snapshot.status === "completed" &&
      snapshot.latestUserCommandId &&
      allVideos.length > 0
    ) {
      return "verification_failed";
    }
    return snapshot.status === "failed" ? "task_failed" : null;
  }
  const verified = videos.some(isVerifiedReadyVideoArtifact);
  if (snapshot.status === "failed" && verified) return "partial";
  if (snapshot.status === "failed") return "failed";
  if (verified) return "verified";
  if (
    videos.some(
      (artifact) =>
        artifact.status === "completed" &&
        !isVerifiedReadyVideoArtifact(artifact),
    )
  ) {
    return "verification_failed";
  }
  if (videos.some((artifact) => artifact.status === "failed")) return "failed";
  if (snapshot.status === "completed") return "verification_failed";
  return "processing";
}
