"use client";

import { CenteredSpinner } from "@/components/ui/spinner";
import { usePod } from "@/stores/pod";
import { AgentPanel } from "./AgentPanel";
import { TerminalPane } from "./TerminalPane";

interface MobileWorkspacePaneProps {
  paneId: string;
  podKey: string;
  onClose: () => void;
}

export function MobileWorkspacePane({
  paneId,
  podKey,
  onClose,
}: MobileWorkspacePaneProps) {
  const pod = usePod(podKey);

  if (!pod) {
    return <CenteredSpinner className="h-full bg-background" />;
  }

  const commonProps = {
    className: "h-full rounded-none border-0 ring-0",
    controlClientLabel: "mobile-workspace",
    isActive: true,
    onClose,
    paneId,
    podKey,
    showHeader: false,
  };

  return pod.interaction_mode === "acp" ? (
    <AgentPanel {...commonProps} />
  ) : (
    <TerminalPane {...commonProps} allowSplit={false} />
  );
}
