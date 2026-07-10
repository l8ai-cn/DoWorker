"use client";

import { useEffect, useState } from "react";
import { AlertCircle } from "lucide-react";
import { usePod, usePodStore } from "@/stores/pod";
import { CenteredSpinner } from "@/components/ui/spinner";
import { TerminalPane } from "@/components/workspace/TerminalPane";
import { AgentPanel } from "@/components/workspace/AgentPanel";
import { getPodDisplayName } from "@/lib/pod-display-name";

interface MobilePodWorkspaceProps {
  podKey: string;
}

function interactionMode(pod: unknown): string {
  if (!pod || typeof pod !== "object") return "";
  const value = (pod as Record<string, unknown>).interaction_mode;
  return typeof value === "string" ? value : "";
}

export function MobilePodWorkspace({ podKey }: MobilePodWorkspaceProps) {
  const pod = usePod(podKey);
  const fetchPod = usePodStore((s) => s.fetchPod);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!podKey || pod) return;
    void fetchPod(podKey).catch((err) => {
      setError(err instanceof Error ? err.message : "Failed to load pod");
    });
  }, [fetchPod, pod, podKey]);

  if (error) {
    return (
      <div className="flex h-full items-center justify-center bg-background p-6 text-center">
        <div className="max-w-sm space-y-3">
          <AlertCircle className="mx-auto h-10 w-10 text-destructive" />
          <p className="text-sm font-medium text-foreground">{error}</p>
        </div>
      </div>
    );
  }

  if (!pod) {
    return (
      <div data-testid="mobile-pod-loading" className="h-full">
        <CenteredSpinner className="h-full bg-background" />
      </div>
    );
  }

  const paneId = `mobile-${podKey}`;
  const commonProps = {
    paneId,
    podKey,
    isActive: true,
    showHeader: false,
    className: "rounded-none border-0 ring-0",
  };

  return (
    <div className="flex h-full min-h-0 flex-col bg-background">
      <div className="flex h-11 shrink-0 items-center border-b border-border/60 px-3">
        <span className="truncate font-mono text-sm font-medium text-foreground">
          {getPodDisplayName(pod)}
        </span>
      </div>
      <div className="min-h-0 flex-1 overflow-hidden">
        {interactionMode(pod) === "acp" ? (
          <AgentPanel {...commonProps} />
        ) : (
          <TerminalPane {...commonProps} allowSplit={false} />
        )}
      </div>
    </div>
  );
}
