import type {
  AgentArtifactItem,
  AgentSessionSnapshot,
} from "./contracts";
import { isVerifiedReadyVideoArtifact } from "./artifactResultTrust";
import {
  type UserVideoTaskState,
  userVideoTaskArtifacts,
  userVideoTaskState,
} from "./userWorkspacePresentation";

export type UserVideoExecutionStepId =
  | "request"
  | "generation"
  | "preview"
  | "verification";

export type UserVideoExecutionDetail =
  | "generation_failed"
  | "generation_ready"
  | "preview_failed"
  | "preview_ready"
  | "preparing_preview"
  | "published"
  | "queued"
  | "rendering"
  | "request_received"
  | "verification_failed"
  | "verification_incomplete"
  | "verifying"
  | "waiting";

export interface UserVideoExecutionStep {
  detail: UserVideoExecutionDetail;
  id: UserVideoExecutionStepId;
  progress?: number;
  status: "pending" | "running" | "completed" | "failed";
}

export function userVideoExecutionSteps(
  snapshot: AgentSessionSnapshot,
  artifacts: readonly AgentArtifactItem[],
): UserVideoExecutionStep[] {
  const state = userVideoTaskState(snapshot, artifacts);
  const video = latestVideo(userVideoTaskArtifacts(snapshot, artifacts), state);
  if (!state || !video || video.manifest?.kind !== "video") return [];

  const steps = [
    step("request", "completed", "request_received"),
    step("generation", "pending", "waiting"),
    step("preview", "pending", "waiting"),
    step("verification", "pending", "waiting"),
  ];
  const phase = video.manifest.stage;
  const active = sessionActive(snapshot.status);
  if (phase === "failed") return failedGeneration(steps);
  if (phase === "queued") {
    return setStep(
      steps,
      "generation",
      active ? "pending" : "failed",
      active ? "queued" : "generation_failed",
    );
  }
  if (phase === "rendering") {
    return setStep(
      steps,
      "generation",
      active ? "running" : "failed",
      active ? "rendering" : "generation_failed",
      video.manifest.progressFraction,
    );
  }
  if (phase === "transcoding") {
    setStep(steps, "generation", "completed", "generation_ready");
    return setStep(
      steps,
      "preview",
      active ? "running" : "failed",
      active ? "preparing_preview" : "preview_failed",
    );
  }
  if (phase !== "ready") return steps;

  setStep(steps, "generation", "completed", "generation_ready");
  setStep(steps, "preview", "completed", "preview_ready");
  if (state === "verified" || state === "partial") {
    return setStep(steps, "verification", "completed", "published");
  }
  if (state === "verification_failed") {
    return setStep(steps, "verification", "failed", "verification_failed");
  }
  return setStep(
    steps,
    "verification",
    snapshot.status === "failed" ? "failed" : "running",
    snapshot.status === "failed" ? "verification_incomplete" : "verifying",
  );
}

function sessionActive(status: AgentSessionSnapshot["status"]): boolean {
  return status === "launching" || status === "running" || status === "waiting";
}

function failedGeneration(
  steps: UserVideoExecutionStep[],
): UserVideoExecutionStep[] {
  return setStep(steps, "generation", "failed", "generation_failed");
}

function latestVideo(
  videos: readonly AgentArtifactItem[],
  state: UserVideoTaskState | null,
): AgentArtifactItem | undefined {
  if (state === "verified" || state === "partial") {
    return videos.find(isVerifiedReadyVideoArtifact) ?? videos.at(-1);
  }
  return videos.reduce<AgentArtifactItem | undefined>(
    (latest, video) => (!latest || video.revision > latest.revision ? video : latest),
    undefined,
  );
}

function setStep(
  steps: UserVideoExecutionStep[],
  id: UserVideoExecutionStepId,
  status: UserVideoExecutionStep["status"],
  detail: UserVideoExecutionDetail,
  progress?: number,
): UserVideoExecutionStep[] {
  const index = steps.findIndex((step) => step.id === id);
  steps[index] = step(id, status, detail, progress);
  return steps;
}

function step(
  id: UserVideoExecutionStepId,
  status: UserVideoExecutionStep["status"],
  detail: UserVideoExecutionDetail,
  progress?: number,
): UserVideoExecutionStep {
  return {
    detail,
    id,
    ...(validProgress(progress) ? { progress } : {}),
    status,
  };
}

function validProgress(value: number | undefined): value is number {
  return value !== undefined && value >= 0 && value <= 1;
}
