import { POD_MODE_PTY } from "@/lib/pod-modes";
import type { PodMode } from "@/lib/pod-modes";
import { resolveDurationKind } from "./WorkerDurationPreset";

export function step1Summary(
  imageSlug: string | null,
  interactionMode: PodMode,
  perpetual: boolean,
  destroyPolicy: string,
  repositorySlug: string | undefined,
  branchName: string | undefined,
  credentialName: string | undefined,
  t: (key: string, values?: Record<string, string | number>) => string,
): string | undefined {
  if (!imageSlug) return undefined;
  const mode =
    interactionMode === POD_MODE_PTY
      ? t("ide.createPod.modePty")
      : t("ide.createPod.modeAcp");
  const duration = durationLabel(perpetual, destroyPolicy, t);
  const parts = [`${imageSlug} · ${duration} · ${mode}`];
  if (credentialName?.trim()) {
    parts.push(credentialName.trim());
  }
  if (repositorySlug) {
    const branch = branchName?.trim();
    parts.push(branch ? `${repositorySlug}@${branch}` : repositorySlug);
  }
  return parts.join(" · ");
}

export function step2Summary(
  knowledgeCount: number,
  skillCount: number,
  t: (key: string, values?: Record<string, string | number>) => string,
): string | undefined {
  if (knowledgeCount === 0 && skillCount === 0) return undefined;
  const parts: string[] = [];
  if (knowledgeCount > 0) {
    parts.push(t("ide.createPod.stepSummaryKnowledge", { count: knowledgeCount }));
  }
  if (skillCount > 0) {
    parts.push(t("ide.createPod.stepSummarySkills", { count: skillCount }));
  }
  return parts.join(" · ");
}

export function step3Summary(
  rawMode: boolean,
  hasLayer: boolean,
  t: (key: string) => string,
): string | undefined {
  if (rawMode) return t("ide.createPod.stepSummaryCustomAgentFile");
  if (hasLayer) return t("ide.createPod.stepSummaryAutoAgentFile");
  return undefined;
}

export function durationLabel(
  perpetual: boolean,
  destroyPolicy: string,
  t: (key: string) => string,
): string {
  const kind = resolveDurationKind(
    perpetual,
    destroyPolicy as "manual" | "idle" | "completed",
  );
  return kind === "long"
    ? t("ide.createPod.durationLongTitle")
    : t("ide.createPod.durationShortTitle");
}
