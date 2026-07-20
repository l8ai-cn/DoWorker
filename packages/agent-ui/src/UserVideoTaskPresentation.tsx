import type {
  AgentArtifactItem,
  AgentSessionSnapshot,
} from "./contracts";
import {
  AgentWorkspaceLocaleProvider,
} from "./AgentWorkspaceLocaleContext";
import type { AgentWorkspaceLocale } from "./agentWorkspaceText";
import { UserTaskStatus } from "./UserTaskStatus";
import { UserVideoExecutionTrace } from "./VideoExecutionTrace";
import { userVideoExecutionSteps } from "./userVideoExecutionTrace";

export interface UserVideoTaskPresentationProps {
  artifacts: readonly AgentArtifactItem[];
  locale?: AgentWorkspaceLocale;
  snapshot: AgentSessionSnapshot;
}

export function UserVideoTaskPresentation({
  artifacts,
  locale = "en-US",
  snapshot,
}: UserVideoTaskPresentationProps) {
  const steps = userVideoExecutionSteps(snapshot, artifacts);
  return (
    <AgentWorkspaceLocaleProvider locale={locale}>
      <UserTaskStatus artifacts={artifacts} snapshot={snapshot} />
      <UserVideoExecutionTrace steps={steps} />
    </AgentWorkspaceLocaleProvider>
  );
}
