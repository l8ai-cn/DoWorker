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
  return items.filter(userVisibleConversationItem);
}

function userVisibleConversationItem(item: AgentTimelineItem): boolean {
  return (
    (item.kind === "message" &&
      (item.role === "user" || item.role === "assistant")) ||
    item.kind === "attachment"
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
  | "provider_auth_failed"
  | "provider_unavailable"
  | "task_failed"
  | "verified"
  | "verification_failed";

export function userVideoTaskArtifacts(
  snapshot: AgentSessionSnapshot,
  artifacts: readonly AgentArtifactItem[],
): AgentArtifactItem[] {
  return artifacts.filter(
    (artifact) =>
      isVideoArtifact(artifact) &&
      (!snapshot.latestUserCommandId ||
        artifact.provenance?.commandId === snapshot.latestUserCommandId),
  );
}

export function userVideoTaskState(
  snapshot: AgentSessionSnapshot,
  artifacts: readonly AgentArtifactItem[],
): UserVideoTaskState | null {
  const allVideos = artifacts.filter(isVideoArtifact);
  const videos = userVideoTaskArtifacts(snapshot, artifacts);
  if (videos.length === 0) {
    if (
      snapshot.status === "completed" &&
      snapshot.latestUserCommandId &&
      allVideos.length > 0
    ) {
      return "verification_failed";
    }
    if (snapshot.status !== "failed") return null;
    return userVideoProviderFailure(snapshot.error) ?? "task_failed";
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

export function userVideoProviderFailure(
  error: string | null,
): Extract<UserVideoTaskState, "provider_auth_failed" | "provider_unavailable"> | null {
  const normalized = error?.toLowerCase() ?? "";
  if (!normalized) return null;
  if (
    normalized.includes("creative_no_account_available") ||
    normalized.includes("no_account_available")
  ) {
    return "provider_unavailable";
  }
  if (
    normalized.includes("invalid_api_key") ||
    normalized.includes("invalid api key")
  ) {
    return "provider_auth_failed";
  }
  return null;
}
