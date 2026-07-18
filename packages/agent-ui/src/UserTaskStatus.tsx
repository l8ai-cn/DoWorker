import {
  AlertTriangle,
  CheckCircle2,
  LoaderCircle,
} from "lucide-react";

import type {
  AgentArtifactItem,
  AgentSessionSnapshot,
} from "./contracts";
import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import { userVideoTaskState } from "./userWorkspacePresentation";

export function UserTaskStatus({
  artifacts,
  snapshot,
}: {
  artifacts: readonly AgentArtifactItem[];
  snapshot: AgentSessionSnapshot;
}) {
  const state = userVideoTaskState(snapshot, artifacts);
  const text = useAgentWorkspaceText();
  if (!state) return null;
  const failed =
    state === "failed" ||
    state === "partial" ||
    state === "task_failed" ||
    state === "verification_failed";
  const Icon =
    state === "processing"
      ? LoaderCircle
      : state === "verified"
        ? CheckCircle2
        : AlertTriangle;
  return (
    <section
      aria-live={failed ? "assertive" : "polite"}
      className={`border-b px-4 py-2.5 ${
        failed
          ? "border-amber-300 bg-amber-50 text-amber-950"
          : state === "verified"
            ? "border-emerald-200 bg-emerald-50 text-emerald-950"
            : "border-border bg-muted/30"
      }`}
      role={failed ? "alert" : "status"}
    >
      <div className="mx-auto flex max-w-4xl items-start gap-2.5 sm:items-center">
        <Icon
          aria-hidden="true"
          className={`mt-0.5 size-4 shrink-0 ${
            state === "processing"
              ? "animate-spin motion-reduce:animate-none"
              : ""
          }`}
        />
        <div className="sm:flex sm:items-baseline sm:gap-2">
          <div className="text-sm font-medium">
            {text.videoTaskStatus[state].title}
          </div>
          <div className="mt-0.5 text-xs opacity-80 sm:mt-0">
            {text.videoTaskStatus[state].detail}
          </div>
        </div>
      </div>
    </section>
  );
}
