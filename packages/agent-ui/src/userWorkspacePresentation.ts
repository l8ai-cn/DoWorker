import type {
  AgentActivityItem,
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
  progressTitle = "Progress",
): AgentTimelineItem[] {
  const visible: AgentTimelineItem[] = [];
  let progress: AgentActivityItem | null = null;
  for (const item of items) {
    if (userVisibleConversationItem(item)) {
      if (progress) visible.push(progress);
      progress = null;
      visible.push(item);
      continue;
    }
    if (userProgressCandidate(item)) {
      progress = mergeUserProgress(progress, item, progressTitle);
    }
  }
  if (progress) visible.push(progress);
  return visible;
}

function userVisibleConversationItem(item: AgentTimelineItem): boolean {
  return (
    (item.kind === "message" &&
      (item.role === "user" || item.role === "assistant")) ||
    item.kind === "attachment"
  );
}

function userProgressCandidate(item: AgentTimelineItem): boolean {
  return (
    item.kind === "tool" ||
    item.kind === "system" ||
    item.kind === "reasoning" ||
    item.kind === "error"
  );
}

function mergeUserProgress(
  current: AgentActivityItem | null,
  item: AgentTimelineItem,
  title: string,
): AgentActivityItem {
  return {
    id: current?.id ?? `${item.id}:user-progress`,
    kind: "system",
    title,
    status: strongerProgressStatus(current?.status, userProgressStatus(item)),
  };
}

function userProgressStatus(
  item: AgentTimelineItem,
): "pending" | "running" | "completed" | "failed" {
  if (item.kind === "error") return "failed";
  if (
    item.kind === "system" ||
    item.kind === "reasoning" ||
    item.kind === "tool"
  ) {
    return item.status;
  }
  return "completed";
}

function strongerProgressStatus(
  current: "pending" | "running" | "completed" | "failed" | undefined,
  next: "pending" | "running" | "completed" | "failed",
) {
  const rank = { completed: 0, pending: 1, running: 2, failed: 3 };
  if (!current) return next;
  return rank[next] > rank[current] ? next : current;
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
